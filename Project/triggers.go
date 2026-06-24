package Project

import (
	"Disconnect/Engine"
)

func SetupCheckpointCallback(trigger *Engine.Trigger, triggerEntity Engine.Entity) {
	trigger.TriggerType = "checkpoint"
	trigger.OnCollision = func(triggerEntity, colliderEntity Engine.Entity, trigger *Engine.Trigger) {
		if colliderEntity == player.Entity {
			if cp, exists := player.World.Checkpoints[triggerEntity]; exists {
				cp.Activated = true
				player.lastTouchedCheckPointEntity = &cp.Entity
				player.InputReleaseLimit = player.World.Checkpoints[cp.Entity].InputReleaseLimit
				player.RemainingReleases = player.World.Checkpoints[cp.Entity].InputReleaseLimit
				player.IsFrozen = false
			}
		}
	}
}

func SetupKillTriggerCallback(trigger *Engine.Trigger) {
	trigger.TriggerType = "kill"
	trigger.OnCollision = func(triggerEntity, colliderEntity Engine.Entity, trigger *Engine.Trigger) {
		if colliderEntity == player.Entity {
			player.KillPlayer()
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
