package Project

import (
	"Disconnect/Engine"
	"math"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

var playerEntity Engine.Entity

var playerGravity float64 = 1200

var playerMoveSpeed float64 = 500
var playerJumpSpeed float64 = 500

var jumpBufferTime time.Duration = time.Duration(0.15 * float64(time.Second))
var jumpBufferTimerSystem Engine.TimerSystem

var coyeteTime time.Duration = time.Duration(0.15 * float64(time.Second))
var coyeteBufferTimerSystem Engine.TimerSystem
var oldFrameGrounded bool = false

var playerWallSlideSpeed float64 = 10.0
var playerWallClimbSpeed float64 = 100.0
var playerWallJumpSpeedX float64 = 400.0
var playerWallJumpSpeedY float64 = 600.0
var wallJumpLockInputForSeconds time.Duration = time.Duration(0.25 * float64(time.Second))
var wallJumpInputLockTimerSystem Engine.TimerSystem
var wallJumpInputLock bool = false

var playerDashSpeed float64 = 600.0
var playerDashTime time.Duration = time.Duration(0.15 * float64(time.Second))
var playerDashTimerSystem Engine.TimerSystem
var dashInputLock bool = false
var canDash bool = true

func PlayerInitFunc(world *Engine.World) {

	screenWidth, screenHeight := ebiten.WindowSize()

	jumpBufferTimerSystem.InitWithTimers("Jump Buffer Timer System", 4)
	coyeteBufferTimerSystem.InitWithTimers("Coyete Buffer Timer System", 1)
	wallJumpInputLockTimerSystem.InitWithTimers("Wall Jump Input Lock Timer System", 2)
	playerDashTimerSystem.InitWithTimers("Player Dash Timer System", 1)

	playerEntity = world.CreateSpriteFromFileInScene(&world.Scene, "Assets/player/idle/00.png")
	world.Transforms[playerEntity].Scale = Engine.Vector2{X: 2.0, Y: 2.0}
	world.Transforms[playerEntity].Position = Engine.Vector3{X: float64(screenWidth) / 2.0, Y: float64(screenHeight) / 2.0, Z: 100}

	playerImageSize := world.Sprites[playerEntity].Tex.Bounds().Size()
	world.Transforms[playerEntity].Pivot = Engine.Vector2{X: float64(playerImageSize.X) / 2.0, Y: float64(playerImageSize.Y) / 2.0}

	physicsWidth :=
		float64(playerImageSize.X) *
			world.Transforms[playerEntity].Scale.X

	physicsHeight :=
		float64(playerImageSize.Y) *
			world.Transforms[playerEntity].Scale.Y

	world.CharacterControllers[playerEntity] = &Engine.CharacterController{
		Width:  physicsWidth,
		Height: physicsHeight,

		Gravity:   playerGravity,
		MoveSpeed: playerMoveSpeed,
		JumpSpeed: playerJumpSpeed,

		Grounded:                  false,
		CollideWithOneWayPlatform: true,

		Velocity: Engine.Vector2{
			X: 0,
			Y: 0,
		},
	}

	gunEntity := world.CreateSpriteFromFileInSceneWithParentEntity(&world.Scene, "Assets/gun.png", playerEntity)
	world.Transforms[gunEntity].Scale = Engine.Vector2{X: 1.0, Y: 1.0}
	world.Transforms[gunEntity].Position = Engine.Vector3{X: 6.0, Y: 0.0, Z: 200}

	gunImageSize := world.Sprites[gunEntity].Tex.Bounds().Size()
	world.Transforms[gunEntity].Pivot = Engine.Vector2{X: float64(gunImageSize.X) / 2.0, Y: float64(gunImageSize.Y) / 2.0}

	// rotatePlayerToValue := Engine.DegreesToRadians(360.0)
	// timeToRotatePlayer := 60.0
	// floatInterpolationSystem.CreateNewInterpolation(
	// 	world.Transforms[playerEntity].Rotation,
	// 	&world.Transforms[playerEntity].Rotation,
	// 	&rotatePlayerToValue,
	// 	time.Duration(timeToRotatePlayer*float64(time.Second)),
	// 	true,
	// 	Engine.LinearInterpolationFloat,
	// 	func() { fmt.Println("Finished rotating sprite.") },
	// )

}

func PlayerUpdateFunc(world *Engine.World, dt float64) {
	jumpBufferTimerSystem.UpdateAllTimerDeltasAndStates()
	coyeteBufferTimerSystem.UpdateAllTimerDeltasAndStates()
	wallJumpInputLockTimerSystem.UpdateAllTimerDeltasAndStates()
	playerDashTimerSystem.UpdateAllTimerDeltasAndStates()

	playerCharacterController := world.CharacterControllers[playerEntity]

	jumpJustPressed := inpututil.IsKeyJustPressed(ebiten.KeyC)
	jumpJustReleased := inpututil.IsKeyJustReleased(ebiten.KeyC)

	arrowDownJustPressed := inpututil.IsKeyJustPressed(ebiten.KeyArrowDown)
	arrowDownJustReleased := inpututil.IsKeyJustReleased(ebiten.KeyArrowDown)

	leftArrowHeld := ebiten.IsKeyPressed(ebiten.KeyArrowLeft)
	rightArrowHeld := ebiten.IsKeyPressed(ebiten.KeyArrowRight)
	upArrowHeld := ebiten.IsKeyPressed(ebiten.KeyArrowUp)
	downArrowHeld := ebiten.IsKeyPressed(ebiten.KeyArrowDown)

	dashJustPressed := inpututil.IsKeyJustPressed(ebiten.KeyX)

	touchingWallLeft := playerCharacterController.TouchingWall(-1, playerEntity, world)
	touchingWallRight := playerCharacterController.TouchingWall(1, playerEntity, world)
	touchingWall := touchingWallLeft || touchingWallRight

	if dashJustPressed && !playerDashTimerSystem.IsTimerSystemPoolFilled() {
		dashInputLock = true
		dashDirection := Engine.Vector2{}
		if leftArrowHeld {
			dashDirection.X = -1
		} else if rightArrowHeld {
			dashDirection.X = 1
		}
		if upArrowHeld {
			dashDirection.Y = -1
		} else if downArrowHeld {
			dashDirection.Y = 1
		}

		if dashDirection.X == 0 && dashDirection.Y == 0 {
			dashDirection.X = 1
		}

		playerCharacterController.Gravity = 0.0

		dashDirection = dashDirection.Normalize()

		playerCharacterController.Velocity = dashDirection.MultiplyFloat(playerDashSpeed)

		playerDashTimerSystem.SetTimerFromPoolWithDurationLoopAndFunc(
			playerDashTime,
			false,
			func() {
				dashInputLock = false
				playerCharacterController.Gravity = playerGravity
				playerCharacterController.Velocity = playerCharacterController.Velocity.MultiplyFloat(0.3)
			},
		)
	}

	if !wallJumpInputLock && !dashInputLock {
		if leftArrowHeld {
			playerCharacterController.Velocity.X = -playerCharacterController.MoveSpeed
		} else if rightArrowHeld {
			playerCharacterController.Velocity.X = playerCharacterController.MoveSpeed
		} else {
			playerCharacterController.Velocity.X = 0
		}
	}

	dir := math.Copysign(1.0, playerCharacterController.Velocity.X)
	if !playerCharacterController.Grounded &&
		playerCharacterController.Velocity.Y > 0 &&
		playerCharacterController.TouchingWall(dir, playerEntity, world) {

		playerCharacterController.Velocity.Y = min(playerCharacterController.Velocity.Y, playerWallSlideSpeed)
	}

	if jumpJustPressed {
		if !jumpBufferTimerSystem.IsTimerSystemPoolFilled() {
			jumpBufferTimerSystem.SetTimerFromPoolWithDurationLoopAndFunc(jumpBufferTime, false, func() {})
		}
	}

	if (jumpBufferTimerSystem.TimerSystemPoolHasRunningTimers() && playerCharacterController.Grounded) ||
		(jumpJustPressed && playerCharacterController.Grounded) ||
		(jumpJustPressed && coyeteBufferTimerSystem.TimerSystemPoolHasRunningTimers()) {

		playerCharacterController.Velocity.Y = -playerCharacterController.JumpSpeed
	}

	if jumpJustReleased &&
		playerCharacterController.Velocity.Y < 0 {

		playerCharacterController.Velocity.Y *= 0.5
	}

	if jumpJustPressed && touchingWallLeft {

		playerCharacterController.Velocity.X = playerWallJumpSpeedX
		playerCharacterController.Velocity.Y = -playerWallJumpSpeedY

		wallJumpInputLock = true
		wallJumpInputLockTimerSystem.SetTimerFromPoolWithDurationLoopAndFunc(
			wallJumpLockInputForSeconds,
			false,
			func() {
				wallJumpInputLock = false
			},
		)
	}

	if jumpJustPressed && touchingWallRight {

		playerCharacterController.Velocity.X = -playerWallJumpSpeedX
		playerCharacterController.Velocity.Y = -playerWallJumpSpeedY

		wallJumpInputLock = true
		wallJumpInputLockTimerSystem.SetTimerFromPoolWithDurationLoopAndFunc(
			wallJumpLockInputForSeconds,
			false,
			func() {
				wallJumpInputLock = false
			},
		)
	}

	if !playerCharacterController.Grounded && touchingWall && upArrowHeld {
		playerCharacterController.Velocity.Y = -playerWallClimbSpeed
	}

	if arrowDownJustPressed {
		playerCharacterController.CollideWithOneWayPlatform = false
	}
	if arrowDownJustReleased {
		playerCharacterController.CollideWithOneWayPlatform = true
	}

	oldFrameGrounded = playerCharacterController.Grounded

	playerCharacterController.UpdateCharacter(playerEntity, world, world.Transforms[playerEntity], dt)

	if playerCharacterController.Grounded {
		jumpBufferTimerSystem.ForceEndAllTimersForNextUpdate()
		coyeteBufferTimerSystem.ForceEndAllTimersForNextUpdate()
		// playerDashTimerSystem.ForceEndAllTimersForNextUpdate()
		// wallJumpInputLockTimerSystem.ForceEndAllTimersForNextUpdate()
	}

	if oldFrameGrounded && !playerCharacterController.Grounded && playerCharacterController.Velocity.Y >= 0 {
		coyeteBufferTimerSystem.SetTimerFromPoolWithDurationLoopAndFunc(coyeteTime, false, func() {})
	}

}
