# LayerScope

LayerScope is a container layer inspector, multi-architecture diff tool, secret leak auditor, and SBOM generator for OCI and Docker images. It packages an interactive terminal interface, a headless CI scanner, and an embedded web studio into a single Go binary.

## Why LayerScope

Modern container development requires building multi-architecture images for both x86_64 and ARM64 platforms. While earlier tools visualized layer sizes in the terminal, they did not support side-by-side architecture comparison, intermediate layer credential scanning, or standardized SBOM generation.

LayerScope addresses these gaps:

- **Intermediate Layer Secret Auditing**: Scans every layer tarball in an image history to detect credentials committed in intermediate steps, even when removed by subsequent `RUN rm` commands.
- **Side-by-Side Tag and Architecture Diffing**: Compares two images or architecture variants (such as `linux/amd64` against `linux/arm64`) to pinpoint size deltas, modified binaries, and added files.
- **Multi-Source Image Ingestion**: Reads directly from local Docker or Podman daemons, exported tarballs, or remote registries (Docker Hub, GitHub Container Registry, Quay) without needing local daemon storage.
- **Automated SBOM Generation**: Extracts packages from Alpine apk, Debian/Ubuntu dpkg, and language lockfiles into CycloneDX v1.5 and SPDX v2.3 JSON.
- **Actionable Optimization Heuristics**: Calculates an efficiency score and suggests concrete Dockerfile adjustments, including cache cleanup and multi-stage reordering.
- **Single-Binary Delivery**: Compiles the Go engine and embedded React studio into an independent binary with zero runtime dependencies.

## Installation

### From Source

Ensure Go 1.23 or newer is installed:

```bash
git clone https://github.com/alexandrmotologa/layerscope.git
cd layerscope
go build -o bin/layerscope ./cmd/layerscope
```

To run with the embedded web studio:

```bash
cd ui
npm install
npm run build
cd ..
go build -tags embed_ui -o bin/layerscope ./cmd/layerscope
```

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

### 2. Terminal TUI Mode

```bash
layerscope tui myapp:latest
```

Navigate layers with the arrow keys, toggle file tree expansion with `Enter`, switch panes with `Tab`, and toggle wasted space view with `W`.

### 3. Compare Tags or Architectures

```bash
# Compare two image tags
layerscope diff myapp:1.0.0 myapp:1.1.0

# Compare architecture variants for a multi-arch image
layerscope diff myapp:latest --arch-a linux/amd64 --arch-b linux/arm64
```

### 4. Continuous Integration Audit

```bash
layerscope audit myapp:latest \
  --min-efficiency 85 \
  --fail-on-secrets \
  --sbom-out ./dist/sbom.json \
  --format json
```

If image efficiency drops below 85% or if intermediate layers contain leaked secrets, LayerScope returns a non-zero exit code.

## Architecture

```
LayerScope CLI (analyze, diff, audit, tui, serve)
  ├── Terminal TUI (Bubbletea dual-pane interface)
  ├── Embedded Web Studio (React 19, Monaco, Lucide)
  └── Go Engine
        ├── Image Sources (Docker socket, OCI remote registry, tarball, fixtures)
        ├── Virtual File System (Layer accumulation, whiteout resolution, diffing)
        ├── Security Scanner (Intermediate layer credential detector)
        ├── SBOM Engine (Alpine, Debian, language package manifests)
        └── Advisor (Efficiency scoring and remediation rules)
```

## Documentation

- [Architecture Guide](file:///B:/workgit/layerscope/docs/ARCHITECTURE.md)
- [CLI Reference](file:///B:/workgit/layerscope/docs/CLI_REFERENCE.md)
- [Specification and Roadmap](file:///B:/workgit/layerscope/LAYERSCOPE_SPEC_AND_ROADMAP.md)

## License

This project is licensed under the MIT License. See [LICENSE](file:///B:/workgit/layerscope/LICENSE) for details.
