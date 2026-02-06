package renderer

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/Crank-Git/FSNRedux/internal/color"
)

// ColorBuckets holds pre-computed colors for each age bucket.
type ColorBuckets struct {
	Colors [32]rl.Color
}

// NewColorBuckets initializes the 32 age-based color buckets.
func NewColorBuckets() *ColorBuckets {
	cb := &ColorBuckets{}
	for i := 0; i < 32; i++ {
		cb.Colors[i] = color.BucketColor(i)
	}
	return cb
}

// DrawConnectionLine draws a line between parent and child nodes (for TreeV mode).
func DrawConnectionLine(parent, child rl.Vector3, lineColor rl.Color) {
	// Draw a horizontal line from parent to child's X position, then vertical down
	midY := parent.Y
	rl.DrawLine3D(parent, rl.NewVector3(child.X, midY, child.Z), lineColor)
	rl.DrawLine3D(rl.NewVector3(child.X, midY, child.Z), child, lineColor)
}
