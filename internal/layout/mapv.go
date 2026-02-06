package layout

import (
	"math"
	"sort"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/Crank-Git/FSNRedux/internal/color"
	"github.com/Crank-Git/FSNRedux/internal/fs"
)

// computeMapV generates a squarified treemap layout with 3D extrusion.
func computeMapV(tree *fs.Tree, opts Options) *Node {
	// The root occupies a square on the ground plane
	totalArea := float32(30.0) // base size of the visualization
	rootRect := Rect2D{
		X: -totalArea / 2,
		Y: -totalArea / 2,
		W: totalArea,
		H: totalArea,
	}

	return layoutMapVNode(tree.Root, rootRect, 0, opts)
}

func layoutMapVNode(entry *fs.Entry, rect Rect2D, depth int, opts Options) *Node {
	if opts.MaxDepth > 0 && depth > opts.MaxDepth {
		return nil
	}

	height := scaleHeight(entry.Size, opts)
	nodeColor := color.DirColor
	if entry.Type != fs.TypeDir {
		nodeColor = color.ColorFromAge(entry.ModTime)
	}

	node := &Node{
		Entry: entry,
		Position: rl.NewVector3(
			rect.X+rect.W/2,
			height/2,
			rect.Y+rect.H/2,
		),
		Size: rl.NewVector3(
			rect.W*(1-opts.PaddingRatio),
			height,
			rect.H*(1-opts.PaddingRatio),
		),
		Color: nodeColor,
		Depth: depth,
	}

	if entry.Type == fs.TypeDir && len(entry.Children) > 0 {
		// Apply padding to create the inner rect for children
		padding := rect.W * opts.PaddingRatio
		innerRect := Rect2D{
			X: rect.X + padding,
			Y: rect.Y + padding,
			W: rect.W - 2*padding,
			H: rect.H - 2*padding,
		}

		// Filter to children with size > 0
		sizedChildren := make([]*fs.Entry, 0, len(entry.Children))
		for _, child := range entry.Children {
			if child.Size > 0 {
				sizedChildren = append(sizedChildren, child)
			}
		}
		// Also add zero-size children so they still appear
		for _, child := range entry.Children {
			if child.Size == 0 {
				sizedChildren = append(sizedChildren, child)
			}
		}

		if len(sizedChildren) > 0 {
			rects := squarify(sizedChildren, innerRect, entry.Size)
			for i, child := range sizedChildren {
				if i < len(rects) {
					childNode := layoutMapVNode(child, rects[i], depth+1, opts)
					if childNode != nil {
						// Raise children above the parent pedestal
						childNode.Position.Y += height
						raiseChildren(childNode, height)
						node.Children = append(node.Children, childNode)
					}
				}
			}
		}
	}

	return node
}

// raiseChildren recursively raises all child nodes by a given Y offset.
func raiseChildren(node *Node, offset float32) {
	for _, child := range node.Children {
		child.Position.Y += offset
		raiseChildren(child, offset)
	}
}

// indexedArea pairs an original index with its computed area for sorting.
type indexedArea struct {
	index int
	area  float64
}

// squarify implements the squarified treemap algorithm.
// Returns a slice of Rect2D, one per child, proportional to child size.
func squarify(children []*fs.Entry, rect Rect2D, parentSize int64) []Rect2D {
	if len(children) == 0 {
		return nil
	}

	// Assign areas proportional to size
	totalSize := float64(0)
	for _, c := range children {
		totalSize += math.Max(float64(c.Size), 1) // minimum 1 to avoid zero-area
	}

	totalArea := float64(rect.W) * float64(rect.H)
	areas := make([]float64, len(children))
	for i, c := range children {
		areas[i] = (math.Max(float64(c.Size), 1) / totalSize) * totalArea
	}

	// Sort areas descending (children should already be sorted, but ensure)
	sorted := make([]indexedArea, len(areas))
	for i, a := range areas {
		sorted[i] = indexedArea{i, a}
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].area > sorted[j].area
	})

	rects := make([]Rect2D, len(children))
	remaining := Rect2D{X: rect.X, Y: rect.Y, W: rect.W, H: rect.H}

	i := 0
	for i < len(sorted) {
		// Determine the shorter side of the remaining rectangle
		shortSide := remaining.W
		if remaining.H < remaining.W {
			shortSide = remaining.H
		}

		// Greedily add items to the current row while aspect ratio improves
		row := []indexedArea{sorted[i]}
		rowArea := sorted[i].area
		i++

		for i < len(sorted) {
			testArea := rowArea + sorted[i].area
			if worstAspectRatio(row, rowArea, shortSide) <=
				worstAspectRatio(append(row, sorted[i]), testArea, shortSide) {
				break
			}
			row = append(row, sorted[i])
			rowArea += sorted[i].area
			i++
		}

		// Lay out the row
		remaining = layoutRow(row, rowArea, remaining, rects, shortSide)
	}

	return rects
}

// worstAspectRatio computes the worst aspect ratio in a row for the squarify algorithm.
func worstAspectRatio(row []indexedArea, rowArea float64, shortSide float32) float64 {
	if len(row) == 0 || shortSide == 0 || rowArea == 0 {
		return math.MaxFloat64
	}

	s2 := float64(shortSide) * float64(shortSide)
	worst := 0.0
	for _, item := range row {
		r1 := (s2 * item.area) / (rowArea * rowArea)
		r2 := (rowArea * rowArea) / (s2 * item.area)
		ratio := math.Max(r1, r2)
		if ratio > worst {
			worst = ratio
		}
	}
	return worst
}

// layoutRow positions items in a row and returns the remaining rectangle.
func layoutRow(row []indexedArea, rowArea float64, rect Rect2D, rects []Rect2D, shortSide float32) Rect2D {
	if len(row) == 0 {
		return rect
	}

	if rect.W < rect.H {
		// Lay out horizontally (row fills top portion)
		rowHeight := float32(rowArea / float64(rect.W))
		x := rect.X
		for _, item := range row {
			w := float32(item.area / float64(rowHeight))
			rects[item.index] = Rect2D{X: x, Y: rect.Y, W: w, H: rowHeight}
			x += w
		}
		return Rect2D{X: rect.X, Y: rect.Y + rowHeight, W: rect.W, H: rect.H - rowHeight}
	}

	// Lay out vertically (row fills left portion)
	rowWidth := float32(rowArea / float64(rect.H))
	y := rect.Y
	for _, item := range row {
		h := float32(item.area / float64(rowWidth))
		rects[item.index] = Rect2D{X: rect.X, Y: y, W: rowWidth, H: h}
		y += h
	}
	return Rect2D{X: rect.X + rowWidth, Y: rect.Y, W: rect.W - rowWidth, H: rect.H}
}
