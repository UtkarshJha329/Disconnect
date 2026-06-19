package Engine

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

func (scene *Scene) RenderSceneHierarchy(cameraEntity Entity, world *World) {

	scene.CreateEntitiesToDrawSortedForScene()

	camera := world.Cameras[cameraEntity]
	op := &ebiten.DrawImageOptions{}
	for _, entity := range scene.EntitiesToDrawSorted {
		op.GeoM = scene.world.Transforms[entity].WorldTransformMatrix
		op.GeoM.Concat(*camera.GetCameraTransformMatrix(cameraEntity, world))

		curSprite := scene.world.Sprites[entity]
		camera.RenderTexture.DrawImage(
			curSprite.Tex.SubImage(
				image.Rect(
					int(curSprite.FramePosition.X),
					int(curSprite.FramePosition.Y),
					int(curSprite.FramePosition.X+curSprite.FrameSize.X),
					int(curSprite.FramePosition.Y+curSprite.FrameSize.Y))).(*ebiten.Image),
			op)
	}

}
