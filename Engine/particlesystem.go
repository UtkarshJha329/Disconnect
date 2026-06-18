package Engine

import (
	"fmt"
	"image/color"
	"math"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2/vector"
)

type ParticleEmissionConfig struct {
	ConfigName        string
	EmissionRadius    float64
	SpeedMin          float64
	SpeedMax          float64
	LifeMin           float64
	LifeMax           float64
	StartSize         float64
	EndSize           float64
	StartColor        color.RGBA
	EndColor          color.RGBA
	Gravity           float64
	Drag              float64
	TotalNumParticles int
}

type Particle struct {
	Alive    bool
	Life     float64
	MaxLife  float64
	Position Vector2
	Velocity Vector2
	Size     float64
	EndSize  float64
	Color    color.RGBA
	EndColor color.RGBA
}

type ParticleSystem struct {
	Name      string
	Particles map[*ParticleEmissionConfig]*Pool[Particle]
	Owner     Entity
	Offset    Vector2
}

func (ps *ParticleSystem) InitParticleSystem(name string, capacity int) {
	ps.Name = name
	ps.Particles = make(map[*ParticleEmissionConfig]*Pool[Particle])
}

func (ps *ParticleSystem) SpawnBurst(world *World, count int, cfg *ParticleEmissionConfig) {
	origin := Vector2{}
	if t, ok := world.Transforms[ps.Owner]; ok {
		origin = Vector2{X: t.Position.X + ps.Offset.X, Y: t.Position.Y + ps.Offset.Y}
	}

	if _, ok := ps.Particles[cfg]; !ok {
		ps.Particles[cfg] = &Pool[Particle]{}
		ps.Particles[cfg].InitPool(ps.Name+"'s "+cfg.ConfigName+" Particle Pool.", cfg.TotalNumParticles)
	}

	currentParticlesPool := ps.Particles[cfg]

	for i := 0; i < count; i++ {
		if currentParticlesPool.IsPoolFilled() {
			fmt.Println(currentParticlesPool.PoolName + "Too many particles unable to spawn more.")
			break // Pool is full, stop spawning
		}

		poolItem := currentParticlesPool.GetAnUnusedItemFromPool()
		p := &poolItem.Item

		p.Alive = true

		theta := rand.Float64() * 2 * math.Pi
		r := cfg.EmissionRadius * math.Sqrt(rand.Float64())
		p.Position = Vector2{
			X: origin.X + r*math.Cos(theta),
			Y: origin.Y + r*math.Sin(theta),
		}

		speed := cfg.SpeedMin + rand.Float64()*(cfg.SpeedMax-cfg.SpeedMin)
		p.Velocity = Vector2{
			X: math.Cos(theta) * speed,
			Y: math.Sin(theta) * speed,
		}

		life := cfg.LifeMin + rand.Float64()*(cfg.LifeMax-cfg.LifeMin)
		p.Life = life
		p.MaxLife = life

		p.Size = cfg.StartSize
		p.EndSize = cfg.EndSize
		p.Color = cfg.StartColor
		p.EndColor = cfg.EndColor
	}
}

func (ps *ParticleSystem) UpdateParticleSystem(dt float64) {

	for cfg, particlePool := range ps.Particles {

		particlePool.PerformOperationOnAlivePoolItemsBackwards(func(poolItem *PoolItem[Particle]) {
			p := &poolItem.Item

			p.Life -= dt
			if p.Life <= 0 {
				p.Alive = false
				particlePool.KillItemInPool(poolItem)
				return
			}

			p.Velocity.Y += cfg.Gravity * dt
			if cfg.Drag > 0 {
				dragFactor := math.Pow(1.0-cfg.Drag, dt*60.0)
				p.Velocity.X *= dragFactor
				p.Velocity.Y *= dragFactor
			}

			p.Position.X += p.Velocity.X * dt
			p.Position.Y += p.Velocity.Y * dt
		})
	}
}

func ParticlesUpdateFunc(world *World, dt float64) {
	for _, ps := range world.ParticleSystems {
		ps.UpdateParticleSystem(dt)
	}
}

func DrawParticleSystems(camera *Camera, cameraEntity Entity, world *World) {
	cameraMatrix := camera.GetCameraTransformMatrix(cameraEntity, world)

	for _, ps := range world.ParticleSystems {

		for _, particlePool := range ps.Particles {

			particlePool.PerformOperationOnAlivePoolItemsBackwards(func(poolItem *PoolItem[Particle]) {
				p := &poolItem.Item

				t := 1.0 - (p.Life / p.MaxLife)

				curSize := p.Size + (p.EndSize-p.Size)*t
				if curSize <= 0 {
					return
				}

				curColor := color.RGBA{
					R: uint8(float64(p.Color.R) + (float64(p.EndColor.R)-float64(p.Color.R))*t),
					G: uint8(float64(p.Color.G) + (float64(p.EndColor.G)-float64(p.Color.G))*t),
					B: uint8(float64(p.Color.B) + (float64(p.EndColor.B)-float64(p.Color.B))*t),
					A: uint8(float64(p.Color.A) + (float64(p.EndColor.A)-float64(p.Color.A))*t),
				}

				screenX, screenY := cameraMatrix.Apply(p.Position.X, p.Position.Y)

				halfSize := float32(curSize / 2.0)
				vector.FillRect(
					camera.RenderTexture,
					float32(screenX)-halfSize,
					float32(screenY)-halfSize,
					float32(curSize),
					float32(curSize),
					curColor,
					false,
				)
			})
		}
	}
}
