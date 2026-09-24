package sbom

// PackageType identifies the ecosystem where the dependency originated.
type PackageType string

const (
	TypeAlpine PackageType = "os:alpine"
	TypeDebian PackageType = "os:debian"
	TypeRPM    PackageType = "os:rpm"
	TypeNPM    PackageType = "npm"
	TypePyPI   PackageType = "pypi"
	TypeGo     PackageType = "golang"
	TypeCargo  PackageType = "cargo"
)

// Package represents an individual software component discovered in the container.
type Package struct {
	Name          string      `json:"name"`
	Version       string      `json:"version"`
	Type          PackageType `json:"type"`
	License       string      `json:"license,omitempty"`
	Description   string      `json:"description,omitempty"`
	PURL          string      `json:"purl,omitempty"`
	Size          int64       `json:"size,omitempty"`
	InstalledPath string      `json:"installedPath,omitempty"`
}

// SBOMReport contains the complete package inventory and counts.
type SBOMReport struct {
	ImageName     string     `json:"imageName"`
	TotalPackages int        `json:"totalPackages"`
	OSPackages    int        `json:"osPackages"`
	AppPackages   int        `json:"appPackages"`
	Packages      []*Package `json:"packages"`
}
