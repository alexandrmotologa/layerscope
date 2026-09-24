package sbom

import (
	"bytes"
	"context"
	"io"
	"path"
	"strings"

	"github.com/alexandrmotologa/layerscope/pkg/oci"
	"github.com/alexandrmotologa/layerscope/pkg/vfs"
	"github.com/alexandrmotologa/layerscope/pkg/vuln"
)

// Extractor captures and processes software package manifests across container layers.
type Extractor struct {
	manifestBuffers map[string][]byte
}

// NewExtractor creates a new SBOM manifest collector.
func NewExtractor() *Extractor {
	return &Extractor{
		manifestBuffers: make(map[string][]byte),
	}
}

// IsManifestPath checks if a file path is a package database or dependency lockfile.
func IsManifestPath(p string) bool {
	clean := path.Clean("/" + p)
	base := strings.ToLower(path.Base(clean))

	if clean == "/lib/apk/db/installed" || clean == "/var/lib/dpkg/status" {
		return true
	}
	if base == "package-lock.json" || base == "requirements.txt" {
		return true
	}
	return false
}

// Hook returns an oci.FileInspectHook capturing relevant manifest streams during layer unpacking.
func (e *Extractor) Hook() oci.FileInspectHook {
	return func(layerIndex int, file *oci.LayerFile, reader io.Reader) error {
		if !IsManifestPath(file.Path) {
			return nil
		}

		data, err := io.ReadAll(reader)
		if err != nil {
			return err
		}

		// Keep the latest version of the manifest written across layers
		e.manifestBuffers[file.Path] = data
		return nil
	}
}

// FinalizeReport parses all collected manifest streams into a unified SBOMReport.
func (e *Extractor) FinalizeReport(imageName string, finalTree *vfs.VFSTree) (*SBOMReport, error) {
	report := &SBOMReport{
		ImageName: imageName,
	}

	seenPackages := make(map[string]bool)

	for manifestPath, data := range e.manifestBuffers {
		// If finalTree is provided, verify manifest wasn't deleted in final container state
		if finalTree != nil && finalTree.Lookup(manifestPath) == nil {
			continue
		}

		clean := path.Clean(manifestPath)
		base := strings.ToLower(path.Base(clean))
		reader := bytes.NewReader(data)

		var parsed []*Package
		var err error

		if clean == "/lib/apk/db/installed" {
			parsed, err = ParseAlpineInstalled(reader)
		} else if clean == "/var/lib/dpkg/status" {
			parsed, err = ParseDebianStatus(reader)
		} else if base == "package-lock.json" {
			parsed, err = ParseNPMLockfile(reader, clean)
		} else if base == "requirements.txt" {
			parsed, err = ParsePythonRequirements(reader, clean)
		}

		if err != nil {
			continue // Skip corrupted or partial manifests
		}

		for _, pkg := range parsed {
			key := string(pkg.Type) + ":" + pkg.Name + "@" + pkg.Version
			if seenPackages[key] {
				continue
			}
			seenPackages[key] = true

			if pkg.Type == TypeAlpine || pkg.Type == TypeDebian || pkg.Type == TypeRPM {
				report.OSPackages++
			} else {
				report.AppPackages++
			}

			report.Packages = append(report.Packages, pkg)
		}
	}

	report.TotalPackages = len(report.Packages)
	return report, nil
}

// EnrichWithVulnerabilities queries OSV.dev to populate CVE and security advisory metadata.
func EnrichWithVulnerabilities(ctx context.Context, report *SBOMReport) {
	if report == nil || len(report.Packages) == 0 {
		return
	}

	var purls []string
	for _, pkg := range report.Packages {
		if pkg.PURL != "" {
			purls = append(purls, pkg.PURL)
		}
	}

	client := vuln.NewClient()
	results, err := client.QueryBatch(ctx, purls)
	if err != nil || len(results) == 0 {
		return // gracefully continue if offline or network unreachable
	}

	totalVulns := 0
	for _, pkg := range report.Packages {
		if vList, ok := results[pkg.PURL]; ok && len(vList) > 0 {
			pkg.Vulnerabilities = vList
			totalVulns += len(vList)
		}
	}
	report.TotalVulnerabilities = totalVulns
}

// EnrichWithVulnerabilities delegates to the package-level EnrichWithVulnerabilities function.
func (e *Extractor) EnrichWithVulnerabilities(ctx context.Context, report *SBOMReport) {
	EnrichWithVulnerabilities(ctx, report)
}

