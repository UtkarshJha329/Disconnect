package Project

import (
	"Disconnect/Engine"
	"encoding/json"
	"fmt"
	"image/color"
	"math"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	EditorGridSize           = 20.0
	EditorPlatformW          = 60.0
	EditorPlatformH          = 20.0
	EditorTriggerW           = 32.0
	EditorTriggerH           = 32.0
	EditorVelocityStep       = 20.0
	EditorVelocityArrowScale = 0.25
	EditorVelocityArrowMax   = 60.0
	EditorTimerStep          = 0.25
	EditorTimerMin           = 0.25
	EditorStopDurationStep   = 0.1
	EditorStopDurationMin    = 0.1
	LevelSaveFile            = "Assets/level.json"
)

type EditorMode int

const (
	EditorModePlatform EditorMode = iota
	EditorModeTrigger
)

type EditorSelectionType int

const (
	EditorSelectionNone EditorSelectionType = iota
	EditorSelectionPlatform
	EditorSelectionTrigger
)

type EditorState struct {
	Active         bool
	Mode           EditorMode
	HasSelection   bool
	SelectionType  EditorSelectionType
	SelectedEntity Engine.Entity
	StatusMessage  string
}

var Editor EditorState

// --- coordinate helpers ---

func ScreenToWorld(camera *Engine.Camera, cameraEntity Engine.Entity, world *Engine.World, screenX, screenY float64) Engine.Vector2 {
	cameraMatrix := camera.GetCameraTransformMatrix(cameraEntity, world)
	if cameraMatrix.IsInvertible() {
		cameraMatrix.Invert()
	}
	worldX, worldY := cameraMatrix.Apply(screenX, screenY)
	return Engine.Vector2{X: worldX, Y: worldY}
}

func SnapToGrid(value float64) float64 {
	return math.Round(value/EditorGridSize) * EditorGridSize
}

func pointInAABB(x, y float64, aabb *Engine.AABB) bool {
	return x >= aabb.X && x <= aabb.Right() && y >= aabb.Y && y <= aabb.Bottom()
}

// --- update ---

func EditorUpdateFunc(world *Engine.World, dt float64) {
	camera := world.Cameras[MainCameraEntity]

	mx, my := ebiten.CursorPosition()
	worldMouse := ScreenToWorld(camera, MainCameraEntity, world, float64(mx), float64(my))

	// Mode switching
	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		Editor.Mode = EditorModePlatform
		Editor.StatusMessage = "Mode: Platform"
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyK) {
		Editor.Mode = EditorModeTrigger
		Editor.StatusMessage = "Mode: Kill Trigger"
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		handleEditorClick(world, worldMouse)
	}

	// Validate selection still exists
	if Editor.HasSelection {
		switch Editor.SelectionType {
		case EditorSelectionPlatform:
			if _, ok := world.Platforms[Editor.SelectedEntity]; !ok {
				Editor.HasSelection = false
				Editor.SelectionType = EditorSelectionNone
			}
		case EditorSelectionTrigger:
			if _, ok := world.Triggers[Editor.SelectedEntity]; !ok {
				Editor.HasSelection = false
				Editor.SelectionType = EditorSelectionNone
			}
		}
	}

	ctrlHeld := ebiten.IsKeyPressed(ebiten.KeyControlLeft) || ebiten.IsKeyPressed(ebiten.KeyControlRight)

	if Editor.HasSelection {
		switch Editor.SelectionType {
		case EditorSelectionPlatform:
			if !ctrlHeld {
				handleResizeInput(world)
			}
			handleVelocityInput(world)
			handleTimerInput(world)
		case EditorSelectionTrigger:
			handleTriggerResizeInput(world)
		}
		handleDeleteInput(world)
	}

	handleSaveLoadInput(world, ctrlHeld)
}

