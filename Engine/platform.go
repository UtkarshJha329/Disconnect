package Engine

import (
	"fmt"
	"math"
)

type Platform struct {
	AABB       *AABB
	Velocity   Vector2
	Remainder  Vector2
	Collidable bool
}

func (platform *Platform) Move(platformEntity Entity, x, y float64, world *World) {
	platform.Remainder.X += x
	platform.Remainder.Y += y

	moveX := math.Round(platform.Remainder.X)
	moveY := math.Round(platform.Remainder.Y)

	if moveX != 0.0 || moveY != 0.0 {
		ridingCCs := platform.GetAllRidingCharacterControllers(platformEntity, world)

		platform.Collidable = false

		if moveX != 0 {

			platform.Remainder.X -= moveX
			platform.AABB.X += moveX

			if moveX > 0 {
				for e, cc := range world.CharacterControllers {
					ccAABB := cc.GetAABB(e, world)
					if AABBOverlap(*ccAABB, *platform.AABB) {
						cc.MoveHorizontal(platform.AABB.Right()-ccAABB.Left(), world.Transforms[e], world, func(collisionResult *CollisionResult) {
							fmt.Println("Squished.")
						})
					} else if _, ok := ridingCCs[e]; ok {
						cc.MoveHorizontal(moveX, world.Transforms[e], world, nil)
					}
				}
			} else {
				for e, cc := range world.CharacterControllers {
					ccAABB := cc.GetAABB(e, world)
					if AABBOverlap(*ccAABB, *platform.AABB) {
						cc.MoveHorizontal(platform.AABB.Left()-ccAABB.Right(), world.Transforms[e], world, func(collisionResult *CollisionResult) {
							fmt.Println("Squished.")
						})
					} else if _, ok := ridingCCs[e]; ok {
						cc.MoveHorizontal(moveX, world.Transforms[e], world, nil)
					}
				}
			}
		}

		if moveY != 0 {

			platform.Remainder.Y -= moveY
			platform.AABB.Y += moveY

			if moveY > 0 {
				if platform.AABB.Layer != ONE_WAY_PLATFORMS {
					for e, cc := range world.CharacterControllers {
						ccAABB := cc.GetAABB(e, world)
						if AABBOverlap(*ccAABB, *platform.AABB) {
							cc.MoveVertical(e, platform.AABB.Bottom()-ccAABB.Top(), world.Transforms[e], world, func(collisionResult *CollisionResult) {
								fmt.Println("Squished.")
							})
						} else if _, ok := ridingCCs[e]; ok {
							cc.MoveVertical(e, moveY, world.Transforms[e], world, nil)
						}
					}
				} else {
					for e, cc := range world.CharacterControllers {
						if _, ok := ridingCCs[e]; ok {
							cc.MoveVertical(e, moveY, world.Transforms[e], world, nil)
						}
					}
				}
			} else {
				for e, cc := range world.CharacterControllers {
					if platform.AABB.Layer == ONE_WAY_PLATFORMS {

						if cc.CollideWithOneWayPlatform {

							if _, ok := ridingCCs[e]; ok {
								cc.MoveVertical(e, moveY, world.Transforms[e], world, nil)
							}
						}

					} else {
						ccAABB := cc.GetAABB(e, world)
						if AABBOverlap(*ccAABB, *platform.AABB) {
							cc.MoveVertical(e, platform.AABB.Top()-ccAABB.Bottom(), world.Transforms[e], world, func(collisionResult *CollisionResult) {
								fmt.Println("Squished.")
							})
						} else if _, ok := ridingCCs[e]; ok {
							cc.MoveVertical(e, moveY, world.Transforms[e], world, nil)
						}

					}
				}
			}
		}

		platform.Collidable = true
	}
}

func (platform *Platform) GetAllRidingCharacterControllers(platformEntity Entity, world *World) map[Entity]struct{} {
	ridingCharacterControllers := make(map[Entity]struct{})
	for e, cc := range world.CharacterControllers {
		ccAABB := cc.GetAABB(e, world)
		if math.Abs(ccAABB.Bottom()-platform.AABB.Top()) <= 1 &&
			ccAABB.Right() > platform.AABB.Left() &&
			ccAABB.Left() < platform.AABB.Right() {

			ridingCharacterControllers[e] = struct{}{}

		}
	}
	return ridingCharacterControllers
}
