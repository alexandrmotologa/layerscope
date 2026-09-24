package oci

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
)

// FileInspectHook is a callback invoked for each regular file during tar extraction.
// This allows scanners (secrets, SBOMs) to process file bytes in a single pass without extra disk I/O.
type FileInspectHook func(layerIndex int, file *LayerFile, reader io.Reader) error

// UnpackLayerTar unpacks a layer tar stream (uncompressed or gzip) and extracts layer files metadata.
func UnpackLayerTar(layerIndex int, r io.Reader, hook FileInspectHook) ([]*LayerFile, int64, error) {
	// Detect gzip magic bytes
	headerBytes := make([]byte, 2)
	n, err := io.ReadFull(r, headerBytes)
	if err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return nil, 0, nil
		}
		return nil, 0, fmt.Errorf("read layer header: %w", err)
	}

	multiReader := io.MultiReader(strings.NewReader(string(headerBytes[:n])), r)
	var tarReader *tar.Reader

	if n >= 2 && headerBytes[0] == 0x1f && headerBytes[1] == 0x8b {
		gzReader, err := gzip.NewReader(multiReader)
		if err != nil {
			return nil, 0, fmt.Errorf("open gzip reader: %w", err)
		}
		defer gzReader.Close()
		tarReader = tar.NewReader(gzReader)
	} else {
		tarReader = tar.NewReader(multiReader)
	}

	var files []*LayerFile
	var totalUncompressedSize int64

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, totalUncompressedSize, fmt.Errorf("read tar entry: %w", err)
		}

		rawPath := path.Clean("/" + strings.TrimPrefix(header.Name, "./"))
		dir := path.Dir(rawPath)
		base := path.Base(rawPath)

		layerFile := &LayerFile{
			Path:      rawPath,
			Size:      header.Size,
			Mode:      os.FileMode(header.Mode),
			ModTime:   header.ModTime,
			IsDir:     header.Typeflag == tar.TypeDir,
			IsSymlink: header.Typeflag == tar.TypeSymlink,
			LinkTarget: header.Linkname,
		}

		// Handle OCI Whiteout files
		if base == ".wh..wh..opq" {
			layerFile.IsOpaque = true
			layerFile.Path = dir
		} else if strings.HasPrefix(base, ".wh.") {
			layerFile.IsWhiteout = true
			targetName := strings.TrimPrefix(base, ".wh.")
			layerFile.WhiteoutTarget = path.Join(dir, targetName)
			layerFile.Path = layerFile.WhiteoutTarget
		}

		totalUncompressedSize += header.Size

		// If regular file and not a whiteout, compute hash and execute hook
		if !layerFile.IsDir && !layerFile.IsWhiteout && !layerFile.IsOpaque && header.Typeflag != tar.TypeSymlink {
			hasher := sha256.New()
			var fileReader io.Reader = io.TeeReader(tarReader, hasher)

			if hook != nil {
				// Let hook read the content stream
				if err := hook(layerIndex, layerFile, fileReader); err != nil {
					return nil, totalUncompressedSize, fmt.Errorf("file hook on %s: %w", layerFile.Path, err)
				}
			} else {
				// Discard remainder to calculate hash
				if _, err := io.Copy(io.Discard, fileReader); err != nil {
					return nil, totalUncompressedSize, fmt.Errorf("hash file %s: %w", layerFile.Path, err)
				}
			}

			layerFile.Digest = hex.EncodeToString(hasher.Sum(nil))
		}

		files = append(files, layerFile)
	}

	return files, totalUncompressedSize, nil
}
