package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/alexandrmotologa/layerscope/pkg/advisor"
	"github.com/alexandrmotologa/layerscope/pkg/oci"
	"github.com/alexandrmotologa/layerscope/pkg/sbom"
	"github.com/alexandrmotologa/layerscope/pkg/security"
	"github.com/alexandrmotologa/layerscope/pkg/vfs"
)

// Server coordinates HTTP endpoints for the visual workbench studio.
type Server struct {
	mu          sync.RWMutex
	router      *chi.Mux
	img         *oci.ImageAnalysis
	snapshots   []*vfs.LayerSnapshot
	waste       *vfs.WastedSpaceSummary
	secReport   *security.SecurityAuditReport
	sbomReport  *sbom.SBOMReport
	advReport   *advisor.AdvisorReport
	diffReport  *vfs.ImageDiffReport
	staticFS    http.FileSystem
}

// NewServer initializes the REST router and registers API endpoints.
func NewServer(
	img *oci.ImageAnalysis,
	snapshots []*vfs.LayerSnapshot,
	waste *vfs.WastedSpaceSummary,
	secReport *security.SecurityAuditReport,
	sbomReport *sbom.SBOMReport,
	advReport *advisor.AdvisorReport,
	staticFS http.FileSystem,
) *Server {
	s := &Server{
		img:        img,
		snapshots:  snapshots,
		waste:      waste,
		secReport:  secReport,
		sbomReport: sbomReport,
		advReport:  advReport,
		staticFS:   staticFS,
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	s.registerRoutes(r)
	s.router = r

	return s
}

// Router returns the underlying Chi router.
func (s *Server) Router() *chi.Mux {
	return s.router
}

func (s *Server) registerRoutes(r *chi.Mux) {
	r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "version": "1.0.0"})
	})

	r.Get("/api/image", s.handleGetImage)
	r.Get("/api/presets", s.handleGetPresets)
	r.Get("/api/layers/{index}/tree", s.handleGetLayerTree)
	r.Get("/api/layers/{index}/file", s.handleGetLayerFile)
	r.Get("/api/waste", s.handleGetWaste)
	r.Get("/api/security", s.handleGetSecurity)
	r.Get("/api/sbom", s.handleGetSBOM)
	r.Get("/api/sbom/cyclonedx", s.handleExportCycloneDX)
	r.Get("/api/sbom/spdx", s.handleExportSPDX)
	r.Get("/api/advisor", s.handleGetAdvisor)
	r.Get("/api/advisor/dockerfile", s.handleGetAdvisorDockerfile)
	r.Get("/api/diff", s.handleGetDiff)
	r.Post("/api/analyze", s.handlePostAnalyze)
	r.Post("/api/diff", s.handlePostDiff)

	// Static UI handler
	r.NotFound(s.handleStaticOrFallback)
}

func (s *Server) handleGetImage(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.img == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "no image loaded"})
		return
	}

	type layerMeta struct {
		Index       int    `json:"index"`
		Digest      string `json:"digest"`
		Command     string `json:"command"`
		Size        int64  `json:"size"`
		WastedBytes int64  `json:"wastedBytes"`
		FileCount   int    `json:"fileCount"`
	}

	var layersMeta []layerMeta
	for _, snap := range s.snapshots {
		layersMeta = append(layersMeta, layerMeta{
			Index:       snap.LayerIndex,
			Digest:      snap.Digest,
			Command:     snap.Command,
			Size:        snap.Size,
			WastedBytes: snap.WastedBytes,
			FileCount:   len(snap.DeltaFiles),
		})
	}

	resp := map[string]interface{}{
		"reference":       s.img.Reference,
		"config":          s.img.Config,
		"totalSizeBytes":  s.img.TotalSizeBytes,
		"totalFilesCount": s.img.TotalFilesCount,
		"efficiencyScore": s.advReport.EfficiencyScore,
		"grade":           s.advReport.Grade,
		"layers":          layersMeta,
		"analyzedAt":      s.img.AnalyzedAt,
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleGetLayerTree(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	idxStr := chi.URLParam(r, "index")
	idx, err := strconv.Atoi(idxStr)
	if err != nil || idx < 0 || idx >= len(s.snapshots) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid layer index"})
		return
	}

	snap := s.snapshots[idx]
	wastedOnly := r.URL.Query().Get("wastedOnly") == "true"
	prefix := r.URL.Query().Get("prefix")

	type nodeDTO struct {
		Path        string        `json:"path"`
		Name        string        `json:"name"`
		Size        int64         `json:"size"`
		IsDir       bool          `json:"isDir"`
		ChangeType  vfs.ChangeType `json:"changeType"`
		IsWasted    bool          `json:"isWasted"`
		WastedBytes int64         `json:"wastedBytes"`
		WasteReason string        `json:"wasteReason,omitempty"`
		Children    []string      `json:"children,omitempty"`
	}

	nodes := make(map[string]nodeDTO)
	for p, n := range snap.Tree.Index {
		if prefix != "" && !strings.HasPrefix(p, prefix) {
			continue
		}
		if wastedOnly && !n.IsWasted && !n.IsDir {
			continue
		}

		var childPaths []string
		if n.IsDir {
			for _, child := range n.SortedChildren() {
				childPaths = append(childPaths, child.Path)
			}
		}

		nodes[p] = nodeDTO{
			Path:        n.Path,
			Name:        n.Name,
			Size:        n.Size,
			IsDir:       n.IsDir,
			ChangeType:  n.ChangeType,
			IsWasted:    n.IsWasted,
			WastedBytes: n.WastedBytes,
			WasteReason: n.WasteReason,
			Children:    childPaths,
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"layerIndex": idx,
		"command":    snap.Command,
		"totalNodes": len(nodes),
		"nodes":      nodes,
		"deltaFiles": snap.DeltaFiles,
	})
}

