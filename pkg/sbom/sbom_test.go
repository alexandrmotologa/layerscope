package sbom

import (
	"strings"
	"testing"

	"github.com/alexandrmotologa/layerscope/pkg/oci"
	"github.com/alexandrmotologa/layerscope/pkg/vfs"
)

func TestParseAlpineInstalled(t *testing.T) {
	apkData := `P:musl
V:1.2.5-r0
T:the musl c-library
L:MIT
I:624512

P:zlib
V:1.3.1-r0
T:Compression library
L:Zlib
I:112400
`
	pkgs, err := ParseAlpineInstalled(strings.NewReader(apkData))
	if err != nil {
		t.Fatalf("ParseAlpineInstalled failed: %v", err)
	}

	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}

	if pkgs[0].Name != "musl" || pkgs[0].Version != "1.2.5-r0" || pkgs[0].Type != TypeAlpine {
		t.Errorf("unexpected first package: %+v", pkgs[0])
	}
	if pkgs[0].PURL != "pkg:apk/alpine/musl@1.2.5-r0" {
		t.Errorf("expected PURL pkg:apk/alpine/musl@1.2.5-r0, got %s", pkgs[0].PURL)
	}
}

func TestParseNPMLockfile(t *testing.T) {
	lockJSON := `{
  "name": "myapp",
  "version": "1.0.0",
  "packages": {
    "node_modules/express": {
      "version": "4.19.2"
    },
    "node_modules/lodash": {
      "version": "4.17.21"
    }
  }
}`
	pkgs, err := ParseNPMLockfile(strings.NewReader(lockJSON), "/app/package-lock.json")
	if err != nil {
		t.Fatalf("ParseNPMLockfile failed: %v", err)
	}

	if len(pkgs) != 2 {
		t.Fatalf("expected 2 packages, got %d", len(pkgs))
	}
}

func TestExportCycloneDXAndSPDX(t *testing.T) {
	report := &SBOMReport{
		ImageName:     "alpine:3.20",
		TotalPackages: 2,
		OSPackages:    2,
		Packages: []*Package{
			{Name: "musl", Version: "1.2.5-r0", Type: TypeAlpine, PURL: "pkg:apk/alpine/musl@1.2.5-r0", License: "MIT"},
			{Name: "zlib", Version: "1.3.1-r0", Type: TypeAlpine, PURL: "pkg:apk/alpine/zlib@1.3.1-r0", License: "Zlib"},
		},
	}

	cdxBytes, err := ExportCycloneDXJSON(report)
	if err != nil {
		t.Fatalf("ExportCycloneDXJSON failed: %v", err)
	}
	if !strings.Contains(string(cdxBytes), "CycloneDX") || !strings.Contains(string(cdxBytes), "musl") {
		t.Errorf("CycloneDX JSON missing expected content: %s", string(cdxBytes))
	}

	spdxBytes, err := ExportSPDXJSON(report)
	if err != nil {
		t.Fatalf("ExportSPDXJSON failed: %v", err)
	}
	if !strings.Contains(string(spdxBytes), "SPDX-2.3") || !strings.Contains(string(spdxBytes), "musl") {
		t.Errorf("SPDX JSON missing expected content: %s", string(spdxBytes))
	}
}

func TestSBOMExtractor_WithSampleImage(t *testing.T) {
	extractor := NewExtractor()
	hook := extractor.Hook()

	sampleImg, err := oci.GenerateSampleImage(hook)
	if err != nil {
		t.Fatalf("GenerateSampleImage failed: %v", err)
	}

	snapshots := vfs.BuildLayerSnapshots(sampleImg)
	finalTree := snapshots[len(snapshots)-1].Tree

	report, err := extractor.FinalizeReport(sampleImg.Reference.Original, finalTree)
	if err != nil {
		t.Fatalf("FinalizeReport failed: %v", err)
	}

	if report.TotalPackages == 0 {
		t.Errorf("expected packages from apk database and package-lock.json in sample image, got 0")
	}

	if report.OSPackages == 0 {
		t.Errorf("expected OS packages from Alpine /lib/apk/db/installed, got 0")
	}

	if report.AppPackages == 0 {
		t.Errorf("expected NPM packages from /app/package-lock.json, got 0")
	}
}
