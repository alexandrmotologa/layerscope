package oci

import (
	"os"
	"time"
)

// SourceType represents the mechanism used to fetch the image.
type SourceType string

const (
	SourceDockerDaemon   SourceType = "docker-daemon"
	SourcePodman         SourceType = "podman"
	SourceRemoteRegistry SourceType = "remote-registry"
	SourceTarball        SourceType = "tarball"
	SourceSample         SourceType = "sample"
)

// ImageReference contains normalized identifiers for an analyzed image.
type ImageReference struct {
	Original     string     `json:"original"`
	Registry     string     `json:"registry"`
	Repository   string     `json:"repository"`
	Tag          string     `json:"tag"`
	Digest       string     `json:"digest"`
	Architecture string     `json:"architecture"`
	OS           string     `json:"os"`
	Source       SourceType `json:"source"`
}

// ContainerConfig captures runtime configuration stored in the image JSON.
type ContainerConfig struct {
	User         string              `json:"user,omitempty"`
	ExposedPorts map[string]struct{} `json:"exposedPorts,omitempty"`
	Env          []string            `json:"env,omitempty"`
	Entrypoint   []string            `json:"entrypoint,omitempty"`
	Cmd          []string            `json:"cmd,omitempty"`
	Volumes      map[string]struct{} `json:"volumes,omitempty"`
	WorkingDir   string              `json:"workingDir,omitempty"`
	Labels       map[string]string   `json:"labels,omitempty"`
	StopSignal   string              `json:"stopSignal,omitempty"`
}

// HistoryEntry maps to an individual line in the image history.
type HistoryEntry struct {
	Created    time.Time `json:"created"`
	CreatedBy  string    `json:"createdBy"`
	EmptyLayer bool      `json:"emptyLayer"`
	Comment    string    `json:"comment,omitempty"`
	Author     string    `json:"author,omitempty"`
}

// ImageConfig represents the parsed OCI/Docker container configuration blob.
type ImageConfig struct {
	Architecture string          `json:"architecture"`
	OS           string          `json:"os"`
	Created      time.Time       `json:"created"`
	Author       string          `json:"author,omitempty"`
	Config       ContainerConfig `json:"config"`
	RootFS       struct {
		Type    string   `json:"type"`
		DiffIDs []string `json:"diffIds"`
	} `json:"rootfs"`
	History []HistoryEntry `json:"history"`
}

// LayerFile describes a single file or directory inside an image layer tar archive.
type LayerFile struct {
	Path           string      `json:"path"`
	Size           int64       `json:"size"`
	Mode           os.FileMode `json:"mode"`
	ModTime        time.Time   `json:"modTime"`
	IsDir          bool        `json:"isDir"`
	IsSymlink      bool        `json:"isSymlink"`
	LinkTarget     string      `json:"linkTarget,omitempty"`
	IsWhiteout     bool        `json:"isWhiteout"`
	IsOpaque       bool        `json:"isOpaque"`
	WhiteoutTarget string      `json:"whiteoutTarget,omitempty"`
	Digest         string      `json:"digest,omitempty"`
	Data           []byte      `json:"-"`
}

// Layer represents one individual filesystem layer in the container image.
type Layer struct {
	Index        int          `json:"index"`
	Digest       string       `json:"digest"`
	DiffID       string       `json:"diffId"`
	Size         int64        `json:"size"`
	Command      string       `json:"command"`
	CreatedBy    string       `json:"createdBy"`
	EmptyLayer   bool         `json:"emptyLayer"`
	Files        []*LayerFile `json:"files"`
	FileCount    int          `json:"fileCount"`
	WastedBytes  int64        `json:"wastedBytes"`
}

// ImageAnalysis captures complete extracted information about an image.
type ImageAnalysis struct {
	Reference       ImageReference `json:"reference"`
	Config          ImageConfig    `json:"config"`
	Layers          []*Layer       `json:"layers"`
	TotalSizeBytes  int64          `json:"totalSizeBytes"`
	TotalFilesCount int            `json:"totalFilesCount"`
	AnalyzedAt      time.Time      `json:"analyzedAt"`
}
