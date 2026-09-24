package security

import (
	"bufio"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/alexandrmotologa/layerscope/pkg/oci"
	"github.com/alexandrmotologa/layerscope/pkg/vfs"
)

// Scanner inspects layer files for sensitive credentials.
type Scanner struct {
	rules    []SecretRule
	findings []*SecretFinding
}

// NewScanner initializes a secret scanner with standard rules.
func NewScanner() *Scanner {
	return &Scanner{
		rules: BuiltinSecretRules(),
	}
}

// ShouldScanPath checks if a file path is a viable candidate for secrets (skipping binary assets).
func ShouldScanPath(filePath string) bool {
	clean := path.Clean(filePath)
	base := strings.ToLower(path.Base(clean))
	ext := strings.ToLower(path.Ext(clean))

	// Always scan known sensitive filenames
	if base == ".env" || strings.HasPrefix(base, ".env.") ||
		base == ".npmrc" || base == ".netrc" || base == "id_rsa" ||
		base == "id_ed25519" || base == "config.json" || base == "credentials" {
		return true
	}

	// Skip common binary media and compiled extensions
	binaryExts := map[string]bool{
		".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".webp": true,
		".ico": true, ".pdf": true, ".zip": true, ".gz": true, ".tar": true,
		".so": true, ".dylib": true, ".dll": true, ".exe": true, ".bin": true,
		".woff": true, ".woff2": true, ".ttf": true, ".eot": true,
	}

	if binaryExts[ext] {
		return false
	}

	return true
}

// ScanStream reads content line by line and evaluates against all secret rules.
func (s *Scanner) ScanStream(layerIndex int, command string, filePath string, r io.Reader) ([]*SecretFinding, error) {
	if !ShouldScanPath(filePath) {
		return nil, nil
	}

	scanner := bufio.NewScanner(r)
	// Cap line size to 64KB
	buf := make([]byte, 64*1024)
	scanner.Buffer(buf, 64*1024)

	var findings []*SecretFinding
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		for _, rule := range s.rules {
			// If rule specifies a path regex, verify path match first
			if rule.PathRegex != nil && !rule.PathRegex.MatchString(filePath) {
				continue
			}

			loc := rule.Regex.FindStringSubmatchIndex(line)
			if loc != nil {
				matchedStr := line[loc[0]:loc[1]]
				// If regex has a capturing group, mask the group instead of entire pattern
				if len(loc) >= 4 && loc[2] >= 0 && loc[3] >= 0 {
					matchedStr = line[loc[2]:loc[3]]
				}

				findingID := fmt.Sprintf("sec-%d-%d-%s", layerIndex, lineNum, rule.ID)
				masked := MaskSecret(matchedStr)

				findings = append(findings, &SecretFinding{
					ID:          findingID,
					RuleName:    rule.Name,
					Severity:    rule.Severity,
					FilePath:    filePath,
					LayerIndex:  layerIndex,
					Command:     command,
					LineNumber:  lineNum,
					MatchMasked: masked,
					Description: rule.Description,
				})
			}
		}
	}

	if len(findings) > 0 {
		s.findings = append(s.findings, findings...)
	}

	if err := scanner.Err(); err != nil && err != io.EOF {
		return findings, fmt.Errorf("read stream for %s: %w", filePath, err)
	}

	return findings, nil
}

// AddFindings appends findings manually to the scanner.
func (s *Scanner) AddFindings(f ...*SecretFinding) {
	s.findings = append(s.findings, f...)
}

// Hook returns an oci.FileInspectHook that runs the secret scanner during layer extraction.
func (s *Scanner) Hook(commands []string) oci.FileInspectHook {
	return func(layerIndex int, file *oci.LayerFile, reader io.Reader) error {
		cmd := ""
		if layerIndex < len(commands) {
			cmd = commands[layerIndex]
		}

		_, err := s.ScanStream(layerIndex, cmd, file.Path, reader)
		return err
	}
}

// FinalizeAudit compiles the security findings and marks intermediate leaks using the final VFS state.
func (s *Scanner) FinalizeAudit(finalTree *vfs.VFSTree, runsAsRoot bool, suidBinaries []string) *SecurityAuditReport {
	report := &SecurityAuditReport{
		RunsAsRoot:        runsAsRoot,
		SuidBinariesCount: len(suidBinaries),
		SuidBinaries:      suidBinaries,
		Findings:          s.findings,
		TotalFindings:     len(s.findings),
	}

	for _, f := range s.findings {
		switch f.Severity {
		case SeverityCritical:
			report.CriticalCount++
		case SeverityHigh:
			report.HighCount++
		case SeverityMedium:
			report.MediumCount++
		case SeverityLow:
			report.LowCount++
		}

		// Check if file exists in the final cumulative tree
		if finalTree != nil {
			finalNode := finalTree.Lookup(f.FilePath)
			if finalNode == nil {
				f.IsDeletedInFinal = true
				report.IntermediateLeaks++
			}
		}
	}

	return report
}
