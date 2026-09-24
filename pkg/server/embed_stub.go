//go:build !embed_ui

package server

import "net/http"

// GetEmbeddedUIFileSystem returns nil when built without the embed_ui build tag.
func GetEmbeddedUIFileSystem() http.FileSystem {
	return nil
}
