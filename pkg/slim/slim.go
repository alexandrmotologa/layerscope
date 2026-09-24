package slim

import (
	"archive/tar"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/alexandrmotologa/layerscope/pkg/oci"
	"github.com/alexandrmotologa/layerscope/pkg/vfs"
)

// Options configures image squashing and cache pruning behavior.
type Options struct {
	Tag         string
	PurgeCaches bool
}

// Result contains metrics on image reduction and the generated squashed archive.
type Result struct {
	OriginalSize    int64   `json:"originalSize"`
	SlimSize        int64   `json:"slimSize"`
	SavedBytes      int64   `json:"savedBytes"`
	SavedPercentage float64 `json:"savedPercentage"`
	OutputFile      string  `json:"outputFile"`
	LayerCount      int     `json:"layerCount"`
}

// ExportSquashedImage squashes multi-layer cumulative files into a clean, single-layer Docker/OCI tar archive.
// It purges intermediate whiteout-deleted artifacts (including leaked credentials),
// removes build cache files, and creates a loadable image for Docker and Podman.
func ExportSquashedImage(img *oci.ImageAnalysis, finalTree *vfs.VFSTree, outPath string, opts Options) (*Result, error) {
	if img == nil || finalTree == nil {
		return nil, fmt.Errorf("nil image or final tree provided")
	}

	if outPath == "" {
		outPath = "slim-image.tar"
	}

	// Ensure parent dir exists
	if dir := filepath.Dir(outPath); dir != "." && dir != "" {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("create output directory: %w", err)
		}
	}

	outFile, err := os.Create(outPath)
	if err != nil {
		return nil, fmt.Errorf("create slim tar file: %w", err)
	}
	defer outFile.Close()

	tw := tar.NewWriter(outFile)
	defer tw.Close()

	// 1. Build squashed single layer tar
	layerTarPath := "layer.tar"
	tempLayerFile, err := os.CreateTemp("", "layerscope-layer-*.tar")
	if err != nil {
		return nil, fmt.Errorf("create temp layer file: %w", err)
	}
	defer os.Remove(tempLayerFile.Name())
	defer tempLayerFile.Close()

	layerTW := tar.NewWriter(tempLayerFile)
	layerHasher := sha256.New()

	// Write directories first, then files
	nodes := finalTree.Index
	sortedPaths := make([]string, 0, len(nodes))
	for p := range nodes {
		sortedPaths = append(sortedPaths, p)
	}

	for _, p := range sortedPaths {
		node := nodes[p]
		if node.Path == "/" || node.Path == "" {
			continue
		}

		// Optional cache purging
		if opts.PurgeCaches && isCachePath(node.Path) {
			continue
		}

		relPath := strings.TrimPrefix(node.Path, "/")
		modTime := node.ModTime
		if modTime.IsZero() {
			modTime = time.Now()
		}

		if node.IsDir {
			hdr := &tar.Header{
				Name:     relPath + "/",
				Mode:     0755,
				Typeflag: tar.TypeDir,
				ModTime:  modTime,
			}
			if err := layerTW.WriteHeader(hdr); err != nil {
				return nil, fmt.Errorf("write dir header %s: %w", p, err)
			}
		} else if node.IsSymlink {
			hdr := &tar.Header{
				Name:     relPath,
				Mode:     0777,
				Typeflag: tar.TypeSymlink,
				Linkname: node.LinkTarget,
				ModTime:  modTime,
			}
			if err := layerTW.WriteHeader(hdr); err != nil {
				return nil, fmt.Errorf("write symlink header %s: %w", p, err)
			}
		} else {
			// Regular file
			content := node.Data
			hdr := &tar.Header{
				Name:     relPath,
				Mode:     int64(node.Mode.Perm()),
				Size:     int64(len(content)),
				Typeflag: tar.TypeReg,
				ModTime:  modTime,
			}
			if hdr.Mode == 0 {
				hdr.Mode = 0644
			}
			if err := layerTW.WriteHeader(hdr); err != nil {
				return nil, fmt.Errorf("write file header %s: %w", p, err)
			}
			if len(content) > 0 {
				if _, err := layerTW.Write(content); err != nil {
					return nil, fmt.Errorf("write file content %s: %w", p, err)
				}
			}
		}
	}

	if err := layerTW.Close(); err != nil {
		return nil, fmt.Errorf("close layer tar writer: %w", err)
	}

	// Calculate layer SHA256 diff_id
	if _, err := tempLayerFile.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("seek layer tar: %w", err)
	}
	layerStat, err := tempLayerFile.Stat()
	if err != nil {
		return nil, fmt.Errorf("stat layer tar: %w", err)
	}

	layerBuf := make([]byte, 32*1024)
	for {
		n, rErr := tempLayerFile.Read(layerBuf)
		if n > 0 {
			layerHasher.Write(layerBuf[:n])
		}
		if rErr != nil {
			break
		}
	}
	diffID := "sha256:" + hex.EncodeToString(layerHasher.Sum(nil))

	// 2. Generate OCI config.json
	repoTag := opts.Tag
	if repoTag == "" {
		if img.Reference.Repository != "" {
			repoTag = img.Reference.Repository + ":slim"
		} else {
			repoTag = "layerscope-squashed:latest"
		}
	}

	type rootFS struct {
		Type    string   `json:"type"`
		DiffIDs []string `json:"diff_ids"`
	}

	type slimConfig struct {
		Architecture string              `json:"architecture"`
		OS           string              `json:"os"`
		Created      time.Time           `json:"created"`
		Author       string              `json:"author"`
		Config       oci.ContainerConfig `json:"config"`
		RootFS       rootFS              `json:"rootfs"`
		History      []oci.HistoryEntry  `json:"history"`
	}

	cfg := slimConfig{
		Architecture: img.Config.Architecture,
		OS:           img.Config.OS,
		Created:      time.Now(),
		Author:       "LayerScope Slim Engine",
		Config:       img.Config.Config,
		RootFS: rootFS{
			Type:    "layers",
			DiffIDs: []string{diffID},
		},
		History: []oci.HistoryEntry{
			{
				Created:    time.Now(),
				CreatedBy:  "LayerScope slim: squashed into single clean layer without intermediate cache or whiteouts",
				EmptyLayer: false,
			},
		},
	}
	if cfg.Architecture == "" {
		cfg.Architecture = "amd64"
	}
	if cfg.OS == "" {
		cfg.OS = "linux"
	}

	cfgBytes, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal config: %w", err)
	}

	cfgHasher := sha256.New()
	cfgHasher.Write(cfgBytes)
	cfgFilename := hex.EncodeToString(cfgHasher.Sum(nil)) + ".json"

	// 3. Generate Docker manifest.json
	type dockerManifestItem struct {
		Config   string   `json:"Config"`
		RepoTags []string `json:"RepoTags"`
		Layers   []string `json:"Layers"`
	}

	manifest := []dockerManifestItem{
		{
			Config:   cfgFilename,
			RepoTags: []string{repoTag},
			Layers:   []string{layerTarPath},
		},
	}
	manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal manifest: %w", err)
	}

	// 4. Write manifest.json to main tar
	_ = tw.WriteHeader(&tar.Header{
		Name:     "manifest.json",
		Mode:     0644,
		Size:     int64(len(manifestBytes)),
		Typeflag: tar.TypeReg,
		ModTime:  time.Now(),
	})
	_, _ = tw.Write(manifestBytes)

	// 5. Write config.json to main tar
	_ = tw.WriteHeader(&tar.Header{
		Name:     cfgFilename,
		Mode:     0644,
		Size:     int64(len(cfgBytes)),
		Typeflag: tar.TypeReg,
		ModTime:  time.Now(),
	})
	_, _ = tw.Write(cfgBytes)

	// 6. Write layer.tar to main tar
	_ = tw.WriteHeader(&tar.Header{
		Name:     layerTarPath,
		Mode:     0644,
		Size:     layerStat.Size(),
		Typeflag: tar.TypeReg,
		ModTime:  time.Now(),
	})

	if _, err := tempLayerFile.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("rewind temp layer: %w", err)
	}
	for {
		n, rErr := tempLayerFile.Read(layerBuf)
		if n > 0 {
			if _, wErr := tw.Write(layerBuf[:n]); wErr != nil {
				return nil, fmt.Errorf("write layer tar into container archive: %w", wErr)
			}
		}
		if rErr != nil {
			break
		}
	}

	slimSize := layerStat.Size() + int64(len(manifestBytes)) + int64(len(cfgBytes))
	savedBytes := img.TotalSizeBytes - slimSize
	savedPct := 0.0
	if img.TotalSizeBytes > 0 && savedBytes > 0 {
		savedPct = (float64(savedBytes) / float64(img.TotalSizeBytes)) * 100.0
	}

	return &Result{
		OriginalSize:    img.TotalSizeBytes,
		SlimSize:        slimSize,
		SavedBytes:      savedBytes,
		SavedPercentage: savedPct,
		OutputFile:      outPath,
		LayerCount:      1,
	}, nil
}

func isCachePath(p string) bool {
	cleaned := path.Clean(p)
	return strings.HasPrefix(cleaned, "/var/cache") ||
		strings.HasPrefix(cleaned, "/tmp") ||
		strings.HasPrefix(cleaned, "/root/.cache") ||
		strings.HasPrefix(cleaned, "/root/.npm") ||
		strings.HasPrefix(cleaned, "/var/lib/apt/lists")
}
