package Engine

import "github.com/hajimehoshi/ebiten/v2"

type Camera struct {
	RenderTexture *ebiten.Image
	ScreenDims    Vector2
}

func (camera *Camera) FollowEntity(cameraEntity Entity, followEnity Entity, world *World) {
	world.Transforms[cameraEntity].Position.X = world.Transforms[followEnity].Position.X
	world.Transforms[cameraEntity].Position.Y = world.Transforms[followEnity].Position.Y
}

func (camera *Camera) GetCameraTransformMatrix(cameraEntity Entity, world *World) *ebiten.GeoM {
	cameraMatrix := ebiten.GeoM{}

	cameraMatrix.Translate(-world.Transforms[cameraEntity].Position.X, -world.Transforms[cameraEntity].Position.Y)
	cameraMatrix.Scale(world.Transforms[cameraEntity].Scale.X, world.Transforms[cameraEntity].Scale.Y)
	cameraMatrix.Translate(camera.ScreenDims.X/2.0, camera.ScreenDims.Y/2.0)

	return &cameraMatrix
}