func handleEditorClick(world *Engine.World, worldMouse Engine.Vector2) {
	// First check if clicking on existing entity (check triggers first since they overlay)
	for e, trigger := range world.Triggers {
		if pointInAABB(worldMouse.X, worldMouse.Y, trigger.AABB) {
			Editor.SelectedEntity = e
			Editor.HasSelection = true
			Editor.SelectionType = EditorSelectionTrigger
			Editor.StatusMessage = "Selected trigger"
			return
		}
	}

	for e, platform := range world.Platforms {
		if pointInAABB(worldMouse.X, worldMouse.Y, platform.AABB) {
			Editor.SelectedEntity = e
			Editor.HasSelection = true
			Editor.SelectionType = EditorSelectionPlatform
			Editor.StatusMessage = "Selected platform"
			return
		}
	}

	// Place new entity based on mode
	snappedX := SnapToGrid(worldMouse.X)
	snappedY := SnapToGrid(worldMouse.Y)

	switch Editor.Mode {
	case EditorModePlatform:
		newEntity := world.CreateNewLevelColliderInScene(
			&world.Scene,
			snappedX,
			snappedY,
			EditorPlatformW,
			EditorPlatformH,
			Engine.DEFAULT,
		)
		Editor.SelectedEntity = newEntity
		Editor.HasSelection = true
		Editor.SelectionType = EditorSelectionPlatform
		Editor.StatusMessage = "Placed new platform"

	case EditorModeTrigger:
		newEntity := world.CreateNewTriggerColliderInScene(
			&world.Scene,
			snappedX,
			snappedY,
			EditorTriggerW,
			EditorTriggerH,
			Engine.DEFAULT,
			nil,
		)
		SetupKillTriggerCallback(world.Triggers[newEntity])
		Editor.SelectedEntity = newEntity
		Editor.HasSelection = true
		Editor.SelectionType = EditorSelectionTrigger
		Editor.StatusMessage = "Placed kill trigger"
	}
}

func handleResizeInput(world *Engine.World) {
	aabb := world.Platforms[Editor.SelectedEntity].AABB

	if inpututil.IsKeyJustPressed(ebiten.KeyD) {
		aabb.Width += EditorGridSize
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyA) {
		aabb.Width = math.Max(EditorGridSize, aabb.Width-EditorGridSize)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyW) {
		aabb.Height += EditorGridSize
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyS) {
		aabb.Height = math.Max(EditorGridSize, aabb.Height-EditorGridSize)
	}
}

func handleTriggerResizeInput(world *Engine.World) {
	aabb := world.Triggers[Editor.SelectedEntity].AABB

	if inpututil.IsKeyJustPressed(ebiten.KeyD) {
		aabb.Width += EditorGridSize
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyA) {
		aabb.Width = math.Max(EditorGridSize, aabb.Width-EditorGridSize)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyW) {
		aabb.Height += EditorGridSize
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyS) {
		aabb.Height = math.Max(EditorGridSize, aabb.Height-EditorGridSize)
	}
}

func handleVelocityInput(world *Engine.World) {
	platform := world.Platforms[Editor.SelectedEntity]

	if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) {
		platform.Velocity.X += EditorVelocityStep
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) {
		platform.Velocity.X -= EditorVelocityStep
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
		platform.Velocity.Y += EditorVelocityStep
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
		platform.Velocity.Y -= EditorVelocityStep
	}
}

