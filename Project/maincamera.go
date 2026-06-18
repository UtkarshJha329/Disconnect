package Project

import "Disconnect/Engine"

var MainCameraEntity Engine.Entity

func MainCameraInitFunc(world *Engine.World) {
	MainCameraEntity = world.NewCameraInScene(&world.Scene, 640, 480)
	world.Transforms[MainCameraEntity].Position = Engine.Vector3{X: 320.0, Y: 240, Z: 100.0}
}

func MainCameraUpdateFunc(world *Engine.World, dt float64) {
	// world.Cameras[MainCameraEntity].FollowEntity(MainCameraEntity, player.Entity, world)
}

type MainCameraEntityTransformPair struct {
	Entity    Engine.Entity
	Transform *Engine.Transform
	Camera    *Engine.Camera
}
