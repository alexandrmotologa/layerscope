package vfs

import (
	"path"

	"github.com/alexandrmotologa/layerscope/pkg/oci"
)

// LayerSnapshot captures the full virtual filesystem state after applying a specific layer.
type LayerSnapshot struct {
	LayerIndex     int         `json:"layerIndex"`
	Command        string      `json:"command"`
	Digest         string      `json:"digest"`
	Size           int64       `json:"size"`
	WastedBytes    int64       `json:"wastedBytes"`
	AddedCount     int         `json:"addedCount"`
	ModifiedCount  int         `json:"modifiedCount"`
	DeletedCount   int         `json:"deletedCount"`
	UnchangedCount int         `json:"unchangedCount"`
	Tree           *VFSTree    `json:"tree"`
	DeltaFiles     []*FileNode `json:"deltaFiles"`
}

// BuildLayerSnapshots computes cumulative filesystem trees and delta changes for all layers.
func BuildLayerSnapshots(img *oci.ImageAnalysis) []*LayerSnapshot {
	snapshots := make([]*LayerSnapshot, len(img.Layers))

	// Track file history across layers: path -> slice of layer indices where it was written
	type fileHistory struct {
		layerIndex int
		size       int64
		digest     string
	}
	historyMap := make(map[string][]fileHistory)

	var currentTree = NewVFSTree()

	for i, layer := range img.Layers {
		nextTree := currentTree.Clone()

		// Reset change types from previous layer to unchanged in the active tree
		for _, node := range nextTree.Index {
			node.ChangeType = ChangeUnchanged
		}

		var deltaFiles []*FileNode
		var addedCount, modifiedCount, deletedCount int

		// First pass: handle whiteouts
		for _, f := range layer.Files {
			if f.IsOpaque {
				nextTree.DeleteChildren(f.Path)
				deltaFiles = append(deltaFiles, &FileNode{
					Path:        f.Path,
					Name:        path.Base(f.Path),
					IsDir:       true,
					LayerIndex:  i,
					ChangeType:  ChangeModified,
					WasteReason: "Opaque directory marker cleared previous contents",
				})
			} else if f.IsWhiteout {
				target := f.WhiteoutTarget
				if existing := nextTree.Lookup(target); existing != nil {
					deletedNode := existing.Clone()
					deletedNode.ChangeType = ChangeDeleted
					deletedNode.LayerIndex = i
					deltaFiles = append(deltaFiles, deletedNode)
					deletedCount++
					nextTree.Delete(target)
				}
			}
		}

		// Second pass: handle regular files and directories
		for _, f := range layer.Files {
			if f.IsOpaque || f.IsWhiteout {
				continue
			}

			existing := nextTree.Lookup(f.Path)
			changeType := ChangeAdded

			if existing != nil && !existing.IsDir && !f.IsDir {
				changeType = ChangeModified
				modifiedCount++
			} else if existing == nil {
				addedCount++
			}

			node := &FileNode{
				Path:       f.Path,
				Name:       path.Base(f.Path),
				Size:       f.Size,
				Mode:       f.Mode,
				ModTime:    f.ModTime,
				IsDir:      f.IsDir,
				IsSymlink:  f.IsSymlink,
				LinkTarget: f.LinkTarget,
				Digest:     f.Digest,
				Data:       f.Data,
				LayerIndex: i,
				ChangeType: changeType,
			}

			nextTree.Insert(node)
			deltaFiles = append(deltaFiles, node)

			if !f.IsDir {
				historyMap[f.Path] = append(historyMap[f.Path], fileHistory{
					layerIndex: i,
					size:       f.Size,
					digest:     f.Digest,
				})
			}
		}

		// Count unchanged files
		unchangedCount := len(nextTree.Index) - (addedCount + modifiedCount)
		if unchangedCount < 0 {
			unchangedCount = 0
		}

		snapshot := &LayerSnapshot{
			LayerIndex:     i,
			Command:        layer.Command,
			Digest:         layer.Digest,
			Size:           layer.Size,
			AddedCount:     addedCount,
			ModifiedCount:  modifiedCount,
			DeletedCount:   deletedCount,
			UnchangedCount: unchangedCount,
			Tree:           nextTree,
			DeltaFiles:     deltaFiles,
		}

		snapshots[i] = snapshot
		currentTree = nextTree
	}

	// Compute wasted space across layers
	CalculateWastedSpace(snapshots, historyMap)

	return snapshots
}
