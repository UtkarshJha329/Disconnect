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

	// Calculate which room the player is in
	world.CurrentRoom = Engine.GetRoomKeyFromPosition(playerPos.X, playerPos.Y)

	// Get exact camera coordinates for this room
	targetX, targetY := Engine.GetRoomCameraPosition(world.CurrentRoom)
	camTrans := world.Transforms[MainCameraEntity]

	// Smoothly snap camera to room
	lerpFactor := 1.0 - math.Pow(0.0001, dt)
	camTrans.Position.X += (targetX - camTrans.Position.X) * lerpFactor
	camTrans.Position.Y += (targetY - camTrans.Position.Y) * lerpFactor
}

type MainCameraEntityTransformPair struct {
	Entity    Engine.Entity
	Transform *Engine.Transform
	Camera    *Engine.Camera
}
