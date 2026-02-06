package layout

import (
	"testing"
	"time"

	"github.com/Crank-Git/FSNRedux/internal/fs"
)

func TestComputeTreeV_NilTree(t *testing.T) {
	result := Compute(nil, DefaultOptions(ModeTreeV))
	if result != nil {
		t.Error("expected nil for nil tree")
	}
}

func TestComputeTreeV_SingleDir(t *testing.T) {
	tree := &fs.Tree{
		Root: &fs.Entry{
			Name: "root",
			Type: fs.TypeDir,
			Size: 0,
		},
		TotalSize: 0,
	}

	result := Compute(tree, DefaultOptions(ModeTreeV))
	if result == nil {
		t.Fatal("expected non-nil")
	}
	if result.Entry.Name != "root" {
		t.Errorf("expected root, got %s", result.Entry.Name)
	}
}

func TestComputeTreeV_FilesOnPedestal(t *testing.T) {
	tree := &fs.Tree{
		Root: &fs.Entry{
			Name: "root",
			Type: fs.TypeDir,
			Size: 2000,
			Children: []*fs.Entry{
				{Name: "a.txt", Type: fs.TypeFile, Size: 1000, ModTime: time.Now(), Depth: 1},
				{Name: "b.txt", Type: fs.TypeFile, Size: 1000, ModTime: time.Now(), Depth: 1},
			},
		},
		TotalSize: 2000,
	}

	result := Compute(tree, DefaultOptions(ModeTreeV))
	if result == nil {
		t.Fatal("expected non-nil")
	}

	// Files should be children positioned above the pedestal
	if len(result.Children) != 2 {
		t.Errorf("expected 2 file children, got %d", len(result.Children))
	}

	pedestalTop := result.Position.Y + result.Size.Y/2
	for _, child := range result.Children {
		if child.Position.Y < pedestalTop {
			t.Errorf("file %s (Y=%f) should be above pedestal top (%f)",
				child.Entry.Name, child.Position.Y, pedestalTop)
		}
	}
}

func TestComputeTreeV_SubdirsSpreadHorizontally(t *testing.T) {
	tree := &fs.Tree{
		Root: &fs.Entry{
			Name: "root",
			Type: fs.TypeDir,
			Size: 2000,
			Children: []*fs.Entry{
				{Name: "dir1", Type: fs.TypeDir, Size: 1000, Depth: 1},
				{Name: "dir2", Type: fs.TypeDir, Size: 1000, Depth: 1},
			},
		},
		TotalSize: 2000,
	}

	result := Compute(tree, DefaultOptions(ModeTreeV))
	if result == nil {
		t.Fatal("expected non-nil")
	}

	if len(result.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(result.Children))
	}

	// Child directories should be at different X positions
	x1 := result.Children[0].Position.X
	x2 := result.Children[1].Position.X
	if x1 == x2 {
		t.Error("child directories should have different X positions")
	}

	// Child directories should be at a deeper Z position (negative Z, matching fsnav)
	for _, child := range result.Children {
		if child.Position.Z >= result.Position.Z {
			t.Errorf("child dir %s (Z=%f) should be deeper (more negative) than parent (Z=%f)",
				child.Entry.Name, child.Position.Z, result.Position.Z)
		}
	}
}
