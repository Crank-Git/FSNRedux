package scene

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/Crank-Git/FSNRedux/internal/fs"
)

// SceneNode is a renderable entity in the 3D scene.
type SceneNode struct {
	ID       uint32
	Entry    *fs.Entry
	Position rl.Vector3
	Size     rl.Vector3
	Color    rl.Color
	Bounds   rl.BoundingBox
	Visible  bool
	Expanded bool
	Depth    int
	Children []*SceneNode
	Parent   *SceneNode
}

// ComputeBounds calculates the axis-aligned bounding box from position and size.
func (n *SceneNode) ComputeBounds() {
	halfSize := rl.NewVector3(n.Size.X/2, n.Size.Y/2, n.Size.Z/2)
	n.Bounds = rl.BoundingBox{
		Min: rl.NewVector3(n.Position.X-halfSize.X, n.Position.Y-halfSize.Y, n.Position.Z-halfSize.Z),
		Max: rl.NewVector3(n.Position.X+halfSize.X, n.Position.Y+halfSize.Y, n.Position.Z+halfSize.Z),
	}
}

// ContainsPoint checks if a point is inside this node's bounds.
func (n *SceneNode) ContainsPoint(point rl.Vector3) bool {
	return point.X >= n.Bounds.Min.X && point.X <= n.Bounds.Max.X &&
		point.Y >= n.Bounds.Min.Y && point.Y <= n.Bounds.Max.Y &&
		point.Z >= n.Bounds.Min.Z && point.Z <= n.Bounds.Max.Z
}
