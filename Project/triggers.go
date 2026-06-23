package Project

import (
	"Disconnect/Engine"
)

var playerRespawnTimerSystem Engine.TimerSystem

func SetupCheckpointCallback(trigger *Engine.Trigger, triggerEntity Engine.Entity) {
	trigger.TriggerType = "checkpoint"
	trigger.OnCollision = func(triggerEntity, colliderEntity Engine.Entity, trigger *Engine.Trigger) {
		if colliderEntity == player.Entity {
			if cp, exists := Engine.WorldInstance.Checkpoints[triggerEntity]; exists {
				cp.Activated = true
			}
		}
	}
}

func SetupKillTriggerCallback(trigger *Engine.Trigger) {
	trigger.TriggerType = "kill"
	trigger.OnCollision = func(triggerEntity, colliderEntity Engine.Entity, trigger *Engine.Trigger) {
		if colliderEntity == player.Entity {
			Engine.WorldInstance.KillEntity(colliderEntity)

			ps := Engine.WorldInstance.ParticleSystems[colliderEntity]
			ps.SpawnBurst(Engine.WorldInstance, 40, &CelesteDeathConfig)
			ps.SpawnBurst(Engine.WorldInstance, 15, &CelesteDeathSparkConfig)

			var respawnPos Engine.Vector2 = Engine.Vector2{X: 320.0, Y: 240.0}
			for _, cp := range Engine.WorldInstance.Checkpoints {
				if cp.Activated {
					respawnPos = cp.Position
				}
			}

			playerRespawnTimerSystem.SetTimerFromPoolWithDurationLoopAndFunc(
				2.0,
				false,
				func() {
					Engine.WorldInstance.Transforms[colliderEntity].Position = Engine.Vector3{
						X: respawnPos.X,
						Y: respawnPos.Y,
						Z: Engine.WorldInstance.Transforms[colliderEntity].Position.Z,
					}
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
		nil,
	)
	SetupKillTriggerCallback(world.Triggers[triggerEntity])
}

func TriggersUpdateFunc(world *Engine.World, dt float64) {
	world.PerformTriggerCharacterColliderChecks()
	playerRespawnTimerSystem.UpdateAllTimerDeltasAndStates(dt)
}
