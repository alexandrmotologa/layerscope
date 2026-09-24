# Command Line Reference

LayerScope provides subcommands for interactive exploration, terminal inspection, image diffing, and automated CI auditing.

## Global Options

- `--help, -h`: Show help for any command.
- `--verbose, -v`: Enable verbose debug logging.
- `--format`: Specify output format for machine-readable commands (`text`, `json`, `markdown`, `junit`, `html`).

## Subcommands

### 1. `layerscope analyze`

Analyzes an image and launches the embedded web studio in your browser.

```bash
# Analyze a local image from Docker daemon
layerscope analyze myapp:latest

# Analyze a remote OCI image without pulling to local daemon
layerscope analyze ghcr.io/org/repo:v1.2.0 --remote

# Analyze an exported image tarball
layerscope analyze ./saved-image.tar

# Run with demo fixture data (no docker daemon required)
layerscope analyze --demo

# Specify studio port
layerscope analyze myapp:latest --port 50060
```

### 2. `layerscope tui`

Launches a terminal user interface modeled after classic layer visualizers, with keyboard navigation, file tree exploration, and color-coded file statuses.

```bash
# Inspect image in terminal
layerscope tui alpine:latest

# Filter wasted files immediately
layerscope tui myapp:latest --wasted-only
```

### 3. `layerscope diff`

Compares two container images side by side. It highlights files added, deleted, modified, or grown between tags or architecture variants.

```bash
# Compare two release tags
layerscope diff myapp:1.0.0 myapp:1.1.0

# Compare architecture targets for a multi-arch manifest
layerscope diff myapp:latest --arch-a linux/amd64 --arch-b linux/arm64
```

### 4. `layerscope audit`

Headless verification intended for continuous integration. Returns exit code `1` if efficiency falls below a minimum threshold or if secrets are discovered.

```bash
# Fail CI build if efficiency is below 85% or if secrets are present
layerscope audit myapp:latest \
  --min-efficiency 85 \
  --fail-on-secrets \
  --sbom-out ./dist/sbom.json \
  --format json
```

### 5. `layerscope sbom`

Extracts system and application dependencies from an image and outputs a standardized bill of materials.

```bash
# Output CycloneDX 1.5 JSON to stdout
layerscope sbom myapp:latest --format cyclonedx

# Output SPDX 2.3 JSON to a file
layerscope sbom myapp:latest --format spdx --output sbom.json
```

### 6. `layerscope slim`

Squashes multi-layer cumulative files into a clean, minimal single-layer Docker/OCI tar archive, purging whiteout-deleted remnants (including leaked credentials) and build caches.

```bash
# Squash multi-layer image into a clean single layer
layerscope slim myapp:latest -o myapp-slim.tar

# Keep package caches instead of purging
layerscope slim myapp:latest --purge-caches=false -o myapp-full.tar

# Import directly into Docker
docker load -i myapp-slim.tar
```
