package main

import (
	"CookieCrumbles/Engine"
	"fmt"
	"log"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

// Game implements ebiten.Game interface.
type Game struct{}

var world *Engine.World = Engine.NewWorld()
var floatInterpolationSystem Engine.InterpolationSystem[float64]

func (g *Game) Update() error {
	floatInterpolationSystem.UpdateAllInterpolationDeltasAndStates()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	world.Scene.RenderSceneHierarchy(screen, world)
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

	playerEntity := Engine.CreateSpriteFromFileInScene(&world.Scene, world, "Assets/player/idle/00.png")
	world.Transforms[playerEntity].Scale = Engine.Vector2{X: 4.0, Y: 4.0}
	world.Transforms[playerEntity].Position = Engine.Vector3{X: float64(screenWidth) / 2.0, Y: float64(screenHeight) / 2.0, Z: 100}

	playerImageSize := world.Sprites[playerEntity].Tex.Bounds().Size()
	world.Transforms[playerEntity].Pivot = Engine.Vector2{X: float64(playerImageSize.X) / 2.0, Y: float64(playerImageSize.Y) / 2.0}

	gunEntity := Engine.CreateSpriteFromFileInSceneWithParentEntity(&world.Scene, world, "Assets/gun.png", playerEntity)
	world.Transforms[gunEntity].Scale = Engine.Vector2{X: 1.0, Y: 1.0}
	world.Transforms[gunEntity].Position = Engine.Vector3{X: 6.0, Y: 0.0, Z: 200}

	gunImageSize := world.Sprites[gunEntity].Tex.Bounds().Size()
	world.Transforms[gunEntity].Pivot = Engine.Vector2{X: float64(gunImageSize.X) / 2.0, Y: float64(gunImageSize.Y) / 2.0}

	rotatePlayerToValue := Engine.DegreesToRadians(360.0)
	timeToRotatePlayer := 60.0
	floatInterpolationSystem.CreateNewInterpolation(
		world.Transforms[playerEntity].Rotation,
		&world.Transforms[playerEntity].Rotation,
		&rotatePlayerToValue,
		time.Duration(timeToRotatePlayer*float64(time.Second)),
		true,
		Engine.LinearInterpolationFloat,
		func() { fmt.Println("Finished rotating sprite.") },
	)

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
