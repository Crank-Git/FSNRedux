package input

import (
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// OrbitalCamera matches fsnav's orbital camera exactly.
// Left-drag: rotate, Right-drag: zoom, Middle-drag: Y-adjust, Scroll: zoom.
type OrbitalCamera struct {
	Camera rl.Camera3D

	Target   rl.Vector3
	Distance float32
	Theta    float32 // horizontal angle (degrees)
	Phi      float32 // vertical angle (degrees), clamped 5-90

	// Animation: lerp orbit center over 0.8s (matching fsnav TRANS_TIME)
	animFrom  rl.Vector3
	animTo    rl.Vector3
	animStart float64
	animating bool

	// When false, skip WASD/arrow/+/- keyboard input (text input active)
	KeyboardEnabled bool

	// Reference to keymap for configurable bindings
	Keys *KeyMap
}

// NewOrbitalCamera creates a camera matching fsnav defaults.
func NewOrbitalCamera() *OrbitalCamera {
	cam := &OrbitalCamera{
		Target:          rl.NewVector3(0, 0, 0),
		Distance:        5,
		Theta:           90, // camera at +Z, looking toward -Z (matches fsnav)
		Phi:             25,
		KeyboardEnabled: true,
	}
	cam.Camera = rl.Camera3D{
		Up:         rl.NewVector3(0, 1, 0),
		Fovy:       50.0, // fsnav uses 50
		Projection: rl.CameraPerspective,
	}
	cam.updatePosition()
	return cam
}

// Update processes mouse input and animation.
func (c *OrbitalCamera) Update() {
	// Animate orbit center
	if c.animating {
		t := float32((rl.GetTime() - c.animStart) / 0.8)
		if t >= 1.0 {
			t = 1.0
			c.animating = false
		}
		c.Target = lerpVec3(c.animFrom, c.animTo, t)
	}

	// Left drag: rotate (matching fsnav: cam_theta += dx * 0.5)
	if rl.IsMouseButtonDown(rl.MouseButtonLeft) {
		delta := rl.GetMouseDelta()
		c.Theta += delta.X * 0.5
		c.Phi += delta.Y * 0.5
		if c.Phi < 5 {
			c.Phi = 5
		}
		if c.Phi > 90 {
			c.Phi = 90
		}
	}

	// Middle drag: Y adjust (matching fsnav: cam_y += (prev_y - y) * 0.1)
	if rl.IsMouseButtonDown(rl.MouseButtonMiddle) {
		delta := rl.GetMouseDelta()
		c.Target.Y -= delta.Y * 0.1
	}

	// Right drag: zoom distance (matching fsnav: cam_dist += (y - prev_y) * 0.1)
	if rl.IsMouseButtonDown(rl.MouseButtonRight) {
		delta := rl.GetMouseDelta()
		c.Distance += delta.Y * 0.1
		if c.Distance < 0.5 {
			c.Distance = 0.5
		}
	}

	// Scroll: zoom (matching fsnav: cam_dist -= 0.5 per step)
	wheel := rl.GetMouseWheelMove()
	if wheel != 0 {
		c.Distance -= wheel * 0.5
		if c.Distance < 0.5 {
			c.Distance = 0.5
		}
	}

	// Keyboard controls (disabled when text input is active)
	if c.KeyboardEnabled && c.Keys != nil {
		// +/- zoom
		if c.Keys.IsDown(ActionZoomIn) {
			c.Distance -= 0.15
			if c.Distance < 0.5 {
				c.Distance = 0.5
			}
		}
		if c.Keys.IsDown(ActionZoomOut) {
			c.Distance += 0.15
		}

		// WASD / arrow keys pan the camera target
		panSpeed := float32(0.15) * c.Distance / 5 // scale with zoom
		theta := float64(c.Theta) * math.Pi / 180
		fwdX := float32(-math.Cos(theta)) * panSpeed
		fwdZ := float32(-math.Sin(theta)) * panSpeed
		rightX := float32(-math.Sin(theta)) * panSpeed
		rightZ := float32(math.Cos(theta)) * panSpeed

		if c.Keys.IsDown(ActionPanForward) {
			c.Target.X += fwdX
			c.Target.Z += fwdZ
		}
		if c.Keys.IsDown(ActionPanBack) {
			c.Target.X -= fwdX
			c.Target.Z -= fwdZ
		}
		if c.Keys.IsDown(ActionPanLeft) {
			c.Target.X -= rightX
			c.Target.Z -= rightZ
		}
		if c.Keys.IsDown(ActionPanRight) {
			c.Target.X += rightX
			c.Target.Z += rightZ
		}
	}

	c.updatePosition()
}

// updatePosition computes camera world position from orbital parameters.
// Matches fsnav's GL transforms: translate(-dist,Z) * rotate(phi,X) * rotate(theta,Y) * translate(-pos)
func (c *OrbitalCamera) updatePosition() {
	theta := float64(c.Theta) * math.Pi / 180
	phi := float64(c.Phi) * math.Pi / 180

	c.Camera.Position = rl.NewVector3(
		c.Target.X+float32(math.Cos(theta)*math.Cos(phi))*c.Distance,
		c.Target.Y+float32(math.Sin(phi))*c.Distance,
		c.Target.Z+float32(math.Sin(theta)*math.Cos(phi))*c.Distance,
	)
	c.Camera.Target = c.Target
}

// AnimateTo smoothly moves the orbit center to a new target (0.8s, matching fsnav).
func (c *OrbitalCamera) AnimateTo(target rl.Vector3) {
	c.animFrom = c.Target
	c.animTo = target
	c.animStart = rl.GetTime()
	c.animating = true
}

// IsAnimating returns true if a camera transition is in progress.
func (c *OrbitalCamera) IsAnimating() bool {
	return c.animating
}

// GetRay returns a picking ray from the current mouse position.
func (c *OrbitalCamera) GetRay() rl.Ray {
	return rl.GetMouseRay(rl.GetMousePosition(), c.Camera)
}

// FrameScene positions the camera to see the entire scene.
func (c *OrbitalCamera) FrameScene(minBounds, maxBounds rl.Vector3) {
	centerX := (minBounds.X + maxBounds.X) / 2
	centerZ := (minBounds.Z + maxBounds.Z) / 2
	c.Target = rl.NewVector3(centerX, 0, centerZ)

	sceneW := maxBounds.X - minBounds.X
	sceneD := maxBounds.Z - minBounds.Z
	extent := sceneW
	if sceneD > extent {
		extent = sceneD
	}
	c.Distance = extent*0.5 + 5
	if c.Distance < 5 {
		c.Distance = 5
	}

	c.Theta = 90
	c.Phi = 25
	c.updatePosition()
}

// Birdseye positions the camera directly overhead looking down at the scene.
func (c *OrbitalCamera) Birdseye(minBounds, maxBounds rl.Vector3) {
	centerX := (minBounds.X + maxBounds.X) / 2
	centerZ := (minBounds.Z + maxBounds.Z) / 2
	c.Target = rl.NewVector3(centerX, 0, centerZ)

	sceneW := maxBounds.X - minBounds.X
	sceneD := maxBounds.Z - minBounds.Z
	extent := sceneW
	if sceneD > extent {
		extent = sceneD
	}
	c.Distance = extent*0.6 + 5
	if c.Distance < 5 {
		c.Distance = 5
	}

	c.Theta = 90
	c.Phi = 85 // near-vertical overhead
	c.updatePosition()
}

func lerpVec3(a, b rl.Vector3, t float32) rl.Vector3 {
	return rl.NewVector3(
		a.X+(b.X-a.X)*t,
		a.Y+(b.Y-a.Y)*t,
		a.Z+(b.Z-a.Z)*t,
	)
}
