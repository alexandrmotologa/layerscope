package advisor

import (
	"testing"

	"github.com/alexandrmotologa/layerscope/pkg/oci"
	"github.com/alexandrmotologa/layerscope/pkg/security"
	"github.com/alexandrmotologa/layerscope/pkg/vfs"
)

func TestCalculateEfficiency(t *testing.T) {
	score, grade := CalculateEfficiency(1000, 100)
	if score != 90.0 || grade != "A" {
		t.Errorf("expected 90%% and A, got %.1f%% and %s", score, grade)
	}

	score0, grade0 := CalculateEfficiency(1000, 950)
	if score0 != 5.0 || grade0 != "F" {
		t.Errorf("expected 5%% and F, got %.1f%% and %s", score0, grade0)
	}
}

func TestAdvisorRules_WithSampleImage(t *testing.T) {
	sampleImg, err := oci.GenerateSampleImage(nil)
	if err != nil {
		t.Fatalf("GenerateSampleImage failed: %v", err)
	}

	snapshots := vfs.BuildLayerSnapshots(sampleImg)
	wasteSummary := vfs.CalculateWastedSpace(snapshots, nil)
	finalTree := snapshots[len(snapshots)-1].Tree

	secReport := &security.SecurityAuditReport{
		IntermediateLeaks: 1,
		RunsAsRoot:        true,
	}

	report := AnalyzeImageHeuristics(sampleImg, wasteSummary, secReport, finalTree)

	if len(report.Recommendations) == 0 {
		t.Fatalf("expected recommendations for sample image, got 0")
	}

	var hasIntermediateLeakRec, hasRootRec, hasPkgCacheRec bool
	for _, r := range report.Recommendations {
		if r.ID == "intermediate-secret-leak" {
			hasIntermediateLeakRec = true
		}
		if r.ID == "run-as-root" {
			hasRootRec = true
		}
		if r.ID == "uncleaned-pkg-cache" {
			hasPkgCacheRec = true
		}
	}

	if !hasIntermediateLeakRec {
		t.Errorf("expected recommendation for intermediate secret leak")
	}
	if !hasRootRec {
		t.Errorf("expected recommendation for running as root")
	}
	if !hasPkgCacheRec {
		t.Errorf("expected recommendation for uncleaned package cache (/root/.npm)")
	}
}
