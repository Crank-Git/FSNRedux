package ui

import (
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/Crank-Git/FSNRedux/internal/color"
)

// InputBarMode distinguishes path entry from search.
type InputBarMode int

const (
	InputBarNone   InputBarMode = iota
	InputBarPath                // Ctrl+L: type a filesystem path
	InputBarSearch              // Ctrl+F / F: search by name
)

// InputBar is a text input overlay for path entry and search.
type InputBar struct {
	Active   bool
	Mode     InputBarMode
	Text     string
	cursor   int
	submitted bool
}

// Open activates the input bar with the given mode and optional initial text.
func (b *InputBar) Open(mode InputBarMode, initial string) {
	b.Active = true
	b.Mode = mode
	b.Text = initial
	b.cursor = len(initial)
	b.submitted = false
}

// Close deactivates the input bar.
func (b *InputBar) Close() {
	b.Active = false
	b.Mode = InputBarNone
	b.Text = ""
	b.cursor = 0
	b.submitted = false
}

// Update processes keyboard input for the bar. Returns true if submitted.
func (b *InputBar) Update() bool {
	if !b.Active {
		return false
	}
	b.submitted = false

	// Escape closes
	if rl.IsKeyPressed(rl.KeyEscape) {
		b.Close()
		return false
	}

	// Enter submits
	if rl.IsKeyPressed(rl.KeyEnter) || rl.IsKeyPressed(rl.KeyKpEnter) {
		b.submitted = true
		return true
	}

	// Backspace
	if rl.IsKeyPressed(rl.KeyBackspace) || rl.IsKeyPressedRepeat(rl.KeyBackspace) {
		if b.cursor > 0 {
			b.Text = b.Text[:b.cursor-1] + b.Text[b.cursor:]
			b.cursor--
		}
	}

	// Delete
	if rl.IsKeyPressed(rl.KeyDelete) || rl.IsKeyPressedRepeat(rl.KeyDelete) {
		if b.cursor < len(b.Text) {
			b.Text = b.Text[:b.cursor] + b.Text[b.cursor+1:]
		}
	}

	// Cursor movement
	if rl.IsKeyPressed(rl.KeyLeft) || rl.IsKeyPressedRepeat(rl.KeyLeft) {
		if b.cursor > 0 {
			b.cursor--
		}
	}
	if rl.IsKeyPressed(rl.KeyRight) || rl.IsKeyPressedRepeat(rl.KeyRight) {
		if b.cursor < len(b.Text) {
			b.cursor++
		}
	}
	if rl.IsKeyPressed(rl.KeyHome) {
		b.cursor = 0
	}
	if rl.IsKeyPressed(rl.KeyEnd) {
		b.cursor = len(b.Text)
	}

	// Character input
	for {
		ch := rl.GetCharPressed()
		if ch == 0 {
			break
		}
		c := string(rune(ch))
		b.Text = b.Text[:b.cursor] + c + b.Text[b.cursor:]
		b.cursor++
	}

	return false
}

// Submitted returns true on the frame the user pressed Enter.
func (b *InputBar) Submitted() bool {
	return b.submitted
}

// Draw renders the input bar at the top of the screen.
func (b *InputBar) Draw(screenWidth int32) {
	if !b.Active {
		return
	}

	barH := int32(28)
	barY := BreadcrumbHeight
	barX := SidebarWidth

	// Background
	rl.DrawRectangle(barX, barY, screenWidth-barX, barH, rl.NewColor(
		color.Active.SidebarBg.R,
		color.Active.SidebarBg.G,
		color.Active.SidebarBg.B,
		240,
	))
	rl.DrawRectangle(barX, barY+barH-1, screenWidth-barX, 1, color.BorderColor)

	// Label
	label := "Path: "
	if b.Mode == InputBarSearch {
		label = "Search: "
	}
	labelW := MeasureTextUI(label, FontSize)
	textY := barY + 6

	DrawTextUI(label, barX+8, textY, FontSize, color.Active.LinkAccent)

	// Text content
	textX := barX + 8 + labelW + 4
	DrawTextUI(b.Text, textX, textY, FontSize, color.TextPrimary)

	// Cursor (blinking)
	if int(rl.GetTime()*3)%2 == 0 {
		cursorX := textX + MeasureTextUI(b.Text[:b.cursor], FontSize)
		rl.DrawRectangle(cursorX, textY, 1, int32(FontSize), color.TextPrimary)
	}

	// Hint text
	hint := "Enter to navigate | Esc to cancel"
	if b.Mode == InputBarSearch {
		hint = "Enter to find | Esc to cancel"
	}
	hintW := MeasureTextUI(hint, SmallFontSize)
	DrawTextUI(hint, screenWidth-hintW-8, textY+2, SmallFontSize, color.TextDim)
}

// SearchResults holds search matches.
type SearchResults struct {
	Matches []string // paths that match
	Current int      // index of currently focused match
	Query   string   // the search query
}

// SearchTree searches the fs tree for entries matching the query (case-insensitive substring).
func SearchTree(root interface{ GetPath() string }, query string) []string {
	return nil // placeholder - search is done via scene graph
}

// SearchSceneByName returns paths of nodes whose name contains the query (case-insensitive).
func SearchSceneByName(query string, allPaths map[string]bool) []string {
	if query == "" {
		return nil
	}
	q := strings.ToLower(query)
	var results []string
	for path := range allPaths {
		name := path
		if idx := strings.LastIndex(path, "/"); idx >= 0 {
			name = path[idx+1:]
		}
		if strings.Contains(strings.ToLower(name), q) {
			results = append(results, path)
		}
	}
	return results
}