func (s *Server) handleGetPresets(w http.ResponseWriter, r *http.Request) {
	presets := []map[string]string{
		{
			"id":          "demo",
			"name":        "LayerScope Demo (Node.js + Alpine with Leaked Secret & Caches)",
			"target":      "demo",
			"description": "Multi-layer synthetic fixture with leaked credentials (.env), apk/npm wasted caches, and high-impact CVEs",
			"isDemo":      "true",
		},
		{
			"id":          "alpine",
			"name":        "Alpine Linux (alpine:3.20)",
			"target":      "alpine:3.20",
			"description": "Ultra-lightweight Linux distribution base image",
			"isDemo":      "false",
		},
		{
			"id":          "node-alpine",
			"name":        "Node.js Alpine (node:20-alpine)",
			"target":      "node:20-alpine",
			"description": "Official Node.js 20 LTS runtime on Alpine Linux",
			"isDemo":      "false",
		},
		{
			"id":          "python-slim",
			"name":        "Python Slim (python:3.11-slim)",
			"target":      "python:3.11-slim",
			"description": "Debian-based minimal Python 3.11 runtime environment",
			"isDemo":      "false",
		},
		{
			"id":          "golang-alpine",
			"name":        "Go Alpine (golang:1.22-alpine)",
			"target":      "golang:1.22-alpine",
			"description": "Go compiler and runtime toolchain on Alpine Linux",
			"isDemo":      "false",
		},
	}
	writeJSON(w, http.StatusOK, presets)
}

func (s *Server) handleGetLayerFile(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	idxStr := chi.URLParam(r, "index")
	filePath := r.URL.Query().Get("path")
	if filePath == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "path query parameter is required"})
		return
	}

	idx, err := strconv.Atoi(idxStr)
	if err != nil || idx < 0 || idx >= len(s.snapshots) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid layer index"})
		return
	}

	snap := s.snapshots[idx]
	node := snap.Tree.Lookup(filePath)
	if node == nil {
		for _, df := range snap.DeltaFiles {
			if df.Path == filePath {
				node = df
				break
			}
		}
	}

	if node == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": fmt.Sprintf("file %q not found in layer %d", filePath, idx)})
		return
	}

	isText := isTextData(node.Data)
	content := ""
	lineCount := 0
	truncated := false

	if isText && len(node.Data) > 0 {
		maxLen := 64 * 1024
		if len(node.Data) > maxLen {
			content = string(node.Data[:maxLen])
			truncated = true
		} else {
			content = string(node.Data)
		}
		lineCount = strings.Count(content, "\n") + 1
	}

	mimeType := detectMimeType(node.Path)

	resp := map[string]interface{}{
		"path":                node.Path,
		"name":                node.Name,
		"size":                node.Size,
		"mode":                node.Mode.String(),
		"modTime":             node.ModTime,
		"isDir":               node.IsDir,
		"isSymlink":           node.IsSymlink,
		"linkTarget":          node.LinkTarget,
		"digest":              node.Digest,
		"layerIndex":          idx,
		"changeType":          node.ChangeType,
		"isWasted":            node.IsWasted,
		"wastedBytes":         node.WastedBytes,
		"wasteReason":         node.WasteReason,
		"overwrittenInLayers": node.OverwrittenInLayers,
		"isText":              isText,
		"content":             content,
		"lineCount":           lineCount,
		"truncated":           truncated,
		"mimeType":            mimeType,
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleGetWaste(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	writeJSON(w, http.StatusOK, s.waste)
}

func (s *Server) handleGetSecurity(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	writeJSON(w, http.StatusOK, s.secReport)
}

func (s *Server) handleGetSBOM(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.sbomReport == nil {
		writeJSON(w, http.StatusOK, &sbom.SBOMReport{ImageName: s.img.Reference.Original})
		return
	}

	writeJSON(w, http.StatusOK, s.sbomReport)
}

func (s *Server) handleExportCycloneDX(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.sbomReport == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "no sbom data available"})
		return
	}

	data, err := sbom.ExportCycloneDXJSON(s.sbomReport)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=\"cyclonedx.json\"")
	_, _ = w.Write(data)
}

