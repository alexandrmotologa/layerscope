package security

import (
	"strings"
	"testing"

	"github.com/alexandrmotologa/layerscope/pkg/oci"
	"github.com/alexandrmotologa/layerscope/pkg/vfs"
)

func TestSecretRulesAndMasking(t *testing.T) {
	scanner := NewScanner()

	sampleData := `
AWS_KEY=AKIAIOSFODNN7EXAMPLE
SECRET=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
GH_TOKEN=ghp_ABC123456789012345678901234567890123456
PG_URL=postgres://app_user:SuperSecretPassword123!@db.internal:5432/payments
`
	findings, err := scanner.ScanStream(0, "COPY .env /app/.env", "/app/.env", strings.NewReader(sampleData))
	if err != nil {
		t.Fatalf("ScanStream returned error: %v", err)
	}

	if len(findings) == 0 {
		t.Fatalf("expected to find credentials in sample data, got 0")
	}

	var foundAWS, foundGH, foundDB bool
	for _, f := range findings {
		if f.RuleName == "AWS Access Key ID" {
			foundAWS = true
			if !strings.HasPrefix(f.MatchMasked, "AKIA") {
				t.Errorf("expected masked AWS key to preserve AKIA prefix, got %s", f.MatchMasked)
			}
			if strings.Contains(f.MatchMasked, "EXAMPLE") {
				t.Errorf("masked key must not reveal secret suffix, got %s", f.MatchMasked)
			}
		}
		if f.RuleName == "GitHub Personal Access Token" {
			foundGH = true
			if !strings.HasPrefix(f.MatchMasked, "ghp_") {
				t.Errorf("expected masked GH token to preserve ghp_ prefix, got %s", f.MatchMasked)
			}
		}
		if f.RuleName == "Database Connection URI with Password" {
			foundDB = true
		}
	}

	if !foundAWS {
		t.Errorf("expected AWS Access Key ID to be detected")
	}
	if !foundGH {
		t.Errorf("expected GitHub PAT to be detected")
	}
	if !foundDB {
		t.Errorf("expected Database URI to be detected")
	}
}

func TestIntermediateSecretLeakDetection(t *testing.T) {
	scanner := NewScanner()

	// Ingest sample image with scanner hook
	var commands []string
	sampleImg, err := oci.GenerateSampleImage(nil)
	if err != nil {
		t.Fatalf("failed to generate sample: %v", err)
	}
	for _, l := range sampleImg.Layers {
		commands = append(commands, l.Command)
	}

	hook := scanner.Hook(commands)
	_, err = oci.GenerateSampleImage(hook)
	if err != nil {
		t.Fatalf("sample generation with hook failed: %v", err)
	}

	snapshots := vfs.BuildLayerSnapshots(sampleImg)
	finalTree := snapshots[len(snapshots)-1].Tree

	runsAsRoot, suid := AuditPrivileges(&sampleImg.Config, finalTree)
	report := scanner.FinalizeAudit(finalTree, runsAsRoot, suid)

	if report.TotalFindings == 0 {
		t.Fatalf("expected secret findings from layer 2, got 0")
	}

	// Verify that the secret was flagged as an intermediate leak because .env was deleted in layer 4
	if report.IntermediateLeaks == 0 {
		t.Errorf("expected intermediate leaks count > 0 because .env was deleted in layer 4")
	}

	var foundDeletedSecret bool
	for _, f := range report.Findings {
		if f.FilePath == "/app/.env" && f.IsDeletedInFinal {
			foundDeletedSecret = true
			break
		}
	}

	if !foundDeletedSecret {
		t.Errorf("expected /app/.env finding to have IsDeletedInFinal = true")
	}

	if !report.RunsAsRoot {
		t.Errorf("expected sample image to be flagged as running as root (User 0)")
	}
}
