package Engine

import (
	"math"
)

type CharacterController struct {
	Velocity Vector2

	Width  float64
	Height float64

	Grounded                  bool
	CollideWithOneWayPlatform bool

	Gravity   float64
	JumpSpeed float64
	MoveSpeed float64

	CurrentlyOnPlatform *Entity

	oldBottom float64
	remainder Vector2
}

type CollisionResult struct {
	aabb   *AABB
	entity Entity
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

	// transform.Position.X += cc.Velocity.X * dt

	// cc.ResolveHorizontal(
	// 	world,
	// 	transform,
	// )

	cc.MoveHorizontal(cc.Velocity.X*dt, transform, world, func(collisionResult *CollisionResult) {})

	cc.oldBottom = transform.Position.Y + (cc.Height / 2.0)

	// transform.Position.Y += cc.Velocity.Y * dt

	// cc.ResolveVertical(
	// 	world,
	// 	transform,
	// )
	cc.MoveVertical(cc.Velocity.Y*dt, transform, world, func(collisionResult *CollisionResult) {})
}

func (cc *CharacterController) MoveHorizontal(amount float64, transform *Transform, world *World, OnCollide func(collisionResult *CollisionResult)) {

	cc.remainder.X += amount
	move := math.Round(cc.remainder.X)

	if move != 0 {
		cc.remainder.X -= move
		sign := math.Copysign(1.0, move)
		for move != 0 {

			playerCollider := AABB{
				X:      transform.Position.X - (cc.Width / 2.0) + sign,
				Y:      transform.Position.Y - (cc.Height / 2.0),
				Width:  cc.Width,
				Height: cc.Height,
			}

			collisionResult := cc.CollidesWithObstaclesAtPositionHorizontally(&playerCollider, sign, world)
			if collisionResult == nil {
				transform.Position.X = transform.Position.X + sign
				move -= sign
			} else {
				if OnCollide != nil {
					OnCollide(collisionResult)
				}
				break
			}

		}
	}
}

func (cc *CharacterController) MoveVertical(amount float64, transform *Transform, world *World, OnCollide func(collisionResult *CollisionResult)) {

	cc.remainder.Y += amount
	move := math.Round(cc.remainder.Y)
	if move != 0 {
		cc.remainder.Y -= move
		sign := math.Copysign(1.0, move)
		for move != 0 {

			playerCollider := AABB{
				X:      transform.Position.X - (cc.Width / 2.0),
				Y:      transform.Position.Y - (cc.Height / 2.0) + sign,
				Width:  cc.Width,
				Height: cc.Height,
			}

			collisionResult := cc.CollidesWithObstaclesAtPositionVertically(&playerCollider, sign, world)
			if collisionResult == nil {
				cc.oldBottom = transform.Position.Y + (cc.Height / 2.0)
				transform.Position.Y = transform.Position.Y + sign
				move -= sign
			} else {
				if sign > 0 {
					cc.Grounded = true
				}
				cc.Velocity.Y = 0.0
				if OnCollide != nil {
					OnCollide(collisionResult)
				}
				break
			}

		}
	}
}

func (cc *CharacterController) CollidesWithObstaclesAtPositionHorizontally(characterCollider *AABB, direction float64, world *World) *CollisionResult {

	for e, platform := range world.Platforms {

		if !platform.Collidable || !AABBOverlap(*characterCollider, *platform.AABB) {
			continue
		}

		if platform.AABB.Layer == ONE_WAY_PLATFORMS {
			continue
		}

		if platform.AABB.Layer == DEAFULT {
			return &CollisionResult{platform.AABB, e}
		}

	}
	return nil
}

func (cc *CharacterController) CollidesWithObstaclesAtPositionVertically(characterCollider *AABB, direction float64, world *World) *CollisionResult {

	for e, platform := range world.Platforms {

		if !platform.Collidable || !AABBOverlap(*characterCollider, *platform.AABB) {
			continue
		}

		if platform.AABB.Layer == DEAFULT {
			return &CollisionResult{platform.AABB, e}
		}

		if cc.CollideWithOneWayPlatform &&
			platform.AABB.Layer == ONE_WAY_PLATFORMS {

			if direction < 0 {
				continue
			}

			if cc.oldBottom <= platform.AABB.Top() &&
				characterCollider.Bottom() >= platform.AABB.Top() {
				return &CollisionResult{platform.AABB, e}
			}
		}
	}
	return nil
}

func (cc *CharacterController) GetAABB(e Entity, world *World) *AABB {
	return &AABB{
		X:      world.Transforms[e].Position.X - (cc.Width / 2.0),
		Y:      world.Transforms[e].Position.Y - (cc.Height / 2.0),
		Width:  cc.Width,
		Height: cc.Height,
	}
}

// func (cc *CharacterController) ResolveHorizontal(
// 	world *World,
// 	transform *Transform,
// ) {

// 	playerCollider := AABB{
// 		X:      transform.Position.X - (cc.Width / 2.0),
// 		Y:      transform.Position.Y - (cc.Height / 2.0),
// 		Width:  cc.Width,
// 		Height: cc.Height,
// 	}

// 	for _, collider := range world.LevelColliders {

// 		if collider.Layer == DEAFULT {

// 			if !AABBOverlap(playerCollider, *collider) {
// 				continue
// 			}

// 			if cc.Velocity.X > 0 {

// 				transform.Position.X = collider.X - (cc.Width / 2.0)

// 			} else if cc.Velocity.X < 0 {

// 				transform.Position.X = collider.X + collider.Width + (cc.Width / 2.0)
// 			}

// 			cc.Velocity.X = 0

// 			playerCollider.X = transform.Position.X
// 		}
// 	}
// }

// func (cc *CharacterController) ResolveVertical(
// 	world *World,
// 	transform *Transform,
// ) {

// 	cc.Grounded = false

// 	playerCollider := AABB{
// 		X:      transform.Position.X - (cc.Width / 2.0),
// 		Y:      transform.Position.Y - (cc.Height / 2.0),
// 		Width:  cc.Width,
// 		Height: cc.Height,
// 	}

// 	for e, collider := range world.LevelColliders {

// 		if collider.Layer == DEAFULT {
// 			if !AABBOverlap(playerCollider, *collider) {
// 				continue
// 			}

// 			if cc.Velocity.Y > 0 {

// 				transform.Position.Y = collider.Y - (cc.Height / 2.0)

// 				cc.Grounded = true
// 				cc.CurrentlyOnPlatform = &e

// 			} else if cc.Velocity.Y < 0 {

// 				transform.Position.Y = collider.Y + collider.Height + (cc.Height / 2.0)
// 			}

// 			cc.Velocity.Y = 0

// 			playerCollider.Y = transform.Position.Y

// 		} else if collider.Layer == ONE_WAY_PLATFORMS && cc.CollideWithOneWayPlatform {

// 			if playerCollider.Right() > collider.Left() &&
// 				playerCollider.Left() < collider.Right() {

// 				if cc.Velocity.Y > 0 &&
// 					cc.oldBottom <= collider.Top() &&
// 					playerCollider.Bottom() >= collider.Top() {

// 					transform.Position.Y = collider.Top() - (cc.Height / 2.0)

// 					cc.Velocity.Y = 0
// 					cc.Grounded = true
// 					cc.CurrentlyOnPlatform = &e

// 					playerCollider.Y = transform.Position.Y - (cc.Height / 2.0)
// 				}
// 			}
// 		}
// 	}

// 	if !cc.Grounded {
// 		cc.CurrentlyOnPlatform = nil
// 	}
// }
