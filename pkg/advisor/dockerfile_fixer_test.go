package advisor

import (
	"strings"
	"testing"

	"github.com/alexandrmotologa/layerscope/pkg/oci"
	"github.com/alexandrmotologa/layerscope/pkg/vfs"
)

func TestGenerateDockerfileOptimization(t *testing.T) {
	img, err := oci.GenerateSampleImage(nil)
	if err != nil {
		t.Fatalf("GenerateSampleImage failed: %v", err)
	}

	snapshots := vfs.BuildLayerSnapshots(img)
	waste := vfs.CalculateWastedSpace(snapshots, nil)
	finalTree := snapshots[len(snapshots)-1].Tree

	opt := GenerateDockerfileOptimization(img, waste, finalTree)
	if opt == nil {
		t.Fatal("expected non-nil DockerfileOptimization")
	}

	if !strings.Contains(opt.OptimizedDockerfile, "FROM node:20-alpine") {
		t.Errorf("expected node base image, got:\n%s", opt.OptimizedDockerfile)
	}

	if !strings.Contains(opt.OptimizedDockerfile, "USER appuser") {
		t.Errorf("expected non-root user appuser, got:\n%s", opt.OptimizedDockerfile)
	}

	if !strings.Contains(opt.GeneratedDockerignore, "node_modules") {
		t.Errorf("expected node_modules in dockerignore, got:\n%s", opt.GeneratedDockerignore)
	}

	if len(opt.Improvements) == 0 {
		t.Errorf("expected non-empty improvements list")
	}

	if opt.EstimatedSavingsMB <= 0 {
		t.Errorf("expected estimated savings > 0, got %f", opt.EstimatedSavingsMB)
	}
}
