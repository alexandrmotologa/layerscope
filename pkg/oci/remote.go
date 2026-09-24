package oci

import (
	"fmt"
	"io"
	"time"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	v1 "github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
)

// RemoteFetchOptions configures fetching an image from a remote registry.
type RemoteFetchOptions struct {
	Architecture string
	OS           string
	Insecure     bool
	Keychain     authn.Keychain
}

// FetchRemoteImage streams and analyzes an image directly from a remote OCI registry.
func FetchRemoteImage(imageRef string, opts RemoteFetchOptions, hook FileInspectHook) (*ImageAnalysis, error) {
	var nameOpts []name.Option
	if opts.Insecure {
		nameOpts = append(nameOpts, name.Insecure)
	}

	ref, err := name.ParseReference(imageRef, nameOpts...)
	if err != nil {
		return nil, fmt.Errorf("parse image reference %q: %w", imageRef, err)
	}

	keychain := opts.Keychain
	if keychain == nil {
		keychain = authn.DefaultKeychain
	}

	var remoteOpts []remote.Option
	remoteOpts = append(remoteOpts, remote.WithAuthFromKeychain(keychain))

	targetArch := opts.Architecture
	if targetArch == "" {
		targetArch = "amd64"
	}
	targetOS := opts.OS
	if targetOS == "" {
		targetOS = "linux"
	}

	platform := v1.Platform{
		Architecture: targetArch,
		OS:           targetOS,
	}
	remoteOpts = append(remoteOpts, remote.WithPlatform(platform))

	img, err := remote.Image(ref, remoteOpts...)
	if err != nil {
		return nil, fmt.Errorf("fetch remote image %s: %w", imageRef, err)
	}

	cfgFile, err := img.ConfigFile()
	if err != nil {
		return nil, fmt.Errorf("fetch image config file: %w", err)
	}

	digest, err := img.Digest()
	if err != nil {
		return nil, fmt.Errorf("get image digest: %w", err)
	}

	// Map config to our internal struct
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
		return nil, fmt.Errorf("fetch image layers: %w", err)
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
			Original:     imageRef,
			Registry:     ref.Context().RegistryStr(),
			Repository:   ref.Context().RepositoryStr(),
			Tag:          ref.Identifier(),
			Digest:       digest.String(),
			Architecture: cfgFile.Architecture,
			OS:           cfgFile.OS,
			Source:       SourceRemoteRegistry,
		},
		Config:          imgConfig,
		Layers:          layers,
		TotalSizeBytes:  totalSize,
		TotalFilesCount: totalFiles,
		AnalyzedAt:      time.Now(),
	}, nil
}

// ListRemoteArchitectures queries an image index or manifest list and returns available architectures.
func ListRemoteArchitectures(imageRef string, opts RemoteFetchOptions) ([]string, error) {
	var nameOpts []name.Option
	if opts.Insecure {
		nameOpts = append(nameOpts, name.Insecure)
	}

	ref, err := name.ParseReference(imageRef, nameOpts...)
	if err != nil {
		return nil, fmt.Errorf("parse image reference %q: %w", imageRef, err)
	}

	keychain := opts.Keychain
	if keychain == nil {
		keychain = authn.DefaultKeychain
	}

	desc, err := remote.Get(ref, remote.WithAuthFromKeychain(keychain))
	if err != nil {
		return nil, fmt.Errorf("get remote descriptor: %w", err)
	}

	var archs []string
	if desc.MediaType.IsIndex() {
		idx, err := desc.ImageIndex()
		if err != nil {
			return nil, fmt.Errorf("parse image index: %w", err)
		}
		manifest, err := idx.IndexManifest()
		if err != nil {
			return nil, fmt.Errorf("parse index manifest: %w", err)
		}
		for _, m := range manifest.Manifests {
			if m.Platform != nil {
				archs = append(archs, fmt.Sprintf("%s/%s", m.Platform.OS, m.Platform.Architecture))
			}
		}
	} else {
		// Single architecture image
		img, err := desc.Image()
		if err == nil {
			cfg, err := img.ConfigFile()
			if err == nil {
				archs = append(archs, fmt.Sprintf("%s/%s", cfg.OS, cfg.Architecture))
			}
		}
	}

	if len(archs) == 0 {
		archs = []string{"linux/amd64"}
	}

	return archs, nil
}
