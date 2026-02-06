package renderer

import (
	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/Crank-Git/FSNRedux/internal/color"
	"github.com/Crank-Git/FSNRedux/internal/scene"
)

// Link color matching fsnav: glColor3f(0.1, 0.75, 0.2)
var linkColor = rl.NewColor(26, 191, 51, 255)

// Renderer handles all 3D drawing.
type Renderer struct{}

// New creates a renderer.
func New() *Renderer {
	return &Renderer{}
}

// DrawScene renders the entire scene graph (matching fsnav's root->draw()).
func (r *Renderer) DrawScene(graph *scene.Graph, selected *scene.SceneNode, hovered *scene.SceneNode) {
	if graph == nil || graph.Root == nil {
		return
	}
	// fsnav draws post-order (children first, then parent) for correct transparency.
	// We do the same via traversal.
	graph.Traverse(func(node *scene.SceneNode) bool {
		r.drawNode(node, selected, hovered)
		return true
	})
}

func (r *Renderer) drawNode(node *scene.SceneNode, selected *scene.SceneNode, hovered *scene.SceneNode) {
	if node.Size.X < 0.01 || node.Size.Y < 0.01 || node.Size.Z < 0.01 {
		return
	}

	isDir := node.Entry != nil && node.Entry.IsDir()

	// Color-based selection/hover (matching fsnav get_color)
	drawColor := node.Color
	if node == selected {
		if isDir {
			drawColor = color.DirSelected
		} else {
			drawColor = color.FileSelected
		}
	} else if node == hovered {
		if isDir {
			drawColor = color.DirHover
		} else {
			drawColor = color.FileHover
		}
	}

	// Draw solid cube (matching fsnav draw_node -> draw_cube)
	rl.DrawCubeV(node.Position, node.Size, drawColor)

	// Connection lines from parent center to child center (matching fsnav)
	if isDir && node.Expanded {
		for _, child := range node.Children {
			if child.Visible && child.Entry != nil && child.Entry.IsDir() {
				rl.DrawLine3D(node.Position, child.Position, linkColor)
			}
		}
	}
}
