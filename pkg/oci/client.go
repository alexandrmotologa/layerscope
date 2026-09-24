package oci

import (
	"os"
	"strings"
)

// IngestionOptions specifies parameters for acquiring and inspecting an image.
type IngestionOptions struct {
	ForceRemote  bool
	ForceDaemon  bool
	ForceDemo    bool
	Architecture string
	OS           string
	Insecure     bool
}

// LoadImage resolves the source automatically and returns the parsed ImageAnalysis.
func LoadImage(target string, opts IngestionOptions, hook FileInspectHook) (*ImageAnalysis, error) {
	if opts.ForceDemo || target == "demo" || target == "layerscope:demo" {
		return GenerateSampleImage(hook)
	}

	// Check if target is a local file tarball
	if fileInfo, err := os.Stat(target); err == nil && !fileInfo.IsDir() {
		if strings.HasSuffix(strings.ToLower(target), ".tar") ||
			strings.HasSuffix(strings.ToLower(target), ".tar.gz") ||
			strings.HasSuffix(strings.ToLower(target), ".tgz") {
			return LoadImageFromTarball(target, hook)
		}
	}

	// If explicit remote requested
	if opts.ForceRemote {
		return FetchRemoteImage(target, RemoteFetchOptions{
			Architecture: opts.Architecture,
			OS:           opts.OS,
			Insecure:     opts.Insecure,
		}, hook)
	}

	// If explicit daemon requested
	if opts.ForceDaemon {
		return FetchDaemonImage(target, hook)
	}

	// Smart detection:
	// If Docker daemon is responsive, try daemon first
	if CheckDockerDaemon() {
		analysis, err := FetchDaemonImage(target, hook)
		if err == nil {
			return analysis, nil
		}
		// If daemon fails (e.g. image not found locally), fall through to remote registry
	}

	// Fallback to remote registry
	return FetchRemoteImage(target, RemoteFetchOptions{
		Architecture: opts.Architecture,
		OS:           opts.OS,
		Insecure:     opts.Insecure,
	}, hook)
}
