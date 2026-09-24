package oci

import (
	"fmt"
	"io"
	"net"
	"os"
	"runtime"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/daemon"
)

// CheckDockerDaemon checks whether a local Docker daemon socket or named pipe is accessible.
func CheckDockerDaemon() bool {
	if host := os.Getenv("DOCKER_HOST"); host != "" {
		return true
	}

	if runtime.GOOS == "windows" {
		conn, err := net.DialTimeout("tcp", "127.0.0.1:2375", 500*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return true
		}
		// Windows named pipe existence check
		f, err := os.OpenFile(`\\.\pipe\docker_engine`, os.O_RDWR, 0)
		if err == nil {
			_ = f.Close()
			return true
		}
		return false
	}

	// UNIX socket check
	if _, err := os.Stat("/var/run/docker.sock"); err == nil {
		conn, err := net.DialTimeout("unix", "/var/run/docker.sock", 500*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return true
		}
	}

	return false
}

// FetchDaemonImage loads and analyzes an image from the local Docker daemon.
func FetchDaemonImage(imageRef string, hook FileInspectHook) (*ImageAnalysis, error) {
	tag, err := name.NewTag(imageRef, name.WeakValidation)
	if err != nil {
		return nil, fmt.Errorf("parse daemon image tag %q: %w", imageRef, err)
	}

	img, err := daemon.Image(tag)
	if err != nil {
		return nil, fmt.Errorf("connect to docker daemon for %q (ensure Docker Desktop/daemon is running, or use --remote / --demo): %w", imageRef, err)
	}

	return analyzeV1Image(img, imageRef, SourceDockerDaemon, hook)
}

func analyzeV1Image(img v1.Image, originalRef string, source SourceType, hook FileInspectHook) (*ImageAnalysis, error) {
	cfgFile, err := img.ConfigFile()
	if err != nil {
		return nil, fmt.Errorf("fetch daemon image config: %w", err)
	}

	digest, err := img.Digest()
	if err != nil {
		return nil, fmt.Errorf("get daemon image digest: %w", err)
	}

	var imgConfig ImageConfig
	imgConfig.Architecture = cfgFile.Architecture
	imgConfig.OS = cfgFile.OS
	imgConfig.Created = cfgFile.Created.Time
	imgConfig.Author = cfgFile.Author
	imgConfig.Config.User = cfgFile.Config.User
	imgConfig.Config.Env = cfgFile.Config.Env
	imgConfig.Config.Entrypoint = cfgFile.Config.Entrypoint
	imgConfig.Config.Cmd = cfgFile.Config.Cmd
	imgConfig.Config.WorkingDir = cfgFile.Config.WorkingDir
	imgConfig.Config.Labels = cfgFile.Config.Labels
	imgConfig.Config.StopSignal = cfgFile.Config.StopSignal

	for _, d := range cfgFile.RootFS.DiffIDs {
		imgConfig.RootFS.DiffIDs = append(imgConfig.RootFS.DiffIDs, d.String())
	}

	for _, h := range cfgFile.History {
		imgConfig.History = append(imgConfig.History, HistoryEntry{
			Created:    h.Created.Time,
			CreatedBy:  h.CreatedBy,
			EmptyLayer: h.EmptyLayer,
			Comment:    h.Comment,
			Author:     h.Author,
		})
	}

	v1Layers, err := img.Layers()
	if err != nil {
		return nil, fmt.Errorf("fetch daemon layers: %w", err)
	}

	var layers []*Layer
	var totalSize int64
	var totalFiles int
	historyIndex := 0

	for i, v1Layer := range v1Layers {
		lDigest, _ := v1Layer.Digest()
		lDiffID, _ := v1Layer.DiffID()

		uncompressedStream, err := v1Layer.Uncompressed()
		if err != nil {
			return nil, fmt.Errorf("open layer %d uncompressed stream: %w", i, err)
		}

		files, uncompressedSize, err := UnpackLayerTar(i, uncompressedStream, hook)
		_ = uncompressedStream.Close()
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("unpack layer %d: %w", i, err)
		}

		cmd := ""
		for historyIndex < len(imgConfig.History) {
			h := imgConfig.History[historyIndex]
			historyIndex++
			if !h.EmptyLayer {
				cmd = h.CreatedBy
				break
			}
		}

		layer := &Layer{
			Index:       i,
			Digest:      lDigest.String(),
			DiffID:      lDiffID.String(),
			Size:        uncompressedSize,
			Command:     cmd,
			CreatedBy:   cmd,
			Files:       files,
			FileCount:   len(files),
			WastedBytes: 0,
		}

		layers = append(layers, layer)
		totalSize += uncompressedSize
		totalFiles += len(files)
	}

	return &ImageAnalysis{
		Reference: ImageReference{
			Original:     originalRef,
			Repository:   originalRef,
			Tag:          "latest",
			Digest:       digest.String(),
			Architecture: cfgFile.Architecture,
			OS:           cfgFile.OS,
			Source:       source,
		},
		Config:          imgConfig,
		Layers:          layers,
		TotalSizeBytes:  totalSize,
		TotalFilesCount: totalFiles,
		AnalyzedAt:      time.Now(),
	}, nil
}
