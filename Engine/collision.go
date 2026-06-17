package Engine

const (
	DEAFULT = iota
	ONE_WAY_PLATFORMS
)

type AABB struct {
	X float64
	Y float64

	Width  float64
	Height float64

	Layer uint
}

func AABBOverlap(a, b AABB) bool {
	return a.X < b.X+b.Width &&
		a.X+a.Width > b.X &&
		a.Y < b.Y+b.Height &&
		a.Y+a.Height > b.Y
}

func (aabb *AABB) Top() float64 {
	return aabb.Y
}

func (aabb *AABB) Bottom() float64 {
	return aabb.Y + aabb.Height
}

func (aabb *AABB) Left() float64 {
	return aabb.X
}

func (aabb *AABB) Right() float64 {
	return aabb.X + aabb.Width
}
