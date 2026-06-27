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

				player.CanDash = true
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

func SetUpDashPowerUpTriggerCallback(trigger *Engine.Trigger) {
	trigger.TriggerType = "DashPowerUpTrigger"
	trigger.OnCollision = func(triggerEntity, colliderEntity Engine.Entity, trigger *Engine.Trigger) {
		if colliderEntity == player.Entity {
			player.PickedUpDashPowerUp = true
			player.SpawnPlayerPowerPickUpParticles()
			player.World.RemoveTrigger(triggerEntity)
		}
	}
}

func SetUpWallClimbPowerUpTriggerCallback(trigger *Engine.Trigger) {
	trigger.TriggerType = "WallClimbPowerUpTrigger"
	trigger.OnCollision = func(triggerEntity, colliderEntity Engine.Entity, trigger *Engine.Trigger) {
		if colliderEntity == player.Entity {
			player.PickedUpWallClimbPowerUp = true
			player.SpawnPlayerPowerPickUpParticles()
			player.World.RemoveTrigger(triggerEntity)
		}
	}
}

func TriggersInitFunc(world *Engine.World) {

	// // Dash Power Up Trigger.
	// dashPowerUpTriggerEntity := world.CreateNewTriggerColliderInScene(
	// 	&world.Scene,
	// 	-1560,
	// 	220,
	// 	20.0,
	// 	20.0,
	// 	Engine.DEFAULT,
	// 	func(triggerEntity, colliderEntity Engine.Entity, trigger *Engine.Trigger) {
	// 		if colliderEntity == player.Entity {
	// 			player.PickedUpDashPowerUp = true
	// 		}
	// 	},
	// )
	// world.Triggers[dashPowerUpTriggerEntity].TriggerType = "DashPowerUpTrigger"

	// // Wall Climb Power Up Trigger.
	// wallClimbPowerUpTriggerEntity := world.CreateNewTriggerColliderInScene(
	// 	&world.Scene,
	// 	2900.0,
	// 	400.0,
	// 	20.0,
	// 	20.0,
	// 	Engine.DEFAULT,
	// 	func(triggerEntity, colliderEntity Engine.Entity, trigger *Engine.Trigger) {
	// 		if colliderEntity == player.Entity {
	// 			player.PickedUpWallClimbPowerUp = true
	// 		}
	// 	},
	// )
	// world.Triggers[wallClimbPowerUpTriggerEntity].TriggerType = "WallClimbPowerUpTrigger"

}

func TriggersUpdateFunc(world *Engine.World, dt float64) {
	world.PerformTriggerCharacterColliderChecks()
	playerRespawnTimerSystem.UpdateAllTimerDeltasAndStates(dt)
}
