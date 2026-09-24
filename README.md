<p align="center">
  <img src="docs/images/logo.png?raw=true" alt="LayerScope Mascot Logo" width="128" style="border-radius: 28px;" />
</p>

<h1 align="center">LayerScope</h1>

<p align="center">
  <strong>Container layer inspector, intermediate secret auditor, multi-arch comparator, SBOM generator, and image squasher for OCI and Docker images.</strong>
</p>

<p align="center">
  <a href="https://golang.org"><img src="https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat-square&logo=go" alt="Go Version" /></a>
  <a href="https://github.com/alexandrmotologa/layerscope/blob/main/LICENSE"><img src="https://img.shields.io/badge/License-MIT-blue?style=flat-square" alt="License" /></a>
  <a href="https://cyclonedx.org"><img src="https://img.shields.io/badge/SBOM-CycloneDX%20%7C%20SPDX-green?style=flat-square" alt="SBOM" /></a>
  <a href="https://osv.dev"><img src="https://img.shields.io/badge/Vulnerabilities-OSV.dev-purple?style=flat-square" alt="OSV Database" /></a>
  <a href="https://github.com/alexandrmotologa/layerscope/actions"><img src="https://img.shields.io/badge/CI-GitHub%20Actions-2088FF?style=flat-square&logo=githubactions" alt="CI" /></a>
</p>

<p align="center">
  <img src="docs/images/layerscope_demo.gif?raw=true" alt="LayerScope Studio Demo" width="100%" />
</p>

---

## Why LayerScope

Modern container workflows build multi-architecture images across x86_64 and ARM64. While earlier CLI tools rendered static layer sizes in the terminal, they did not detect credentials baked into intermediate layers, lacked side-by-side architecture diffing, could not squash wasted cache remnants, and did not produce standards-compliant SBOM files with real-time CVE correlation.

LayerScope addresses these gaps in a single, zero-dependency Go binary:

- **Intermediate Layer Secret Auditing**: Scans every layer tarball in an image history to detect credentials committed in intermediate steps, even when removed by subsequent `RUN rm` commands or `.dockerignore` mistakes.
- **Side-by-Side Tag and Architecture Diffing**: Compares two image tags or architecture variants (such as `linux/amd64` against `linux/arm64`) to pinpoint size deltas, modified binaries, and added files.
- **Image Squasher & Purger (`layerscope slim`)**: Squashes bloated multi-layer images into a clean single layer archive, purging whiteout-deleted remnants and build caches (`docker load -i slim.tar`).
- **Real-Time OSV.dev CVE Matching**: Enriches SBOM software components with known CVE and GHSA vulnerability advisories and fix versions directly from the OSV database.
- **Dockerfile Auto-Fixer & Optimizer**: Generates production-hardened `Dockerfile.optimized` and `.dockerignore` files with layer cache reordering, package cache purging, and non-root execution.
- **Interactive File Inspection Drawer**: Inspects text files, scripts, and configuration contents directly from the web studio without extracting layers manually.
- **Multi-Source Image Ingestion**: Reads directly from local Docker or Podman daemons, exported tarballs, or remote registries (Docker Hub, GitHub Container Registry, Quay) without needing local daemon storage.
- **Automated SBOM Generation**: Extracts packages from Alpine apk, Debian/Ubuntu dpkg, and language lockfiles into CycloneDX v1.5 and SPDX v2.3 JSON.
- **Single-Binary Delivery**: Compiles the Go engine and embedded React studio into an independent binary with zero runtime dependencies.

---

## Visual Overview

### 1. Interactive Layer Timeline & File Inspection Drawer
Inspect every layer modification, filter files by category (binaries, caches, configuration, libraries, or wasted space), and click any file to open the slide-out viewer with syntax detection and mode metadata.

<p align="center">
  <img src="docs/images/layerscope-preview-drawer.png?raw=true" alt="LayerScope File Preview Drawer" width="100%" />
</p>

### 2. Intermediate Layer Secret Auditor
Identifies leaked API keys, tokens, SSH keys, database credentials, and `.env` files buried in historical layer tarballs that remain accessible to anyone pulling the image, even after deletion.

<p align="center">
  <img src="docs/images/layerscope-secrets.png?raw=true" alt="LayerScope Secret Auditor" width="100%" />
</p>

### 3. Software Bill of Materials (SBOM) & OSV.dev CVE Tracker
Generates package inventories from OS distributions (Alpine apk, Debian dpkg) and application runtimes (npm, pip, go.mod), querying the OSV.dev batch API for unpatched CVE advisories.

