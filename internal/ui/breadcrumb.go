package ui

import (
	"path/filepath"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/Crank-Git/FSNRedux/internal/color"
)

// DrawBreadcrumb renders the path breadcrumb bar at the top of the window.
// Returns the clicked path segment if any was clicked, empty string otherwise.
func DrawBreadcrumb(currentPath string, rootPath string, screenWidth int32) string {
	DrawPanel(0, 0, screenWidth, BreadcrumbHeight, color.SidebarBg)

	if currentPath == "" {
		return ""
	}

	// Build segment list: [rootName, relative, path, parts...]
	rootName := filepath.Base(rootPath)
	if rootName == "/" || rootName == "." {
		rootName = rootPath // show full root for "/" or "."
	}

	segments := []string{rootName}
	// Paths for each segment (absolute)
	segPaths := []string{rootPath}

	if currentPath != rootPath {
		relPath, _ := filepath.Rel(rootPath, currentPath)
		if relPath != "" && relPath != "." {
			parts := strings.Split(relPath, string(filepath.Separator))
			for i, part := range parts {
				if part == "" {
					continue
				}
				segments = append(segments, part)
				segPaths = append(segPaths, filepath.Join(rootPath, filepath.Join(parts[:i+1]...)))
			}
		}
	}

	x := int32(8)
	y := int32(float32(BreadcrumbHeight)/2 - FontSize/2)
	clicked := ""
	mousePos := rl.GetMousePosition()

	DrawTextUI("[", x, y, FontSize, color.TextDim)
	x += 8

	for i, seg := range segments {
		segWidth := MeasureTextUI(seg, FontSize)
		segRect := rl.NewRectangle(float32(x-2), 2, float32(segWidth+4), float32(BreadcrumbHeight-4))

		isHovered := rl.CheckCollisionPointRec(mousePos, segRect)
		if isHovered {
			rl.DrawRectangleRec(segRect, color.HoverBg)
		}

		segColor := color.TextPrimary
		if i == len(segments)-1 {
			segColor = color.Active.LinkAccent
		}
		DrawTextUI(seg, x, y, FontSize, segColor)

		if isHovered && rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
			clicked = segPaths[i]
		}

		x += segWidth

		if i < len(segments)-1 {
			DrawTextUI(" > ", x, y, FontSize, color.TextDim)
			x += MeasureTextUI(" > ", FontSize)
		}
	}

	DrawTextUI("]", x, y, FontSize, color.TextDim)

	return clicked
}
