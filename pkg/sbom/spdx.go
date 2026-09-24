package sbom

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type spdxDocument struct {
	SPDXVersion       string         `json:"spdxVersion"`
	DataLicense       string         `json:"dataLicense"`
	SPDXID            string         `json:"SPDXID"`
	Name              string         `json:"name"`
	DocumentNamespace string         `json:"documentNamespace"`
	CreationInfo      spdxCreation   `json:"creationInfo"`
	Packages          []spdxPackage  `json:"packages"`
}

type spdxCreation struct {
	Created  string   `json:"created"`
	Creators []string `json:"creators"`
}

type spdxPackage struct {
	SPDXID           string            `json:"SPDXID"`
	Name             string            `json:"name"`
	VersionInfo      string            `json:"versionInfo"`
	PackageFileName  string            `json:"packageFileName,omitempty"`
	DownloadLocation string            `json:"downloadLocation"`
	LicenseConcluded string            `json:"licenseConcluded"`
	LicenseDeclared  string            `json:"licenseDeclared"`
	ExternalRefs     []spdxExternalRef `json:"externalRefs,omitempty"`
}

type spdxExternalRef struct {
	ReferenceCategory string `json:"referenceCategory"`
	ReferenceType     string `json:"referenceType"`
	ReferenceLocator  string `json:"referenceLocator"`
}

// ExportSPDXJSON formats the package report as standard SPDX v2.3 JSON.
func ExportSPDXJSON(report *SBOMReport) ([]byte, error) {
	doc := spdxDocument{
		SPDXVersion:       "SPDX-2.3",
		DataLicense:       "CC0-1.0",
		SPDXID:            "SPDXRef-DOCUMENT",
		Name:              report.ImageName,
		DocumentNamespace: fmt.Sprintf("https://spdx.org/spdxdocs/layerscope-%d", time.Now().UnixNano()),
		CreationInfo: spdxCreation{
			Created:  time.Now().UTC().Format(time.RFC3339),
			Creators: []string{"Tool: LayerScope-1.0.0"},
		},
	}

	for i, pkg := range report.Packages {
		cleanName := strings.ReplaceAll(pkg.Name, "/", "-")
		spdxID := fmt.Sprintf("SPDXRef-Package-%d-%s", i, cleanName)

		sp := spdxPackage{
			SPDXID:           spdxID,
			Name:             pkg.Name,
			VersionInfo:      pkg.Version,
			DownloadLocation: "NOASSERTION",
			LicenseConcluded: "NOASSERTION",
			LicenseDeclared:  "NOASSERTION",
		}

		if pkg.License != "" {
			sp.LicenseDeclared = pkg.License
		}

		if pkg.PURL != "" {
			sp.ExternalRefs = []spdxExternalRef{
				{
					ReferenceCategory: "PACKAGE-MANAGER",
					ReferenceType:     "purl",
					ReferenceLocator:  pkg.PURL,
				},
			}
		}

		doc.Packages = append(doc.Packages, sp)
	}

	return json.MarshalIndent(doc, "", "  ")
}
