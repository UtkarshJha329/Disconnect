package main

import (
	"Disconnect/Engine"
	"Disconnect/Project"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Game struct{}

var world *Engine.World = Engine.NewWorld()

func (g *Game) Update() error {
	dt := 1.0 / 60.0
	for _, updateFunc := range world.EntityUpdateFuncs {
		updateFunc(world, dt)
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	world.Scene.RenderSceneHierarchy(screen, world)

	for _, platform := range world.Platforms {

		collider := platform.AABB
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

	world.EntityInitfuncs = append(world.EntityInitfuncs,
		Project.PlatformsInitFunc,
		Project.PlayerInitFunc,
	)

	world.EntityUpdateFuncs = append(world.EntityUpdateFuncs,
		Project.PlatformsUpdateFunc,
		Project.PlayerUpdateFunc,
	)

	for _, initFunc := range world.EntityInitfuncs {
		initFunc(world)
	}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
