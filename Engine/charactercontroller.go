package Engine

type CharacterController struct {
	Velocity Vector2

	Width  float64
	Height float64

	Grounded bool

	Gravity   float64
	JumpSpeed float64
	MoveSpeed float64
}

func UpdateCharacter(
	world *World,
	controller *CharacterController,
	transform *Transform,
	dt float64,
) {

	gravity := controller.Gravity

	if controller.Velocity.Y > 0 {
		gravity *= 1.8
	}

	controller.Velocity.Y += gravity * dt

	transform.Position.X += controller.Velocity.X * dt

	ResolveHorizontal(
		world,
		controller,
		transform,
	)

	transform.Position.Y += controller.Velocity.Y * dt

	ResolveVertical(
		world,
		controller,
		transform,
	)
}
