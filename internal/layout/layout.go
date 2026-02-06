package layout

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/Crank-Git/FSNRedux/internal/fs"
)

// Mode selects the visualization algorithm.
type Mode uint8

const (
	ModeMapV  Mode = iota // Squarified treemap with 3D extrusion
	ModeTreeV             // Hierarchical tree with pedestals and columns
)

// String returns the mode name.
func (m Mode) String() string {
	switch m {
	case ModeMapV:
		return "MapV"
	case ModeTreeV:
		return "TreeV"
	default:
		return "Unknown"
	}
}

// Options controls layout parameters.
type Options struct {
	Mode          Mode
	MaxDepth      int              // limit visible depth (0 = unlimited)
	PaddingRatio  float32          // spacing between sibling cuboids (default 0.02)
	HeightScale   float32          // multiplier for file-size-to-height mapping (default 1.0)
	MinHeight     float32          // minimum cuboid height (default 0.1)
	MaxHeight     float32          // maximum cuboid height (default 20.0)
	ExpandedPaths map[string]bool  // which directories are expanded (nil = all)
}

// DefaultOptions returns sensible default layout options.
func DefaultOptions(mode Mode) Options {
	return Options{
		Mode:         mode,
		MaxDepth:     0,
		PaddingRatio: 0.02,
		HeightScale:  1.0,
		MinHeight:    0.1,
		MaxHeight:    20.0,
	}
}

// Node is a positioned element in the layout with its computed geometry.
type Node struct {
	Entry    *fs.Entry
	Position rl.Vector3 // center of the cuboid
	Size     rl.Vector3 // width (X), height (Y), depth (Z)
	Color    rl.Color
	Children []*Node
	Depth    int
}

// Rect2D is a 2D rectangle used for treemap subdivision.
type Rect2D struct {
	X, Y, W, H float32
}

// Compute transforms an fs.Tree into a positioned layout tree.
func Compute(tree *fs.Tree, opts Options) *Node {
	if tree == nil || tree.Root == nil {
		return nil
	}
	switch opts.Mode {
	case ModeMapV:
		return computeMapV(tree, opts)
	case ModeTreeV:
		return computeTreeV(tree, opts)
	default:
		return computeMapV(tree, opts)
	}
}
