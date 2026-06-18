package Engine

import (
	"math"
)

type CharacterController struct {
	Velocity                    Vector2
	PlatformMoveAmountLastFrame Vector2

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
	characterControllerEntity Entity,
	world *World,
	transform *Transform,
	dt float64,
) {

	gravity := cc.Gravity

	if cc.Velocity.Y > 0 {
		gravity *= 1.8
	}

	cc.Velocity.Y += gravity * dt

	moveX := cc.Velocity.X * dt
	moveY := cc.Velocity.Y * dt

	if !cc.Grounded {
		moveX += cc.PlatformMoveAmountLastFrame.X
		moveY += cc.PlatformMoveAmountLastFrame.Y

		horizontalDecay := math.Pow(0.95, dt*60.0)
		// horizontalDecay := 1.0
		verticalDecay := math.Pow(0.95, dt*60.0)
		// verticalDecay := 1.0
		cc.PlatformMoveAmountLastFrame.X *= horizontalDecay
		cc.PlatformMoveAmountLastFrame.Y *= verticalDecay
	}

	cc.MoveHorizontal(moveX, transform, world, func(collisionResult *CollisionResult) {})

	cc.oldBottom = transform.Position.Y + (cc.Height / 2.0)

	// transform.Position.Y += cc.Velocity.Y * dt

	// cc.ResolveVertical(
	// 	world,
	// 	transform,
	// )
	cc.MoveVertical(characterControllerEntity, moveY, true, transform, world, func(collisionResult *CollisionResult) {})
}

func (cc *CharacterController) MoveHorizontal(amount float64, transform *Transform, world *World, OnResolutionFailure func(collisionResult *CollisionResult)) {

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

			collisionResult := cc.CollidesWithObstaclesAtPositionHorizontally(&playerCollider, world)
			if collisionResult == nil {
				transform.Position.X = transform.Position.X + sign
				move -= sign
			} else {
				if OnResolutionFailure != nil {
					OnResolutionFailure(collisionResult)
				}
				break
			}

		}
	}
}

func (cc *CharacterController) MoveVertical(characterControllerEntity Entity, amount float64, changeGrouned bool, transform *Transform, world *World, OnResolutionFailure func(collisionResult *CollisionResult)) {

	if changeGrouned {
		cc.Grounded = false
	}
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
					// fmt.Println("Setting grounded to true because colliding with entity : ", collisionResult.entity)

					if changeGrouned {
						cc.Grounded = true
					}
				}
				cc.Velocity.Y = 0.0
				if OnResolutionFailure != nil {
					OnResolutionFailure(collisionResult)
				}
				break
			}

		}
	} else {
		player := cc.GetAABB(characterControllerEntity, world)
		player.Y += 1

		cc.Grounded = cc.CollidesWithObstaclesAtPositionVertically(player, 1, world) != nil
	}
}

func (cc *CharacterController) CollidesWithObstaclesAtPositionHorizontally(characterCollider *AABB, world *World) *CollisionResult {

	for e, platform := range world.Platforms {

		if !platform.Collidable || !AABBOverlap(*characterCollider, *platform.AABB) {
			continue
		}

		if platform.AABB.Layer == ONE_WAY_PLATFORMS {
			continue
		}

		if platform.AABB.Layer == DEFAULT {
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

		if platform.AABB.Layer == DEFAULT {
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

func (cc *CharacterController) TouchingWall(dir float64, e Entity, world *World) bool {

	aabb := cc.GetAABB(e, world)
	aabb.X += dir

	collisionResult := cc.CollidesWithObstaclesAtPositionHorizontally(aabb, world)
	return collisionResult != nil
}

func (cc *CharacterController) GetAABB(e Entity, world *World) *AABB {
	return &AABB{
		X:      world.Transforms[e].Position.X - (cc.Width / 2.0),
		Y:      world.Transforms[e].Position.Y - (cc.Height / 2.0),
		Width:  cc.Width,
		Height: cc.Height,
	}
}
