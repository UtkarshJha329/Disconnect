package Engine

import (
	"github.com/hajimehoshi/ebiten/v2"
)

func (scene *Scene) RenderSceneHierarchy(screen *ebiten.Image, world *World) {

	scene.CreateEntitiesToDrawSortedForScene()

	op := &ebiten.DrawImageOptions{}
	for _, entity := range scene.EntitiesToDrawSorted {
		op.GeoM = scene.world.Transforms[entity].WorldTransformMatrix
		screen.DrawImage(scene.world.Sprites[entity].Tex, op)
	}

}
