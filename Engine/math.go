package Engine

import "math"

func DegreesToRadians(degrees float64) float64 {
	return degrees * math.Pi / 180.0
}

type Vector2 struct {
	X float64
	Y float64
}

func (a *Vector2) Add(b *Vector2) Vector2 {
	return Vector2{a.X + b.X, a.Y + b.Y}
}

func (a *Vector2) Subtract(b *Vector2) Vector2 {
	return Vector2{a.X - b.X, a.Y - b.Y}
}

func (a *Vector2) Multiply(b *Vector2) Vector2 {
	return Vector2{a.X * b.X, a.Y * b.Y}
}

func (a *Vector2) Divide(b *Vector2) Vector2 {
	return Vector2{a.X / b.X, a.Y / b.Y}
}

func (a *Vector2) Dot(b *Vector2) float64 {
	return a.X*b.X + a.Y*b.Y
}

func (a *Vector2) MultiplyFloat(b float64) Vector2 {
	return Vector2{a.X * b, a.Y * b}
}

func (a *Vector2) DivideFloat(b float64) Vector2 {
	return Vector2{a.X / b, a.Y / b}
}

func (a *Vector2) Normalize() Vector2 {
	normalizingFactor := math.Sqrt(math.Pow(a.X, 2) + math.Pow(a.Y, 2))
	return a.DivideFloat(normalizingFactor)
}

type Vector3 struct {
	X float64
	Y float64
	Z float64
}

func (a *Vector3) Add(b *Vector3) Vector3 {
	return Vector3{a.X + b.X, a.Y + b.Y, a.Z + b.Z}
}

func (a *Vector3) Subtract(b *Vector3) Vector3 {
	return Vector3{a.X - b.X, a.Y - b.Y, a.Z - b.Z}
}

func (a *Vector3) Multiply(b *Vector3) Vector3 {
	return Vector3{a.X * b.X, a.Y * b.Y, a.Z * b.Z}
}

func (a *Vector3) Divide(b *Vector3) Vector3 {
	return Vector3{a.X / b.X, a.Y / b.Y, a.Z / b.Z}
}

func (a *Vector3) Dot(b *Vector3) float64 {
	return a.X*b.X + a.Y*b.Y + a.Z*b.Z
}

func (a *Vector3) MultiplyFloat(b float64) Vector3 {
	return Vector3{a.X * b, a.Y * b, a.Z * b}
}

func (a *Vector3) DivideFloat(b float64) Vector3 {
	return Vector3{a.X / b, a.Y / b, a.Z / b}
}

func (a *Vector3) AddVector2(b *Vector2) Vector3 {
	return Vector3{a.X + b.X, a.Y + b.Y, a.Z}
}

func (a *Vector3) SubtractVector2(b *Vector2) Vector3 {
	return Vector3{a.X - b.X, a.Y - b.Y, a.Z}
}

func (a *Vector3) MultiplyVector2(b *Vector2) Vector3 {
	return Vector3{a.X * b.X, a.Y * b.Y, a.Z}
}

func (a *Vector3) DivideVector2(b *Vector2) Vector3 {
	return Vector3{a.X / b.X, a.Y / b.Y, a.Z}
}
