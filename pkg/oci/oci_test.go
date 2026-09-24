package oci

import (
	"archive/tar"
	"bytes"
	"io"
	"testing"
	"time"
)

func TestSampleImageGeneration(t *testing.T) {
	hookCount := 0
	hook := func(layerIndex int, file *LayerFile, reader io.Reader) error {
		hookCount++
		// verify stream can be read
		_, err := io.ReadAll(reader)
		return err
	}

	analysis, err := GenerateSampleImage(hook)
	if err != nil {
		t.Fatalf("GenerateSampleImage returned error: %v", err)
	}

	if len(analysis.Layers) != 6 {
		t.Errorf("expected 6 layers, got %d", len(analysis.Layers))
	}

	if analysis.TotalSizeBytes <= 0 {
		t.Errorf("expected positive total size, got %d", analysis.TotalSizeBytes)
	}

	if hookCount == 0 {
		t.Errorf("expected file hook to be called, got 0 calls")
	}

	// Verify whiteout was detected in layer 4
	layer4 := analysis.Layers[4]
	var foundWhiteout bool
	for _, f := range layer4.Files {
		if f.IsWhiteout && f.WhiteoutTarget == "/app/.env" {
			foundWhiteout = true
			break
		}
	}

	if !foundWhiteout {
		t.Errorf("expected to find whiteout for /app/.env in layer 4")
	}
}

func TestUnpackLayerTar_Whiteouts(t *testing.T) {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	now := time.Now()

	// Normal directory
	_ = tw.WriteHeader(&tar.Header{
		Name:     "etc/",
		Typeflag: tar.TypeDir,
		Mode:     0755,
		ModTime:  now,
	})

	// Whiteout file
	_ = tw.WriteHeader(&tar.Header{
		Name:     "etc/.wh.shadow",
		Typeflag: tar.TypeReg,
		Size:     0,
		Mode:     0644,
		ModTime:  now,
	})

	// Opaque marker
	_ = tw.WriteHeader(&tar.Header{
		Name:     "etc/.wh..wh..opq",
		Typeflag: tar.TypeReg,
		Size:     0,
		Mode:     0644,
		ModTime:  now,
	})

	_ = tw.Close()

	files, _, err := UnpackLayerTar(0, buf, nil)
	if err != nil {
		t.Fatalf("UnpackLayerTar failed: %v", err)
	}

	if len(files) != 3 {
		t.Fatalf("expected 3 files, got %d", len(files))
	}

	var hasWhiteout, hasOpaque bool
	for _, f := range files {
		if f.IsWhiteout && f.WhiteoutTarget == "/etc/shadow" {
			hasWhiteout = true
		}
		if f.IsOpaque && f.Path == "/etc" {
			hasOpaque = true
		}
	}

	if !hasWhiteout {
		t.Errorf("expected whiteout for /etc/shadow")
	}
	if !hasOpaque {
		t.Errorf("expected opaque marker for /etc")
	}
}
