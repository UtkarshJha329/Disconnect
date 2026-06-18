package Engine

import (
	"github.com/hajimehoshi/ebiten/v2"
)

func (scene *Scene) RenderSceneHierarchy(cameraEntity Entity, world *World) {

	scene.CreateEntitiesToDrawSortedForScene()

	camera := world.Cameras[cameraEntity]
	op := &ebiten.DrawImageOptions{}
	for _, entity := range scene.EntitiesToDrawSorted {
		op.GeoM = scene.world.Transforms[entity].WorldTransformMatrix
		op.GeoM.Concat(*camera.GetCameraTransformMatrix(cameraEntity, world))
		camera.RenderTexture.DrawImage(scene.world.Sprites[entity].Tex, op)
	}

}
