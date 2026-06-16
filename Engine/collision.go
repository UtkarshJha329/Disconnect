package Engine

type AABB struct {
	X float64
	Y float64

	Width  float64
	Height float64
}

func AABBOverlap(a, b AABB) bool {
	return a.X < b.X+b.Width &&
		a.X+a.Width > b.X &&
		a.Y < b.Y+b.Height &&
		a.Y+a.Height > b.Y
}

func ResolveHorizontal(
	world *World,
	controller *CharacterController,
	transform *Transform,
) {

	player := AABB{
		X:      transform.Position.X - (controller.Width / 2.0),
		Y:      transform.Position.Y - (controller.Height / 2.0),
		Width:  controller.Width,
		Height: controller.Height,
	}

	for _, wall := range world.LevelColliders {

		if !AABBOverlap(player, wall) {
			continue
		}

		if controller.Velocity.X > 0 {

			transform.Position.X = wall.X - (controller.Width / 2.0)

		} else if controller.Velocity.X < 0 {

			transform.Position.X = wall.X + wall.Width + (controller.Width / 2.0)
		}

		controller.Velocity.X = 0

		player.X = transform.Position.X
	}
}

func ResolveVertical(
	world *World,
	controller *CharacterController,
	transform *Transform,
) {

	controller.Grounded = false

	player := AABB{
		X:      transform.Position.X - (controller.Width / 2.0),
		Y:      transform.Position.Y - (controller.Height / 2.0),
		Width:  controller.Width,
		Height: controller.Height,
	}

	for _, floor := range world.LevelColliders {

		if !AABBOverlap(player, floor) {
			continue
		}

		if controller.Velocity.Y > 0 {

			transform.Position.Y = floor.Y - (controller.Height / 2.0)

			controller.Grounded = true

		} else if controller.Velocity.Y < 0 {

			transform.Position.Y = floor.Y + floor.Height + (controller.Height / 2.0)
		}

		controller.Velocity.Y = 0

		player.Y = transform.Position.Y
	}
}