func (s *Server) handleExportSPDX(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.sbomReport == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "no sbom data available"})
		return
	}

	data, err := sbom.ExportSPDXJSON(s.sbomReport)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=\"spdx.json\"")
	_, _ = w.Write(data)
}

func (s *Server) handleGetAdvisor(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	writeJSON(w, http.StatusOK, s.advReport)
}

func (s *Server) handleGetAdvisorDockerfile(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.img == nil || len(s.snapshots) == 0 {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "no image loaded"})
		return
	}

	finalTree := s.snapshots[len(s.snapshots)-1].Tree
	optimization := advisor.GenerateDockerfileOptimization(s.img, s.waste, finalTree)
	writeJSON(w, http.StatusOK, optimization)
}

func (s *Server) handleGetDiff(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.diffReport == nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{"status": "no_diff_loaded"})
		return
	}

	writeJSON(w, http.StatusOK, s.diffReport)
}

type analyzeRequest struct {
	Target       string `json:"target"`
	ForceRemote  bool   `json:"forceRemote"`
	Architecture string `json:"architecture"`
	Demo         bool   `json:"demo"`
}

func (s *Server) handlePostAnalyze(w http.ResponseWriter, r *http.Request) {
	var req analyzeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json payload"})
		return
	}

	scanner := security.NewScanner()
	extractor := sbom.NewExtractor()

	var commands []string
	hook := func(layerIndex int, file *oci.LayerFile, reader io.Reader) error {
		// Run scanner and extractor in a single stream pass
		var buf bytesBuffer
		tr := io.TeeReader(reader, &buf)

		cmd := ""
		if layerIndex < len(commands) {
			cmd = commands[layerIndex]
		}
		if _, err := scanner.ScanStream(layerIndex, cmd, file.Path, tr); err != nil {
			return err
		}
		if sbom.IsManifestPath(file.Path) {
			return extractor.Hook()(layerIndex, file, &buf)
		}
		return nil
	}

	var targetImg *oci.ImageAnalysis
	var err error

	if req.Demo || req.Target == "demo" {
		targetImg, err = oci.GenerateSampleImage(hook)
	} else {
		targetImg, err = oci.LoadImage(req.Target, oci.IngestionOptions{
			ForceRemote:  req.ForceRemote,
			Architecture: req.Architecture,
		}, hook)
	}

	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	snapshots := vfs.BuildLayerSnapshots(targetImg)
	wasteSummary := vfs.CalculateWastedSpace(snapshots, nil)
	finalTree := snapshots[len(snapshots)-1].Tree
	runsAsRoot, suid := security.AuditPrivileges(&targetImg.Config, finalTree)
	secReport := scanner.FinalizeAudit(finalTree, runsAsRoot, suid)
	sbomReport, _ := extractor.FinalizeReport(targetImg.Reference.Original, finalTree)

	// Enrich SBOM components with real-time OSV.dev CVE vulnerabilities (5s timeout)
	enrichCtx, enrichCancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer enrichCancel()
	extractor.EnrichWithVulnerabilities(enrichCtx, sbomReport)

	advReport := advisor.AnalyzeImageHeuristics(targetImg, wasteSummary, secReport, finalTree)

	s.mu.Lock()
	s.img = targetImg
	s.snapshots = snapshots
	s.waste = wasteSummary
	s.secReport = secReport
	s.sbomReport = sbomReport
	s.advReport = advReport
	s.mu.Unlock()

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":          "success",
		"reference":       targetImg.Reference,
		"efficiencyScore": advReport.EfficiencyScore,
		"grade":           advReport.Grade,
		"totalSizeBytes":  targetImg.TotalSizeBytes,
		"totalFilesCount": targetImg.TotalFilesCount,
		"layersCount":     len(snapshots),
	})
}

