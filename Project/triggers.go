package Project

import (
	"Disconnect/Engine"
	"fmt"
)

var playerRespawnTimerSystem Engine.TimerSystem

func TriggersInitFunc(world *Engine.World) {

	playerRespawnTimerSystem.InitWithTimers("Player Respawn Timer System", 1)

	world.CreateNewTriggerColliderInScene(
		&world.Scene,
		500.0,
		418.0,
		32.0,
		32.0,
		Engine.DEFAULT,
		func(triggerEntity, colliderEntity Engine.Entity, trigger *Engine.Trigger) {
			if colliderEntity == player.Entity {
				fmt.Println("Collided with player entity : ", colliderEntity)
				world.KillEntity(colliderEntity)

				playerRespawnTimerSystem.SetTimerFromPoolWithDurationLoopAndFunc(
					2.0,
					false,
					func() {

						mainCamera := world.Cameras[MainCameraEntity]

						newPlayerPos := mainCamera.GetScreenCenterInWorld(MainCameraEntity, world)
						world.Transforms[colliderEntity].Position = Engine.Vector3{X: newPlayerPos.X, Y: newPlayerPos.Y, Z: world.Transforms[colliderEntity].Position.Z}
						world.SetEntityAlive(colliderEntity)
					},
				)
			}
		},
	)
}

func TriggersUpdateFunc(world *Engine.World, dt float64) {
	world.PerformTriggerCharacterColliderChecks()
	playerRespawnTimerSystem.UpdateAllTimerDeltasAndStates(dt)
}
