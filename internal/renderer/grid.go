package renderer

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

// Ground color - dark muted teal-gray to complement the scene.
var groundColor = rl.NewColor(22, 38, 36, 255)

// DrawGround renders the green ground plane (matching fsnav draw_env).
func DrawGround() {
	// Slightly below y=0 to avoid z-fighting with pedestal bottoms
	rl.DrawPlane(
		rl.NewVector3(0, -0.01, 0),
		rl.NewVector2(1000, 1000),
		groundColor,
	)
}
