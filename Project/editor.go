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
	EditorPlatformW          = 20.0
	EditorPlatformH          = 20.0
	EditorTriggerW           = 20.0
	EditorTriggerH           = 20.0
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
	EditorModeCheckpoint
)

type EditorSelectionType int

const (
	EditorSelectionNone EditorSelectionType = iota
	EditorSelectionPlatform
	EditorSelectionTrigger
	EditorSelectionCheckpoint
)

type EditorState struct {
	Active         bool
	Mode           EditorMode
	HasSelection   bool
	SelectionType  EditorSelectionType
	SelectedEntity Engine.Entity
	StatusMessage  string
	ViewedRoom     Engine.Vector2
}

var Editor = EditorState{ViewedRoom: Engine.Vector2{X: 0, Y: 0}}
var (
	editorIsPanning  bool
	editorLastMouseX int
	editorLastMouseY int
)

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

	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
		if !editorIsPanning {
			editorIsPanning = true
			editorLastMouseX, editorLastMouseY = mx, my
		} else {
			dx := float64(mx - editorLastMouseX)
			dy := float64(my - editorLastMouseY)

			world.Transforms[MainCameraEntity].Position.X -= dx
			world.Transforms[MainCameraEntity].Position.Y -= dy

			editorLastMouseX, editorLastMouseY = mx, my
		}
	} else {
		editorIsPanning = false
	}

	panSpeed := 500.0 * dt
	if ebiten.IsKeyPressed(ebiten.KeyNumpad8) {
		world.Transforms[MainCameraEntity].Position.Y -= panSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyNumpad2) {
		world.Transforms[MainCameraEntity].Position.Y += panSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyNumpad4) {
		world.Transforms[MainCameraEntity].Position.X -= panSpeed
	}
	if ebiten.IsKeyPressed(ebiten.KeyNumpad6) {
		world.Transforms[MainCameraEntity].Position.X += panSpeed
	}
	// ---------------------------

	worldMouse := ScreenToWorld(camera, MainCameraEntity, world, float64(mx), float64(my))

	if inpututil.IsKeyJustPressed(ebiten.KeyP) {
		Editor.Mode = EditorModePlatform
		Editor.StatusMessage = "Mode: Platform"
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyK) {
		Editor.Mode = EditorModeTrigger
		Editor.StatusMessage = "Mode: Kill Trigger"
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyC) {
		Editor.Mode = EditorModeCheckpoint
		Editor.StatusMessage = "Mode: Checkpoint"
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		handleEditorClick(world, worldMouse)
	}

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
		case EditorSelectionCheckpoint:
			if _, ok := world.Checkpoints[Editor.SelectedEntity]; !ok {
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
				handleLayerInput(world)
			}
			handleVelocityInput(world)
			handleTimerInput(world)
		case EditorSelectionTrigger:
			handleTriggerResizeInput(world)
		case EditorSelectionCheckpoint:
			handleTriggerResizeInput(world)
			handleCheckpointLimitInput(world)
		}
		handleDeleteInput(world)
	}

	handleSaveLoadInput(world, ctrlHeld)
}

func moveEditorRoom(world *Engine.World, dx, dy int) {
	Editor.ViewedRoom.X += float64(dx)
	Editor.ViewedRoom.Y += float64(dy)
	Editor.HasSelection = false
	Editor.StatusMessage = fmt.Sprintf("Viewing Room %.0f, %.0f", Editor.ViewedRoom.X, Editor.ViewedRoom.Y)
}

