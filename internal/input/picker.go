package input

import (
	"github.com/Crank-Git/FSNRedux/internal/scene"
)

// Picker handles ray-based node selection.
type Picker struct {
	SelectedNode *scene.SceneNode
	HoveredNode  *scene.SceneNode
}

// NewPicker creates a picker.
func NewPicker() *Picker {
	return &Picker{}
}

// ClearSelection deselects the current node.
func (p *Picker) ClearSelection() {
	p.SelectedNode = nil
}
