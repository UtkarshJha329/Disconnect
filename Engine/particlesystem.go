package Engine

import (
	"image/color"
	"math"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2/vector"
)

// ParticleConfig holds the parameters for a specific burst of particles.
// You can create different configs for explosions, smoke, sparks, etc.
type ParticleConfig struct {
	EmissionRadius float64
	SpeedMin       float64
	SpeedMax       float64
	LifeMin        float64
	LifeMax        float64
	StartSize      float64
	EndSize        float64
	StartColor     color.RGBA
	EndColor       color.RGBA
	Gravity        float64
	Drag           float64
}

// Particle holds the raw data for a single particle.
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
	Gravity  float64 // Per-particle gravity
	Drag     float64 // Per-particle drag
}

// ParticleSystem now just manages the pool, the owner, and the offset.
type ParticleSystem struct {
	Name      string
	Particles Pool[Particle]
	Owner     Entity
	Offset    Vector2
}

// InitParticleSystem sets up the pool
func (ps *ParticleSystem) InitParticleSystem(name string, capacity int) {
	ps.Name = name
	ps.Particles.InitPool(name, capacity)
}

// SpawnBurst emits a number of particles using the provided ParticleConfig
func (ps *ParticleSystem) SpawnBurst(world *World, count int, cfg ParticleConfig) {
	// Get the owner's position
	origin := Vector2{}
	if t, ok := world.Transforms[ps.Owner]; ok {
		origin = Vector2{X: t.Position.X + ps.Offset.X, Y: t.Position.Y + ps.Offset.Y}
	}

	for i := 0; i < count; i++ {
		if ps.Particles.IsPoolFilled() {
			break // Pool is full, stop spawning
		}

		poolItem := ps.Particles.GetAnUnusedItemFromPool()
		p := &poolItem.Item

		p.Alive = true

		// Random point in a circle for spawn position
		theta := rand.Float64() * 2 * math.Pi
		r := cfg.EmissionRadius * math.Sqrt(rand.Float64())
		p.Position = Vector2{
			X: origin.X + r*math.Cos(theta),
			Y: origin.Y + r*math.Sin(theta),
		}

		// Outward velocity
		speed := cfg.SpeedMin + rand.Float64()*(cfg.SpeedMax-cfg.SpeedMin)
		p.Velocity = Vector2{
			X: math.Cos(theta) * speed,
			Y: math.Sin(theta) * speed,
		}

		// Life
		life := cfg.LifeMin + rand.Float64()*(cfg.LifeMax-cfg.LifeMin)
		p.Life = life
		p.MaxLife = life

		// Visuals
		p.Size = cfg.StartSize
		p.EndSize = cfg.EndSize
		p.Color = cfg.StartColor
		p.EndColor = cfg.EndColor

		// Dynamics (stored per-particle so one system can handle mixed types)
		p.Gravity = cfg.Gravity
		p.Drag = cfg.Drag
	}
}

// UpdateParticleSystem moves particles and returns dead ones to the pool
func (ps *ParticleSystem) UpdateParticleSystem(dt float64) {
	// Iterate backwards so we can safely kill items while iterating
	ps.Particles.PerformOperationOnAlivePoolItemsBackwards(func(poolItem *PoolItem[Particle]) {
		p := &poolItem.Item

		p.Life -= dt
		if p.Life <= 0 {
			p.Alive = false
			ps.Particles.KillItemInPool(poolItem)
			return
		}

		// Apply gravity and drag (using the per-particle values)
		p.Velocity.Y += p.Gravity * dt
		if p.Drag > 0 {
			dragFactor := math.Pow(1.0-p.Drag, dt*60.0)
			p.Velocity.X *= dragFactor
			p.Velocity.Y *= dragFactor
		}

		// Move
		p.Position.X += p.Velocity.X * dt
		p.Position.Y += p.Velocity.Y * dt
	})
}

// ParticlesUpdateFunc is called by your world loop
func ParticlesUpdateFunc(world *World, dt float64) {
	for _, ps := range world.ParticleSystems {
		ps.UpdateParticleSystem(dt)
	}
}

// DrawParticleSystems renders all active particles
func DrawParticleSystems(camera *Camera, cameraEntity Entity, world *World) {
	cameraMatrix := camera.GetCameraTransformMatrix(cameraEntity, world)

	for _, ps := range world.ParticleSystems {
		ps.Particles.PerformOperationOnAlivePoolItems(func(poolItem *PoolItem[Particle]) {
			p := &poolItem.Item

			// Calculate progress 0.0 to 1.0
			t := 1.0 - (p.Life / p.MaxLife)

			// Interpolate size
			curSize := p.Size + (p.EndSize-p.Size)*t
			if curSize <= 0 {
				return
			}

			// Interpolate color
			curColor := color.RGBA{
				R: uint8(float64(p.Color.R) + (float64(p.EndColor.R)-float64(p.Color.R))*t),
				G: uint8(float64(p.Color.G) + (float64(p.EndColor.G)-float64(p.Color.G))*t),
				B: uint8(float64(p.Color.B) + (float64(p.EndColor.B)-float64(p.Color.B))*t),
				A: uint8(float64(p.Color.A) + (float64(p.EndColor.A)-float64(p.Color.A))*t),
			}

			// Apply camera transform
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
