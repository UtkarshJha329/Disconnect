package Engine

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type World struct {
	Names                map[Entity]*string
	Transforms           map[Entity]*Transform
	Sprites              map[Entity]*Sprite
	CharacterControllers map[Entity]*CharacterController

	Parents  map[Entity]*Entity
	Children map[Entity][]*Entity
	Alive    map[Entity]bool

	LevelColliders []*AABB

	Scene Scene
}

func NewWorld() *World {

	world := World{
		Names:                make(map[Entity]*string),
		Transforms:           make(map[Entity]*Transform),
		Sprites:              make(map[Entity]*Sprite),
		CharacterControllers: make(map[Entity]*CharacterController),

		Parents:  make(map[Entity]*Entity),
		Children: make(map[Entity][]*Entity),
		Alive:    make(map[Entity]bool),

		LevelColliders: make([]*AABB, 0),
	}

	world.Scene = Scene{
		Root:                 world.NewEntity(),
		world:                &world,
		EntitiesToDrawSorted: make([]Entity, 0),
	}

	return &world
}

func (world *World) MakeEntityAChildOfB(a, b Entity) {
	world.Children[b] = append(world.Children[b], &a)
	world.Parents[a] = &b
}

func (e Entity) CalculateHierarchyWorldTransform(world *World, parentWorldTransform ebiten.GeoM) {
	local := world.Transforms[e].CalculateLocalMatrix()

	if parentEntity, ok := world.Parents[e]; ok {
		parentPivot := world.Transforms[*parentEntity].Pivot
		local.Translate(parentPivot.X, parentPivot.Y)
	}

	local.Concat(parentWorldTransform)

	world.Transforms[e].WorldTransformMatrix = local
}

func (world *World) NewEntity() Entity {
	e := nextEntity
	world.Alive[e] = true
	world.Transforms[e] = &Transform{
		Position: Vector3{X: 0, Y: 0, Z: 0},
		Rotation: 0,
		Scale:    Vector2{X: 1, Y: 1},
	}
	world.Children[e] = make([]*Entity, 0)
	nextEntity++
	return e
}

func (world *World) CreateEntityInScene(scene *Scene) Entity {
	e := world.NewEntity()
	world.MakeEntityAChildOfB(e, scene.Root)
	return e
}

func (world *World) CreateEntityInSceneWithParent(scene *Scene, parentEntity Entity) Entity {
	e := world.NewEntity()
	world.MakeEntityAChildOfB(e, parentEntity)
	return e
}

func (world *World) CreateSpriteFromFileInScene(scene *Scene, src string) Entity {
	e := world.CreateEntityInScene(scene)

	world.Sprites[e] = &Sprite{}
	world.Sprites[e].SetImageFromFile(src)

	return e
}

func (world *World) CreateSpriteFromFileInSceneWithParentEntity(scene *Scene, src string, parentEntity Entity) Entity {
	e := world.CreateEntityInSceneWithParent(scene, parentEntity)

	world.Sprites[e] = &Sprite{}
	world.Sprites[e].SetImageFromFile(src)

	return e
}

func (world *World) CreateNewLevelCollider(X, Y, Width, Height float64) int {
	world.LevelColliders = append(world.LevelColliders, &AABB{
		X:      X,
		Y:      Y,
		Width:  Width,
		Height: Height,
	})
	return len(world.LevelColliders) - 1
}
