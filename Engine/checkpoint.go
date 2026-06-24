package Engine

type Checkpoint struct {
	Entity            Entity
	Position          Vector2
	Activated         bool
	InputReleaseLimit int
}
