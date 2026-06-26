package Project

import (
	"Disconnect/Engine"
)

const maxPlatformTimers = 128

const DefaultVelTimerStopDuration = 0.4

var platformVelocityChangeTimer Engine.TimerSystem

func PlatformsInitFunc(world *Engine.World) {
	platformVelocityChangeTimer.InitWithTimers("Platform Velocity Change Timer", maxPlatformTimers)
}

func RegisterPlatformTimer(p *Engine.Platform) {
	if p.VelTimerPoolItem() != nil {
		platformVelocityChangeTimer.KillTimer(p.VelTimerPoolItem())
		p.SetVelTimerPoolItem(nil)
	}

	if p.VelTimerDuration <= 0 {
		return
	}

	onReversal := func() {
		if !p.VelTimerStopAtEnds {
			reverseVelocity(p)
			return
		}

		savedVX := p.Velocity.X
		savedVY := p.Velocity.Y

		switch p.VelTimerAxis {
		case Engine.VelTimerAxisX:
			p.Velocity.X = 0
		case Engine.VelTimerAxisY:
			p.Velocity.Y = 0
		case Engine.VelTimerAxisBoth:
			p.Velocity.X = 0
			p.Velocity.Y = 0
		}

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
