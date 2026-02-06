package input

import (
	"encoding/json"
	"os"
	"path/filepath"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Action is a named input action.
type Action string

const (
	ActionPanForward  Action = "pan_forward"
	ActionPanBack     Action = "pan_back"
	ActionPanLeft     Action = "pan_left"
	ActionPanRight    Action = "pan_right"
	ActionZoomIn      Action = "zoom_in"
	ActionZoomOut     Action = "zoom_out"
	ActionExpand      Action = "expand"       // Enter: expand selected dir
	ActionBack        Action = "back"         // Escape: collapse / go to parent
	ActionNextNode    Action = "next_node"    // Tab: select next visible node
	ActionPrevNode    Action = "prev_node"    // Shift+Tab: select previous visible node
	ActionSearch      Action = "search"       // F: open search
	ActionPathBar     Action = "path_bar"     // Ctrl+L: open path bar
	ActionToggleHelp  Action = "toggle_help"  // H: toggle help
	ActionHome        Action = "home"         // Home: focus on root
	ActionInspect     Action = "inspect"      // Space: inspect selected node
	ActionSettings    Action = "settings"    // Comma: open settings menu
	ActionOpenFile    Action = "open_file"   // O: open file with default app
	ActionBirdseye    Action = "birdseye"   // B: birdseye view of all expanded dirs
)

// KeyMap maps actions to raylib key codes.
type KeyMap struct {
	Bindings map[Action][]int32 `json:"bindings"`
}

// DefaultKeyMap returns the default key bindings.
func DefaultKeyMap() *KeyMap {
	return &KeyMap{
		Bindings: map[Action][]int32{
			ActionPanForward: {rl.KeyW, rl.KeyUp},
			ActionPanBack:    {rl.KeyS, rl.KeyDown},
			ActionPanLeft:    {rl.KeyA, rl.KeyLeft},
			ActionPanRight:   {rl.KeyD, rl.KeyRight},
			ActionZoomIn:     {rl.KeyEqual, rl.KeyKpAdd},
			ActionZoomOut:    {rl.KeyMinus, rl.KeyKpSubtract},
			ActionExpand:     {rl.KeyEnter, rl.KeyKpEnter},
			ActionBack:       {rl.KeyEscape},
			ActionNextNode:   {rl.KeyTab},
			ActionPrevNode:   {}, // Shift+Tab handled specially
			ActionSearch:     {rl.KeyF},
			ActionPathBar:    {rl.KeyL}, // requires Ctrl/Cmd modifier
			ActionToggleHelp: {rl.KeyH},
			ActionHome:       {rl.KeyHome},
			ActionInspect:    {rl.KeySpace},
			ActionSettings:   {rl.KeyComma},
			ActionOpenFile:   {rl.KeyO},
			ActionBirdseye:   {rl.KeyB},
		},
	}
}

// IsPressed returns true if any key bound to the action was just pressed.
func (km *KeyMap) IsPressed(action Action) bool {
	keys, ok := km.Bindings[action]
	if !ok {
		return false
	}
	for _, k := range keys {
		if rl.IsKeyPressed(k) {
			return true
		}
	}
	return false
}

// IsDown returns true if any key bound to the action is currently held.
func (km *KeyMap) IsDown(action Action) bool {
	keys, ok := km.Bindings[action]
	if !ok {
		return false
	}
	for _, k := range keys {
		if rl.IsKeyDown(k) {
			return true
		}
	}
	return false
}

// LoadKeyMap loads a keymap from a JSON file, falling back to defaults.
// Config file location: ~/.config/fsnredux/keys.json
func LoadKeyMap() *KeyMap {
	km := DefaultKeyMap()

	configDir, err := os.UserConfigDir()
	if err != nil {
		return km
	}
	path := filepath.Join(configDir, "fsnredux", "keys.json")

	data, err := os.ReadFile(path)
	if err != nil {
		return km
	}

	var userBindings map[Action][]int32
	if err := json.Unmarshal(data, &userBindings); err != nil {
		return km
	}

	// Merge user bindings over defaults
	for action, keys := range userBindings {
		km.Bindings[action] = keys
	}
	return km
}

// SaveDefaultKeyMap writes the default keymap to the config file for editing.
func SaveDefaultKeyMap() error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(configDir, "fsnredux")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	km := DefaultKeyMap()
	data, err := json.MarshalIndent(km.Bindings, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, "keys.json"), data, 0644)
}
