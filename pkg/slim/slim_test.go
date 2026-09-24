package slim

import (
	"archive/tar"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/alexandrmotologa/layerscope/pkg/oci"
	"github.com/alexandrmotologa/layerscope/pkg/vfs"
)

func TestExportSquashedImage(t *testing.T) {
	img, err := oci.GenerateSampleImage(nil)
	if err != nil {
		t.Fatalf("GenerateSampleImage failed: %v", err)
	}

	snapshots := vfs.BuildLayerSnapshots(img)
	finalTree := snapshots[len(snapshots)-1].Tree

	tmpDir, err := os.MkdirTemp("", "layerscope-slim-test-*")
	if err != nil {
		t.Fatalf("MkdirTemp failed: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	outTar := filepath.Join(tmpDir, "test-slim.tar")
	res, err := ExportSquashedImage(img, finalTree, outTar, Options{
		Tag:         "test/app:slim",
		PurgeCaches: true,
	})
	if err != nil {
		t.Fatalf("ExportSquashedImage failed: %v", err)
	}

	if res.LayerCount != 1 {
		t.Errorf("expected 1 layer in slim image, got %d", res.LayerCount)
	}
	if res.SlimSize <= 0 {
		t.Errorf("expected positive slim size, got %d", res.SlimSize)
	}

	// Verify tar structure
	f, err := os.Open(outTar)
	if err != nil {
		t.Fatalf("failed to open output tar: %v", err)
	}
	defer f.Close()

	tr := tar.NewReader(f)
	foundManifest := false
	foundLayerTar := false

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("tar read error: %v", err)
		}
		if hdr.Name == "manifest.json" {
			foundManifest = true
		}
		if hdr.Name == "layer.tar" {
			foundLayerTar = true
		}
	}

	if !foundManifest {
		t.Errorf("expected manifest.json in slim archive")
	}
	if !foundLayerTar {
		t.Errorf("expected layer.tar in slim archive")
	}
}
