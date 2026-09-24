# Architecture Overview

LayerScope inspects container images down to the layer level, tracks file modifications across cumulative states, scans intermediate layers for secrets, and exports package inventories in standard SBOM formats.

## System Components

```
+-------------------------------------------------------------+
|                       LayerScope CLI                        |
|  (analyze, diff, audit, tui, serve subcommands via Cobra)   |
+-------------------------------------------------------------+
                               |
        +----------------------+----------------------+
        |                                             |
        v                                             v
+-------------------------+               +-------------------------+
|     Terminal TUI        |               |   Embedded Web Studio   |
| (Bubbletea dual-pane)   |               |   (React 19 + Monaco)   |
+-------------------------+               +-------------------------+
        |                                             |
        +----------------------+----------------------+
                               |
                               v
+-------------------------------------------------------------+
|                      Core Go Engine                         |
|                                                             |
|  +--------------------+  +-------------------------------+  |
|  | OCI Image Source   |  | Virtual File System (VFS)     |  |
|  | - Docker daemon    |  | - Layer accumulator           |  |
|  | - Remote registry  |  | - Whiteout resolution         |  |
|  | - Image tarball    |  | - Multi-arch diff engine      |  |
|  | - Offline fixtures |  | - Wasted space calculator     |  |
|  +--------------------+  +-------------------------------+  |
|                                                             |
|  +--------------------+  +-------------------------------+  |
|  | Security Scanner   |  | SBOM Extractor                |  |
|  | - Intermediate tar |  | - Alpine apk                  |  |
|  | - Gitleaks rules   |  | - Debian dpkg                 |  |
|  | - Secret masking   |  | - Language runtimes           |  |
|  | - Privilege checks |  | - CycloneDX / SPDX output     |  |
|  +--------------------+  +-------------------------------+  |
|                                                             |
|  +--------------------+  +-------------------------------+  |
|  | Advisor Engine     |  | Audit Exporters               |  |
|  | - Efficiency score |  | - JSON & JUnit XML            |  |
|  | - Dockerfile tips  |  | - Markdown & HTML reports     |  |
|  +--------------------+  +-------------------------------+  |
+-------------------------------------------------------------+
```

## Image Acquisition

LayerScope loads images from four distinct sources:

1. **Remote OCI Registries**: Pulls manifests and layer blobs directly via `google/go-containerregistry`. This works on machines without Docker installed and avoids writing full uncompressed layers to disk when only metadata or layer streams are required.
2. **Local Docker Daemon**: Connects through `/var/run/docker.sock` on Linux/macOS or `//./pipe/docker_engine` on Windows using Docker Engine SDK.
3. **Exported Tarballs**: Reads archives created by `docker save` or OCI image layout directories.
4. **Built-in Fixture Engine**: Generates deterministic synthetic multi-layer images for offline testing, CI verification, and immediate local evaluation.

## Virtual File System and Whiteout Resolution

OCI images represent file deletions and directory shadows through whiteout files:

- `.wh.<filename>`: Indicates that `<filename>` was deleted in this layer.
- `.wh..wh..opq`: Indicates that the directory is opaque, hiding all files from underlying layers within that directory.

The VFS Layer Accumulator iterates from base layer (index 0) to top layer (index N). At each step, it records:
- The delta of files introduced by that layer.
- The cumulative filesystem state up to that point.
- Changes classified as `Added`, `Modified`, `Deleted`, or `Unchanged`.
- Wasted bytes, such as files created in layer K and subsequently modified or removed in layer K+1.

## Intermediate Layer Security Scanner

Standard container scanners only inspect the final flattened filesystem. If a secret was copied into layer 2 and deleted in layer 4, traditional runtime tools will not report it. However, the secret remains intact in layer 2's tarball and can be extracted by any user with read access to the image repository.

LayerScope reads each layer archive stream independently during ingestion. It matches file paths and contents against regex and high-entropy patterns. When a match occurs, it attributes the leak to the specific layer index and Dockerfile command.

## Software Bill of Materials (SBOM)

LayerScope scans the resolved root filesystem for standard package manager databases:
- `/lib/apk/db/installed` for Alpine Linux.
- `/var/lib/dpkg/status` for Debian, Ubuntu, and derivatives.
- Common language dependency files (`package-lock.json`, `pnpm-lock.yaml`, `requirements.txt`, `Cargo.lock`, Go module tables in binaries).

The inventory can be exported in CycloneDX v1.5 JSON or SPDX v2.3 format for supply-chain compliance pipelines.
