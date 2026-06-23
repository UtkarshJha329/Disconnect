package Engine

type Trigger struct {
	AABB        *AABB
	Collidable  bool
	RoomKey     Vector2
	OnCollision func(triggerEntity Entity, colliderEntity Entity, trigger *Trigger)
	TriggerType string
}
