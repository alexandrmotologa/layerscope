package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexandrmotologa/layerscope/pkg/advisor"
	"github.com/alexandrmotologa/layerscope/pkg/oci"
	"github.com/alexandrmotologa/layerscope/pkg/sbom"
	"github.com/alexandrmotologa/layerscope/pkg/security"
	"github.com/alexandrmotologa/layerscope/pkg/vfs"
)

func TestServerEndpoints(t *testing.T) {
	sampleImg, err := oci.GenerateSampleImage(nil)
	if err != nil {
		t.Fatalf("GenerateSampleImage failed: %v", err)
	}

	snapshots := vfs.BuildLayerSnapshots(sampleImg)
	wasteSummary := vfs.CalculateWastedSpace(snapshots, nil)
	finalTree := snapshots[len(snapshots)-1].Tree
	runsAsRoot, suid := security.AuditPrivileges(&sampleImg.Config, finalTree)
	secReport := &security.SecurityAuditReport{
		TotalFindings:     1,
		RunsAsRoot:        runsAsRoot,
		SuidBinariesCount: len(suid),
	}
	sbomReport := &sbom.SBOMReport{
		ImageName:     "sample:demo",
		TotalPackages: 2,
		Packages: []*sbom.Package{
			{Name: "musl", Version: "1.2.5-r0", Type: sbom.TypeAlpine},
		},
	}
	advReport := advisor.AnalyzeImageHeuristics(sampleImg, wasteSummary, secReport, finalTree)

	srv := NewServer(sampleImg, snapshots, wasteSummary, secReport, sbomReport, advReport, nil)

	// 1. Health check
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()
	srv.Router().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 on /api/health, got %d", w.Code)
	}

	// 2. Image metadata
	req = httptest.NewRequest(http.MethodGet, "/api/image", nil)
	w = httptest.NewRecorder()
	srv.Router().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 on /api/image, got %d", w.Code)
	}
	var imgResp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &imgResp); err != nil {
		t.Fatalf("parse /api/image response: %v", err)
	}
	if imgResp["efficiencyScore"] == nil {
		t.Errorf("expected efficiencyScore in /api/image response")
	}

	// 3. Layer tree
	req = httptest.NewRequest(http.MethodGet, "/api/layers/0/tree", nil)
	w = httptest.NewRecorder()
	srv.Router().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 on /api/layers/0/tree, got %d", w.Code)
	}

	// 4. SBOM CycloneDX Export
	req = httptest.NewRequest(http.MethodGet, "/api/sbom/cyclonedx", nil)
	w = httptest.NewRecorder()
	srv.Router().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 on /api/sbom/cyclonedx, got %d", w.Code)
	}

	// 5. HTML Studio Fallback
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	w = httptest.NewRecorder()
	srv.Router().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 on /, got %d", w.Code)
	}
	if w.Header().Get("Content-Type") != "text/html; charset=utf-8" {
		t.Errorf("expected text/html content type, got %s", w.Header().Get("Content-Type"))
	}

	// 6. Presets list
	req = httptest.NewRequest(http.MethodGet, "/api/presets", nil)
	w = httptest.NewRecorder()
	srv.Router().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 on /api/presets, got %d", w.Code)
	}

	// 7. File Inspection Drawer Preview
	req = httptest.NewRequest(http.MethodGet, "/api/layers/0/file?path=/etc/os-release", nil)
	w = httptest.NewRecorder()
	srv.Router().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 on /api/layers/0/file, got %d", w.Code)
	}
	var fileResp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &fileResp); err != nil {
		t.Fatalf("parse file inspect response: %v", err)
	}
	if fileResp["path"] != "/etc/os-release" {
		t.Errorf("expected /etc/os-release, got %v", fileResp["path"])
	}
	if fileResp["isText"] != true {
		t.Errorf("expected isText=true")
	}

	// 8. Dockerfile Optimizer
	req = httptest.NewRequest(http.MethodGet, "/api/advisor/dockerfile", nil)
	w = httptest.NewRecorder()
	srv.Router().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 on /api/advisor/dockerfile, got %d", w.Code)
	}
	var optResp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &optResp); err != nil {
		t.Fatalf("parse dockerfile opt response: %v", err)
	}
	if optResp["optimizedDockerfile"] == nil {
		t.Errorf("expected optimizedDockerfile field")
	}
}
