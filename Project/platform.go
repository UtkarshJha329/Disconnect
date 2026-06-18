package Project

import (
	"Disconnect/Engine"
	"time"
)

var platformVelocityChangeTimer Engine.TimerSystem

func PlatformsInitFunc(world *Engine.World) {
	platformVelocityChangeTimer.InitWithTimers("Platform Velocity Change Timer", 10)

	world.CreateNewLevelColliderInScene(&world.Scene, 0, 450, 640, 32, Engine.DEFAULT)
	world.CreateNewLevelColliderInScene(&world.Scene, -16, 0, 16, 480, Engine.DEFAULT)
	world.CreateNewLevelColliderInScene(&world.Scene, 639, 0, 16, 480, Engine.DEFAULT)
	world.CreateNewLevelColliderInScene(&world.Scene, 100, 380, 120, 20, Engine.DEFAULT)
	platform2Index := world.CreateNewLevelColliderInScene(&world.Scene, 300, 220, 120, 20, Engine.DEFAULT)
	world.Platforms[platform2Index].Velocity.X = 100.0
	platformVelocityChangeTimer.SetTimerFromPoolWithDurationLoopAndFunc(
		time.Duration(2.0*float64(time.Second)),
		true,
		func() {
			world.Platforms[platform2Index].Velocity.X *= -1
		},
	)

	//world.CreateNewLevelColliderInScene(&world.Scene, 100, 0, 120, 32, Engine.DEAFULT)
	world.CreateNewLevelColliderInScene(&world.Scene, 100, 280, 120, 20, Engine.ONE_WAY_PLATFORMS)

	platform4Index := world.CreateNewLevelColliderInScene(&world.Scene, 100, 100, 120, 20, Engine.ONE_WAY_PLATFORMS)

	world.Platforms[platform4Index].Velocity.Y = -400.0
	platformVelocityChangeTimer.SetTimerFromPoolWithDurationLoopAndFunc(
		time.Duration(0.5*float64(time.Second)),
		true,
		func() {
			world.Platforms[platform4Index].Velocity.Y *= -1
		},
	)
}

func PlatformsUpdateFunc(world *Engine.World, dt float64) {
	platformVelocityChangeTimer.UpdateAllTimerDeltasAndStates()
	for e, platform := range world.Platforms {
		platform.Move(e, platform.Velocity.X*dt, platform.Velocity.Y*dt, world)
	}
}
