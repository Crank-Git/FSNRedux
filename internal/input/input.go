package input

import (
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/Crank-Git/FSNRedux/internal/scene"
)

// InputState tracks all input handling state.
type InputState struct {
	Camera   *OrbitalCamera
	Picker   *Picker
	Keys     *KeyMap
	ShowHelp bool

	// Double-click tracking (matching fsnav DOUBLE_CLICK_INTERVAL 400ms)
	lastClickTime time.Time
	lastClickX    float32
	lastClickY    float32
	lastClickNode *scene.SceneNode
	doubleClicked bool

	// Drag detection: distinguish click from drag
	leftPressX  float32
	leftPressY  float32
	leftDragged bool

	// Signals to app.go
	ExpandRequested bool   // Enter was pressed on selected dir
	BackRequested   bool   // Escape was pressed
	HomeRequested   bool   // Home key pressed
	SearchRequested bool   // F key pressed
	PathBarRequested bool  // Ctrl+L pressed
	NextNodeRequested bool // Tab pressed
	PrevNodeRequested bool // Shift+Tab pressed
	InspectRequested  bool // Space pressed
	SettingsRequested bool // Comma pressed
	OpenFileRequested bool // O pressed
	BirdseyeRequested bool // B pressed

	// When true, keyboard input goes to a text field - skip camera/shortcut keys
	TextInputActive bool
}

// NewInputState creates the input handler.
func NewInputState() *InputState {
	km := LoadKeyMap()
	cam := NewOrbitalCamera()
	cam.Keys = km
	return &InputState{
		Camera:   cam,
		Picker:   NewPicker(),
		Keys:     km,
		ShowHelp: true,
	}
}

// Update processes all input for a frame. Returns the path that was double-clicked, if any.
func (s *InputState) Update(graph *scene.Graph, sidebarWidth int32) string {
	s.doubleClicked = false
	s.ExpandRequested = false
	s.BackRequested = false
	s.HomeRequested = false
	s.SearchRequested = false
	s.PathBarRequested = false
	s.NextNodeRequested = false
	s.PrevNodeRequested = false
	s.InspectRequested = false
	s.SettingsRequested = false
	s.OpenFileRequested = false
	s.BirdseyeRequested = false

	mousePos := rl.GetMousePosition()
	inViewport := mousePos.X > float32(sidebarWidth)

	if inViewport {
		// Camera always updates (handles animation + user input, matching fsnav)
		s.Camera.Update()

		// Hover: pick on every frame (matching fsnav passive_motion)
		s.Picker.HoveredNode = nil
		if graph != nil {
			s.Picker.HoveredNode = graph.Pick(s.Camera.GetRay())
		}

		// Track left press position for drag detection
		if rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
			s.leftPressX = mousePos.X
			s.leftPressY = mousePos.Y
			s.leftDragged = false
		}
		if rl.IsMouseButtonDown(rl.MouseButtonLeft) {
			dx := mousePos.X - s.leftPressX
			dy := mousePos.Y - s.leftPressY
			if dx*dx+dy*dy > 9 { // 3px threshold
				s.leftDragged = true
			}
		}

		// Click = select on release without drag
		if rl.IsMouseButtonReleased(rl.MouseButtonLeft) && !s.leftDragged {
			if graph != nil && s.Picker.HoveredNode != nil {
				hit := s.Picker.HoveredNode
				now := time.Now()

				dx := mousePos.X - s.lastClickX
				dy := mousePos.Y - s.lastClickY
				if hit == s.lastClickNode &&
					now.Sub(s.lastClickTime) < 400*time.Millisecond &&
					dx*dx+dy*dy < 9 {
					s.doubleClicked = true
				}

				s.lastClickTime = now
				s.lastClickX = mousePos.X
				s.lastClickY = mousePos.Y
				s.lastClickNode = hit
				s.Picker.SelectedNode = hit
			}
		}
	}

	// Keyboard shortcuts (disabled when text input is active)
	if !s.TextInputActive {
		if s.Keys.IsPressed(ActionToggleHelp) {
			s.ShowHelp = !s.ShowHelp
		}
		if s.Keys.IsPressed(ActionBack) {
			s.BackRequested = true
		}
		if s.Keys.IsPressed(ActionExpand) {
			s.ExpandRequested = true
		}
		if s.Keys.IsPressed(ActionHome) {
			s.HomeRequested = true
		}
		if s.Keys.IsPressed(ActionNextNode) {
			s.NextNodeRequested = true
		}
		// Shift+Tab for prev node
		if (rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift)) && rl.IsKeyPressed(rl.KeyTab) {
			s.PrevNodeRequested = true
			s.NextNodeRequested = false // override
		}

		ctrlDown := rl.IsKeyDown(rl.KeyLeftControl) || rl.IsKeyDown(rl.KeyRightControl) ||
			rl.IsKeyDown(rl.KeyLeftSuper) || rl.IsKeyDown(rl.KeyRightSuper)
		if ctrlDown && s.Keys.IsPressed(ActionPathBar) {
			s.PathBarRequested = true
		}
		if !ctrlDown && s.Keys.IsPressed(ActionSearch) {
			s.SearchRequested = true
		}
		if s.Keys.IsPressed(ActionInspect) {
			s.InspectRequested = true
		}
		if s.Keys.IsPressed(ActionSettings) {
			s.SettingsRequested = true
		}
		if s.Keys.IsPressed(ActionOpenFile) {
			s.OpenFileRequested = true
		}
		if s.Keys.IsPressed(ActionBirdseye) {
			s.BirdseyeRequested = true
		}
	}

	// Double-click: navigate to node
	if s.doubleClicked && s.Picker.SelectedNode != nil {
		s.FocusOnNode(s.Picker.SelectedNode)
		if s.Picker.SelectedNode.Entry != nil {
			return s.Picker.SelectedNode.Entry.Path
		}
	}

	return ""
}

// FocusOnNode animates the camera to orbit around a node (matching fsnav cam_focus).
func (s *InputState) FocusOnNode(node *scene.SceneNode) {
	if node == nil {
		return
	}
	s.Camera.AnimateTo(node.Position)
}

// FocusOnPath finds a node by path and animates to it.
func (s *InputState) FocusOnPath(graph *scene.Graph, path string) {
	if graph == nil {
		return
	}
	node := graph.FindByPath(path)
	if node != nil {
		s.Picker.SelectedNode = node
		s.FocusOnNode(node)
	}
}
