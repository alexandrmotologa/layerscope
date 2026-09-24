package vfs

import (
	"sort"
)

// DiffChange represents the comparison result for a single file between two images.
type DiffChange string

const (
	DiffAdded    DiffChange = "added"    // present in B, missing in A
	DiffRemoved  DiffChange = "removed"  // present in A, missing in B
	DiffModified DiffChange = "modified" // present in both, but size, digest, or permissions differ
	DiffSame     DiffChange = "same"     // identical
)

// FileDiff describes how an individual file differs between Image A and Image B.
type FileDiff struct {
	Path       string     `json:"path"`
	Status     DiffChange `json:"status"`
	SizeA      int64      `json:"sizeA"`
	SizeB      int64      `json:"sizeB"`
	SizeDelta  int64      `json:"sizeDelta"` // SizeB - SizeA
	DigestA    string     `json:"digestA,omitempty"`
	DigestB    string     `json:"digestB,omitempty"`
	ModeA      string     `json:"modeA,omitempty"`
	ModeB      string     `json:"modeB,omitempty"`
	IsDir      bool       `json:"isDir"`
}

// ImageDiffReport holds the complete side-by-side comparison between two container images.
type ImageDiffReport struct {
	ImageAName     string      `json:"imageAName"`
	ImageBName     string      `json:"imageBName"`
	TotalSizeA     int64       `json:"totalSizeA"`
	TotalSizeB     int64       `json:"totalSizeB"`
	TotalSizeDelta int64       `json:"totalSizeDelta"`
	AddedCount     int         `json:"addedCount"`
	RemovedCount   int         `json:"removedCount"`
	ModifiedCount  int         `json:"modifiedCount"`
	SameCount      int         `json:"sameCount"`
	Files          []*FileDiff `json:"files"`
}

// DiffVFSTrees compares the resolved filesystems of two images.
func DiffVFSTrees(nameA string, treeA *VFSTree, totalSizeA int64, nameB string, treeB *VFSTree, totalSizeB int64) *ImageDiffReport {
	allPaths := make(map[string]bool)

	for p := range treeA.Index {
		allPaths[p] = true
	}
	for p := range treeB.Index {
		allPaths[p] = true
	}

	report := &ImageDiffReport{
		ImageAName:     nameA,
		ImageBName:     nameB,
		TotalSizeA:     totalSizeA,
		TotalSizeB:     totalSizeB,
		TotalSizeDelta: totalSizeB - totalSizeA,
	}

	for p := range allPaths {
		if p == "/" {
			continue
		}

		nodeA := treeA.Lookup(p)
		nodeB := treeB.Lookup(p)

		if nodeA == nil && nodeB != nil {
			report.AddedCount++
			report.Files = append(report.Files, &FileDiff{
				Path:      p,
				Status:    DiffAdded,
				SizeB:     nodeB.Size,
				SizeDelta: nodeB.Size,
				DigestB:   nodeB.Digest,
				ModeB:     nodeB.Mode.String(),
				IsDir:     nodeB.IsDir,
			})
		} else if nodeA != nil && nodeB == nil {
			report.RemovedCount++
			report.Files = append(report.Files, &FileDiff{
				Path:      p,
				Status:    DiffRemoved,
				SizeA:     nodeA.Size,
				SizeDelta: -nodeA.Size,
				DigestA:   nodeA.Digest,
				ModeA:     nodeA.Mode.String(),
				IsDir:     nodeA.IsDir,
			})
		} else if nodeA != nil && nodeB != nil {
			isSame := nodeA.Size == nodeB.Size && nodeA.IsDir == nodeB.IsDir
			if nodeA.Digest != "" && nodeB.Digest != "" {
				isSame = isSame && (nodeA.Digest == nodeB.Digest)
			}

			if isSame {
				report.SameCount++
			} else {
				report.ModifiedCount++
				report.Files = append(report.Files, &FileDiff{
					Path:      p,
					Status:    DiffModified,
					SizeA:     nodeA.Size,
					SizeB:     nodeB.Size,
					SizeDelta: nodeB.Size - nodeA.Size,
					DigestA:   nodeA.Digest,
					DigestB:   nodeB.Digest,
					ModeA:     nodeA.Mode.String(),
					ModeB:     nodeB.Mode.String(),
					IsDir:     nodeA.IsDir,
				})
			}
		}
	}

	// Sort diffs: modified/added/removed first, sorted by absolute delta size descending
	sort.Slice(report.Files, func(i, j int) bool {
		absI := report.Files[i].SizeDelta
		if absI < 0 {
			absI = -absI
		}
		absJ := report.Files[j].SizeDelta
		if absJ < 0 {
			absJ = -absJ
		}
		if absI != absJ {
			return absI > absJ
		}
		return report.Files[i].Path < report.Files[j].Path
	})

	return report
}
