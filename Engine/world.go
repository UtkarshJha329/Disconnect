package Engine

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type World struct {
	Names      map[Entity]*string
	Transforms map[Entity]*Transform
	Sprites    map[Entity]*Sprite
	Parents    map[Entity]*Entity
	Children   map[Entity][]*Entity
	Alive      map[Entity]bool

	Scene Scene
}

func NewWorld() *World {

	world := World{
		Names:      make(map[Entity]*string),
		Transforms: make(map[Entity]*Transform),
		Sprites:    make(map[Entity]*Sprite),
		Parents:    make(map[Entity]*Entity),
		Children:   make(map[Entity][]*Entity),
		Alive:      make(map[Entity]bool),
	}

	world.Scene = Scene{
		Root:                 NewEntity(&world),
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
	// totalWorldTransform := parentWorldTransform
	// totalWorldTransform.Concat(world.Transforms[e].CalculateLocalMatrix())

	// world.Transforms[e].WorldTransformMatrix = totalWorldTransform

	local := world.Transforms[e].CalculateLocalMatrix()

	if parentEntity, ok := world.Parents[e]; ok {
		parentPivot := world.Transforms[*parentEntity].Pivot
		local.Translate(parentPivot.X, parentPivot.Y)
	}

	local.Concat(parentWorldTransform)

	world.Transforms[e].WorldTransformMatrix = local

	// fmt.Print("Entity ")
	// fmt.Print(e)
	// fmt.Println(" world transform matrix: ")
	// fmt.Println(world.Transforms[e].WorldTransformMatrix)

	// pivotX, pivotY := world.Transforms[e].WorldTransformMatrix.Apply(
	// 	world.Transforms[e].Pivot.X,
	// 	world.Transforms[e].Pivot.Y,
	// )

	// fmt.Printf(
	// 	"Entity %d pivot world position = (%f, %f)\n",
	// 	e,
	// 	pivotX,
	// 	pivotY,
	// )
}

func NewEntity(world *World) Entity {
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

func CreateEntityInScene(scene *Scene, world *World) Entity {
	e := NewEntity(world)
	world.MakeEntityAChildOfB(e, scene.Root)
	return e
}

func CreateEntityInSceneWithParent(scene *Scene, world *World, parentEntity Entity) Entity {
	e := NewEntity(world)
	world.MakeEntityAChildOfB(e, parentEntity)
	return e
}

func CreateSpriteFromFileInScene(scene *Scene, world *World, src string) Entity {
	e := CreateEntityInScene(scene, world)

	world.Sprites[e] = &Sprite{}
	world.Sprites[e].SetImageFromFile(src)

	return e
}

func CreateSpriteFromFileInSceneWithParentEntity(scene *Scene, world *World, src string, parentEntity Entity) Entity {
	e := CreateEntityInSceneWithParent(scene, world, parentEntity)

	world.Sprites[e] = &Sprite{}
	world.Sprites[e].SetImageFromFile(src)

	return e
}
