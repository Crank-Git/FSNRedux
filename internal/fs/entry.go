package fs

import (
	"os"
	"path/filepath"
	"time"
)

// EntryType distinguishes files from directories and special nodes.
type EntryType uint8

const (
	TypeFile    EntryType = iota
	TypeDir
	TypeSymlink
	TypeOther
)

// String returns a human-readable name for the entry type.
func (t EntryType) String() string {
	switch t {
	case TypeFile:
		return "file"
	case TypeDir:
		return "directory"
	case TypeSymlink:
		return "symlink"
	default:
		return "other"
	}
}

// Entry is an immutable node in the scanned filesystem tree.
type Entry struct {
	Name     string
	Path     string    // absolute path
	Type     EntryType
	Size     int64     // for files: file size; for dirs: recursive sum
	ModTime  time.Time // last modification time
	Children []*Entry  // nil for files; sorted by Size descending for layout
	Depth    int       // distance from scan root
	Error    string    // non-empty if this entry had a scan error
	Loaded   bool      // true if this dir's children have been scanned
}

// IsDir returns true if this entry is a directory.
func (e *Entry) IsDir() bool {
	return e.Type == TypeDir
}

// FileCount returns the total number of files in this subtree (recursive).
func (e *Entry) FileCount() int {
	if !e.IsDir() {
		return 1
	}
	count := 0
	for _, child := range e.Children {
		count += child.FileCount()
	}
	return count
}

// DirCount returns the total number of directories in this subtree (recursive, including self).
func (e *Entry) DirCount() int {
	if !e.IsDir() {
		return 0
	}
	count := 1
	for _, child := range e.Children {
		count += child.DirCount()
	}
	return count
}

// InspectInfo holds detailed metadata gathered on-demand when the user inspects a node.
type InspectInfo struct {
	Name       string
	Path       string
	TypeStr    string
	Extension  string
	Size       int64
	Perms      string // e.g. "-rwxr-xr-x"
	ModTime    time.Time
	IsDir      bool
	FileCount  int
	DirCount   int
	ChildCount int // direct children count
	Loaded     bool
}

// Inspect gathers detailed info about this entry from the filesystem.
func (e *Entry) Inspect() InspectInfo {
	info := InspectInfo{
		Name:    e.Name,
		Path:    e.Path,
		TypeStr: e.Type.String(),
		Size:    e.Size,
		ModTime: e.ModTime,
		IsDir:   e.IsDir(),
		Loaded:  e.Loaded,
	}

	// Get permissions from filesystem
	if stat, err := os.Lstat(e.Path); err == nil {
		info.Perms = stat.Mode().Perm().String()
	}

	if e.IsDir() {
		info.ChildCount = len(e.Children)
		if e.Loaded {
			info.FileCount = e.FileCount()
			info.DirCount = e.DirCount() - 1 // exclude self
		}
	} else {
		info.Extension = filepath.Ext(e.Name)
	}

	return info
}
