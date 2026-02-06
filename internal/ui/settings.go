package ui

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/Crank-Git/FSNRedux/internal/color"
)

// SettingsAction is returned when the user changes a setting.
type SettingsAction int

const (
	SettingsNone       SettingsAction = iota
	SettingsToggleHidden              // ShowHidden changed
	SettingsCycleTheme                // Theme changed
	SettingsDepthUp                   // MaxDepth increased
	SettingsDepthDown                 // MaxDepth decreased
	SettingsToggleLegend              // ShowLegend changed
)

// SettingsState holds runtime-modifiable settings and menu state.
type SettingsState struct {
	Open       bool
	ShowHidden bool
	ShowLegend bool
	Theme      string // "dark", "light", "auto"
	MaxDepth   int
	hoverIndex int // which row is hovered (-1 = none)
}

// NewSettingsState creates settings from the initial config values.
func NewSettingsState(showHidden bool, theme string, maxDepth int, showLegend bool) *SettingsState {
	if theme == "" {
		theme = "auto"
	}
	return &SettingsState{
		ShowHidden: showHidden,
		ShowLegend: showLegend,
		Theme:      theme,
		MaxDepth:   maxDepth,
		hoverIndex: -1,
	}
}

// settingsRow defines a row in the settings panel.
type settingsRow struct {
	label string
	value string
}

// DrawSettingsPanel renders the settings menu and returns any action taken.
func DrawSettingsPanel(state *SettingsState, screenW, screenH int32) SettingsAction {
	if state == nil || !state.Open {
		return SettingsNone
	}

	action := SettingsNone

	// Build rows
	hiddenStr := "Off"
	if state.ShowHidden {
		hiddenStr = "On"
	}
	legendStr := "Off"
	if state.ShowLegend {
		legendStr = "On"
	}
	depthStr := fmt.Sprintf("%d", state.MaxDepth)
	if state.MaxDepth == 0 {
		depthStr = "Unlimited"
	}

	rows := []settingsRow{
		{"Show Hidden Files", hiddenStr},
		{"Show Legend", legendStr},
		{"Theme", state.Theme},
		{"Max Scan Depth", depthStr},
	}

	// Panel dimensions
	panelW := int32(320)
	rowH := int32(32)
	headerH := int32(36)
	panelH := headerH + int32(len(rows))*rowH + 24 // +24 for padding + hint
	panelX := (screenW - panelW) / 2
	panelY := (screenH - panelH) / 2

	// Dimmed background
	rl.DrawRectangle(0, 0, screenW, screenH, rl.NewColor(0, 0, 0, 100))

	// Panel
	rl.DrawRectangle(panelX, panelY, panelW, panelH, color.SidebarBg)
	rl.DrawRectangleLines(panelX, panelY, panelW, panelH, color.Active.LinkAccent)

	// Title
	DrawTextUI("Settings", panelX+12, panelY+10, FontSize+2, color.TextPrimary)
	rl.DrawRectangle(panelX+12, panelY+headerH-2, panelW-24, 1, color.BorderColor)

	// Mouse interaction
	mousePos := rl.GetMousePosition()
	mouseClicked := rl.IsMouseButtonPressed(rl.MouseButtonLeft)
	state.hoverIndex = -1

	// Draw rows
	for i, row := range rows {
		ry := panelY + headerH + int32(i)*rowH
		rx := panelX

		// Hover detection
		inRow := int32(mousePos.X) >= rx && int32(mousePos.X) < rx+panelW &&
			int32(mousePos.Y) >= ry && int32(mousePos.Y) < ry+rowH
		if inRow {
			state.hoverIndex = i
			rl.DrawRectangle(rx+4, ry, panelW-8, rowH, color.HoverBg)
		}

		// Label
		DrawTextUI(row.label, rx+16, ry+8, FontSize, color.TextPrimary)

		// Value (right-aligned, with accent color)
		valColor := color.Active.LinkAccent
		valW := MeasureTextUI(row.value, FontSize)
		DrawTextUI(row.value, rx+panelW-valW-16, ry+8, FontSize, valColor)

		// Separator
		if i < len(rows)-1 {
			rl.DrawRectangle(rx+12, ry+rowH-1, panelW-24, 1, color.BorderColor)
		}

		// Handle clicks
		if inRow && mouseClicked {
			switch i {
			case 0: // Toggle hidden
				state.ShowHidden = !state.ShowHidden
				action = SettingsToggleHidden
			case 1: // Toggle legend
				state.ShowLegend = !state.ShowLegend
				action = SettingsToggleLegend
			case 2: // Cycle theme
				switch state.Theme {
				case "auto":
					state.Theme = "dark"
				case "dark":
					state.Theme = "light"
				case "light":
					state.Theme = "auto"
				}
				action = SettingsCycleTheme
			case 3: // Depth - left half = decrease, right half = increase
				midX := rx + panelW/2
				if int32(mousePos.X) < midX {
					if state.MaxDepth > 0 {
						state.MaxDepth--
					}
					action = SettingsDepthDown
				} else {
					state.MaxDepth++
					action = SettingsDepthUp
				}
			}
		}
	}

	// Keyboard handling for the rows
	if rl.IsKeyPressed(rl.KeyOne) || rl.IsKeyPressed(rl.KeyKp1) {
		state.ShowHidden = !state.ShowHidden
		action = SettingsToggleHidden
	}
	if rl.IsKeyPressed(rl.KeyTwo) || rl.IsKeyPressed(rl.KeyKp2) {
		state.ShowLegend = !state.ShowLegend
		action = SettingsToggleLegend
	}
	if rl.IsKeyPressed(rl.KeyThree) || rl.IsKeyPressed(rl.KeyKp3) {
		switch state.Theme {
		case "auto":
			state.Theme = "dark"
		case "dark":
			state.Theme = "light"
		case "light":
			state.Theme = "auto"
		}
		action = SettingsCycleTheme
	}
	if rl.IsKeyPressed(rl.KeyFour) || rl.IsKeyPressed(rl.KeyKp4) {
		state.MaxDepth++
		action = SettingsDepthUp
	}

	// Depth controls hint for row 4
	depthHintY := panelY + headerH + int32(len(rows))*rowH + 4
	DrawTextUI("Depth: click left(-) / right(+) or press 4", panelX+12, depthHintY, SmallFontSize, color.TextDim)

	// Close hint
	hintY := panelY + panelH - 16
	hint := "Press Comma or Escape to close"
	hintW := MeasureTextUI(hint, SmallFontSize)
	DrawTextUI(hint, panelX+(panelW-hintW)/2, hintY, SmallFontSize, color.TextDim)

	return action
}
