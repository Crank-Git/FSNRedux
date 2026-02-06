package layout

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/Crank-Git/FSNRedux/internal/color"
	"github.com/Crank-Git/FSNRedux/internal/fs"
)

// Layout parameters matching fsnav exactly.
const (
	lpFileSize    = 0.5
	lpFileSpacing = 0.1
	lpFileHeight  = 0.1
	lpDirSize     = 0.7 // minimum dir size (fsnav: 0.5 + 0.2)
	lpDirSpacing  = 0.5
	lpDirHeight   = 0.1
	lpDirDist     = 5.0
)

// dirBounds stores layout bounds for a directory (matching fsnav Dir::min_x/max_x/vis_size).
type dirBounds struct {
	minX, maxX float32
	size       rl.Vector3 // visual size of the dir pedestal
}

// computeTreeV generates the FSN-style hierarchical layout matching fsnav.
func computeTreeV(tree *fs.Tree, opts Options) *Node {
	bounds := make(map[*fs.Entry]*dirBounds)
	calcBounds(tree.Root, bounds, opts)
	return place(tree.Root, rl.NewVector3(0, lpDirHeight/2, 0), bounds, opts)
}

// calcDirSize computes pedestal size based on file count (matching fsnav calc_dir_size).
func calcDirSize(numFiles int) (float32, float32) {
	if numFiles == 0 {
		return lpDirSize, lpDirSize
	}
	filesX := int(math.Ceil(math.Sqrt(float64(numFiles))))
	filesY := int(math.Ceil(float64(numFiles) / float64(filesX)))

	xsz := float32(filesX)*lpFileSize + float32(filesX+1)*lpFileSpacing
	ysz := float32(filesY)*lpFileSize + float32(filesY+1)*lpFileSpacing

	if xsz < lpDirSize {
		xsz = lpDirSize
	}
	if ysz < lpDirSize {
		ysz = lpDirSize
	}
	return xsz, ysz
}

// calcBounds recursively computes width bounds for each directory (matching fsnav Dir::calc_bounds).
func calcBounds(entry *fs.Entry, bounds map[*fs.Entry]*dirBounds, opts Options) {
	if entry.Type != fs.TypeDir {
		return
	}

	isExpanded := opts.ExpandedPaths == nil || opts.ExpandedPaths[entry.Path]

	// Collapsed directories use minimum size (no files shown)
	if !isExpanded {
		b := &dirBounds{
			size: rl.NewVector3(lpDirSize, lpDirHeight, lpDirSize),
		}
		b.minX = -(lpDirSize + lpDirSpacing) / 2
		b.maxX = (lpDirSize + lpDirSpacing) / 2
		bounds[entry] = b
		return
	}

	// Count files for expanded directories
	numFiles := 0
	for _, child := range entry.Children {
		if child.Type != fs.TypeDir {
			numFiles++
		}
	}

	dirW, dirD := calcDirSize(numFiles)
	b := &dirBounds{
		size: rl.NewVector3(dirW, lpDirHeight, dirD),
	}

	if opts.MaxDepth > 0 && entry.Depth >= opts.MaxDepth {
		b.minX = -(dirW + lpDirSpacing) / 2
		b.maxX = (dirW + lpDirSpacing) / 2
		bounds[entry] = b
		return
	}

	// Recurse into subdirs
	childWidth := float32(0)
	for _, child := range entry.Children {
		if child.Type == fs.TypeDir {
			calcBounds(child, bounds, opts)
			cb := bounds[child]
			childWidth += cb.maxX - cb.minX
		}
	}

	width := dirW
	if childWidth > width {
		width = childWidth
	}

	b.minX = -(width + lpDirSpacing) / 2
	b.maxX = (width + lpDirSpacing) / 2
	bounds[entry] = b
}

// place recursively positions nodes (matching fsnav Dir::place).
func place(entry *fs.Entry, pos rl.Vector3, bounds map[*fs.Entry]*dirBounds, opts Options) *Node {
	if opts.MaxDepth > 0 && entry.Depth > opts.MaxDepth {
		return nil
	}

	b := bounds[entry]
	if b == nil {
		// Non-directory entries shouldn't reach here, but handle gracefully
		return &Node{
			Entry:    entry,
			Position: pos,
			Size:     rl.NewVector3(lpFileSize, lpFileHeight, lpFileSize),
			Color:    color.FileColor,
			Depth:    entry.Depth,
		}
	}

	node := &Node{
		Entry:    entry,
		Position: pos,
		Size:     b.size,
		Color:    color.DirColor,
		Depth:    entry.Depth,
	}

	if entry.Type != fs.TypeDir {
		return node
	}

	// Only show contents for expanded directories
	isExpanded := opts.ExpandedPaths == nil || opts.ExpandedPaths[entry.Path]
	if !isExpanded {
		return node
	}

	// Separate files and subdirs
	var files []*fs.Entry
	var dirs []*fs.Entry
	for _, child := range entry.Children {
		if child.Type == fs.TypeDir {
			dirs = append(dirs, child)
		} else {
			files = append(files, child)
		}
	}

	// Place files in grid on top of pedestal (matching fsnav)
	if len(files) > 0 {
		sideFiles := int(math.Ceil(math.Sqrt(float64(len(files)))))
		fRowWidth := float32(sideFiles)*lpFileSize + float32(sideFiles-1)*lpFileSpacing

		offs := float32(lpFileSize/2 + lpFileSpacing)
		fStartX := pos.X - b.size.X/2 + offs
		fStartZ := pos.Z - b.size.Z/2 + offs
		fStartY := pos.Y + b.size.Y/2 + float32(lpFileHeight/2) // on top of pedestal

		fPosX := fStartX
		fPosZ := fStartZ
		_ = fRowWidth

		// Find max file size for color scaling
		var maxFileSize int64
		for _, file := range files {
			if file.Size > maxFileSize {
				maxFileSize = file.Size
			}
		}

		for i, file := range files {
			col := i % sideFiles

			fileColor := color.ColorFromSize(file.Size, maxFileSize)

			fileNode := &Node{
				Entry:    file,
				Position: rl.NewVector3(fPosX, fStartY, fPosZ),
				Size:     rl.NewVector3(lpFileSize, lpFileHeight, lpFileSize),
				Color:    fileColor,
				Depth:    file.Depth,
			}
			node.Children = append(node.Children, fileNode)

			fPosX += lpFileSize + lpFileSpacing
			if col == sideFiles-1 {
				fPosX = fStartX
				fPosZ += lpFileSize + lpFileSpacing
			}
		}
	}

	// Place child directories behind the pedestal
	if len(dirs) > 0 && (opts.MaxDepth == 0 || entry.Depth < opts.MaxDepth) {
		x := b.minX - lpDirSpacing/2
		for _, dir := range dirs {
			cb := bounds[dir]
			if cb == nil {
				continue
			}
			width := cb.maxX - cb.minX

			childPos := rl.NewVector3(
				pos.X+x+width/2,
				pos.Y, // same Y as parent (matching fsnav)
				pos.Z-(b.size.Z/2+lpDirDist), // negative Z (matching fsnav)
			)

			childNode := place(dir, childPos, bounds, opts)
			if childNode != nil {
				node.Children = append(node.Children, childNode)
			}

			x += width + lpDirSpacing
		}
	}

	return node
}