type diffRequest struct {
	ImageA string `json:"imageA"`
	ImageB string `json:"imageB"`
	ArchA  string `json:"archA"`
	ArchB  string `json:"archB"`
}

func (s *Server) handlePostDiff(w http.ResponseWriter, r *http.Request) {
	var req diffRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid json payload"})
		return
	}

	imgA, errA := oci.LoadImage(req.ImageA, oci.IngestionOptions{Architecture: req.ArchA}, nil)
	if errA != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("load image A: %v", errA)})
		return
	}
	snapsA := vfs.BuildLayerSnapshots(imgA)
	treeA := snapsA[len(snapsA)-1].Tree

	imgB, errB := oci.LoadImage(req.ImageB, oci.IngestionOptions{Architecture: req.ArchB}, nil)
	if errB != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": fmt.Sprintf("load image B: %v", errB)})
		return
	}
	snapsB := vfs.BuildLayerSnapshots(imgB)
	treeB := snapsB[len(snapsB)-1].Tree

	report := vfs.DiffVFSTrees(
		imgA.Reference.Original, treeA, imgA.TotalSizeBytes,
		imgB.Reference.Original, treeB, imgB.TotalSizeBytes,
	)

	s.mu.Lock()
	s.diffReport = report
	s.mu.Unlock()

	writeJSON(w, http.StatusOK, report)
}

func (s *Server) handleStaticOrFallback(w http.ResponseWriter, r *http.Request) {
	if s.staticFS != nil {
		// Serve embedded files
		f, err := s.staticFS.Open(r.URL.Path)
		if err == nil {
			_ = f.Close()
			http.FileServer(s.staticFS).ServeHTTP(w, r)
			return
		}
		// Fallback to index.html for SPA routes
		indexFile, err := s.staticFS.Open("/index.html")
		if err == nil {
			_ = indexFile.Close()
			r.URL.Path = "/index.html"
			http.FileServer(s.staticFS).ServeHTTP(w, r)
			return
		}
	}

	// Dynamic fallback HTML workbench if UI bundle is not embedded
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(GetFallbackStudioHTML(s.img, s.advReport, s.waste, s.secReport)))
}

type bytesBuffer struct {
	buf []byte
}

func (b *bytesBuffer) Write(p []byte) (n int, err error) {
	b.buf = append(b.buf, p...)
	return len(p), nil
}

func (b *bytesBuffer) Read(p []byte) (n int, err error) {
	if len(b.buf) == 0 {
		return 0, io.EOF
	}
	n = copy(p, b.buf)
	b.buf = b.buf[n:]
	return n, nil
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func isTextData(data []byte) bool {
	if len(data) == 0 {
		return true
	}
	checkLen := len(data)
	if checkLen > 1024 {
		checkLen = 1024
	}
	for i := 0; i < checkLen; i++ {
		b := data[i]
		if b == 0 {
			return false
		}
	}
	return true
}

func detectMimeType(p string) string {
	ext := strings.ToLower(path.Ext(p))
	switch ext {
	case ".json":
		return "application/json"
	case ".js", ".mjs", ".ts", ".tsx":
		return "application/javascript"
	case ".html", ".htm":
		return "text/html"
	case ".css":
		return "text/css"
	case ".yaml", ".yml":
		return "application/yaml"
	case ".xml":
		return "application/xml"
	case ".sh", ".bash":
		return "application/x-sh"
	case ".py":
		return "text/x-python"
	case ".go":
		return "text/x-go"
	case ".md", ".txt":
		return "text/plain"
	case ".conf", ".ini", ".env":
		return "text/plain"
	default:
		return "text/plain"
	}
}

