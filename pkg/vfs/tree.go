package vfs

import (
	"path"
	"strings"
	"time"
)

// VFSTree represents a complete virtual filesystem directory hierarchy.
type VFSTree struct {
	Root  *FileNode            `json:"root"`
	Index map[string]*FileNode `json:"-"`
}

// NewVFSTree initializes an empty virtual filesystem with a root directory.
func NewVFSTree() *VFSTree {
	root := &FileNode{
		Path:       "/",
		Name:       "/",
		IsDir:      true,
		ChangeType: ChangeUnchanged,
		Children:   make(map[string]*FileNode),
		ModTime:    time.Now(),
	}
	t := &VFSTree{
		Root:  root,
		Index: make(map[string]*FileNode),
	}
	t.Index["/"] = root
	return t
}

// Lookup finds a node by its exact path.
func (t *VFSTree) Lookup(p string) *FileNode {
	clean := CleanPath(p)
	return t.Index[clean]
}

// Insert adds a node to the tree, automatically creating missing intermediate directories.
func (t *VFSTree) Insert(node *FileNode) {
	node.Path = CleanPath(node.Path)
	node.Name = path.Base(node.Path)
	if node.Children == nil && node.IsDir {
		node.Children = make(map[string]*FileNode)
	}

	if node.Path == "/" {
		t.Root = node
		t.Index["/"] = node
		return
	}

	parentPath := path.Dir(node.Path)
	parentNode := t.ensureDir(parentPath, node.LayerIndex)

	parentNode.Children[node.Name] = node
	t.Index[node.Path] = node

	// If inserting a directory, also index any existing children if present
	if node.IsDir {
		for _, child := range node.Children {
			t.Index[child.Path] = child
		}
	}
}

// ensureDir traverses down and creates any missing parent directory nodes.
func (t *VFSTree) ensureDir(dirPath string, layerIndex int) *FileNode {
	dirPath = CleanPath(dirPath)
	if node, exists := t.Index[dirPath]; exists {
		return node
	}

	parentPath := path.Dir(dirPath)
	parentNode := t.ensureDir(parentPath, layerIndex)

	name := path.Base(dirPath)
	dirNode := &FileNode{
		Path:       dirPath,
		Name:       name,
		IsDir:      true,
		Mode:       0755,
		LayerIndex: layerIndex,
		ChangeType: ChangeUnchanged,
		Children:   make(map[string]*FileNode),
		ModTime:    time.Now(),
	}

	parentNode.Children[name] = dirNode
	t.Index[dirPath] = dirNode
	return dirNode
}

// Delete removes a node and all of its descendants from the tree.
func (t *VFSTree) Delete(p string) {
	clean := CleanPath(p)
	if clean == "/" {
		// Cannot delete root, just clear children
		t.Root.Children = make(map[string]*FileNode)
		t.Index = map[string]*FileNode{"/": t.Root}
		return
	}

	node, exists := t.Index[clean]
	if !exists {
		return
	}

	// Remove from parent
	parentPath := path.Dir(clean)
	if parent, pExists := t.Index[parentPath]; pExists {
		delete(parent.Children, node.Name)
	}

	// Recursively delete from index
	t.removeFromIndex(node)
}

// DeleteChildren removes all children inside an opaque directory.
func (t *VFSTree) DeleteChildren(dirPath string) {
	clean := CleanPath(dirPath)
	dirNode, exists := t.Index[clean]
	if !exists || !dirNode.IsDir {
		return
	}

	for _, child := range dirNode.Children {
		t.removeFromIndex(child)
	}
	dirNode.Children = make(map[string]*FileNode)
}

func (t *VFSTree) removeFromIndex(node *FileNode) {
	delete(t.Index, node.Path)
	for _, child := range node.Children {
		t.removeFromIndex(child)
	}
}

// Clone creates a deep copy of the tree structure.
func (t *VFSTree) Clone() *VFSTree {
	newTree := &VFSTree{
		Index: make(map[string]*FileNode, len(t.Index)),
	}

	newTree.Root = t.cloneSubtree(t.Root, newTree.Index)
	return newTree
}

func (t *VFSTree) cloneSubtree(current *FileNode, index map[string]*FileNode) *FileNode {
	newNode := current.Clone()
	index[newNode.Path] = newNode

	for name, child := range current.Children {
		clonedChild := t.cloneSubtree(child, index)
		newNode.Children[name] = clonedChild
	}

	return newNode
}

// Flatten returns all nodes matching an optional prefix.
func (t *VFSTree) Flatten(prefix string) []*FileNode {
	var list []*FileNode
	prefix = CleanPath(prefix)

	for p, node := range t.Index {
		if prefix == "/" || strings.HasPrefix(p, prefix) {
			list = append(list, node)
		}
	}
	return list
}
