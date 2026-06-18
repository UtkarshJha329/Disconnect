package Engine

type Trigger struct {
	AABB        *AABB
	Collidable  bool
	OnCollision func(triggerEntity Entity, colliderEntity Entity, trigger *Trigger)
}
