package fs

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestScanSync_BasicTree(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()

	// Create subdirectories
	os.MkdirAll(filepath.Join(tmpDir, "dir1", "subdir1"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "dir2"), 0755)

	// Create files with known sizes
	writeFile(t, filepath.Join(tmpDir, "root.txt"), 100)
	writeFile(t, filepath.Join(tmpDir, "dir1", "file1.txt"), 200)
	writeFile(t, filepath.Join(tmpDir, "dir1", "subdir1", "deep.txt"), 300)
	writeFile(t, filepath.Join(tmpDir, "dir2", "file2.txt"), 400)

	scanner := NewScanner(ScannerOptions{})
	tree, err := scanner.ScanSync(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("ScanSync failed: %v", err)
	}

	// Verify tree stats
	if tree.FileCount != 4 {
		t.Errorf("expected 4 files, got %d", tree.FileCount)
	}
	if tree.DirCount != 4 { // root + dir1 + subdir1 + dir2
		t.Errorf("expected 4 dirs, got %d", tree.DirCount)
	}
	if tree.TotalSize != 1000 {
		t.Errorf("expected total size 1000, got %d", tree.TotalSize)
	}
	if tree.MaxDepth != 3 {
		t.Errorf("expected max depth 3, got %d", tree.MaxDepth)
	}
}

func TestScanSync_EmptyDir(t *testing.T) {
	tmpDir := t.TempDir()

	scanner := NewScanner(ScannerOptions{})
	tree, err := scanner.ScanSync(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("ScanSync failed: %v", err)
	}

	if tree.FileCount != 0 {
		t.Errorf("expected 0 files, got %d", tree.FileCount)
	}
	if tree.Root.Size != 0 {
		t.Errorf("expected root size 0, got %d", tree.Root.Size)
	}
}

func TestScanSync_IgnorePatterns(t *testing.T) {
	tmpDir := t.TempDir()

	os.MkdirAll(filepath.Join(tmpDir, ".git", "objects"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "node_modules", "pkg"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "src"), 0755)

	writeFile(t, filepath.Join(tmpDir, ".git", "HEAD"), 50)
	writeFile(t, filepath.Join(tmpDir, "node_modules", "pkg", "index.js"), 100)
	writeFile(t, filepath.Join(tmpDir, "src", "main.go"), 200)

	scanner := NewScanner(ScannerOptions{})
	tree, err := scanner.ScanSync(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("ScanSync failed: %v", err)
	}

	// Only src/main.go should be counted (not .git or node_modules contents)
	if tree.FileCount != 1 {
		t.Errorf("expected 1 file (ignoring .git and node_modules), got %d", tree.FileCount)
	}
}

func TestScanSync_Cancellation(t *testing.T) {
	tmpDir := t.TempDir()

	// Create many directories to make scan take some time
	for i := 0; i < 50; i++ {
		dir := filepath.Join(tmpDir, fmt.Sprintf("dir%d", i))
		os.MkdirAll(dir, 0755)
		for j := 0; j < 10; j++ {
			writeFile(t, filepath.Join(dir, fmt.Sprintf("file%d.txt", j)), 100)
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	scanner := NewScanner(ScannerOptions{})
	tree, err := scanner.ScanSync(ctx, tmpDir)
	if err != nil {
		t.Fatalf("ScanSync failed: %v", err)
	}

	// With immediate cancellation, we should have very few files
	// (might get some due to race, but should be far less than 500)
	if tree.FileCount >= 500 {
		t.Errorf("expected cancellation to limit scanning, but got all %d files", tree.FileCount)
	}
}

func TestScanSync_MaxDepth(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a deep tree
	deep := tmpDir
	for i := 0; i < 10; i++ {
		deep = filepath.Join(deep, fmt.Sprintf("level%d", i))
		os.MkdirAll(deep, 0755)
		writeFile(t, filepath.Join(deep, "file.txt"), 100)
	}

	scanner := NewScanner(ScannerOptions{MaxDepth: 3})
	tree, err := scanner.ScanSync(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("ScanSync failed: %v", err)
	}

	// Should only have files up to depth 3
	if tree.MaxDepth > 3 {
		t.Errorf("expected max depth <= 3, got %d", tree.MaxDepth)
	}
}

func TestScanSync_Progress(t *testing.T) {
	tmpDir := t.TempDir()

	os.MkdirAll(filepath.Join(tmpDir, "a"), 0755)
	writeFile(t, filepath.Join(tmpDir, "a", "f.txt"), 500)

	scanner := NewScanner(ScannerOptions{})
	_, err := scanner.ScanSync(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("ScanSync failed: %v", err)
	}

	progress := scanner.Progress()
	if progress.FilesFound < 1 {
		t.Errorf("expected at least 1 file in progress, got %d", progress.FilesFound)
	}
	if progress.BytesTotal < 500 {
		t.Errorf("expected at least 500 bytes in progress, got %d", progress.BytesTotal)
	}
}

func TestScanSync_InvalidPath(t *testing.T) {
	scanner := NewScanner(ScannerOptions{})
	_, err := scanner.ScanSync(context.Background(), "/nonexistent/path/12345")
	if err == nil {
		t.Error("expected error for nonexistent path")
	}
}

func TestScanSync_ChildrenSortedBySize(t *testing.T) {
	tmpDir := t.TempDir()

	writeFile(t, filepath.Join(tmpDir, "small.txt"), 10)
	writeFile(t, filepath.Join(tmpDir, "medium.txt"), 100)
	writeFile(t, filepath.Join(tmpDir, "large.txt"), 1000)

	scanner := NewScanner(ScannerOptions{})
	tree, err := scanner.ScanSync(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("ScanSync failed: %v", err)
	}

	children := tree.Root.Children
	if len(children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(children))
	}

	// Should be sorted descending by size
	for i := 1; i < len(children); i++ {
		if children[i].Size > children[i-1].Size {
			t.Errorf("children not sorted by size: %s(%d) > %s(%d)",
				children[i].Name, children[i].Size,
				children[i-1].Name, children[i-1].Size)
		}
	}
}

func TestEntryFileCount(t *testing.T) {
	entry := &Entry{
		Type: TypeDir,
		Children: []*Entry{
			{Type: TypeFile},
			{Type: TypeFile},
			{Type: TypeDir, Children: []*Entry{
				{Type: TypeFile},
			}},
		},
	}

	if got := entry.FileCount(); got != 3 {
		t.Errorf("expected 3 files, got %d", got)
	}
}

func TestEntryDirCount(t *testing.T) {
	entry := &Entry{
		Type: TypeDir,
		Children: []*Entry{
			{Type: TypeFile},
			{Type: TypeDir, Children: []*Entry{
				{Type: TypeDir},
			}},
		},
	}

	if got := entry.DirCount(); got != 3 {
		t.Errorf("expected 3 dirs, got %d", got)
	}
}

// writeFile creates a file with exactly the specified size.
func writeFile(t *testing.T, path string, size int) {
	t.Helper()
	data := make([]byte, size)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatalf("failed to write %s: %v", path, err)
	}
}