func handleTimerInput(world *Engine.World) {
	platform := world.Platforms[Editor.SelectedEntity]

	changed := false

	shiftHeld := ebiten.IsKeyPressed(ebiten.KeyShiftLeft) || ebiten.IsKeyPressed(ebiten.KeyShiftRight)

	if !shiftHeld && inpututil.IsKeyJustPressed(ebiten.KeyBracketLeft) {
		newDur := platform.VelTimerDuration - EditorTimerStep
		if newDur < EditorTimerMin {
			newDur = 0
		}
		platform.VelTimerDuration = newDur
		changed = true
		if newDur == 0 {
			Editor.StatusMessage = "Timer removed"
		} else {
			Editor.StatusMessage = fmt.Sprintf("Timer duration: %.2fs", newDur)
		}
	}
	if !shiftHeld && inpututil.IsKeyJustPressed(ebiten.KeyBracketRight) {
		newDur := platform.VelTimerDuration + EditorTimerStep
		if platform.VelTimerDuration == 0 {
			newDur = EditorTimerMin
		}
		platform.VelTimerDuration = newDur
		changed = true
		Editor.StatusMessage = fmt.Sprintf("Timer duration: %.2fs", newDur)
	}

	if shiftHeld {
		if inpututil.IsKeyJustPressed(ebiten.KeyBracketLeft) {
			cur := platform.VelTimerStopDuration
			if cur <= 0 {
				cur = DefaultVelTimerStopDuration
			}
			newDur := cur - EditorStopDurationStep
			if newDur < EditorStopDurationMin {
				newDur = EditorStopDurationMin
			}
			platform.VelTimerStopDuration = newDur
			changed = true
			Editor.StatusMessage = fmt.Sprintf("Stop duration: %.2fs", newDur)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyBracketRight) {
			cur := platform.VelTimerStopDuration
			if cur <= 0 {
				cur = DefaultVelTimerStopDuration
			}
			platform.VelTimerStopDuration = cur + EditorStopDurationStep
			changed = true
			Editor.StatusMessage = fmt.Sprintf("Stop duration: %.2fs", platform.VelTimerStopDuration)
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyT) {
		platform.VelTimerStopAtEnds = !platform.VelTimerStopAtEnds
		changed = true
		if platform.VelTimerStopAtEnds {
			Editor.StatusMessage = "Stop at ends: ON"
		} else {
			Editor.StatusMessage = "Stop at ends: OFF"
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyY) {
		platform.VelTimerAxis = (platform.VelTimerAxis + 1) % 3
		changed = true
		Editor.StatusMessage = "Timer axis: " + axisName(platform.VelTimerAxis)
	}

	if changed {
		RegisterPlatformTimer(platform)
	}
}

func axisName(axis int) string {
	switch axis {
	case Engine.VelTimerAxisX:
		return "X"
	case Engine.VelTimerAxisY:
		return "Y"
	case Engine.VelTimerAxisBoth:
		return "Both"
	}
	return "?"
}

func handleDeleteInput(world *Engine.World) {
	if inpututil.IsKeyJustPressed(ebiten.KeyDelete) || inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
		switch Editor.SelectionType {
		case EditorSelectionPlatform:
			world.RemovePlatform(Editor.SelectedEntity)
			Editor.StatusMessage = "Deleted platform"
		case EditorSelectionTrigger:
			world.RemoveTrigger(Editor.SelectedEntity)
			Editor.StatusMessage = "Deleted trigger"
		}
		Editor.HasSelection = false
		Editor.SelectionType = EditorSelectionNone
	}
}

func handleSaveLoadInput(world *Engine.World, ctrlHeld bool) {
	if ctrlHeld && inpututil.IsKeyJustPressed(ebiten.KeyS) {
		if err := SaveLevel(LevelSaveFile, world); err != nil {
			Editor.StatusMessage = "Save failed: " + err.Error()
		} else {
			Editor.StatusMessage = "Saved to " + LevelSaveFile
		}
	}

	if ctrlHeld && inpututil.IsKeyJustPressed(ebiten.KeyL) {
		if err := LoadLevel(LevelSaveFile, world); err != nil {
			Editor.StatusMessage = "Load failed: " + err.Error()
		} else {
			Editor.HasSelection = false
			Editor.SelectionType = EditorSelectionNone
			Editor.StatusMessage = "Loaded from " + LevelSaveFile
		}
	}
}

// --- serialization ---

type SerializablePlatform struct {
	X, Y, Width, Height  float64
	Layer                uint
	VelX, VelY           float64
	VelTimerDuration     float64
	VelTimerStopAtEnds   bool
	VelTimerStopDuration float64
	VelTimerAxis         int
}

type SerializableTrigger struct {
	X, Y, Width, Height float64
	Layer               uint
	TriggerType         string
}

type SerializableLevel struct {
	Platforms []SerializablePlatform `json:"platforms"`
	Triggers  []SerializableTrigger  `json:"triggers"`
}

func SaveLevel(path string, world *Engine.World) error {
	platforms := make([]SerializablePlatform, 0, len(world.Platforms))
	for _, platform := range world.Platforms {
		platforms = append(platforms, SerializablePlatform{
			X:                    platform.AABB.X,
			Y:                    platform.AABB.Y,
			Width:                platform.AABB.Width,
			Height:               platform.AABB.Height,
			Layer:                platform.AABB.Layer,
			VelX:                 platform.Velocity.X,
			VelY:                 platform.Velocity.Y,
			VelTimerDuration:     platform.VelTimerDuration,
			VelTimerStopAtEnds:   platform.VelTimerStopAtEnds,
			VelTimerStopDuration: platform.VelTimerStopDuration,
			VelTimerAxis:         platform.VelTimerAxis,
		})
	}

	triggers := make([]SerializableTrigger, 0, len(world.Triggers))
	for _, trigger := range world.Triggers {
		triggers = append(triggers, SerializableTrigger{
			X:           trigger.AABB.X,
			Y:           trigger.AABB.Y,
			Width:       trigger.AABB.Width,
			Height:      trigger.AABB.Height,
			Layer:       trigger.AABB.Layer,
			TriggerType: trigger.TriggerType,
		})
	}

	level := SerializableLevel{
		Platforms: platforms,
		Triggers:  triggers,
	}

	data, err := json.MarshalIndent(level, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func LoadLevel(path string, world *Engine.World) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var level SerializableLevel
	if err := json.Unmarshal(data, &level); err != nil {
		return err
	}

	// Clear existing platforms
	for e := range world.Platforms {
		world.RemovePlatform(e)
	}

	// Clear existing triggers (except those created in InitFuncs)
	// We'll track which triggers are from the level file
	for e := range world.Triggers {
		world.RemoveTrigger(e)
	}

	// Load platforms
	for _, p := range level.Platforms {
		e := world.CreateNewLevelColliderInScene(&world.Scene, p.X, p.Y, p.Width, p.Height, p.Layer)
		platform := world.Platforms[e]
		platform.Velocity = Engine.Vector2{X: p.VelX, Y: p.VelY}
		platform.VelTimerDuration = p.VelTimerDuration
		platform.VelTimerStopAtEnds = p.VelTimerStopAtEnds
		platform.VelTimerStopDuration = p.VelTimerStopDuration
		platform.VelTimerAxis = p.VelTimerAxis
		RegisterPlatformTimer(platform)
	}

	// Load triggers
	for _, t := range level.Triggers {
		e := world.CreateNewTriggerColliderInScene(
			&world.Scene,
			t.X, t.Y, t.Width, t.Height, t.Layer,
			nil,
		)
		trigger := world.Triggers[e]
		trigger.TriggerType = t.TriggerType

		// Assign callback based on type
		switch t.TriggerType {
		case "kill":
			SetupKillTriggerCallback(trigger)
		}
	}

	return nil
}

// --- drawing ---

func EditorDrawFunc(camera *Engine.Camera, cameraEntity Engine.Entity, world *Engine.World) {
	cameraMatrix := camera.GetCameraTransformMatrix(cameraEntity, world)

	drawEditorGrid(camera, cameraMatrix)
	drawPlacementPreview(camera, cameraEntity, world, cameraMatrix)
	drawAllTriggersInEditor(camera, cameraMatrix, world)

	if Editor.HasSelection {
		switch Editor.SelectionType {
		case EditorSelectionPlatform:
			if platform, ok := world.Platforms[Editor.SelectedEntity]; ok {
				drawSelectionHighlight(camera, cameraMatrix, platform.AABB, color.RGBA{R: 255, G: 255, B: 0, A: 255})
				drawVelocityArrow(camera, cameraMatrix, platform)
				drawTimerArc(camera, cameraMatrix, platform)
			}
		case EditorSelectionTrigger:
			if trigger, ok := world.Triggers[Editor.SelectedEntity]; ok {
				drawSelectionHighlight(camera, cameraMatrix, trigger.AABB, color.RGBA{R: 255, G: 100, B: 255, A: 255})
			}
		}
	}

	drawEditorHUD(world)
}

func drawEditorGrid(camera *Engine.Camera, cameraMatrix *ebiten.GeoM) {
	inverse := *cameraMatrix
	if inverse.IsInvertible() {
		inverse.Invert()
	}

	topLeftX, topLeftY := inverse.Apply(0, 0)
	bottomRightX, bottomRightY := inverse.Apply(camera.ScreenDims.X, camera.ScreenDims.Y)

	minX := math.Min(topLeftX, bottomRightX)
	maxX := math.Max(topLeftX, bottomRightX)
	minY := math.Min(topLeftY, bottomRightY)
	maxY := math.Max(topLeftY, bottomRightY)

	gridColor := color.RGBA{R: 20, G: 20, B: 20, A: 35}

	for x := SnapToGrid(minX); x <= maxX; x += EditorGridSize {
		sx0, sy0 := cameraMatrix.Apply(x, minY)
		sx1, sy1 := cameraMatrix.Apply(x, maxY)
		vector.StrokeLine(camera.RenderTexture, float32(sx0), float32(sy0), float32(sx1), float32(sy1), 1, gridColor, false)
	}

	for y := SnapToGrid(minY); y <= maxY; y += EditorGridSize {
		sx0, sy0 := cameraMatrix.Apply(minX, y)
		sx1, sy1 := cameraMatrix.Apply(maxX, y)
		vector.StrokeLine(camera.RenderTexture, float32(sx0), float32(sy0), float32(sx1), float32(sy1), 1, gridColor, false)
	}
}

func drawPlacementPreview(camera *Engine.Camera, cameraEntity Engine.Entity, world *Engine.World, cameraMatrix *ebiten.GeoM) {
	mx, my := ebiten.CursorPosition()
	worldMouse := ScreenToWorld(camera, cameraEntity, world, float64(mx), float64(my))

	// Don't show preview if hovering over existing entity
	for _, trigger := range world.Triggers {
		if pointInAABB(worldMouse.X, worldMouse.Y, trigger.AABB) {
			return
		}
	}
	for _, platform := range world.Platforms {
		if pointInAABB(worldMouse.X, worldMouse.Y, platform.AABB) {
			return
		}
	}

	snappedX := SnapToGrid(worldMouse.X)
	snappedY := SnapToGrid(worldMouse.Y)
	sx, sy := cameraMatrix.Apply(snappedX, snappedY)

	var previewW, previewH float64
	var previewColor color.RGBA

	switch Editor.Mode {
	case EditorModePlatform:
		previewW = EditorPlatformW
		previewH = EditorPlatformH
		previewColor = color.RGBA{R: 255, G: 255, B: 255, A: 150}
	case EditorModeTrigger:
		previewW = EditorTriggerW
		previewH = EditorTriggerH
		previewColor = color.RGBA{R: 255, G: 80, B: 80, A: 150}
	}

	vector.StrokeRect(
		camera.RenderTexture,
		float32(sx), float32(sy),
		float32(previewW), float32(previewH),
		1,
		previewColor,
		false,
	)
}

func drawAllTriggersInEditor(camera *Engine.Camera, cameraMatrix *ebiten.GeoM, world *Engine.World) {
	for _, trigger := range world.Triggers {
		aabb := trigger.AABB
		x, y := cameraMatrix.Apply(aabb.X, aabb.Y)

		// Red fill with transparency
		vector.FillRect(
			camera.RenderTexture,
			float32(x), float32(y),
			float32(aabb.Width), float32(aabb.Height),
			color.RGBA{R: 255, G: 0, B: 0, A: 40},
			false,
		)

		// Red outline
		vector.StrokeRect(
			camera.RenderTexture,
			float32(x), float32(y),
			float32(aabb.Width), float32(aabb.Height),
			1,
			color.RGBA{R: 255, G: 0, B: 0, A: 180},
			false,
		)

		// Draw X pattern to indicate danger
		vector.StrokeLine(
			camera.RenderTexture,
			float32(x), float32(y),
			float32(x+aabb.Width), float32(y+aabb.Height),
			1,
			color.RGBA{R: 255, G: 0, B: 0, A: 120},
			false,
		)
		vector.StrokeLine(
			camera.RenderTexture,
			float32(x+aabb.Width), float32(y),
			float32(x), float32(y+aabb.Height),
			1,
			color.RGBA{R: 255, G: 0, B: 0, A: 120},
			false,
		)
	}
}

func drawSelectionHighlight(camera *Engine.Camera, cameraMatrix *ebiten.GeoM, aabb *Engine.AABB, highlightColor color.RGBA) {
	x, y := cameraMatrix.Apply(aabb.X-2, aabb.Y-2)

	vector.StrokeRect(
		camera.RenderTexture,
		float32(x), float32(y),
		float32(aabb.Width+4), float32(aabb.Height+4),
		2,
		highlightColor,
		false,
	)
}

func drawVelocityArrow(camera *Engine.Camera, cameraMatrix *ebiten.GeoM, platform *Engine.Platform) {
	if platform.Velocity.X == 0 && platform.Velocity.Y == 0 {
		return
	}

	aabb := platform.AABB
	centerX := aabb.X + aabb.Width/2.0
	centerY := aabb.Y + aabb.Height/2.0

	dx := platform.Velocity.X * EditorVelocityArrowScale
	dy := platform.Velocity.Y * EditorVelocityArrowScale
	length := math.Hypot(dx, dy)
	if length > EditorVelocityArrowMax {
		scale := EditorVelocityArrowMax / length
		dx *= scale
		dy *= scale
	}

	sx0, sy0 := cameraMatrix.Apply(centerX, centerY)
	sx1, sy1 := cameraMatrix.Apply(centerX+dx, centerY+dy)

	arrowColor := color.RGBA{R: 0, G: 255, B: 255, A: 255}
	vector.StrokeLine(camera.RenderTexture, float32(sx0), float32(sy0), float32(sx1), float32(sy1), 2, arrowColor, false)

	angle := math.Atan2(float64(sy1-sy0), float64(sx1-sx0))
	const headLength = 8.0
	const headAngle = 0.45

	leftX := sx1 - headLength*math.Cos(angle-headAngle)
	leftY := sy1 - headLength*math.Sin(angle-headAngle)
	rightX := sx1 - headLength*math.Cos(angle+headAngle)
	rightY := sy1 - headLength*math.Sin(angle+headAngle)

	vector.StrokeLine(camera.RenderTexture, float32(sx1), float32(sy1), float32(leftX), float32(leftY), 2, arrowColor, false)
	vector.StrokeLine(camera.RenderTexture, float32(sx1), float32(sy1), float32(rightX), float32(rightY), 2, arrowColor, false)

	ebitenutil.DebugPrintAt(
		camera.RenderTexture,
		fmt.Sprintf("vx:%.0f vy:%.0f", platform.Velocity.X, platform.Velocity.Y),
		int(sx1)+6, int(sy1)-6,
	)
}

func drawTimerArc(camera *Engine.Camera, cameraMatrix *ebiten.GeoM, platform *Engine.Platform) {
	if platform.VelTimerDuration <= 0 {
		return
	}

	aabb := platform.AABB
	centerX := aabb.X + aabb.Width/2.0
	centerY := aabb.Y - 16.0

	scx, scy := cameraMatrix.Apply(centerX, centerY)

	const radius = 10.0
	const segments = 24
	const twoPi = math.Pi * 2

	arcColor := color.RGBA{R: 255, G: 165, B: 0, A: 220}
	if platform.VelTimerStopAtEnds {
		arcColor = color.RGBA{R: 220, G: 80, B: 255, A: 220}
	}

	for i := 0; i < segments; i++ {
		a0 := twoPi * float64(i) / float64(segments)
		a1 := twoPi * float64(i+1) / float64(segments)
		x0 := scx + radius*math.Cos(a0)
		y0 := scy + radius*math.Sin(a0)
		x1 := scx + radius*math.Cos(a1)
		y1 := scy + radius*math.Sin(a1)
		vector.StrokeLine(camera.RenderTexture, float32(x0), float32(y0), float32(x1), float32(y1), 1, arcColor, false)
	}

	progress := 0.0
	if item := platform.VelTimerPoolItem(); item != nil {
		elapsed := item.Item.TimerTotalDuration - item.Item.TimerDurationRemaining
		if item.Item.TimerTotalDuration > 0 {
			progress = elapsed / item.Item.TimerTotalDuration
		}
	}

	tickAngle := -math.Pi/2 + twoPi*progress
	tickX := scx + radius*math.Cos(tickAngle)
	tickY := scy + radius*math.Sin(tickAngle)
	vector.StrokeLine(camera.RenderTexture, float32(scx), float32(scy), float32(tickX), float32(tickY), 2, color.RGBA{R: 255, G: 255, B: 255, A: 200}, false)

	label := fmt.Sprintf("%.2fs %s", platform.VelTimerDuration, axisName(platform.VelTimerAxis))
	if platform.VelTimerStopAtEnds {
		stopDur := platform.VelTimerStopDuration
		if stopDur <= 0 {
			stopDur = DefaultVelTimerStopDuration
		}
		label += fmt.Sprintf(" [stop %.2fs]", stopDur)
	}
	ebitenutil.DebugPrintAt(camera.RenderTexture, label, int(scx)-len(label)*3, int(scy)+14)
}

func drawEditorHUD(world *Engine.World) {
	camera := world.Cameras[MainCameraEntity]

	modeStr := "PLATFORM"
	if Editor.Mode == EditorModeTrigger {
		modeStr = "KILL TRIGGER"
	}

	lines := fmt.Sprintf("EDITOR MODE (Tab to exit)\n"+
		"Current: %s [P/K to switch]\n"+
		"Click: place/select\n"+
		"W/S height, A/D width\n",
		modeStr)

	if Editor.Mode == EditorModePlatform {
		lines += "Arrows: nudge velocity\n" +
			"[ / ]: timer duration -/+\n" +
			"Shift+[ / Shift+]: stop duration -/+\n" +
			"T: toggle stop-at-ends\n" +
			"Y: cycle timer axis (X/Y/Both)\n"
	}

	lines += "Delete: remove selected\n" +
		"Ctrl+S save, Ctrl+L load"

	if Editor.HasSelection {
		switch Editor.SelectionType {
		case EditorSelectionPlatform:
			if platform, ok := world.Platforms[Editor.SelectedEntity]; ok {
				aabb := platform.AABB
				timerStr := "none"
				if platform.VelTimerDuration > 0 {
					stopStr := ""
					if platform.VelTimerStopAtEnds {
						stopDur := platform.VelTimerStopDuration
						if stopDur <= 0 {
							stopDur = DefaultVelTimerStopDuration
						}
						stopStr = fmt.Sprintf(" stop(%.2fs)", stopDur)
					}
					timerStr = fmt.Sprintf("%.2fs %s%s", platform.VelTimerDuration, axisName(platform.VelTimerAxis), stopStr)
				}
				lines += fmt.Sprintf(
					"\n\nPlatform: pos(%.0f, %.0f) size(%.0f x %.0f)\nvel(%.0f, %.0f)  timer:%s",
					aabb.X, aabb.Y, aabb.Width, aabb.Height,
					platform.Velocity.X, platform.Velocity.Y,
					timerStr,
				)
			}
		case EditorSelectionTrigger:
			if trigger, ok := world.Triggers[Editor.SelectedEntity]; ok {
				aabb := trigger.AABB
				lines += fmt.Sprintf(
					"\n\nKill Trigger: pos(%.0f, %.0f) size(%.0f x %.0f)",
					aabb.X, aabb.Y, aabb.Width, aabb.Height,
				)
			}
		}
	}

	if Editor.StatusMessage != "" {
		lines += "\n" + Editor.StatusMessage
	}

	ebitenutil.DebugPrintAt(camera.RenderTexture, lines, 4, 4)
}
