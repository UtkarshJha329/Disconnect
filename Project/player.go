package Project

import (
	"Disconnect/Engine"
	"fmt"
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

	FacingDirection int

	InputReleaseLimit int
	RemainingReleases int
	IsFrozen          bool

	playerIdleAnimationIndex      int
	playerJumpAnimationIndex      int
	playerRunAnimationIndex       int
	playerDashAnimationIndex      int
	playerWallSlideAnimationIndex int

	lastTouchedCheckPointEntity *Engine.Entity
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

	if player.IsFrozen {
		player.JumpJustPressed = false
		player.JumpJustReleased = false
		player.DashJustPressed = false
		player.DownJustPressed = false
		player.DownJustReleased = false
		player.LeftHeld = false
		player.RightHeld = false
		player.UpHeld = false
		player.DownHeld = false
		return
	}

	trackedKeys := []ebiten.Key{
		ebiten.KeyArrowLeft,
		ebiten.KeyArrowRight,
		ebiten.KeyArrowUp,
		ebiten.KeyC,
		ebiten.KeyArrowDown,
		ebiten.KeyX,
	}

	for _, key := range trackedKeys {
		if inpututil.IsKeyJustReleased(key) {
			player.RemainingReleases--
			if player.RemainingReleases <= 0 {
				player.RemainingReleases = 0
				player.IsFrozen = true
				break
			}
		}
	}

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
		player.FacingDirection = -1
	} else if player.RightHeld {
		player.CC().Velocity.X = player.CC().MoveSpeed
		player.FacingDirection = 1
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

	if cc.Velocity.X == 0 {
		playerAnimatedSprite := player.World.AnimatedSprites[player.Entity]
		playerAnimatedSprite.ChangeCurrentAnimationToAnimationIndex(player.playerIdleAnimationIndex)
	} else {
		playerAnimatedSprite := player.World.AnimatedSprites[player.Entity]
		playerAnimatedSprite.ChangeCurrentAnimationToAnimationIndex(player.playerRunAnimationIndex)
	}

	if player.JumpJustPressed && !player.JumpBufferTimer.IsTimerSystemPoolFilled() {
		player.JumpBufferTimer.SetTimerFromPoolWithDurationLoopAndFunc(
			player.JumpBufferTime, false, func() {},
		)
	}

	if player.JumpBufferTimer.TimerSystemPoolHasRunningTimers() {
		cc.Velocity.Y = -cc.JumpSpeed

		ps := player.World.ParticleSystems[player.Entity]
		ps.Offset = Engine.Vector2{X: 0, Y: player.CC().Height / 2.0}
		ps.SpawnBurst(player.World, 4, &CelesteJumpDustConfig)
		ps.Offset = Engine.Vector2{X: 0, Y: 0}

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

type AirborneState struct {
	dustParticleSizeOffset float64
}

func (state *AirborneState) Name() string         { return "Airborne" }
func (state *AirborneState) Enter(player *Player) {}

func (state *AirborneState) Update(player *Player, dt float64) {
	cc := player.CC()
	player.ApplyHorizontalInput()

	playerAnimatedSprite := player.World.AnimatedSprites[player.Entity]
	playerAnimatedSprite.ChangeCurrentAnimationToAnimationIndex(player.playerJumpAnimationIndex)

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

	CelesteJumpTrailConfig.StartSize -= 2.0
	ps := player.World.ParticleSystems[player.Entity]
	ps.Offset = Engine.Vector2{X: 0, Y: player.CC().Height / 2.0}
	ps.SpawnBurst(player.World, 2, &CelesteJumpTrailConfig)
	ps.Offset = Engine.Vector2{X: 0, Y: 0}

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

func (state *AirborneState) Exit(player *Player) {
	CelesteJumpTrailConfig.StartSize = 16.0
}

type WallSlideState struct{}

func (state *WallSlideState) Name() string         { return "WallSlide" }
func (state *WallSlideState) Enter(player *Player) {}

func (state *WallSlideState) Update(player *Player, dt float64) {
	cc := player.CC()
	player.ApplyHorizontalInput()

	playerAnimatedSprite := player.World.AnimatedSprites[player.Entity]
	playerAnimatedSprite.ChangeCurrentAnimationToAnimationIndex(player.playerWallSlideAnimationIndex)

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
		ps.SpawnBurst(player.World, 15, &CelesteDashBurstConfig)
		state.hasSpawnedBurst = true
	}

	playerAnimatedSprite := player.World.AnimatedSprites[player.Entity]
	playerAnimatedSprite.ChangeCurrentAnimationToAnimationIndex(player.playerDashAnimationIndex)

	// 2. Spawn the after-image trail EVERY frame
	// We only need 1 or 2 particles per frame to make a solid trail
	ps := player.World.ParticleSystems[player.Entity]
	ps.SpawnBurst(player.World, 2, &CelesteDashTrailConfig)
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
var playerRespawnTimerSystem Engine.TimerSystem

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

		InputReleaseLimit: 10,
		RemainingReleases: 10,
		IsFrozen:          false,
	}

	player.JumpBufferTimer.InitWithTimers("Jump Buffer Timer System", 4)
	player.CoyoteTimer.InitWithTimers("Coyote Buffer Timer System", 1)
	player.WallJumpLockTimer.InitWithTimers("Wall Jump Input Lock Timer System", 2)
	player.DashTimer.InitWithTimers("Player Dash Timer System", 1)

	screenWidth, screenHeight := ebiten.WindowSize()

	player.Entity = world.CreateSpriteFromFileInScene(&world.Scene, "Assets/player/idle/player_idle_sheet.png")
	world.Transforms[player.Entity].Scale = Engine.Vector2{X: 2.0, Y: 2.0}
	world.Transforms[player.Entity].Position = Engine.Vector3{
		X: float64(screenWidth) / 2.0,
		Y: float64(screenHeight) / 2.0,
		Z: 100,
	}

	world.AnimatedSprites[player.Entity] = &Engine.AnimatedSprite{
		CurAnimationIndex:  0,
		TotalNumAnimations: 0,
		Animations:         make(map[int]*Engine.Animation),
	}

	playerAnimatedSprite := world.AnimatedSprites[player.Entity]
	player.playerIdleAnimationIndex = playerAnimatedSprite.CreateNewAnimation(
		"Assets/player/idle/player_idle_sheet.png",
		22,
		Engine.Vector2{X: 14, Y: 18},
		true,
		2.20,
	)

	player.playerRunAnimationIndex = playerAnimatedSprite.CreateNewAnimation(
		"Assets/player/run/player_run_sheet.png",
		8,
		Engine.Vector2{X: 14, Y: 18},
		true,
		0.8,
	)

	player.playerJumpAnimationIndex = playerAnimatedSprite.CreateNewAnimation(
		"Assets/player/jump/0.png",
		1,
		Engine.Vector2{X: 14, Y: 18},
		true,
		1000.0,
	)

	player.playerDashAnimationIndex = playerAnimatedSprite.CreateNewAnimation(
		"Assets/player/slide/0.png",
		1,
		Engine.Vector2{X: 14, Y: 18},
		true,
		1000.0,
	)

	player.playerWallSlideAnimationIndex = playerAnimatedSprite.CreateNewAnimation(
		"Assets/player/wall_slide/0.png",
		1,
		Engine.Vector2{X: 14, Y: 18},
		true,
		1000.0,
	)

	// playerAnimatedSprite.CurAnimationIndex = player.playerIdleAnimationIndex

	playerImageSize := playerAnimatedSprite.Animations[player.playerIdleAnimationIndex].FrameSize
	world.Transforms[player.Entity].Pivot = Engine.Vector2{
		X: float64(playerImageSize.X) / 2.0,
		Y: float64(playerImageSize.Y) / 2.0,
	}

	physicsWidth := float64(playerImageSize.X) * world.Transforms[player.Entity].Scale.X
	physicsHeight := float64(playerImageSize.Y) * world.Transforms[player.Entity].Scale.Y

	world.CharacterControllers[player.Entity] = &Engine.CharacterController{
		Width:                     physicsWidth / 1.50,
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

	player.lastTouchedCheckPointEntity = nil
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

		world.Transforms[player.Entity].Scale.X = math.Copysign(
			math.Abs(world.Transforms[player.Entity].Scale.X),
			float64(player.FacingDirection),
		)

		if player.IsFrozen {
			player.KillPlayer()
		}

	} else {
		player.ForceEndAllTimers()
		player.CC().Velocity = Engine.Vector2{}
	}
}

// Player Particle Effects Configs

// CelesteDashTrailConfig leaves a static, fading after-image behind the player
var CelesteDashTrailConfig = Engine.ParticleEmissionConfig{
	ConfigName:        "Dash Trail",
	EmissionRadius:    8.0,
	SpeedMin:          0.0,
	SpeedMax:          0.0,
	LifeMin:           0.1,
	LifeMax:           0.25,
	StartSize:         16.0,
	EndSize:           0.0,
	StartColor:        color.RGBA{R: 255, G: 40, B: 80, A: 180},
	EndColor:          color.RGBA{R: 255, G: 40, B: 80, A: 0},
	Gravity:           0.0,
	Drag:              0.0,
	TotalNumParticles: 50,
}

// Optional: A few sparks that fly OUTWARD when the dash starts
var CelesteDashBurstConfig = Engine.ParticleEmissionConfig{
	ConfigName:        "Dash Burst",
	EmissionRadius:    4.0,
	SpeedMin:          100.0,
	SpeedMax:          300.0,
	LifeMin:           0.2,
	LifeMax:           0.4,
	StartSize:         4.0,
	EndSize:           0.0,
	StartColor:        color.RGBA{R: 255, G: 255, B: 255, A: 255},
	EndColor:          color.RGBA{R: 255, G: 40, B: 80, A: 0},
	Gravity:           0.0,
	Drag:              0.1,
	TotalNumParticles: 50,
}

// CelesteDeathConfig is the main red shatter effect
var CelesteDeathConfig = Engine.ParticleEmissionConfig{
	ConfigName:        "Death",
	EmissionRadius:    4.0,
	SpeedMin:          150.0,
	SpeedMax:          350.0,
	LifeMin:           0.6,
	LifeMax:           1.2,
	StartSize:         6.0,
	EndSize:           0.0,
	StartColor:        color.RGBA{R: 255, G: 30, B: 80, A: 255},
	EndColor:          color.RGBA{R: 255, G: 30, B: 80, A: 0},
	Gravity:           0.0,
	Drag:              0.08,
	TotalNumParticles: 50,
}

// CelesteDeathSparkConfig is a subtle white/yellow flash for extra impact
var CelesteDeathSparkConfig = Engine.ParticleEmissionConfig{
	ConfigName:        "Death Spark",
	EmissionRadius:    2.0,
	SpeedMin:          250.0,
	SpeedMax:          500.0,
	LifeMin:           0.2,
	LifeMax:           0.4,
	StartSize:         4.0,
	EndSize:           0.0,
	StartColor:        color.RGBA{R: 255, G: 255, B: 200, A: 255},
	EndColor:          color.RGBA{R: 255, G: 255, B: 200, A: 0},
	Gravity:           0.0,
	Drag:              0.15,
	TotalNumParticles: 50,
}

// CelesteJumpTrailConfig leaves static, fading clouds behind the player
var CelesteJumpTrailConfig = Engine.ParticleEmissionConfig{
	ConfigName:        "Jump Trail",
	EmissionRadius:    6.0,
	SpeedMin:          0.0,
	SpeedMax:          0.0,
	LifeMin:           0.25,
	LifeMax:           0.25,
	StartSize:         16.0,
	EndSize:           0.0,
	StartColor:        color.RGBA{R: 230, G: 230, B: 230, A: 200},
	EndColor:          color.RGBA{R: 20, G: 20, B: 20, A: 0},
	Gravity:           0.0,
	Drag:              0.0,
	TotalNumParticles: 50,
}

// CelesteJumpDustConfig creates a gentle puff of dust at the player's feet
var CelesteJumpDustConfig = Engine.ParticleEmissionConfig{
	ConfigName:        "Jump Dust",
	EmissionRadius:    4.0,
	SpeedMin:          40.0,
	SpeedMax:          120.0,
	LifeMin:           0.15,
	LifeMax:           0.3,
	StartSize:         3.0,
	EndSize:           8.0,
	StartColor:        color.RGBA{R: 230, G: 230, B: 230, A: 200},
	EndColor:          color.RGBA{R: 20, G: 20, B: 20, A: 0},
	Gravity:           0.0,
	Drag:              0.25,
	TotalNumParticles: 50,
}

func (player *Player) KillPlayer() {
	player.World.KillEntity(player.Entity)

	ps := player.World.ParticleSystems[player.Entity]
	ps.SpawnBurst(player.World, 40, &CelesteDeathConfig)
	ps.SpawnBurst(player.World, 15, &CelesteDeathSparkConfig)

	var respawnPos Engine.Vector2 = Engine.Vector2{X: 320.0, Y: 240.0}
	var respawnWithInputAmount int = player.InputReleaseLimit
	if player.lastTouchedCheckPointEntity != nil {
		respawnPos = player.World.Checkpoints[*player.lastTouchedCheckPointEntity].Position
		respawnWithInputAmount = player.World.Checkpoints[*player.lastTouchedCheckPointEntity].InputReleaseLimit
	}

	playerRespawnTimerSystem.SetTimerFromPoolWithDurationLoopAndFunc(
		2.0,
		false,
		func() {
			player.World.Transforms[player.Entity].Position = Engine.Vector3{
				X: respawnPos.X,
				Y: respawnPos.Y,
				Z: player.World.Transforms[player.Entity].Position.Z,
			}
			player.InputReleaseLimit = respawnWithInputAmount
			player.RemainingReleases = respawnWithInputAmount
			player.IsFrozen = false
			player.World.SetEntityAlive(player.Entity)
		},
	)
}

func GetPlayerReleaseText() string {
	return fmt.Sprintf("%d/%d", player.RemainingReleases, player.InputReleaseLimit)
}
