package Engine

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type Transform struct {
	Position Vector3
	Rotation float64
	Scale    Vector2
	Pivot    Vector2

	WorldTransformMatrix ebiten.GeoM
}

func (trans *Transform) CalculateLocalMatrix() ebiten.GeoM {

	var localMatrix ebiten.GeoM

	localMatrix.Translate(-trans.Pivot.X, -trans.Pivot.Y)
	localMatrix.Scale(trans.Scale.X, trans.Scale.Y)
	localMatrix.Rotate(trans.Rotation)
	localMatrix.Translate(trans.Position.X, trans.Position.Y)

	return localMatrix
}
