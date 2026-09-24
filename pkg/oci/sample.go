package oci

import (
	"archive/tar"
	"bytes"
	"fmt"
	"time"
)

// GenerateSampleImage creates a synthetic multi-layer container image with realistic OS packages,
// an intermediate leaked secret (.env) that gets removed in a subsequent layer via whiteout,
// and wasted build caches.
func GenerateSampleImage(hook FileInspectHook) (*ImageAnalysis, error) {
	now := time.Now().Add(-24 * time.Hour)

	// Layer 0: Alpine base layer
	layer0Tar := new(bytes.Buffer)
	tw0 := tar.NewWriter(layer0Tar)
	writeTarDir(tw0, "/bin", 0755, now)
	writeTarFile(tw0, "/bin/sh", []byte("#!/bin/sh\necho 'Alpine Shell'\n"), 0755, now)
	writeTarDir(tw0, "/etc", 0755, now)
	writeTarFile(tw0, "/etc/os-release", []byte("NAME=\"Alpine Linux\"\nID=alpine\nVERSION_ID=3.20.0\nPRETTY_NAME=\"Alpine Linux v3.20\"\n"), 0644, now)
	writeTarDir(tw0, "/lib/apk/db", 0755, now)
	apkDbContent := `C:Q1r3w4
P:musl
V:1.2.5-r0
A:x86_64
S:624512
I:624512
T:the musl c-library
U:https://musl.libc.org/
L:MIT

C:Q1abc123
P:busybox
V:1.36.1-r28
A:x86_64
S:892300
I:892300
T:Size optimized toolbox of many common UNIX utilities
U:https://busybox.net/
L:GPL-2.0-only

C:Q1xyz789
P:zlib
V:1.3.1-r0
A:x86_64
S:112400
I:112400
T:A Massively Spiffy Yet Delicately Unobtrusive Compression Library
U:https://zlib.net/
L:Zlib
`
	writeTarFile(tw0, "/lib/apk/db/installed", []byte(apkDbContent), 0644, now)
	tw0.Close()

	// Layer 1: App code and dependency lockfile
	layer1Tar := new(bytes.Buffer)
	tw1 := tar.NewWriter(layer1Tar)
	writeTarDir(tw1, "/app", 0755, now.Add(time.Minute))
	packageJson := `{
  "name": "payments-worker",
  "version": "1.4.2",
  "dependencies": {
    "express": "^4.19.2",
    "axios": "^1.7.2",
    "lodash": "^4.17.21"
  }
}`
	writeTarFile(tw1, "/app/package.json", []byte(packageJson), 0644, now.Add(time.Minute))
	packageLock := `{
  "name": "payments-worker",
  "version": "1.4.2",
  "lockfileVersion": 3,
  "packages": {
    "": {
      "name": "payments-worker",
      "version": "1.4.2",
      "dependencies": {
        "axios": "^1.7.2",
        "express": "^4.19.2",
        "lodash": "^4.17.21"
      }
    },
    "node_modules/axios": {
      "version": "1.7.2",
      "resolved": "https://registry.npmjs.org/axios/-/axios-1.7.2.tgz"
    },
    "node_modules/express": {
      "version": "4.19.2",
      "resolved": "https://registry.npmjs.org/express/-/express-4.19.2.tgz"
    },
    "node_modules/lodash": {
      "version": "4.17.21",
      "resolved": "https://registry.npmjs.org/lodash/-/lodash-4.17.21.tgz"
    }
  }
}`
	writeTarFile(tw1, "/app/package-lock.json", []byte(packageLock), 0644, now.Add(time.Minute))
	tw1.Close()

	// Layer 2: Intermediate secret committed by mistake!
	layer2Tar := new(bytes.Buffer)
	tw2 := tar.NewWriter(layer2Tar)
	envSecrets := `NODE_ENV=production
PORT=8080
DATABASE_URL=postgres://app_user:SuperSecretPassword123!@db.internal:5432/payments
AWS_ACCESS_KEY_ID=AKIAIOSFODNN7EXAMPLE
AWS_SECRET_ACCESS_KEY=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
GITHUB_TOKEN=ghp_ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789
`
	writeTarFile(tw2, "/app/.env", []byte(envSecrets), 0600, now.Add(2*time.Minute))
	tw2.Close()

	// Layer 3: Build step creating production bundle
	layer3Tar := new(bytes.Buffer)
	tw3 := tar.NewWriter(layer3Tar)
	writeTarDir(tw3, "/app/dist", 0755, now.Add(3*time.Minute))
	bundleData := make([]byte, 250000) // 250 KB bundle
	for i := range bundleData {
		bundleData[i] = byte(i % 256)
	}
	writeTarFile(tw3, "/app/dist/server.js", bundleData, 0644, now.Add(3*time.Minute))
	// Uncleaned build cache left in /root/.npm
	writeTarDir(tw3, "/root/.npm", 0700, now.Add(3*time.Minute))
	npmCacheData := make([]byte, 180000) // 180 KB wasted npm cache
	writeTarFile(tw3, "/root/.npm/cache-index.bin", npmCacheData, 0600, now.Add(3*time.Minute))
	tw3.Close()

	// Layer 4: RUN rm /app/.env (secret deleted from filesystem via whiteout .wh..env!)
	layer4Tar := new(bytes.Buffer)
	tw4 := tar.NewWriter(layer4Tar)
	// Whiteout entry for /app/.env is named /app/.wh..env
	writeTarFile(tw4, "/app/.wh..env", []byte{}, 0644, now.Add(4*time.Minute))
	tw4.Close()

	// Layer 5: App configuration override and start script
	layer5Tar := new(bytes.Buffer)
	tw5 := tar.NewWriter(layer5Tar)
	writeTarFile(tw5, "/app/config.json", []byte("{\"telemetry\": false, \"log_level\": \"info\"}\n"), 0644, now.Add(5*time.Minute))
	writeTarFile(tw5, "/app/entrypoint.sh", []byte("#!/bin/sh\nexec node /app/dist/server.js\n"), 0755, now.Add(5*time.Minute))
	tw5.Close()

	layerBuffers := []*bytes.Buffer{layer0Tar, layer1Tar, layer2Tar, layer3Tar, layer4Tar, layer5Tar}
	commands := []string{
		"FROM alpine:3.20",
		"WORKDIR /app && COPY package*.json ./",
		"COPY .env /app/.env",
		"RUN npm run build && npm cache clean --force",
		"RUN rm /app/.env",
		"COPY config.json ./ && CMD [\"sh\", \"/app/entrypoint.sh\"]",
	}

	var layers []*Layer
	var totalSize int64
	var totalFiles int

	for i, buf := range layerBuffers {
		files, uncompressedSize, err := UnpackLayerTar(i, buf, hook)
		if err != nil {
			return nil, fmt.Errorf("unpack sample layer %d: %w", i, err)
		}

		layer := &Layer{
			Index:       i,
			Digest:      fmt.Sprintf("sha256:mocklayer%02d%s", i, "abcdef1234567890"),
			DiffID:      fmt.Sprintf("sha256:diffid%02d%s", i, "abcdef1234567890"),
			Size:        uncompressedSize,
			Command:     commands[i],
			CreatedBy:   commands[i],
			Files:       files,
			FileCount:   len(files),
			WastedBytes: 0,
		}

		layers = append(layers, layer)
		totalSize += uncompressedSize
		totalFiles += len(files)
	}

	config := ImageConfig{
		Architecture: "amd64",
		OS:           "linux",
		Created:      now.Add(5 * time.Minute),
		Author:       "LayerScope Demo Generator",
		Config: ContainerConfig{
			User:       "0", // root user flag
			Cmd:        []string{"sh", "/app/entrypoint.sh"},
			WorkingDir: "/app",
			Env: []string{
				"PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
				"NODE_ENV=production",
			},
		},
	}

	for i, cmd := range commands {
		config.History = append(config.History, HistoryEntry{
			Created:    now.Add(time.Duration(i) * time.Minute),
			CreatedBy:  cmd,
			EmptyLayer: false,
		})
	}

	return &ImageAnalysis{
		Reference: ImageReference{
			Original:     "layerscope-demo/payments-service:v1.4.2",
			Registry:     "local",
			Repository:   "layerscope-demo/payments-service",
			Tag:          "v1.4.2",
			Digest:       "sha256:8f4c2e6d9a1b0c3e7f5a8d2e4b6c9a1b0c3e7f5a8d2e4b6c9a1b0c3e7f5a8d2e",
			Architecture: "linux/amd64",
			OS:           "linux",
			Source:       SourceSample,
		},
		Config:          config,
		Layers:          layers,
		TotalSizeBytes:  totalSize,
		TotalFilesCount: totalFiles,
		AnalyzedAt:      time.Now(),
	}, nil
}

func writeTarDir(tw *tar.Writer, name string, mode int64, modTime time.Time) {
	_ = tw.WriteHeader(&tar.Header{
		Name:     name,
		Mode:     mode,
		Typeflag: tar.TypeDir,
		ModTime:  modTime,
	})
}

func writeTarFile(tw *tar.Writer, name string, content []byte, mode int64, modTime time.Time) {
	_ = tw.WriteHeader(&tar.Header{
		Name:     name,
		Size:     int64(len(content)),
		Mode:     mode,
		Typeflag: tar.TypeReg,
		ModTime:  modTime,
	})
	_, _ = tw.Write(content)
}
