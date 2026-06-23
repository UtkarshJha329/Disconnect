package Project

import (
	"Disconnect/Engine"
)

var playerRespawnTimerSystem Engine.TimerSystem

// SetupKillTriggerCallback assigns the kill-player behavior to a trigger.
// Call this after creating or loading a trigger that should kill on contact.
func SetupKillTriggerCallback(trigger *Engine.Trigger) {
	trigger.TriggerType = "kill"
	trigger.OnCollision = func(triggerEntity, colliderEntity Engine.Entity, trigger *Engine.Trigger) {
		if colliderEntity == player.Entity {
			Engine.WorldInstance.KillEntity(colliderEntity)

			ps := Engine.WorldInstance.ParticleSystems[colliderEntity]
			ps.SpawnBurst(Engine.WorldInstance, 40, &CelesteDeathConfig)
			ps.SpawnBurst(Engine.WorldInstance, 15, &CelesteDeathSparkConfig)

			playerRespawnTimerSystem.SetTimerFromPoolWithDurationLoopAndFunc(
				2.0,
				false,
				func() {
					mainCamera := Engine.WorldInstance.Cameras[MainCameraEntity]

					newPlayerPos := mainCamera.GetScreenCenterInWorld(MainCameraEntity, Engine.WorldInstance)
					Engine.WorldInstance.Transforms[colliderEntity].Position = Engine.Vector3{X: newPlayerPos.X, Y: newPlayerPos.Y, Z: Engine.WorldInstance.Transforms[colliderEntity].Position.Z}
					Engine.WorldInstance.SetEntityAlive(colliderEntity)
				},
			)
		}
	}
}

func TriggersInitFunc(world *Engine.World) {
	playerRespawnTimerSystem.InitWithTimers("Player Respawn Timer System", 1)

	triggerEntity := world.CreateNewTriggerColliderInScene(
		&world.Scene,
		500.0,
		418.0,
		32.0,
		32.0,
		Engine.DEFAULT,
		nil, // callback set below
	)
	SetupKillTriggerCallback(world.Triggers[triggerEntity])
}

func TriggersUpdateFunc(world *Engine.World, dt float64) {
	world.PerformTriggerCharacterColliderChecks()
	playerRespawnTimerSystem.UpdateAllTimerDeltasAndStates(dt)
}