func handleEditorClick(world *Engine.World, worldMouse Engine.Vector2) {
	for _, cp := range world.Checkpoints {
		if pointInAABB(worldMouse.X, worldMouse.Y, world.Triggers[cp.Entity].AABB) {
			Editor.SelectedEntity = cp.Entity
			Editor.HasSelection = true
			Editor.SelectionType = EditorSelectionCheckpoint
			Editor.StatusMessage = "Selected checkpoint"
			return
		}
	}

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

	snappedX := SnapToGrid(worldMouse.X)
	snappedY := SnapToGrid(worldMouse.Y)

	switch Editor.Mode {
	case EditorModePlatform:
		newEntity := world.CreateNewLevelColliderInScene(&world.Scene, snappedX, snappedY, EditorPlatformW, EditorPlatformH, Engine.DEFAULT)
		Editor.SelectedEntity = newEntity
		Editor.HasSelection = true
		Editor.SelectionType = EditorSelectionPlatform
		Editor.StatusMessage = "Placed new platform"

		world.Platforms[newEntity].StartX = snappedX
		world.Platforms[newEntity].StartY = snappedY
		world.Platforms[newEntity].StartVelX = 0.0
		world.Platforms[newEntity].StartVelY = 0.0

	case EditorModeTrigger:
		newEntity := world.CreateNewTriggerColliderInScene(&world.Scene, snappedX, snappedY, EditorTriggerW, EditorTriggerH, Engine.DEFAULT, nil)
		SetupKillTriggerCallback(world.Triggers[newEntity])
		Editor.SelectedEntity = newEntity
		Editor.HasSelection = true
		Editor.SelectionType = EditorSelectionTrigger
		Editor.StatusMessage = "Placed kill trigger"

	case EditorModeCheckpoint:
		newEntity := world.CreateNewTriggerColliderInScene(&world.Scene, snappedX, snappedY, EditorTriggerW, EditorTriggerH, Engine.DEFAULT, nil)
		SetupCheckpointCallback(world.Triggers[newEntity], newEntity)

		world.Checkpoints[newEntity] = &Engine.Checkpoint{
			Entity:            newEntity,
			Position:          Engine.Vector2{X: snappedX, Y: snappedY},
			Activated:         false,
			InputReleaseLimit: 10,
		}
		Editor.SelectedEntity = newEntity
		Editor.HasSelection = true
		Editor.SelectionType = EditorSelectionCheckpoint
		Editor.StatusMessage = "Placed checkpoint"
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
		platform.StartVelX = platform.Velocity.X
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) {
		platform.Velocity.X -= EditorVelocityStep
		platform.StartVelX = platform.Velocity.X
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
		platform.Velocity.Y += EditorVelocityStep
		platform.StartVelY = platform.Velocity.Y
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
		platform.Velocity.Y -= EditorVelocityStep
		platform.StartVelY = platform.Velocity.Y
	}
}

func handleLayerInput(world *Engine.World) {
	if inpututil.IsKeyJustPressed(ebiten.KeyO) {
		platform := world.Platforms[Editor.SelectedEntity]

		if platform.AABB.Layer == Engine.DEFAULT {
			platform.AABB.Layer = Engine.ONE_WAY_PLATFORMS
			Editor.StatusMessage = "Platform type: One-Way"
		} else {
			platform.AABB.Layer = Engine.DEFAULT
			Editor.StatusMessage = "Platform type: Solid"
		}
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
		case EditorSelectionCheckpoint:
			delete(world.Checkpoints, Editor.SelectedEntity)
			world.RemoveTrigger(Editor.SelectedEntity)
			Editor.StatusMessage = "Deleted checkpoint"
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
		if err := LoadLevel(LevelSaveFile, world, true); err != nil {
			Editor.StatusMessage = "Load failed: " + err.Error()
		} else {
			Editor.HasSelection = false
			Editor.SelectionType = EditorSelectionNone
			Editor.StatusMessage = "Loaded from " + LevelSaveFile
		}
	}
}

