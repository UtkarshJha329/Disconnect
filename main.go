package main

import (
	"Disconnect/Engine"
	"Disconnect/Project"
	"cmp"
	"image/color"
	"log"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Game struct{}

var world *Engine.World = Engine.NewWorld()

func (g *Game) Update() error {
	dt := 1.0 / float64(ebiten.TPS())
	for _, updateFunc := range world.EntityUpdateFuncs {
		updateFunc(world, dt)
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {

	camera := world.Cameras[Project.MainCameraEntity]
	camera.RenderTexture.Clear()
	world.Scene.RenderSceneHierarchy(Project.MainCameraEntity, world)

	cameraMatrix := camera.GetCameraTransformMatrix(Project.MainCameraEntity, world)

	for _, platform := range world.Platforms {

		collider := platform.AABB
		x, y := cameraMatrix.Apply(collider.X, collider.Y)

		vector.StrokeRect(
			camera.RenderTexture,
			float32(x),
			float32(y),
			float32(collider.Width),
			float32(collider.Height),
			1,
			color.RGBA{0, 255, 0, 255},
			false,
		)
	}

	for _, trigger := range world.Triggers {

		collider := trigger.AABB
		x, y := cameraMatrix.Apply(collider.X, collider.Y)

		vector.StrokeRect(
			camera.RenderTexture,
			float32(x),
			float32(y),
			float32(collider.Width),
			float32(collider.Height),
			1,
			color.RGBA{255, 0, 0, 255},
			false,
		)
	}

	for e, cc := range world.CharacterControllers {

		x, y := cameraMatrix.Apply(
			world.Transforms[e].Position.X-cc.Width/2.0,
			world.Transforms[e].Position.Y-cc.Height/2.0,
		)

		vector.StrokeRect(
			camera.RenderTexture,
			float32(x),
			float32(y),
			float32(cc.Width),
			float32(cc.Height),
			1,
			color.RGBA{0, 0, 255, 255},
			false,
		)
	}

	Engine.DrawParticleSystems(camera, Project.MainCameraEntity, world)

	cameras := make([]*Project.MainCameraEntityTransformPair, 0, len(world.Cameras))
	for e, camera := range world.Cameras {
		curEle := &Project.MainCameraEntityTransformPair{
			Entity:    e,
			Transform: world.Transforms[e],
			Camera:    camera,
		}
		cameras = append(cameras, curEle)
	}
	slices.SortFunc(cameras, func(a, b *Project.MainCameraEntityTransformPair) int {
		return cmp.Compare(a.Transform.Position.Z, b.Transform.Position.Z)
	})

	op := ebiten.DrawImageOptions{}
	for _, camera := range cameras {
		screen.DrawImage(camera.Camera.RenderTexture, &op)
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
		Project.MainCameraInitFunc,
		Project.TriggersInitFunc,
	)

	world.EntityUpdateFuncs = append(world.EntityUpdateFuncs,
		Project.PlatformsUpdateFunc,
		Project.PlayerUpdateFunc,
		Project.MainCameraUpdateFunc,
		Project.TriggersUpdateFunc,
		Engine.ParticlesUpdateFunc,
	)

	for _, initFunc := range world.EntityInitfuncs {
		initFunc(world)
	}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
