package scene

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// CameraAnimation holds state for smooth camera transitions.
type CameraAnimation struct {
	Active     bool
	From       rl.Vector3
	To         rl.Vector3
	FromTarget rl.Vector3
	ToTarget   rl.Vector3
	Progress   float32
	Duration   float32
}

// Animator handles smooth transitions.
type Animator struct {
	Camera CameraAnimation
}

// NewAnimator creates a new animator.
func NewAnimator() *Animator {
	return &Animator{}
}

// StartCameraMove begins a smooth camera transition.
func (a *Animator) StartCameraMove(fromPos, toPos, fromTarget, toTarget rl.Vector3, duration float32) {
	a.Camera = CameraAnimation{
		Active:     true,
		From:       fromPos,
		To:         toPos,
		FromTarget: fromTarget,
		ToTarget:   toTarget,
		Progress:   0,
		Duration:   duration,
	}
}

// Tick advances animations by dt seconds.
// Returns (currentPos, currentTarget, stillAnimating).
func (a *Animator) Tick(dt float32) (rl.Vector3, rl.Vector3, bool) {
	if !a.Camera.Active {
		return rl.Vector3{}, rl.Vector3{}, false
	}

	a.Camera.Progress += dt / a.Camera.Duration
	if a.Camera.Progress >= 1.0 {
		a.Camera.Progress = 1.0
		a.Camera.Active = false
	}

	// Ease-in-out cubic
	t := easeInOutCubic(a.Camera.Progress)

	pos := lerpVector3(a.Camera.From, a.Camera.To, t)
	target := lerpVector3(a.Camera.FromTarget, a.Camera.ToTarget, t)

	return pos, target, a.Camera.Active
}

// IsAnimating returns true if any animation is in progress.
func (a *Animator) IsAnimating() bool {
	return a.Camera.Active
}

// lerpVector3 linearly interpolates between two vectors.
func lerpVector3(a, b rl.Vector3, t float32) rl.Vector3 {
	return rl.NewVector3(
		a.X+(b.X-a.X)*t,
		a.Y+(b.Y-a.Y)*t,
		a.Z+(b.Z-a.Z)*t,
	)
}

// easeInOutCubic provides smooth acceleration and deceleration.
func easeInOutCubic(t float32) float32 {
	if t < 0.5 {
		return 4 * t * t * t
	}
	return 1 - float32(math.Pow(float64(-2*t+2), 3))/2
}