func handleCheckpointLimitInput(world *Engine.World) {
	if Editor.SelectionType != EditorSelectionCheckpoint {
		return
	}

	cp := world.Checkpoints[Editor.SelectedEntity]
	if cp == nil {
		return
	}

	// Use Numpad + and - to adjust the limit
	if inpututil.IsKeyJustPressed(ebiten.KeyNumpadAdd) {
		cp.InputReleaseLimit++
		Editor.StatusMessage = fmt.Sprintf("Checkpoint Releases: %d", cp.InputReleaseLimit)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyNumpadSubtract) {
		if cp.InputReleaseLimit > 0 {
			cp.InputReleaseLimit--
		}
		Editor.StatusMessage = fmt.Sprintf("Checkpoint Releases: %d", cp.InputReleaseLimit)
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

type SerializableCheckpoint struct {
	X, Y          float64
	Width, Height float64
	ReleaseLimit  int `json:"ReleaseLimit"`
}

type SerializableLevel struct {
	Platforms   []SerializablePlatform   `json:"platforms"`
	Triggers    []SerializableTrigger    `json:"triggers"`
	Checkpoints []SerializableCheckpoint `json:"checkpoints"`
}

func SaveLevel(path string, world *Engine.World) error {
	platforms := make([]SerializablePlatform, 0, len(world.Platforms))
	for _, platform := range world.Platforms {
		platforms = append(platforms, SerializablePlatform{
			X:                    platform.StartX,
			Y:                    platform.StartY,
			Width:                platform.AABB.Width,
			Height:               platform.AABB.Height,
			Layer:                platform.AABB.Layer,
			VelX:                 platform.StartVelX,
			VelY:                 platform.StartVelY,
			VelTimerDuration:     platform.VelTimerDuration,
			VelTimerStopAtEnds:   platform.VelTimerStopAtEnds,
			VelTimerStopDuration: platform.VelTimerStopDuration,
			VelTimerAxis:         platform.VelTimerAxis,
		})
	}

	triggers := make([]SerializableTrigger, 0, len(world.Triggers))
	for _, trigger := range world.Triggers {
		if trigger.TriggerType == "checkpoint" {
			continue
		}
		triggers = append(triggers, SerializableTrigger{
			X: trigger.AABB.X, Y: trigger.AABB.Y, Width: trigger.AABB.Width, Height: trigger.AABB.Height,
			Layer: trigger.AABB.Layer, TriggerType: trigger.TriggerType,
		})
	}

	checkpoints := make([]SerializableCheckpoint, 0, len(world.Checkpoints))
	for _, cp := range world.Checkpoints {
		trigger := world.Triggers[cp.Entity]
		checkpoints = append(checkpoints, SerializableCheckpoint{
			X:            cp.Position.X,
			Y:            cp.Position.Y,
			Width:        trigger.AABB.Width,
			Height:       trigger.AABB.Height,
			ReleaseLimit: cp.InputReleaseLimit,
		})
	}

	level := SerializableLevel{Platforms: platforms, Triggers: triggers, Checkpoints: checkpoints}
	data, err := json.MarshalIndent(level, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func LoadLevel(path string, world *Engine.World, clearLoad bool) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var level SerializableLevel
	if err := json.Unmarshal(data, &level); err != nil {
		return err
	}

	if clearLoad {
		for e := range world.Platforms {
			world.RemovePlatform(e)
		}
		for e := range world.Triggers {
			world.RemoveTrigger(e)
		}
	}

	world.Checkpoints = make(map[Engine.Entity]*Engine.Checkpoint)

	for _, p := range level.Platforms {
		e := world.CreateNewLevelColliderInScene(&world.Scene, p.X, p.Y, p.Width, p.Height, p.Layer)
		platform := world.Platforms[e]

		platform.Velocity = Engine.Vector2{X: p.VelX, Y: p.VelY}

		platform.StartX = p.X
		platform.StartY = p.Y
		platform.StartVelX = p.VelX
		platform.StartVelY = p.VelY

		platform.VelTimerDuration = p.VelTimerDuration
		platform.VelTimerStopAtEnds = p.VelTimerStopAtEnds
		platform.VelTimerStopDuration = p.VelTimerStopDuration
		platform.VelTimerAxis = p.VelTimerAxis
		RegisterPlatformTimer(platform)
	}

	for _, t := range level.Triggers {
		e := world.CreateNewTriggerColliderInScene(&world.Scene, t.X, t.Y, t.Width, t.Height, t.Layer, nil)
		trigger := world.Triggers[e]
		trigger.TriggerType = t.TriggerType
		switch t.TriggerType {
		case "kill":
			SetupKillTriggerCallback(trigger)
		case "DashPowerUpTrigger":
			SetUpDashPowerUpTriggerCallback(trigger)
		case "WallClimbPowerUpTrigger":
			SetUpWallClimbPowerUpTriggerCallback(trigger)
		}
	}

	for _, scp := range level.Checkpoints {
		e := world.CreateNewTriggerColliderInScene(&world.Scene, scp.X, scp.Y, scp.Width, scp.Height, Engine.DEFAULT, nil)
		SetupCheckpointCallback(world.Triggers[e], e)

		world.Checkpoints[e] = &Engine.Checkpoint{
			Entity:            e,
			Position:          Engine.Vector2{X: scp.X, Y: scp.Y},
			Activated:         false,
			InputReleaseLimit: scp.ReleaseLimit,
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
	drawCheckpointsInEditor(camera, cameraMatrix, world)

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
		case EditorSelectionCheckpoint:
			if trigger, ok := world.Triggers[Editor.SelectedEntity]; ok {
				drawSelectionHighlight(camera, cameraMatrix, trigger.AABB, color.RGBA{R: 255, G: 255, B: 100, A: 255})
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

	// 1. Draw the minor 20x20 placement grid
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

	// 2. Draw the 640x480 "Fake Room" boundaries
	// A subtle blue/purple color so it stands out from the dark gray grid
	roomColor := color.RGBA{R: 60, G: 60, B: 120, A: 120}
	roomWidth := 640.0
	roomHeight := 480.0

	// Find the first room boundary line that is visible on screen
	startRoomX := math.Floor(minX/roomWidth) * roomWidth
	startRoomY := math.Floor(minY/roomHeight) * roomHeight

	// Draw vertical room boundaries (every 640 pixels)
	for x := startRoomX; x <= maxX; x += roomWidth {
		sx0, sy0 := cameraMatrix.Apply(x, minY)
		sx1, sy1 := cameraMatrix.Apply(x, maxY)
		// Width is 2 to make it slightly thicker than the placement grid
		vector.StrokeLine(camera.RenderTexture, float32(sx0), float32(sy0), float32(sx1), float32(sy1), 2, roomColor, false)
	}

	// Draw horizontal room boundaries (every 480 pixels)
	for y := startRoomY; y <= maxY; y += roomHeight {
		sx0, sy0 := cameraMatrix.Apply(minX, y)
		sx1, sy1 := cameraMatrix.Apply(maxX, y)
		vector.StrokeLine(camera.RenderTexture, float32(sx0), float32(sy0), float32(sx1), float32(sy1), 2, roomColor, false)
	}
}

func drawPlacementPreview(camera *Engine.Camera, cameraEntity Engine.Entity, world *Engine.World, cameraMatrix *ebiten.GeoM) {
	mx, my := ebiten.CursorPosition()
	worldMouse := ScreenToWorld(camera, cameraEntity, world, float64(mx), float64(my))

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
		previewW, previewH = EditorPlatformW, EditorPlatformH
		previewColor = color.RGBA{R: 255, G: 255, B: 255, A: 150}
	case EditorModeTrigger:
		previewW, previewH = EditorTriggerW, EditorTriggerH
		previewColor = color.RGBA{R: 255, G: 80, B: 80, A: 150}
	case EditorModeCheckpoint: // NEW
		previewW, previewH = EditorTriggerW, EditorTriggerH
		previewColor = color.RGBA{R: 255, G: 255, B: 0, A: 150}
	}

	vector.StrokeRect(camera.RenderTexture, float32(sx), float32(sy), float32(previewW), float32(previewH), 1, previewColor, false)
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

func drawCheckpointsInEditor(camera *Engine.Camera, cameraMatrix *ebiten.GeoM, world *Engine.World) {
	for _, cp := range world.Checkpoints {
		trigger := world.Triggers[cp.Entity]
		aabb := trigger.AABB
		x, y := cameraMatrix.Apply(aabb.X, aabb.Y)

		vector.FillRect(camera.RenderTexture, float32(x), float32(y), float32(aabb.Width), float32(aabb.Height), color.RGBA{R: 255, G: 255, B: 0, A: 40}, false)
		vector.StrokeRect(camera.RenderTexture, float32(x), float32(y), float32(aabb.Width), float32(aabb.Height), 1, color.RGBA{R: 255, G: 255, B: 0, A: 100}, false)

		poleColor := color.RGBA{R: 180, G: 140, B: 0, A: 255}
		flagFillColor := color.RGBA{R: 255, G: 255, B: 0, A: 220}
		flagOutlineColor := color.RGBA{R: 255, G: 255, B: 100, A: 255}
		foldColor := color.RGBA{R: 200, G: 200, B: 0, A: 180}

		poleX := float32(x + 8)
		poleTop := float32(y + 2)
		poleBottom := float32(y + aabb.Height - 2)

		vector.StrokeLine(camera.RenderTexture, poleX, poleTop, poleX, poleBottom, 2, poleColor, false)

		flagX := poleX
		flagY := poleTop
		flagW := float32(20)
		flagH := float32(10)

		vector.FillRect(camera.RenderTexture, flagX, flagY, flagW, flagH, flagFillColor, false)
		vector.StrokeRect(camera.RenderTexture, flagX, flagY, flagW, flagH, 2, flagOutlineColor, false)

		vector.StrokeLine(camera.RenderTexture, flagX+flagW/2, flagY, flagX+flagW/2, flagY+flagH, 1, foldColor, false)
		vector.StrokeLine(camera.RenderTexture, flagX, flagY+flagH/2, flagX+flagW, flagY+flagH, 1, foldColor, false)
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
	if Editor.Mode == EditorModeCheckpoint {
		modeStr = "CHECKPOINT"
	}

	lines := fmt.Sprintf("EDITOR MODE (Tab to exit)\n"+
		"Current: %s [P/K/C to switch]\n"+
		"Click: place/select\n"+
		"W/S height, A/D width\n"+
		"Right-Click Drag: Pan camera\n"+
		"Numpad Arrows: Pan camera\n",
		modeStr)

	if Editor.Mode == EditorModePlatform {
		lines += "O: toggle one-way/solid\n" +
			"Arrows: nudge velocity\n" +
			"[ / ]: timer duration -/+\n" +
			"Shift+[ / Shift+]: stop duration -/+\n" +
			"T: toggle stop-at-ends\n" +
			"Y: cycle timer axis (X/Y/Both)\n"
	}

	if Editor.Mode == EditorModeCheckpoint {
		lines += "Numpad +/-: set release limit\n"
	}

	lines += "Delete: remove selected\n" +
		"Ctrl+S save, Ctrl+L load"

	if Editor.HasSelection {
		switch Editor.SelectionType {
		case EditorSelectionPlatform:
			if platform, ok := world.Platforms[Editor.SelectedEntity]; ok {
				aabb := platform.AABB
				layerStr := "Solid"
				if aabb.Layer == Engine.ONE_WAY_PLATFORMS {
					layerStr = "One-Way"
				}

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
					"\n\nPlatform: pos(%.0f, %.0f) size(%.0f x %.0f)\nLayer: %s\nvel(%.0f, %.0f)  timer:%s",
					aabb.X, aabb.Y, aabb.Width, aabb.Height,
					layerStr,
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
		case EditorSelectionCheckpoint:
			if cp, ok := world.Checkpoints[Editor.SelectedEntity]; ok {
				lines += fmt.Sprintf(
					"\n\nCheckpoint: pos(%.0f, %.0f)\nReleases Granted: %d\nState: %v",
					cp.Position.X, cp.Position.Y, cp.InputReleaseLimit, cp.Activated,
				)
			}
		}
	}

	if Editor.StatusMessage != "" {
		lines += "\n" + Editor.StatusMessage
	}

	ebitenutil.DebugPrintAt(camera.RenderTexture, lines, 4, 4)
}
