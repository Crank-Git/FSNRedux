package scene

import (
	"sync/atomic"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/Crank-Git/FSNRedux/internal/layout"
)

var nextID atomic.Uint32

// Graph is the root of the scene hierarchy.
type Graph struct {
	Root       *SceneNode
	NodeIndex  map[uint32]*SceneNode
	NodeByPath map[string]*SceneNode
	NodeCount  int
}

// NewGraph creates a Graph from a layout tree.
// expandedPaths controls which directories start expanded.
func NewGraph(layoutRoot *layout.Node, expandedPaths map[string]bool) *Graph {
	if layoutRoot == nil {
		return &Graph{
			NodeIndex:  make(map[uint32]*SceneNode),
			NodeByPath: make(map[string]*SceneNode),
		}
	}

	g := &Graph{
		NodeIndex:  make(map[uint32]*SceneNode),
		NodeByPath: make(map[string]*SceneNode),
	}

	g.Root = g.buildNode(layoutRoot, nil, expandedPaths)
	return g
}

func (g *Graph) buildNode(ln *layout.Node, parent *SceneNode, expandedPaths map[string]bool) *SceneNode {
	id := nextID.Add(1)

	expanded := false
	if ln.Entry != nil {
		expanded = expandedPaths[ln.Entry.Path]
	}

	node := &SceneNode{
		ID:       id,
		Entry:    ln.Entry,
		Position: ln.Position,
		Size:     ln.Size,
		Color:    ln.Color,
		Visible:  true,
		Expanded: expanded,
		Depth:    ln.Depth,
		Parent:   parent,
	}
	node.ComputeBounds()

	g.NodeIndex[id] = node
	if ln.Entry != nil {
		g.NodeByPath[ln.Entry.Path] = node
	}
	g.NodeCount++

	for _, childLayout := range ln.Children {
		child := g.buildNode(childLayout, node, expandedPaths)
		node.Children = append(node.Children, child)
	}

	return node
}

// Traverse calls fn for every visible node in depth-first order.
// If fn returns false, children of that node are skipped.
func (g *Graph) Traverse(fn func(node *SceneNode) bool) {
	if g.Root == nil {
		return
	}
	traverseNode(g.Root, fn)
}

func traverseNode(node *SceneNode, fn func(*SceneNode) bool) {
	if !node.Visible {
		return
	}
	if !fn(node) {
		return
	}
	if node.Expanded {
		for _, child := range node.Children {
			traverseNode(child, fn)
		}
	}
}

// Pick returns the closest node intersected by the given ray, or nil.
func (g *Graph) Pick(ray rl.Ray) *SceneNode {
	if g.Root == nil {
		return nil
	}

	var closest *SceneNode
	closestDist := float32(1e30)

	g.Traverse(func(node *SceneNode) bool {
		collision := rl.GetRayCollisionBox(ray, node.Bounds)
		if collision.Hit && collision.Distance < closestDist {
			closestDist = collision.Distance
			closest = node
		}
		return true
	})

	return closest
}

// FindByPath returns the node at the given filesystem path.
func (g *Graph) FindByPath(path string) *SceneNode {
	return g.NodeByPath[path]
}

// VisibleNodeCount returns the number of currently visible nodes.
func (g *Graph) VisibleNodeCount() int {
	count := 0
	g.Traverse(func(node *SceneNode) bool {
		count++
		return true
	})
	return count
}
