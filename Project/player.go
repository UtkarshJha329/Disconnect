package Project

import (
	"Disconnect/Engine"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type PlayerState interface {
	Name() string
	Enter(p *Player)
	Update(p *Player, dt float64)
	PostUpdate(p *Player)
	Exit(p *Player)
}

type Player struct {
	Entity Engine.Entity
	World  *Engine.World

	Gravity   float64
	MoveSpeed float64

	JumpSpeed       float64
	JumpBufferTime  float64
	JumpBufferTimer Engine.TimerSystem

	CoyoteTime  float64
	CoyoteTimer Engine.TimerSystem

	WallSlideSpeed    float64
	WallClimbSpeed    float64
	WallJumpSpeedX    float64
	WallJumpSpeedY    float64
	WallJumpLockTime  float64
	WallJumpLockTimer Engine.TimerSystem

	DashSpeed float64
	DashTime  float64
	DashTimer Engine.TimerSystem

	PlatformMomentumDecay Engine.Vector2

	OldFrameGrounded bool

	JumpJustPressed  bool
	JumpJustReleased bool
	DashJustPressed  bool
	DownJustPressed  bool
	DownJustReleased bool
	LeftHeld         bool
	RightHeld        bool
	UpHeld           bool
	DownHeld         bool

	TouchingWallLeft  bool
	TouchingWallRight bool

	state PlayerState
}

func (player *Player) CC() *Engine.CharacterController {
	return player.World.CharacterControllers[player.Entity]
}

func (player *Player) ChangeState(playerState PlayerState) {
	if player.state != nil {
		player.state.Exit(player)
	}
	player.state = playerState
	playerState.Enter(player)
}

func (player *Player) UpdateTimers(dt float64) {
	player.JumpBufferTimer.UpdateAllTimerDeltasAndStates(dt)
	player.CoyoteTimer.UpdateAllTimerDeltasAndStates(dt)
	player.WallJumpLockTimer.UpdateAllTimerDeltasAndStates(dt)
	player.DashTimer.UpdateAllTimerDeltasAndStates(dt)
}

func (player *Player) ForceEndAllTimers() {
	player.JumpBufferTimer.ForceEndAllTimersForNextUpdate()
	player.CoyoteTimer.ForceEndAllTimersForNextUpdate()
	player.WallJumpLockTimer.ForceEndAllTimersForNextUpdate()
	player.DashTimer.ForceEndAllTimersForNextUpdate()
}

func (player *Player) RefreshInput() {
	player.JumpJustPressed = inpututil.IsKeyJustPressed(ebiten.KeyC)
	player.JumpJustReleased = inpututil.IsKeyJustReleased(ebiten.KeyC)
	player.DashJustPressed = inpututil.IsKeyJustPressed(ebiten.KeyX)
	player.DownJustPressed = inpututil.IsKeyJustPressed(ebiten.KeyArrowDown)
	player.DownJustReleased = inpututil.IsKeyJustReleased(ebiten.KeyArrowDown)
	player.LeftHeld = ebiten.IsKeyPressed(ebiten.KeyArrowLeft)
	player.RightHeld = ebiten.IsKeyPressed(ebiten.KeyArrowRight)
	player.UpHeld = ebiten.IsKeyPressed(ebiten.KeyArrowUp)
	player.DownHeld = ebiten.IsKeyPressed(ebiten.KeyArrowDown)

	cc := player.CC()
	player.TouchingWallLeft = cc.TouchingWall(-1, player.Entity, player.World)
	player.TouchingWallRight = cc.TouchingWall(1, player.Entity, player.World)
}

func (player *Player) ApplyHorizontalInput() {
	if player.LeftHeld {
		player.CC().Velocity.X = -player.CC().MoveSpeed
	} else if player.RightHeld {
		player.CC().Velocity.X = player.CC().MoveSpeed
	} else {
		player.CC().Velocity.X = 0
	}
}

func (player *Player) HandleOneWayPlatform() {
	if player.DownJustPressed {
		player.CC().CollideWithOneWayPlatform = false
	}
	if player.DownJustReleased {
		player.CC().CollideWithOneWayPlatform = true
	}
}

func (player *Player) MustStartDash() bool {
	return player.DashJustPressed && !player.DashTimer.TimerSystemPoolHasRunningTimers()
}

type GroundedState struct{}

func (state *GroundedState) Name() string { return "Grounded" }

func (state *GroundedState) Enter(player *Player) {
	player.JumpBufferTimer.ForceEndAllTimersForNextUpdate()
	player.CoyoteTimer.ForceEndAllTimersForNextUpdate()
}

func (state *GroundedState) Update(player *Player, dt float64) {
	cc := player.CC()
	player.ApplyHorizontalInput()

	if player.JumpJustPressed && !player.JumpBufferTimer.IsTimerSystemPoolFilled() {
		player.JumpBufferTimer.SetTimerFromPoolWithDurationLoopAndFunc(
			player.JumpBufferTime, false, func() {},
		)
	}

	if player.JumpBufferTimer.TimerSystemPoolHasRunningTimers() {
		cc.Velocity.Y = -cc.JumpSpeed
		player.JumpBufferTimer.ForceEndAllTimersForNextUpdate()
		player.ChangeState(&AirborneState{})
		return
	}

	if player.MustStartDash() {
		player.ChangeState(&DashState{})
		return
	}

	player.HandleOneWayPlatform()
}

func (state *GroundedState) PostUpdate(player *Player) {
	cc := player.CC()
	if !cc.Grounded {
		if cc.Velocity.Y >= 0 {
			player.CoyoteTimer.SetTimerFromPoolWithDurationLoopAndFunc(
				player.CoyoteTime, false, func() {},
			)
		}
		player.ChangeState(&AirborneState{})
	}
}

func (state *GroundedState) Exit(player *Player) {}

type AirborneState struct{}

func (state *AirborneState) Name() string         { return "Airborne" }
func (state *AirborneState) Enter(player *Player) {}

func (state *AirborneState) Update(player *Player, dt float64) {
	cc := player.CC()
	player.ApplyHorizontalInput()

	if player.JumpJustPressed && !player.JumpBufferTimer.IsTimerSystemPoolFilled() {
		player.JumpBufferTimer.SetTimerFromPoolWithDurationLoopAndFunc(
			player.JumpBufferTime, false, func() {},
		)
	}

	if player.JumpJustPressed && player.CoyoteTimer.TimerSystemPoolHasRunningTimers() {
		cc.Velocity.Y = -cc.JumpSpeed
		player.CoyoteTimer.ForceEndAllTimersForNextUpdate()
	}

	if player.JumpJustReleased && cc.Velocity.Y < 0 {
		cc.Velocity.Y *= 0.5
	}

	if player.MustStartDash() {
		player.ChangeState(&DashState{})
		return
	}

	player.HandleOneWayPlatform()
}

func (state *AirborneState) PostUpdate(player *Player) {
	cc := player.CC()

	if cc.Grounded {
		player.ChangeState(&GroundedState{})
		return
	}

	if (player.TouchingWallLeft || player.TouchingWallRight) && cc.Velocity.Y > 0 {
		player.ChangeState(&WallSlideState{})
	}
}

func (state *AirborneState) Exit(player *Player) {}

type WallSlideState struct{}

func (state *WallSlideState) Name() string         { return "WallSlide" }
func (state *WallSlideState) Enter(player *Player) {}

func (state *WallSlideState) Update(player *Player, dt float64) {
	cc := player.CC()
	player.ApplyHorizontalInput()

	if player.UpHeld {
		cc.Velocity.Y = -player.WallClimbSpeed
	} else {
		cc.Velocity.Y = math.Min(cc.Velocity.Y, player.WallSlideSpeed)
	}

	if player.JumpJustPressed {
		if player.TouchingWallLeft {
			cc.Velocity.X = player.WallJumpSpeedX
			cc.Velocity.Y = -player.WallJumpSpeedY
			player.ChangeState(&WallJumpState{})
			return
		} else if player.TouchingWallRight {
			cc.Velocity.X = -player.WallJumpSpeedX
			cc.Velocity.Y = -player.WallJumpSpeedY
			player.ChangeState(&WallJumpState{})
			return
		}
	}

	if player.MustStartDash() {
		player.ChangeState(&DashState{})
		return
	}

	player.HandleOneWayPlatform()
}

func (state *WallSlideState) PostUpdate(player *Player) {
	cc := player.CC()

	if cc.Grounded {
		player.ChangeState(&GroundedState{})
		return
	}

	if !player.TouchingWallLeft && !player.TouchingWallRight {
		player.ChangeState(&AirborneState{})
	}
}

func (state *WallSlideState) Exit(player *Player) {}

type WallJumpState struct{}

func (state *WallJumpState) Name() string { return "WallJump" }

func (state *WallJumpState) Enter(player *Player) {
	player.WallJumpLockTimer.SetTimerFromPoolWithDurationLoopAndFunc(
		player.WallJumpLockTime, false, func() {},
	)
}

func (state *WallJumpState) Update(player *Player, dt float64) {
	cc := player.CC()

	if player.JumpJustReleased && cc.Velocity.Y < 0 {
		cc.Velocity.Y *= 0.5
	}

	if player.MustStartDash() {
		player.ChangeState(&DashState{})
		return
	}
}

func (state *WallJumpState) PostUpdate(player *Player) {
	cc := player.CC()

	if cc.Grounded {
		player.ChangeState(&GroundedState{})
		return
	}

	if !player.WallJumpLockTimer.TimerSystemPoolHasRunningTimers() {
		player.ChangeState(&AirborneState{})
	}
}

func (state *WallJumpState) Exit(player *Player) {}

type DashState struct {
	gravityBackup   float64
	hasSpawnedBurst bool
}

func (state *DashState) Name() string { return "Dash" }

func (state *DashState) Enter(player *Player) {
	cc := player.CC()
	state.gravityBackup = cc.Gravity
	cc.Gravity = 0.0
	state.hasSpawnedBurst = false

	directionDash := Engine.Vector2{}
	if player.LeftHeld {
		directionDash.X = -1
	} else if player.RightHeld {
		directionDash.X = 1
	}
	if player.UpHeld {
		directionDash.Y = -1
	} else if player.DownHeld {
		directionDash.Y = 1
	}
	if directionDash.X == 0 && directionDash.Y == 0 {
		directionDash.X = 1
	}
	directionDash = directionDash.Normalize()
	cc.Velocity = directionDash.MultiplyFloat(player.DashSpeed)

	player.DashTimer.SetTimerFromPoolWithDurationLoopAndFunc(
		player.DashTime, false,
		func() {
			cc.Gravity = state.gravityBackup
			cc.Velocity = cc.Velocity.MultiplyFloat(0.3)
		},
	)
}

func (state *DashState) Update(player *Player, dt float64) {
	// 1. Spawn the initial impact burst ONCE
	if !state.hasSpawnedBurst {
		ps := player.World.ParticleSystems[player.Entity]
		ps.SpawnBurst(player.World, 15, CelesteDashBurstConfig)
		state.hasSpawnedBurst = true
	}

	// 2. Spawn the after-image trail EVERY frame
	// We only need 1 or 2 particles per frame to make a solid trail
	ps := player.World.ParticleSystems[player.Entity]
	ps.SpawnBurst(player.World, 2, CelesteDashTrailConfig)
}

func (state *DashState) PostUpdate(player *Player) {

	if player.DashTimer.TimerSystemPoolHasRunningTimers() {
		return
	}

	cc := player.CC()
	switch {
	case cc.Grounded:
		player.ChangeState(&GroundedState{})
	case (player.TouchingWallLeft || player.TouchingWallRight) && cc.Velocity.Y > 0:
		player.ChangeState(&WallSlideState{})
	default:
		player.ChangeState(&AirborneState{})
	}
}

func (state *DashState) Exit(p *Player) {
	p.CC().Gravity = p.Gravity
}

var player *Player

func PlayerInitFunc(world *Engine.World) {
	player = &Player{
		World: world,

		Gravity:        1200,
		MoveSpeed:      500,
		JumpSpeed:      500,
		WallSlideSpeed: 10.0,
		WallClimbSpeed: 100.0,
		WallJumpSpeedX: 400.0,
		WallJumpSpeedY: 600.0,
		DashSpeed:      900.0,

		PlatformMomentumDecay: Engine.Vector2{X: 0.95, Y: 0.95},

		JumpBufferTime:   0.15,
		CoyoteTime:       0.15,
		WallJumpLockTime: 0.25,
		DashTime:         0.15,
	}

	player.JumpBufferTimer.InitWithTimers("Jump Buffer Timer System", 4)
	player.CoyoteTimer.InitWithTimers("Coyote Buffer Timer System", 1)
	player.WallJumpLockTimer.InitWithTimers("Wall Jump Input Lock Timer System", 2)
	player.DashTimer.InitWithTimers("Player Dash Timer System", 1)

	screenWidth, screenHeight := ebiten.WindowSize()

	player.Entity = world.CreateSpriteFromFileInScene(&world.Scene, "Assets/player/idle/00.png")
	world.Transforms[player.Entity].Scale = Engine.Vector2{X: 2.0, Y: 2.0}
	world.Transforms[player.Entity].Position = Engine.Vector3{
		X: float64(screenWidth) / 2.0,
		Y: float64(screenHeight) / 2.0,
		Z: 100,
	}

	playerImageSize := world.Sprites[player.Entity].Tex.Bounds().Size()
	world.Transforms[player.Entity].Pivot = Engine.Vector2{
		X: float64(playerImageSize.X) / 2.0,
		Y: float64(playerImageSize.Y) / 2.0,
	}

	physicsWidth := float64(playerImageSize.X) * world.Transforms[player.Entity].Scale.X
	physicsHeight := float64(playerImageSize.Y) * world.Transforms[player.Entity].Scale.Y

	world.CharacterControllers[player.Entity] = &Engine.CharacterController{
		Width:                     physicsWidth,
		Height:                    physicsHeight,
		Gravity:                   player.Gravity,
		MoveSpeed:                 player.MoveSpeed,
		JumpSpeed:                 player.JumpSpeed,
		Grounded:                  false,
		CollideWithOneWayPlatform: true,
		Velocity:                  Engine.Vector2{X: 0, Y: 0},
		PlatformMomentumDecay:     player.PlatformMomentumDecay,
	}

	world.ParticleSystems[player.Entity] = &Engine.ParticleSystem{}
	world.ParticleSystems[player.Entity] = &Engine.ParticleSystem{
		Owner: player.Entity,
	}
	world.ParticleSystems[player.Entity].InitParticleSystem("Player Death Particles", 1000)

	gunEntity := world.CreateSpriteFromFileInSceneWithParentEntity(&world.Scene, "Assets/gun.png", player.Entity)
	world.Transforms[gunEntity].Scale = Engine.Vector2{X: 1.0, Y: 1.0}
	world.Transforms[gunEntity].Position = Engine.Vector3{X: 6.0, Y: 0.0, Z: 200}

	gunSize := world.Sprites[gunEntity].Tex.Bounds().Size()
	world.Transforms[gunEntity].Pivot = Engine.Vector2{
		X: float64(gunSize.X) / 2.0,
		Y: float64(gunSize.Y) / 2.0,
	}

	player.state = &GroundedState{}
	player.state.Enter(player)
}

func PlayerUpdateFunc(world *Engine.World, dt float64) {

	if world.Alive[player.Entity] {

		player.UpdateTimers(dt)
		player.RefreshInput()

		player.state.Update(player, dt)

		cc := player.CC()
		player.OldFrameGrounded = cc.Grounded
		cc.UpdateCharacter(player.Entity, world, world.Transforms[player.Entity], dt)

		player.state.PostUpdate(player)
	} else {
		player.ForceEndAllTimers()
		player.CC().Velocity = Engine.Vector2{}
	}
}

// Player Particle Effects Configs

// CelesteDashTrailConfig leaves a static, fading after-image behind the player
var CelesteDashTrailConfig = Engine.ParticleConfig{
	EmissionRadius: 8.0, // Spawns roughly within the player's bounding box
	SpeedMin:       0.0, // 0 Speed = They stay exactly where they spawned!
	SpeedMax:       0.0,
	LifeMin:        0.1, // Very short life so the trail drops off quickly
	LifeMax:        0.25,
	StartSize:      16.0,                                     // Roughly the size of the player
	EndSize:        0.0,                                      // Shrinks away
	StartColor:     color.RGBA{R: 255, G: 40, B: 80, A: 180}, // Celeste Red, slightly transparent
	EndColor:       color.RGBA{R: 255, G: 40, B: 80, A: 0},   // Fades to invisible
	Gravity:        0.0,
	Drag:           0.0, // No drag needed since speed is already 0
}

// Optional: A few sparks that fly OUTWARD when the dash starts
var CelesteDashBurstConfig = Engine.ParticleConfig{
	EmissionRadius: 4.0,
	SpeedMin:       100.0,
	SpeedMax:       300.0,
	LifeMin:        0.2,
	LifeMax:        0.4,
	StartSize:      4.0,
	EndSize:        0.0,
	StartColor:     color.RGBA{R: 255, G: 255, B: 255, A: 255}, // White impact sparks
	EndColor:       color.RGBA{R: 255, G: 40, B: 80, A: 0},     // Fade to red
	Gravity:        0.0,
	Drag:           0.1, // High drag so they stop quickly
}

// CelesteDeathConfig is the main red shatter effect
var CelesteDeathConfig = Engine.ParticleConfig{
	EmissionRadius: 4.0,   // Starts very close to the center
	SpeedMin:       150.0, // Fast initial burst
	SpeedMax:       350.0,
	LifeMin:        0.6, // Particles fade out relatively quickly
	LifeMax:        1.2,
	StartSize:      6.0,                                      // Medium sized chunks
	EndSize:        0.0,                                      // Shrink to nothing
	StartColor:     color.RGBA{R: 255, G: 30, B: 80, A: 255}, // Bright Celeste Red/Pink
	EndColor:       color.RGBA{R: 255, G: 30, B: 80, A: 0},   // Fades to transparent
	Gravity:        0.0,                                      // No gravity, they just float and stop
	Drag:           0.08,                                     // High drag so they lose momentum fast (the "shatter" feel)
}

// CelesteDeathSparkConfig is a subtle white/yellow flash for extra impact
var CelesteDeathSparkConfig = Engine.ParticleConfig{
	EmissionRadius: 2.0,
	SpeedMin:       250.0,
	SpeedMax:       500.0,
	LifeMin:        0.2,
	LifeMax:        0.4,
	StartSize:      4.0,
	EndSize:        0.0,
	StartColor:     color.RGBA{R: 255, G: 255, B: 200, A: 255}, // Bright white/yellow
	EndColor:       color.RGBA{R: 255, G: 255, B: 200, A: 0},
	Gravity:        0.0,
	Drag:           0.15, // Even higher drag for sparks
}
