package Engine

type CharacterController struct {
	Velocity Vector2

	Width  float64
	Height float64

	Grounded                  bool
	CollideWithOneWayPlatform bool

	Gravity   float64
	JumpSpeed float64
	MoveSpeed float64

	oldBottom float64
}

func (cc *CharacterController) UpdateCharacter(
	world *World,
	transform *Transform,
	dt float64,
) {

	gravity := cc.Gravity

	if cc.Velocity.Y > 0 {
		gravity *= 1.8
	}

	cc.Velocity.Y += gravity * dt

	transform.Position.X += cc.Velocity.X * dt

	cc.ResolveHorizontal(
		world,
		transform,
	)

	cc.oldBottom = transform.Position.Y + (cc.Height / 2.0)

	transform.Position.Y += cc.Velocity.Y * dt

	cc.ResolveVertical(
		world,
		transform,
	)
}

func (cc *CharacterController) ResolveHorizontal(
	world *World,
	transform *Transform,
) {

	playerCollider := AABB{
		X:      transform.Position.X - (cc.Width / 2.0),
		Y:      transform.Position.Y - (cc.Height / 2.0),
		Width:  cc.Width,
		Height: cc.Height,
	}

	for _, collider := range world.LevelColliders {

		if collider.Layer == DEAFULT {

			if !AABBOverlap(playerCollider, *collider) {
				continue
			}

			if cc.Velocity.X > 0 {

				transform.Position.X = collider.X - (cc.Width / 2.0)

			} else if cc.Velocity.X < 0 {

				transform.Position.X = collider.X + collider.Width + (cc.Width / 2.0)
			}

			cc.Velocity.X = 0

			playerCollider.X = transform.Position.X
		}
	}
}

func (cc *CharacterController) ResolveVertical(
	world *World,
	transform *Transform,
) {

	cc.Grounded = false

	playerCollider := AABB{
		X:      transform.Position.X - (cc.Width / 2.0),
		Y:      transform.Position.Y - (cc.Height / 2.0),
		Width:  cc.Width,
		Height: cc.Height,
	}

	for _, collider := range world.LevelColliders {

		if collider.Layer == DEAFULT {
			if !AABBOverlap(playerCollider, *collider) {
				continue
			}

			if cc.Velocity.Y > 0 {

				transform.Position.Y = collider.Y - (cc.Height / 2.0)

				cc.Grounded = true

			} else if cc.Velocity.Y < 0 {

				transform.Position.Y = collider.Y + collider.Height + (cc.Height / 2.0)
			}

			cc.Velocity.Y = 0

			playerCollider.Y = transform.Position.Y

		} else if collider.Layer == ONE_WAY_PLATFORMS && cc.CollideWithOneWayPlatform {

			if playerCollider.Right() > collider.Left() &&
				playerCollider.Left() < collider.Right() {

				if cc.Velocity.Y > 0 &&
					cc.oldBottom <= collider.Top() &&
					playerCollider.Bottom() >= collider.Top() {

					transform.Position.Y = collider.Top() - (cc.Height / 2.0)

					cc.Velocity.Y = 0
					cc.Grounded = true

					playerCollider.Y = transform.Position.Y - (cc.Height / 2.0)
				}
			}
		}
	}
}
