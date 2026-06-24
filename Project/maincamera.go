package Project

import (
	"Disconnect/Engine"
	"math"
)

var MainCameraEntity Engine.Entity

func MainCameraInitFunc(world *Engine.World) {
	MainCameraEntity = world.NewCameraInScene(&world.Scene, 640, 480)
	world.Transforms[MainCameraEntity].Position = Engine.Vector3{X: 320.0, Y: 240.0, Z: 100.0}
}

func MainCameraUpdateFunc(world *Engine.World, dt float64) {
	if !world.Alive[player.Entity] {
		return
	}

	playerPos := world.Transforms[player.Entity].Position
	camTrans := world.Transforms[MainCameraEntity]

	// 1. Figure out which 640x480 grid cell the player is currently in
	targetRoomX := math.Floor(playerPos.X / 640.0)
	targetRoomY := math.Floor(playerPos.Y / 480.0)

	// 2. Calculate the exact center coordinates of that grid cell
	targetCamX := targetRoomX*640.0 + 320.0
	targetCamY := targetRoomY*480.0 + 240.0

	// 3. Smoothly move the camera towards the center of the room
	lerpFactor := 1.0 - math.Pow(0.001, dt)

	camTrans.Position.X += (targetCamX - camTrans.Position.X) * lerpFactor
	camTrans.Position.Y += (targetCamY - camTrans.Position.Y) * lerpFactor
}

type MainCameraEntityTransformPair struct {
	Entity    Engine.Entity
	Transform *Engine.Transform
	Camera    *Engine.Camera
}
