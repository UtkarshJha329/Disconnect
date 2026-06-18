package Engine

import (
	"math"
	"math/rand/v2"
)

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

// Utility

// RandomPointInCircle generates a uniform random point within a circle.
func RandomPointInCircle(centerX, centerY, radius float64) *Vector2 {
	// Generate a random angle between 0 and 2*Pi
	theta := rand.Float64() * 2 * math.Pi

	// Sqrt ensures the points are uniformly distributed by area
	r := radius * math.Sqrt(rand.Float64())

	// Convert polar coordinates to Cartesian coordinates
	x := centerX + r*math.Cos(theta)
	y := centerY + r*math.Sin(theta)

	return &Vector2{x, y}
}
