package Engine

import (
	"cmp"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
)

type Scene struct {
	world *World
	Root  Entity

	EntitiesToDrawSorted []Entity
}

func (scene *Scene) CreateEntitiesToDrawSortedForScene() {
	scene.EntitiesToDrawSorted = scene.EntitiesToDrawSorted[:0]
	scene.CreateEntitiesToDrawNonSortedFromEntity(scene.Root, ebiten.GeoM{})
	slices.SortFunc(scene.EntitiesToDrawSorted, func(a, b Entity) int {
		return cmp.Compare(scene.world.Transforms[a].Position.Z, scene.world.Transforms[b].Position.Z)
	})
}

func (scene *Scene) CreateEntitiesToDrawNonSortedFromEntity(entity Entity, parentWorldTransform ebiten.GeoM) {

	entity.CalculateHierarchyWorldTransform(scene.world, parentWorldTransform)

	if scene.world.Alive[entity] {
		if _, ok := scene.world.Sprites[entity]; ok {
			scene.EntitiesToDrawSorted = append(scene.EntitiesToDrawSorted, entity)
		}
	}

	for _, child := range scene.world.Children[entity] {
		scene.CreateEntitiesToDrawNonSortedFromEntity(*child, scene.world.Transforms[entity].WorldTransformMatrix)
	}
}
