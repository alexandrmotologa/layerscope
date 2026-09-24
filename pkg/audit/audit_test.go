package audit

import (
	"strings"
	"testing"

	"github.com/alexandrmotologa/layerscope/pkg/oci"
	"github.com/alexandrmotologa/layerscope/pkg/sbom"
	"github.com/alexandrmotologa/layerscope/pkg/security"
	"github.com/alexandrmotologa/layerscope/pkg/vfs"
)

func TestAuditReportFormats(t *testing.T) {
	sampleImg, err := oci.GenerateSampleImage(nil)
	if err != nil {
		t.Fatalf("GenerateSampleImage failed: %v", err)
	}

	snapshots := vfs.BuildLayerSnapshots(sampleImg)
	wasteSummary := vfs.CalculateWastedSpace(snapshots, nil)
	secReport := &security.SecurityAuditReport{
		TotalFindings:     1,
		CriticalCount:     1,
		IntermediateLeaks: 1,
		RunsAsRoot:        true,
		Findings: []*security.SecretFinding{
			{
				ID:               "sec-1",
				RuleName:         "AWS Access Key ID",
				Severity:         security.SeverityCritical,
				FilePath:         "/app/.env",
				LayerIndex:       2,
				MatchMasked:      "AKIAIOSFODNN********",
				IsDeletedInFinal: true,
			},
		},
	}
	sbomReport := &sbom.SBOMReport{
		ImageName:     "sample:latest",
		TotalPackages: 5,
		OSPackages:    3,
		AppPackages:   2,
	}

	res := GenerateFullAudit(sampleImg, snapshots, wasteSummary, secReport, sbomReport)

	// JSON format test
	jsonBytes, err := res.FormatJSON()
	if err != nil {
		t.Fatalf("FormatJSON failed: %v", err)
	}
	if !strings.Contains(string(jsonBytes), "efficiencyScore") {
		t.Errorf("JSON output missing efficiencyScore: %s", string(jsonBytes))
	}

	// Markdown format test
	md := res.FormatMarkdown()
	if !strings.Contains(md, "LayerScope Audit:") || !strings.Contains(md, "AWS Access Key ID") {
		t.Errorf("Markdown output missing expected sections: %s", md)
	}

	// JUnit format test
	junitBytes, err := res.FormatJUnit(85.0, true)
	if err != nil {
		t.Fatalf("FormatJUnit failed: %v", err)
	}
	if !strings.Contains(string(junitBytes), "<testsuites") || !strings.Contains(string(junitBytes), "failure") {
		t.Errorf("JUnit output missing expected XML: %s", string(junitBytes))
	}

	// HTML format test
	html := res.FormatHTML()
	if !strings.Contains(html, "<!DOCTYPE html>") || !strings.Contains(html, "LayerScope Audit Report") {
		t.Errorf("HTML output missing expected template elements: %s", html)
	}
}
