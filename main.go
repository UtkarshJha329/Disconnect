package main

import (
	"CookieCrumbles/Engine"
	"fmt"
	"image/color"
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Game struct{}

var world *Engine.World = Engine.NewWorld()
var floatInterpolationSystem Engine.InterpolationSystem[float64]
var playerEntity Engine.Entity

func (g *Game) Update() error {
	floatInterpolationSystem.UpdateAllInterpolationDeltasAndStates()

	playerCharacterController := world.CharacterControllers[playerEntity]

	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		playerCharacterController.Velocity.X = -playerCharacterController.MoveSpeed
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		playerCharacterController.Velocity.X = playerCharacterController.MoveSpeed
	} else {
		playerCharacterController.Velocity.X = 0
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) &&
		playerCharacterController.Grounded {

		playerCharacterController.Velocity.Y = -playerCharacterController.JumpSpeed
	}

	if inpututil.IsKeyJustReleased(ebiten.KeyArrowUp) &&
		playerCharacterController.Velocity.Y < 0 {

		playerCharacterController.Velocity.Y *= 0.5
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
		playerCharacterController.CollideWithOneWayPlatform = false
	}
	if inpututil.IsKeyJustReleased(ebiten.KeyArrowDown) {
		playerCharacterController.CollideWithOneWayPlatform = true
	}

	dt := 1.0 / 60.0
	playerCharacterController.UpdateCharacter(world, world.Transforms[playerEntity], dt)

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	world.Scene.RenderSceneHierarchy(screen, world)

	for _, collider := range world.LevelColliders {

		vector.StrokeRect(
			screen,
			float32(collider.X),
			float32(collider.Y),
			float32(collider.Width),
			float32(collider.Height),
			1,
			color.RGBA{255, 0, 0, 255},
			false,
		)
	}

	for e, cc := range world.CharacterControllers {
		vector.StrokeRect(
			screen,
			float32(world.Transforms[e].Position.X-(cc.Width/2.0)),
			float32(world.Transforms[e].Position.Y-(cc.Height/2.0)),
			float32(cc.Width),
			float32(cc.Height),
			1,
			color.RGBA{0, 0, 255, 255},
			false,
		)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 640, 480
}

func main() {
	game := &Game{}
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Sora Engine")

	screenWidth, screenHeight := ebiten.WindowSize()

	floatInterpolationSystem.InitWithInterpolationsInPool("Float Interpolation System", 10)

	playerEntity = world.CreateSpriteFromFileInScene(&world.Scene, "Assets/player/idle/00.png")
	world.Transforms[playerEntity].Scale = Engine.Vector2{X: 4.0, Y: 4.0}
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

		Gravity:   1200,
		MoveSpeed: 500,
		JumpSpeed: 500,

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

	world.CreateNewLevelColliderInScene(&world.Scene, 0, 450, 640, 32, Engine.DEAFULT)
	world.CreateNewLevelColliderInScene(&world.Scene, -16, 0, 16, 480, Engine.DEAFULT)
	world.CreateNewLevelColliderInScene(&world.Scene, 639, 0, 16, 480, Engine.DEAFULT)
	world.CreateNewLevelColliderInScene(&world.Scene, 100, 380, 120, 20, Engine.DEAFULT)
	platform2Index := world.CreateNewLevelColliderInScene(&world.Scene, 300, 320, 120, 20, Engine.DEAFULT)
	platform2XFinal := 500.0

	floatInterpolationSystem.CreateNewInterpolation(
		world.LevelColliders[platform2Index].X,
		&world.LevelColliders[platform2Index].X,
		&platform2XFinal,
		time.Duration(2.0*float64(time.Second)),
		false,
		true,
		Engine.LinearInterpolationFloat,
		func() { fmt.Println("Finished interpolating platform position.") },
	)

	world.CreateNewLevelColliderInScene(&world.Scene, 100, 280, 120, 20, Engine.ONE_WAY_PLATFORMS)

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
