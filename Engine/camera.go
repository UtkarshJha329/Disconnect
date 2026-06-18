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

func (camera *Camera) GetScreenCenterInWorld(cameraEntity Entity, world *World) Vector2 {
	screenCenter := Vector2{X: camera.ScreenDims.X / 2.0, Y: camera.ScreenDims.Y / 2.0}

	cameraMatrix := camera.GetCameraTransformMatrix(cameraEntity, world)
	if cameraMatrix.IsInvertible() {
		cameraMatrix.Invert()
	}

	worldScreenCenterX, worldScreenCenterY := cameraMatrix.Apply(screenCenter.X, screenCenter.Y)
	return Vector2{X: worldScreenCenterX, Y: worldScreenCenterY}
}
