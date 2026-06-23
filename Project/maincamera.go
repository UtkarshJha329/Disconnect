package Project

import (
	"Disconnect/Engine"
	"math"
)

var MainCameraEntity Engine.Entity

func MainCameraInitFunc(world *Engine.World) {
	MainCameraEntity = world.NewCameraInScene(&world.Scene, 640, 480)
	// Initialize camera to the center of the screen so it doesn't start at 0,0
	world.Transforms[MainCameraEntity].Position = Engine.Vector3{X: 320.0, Y: 240.0, Z: 100.0}
}

func MainCameraUpdateFunc(world *Engine.World, dt float64) {
	// Don't move the camera if the player is dead
	if !world.Alive[player.Entity] {
		return
	}

	playerPos := world.Transforms[player.Entity].Position
	camTrans := world.Transforms[MainCameraEntity]

	// Framerate-independent smoothing factor.
	// Decrease the number (e.g., 0.0001) for a SNAPPY camera.
	// Increase the number (e.g., 0.05) for a SLOW/DELAYED camera.
	lerpFactor := 1.0 - math.Pow(0.001, dt)

	// Smoothly move the camera towards the player's X and Y
	camTrans.Position.X += (playerPos.X - camTrans.Position.X) * lerpFactor
	camTrans.Position.Y += (playerPos.Y - camTrans.Position.Y) * lerpFactor

	// Note: We intentionally DO NOT update the Z axis here,
	// so the camera stays at Z: 100 and renders behind the player.
}

type MainCameraEntityTransformPair struct {
	Entity    Engine.Entity
	Transform *Engine.Transform
	Camera    *Engine.Camera
}
