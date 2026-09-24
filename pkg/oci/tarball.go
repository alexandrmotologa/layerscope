package oci

import (
	"archive/tar"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"
)

type dockerManifestItem struct {
	Config   string   `json:"Config"`
	RepoTags []string `json:"RepoTags"`
	Layers   []string `json:"Layers"`
}

// LoadImageFromTarball parses an exported image tarball (from `docker save` or OCI archive).
func LoadImageFromTarball(tarballPath string, hook FileInspectHook) (*ImageAnalysis, error) {
	file, err := os.Open(tarballPath)
	if err != nil {
		return nil, fmt.Errorf("open tarball: %w", err)
	}
	defer file.Close()

	// First pass: locate manifest.json and the config JSON
	tarReader := tar.NewReader(file)
	var manifestItems []dockerManifestItem
	var configBytes []byte
	var layerTarStreams = make(map[string][]byte)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read tarball entry: %w", err)
		}

		cleanName := strings.TrimPrefix(path.Clean(header.Name), "/")

		if cleanName == "manifest.json" {
			data, err := io.ReadAll(tarReader)
			if err != nil {
				return nil, fmt.Errorf("read manifest.json: %w", err)
			}
			if err := json.Unmarshal(data, &manifestItems); err != nil {
				return nil, fmt.Errorf("parse manifest.json: %w", err)
			}
		} else if strings.HasSuffix(cleanName, ".json") && !strings.Contains(cleanName, "/") {
			// Potential config file in legacy docker format
			data, err := io.ReadAll(tarReader)
			if err != nil {
				return nil, fmt.Errorf("read config json: %w", err)
			}
			configBytes = data
		}
	}

	if len(manifestItems) == 0 {
		return nil, fmt.Errorf("invalid archive: manifest.json not found in %s", tarballPath)
	}

	activeManifest := manifestItems[0]

	// Reset file pointer to read config and layers
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("seek tarball: %w", err)
	}

	tarReader = tar.NewReader(file)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read tarball entry on pass 2: %w", err)
		}

		cleanName := strings.TrimPrefix(path.Clean(header.Name), "/")
		if cleanName == activeManifest.Config {
			configBytes, err = io.ReadAll(tarReader)
			if err != nil {
				return nil, fmt.Errorf("read manifest config %s: %w", cleanName, err)
			}
		}

		for _, layerPath := range activeManifest.Layers {
			if cleanName == layerPath {
				layerData, err := io.ReadAll(tarReader)
				if err != nil {
					return nil, fmt.Errorf("read layer %s: %w", cleanName, err)
				}
				layerTarStreams[layerPath] = layerData
				break
			}
		}
	}

	var imgConfig ImageConfig
	if len(configBytes) > 0 {
		if err := json.Unmarshal(configBytes, &imgConfig); err != nil {
			return nil, fmt.Errorf("parse image config: %w", err)
		}
	}

	// Unpack layers
	var layers []*Layer
	var totalSize int64
	var totalFiles int

	historyIndex := 0
	for i, layerPath := range activeManifest.Layers {
		layerData, exists := layerTarStreams[layerPath]
		if !exists {
			return nil, fmt.Errorf("missing layer data for %s", layerPath)
		}

		files, uncompressedSize, err := UnpackLayerTar(i, bytes.NewReader(layerData), hook)
		if err != nil {
			return nil, fmt.Errorf("unpack layer %d (%s): %w", i, layerPath, err)
		}

		cmd := ""
		createdBy := ""
		for historyIndex < len(imgConfig.History) {
			h := imgConfig.History[historyIndex]
			historyIndex++
			if !h.EmptyLayer {
				cmd = h.CreatedBy
				createdBy = h.CreatedBy
				break
			}
		}

		layer := &Layer{
			Index:       i,
			Digest:      layerPath,
			Size:        uncompressedSize,
			Command:     cmd,
			CreatedBy:   createdBy,
			Files:       files,
			FileCount:   len(files),
			WastedBytes: 0,
		}

		layers = append(layers, layer)
		totalSize += uncompressedSize
		totalFiles += len(files)
	}

	repoTag := "archive:latest"
	if len(activeManifest.RepoTags) > 0 {
		repoTag = activeManifest.RepoTags[0]
	}

	return &ImageAnalysis{
		Reference: ImageReference{
			Original:     repoTag,
			Repository:   repoTag,
			Tag:          "latest",
			Source:       SourceTarball,
			Architecture: imgConfig.Architecture,
			OS:           imgConfig.OS,
		},
		Config:          imgConfig,
		Layers:          layers,
		TotalSizeBytes:  totalSize,
		TotalFilesCount: totalFiles,
		AnalyzedAt:      time.Now(),
	}, nil
}