<p align="center">
  <img src="docs/images/layerscope-sbom.png?raw=true" alt="LayerScope SBOM and CVE Tracker" width="100%" />
</p>

### 4. Optimization Advisor & Dockerfile Auto-Fixer
Analyzes wasted layer payloads, detects missing package cache cleanups, and generates a drop-in `Dockerfile.optimized` along with a tuned `.dockerignore`.

<p align="center">
  <img src="docs/images/layerscope-advisor.png?raw=true" alt="LayerScope Optimization Advisor" width="100%" />
</p>

### 5. Side-by-Side Tag & Multi-Architecture Comparator
Evaluates differences between releases or cross-compiled architectures (`amd64` vs `arm64`), highlighting binary size discrepancies and file drift.

<p align="center">
  <img src="docs/images/layerscope-diff.png?raw=true" alt="LayerScope Tag and Architecture Comparator" width="100%" />
</p>

---

## Installation

### From Source

Ensure Go 1.23 or newer is installed:

```bash
git clone https://github.com/alexandrmotologa/layerscope.git
cd layerscope
go build -o bin/layerscope ./cmd/layerscope
```

To build with the embedded Web Studio:

```bash
cd ui
npm install
npm run build
cd ..
go build -tags embed_ui -o bin/layerscope ./cmd/layerscope
```

---

## Quick Start

### 1. Analyze an Image in Web Studio

```bash
# Analyze a local Docker image
layerscope analyze myapp:latest

# Analyze a remote image without pulling
layerscope analyze ghcr.io/org/service:v1.0.0 --remote

# Analyze an exported image archive
layerscope analyze ./saved-image.tar

# Run with demo fixture data (works without Docker or network)
layerscope analyze --demo
```

The web studio opens automatically at `http://localhost:50060`.

### 2. Squash and Purge Bloat (`layerscope slim`)

```bash
# Squash multi-layer image and export clean tarball
layerscope slim myapp:latest -o myapp-slim.tar

# Import directly back into Docker
docker load -i myapp-slim.tar
```

### 3. Terminal TUI Mode

```bash
layerscope tui myapp:latest
```

Keyboard controls:
- `Up` / `Down`: Navigate layers or file tree
- `Tab`: Switch pane focus
- `Enter`: Expand or collapse directory
- `W`: Toggle wasted space filter
- `Esc` / `Q`: Exit

### 4. Compare Tags or Architectures

```bash
# Compare two image tags
layerscope diff myapp:1.0.0 myapp:1.1.0

# Compare architecture variants for a multi-arch image
layerscope diff myapp:latest --arch-a linux/amd64 --arch-b linux/arm64
```

### 5. Continuous Integration Audit

```bash
layerscope audit myapp:latest \
  --min-efficiency 85 \
  --fail-on-secrets \
  --sbom-out ./dist/sbom.json \
  --format markdown
```

If image efficiency drops below 85% or if intermediate layers contain leaked secrets, LayerScope returns a non-zero exit code.

### 6. GitHub Action Integration

Add LayerScope to your GitHub Actions workflow:

```yaml
- name: Audit Container Image
  uses: alexandrmotologa/layerscope@main
  with:
    image: 'myapp:latest'
    min-score: '75'
    fail-on-secrets: 'true'
    format: 'markdown'
    output: 'audit-report.md'
```

---

## Architecture

```
LayerScope CLI (analyze, slim, diff, audit, sbom, tui)
  ├── Terminal TUI (Bubbletea dual-pane interface)
  ├── Embedded Web Studio (React 19, Monaco, Lucide)
  └── Go Engine
        ├── Image Sources (Docker socket, OCI remote registry, tarball, fixtures)
        ├── Virtual File System (Layer accumulation, whiteout resolution, diffing)
        ├── Security Scanner (Intermediate layer credential detector)
        ├── SBOM & Vulnerability Engine (OSV.dev batch client, CycloneDX/SPDX)
        ├── Image Squasher (Single-layer clean tarball exporter)
        └── Advisor (Efficiency scoring, Dockerfile Auto-Fixer)
```

---

## Documentation

- [Architecture Guide](docs/ARCHITECTURE.md)
- [CLI Reference](docs/CLI_REFERENCE.md)
- [Specification and Roadmap](LAYERSCOPE_SPEC_AND_ROADMAP.md)

---

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.
