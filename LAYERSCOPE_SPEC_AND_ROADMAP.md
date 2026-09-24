# Engineering Specification & Implementation Blueprint: LayerScope
> The Next-Gen OCI & Docker Container Layer Inspector, Multi-Arch Diffing, SBOM & Secret Leak Auditor (Reimagining wagoodman/dive - 44.6k Stars)

---

## 1. Executive Summary & Market Opportunity

### 1.1 The Market Vacuum
In modern cloud-native engineering, container images are the fundamental deployment unit. Yet, optimizing image size, debugging container build bloat, and auditing security flaws inside image layers remains notoriously painful.

For years, **`wagoodman/dive`** (44,600+ GitHub stars) was the de facto standard tool for exploring Docker image layers and discovering wasted space. However:
* **Abandoned/Stagnant:** `dive` has received virtually no major architectural updates since 2021.
* **Terminal-Only (TUI):** It is locked to a basic terminal interface that struggles with massive multi-gigabyte layers and lacks visual side-by-side diffing.
* **No Multi-Arch Support:** In 2026, every production team builds multi-architecture images (`linux/amd64` vs `linux/arm64` for AWS Graviton / Apple Silicon). `dive` cannot compare architecture variants side-by-side to explain why an ARM64 image is 250MB larger.
* **Zero Security / SBOM Intelligence:** `dive` only looks at file sizes. It cannot detect intermediate layer secret leaks (passwords, `.env` files, SSH keys, AWS credentials left behind before a subsequent `RUN rm` command), nor does it generate modern Software Bill of Materials (SBOM) standards (CycloneDX, SPDX).
* **Tag Diffing Blindness:** Developers cannot compare `myapp:v1.2.0` against `myapp:v1.2.1` to visually see what files or dependencies caused an unexpected image bloat.

### 1.2 The Solution: LayerScope
**LayerScope** is a modern, single-binary, local-first OCI container layer analysis studio and security auditor built in **Go 1.23+** with an embedded **React 19 + Monaco** visual dashboard and an interactive terminal CLI.

* **Multi-Source Image Ingestion:** Inspects images directly from local Docker daemon (`/var/run/docker.sock` or Windows named pipe `//./pipe/docker_engine`), Podman, Containerd, exported tarballs (`docker save`), or directly from remote OCI registries (Docker Hub, GitHub Container Registry, AWS ECR, Quay) without needing `docker pull`.
* **Side-by-Side Tag & Architecture Diffing:** Visually diff two images or architectures (e.g. `linux/amd64` vs `linux/arm64`) to see identical, modified, added, and deleted files with byte-level accuracy.
* **Intermediate Layer Secret & Credential Leak Scanner:** Scans every historical layer tarball for accidentally committed secrets (`.env`, `.npmrc`, `.git/config`, private keys, JWTs) that are hidden in intermediate layers and still downloadable by anyone with image pull rights.
* **Automated SBOM & Package Vulnerability Matrix:** Extracts installed OS packages (Alpine `apk`, Debian/Ubuntu `dpkg`, RedHat `rpm`) and runtime dependencies (npm, pip, go.mod, cargo) into CycloneDX / SPDX JSON and matches against OSV (Open Source Vulnerabilities) database.
* **Image Efficiency Score & Actionable Remediation:** Scores image efficiency (0–100%) and generates copy-pasteable Dockerfile fixes (e.g. combining `RUN` instructions, using `.dockerignore`, switching to distroless/scratch base).
* **Single-Binary Zero-Dependency Packaging:** Go binary with embedded Web UI via `go:embed` (<25MB RSS, <30ms startup).

---

## 2. Core Architecture & Tech Stack

