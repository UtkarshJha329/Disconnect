package Project

import (
	"Disconnect/Engine"
	"fmt"
	"time"
)

var platformVelocityChangeTimer Engine.TimerSystem

func PlatformsInitFunc(world *Engine.World) {
	platformVelocityChangeTimer.InitWithTimers("Platform Velocity Change Timer", 10)

	world.CreateNewLevelColliderInScene(&world.Scene, 0, 450, 640, 32, Engine.DEAFULT)
	world.CreateNewLevelColliderInScene(&world.Scene, -16, 0, 16, 480, Engine.DEAFULT)
	world.CreateNewLevelColliderInScene(&world.Scene, 639, 0, 16, 480, Engine.DEAFULT)
	world.CreateNewLevelColliderInScene(&world.Scene, 100, 380, 120, 20, Engine.DEAFULT)
	platform2Index := world.CreateNewLevelColliderInScene(&world.Scene, 300, 320, 120, 20, Engine.DEAFULT)
	world.Platforms[platform2Index].Velocity.X = 100.0
	platformVelocityChangeTimer.SetTimerFromPoolWithDurationLoopAndFunc(
		time.Duration(2.0*float64(time.Second)),
		true,
		func() {
			world.Platforms[platform2Index].Velocity.X *= -1
			fmt.Println("Changed Platform velocity to : ", world.Platforms[platform2Index].Velocity.X)
		},
	)

	//world.CreateNewLevelColliderInScene(&world.Scene, 100, 0, 120, 32, Engine.DEAFULT)
	world.CreateNewLevelColliderInScene(&world.Scene, 100, 280, 120, 20, Engine.ONE_WAY_PLATFORMS)

	platform4Index := world.CreateNewLevelColliderInScene(&world.Scene, 100, 220, 120, 20, Engine.ONE_WAY_PLATFORMS)

	world.Platforms[platform4Index].Velocity.Y = -100.0
	platformVelocityChangeTimer.SetTimerFromPoolWithDurationLoopAndFunc(
		time.Duration(2.0*float64(time.Second)),
		true,
		func() {
			world.Platforms[platform4Index].Velocity.Y *= -1
			fmt.Println("Changed Platform velocity to : ", world.Platforms[platform2Index].Velocity.X)
		},
	)
}

func PlatformsUpdateFunc(world *Engine.World, dt float64) {
	platformVelocityChangeTimer.UpdateAllTimerDeltasAndStates()
	for e, platform := range world.Platforms {
		platform.Move(e, platform.Velocity.X*dt, platform.Velocity.Y*dt, world)
	}
}
