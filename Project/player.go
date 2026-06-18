package Project

import (
	"Disconnect/Engine"
	"math"
	"time"

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
	JumpBufferTime  time.Duration
	JumpBufferTimer Engine.TimerSystem

	CoyoteTime  time.Duration
	CoyoteTimer Engine.TimerSystem

	WallSlideSpeed    float64
	WallClimbSpeed    float64
	WallJumpSpeedX    float64
	WallJumpSpeedY    float64
	WallJumpLockTime  time.Duration
	WallJumpLockTimer Engine.TimerSystem

	DashSpeed float64
	DashTime  time.Duration
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

func (player *Player) UpdateTimers() {
	player.JumpBufferTimer.UpdateAllTimerDeltasAndStates()
	player.CoyoteTimer.UpdateAllTimerDeltasAndStates()
	player.WallJumpLockTimer.UpdateAllTimerDeltasAndStates()
	player.DashTimer.UpdateAllTimerDeltasAndStates()
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
	gravityBackup float64
}

func (state *DashState) Name() string { return "Dash" }

func (state *DashState) Enter(player *Player) {
	cc := player.CC()
	state.gravityBackup = cc.Gravity
	cc.Gravity = 0.0

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

func (state *DashState) Update(player *Player, dt float64) {}

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

		JumpBufferTime:   time.Duration(0.15 * float64(time.Second)),
		CoyoteTime:       time.Duration(0.15 * float64(time.Second)),
		WallJumpLockTime: time.Duration(0.25 * float64(time.Second)),
		DashTime:         time.Duration(0.15 * float64(time.Second)),
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
	player.UpdateTimers()
	player.RefreshInput()

	player.state.Update(player, dt)

	cc := player.CC()
	player.OldFrameGrounded = cc.Grounded
	cc.UpdateCharacter(player.Entity, world, world.Transforms[player.Entity], dt)

	player.state.PostUpdate(player)
}
