package vfs

import (
	"testing"

	"github.com/alexandrmotologa/layerscope/pkg/oci"
)

func TestVFSTree_BasicOperations(t *testing.T) {
	tree := NewVFSTree()

	node := &FileNode{
		Path:  "/etc/nginx/nginx.conf",
		Size:  1024,
		IsDir: false,
	}
	tree.Insert(node)

	found := tree.Lookup("/etc/nginx/nginx.conf")
	if found == nil {
		t.Fatalf("expected node at /etc/nginx/nginx.conf")
	}
	if found.Size != 1024 {
		t.Errorf("expected size 1024, got %d", found.Size)
	}

	// Parent directories should have been created
	parent := tree.Lookup("/etc/nginx")
	if parent == nil || !parent.IsDir {
		t.Errorf("expected parent directory /etc/nginx to exist")
	}

	// Delete file
	tree.Delete("/etc/nginx/nginx.conf")
	if tree.Lookup("/etc/nginx/nginx.conf") != nil {
		t.Errorf("expected /etc/nginx/nginx.conf to be deleted")
	}
}

func TestVFSTree_OpaqueDirectory(t *testing.T) {
	tree := NewVFSTree()
	tree.Insert(&FileNode{Path: "/app/data/file1.txt", Size: 100})
	tree.Insert(&FileNode{Path: "/app/data/file2.txt", Size: 200})

	if tree.Lookup("/app/data/file1.txt") == nil {
		t.Fatalf("expected file1.txt to exist")
	}

	tree.DeleteChildren("/app/data")
	if tree.Lookup("/app/data/file1.txt") != nil || tree.Lookup("/app/data/file2.txt") != nil {
		t.Errorf("expected children of /app/data to be deleted")
	}
	if tree.Lookup("/app/data") == nil {
		t.Errorf("expected /app/data directory itself to remain")
	}
}

func TestBuildLayerSnapshots_WithSample(t *testing.T) {
	sampleImg, err := oci.GenerateSampleImage(nil)
	if err != nil {
		t.Fatalf("GenerateSampleImage failed: %v", err)
	}

	snapshots := BuildLayerSnapshots(sampleImg)
	if len(snapshots) != 6 {
		t.Fatalf("expected 6 snapshots, got %d", len(snapshots))
	}

	// Layer 2 should have /app/.env
	snap2 := snapshots[2]
	if snap2.Tree.Lookup("/app/.env") == nil {
		t.Errorf("expected /app/.env to exist in layer 2 snapshot")
	}

	// Layer 4 deletes /app/.env via whiteout
	snap4 := snapshots[4]
	if snap4.Tree.Lookup("/app/.env") != nil {
		t.Errorf("expected /app/.env to be deleted in layer 4 snapshot")
	}

	// Final snapshot (layer 5)
	finalSnap := snapshots[5]
	if finalSnap.Tree.Lookup("/app/.env") != nil {
		t.Errorf("expected /app/.env to remain deleted in final layer")
	}
	if finalSnap.Tree.Lookup("/app/entrypoint.sh") == nil {
		t.Errorf("expected /app/entrypoint.sh to exist in final snapshot")
	}

	// Check wasted space
	summary := CalculateWastedSpace(snapshots, nil)
	if summary.TotalWastedBytes <= 0 {
		t.Errorf("expected positive wasted bytes from deleted .env / overwritten files, got %d", summary.TotalWastedBytes)
	}
}

func TestDiffVFSTrees(t *testing.T) {
	treeA := NewVFSTree()
	treeA.Insert(&FileNode{Path: "/app/main.js", Size: 100, Digest: "hash1"})
	treeA.Insert(&FileNode{Path: "/app/old.js", Size: 50, Digest: "hash2"})

	treeB := NewVFSTree()
	treeB.Insert(&FileNode{Path: "/app/main.js", Size: 120, Digest: "hash3"}) // modified
	treeB.Insert(&FileNode{Path: "/app/new.js", Size: 80, Digest: "hash4"})  // added
	// old.js removed

	diff := DiffVFSTrees("v1", treeA, 150, "v2", treeB, 200)

	if diff.AddedCount != 1 {
		t.Errorf("expected 1 added file, got %d", diff.AddedCount)
	}
	if diff.RemovedCount != 1 {
		t.Errorf("expected 1 removed file, got %d", diff.RemovedCount)
	}
	if diff.ModifiedCount != 1 {
		t.Errorf("expected 1 modified file, got %d", diff.ModifiedCount)
	}
	if diff.TotalSizeDelta != 50 {
		t.Errorf("expected total size delta 50, got %d", diff.TotalSizeDelta)
	}
}
