//go:build embed_ui

package server

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed dist/*
var embeddedDist embed.FS

// GetEmbeddedUIFileSystem returns the embedded Vite production build assets.
func GetEmbeddedUIFileSystem() http.FileSystem {
	sub, err := fs.Sub(embeddedDist, "dist")
	if err != nil {
		return nil
	}
	return http.FS(sub)
}
