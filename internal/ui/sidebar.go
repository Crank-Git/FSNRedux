package ui

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/Crank-Git/FSNRedux/internal/color"
	"github.com/Crank-Git/FSNRedux/internal/fs"
)

// TreeViewState holds the sidebar tree state.
type TreeViewState struct {
	ScrollOffset float32
	ExpandedDirs map[string]bool
	SelectedPath string
	HoveredPath  string
	rows         []treeRow // computed each frame

	// Sidebar search
	SearchActive bool
	SearchText   string
	SearchCursor int
	SearchSubmit string // set to query on Enter, cleared by caller
}

type treeRow struct {
	Entry *fs.Entry
	Depth int
	Y     float32
}

// NewTreeViewState creates initial sidebar state with root expanded.
func NewTreeViewState(rootPath string) *TreeViewState {
	return &TreeViewState{
		ExpandedDirs: map[string]bool{rootPath: true},
	}
}

// SidebarSearchState holds the search field state in the sidebar.
type SidebarSearchState struct {
	Active  bool
	Text    string
	cursor  int
}

// DrawSidebar renders the file tree sidebar and returns the selected path if clicked.
// searchState is the sidebar search field state. searchSubmit receives the query on Enter.
func DrawSidebar(tree *fs.Tree, state *TreeViewState, screenHeight int32) string {
	if tree == nil || tree.Root == nil {
		return ""
	}

	panelX := int32(0)
	panelY := BreadcrumbHeight
	panelW := SidebarWidth
	panelH := screenHeight - BreadcrumbHeight - InfoPanelHeight

	// Background
	DrawPanel(panelX, panelY, panelW, panelH, color.SidebarBg)

	// Right edge shadow (gradient effect using 3 lines)
	rl.DrawRectangle(panelX+panelW, panelY, 1, panelH, rl.NewColor(0, 0, 0, 60))
	rl.DrawRectangle(panelX+panelW+1, panelY, 1, panelH, rl.NewColor(0, 0, 0, 30))
	rl.DrawRectangle(panelX+panelW+2, panelY, 1, panelH, rl.NewColor(0, 0, 0, 10))

	// Search field header
	headerH := int32(26)
	searchBoxY := panelY + 3
	searchBoxH := int32(20)
	searchBoxW := panelW - 16

	// Search box background
	boxColor := rl.NewColor(
		color.Active.Background.R,
		color.Active.Background.G,
		color.Active.Background.B,
		255,
	)
	rl.DrawRectangle(panelX+8, searchBoxY, searchBoxW, searchBoxH, boxColor)
	rl.DrawRectangleLines(panelX+8, searchBoxY, searchBoxW, searchBoxH, color.BorderColor)

	// Search icon/placeholder
	if state.SearchText == "" && !state.SearchActive {
		DrawTextUI("Search (F)...", panelX+12, searchBoxY+3, SmallFontSize, color.TextDim)
	} else {
		DrawTextUI(state.SearchText, panelX+12, searchBoxY+3, SmallFontSize, color.TextPrimary)
		if state.SearchActive {
			// Cursor
			if int(rl.GetTime()*3)%2 == 0 {
				cursorX := panelX + 12 + MeasureTextUI(state.SearchText[:state.SearchCursor], SmallFontSize)
				rl.DrawRectangle(cursorX, searchBoxY+3, 1, int32(SmallFontSize), color.TextPrimary)
			}
		}
	}

	// Click to activate search
	mousePos := rl.GetMousePosition()
	searchRect := rl.NewRectangle(float32(panelX+8), float32(searchBoxY), float32(searchBoxW), float32(searchBoxH))
	if rl.CheckCollisionPointRec(mousePos, searchRect) && rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
		state.SearchActive = true
	}

	// Handle search text input when active
	if state.SearchActive {
		if rl.IsKeyPressed(rl.KeyEscape) {
			state.SearchActive = false
		} else if rl.IsKeyPressed(rl.KeyEnter) || rl.IsKeyPressed(rl.KeyKpEnter) {
			state.SearchSubmit = state.SearchText
			state.SearchActive = false
		} else {
			if rl.IsKeyPressed(rl.KeyBackspace) || rl.IsKeyPressedRepeat(rl.KeyBackspace) {
				if state.SearchCursor > 0 {
					state.SearchText = state.SearchText[:state.SearchCursor-1] + state.SearchText[state.SearchCursor:]
					state.SearchCursor--
				}
			}
			if rl.IsKeyPressed(rl.KeyLeft) || rl.IsKeyPressedRepeat(rl.KeyLeft) {
				if state.SearchCursor > 0 {
					state.SearchCursor--
				}
			}
			if rl.IsKeyPressed(rl.KeyRight) || rl.IsKeyPressedRepeat(rl.KeyRight) {
				if state.SearchCursor < len(state.SearchText) {
					state.SearchCursor++
				}
			}
			for {
				ch := rl.GetCharPressed()
				if ch == 0 {
					break
				}
				c := string(rune(ch))
				state.SearchText = state.SearchText[:state.SearchCursor] + c + state.SearchText[state.SearchCursor:]
				state.SearchCursor++
			}
		}
	}

	// Header separator
	rl.DrawRectangle(panelX+8, panelY+headerH-1, panelW-16, 1, color.BorderColor)

	// Compute visible rows
	state.rows = state.rows[:0]
	flattenTree(tree.Root, 0, state, &state.rows)

	// Content area (clipped)
	contentY := panelY + headerH
	contentH := panelH - headerH
	visibleRows := int(float32(contentH) / RowHeight)

	// Handle scroll
	mouseWheel := rl.GetMouseWheelMove()
	mousePos = rl.GetMousePosition()
	if mousePos.X < float32(panelW) && mousePos.Y > float32(contentY) {
		state.ScrollOffset -= mouseWheel * RowHeight * 3
		maxScroll := float32(len(state.rows))*RowHeight - float32(contentH)
		if maxScroll < 0 {
			maxScroll = 0
		}
		if state.ScrollOffset < 0 {
			state.ScrollOffset = 0
		}
		if state.ScrollOffset > maxScroll {
			state.ScrollOffset = maxScroll
		}
	}

	startRow := int(state.ScrollOffset / RowHeight)
	if startRow < 0 {
		startRow = 0
	}

	// Enable scissor clipping
	rl.BeginScissorMode(panelX, contentY, panelW, contentH)

	clickedPath := ""
	state.HoveredPath = ""

	for i := startRow; i < len(state.rows) && i < startRow+visibleRows+2; i++ {
		row := state.rows[i]
		rowY := float32(contentY) + float32(i)*RowHeight - state.ScrollOffset

		if rowY+RowHeight < float32(contentY) || rowY > float32(contentY+contentH) {
			continue
		}

		indent := float32(row.Depth) * IndentWidth
		rowRect := rl.NewRectangle(float32(panelX), rowY, float32(panelW), RowHeight)

		// Hover detection
		isHovered := rl.CheckCollisionPointRec(mousePos, rowRect) && mousePos.X < float32(panelW)
		isSelected := row.Entry.Path == state.SelectedPath

		// Background highlight
		if isSelected {
			rl.DrawRectangle(panelX, int32(rowY), panelW, int32(RowHeight), color.SelectionBg)
		} else if isHovered {
			rl.DrawRectangle(panelX, int32(rowY), panelW, int32(RowHeight), color.HoverBg)
			state.HoveredPath = row.Entry.Path
		}

		// Expand/collapse triangle for directories
		textX := float32(panelX+8) + indent
		if row.Entry.Type == fs.TypeDir {
			triX := textX
			triY := rowY + RowHeight/2

			if state.ExpandedDirs[row.Entry.Path] {
				// Down-pointing triangle (expanded)
				rl.DrawTriangle(
					rl.NewVector2(triX, triY-3),
					rl.NewVector2(triX+8, triY-3),
					rl.NewVector2(triX+4, triY+4),
					color.TextSecondary,
				)
			} else {
				// Right-pointing triangle (collapsed)
				rl.DrawTriangle(
					rl.NewVector2(triX, triY-4),
					rl.NewVector2(triX, triY+4),
					rl.NewVector2(triX+7, triY),
					color.TextSecondary,
				)
			}
			textX += 12
		}

		// Icon (skip for root "/" to avoid showing "//")
		if row.Entry.Type == fs.TypeDir && row.Entry.Name != "/" {
			DrawTextUI("/", int32(textX), int32(rowY+3), FontSize, color.Active.DirAccent)
			textX += 10
		} else if row.Entry.Type != fs.TypeDir {
			DrawTextUI(".", int32(textX), int32(rowY+3), FontSize, color.TextDim)
			textX += 8
		}

		// Name (truncate if too long)
		name := row.Entry.Name
		maxChars := int((float32(panelW) - textX - 8) / 8) // approximate char width
		if maxChars > 0 && len(name) > maxChars {
			name = name[:maxChars-2] + ".."
		}
		textColor := color.TextPrimary
		if row.Entry.Type == fs.TypeDir {
			textColor = color.Active.DirAccent
		}
		DrawTextUI(name, int32(textX), int32(rowY+3), FontSize, textColor)

		// Handle click
		if isHovered && rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
			if row.Entry.Type == fs.TypeDir {
				// Toggle expand/collapse
				state.ExpandedDirs[row.Entry.Path] = !state.ExpandedDirs[row.Entry.Path]
			}
			state.SelectedPath = row.Entry.Path
			clickedPath = row.Entry.Path
		}
	}

	rl.EndScissorMode()

	// Scrollbar
	if len(state.rows) > visibleRows {
		scrollbarH := float32(contentH) * float32(visibleRows) / float32(len(state.rows))
		if scrollbarH < 20 {
			scrollbarH = 20
		}
		maxScroll := float32(len(state.rows))*RowHeight - float32(contentH)
		scrollbarY := float32(contentY) + (state.ScrollOffset/maxScroll)*(float32(contentH)-scrollbarH)
		rl.DrawRectangle(panelX+panelW-6, int32(scrollbarY), 4, int32(scrollbarH),
			rl.NewColor(100, 100, 110, 180))
	}

	return clickedPath
}

// flattenTree builds the visible row list by walking the expanded tree.
func flattenTree(entry *fs.Entry, depth int, state *TreeViewState, rows *[]treeRow) {
	*rows = append(*rows, treeRow{Entry: entry, Depth: depth})

	if entry.Type == fs.TypeDir && state.ExpandedDirs[entry.Path] {
		for _, child := range entry.Children {
			flattenTree(child, depth+1, state, rows)
		}
	}
}

// FormatSize returns a human-readable file size string.
func FormatSize(size int64) string {
	switch {
	case size >= 1<<30:
		return fmt.Sprintf("%.1f GB", float64(size)/float64(1<<30))
	case size >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(size)/float64(1<<20))
	case size >= 1<<10:
		return fmt.Sprintf("%.1f KB", float64(size)/float64(1<<10))
	default:
		return fmt.Sprintf("%d B", size)
	}
}
