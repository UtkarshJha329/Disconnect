package Project

import (
	"Disconnect/Engine"
)

const maxPlatformTimers = 128

const DefaultVelTimerStopDuration = 0.4

var platformVelocityChangeTimer Engine.TimerSystem

func PlatformsInitFunc(world *Engine.World) {
	platformVelocityChangeTimer.InitWithTimers("Platform Velocity Change Timer", maxPlatformTimers)

	world.CreateNewLevelColliderInScene(&world.Scene, 0, 450, 640, 32, Engine.DEFAULT)
	world.CreateNewLevelColliderInScene(&world.Scene, -16, 0, 16, 480, Engine.DEFAULT)
	world.CreateNewLevelColliderInScene(&world.Scene, 639, 0, 16, 480, Engine.DEFAULT)
	world.CreateNewLevelColliderInScene(&world.Scene, 100, 380, 120, 20, Engine.DEFAULT)

	platform2Index := world.CreateNewLevelColliderInScene(&world.Scene, 300, 220, 120, 20, Engine.DEFAULT)
	world.Platforms[platform2Index].Velocity.X = 100.0
	world.Platforms[platform2Index].VelTimerDuration = 2.0
	world.Platforms[platform2Index].VelTimerStopAtEnds = false
	world.Platforms[platform2Index].VelTimerAxis = Engine.VelTimerAxisX
	RegisterPlatformTimer(world.Platforms[platform2Index])

	world.CreateNewLevelColliderInScene(&world.Scene, 100, 280, 120, 20, Engine.ONE_WAY_PLATFORMS)

	platform4Index := world.CreateNewLevelColliderInScene(&world.Scene, 100, 200, 120, 20, Engine.ONE_WAY_PLATFORMS)
	world.Platforms[platform4Index].Velocity.Y = -400.0
	world.Platforms[platform4Index].VelTimerDuration = 0.5
	world.Platforms[platform4Index].VelTimerStopAtEnds = false
	world.Platforms[platform4Index].VelTimerAxis = Engine.VelTimerAxisY
	RegisterPlatformTimer(world.Platforms[platform4Index])

	world.Platforms[platform4Index].StartX = 100.0
	world.Platforms[platform4Index].StartY = 200.0
	world.Platforms[platform4Index].StartVelX = 0.0
	world.Platforms[platform4Index].StartVelY = -400.0
	RegisterPlatformTimer(world.Platforms[platform4Index])
}

// RegisterPlatformTimer wires up a velocity-reversal timer for p.
// Call this whenever VelTimerDuration, VelTimerStopAtEnds, VelTimerStopDuration,
// or VelTimerAxis change. It kills any existing timer first.
func RegisterPlatformTimer(p *Engine.Platform) {
	// Kill any old timer
	if p.VelTimerPoolItem() != nil {
		platformVelocityChangeTimer.KillTimer(p.VelTimerPoolItem())
		p.SetVelTimerPoolItem(nil)
	}

	if p.VelTimerDuration <= 0 {
		return
	}

	// onReversal is called each time the timer fires.
	// When stop-at-ends is enabled we zero the velocity, schedule a short pause
	// timer, then restore and negate the velocity once the pause ends.
	onReversal := func() {
		if !p.VelTimerStopAtEnds {
			reverseVelocity(p)
			return
		}

		// Capture the current velocity so the closure can restore it.
		savedVX := p.Velocity.X
		savedVY := p.Velocity.Y

		// Zero out the relevant axes so the platform stops at this end.
		switch p.VelTimerAxis {
		case Engine.VelTimerAxisX:
			p.Velocity.X = 0
		case Engine.VelTimerAxisY:
			p.Velocity.Y = 0
		case Engine.VelTimerAxisBoth:
			p.Velocity.X = 0
			p.Velocity.Y = 0
		}

		// After the pause, restore and reverse.
		stopDur := p.VelTimerStopDuration
		if stopDur <= 0 {
			stopDur = DefaultVelTimerStopDuration
		}
		platformVelocityChangeTimer.SetTimerFromPoolWithDurationLoopAndFunc(
			stopDur,
			false,
			func() {
				switch p.VelTimerAxis {
				case Engine.VelTimerAxisX:
					p.Velocity.X = -savedVX
				case Engine.VelTimerAxisY:
					p.Velocity.Y = -savedVY
				case Engine.VelTimerAxisBoth:
					p.Velocity.X = -savedVX
					p.Velocity.Y = -savedVY
				}
			},
		)
	}

	item := platformVelocityChangeTimer.SetTimerFromPoolWithDurationLoopAndFunc(
		p.VelTimerDuration,
		true,
		onReversal,
	)
	p.SetVelTimerPoolItem(item)
}

// reverseVelocity negates the platform velocity on the configured axis.
func reverseVelocity(p *Engine.Platform) {
	switch p.VelTimerAxis {
	case Engine.VelTimerAxisX:
		p.Velocity.X *= -1
	case Engine.VelTimerAxisY:
		p.Velocity.Y *= -1
	case Engine.VelTimerAxisBoth:
		p.Velocity.X *= -1
		p.Velocity.Y *= -1
	}
}

func PlatformsUpdateFunc(world *Engine.World, dt float64) {
	platformVelocityChangeTimer.UpdateAllTimerDeltasAndStates(dt)
	for e, platform := range world.Platforms {
		platform.Move(e, platform.Velocity.X*dt, platform.Velocity.Y*dt, world)
	}
}
