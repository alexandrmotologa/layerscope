package sbom

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

type cycloneDXBOM struct {
	BOMFormat    string              `json:"bomFormat"`
	SpecVersion  string              `json:"specVersion"`
	SerialNumber string              `json:"serialNumber"`
	Version      int                 `json:"version"`
	Metadata     cycloneDXMetadata   `json:"metadata"`
	Components   []cycloneDXComp     `json:"components"`
}

type cycloneDXMetadata struct {
	Timestamp string             `json:"timestamp"`
	Tools     []cycloneDXTool    `json:"tools"`
	Component cycloneDXComponent `json:"component"`
}

type cycloneDXTool struct {
	Vendor  string `json:"vendor"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

type cycloneDXComponent struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

type cycloneDXComp struct {
	Type        string               `json:"type"`
	Name        string               `json:"name"`
	Version     string               `json:"version"`
	PURL        string               `json:"purl,omitempty"`
	Description string               `json:"description,omitempty"`
	Licenses    []cycloneDXLicEntry  `json:"licenses,omitempty"`
}

type cycloneDXLicEntry struct {
	License cycloneDXLicense `json:"license"`
}

type cycloneDXLicense struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// ExportCycloneDXJSON formats the package report as standard CycloneDX v1.5 JSON.
func ExportCycloneDXJSON(report *SBOMReport) ([]byte, error) {
	uuidBytes := make([]byte, 16)
	_, _ = rand.Read(uuidBytes)
	serialNumber := fmt.Sprintf("urn:uuid:%s-%s-%s-%s-%s",
		hex.EncodeToString(uuidBytes[0:4]),
		hex.EncodeToString(uuidBytes[4:6]),
		hex.EncodeToString(uuidBytes[6:8]),
		hex.EncodeToString(uuidBytes[8:10]),
		hex.EncodeToString(uuidBytes[10:16]),
	)

	bom := cycloneDXBOM{
		BOMFormat:    "CycloneDX",
		SpecVersion:  "1.5",
		SerialNumber: serialNumber,
		Version:      1,
		Metadata: cycloneDXMetadata{
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Tools: []cycloneDXTool{
				{
					Vendor:  "LayerScope",
					Name:    "layerscope",
					Version: "1.0.0",
				},
			},
			Component: cycloneDXComponent{
				Type: "container",
				Name: report.ImageName,
			},
		},
	}

	for _, pkg := range report.Packages {
		compType := "library"
		if pkg.Type == TypeAlpine || pkg.Type == TypeDebian || pkg.Type == TypeRPM {
			compType = "operating-system"
		}

		comp := cycloneDXComp{
			Type:        compType,
			Name:        pkg.Name,
			Version:     pkg.Version,
			PURL:        pkg.PURL,
			Description: pkg.Description,
		}

		if pkg.License != "" {
			comp.Licenses = []cycloneDXLicEntry{
				{
					License: cycloneDXLicense{
						Name: pkg.License,
					},
				},
			}
		}

		bom.Components = append(bom.Components, comp)
	}

	return json.MarshalIndent(bom, "", "  ")
}
