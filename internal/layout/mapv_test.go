package layout

import (
	"testing"
	"time"

	"github.com/Crank-Git/FSNRedux/internal/fs"
)

func TestComputeMapV_NilTree(t *testing.T) {
	result := Compute(nil, DefaultOptions(ModeMapV))
	if result != nil {
		t.Error("expected nil for nil tree")
	}
}

func TestComputeMapV_SingleFile(t *testing.T) {
	tree := &fs.Tree{
		Root: &fs.Entry{
			Name: "root",
			Type: fs.TypeDir,
			Size: 1000,
			Children: []*fs.Entry{
				{Name: "file.txt", Type: fs.TypeFile, Size: 1000, ModTime: time.Now()},
			},
		},
		TotalSize: 1000,
	}

	opts := DefaultOptions(ModeMapV)
	result := Compute(tree, opts)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Children) != 1 {
		t.Errorf("expected 1 child, got %d", len(result.Children))
	}
	if result.Children[0].Entry.Name != "file.txt" {
		t.Errorf("expected file.txt, got %s", result.Children[0].Entry.Name)
	}
}

func TestComputeMapV_MultipleFiles(t *testing.T) {
	tree := &fs.Tree{
		Root: &fs.Entry{
			Name: "root",
			Type: fs.TypeDir,
			Size: 3000,
			Children: []*fs.Entry{
				{Name: "big.txt", Type: fs.TypeFile, Size: 2000, ModTime: time.Now()},
				{Name: "small.txt", Type: fs.TypeFile, Size: 1000, ModTime: time.Now()},
			},
		},
		TotalSize: 3000,
	}

	result := Compute(tree, DefaultOptions(ModeMapV))

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Children) != 2 {
		t.Errorf("expected 2 children, got %d", len(result.Children))
	}

	// Larger file should have a bigger area (W*H)
	big := result.Children[0]
	small := result.Children[1]
	if big.Entry.Name == "small.txt" {
		big, small = small, big
	}
	bigArea := big.Size.X * big.Size.Z
	smallArea := small.Size.X * small.Size.Z
	if bigArea <= smallArea {
		t.Errorf("big file area (%f) should be > small file area (%f)", bigArea, smallArea)
	}
}

func TestComputeMapV_NestedDirs(t *testing.T) {
	tree := &fs.Tree{
		Root: &fs.Entry{
			Name: "root",
			Type: fs.TypeDir,
			Size: 500,
			Children: []*fs.Entry{
				{
					Name: "subdir",
					Type: fs.TypeDir,
					Size: 500,
					Children: []*fs.Entry{
						{Name: "deep.txt", Type: fs.TypeFile, Size: 500, ModTime: time.Now()},
					},
				},
			},
		},
		TotalSize: 500,
	}

	result := Compute(tree, DefaultOptions(ModeMapV))
	if result == nil {
		t.Fatal("expected non-nil")
	}
	if len(result.Children) != 1 {
		t.Fatalf("expected 1 child dir, got %d", len(result.Children))
	}

	subdir := result.Children[0]
	if len(subdir.Children) != 1 {
		t.Fatalf("expected 1 grandchild, got %d", len(subdir.Children))
	}

	// Grandchild should be positioned above parent
	if subdir.Children[0].Position.Y <= subdir.Position.Y {
		t.Error("grandchild should be above parent directory")
	}
}

func TestComputeMapV_MaxDepth(t *testing.T) {
	tree := &fs.Tree{
		Root: &fs.Entry{
			Name: "root",
			Type: fs.TypeDir,
			Size: 100,
			Depth: 0,
			Children: []*fs.Entry{
				{
					Name:  "level1",
					Type:  fs.TypeDir,
					Size:  100,
					Depth: 1,
					Children: []*fs.Entry{
						{
							Name:  "level2",
							Type:  fs.TypeDir,
							Size:  100,
							Depth: 2,
							Children: []*fs.Entry{
								{Name: "deep.txt", Type: fs.TypeFile, Size: 100, Depth: 3},
							},
						},
					},
				},
			},
		},
		TotalSize: 100,
	}

	opts := DefaultOptions(ModeMapV)
	opts.MaxDepth = 1
	result := Compute(tree, opts)

	if result == nil {
		t.Fatal("expected non-nil")
	}

	// At max depth 1, level1 should exist but level2 should not be expanded
	if len(result.Children) != 1 {
		t.Fatalf("expected 1 child, got %d", len(result.Children))
	}
	// level1 at depth 1 should have no children since depth 2 exceeds MaxDepth 1
	if len(result.Children[0].Children) != 0 {
		t.Errorf("expected 0 grandchildren at max depth, got %d", len(result.Children[0].Children))
	}
}

func TestScaleHeight(t *testing.T) {
	opts := DefaultOptions(ModeMapV)

	// Zero size -> minimum height
	h := scaleHeight(0, opts)
	if h != opts.MinHeight {
		t.Errorf("expected MinHeight for size 0, got %f", h)
	}

	// Larger files should produce taller heights
	h1 := scaleHeight(1024, opts)       // 1KB
	h2 := scaleHeight(1024*1024, opts)  // 1MB
	if h2 <= h1 {
		t.Errorf("1MB height (%f) should be > 1KB height (%f)", h2, h1)
	}

	// Very large file should be capped
	h3 := scaleHeight(1024*1024*1024*1024, opts) // 1TB
	if h3 > opts.MaxHeight {
		t.Errorf("height %f exceeds max %f", h3, opts.MaxHeight)
	}
}

func TestSquarify_SingleItem(t *testing.T) {
	children := []*fs.Entry{
		{Name: "only", Size: 100},
	}
	rect := Rect2D{X: 0, Y: 0, W: 10, H: 10}
	rects := squarify(children, rect, 100)

	if len(rects) != 1 {
		t.Fatalf("expected 1 rect, got %d", len(rects))
	}
	// Single item should fill entire rect
	r := rects[0]
	if r.W < 9.9 || r.H < 9.9 {
		t.Errorf("single item should fill rect, got %fx%f", r.W, r.H)
	}
}

func TestSquarify_ProportionalAreas(t *testing.T) {
	children := []*fs.Entry{
		{Name: "big", Size: 300},
		{Name: "small", Size: 100},
	}
	rect := Rect2D{X: 0, Y: 0, W: 10, H: 10}
	rects := squarify(children, rect, 400)

	if len(rects) != 2 {
		t.Fatalf("expected 2 rects, got %d", len(rects))
	}

	bigArea := rects[0].W * rects[0].H
	smallArea := rects[1].W * rects[1].H
	ratio := bigArea / smallArea
	// Should be roughly 3:1
	if ratio < 2.0 || ratio > 4.0 {
		t.Errorf("area ratio should be ~3, got %f", ratio)
	}
}