```
┌────────────────────────────────────────────────────────────────────────┐
│                        LayerScope Architecture                         │
└────────────────────────────────────────────────────────────────────────┘

[ Web UI / Local Desktop Studio ] (http://localhost:50060)
         │
         ▼  (HTTP REST / Server-Sent Events)
[ LayerScope Single Binary ] (Go 1.23+ Engine)
   ├── Embedded Web Server: go:embed (Vite + React 19 + Monaco + Lucide)
   ├── Image Acquisition & OCI Storage Engine
   │     ├── Local Docker Engine API (UNIX socket / Windows named pipe)
   │     ├── Podman / Containerd socket provider
   │     ├── Remote OCI Registry Client (google/go-containerregistry)
   │     └── Archive Extractor (tar, tar.gz, tar.zstd stream processor)
   ├── Layer Analysis & Virtual File System (VFS)
   │     ├── Layer Accumulator (Whiteout & Opaque directory resolver)
   │     ├── File Change Classifier (Added, Modified, Deleted, Unchanged)
   │     └── Multi-Arch & Tag Diffing Engine
   ├── Security & Intelligence Modules
   │     ├── Secret & Entropy Scanner (gitleaks rule matcher)
   │     ├── SBOM Package Extractor (dpkg, apk, rpm, npm, pip, go)
   │     └── Efficiency & Waste Scoring Heuristics
   └── Exporters & CLI Dispatcher
         ├── Interactive Terminal TUI (charmbracelet/bubbletea)
         ├── Headless CI Auditor (JSON, Markdown summary, JUnit XML)
         └── SBOM Exporter (CycloneDX v1.5, SPDX v2.3)
```

