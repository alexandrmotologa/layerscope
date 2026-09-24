package vfs

import (
	"fmt"
	"sort"
)

// WastedSpaceSummary aggregates all wasted bytes and inefficiencies found in an image.
type WastedSpaceSummary struct {
	TotalWastedBytes   int64       `json:"totalWastedBytes"`
	TotalImageBytes    int64       `json:"totalImageBytes"`
	WastedPercentage   float64     `json:"wastedPercentage"`
	OverwrittenFiles   int         `json:"overwrittenFiles"`
	DeletedFiles       int         `json:"deletedFiles"`
	DuplicateFiles     int         `json:"duplicateFiles"`
	TopWastedFiles     []*FileNode `json:"topWastedFiles"`
}

type fileOccurrence struct {
	layerIndex int
	size       int64
	digest     string
}

// CalculateWastedSpace analyzes file history across layers and tags wasted nodes.
func CalculateWastedSpace(snapshots []*LayerSnapshot, historyMap interface{}) *WastedSpaceSummary {
	var totalWasted int64
	var totalImageBytes int64
	var overwrittenCount, deletedCount int
	var allWastedNodes []*FileNode

	// Calculate total image bytes from layers
	for _, snap := range snapshots {
		totalImageBytes += snap.Size
	}

	// Analyze each layer's delta files against the final tree state
	if len(snapshots) == 0 {
		return &WastedSpaceSummary{}
	}

	finalTree := snapshots[len(snapshots)-1].Tree

	for layerIdx := 0; layerIdx < len(snapshots)-1; layerIdx++ {
		snap := snapshots[layerIdx]

		for _, node := range snap.DeltaFiles {
			if node.IsDir || node.ChangeType == ChangeDeleted {
				continue
			}

			// Check where the file ends up in later layers
			var overwrittenLater bool
			var deletedLater bool
			var laterLayerIdx int

			// Check subsequent layers
			for nextIdx := layerIdx + 1; nextIdx < len(snapshots); nextIdx++ {
				nextSnap := snapshots[nextIdx]
				for _, nextDelta := range nextSnap.DeltaFiles {
					if nextDelta.Path == node.Path {
						if nextDelta.ChangeType == ChangeDeleted {
							deletedLater = true
							laterLayerIdx = nextIdx
							break
						} else if nextDelta.ChangeType == ChangeModified || nextDelta.ChangeType == ChangeAdded {
							overwrittenLater = true
							laterLayerIdx = nextIdx
							break
						}
					}
				}
				if overwrittenLater || deletedLater {
					break
				}
			}

			// Check if deleted in final tree
			finalNode := finalTree.Lookup(node.Path)
			if finalNode == nil {
				deletedLater = true
			} else if finalNode.LayerIndex > layerIdx {
				overwrittenLater = true
				laterLayerIdx = finalNode.LayerIndex
			}

			if overwrittenLater {
				node.IsWasted = true
				node.WastedBytes = node.Size
				node.WasteReason = fmt.Sprintf("Overwritten in layer %d", laterLayerIdx)
				snap.WastedBytes += node.Size
				totalWasted += node.Size
				overwrittenCount++
				allWastedNodes = append(allWastedNodes, node)
			} else if deletedLater {
				node.IsWasted = true
				node.WastedBytes = node.Size
				node.WasteReason = fmt.Sprintf("Deleted in later layer (historical layer remnant)")
				snap.WastedBytes += node.Size
				totalWasted += node.Size
				deletedCount++
				allWastedNodes = append(allWastedNodes, node)
			}
		}
	}

	// Sort wasted nodes descending by size
	sort.Slice(allWastedNodes, func(i, j int) bool {
		return allWastedNodes[i].WastedBytes > allWastedNodes[j].WastedBytes
	})

	topCount := 50
	if len(allWastedNodes) < topCount {
		topCount = len(allWastedNodes)
	}

	var wastedPct float64
	if totalImageBytes > 0 {
		wastedPct = (float64(totalWasted) / float64(totalImageBytes)) * 100.0
	}

	return &WastedSpaceSummary{
		TotalWastedBytes:   totalWasted,
		TotalImageBytes:    totalImageBytes,
		WastedPercentage:   wastedPct,
		OverwrittenFiles:   overwrittenCount,
		DeletedFiles:       deletedCount,
		TopWastedFiles:     allWastedNodes[:topCount],
	}
}
