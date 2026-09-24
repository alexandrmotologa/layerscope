package vfs

import (
	"os"
	"path"
	"sort"
	"time"
)

// ChangeType represents how a file changed in a particular layer compared to previous layers.
type ChangeType string

const (
	ChangeAdded     ChangeType = "added"
	ChangeModified  ChangeType = "modified"
	ChangeDeleted   ChangeType = "deleted"
	ChangeUnchanged ChangeType = "unchanged"
)

// FileNode represents an individual file or directory in the cumulative virtual filesystem.
type FileNode struct {
	Path                string               `json:"path"`
	Name                string               `json:"name"`
	Size                int64                `json:"size"`
	Mode                os.FileMode          `json:"mode"`
	ModTime             time.Time            `json:"modTime"`
	IsDir               bool                 `json:"isDir"`
	IsSymlink           bool                 `json:"isSymlink"`
	LinkTarget          string               `json:"linkTarget,omitempty"`
	Digest              string               `json:"digest,omitempty"`
	LayerIndex          int                  `json:"layerIndex"`
	ChangeType          ChangeType           `json:"changeType"`
	WastedBytes         int64                `json:"wastedBytes"`
	IsWasted            bool                 `json:"isWasted"`
	WasteReason         string               `json:"wasteReason,omitempty"`
	OverwrittenInLayers []int                `json:"overwrittenInLayers,omitempty"`
	Children            map[string]*FileNode `json:"children,omitempty"`
}

// Clone creates a shallow copy of the node with a new children map.
func (n *FileNode) Clone() *FileNode {
	clone := &FileNode{
		Path:                n.Path,
		Name:                n.Name,
		Size:                n.Size,
		Mode:                n.Mode,
		ModTime:             n.ModTime,
		IsDir:               n.IsDir,
		IsSymlink:           n.IsSymlink,
		LinkTarget:          n.LinkTarget,
		Digest:              n.Digest,
		LayerIndex:          n.LayerIndex,
		ChangeType:          n.ChangeType,
		WastedBytes:         n.WastedBytes,
		IsWasted:            n.IsWasted,
		WasteReason:         n.WasteReason,
		OverwrittenInLayers: append([]int(nil), n.OverwrittenInLayers...),
		Children:            make(map[string]*FileNode, len(n.Children)),
	}
	return clone
}

// SortedChildren returns the node's children sorted with directories first, then alphabetically.
func (n *FileNode) SortedChildren() []*FileNode {
	if len(n.Children) == 0 {
		return nil
	}

	result := make([]*FileNode, 0, len(n.Children))
	for _, child := range n.Children {
		result = append(result, child)
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].IsDir != result[j].IsDir {
			return result[i].IsDir // directories first
		}
		return result[i].Name < result[j].Name
	})

	return result
}

// TotalSubtreeSize calculates the sum of all file sizes under this node.
func (n *FileNode) TotalSubtreeSize() int64 {
	if !n.IsDir {
		return n.Size
	}
	var total int64
	for _, child := range n.Children {
		total += child.TotalSubtreeSize()
	}
	return total
}

// CleanPath normalizes a filesystem path to start with / and have no trailing slashes.
func CleanPath(p string) string {
	cleaned := path.Clean("/" + p)
	return cleaned
}
