package main

import (
	"CookieCrumbles/Engine"
	"image/color"
	"log"

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

	Engine.UpdateCharacter(world, playerCharacterController, world.Transforms[playerEntity], 1.0/60.0)

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

	playerEntity = Engine.CreateSpriteFromFileInScene(&world.Scene, world, "Assets/player/idle/00.png")
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

		Grounded: false,

		Velocity: Engine.Vector2{
			X: 0,
			Y: 0,
		},
	}

	gunEntity := Engine.CreateSpriteFromFileInSceneWithParentEntity(&world.Scene, world, "Assets/gun.png", playerEntity)
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

	floor := Engine.AABB{
		X:      0,
		Y:      450,
		Width:  640,
		Height: 32,
	}

	leftWall := Engine.AABB{
		X:      -16,
		Y:      0,
		Width:  16,
		Height: 480,
	}

	rightWall := Engine.AABB{
		X:      639,
		Y:      0,
		Width:  16,
		Height: 480,
	}

	world.LevelColliders = append(world.LevelColliders, floor)
	world.LevelColliders = append(world.LevelColliders, leftWall)
	world.LevelColliders = append(world.LevelColliders, rightWall)

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