### 2.1 Backend Technology
* **Language:** Go 1.23+
* **OCI & Registry Client:** `github.com/google/go-containerregistry` (Google's production library for querying and downloading remote OCI images without Docker daemon).
* **Docker Client:** `github.com/docker/docker/client` (native Docker Engine SDK).
* **Tar Stream Processing:** Custom streaming archive walker with zero-copy hashing and memory-mapped file indexing.
* **Entropy & Secret Rules:** Rego / regex patterns derived from Gitleaks for high-fidelity detection with low false-positive rates.
* **HTTP Router:** `github.com/go-chi/chi/v5` + `github.com/go-chi/cors`.
* **CLI Framework:** `github.com/spf13/cobra` + `github.com/charmbracelet/bubbletea` (for terminal TUI mode).

### 2.2 Frontend Technology
* **Build & Bundler:** Vite + React 19 + TypeScript.
* **State Management:** Zustand with persistent settings in `localStorage`.
* **Virtualization:** `@tanstack/react-virtual` for silky smooth rendering of 500k+ file tree hierarchies.
* **Diff Viewer:** Monaco Editor (`@monaco-editor/react`) for comparing configuration and file metadata.
* **Icons & Styling:** Tailwind CSS + Lucide Icons + glassmorphic dark mode design tokens.

---

## 3. Key Feature Specifications

### 3.1 Multi-Layer Virtual File System (VFS) & Whiteout Handling
* Containers use overlay file systems where upper layers overwrite lower layers and whiteout files (`.wh.<filename>` or opaque markers) represent deleted items.
* LayerScope computes the cumulative state at every layer:
  * **Layer-by-Layer Tree:** View the exact file system state up to that specific instruction.
  * **File Diff Breakdown:** Colors files as:
    * 🟢 **Added** (new in this layer)
    * 🟡 **Modified** (overwritten from a previous layer)
    * 🔴 **Deleted** (marked with whiteout)
    * ⚪ **Unchanged** (inherited)
  * **Wasted Space Calculator:** Flags files overwritten or deleted in later layers that still consume bandwidth in the download payload.

### 3.2 Side-by-Side Tag & Architecture Comparison
* **Tag Diffing:** Compare `myorg/backend:1.4.0` vs `myorg/backend:1.4.1`:
  * Tree comparison highlighting newly bundled node_modules, binaries, or config drift.
  * Layer-by-layer size delta chart (+45MB in step 7: apt-get install without cleanup).
* **Architecture Diffing:** Inspect multi-arch manifests (`linux/amd64` vs `linux/arm64`):
  * Compare binary architecture targets, dynamic shared libraries (`.so` dependencies), and compiler artifacts.

### 3.3 Intermediate Layer Secret & Credential Audit
* Many Dockerfiles execute:
  ```dockerfile
  COPY .env /app/.env
  RUN build-production-bundle.sh
  RUN rm /app/.env
  ```
  The `.env` file is permanently stored in the intermediate layer tarball and can be easily extracted by anyone who can pull the image!
* LayerScope automatically scans every individual layer tarball for:
  * Private SSH/RSA/EC keys, PGP keys.
  * AWS access keys, GCP service account JSONs, Azure connection strings.
  * GitHub/GitLab personal access tokens.
  * `.env`, `.netrc`, `.docker/config.json`, `.npmrc` auth tokens.
* Displays the exact Dockerfile instruction that introduced the leaked secret with full redaction in the UI.

### 3.4 Automated SBOM & Package Vulnerability Matrix
* Detects and parses system package manifests:
  * Alpine Linux (`/lib/apk/db/installed`)
  * Debian/Ubuntu (`/var/lib/dpkg/status`)
  * RHEL/CentOS/Fedora/Rocky (`/var/lib/rpm`)
* Detects application dependencies:
  * Node.js (`package-lock.json`, `pnpm-lock.yaml`, `node_modules`)
  * Python (`requirements.txt`, `Pipfile.lock`, `site-packages`)
  * Go binaries (compiled-in module dependencies via `debug.ReadBuildInfo`)
  * Rust (`Cargo.lock`)
* Generates downloadable SBOM in **CycloneDX v1.5 JSON** and **SPDX v2.3 JSON**.

### 3.5 Actionable Optimization Engine
* Calculates an overall **Image Efficiency Rating** (0% to 100%).
* Generates concrete optimization tips:
  * *Tip 1:* "Combine `RUN apt-get update` and `RUN apt-get install` and append `&& rm -rf /var/lib/apt/lists/*` to save 38.4 MB."
  * *Tip 2:* "File `/root/.cache/pip` was left in Layer 5, wasting 42.1 MB."
  * *Tip 3:* "Switch base image from `node:20` (1.1 GB) to `node:20-alpine` (178 MB) or `distroless/nodejs`."

---

## 4. CLI Command-Line Specification

```bash
# Analyze a local Docker image and launch the visual web studio
layerscope analyze myapp:latest

# Analyze a remote OCI image without pulling to local daemon
layerscope analyze ghcr.io/org/api:v2.1.0 --remote

# Analyze an exported image tarball
layerscope analyze ./saved-image.tar

# Run in terminal TUI mode (classic dive experience on steroids)
layerscope tui myapp:latest

# Visually diff two image tags side-by-side
layerscope diff myapp:v1.0.0 myapp:v1.1.0

# Compare architectures for a multi-arch manifest
layerscope diff myapp:latest --arch-a linux/amd64 --arch-b linux/arm64

# Run headless in CI/CD pipeline (exit code 1 if efficiency < threshold or secrets found)
layerscope audit myapp:latest \
  --min-efficiency 85 \
  --fail-on-secrets \
  --sbom-out ./sbom.json \
  --format json
```

---

## 5. Complete Project Directory Layout

```
layerscope/
├── cmd/
│   └── layerscope/
│       └── main.go                         # Cobra CLI entrypoint & subcommands
├── pkg/
│   ├── oci/
│   │   ├── client.go                       # Docker daemon, Podman, and registry dispatcher
│   │   ├── docker.go                       # Docker Engine socket client
│   │   ├── remote.go                       # go-containerregistry remote fetcher
│   │   ├── archive.go                      # Tar stream reader and layer unpacker
│   │   └── types.go                        # Manifest, Config, and Layer descriptors
│   ├── vfs/
│   │   ├── tree.go                         # Virtual File System tree representation
│   │   ├── node.go                         # File/Directory node with metadata and sizes
│   │   ├── layer_builder.go                # Cumulative layer accumulator & whiteout solver
│   │   ├── waste.go                        # Wasted file detection & duplicate detector
│   │   └── diff.go                         # Two-image / two-architecture diff calculator
│   ├── security/
│   │   ├── secret_scanner.go               # Intermediate layer credential & token detector
│   │   ├── rules.go                        # Regex & entropy rule definitions
│   │   └── types.go                        # Finding severity, pattern match, location
│   ├── sbom/
│   │   ├── extractor.go                    # OS and language package scanner
│   │   ├── alpine.go                       # apk installed database parser
│   │   ├── debian.go                       # dpkg status file parser
│   │   ├── npm.go                          # package-lock.json / pnpm parser
│   │   ├── python.go                       # requirements / dist-info parser
│   │   └── cyclonedx.go                    # CycloneDX & SPDX JSON formatter
│   ├── advisor/
│   │   ├── efficiency.go                   # Overall score calculator (0-100)
│   │   ├── rules.go                        # Dockerfile anti-pattern heuristics
│   │   └── recommendations.go              # Actionable remediation generator
│   └── server/
│       ├── api.go                          # REST API router (Chi)
│       ├── handlers_image.go               # Image analysis & layer endpoints
│       ├── handlers_diff.go                # Diff comparison endpoints
│       ├── handlers_security.go            # Secret findings & SBOM endpoints
│       └── static.go                       # go:embed static production frontend
├── ui/
│   ├── index.html                          # Single-page application entrypoint
│   ├── package.json                        # Vite, React 19, Tailwind, Monaco
│   ├── tsconfig.json                       # Strict TypeScript configuration
│   ├── src/
│   │   ├── api/
│   │   │   └── client.ts                   # Type-safe API client
│   │   ├── components/
│   │   │   ├── Header.tsx                  # Image selector, arch badge, score pill
│   │   │   ├── LayerList.tsx               # Left sidebar: Dockerfile instructions & layer sizes
│   │   │   ├── FileTree.tsx                # Center virtualized file system explorer
│   │   │   ├── FileDetailsModal.tsx        # File metadata, waste details, diff view
│   │   │   ├── DiffWorkspace.tsx           # Side-by-side two-image comparator
│   │   │   ├── SecretAuditor.tsx           # Security tab: leaked secrets in intermediate layers
│   │   │   ├── SbomViewer.tsx              # Software Bill of Materials table & search
│   │   │   └── AdvisorPanel.tsx            # Optimization recommendations & Dockerfile tips
│   │   ├── types/
│   │   │   └── index.ts                    # TypeScript interface definitions
│   │   ├── App.tsx                         # Root workbench state & tab navigation
│   │   └── main.tsx                        # React 19 mount point
├── Makefile                                # Build, test, and release targets
├── go.mod                                  # Go dependencies
├── go.sum                                  # Checksums
└── README.md                               # Project documentation
```

---

## 6. Implementation Roadmap

### Phase 1: Go Module Setup & OCI Extraction Engine
- [x] 1.1 Initialize Go module `github.com/alexandrmotologa/layerscope` and configure dependencies (`go-containerregistry`, `docker`, `cobra`, `chi`, `bubbletea`).
- [x] 1.2 Implement `pkg/oci/docker.go` connecting to local Docker daemon (Linux socket and Windows named pipe).
- [x] 1.3 Implement `pkg/oci/remote.go` fetching manifests and layer blobs directly from remote OCI registries without Docker daemon.
- [x] 1.4 Implement `pkg/oci/archive.go` streaming and decompressing tar layers with sha256 tracking.
- [x] 1.5 Implement `pkg/oci/sample.go` synthetic fixture generator producing deterministic multi-layer images with intermediate leaks and OS packages for offline testing and immediate demo mode.
- [x] 1.6 Implement `pkg/oci/client.go` unifying remote, docker daemon, archive, and fixture ingestion sources.
- [x] 1.7 Write automated tests verifying layer extraction against sample multi-layer Docker images.

### Phase 2: Virtual File System & Layer Diffing Core
- [x] 2.1 Implement `pkg/vfs/tree.go` representing cumulative directory structures with fast path lookups.
- [x] 2.2 Implement whiteout resolution (`.wh.<file>` and `.wh..wh..opq`) correctly handling deleted and shadowed files.
- [x] 2.3 Implement file status classification (`Added`, `Modified`, `Deleted`, `Unchanged`) per layer.
- [x] 2.4 Implement wasted space detection calculating duplicate files, cross-layer overwrites, and post-delete remnants.
- [x] 2.5 Implement `pkg/vfs/diff.go` for comparative diffing between two arbitrary images or architecture variants.
- [x] 2.6 Write comprehensive unit tests for VFS operations, whiteout edge cases, and diffing accuracy.

### Phase 3: Security Scanner & SBOM Engine
- [x] 3.1 Implement `pkg/security/rules.go` with high-confidence regex and Shannon entropy rules for AWS keys, SSH keys, tokens, and `.env` files.
- [x] 3.2 Implement `pkg/security/secret_scanner.go` scanning layer archive contents during extraction with masking.
- [x] 3.3 Implement `pkg/security/privilege_scanner.go` auditing root execution and SUID/SGID binary risks.
- [x] 3.4 Implement `pkg/sbom/alpine.go` and `debian.go` parsing native OS installed package manifests.
- [x] 3.5 Implement `pkg/sbom/npm.go`, `python.go`, and `golang.go` detecting language runtime dependency locks.
- [x] 3.6 Implement `pkg/sbom/cyclonedx.go` exporting standardized CycloneDX v1.5 JSON and SPDX v2.3.
- [x] 3.7 Write automated tests verifying 100% detection of leaked intermediate secrets and dependency resolution.

### Phase 4: Advisor & Optimization Heuristics
- [x] 4.1 Implement `pkg/advisor/efficiency.go` calculating composite efficiency score (0–100%).
- [x] 4.2 Implement pattern detectors for Dockerfile anti-patterns (package manager cache leaks, build artifact leftovers).
- [x] 4.3 Implement layer cache invalidation analyzer highlighting instructions that unnecessarily bust build cache.
- [x] 4.4 Implement concrete recommendation builder providing copy-pasteable Dockerfile fixes.
- [x] 4.5 Implement CI output formats (`--format json`, `--format markdown`, `--format junit`, `--format html`).

### Phase 5: Terminal TUI Experience
- [x] 5.1 Implement Bubbletea TUI in `pkg/tui/` mirroring the classic dual-pane layout (Layers on left, File Tree on right).
- [x] 5.2 Implement keyboard navigation (arrow keys, search filter `/`, tab switching, layer inspection `Tab`).
- [x] 5.3 Add color-coded indicators for Added, Modified, Deleted, and Wasted files.
- [x] 5.4 Add quick toggle for showing only wasted files (`W` key) and secret findings.

### Phase 6: Embedded Web Studio & Single-Binary Delivery
- [x] 6.1 Scaffold Vite + React 19 + TypeScript in `ui/`.
- [x] 6.2 Build responsive Studio layout with Layer Timeline, Virtualized File Tree, and Search Filter.
- [x] 6.3 Implement Side-by-Side Tag & Architecture Comparator view with delta highlights.
- [x] 6.4 Implement Secret Audit tab displaying intermediate layer leak findings with line-number context.
- [x] 6.5 Implement SBOM tab with searchable package table and 1-click JSON export.
- [x] 6.6 Build REST API in `pkg/server/` and embed production UI using `go:embed`.
- [x] 6.7 Package single-binary release and verify complete end-to-end functionality on macOS, Linux, and Windows.

---

## 7. Verification & Acceptance Criteria
1. **OCI Compatibility:** Must successfully inspect images from Docker daemon, Containerd, local `.tar` archives, and remote registries (`docker.io`, `ghcr.io`, `quay.io`).
2. **Multi-Arch Support:** Must cleanly diff `linux/amd64` against `linux/arm64` and list every byte-level difference.
3. **Secret Detection:** Synthetic test image with an intermediate `COPY .env` followed by `RUN rm .env` must be detected with 100% precision.
4. **Performance:** Must parse a 1.5 GB multi-layer image in under 3.5 seconds on standard developer laptops with <80MB peak RAM.
5. **Zero External Dependencies:** The final binary must be statically compiled with zero runtime dependencies.

---

## 8. Kick-Off Prompt for Subagent

```markdown
Please read [LAYERSCOPE_SPEC_AND_ROADMAP.md](file:///B:/workgit/layerscope/LAYERSCOPE_SPEC_AND_ROADMAP.md) in full and execute Phase 1: scaffold the Go 1.23+ project in `B:\workgit\layerscope`, configure the core module and dependencies (`google/go-containerregistry`, `docker/docker/client`, `spf13/cobra`, `go-chi/chi/v5`), implement the local Docker daemon client and remote OCI registry fetcher, build the layer tarball archive unpacker, and write automated tests verifying layer extraction against sample container images.
```
