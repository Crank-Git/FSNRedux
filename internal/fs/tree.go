package fs

import (
	"sort"
	"time"
)

// Tree is the result of a complete filesystem scan.
type Tree struct {
	Root      *Entry
	ScannedAt time.Time
	TotalSize int64
	FileCount int
	DirCount  int
	MaxDepth  int
	Errors    []ScanError
}

// ScanError records an error encountered during scanning.
type ScanError struct {
	Path    string
	Message string
}

// buildTree computes aggregate statistics on a scanned root entry.
func buildTree(root *Entry) *Tree {
	tree := &Tree{
		Root:      root,
		ScannedAt: time.Now(),
	}
	tree.aggregate(root)
	tree.TotalSize = root.Size
	return tree
}

// aggregate recursively computes sizes and stats.
func (t *Tree) aggregate(entry *Entry) {
	if entry.Type == TypeDir {
		t.DirCount++
		var totalSize int64
		for _, child := range entry.Children {
			t.aggregate(child)
			totalSize += child.Size
		}
		entry.Size = totalSize

		// Sort children by size descending (for layout algorithms)
		sort.Slice(entry.Children, func(i, j int) bool {
			return entry.Children[i].Size > entry.Children[j].Size
		})

		// Track max depth
		if entry.Depth > t.MaxDepth {
			t.MaxDepth = entry.Depth
		}
	} else {
		t.FileCount++
		if entry.Depth > t.MaxDepth {
			t.MaxDepth = entry.Depth
		}
	}

	if entry.Error != "" {
		t.Errors = append(t.Errors, ScanError{
			Path:    entry.Path,
			Message: entry.Error,
		})
	}
}
