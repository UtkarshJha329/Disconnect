package Engine

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type EntityInitFunc func(*World)
type EntityUpdateFunc func(*World, float64)

// WorldInstance is set during NewWorld so that callbacks can reference it.
var WorldInstance *World

type World struct {
	Names map[Entity]*string

	Transforms      map[Entity]*Transform
	Sprites         map[Entity]*Sprite
	AnimatedSprites map[Entity]*AnimatedSprite

	CharacterControllers map[Entity]*CharacterController
	Platforms            map[Entity]*Platform
	Triggers             map[Entity]*Trigger

	ParticleSystems map[Entity]*ParticleSystem

	Cameras map[Entity]*Camera

	Parents  map[Entity]Entity
	Children map[Entity][]Entity
	Alive    map[Entity]bool

	Scene             Scene
	EntityInitfuncs   []EntityInitFunc
	EntityUpdateFuncs []EntityUpdateFunc
}

func NewWorld() *World {

	world := World{
		Names: make(map[Entity]*string),

		Transforms:      make(map[Entity]*Transform),
		Sprites:         make(map[Entity]*Sprite),
		AnimatedSprites: make(map[Entity]*AnimatedSprite),

		CharacterControllers: make(map[Entity]*CharacterController),
		Platforms:            make(map[Entity]*Platform),
		Triggers:             make(map[Entity]*Trigger),

		ParticleSystems: make(map[Entity]*ParticleSystem),

		Cameras: make(map[Entity]*Camera),

		Parents:  make(map[Entity]Entity),
		Children: make(map[Entity][]Entity),
		Alive:    make(map[Entity]bool),
	}

	world.Scene = Scene{
		Root:                 world.NewEntity(),
		world:                &world,
		EntitiesToDrawSorted: make([]Entity, 0),
	}

	WorldInstance = &world
	return &world
}

func (world *World) MakeEntityAChildOfB(a, b Entity) {
	world.Children[b] = append(world.Children[b], a)
	world.Parents[a] = b
}

func (e Entity) CalculateHierarchyWorldTransform(world *World, parentWorldTransform ebiten.GeoM) {
	local := world.Transforms[e].CalculateLocalMatrix()

	if parentEntity, ok := world.Parents[e]; ok {
		parentPivot := world.Transforms[parentEntity].Pivot
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
	world.Children[e] = make([]Entity, 0)
	nextEntity++
	return e
}

func (world *World) SetEntityAlive(entityToAlive Entity) {

	world.Alive[entityToAlive] = true

	for _, child := range world.Children[entityToAlive] {
		world.SetEntityAlive(child)
	}

}

func (world *World) KillEntity(entityToKill Entity) {

	world.Alive[entityToKill] = false

	for _, child := range world.Children[entityToKill] {
		world.KillEntity(child)
	}

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

func (world *World) CreateNewLevelColliderInScene(scene *Scene, X, Y, Width, Height float64, Layer uint) Entity {
	e := world.CreateEntityInScene(scene)
	world.Platforms[e] = &Platform{
		AABB: &AABB{
			X:      X,
			Y:      Y,
			Width:  Width,
			Height: Height,
			Layer:  Layer,
		},
		Collidable: true,
	}
	return e
}

func (world *World) CreateNewTriggerColliderInScene(scene *Scene, X, Y, Width, Height float64, Layer uint, OnCollision func(triggerEntity Entity, colliderEntity Entity, trigger *Trigger)) Entity {
	e := world.CreateEntityInScene(scene)
	world.Triggers[e] = &Trigger{
		AABB: &AABB{
			X:      X,
			Y:      Y,
			Width:  Width,
			Height: Height,
			Layer:  Layer,
		},
		Collidable:  true,
		OnCollision: OnCollision,
	}
	return e
}

func (world *World) NewCameraInScene(scene *Scene, ScreenWidth, ScreenHeight float64) Entity {
	e := world.CreateEntityInScene(scene)
	world.Cameras[e] = &Camera{
		ScreenDims:    Vector2{X: ScreenWidth, Y: ScreenHeight},
		RenderTexture: ebiten.NewImage(int(ScreenWidth), int(ScreenHeight)),
	}
	return e
}

func (world *World) PerformTriggerCharacterColliderChecks() {

	for triggerEntity, trigger := range world.Triggers {

		if !world.Alive[triggerEntity] || !trigger.Collidable {
			continue
		}

		for characterEntity, character := range world.CharacterControllers {

			if !world.Alive[characterEntity] {
				continue
			}

			if AABBOverlap(*trigger.AABB, *character.GetAABB(characterEntity, world)) {
				trigger.OnCollision(triggerEntity, characterEntity, trigger)
			}

		}
	}
}

func (world *World) RemovePlatform(e Entity) {
	delete(world.Platforms, e)
	delete(world.Transforms, e)
	delete(world.Alive, e)
	delete(world.Children, e)

	if parent, ok := world.Parents[e]; ok {
		siblings := world.Children[parent]
		for i, child := range siblings {
			if child == e {
				world.Children[parent] = append(siblings[:i], siblings[i+1:]...)
				break
			}
		}
		delete(world.Parents, e)
	}
}

func (world *World) RemoveTrigger(e Entity) {
	delete(world.Triggers, e)
	delete(world.Transforms, e)
	delete(world.Alive, e)
	delete(world.Children, e)

	if parent, ok := world.Parents[e]; ok {
		siblings := world.Children[parent]
		for i, child := range siblings {
			if child == e {
				world.Children[parent] = append(siblings[:i], siblings[i+1:]...)
				break
			}
		}
		delete(world.Parents, e)
	}
}
