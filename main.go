package main

import (
	"Disconnect/Engine"
	"Disconnect/Project"
	"bytes"
	"cmp"
	"errors"
	"image/color"
	"log"
	"slices"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Game struct{}

var world *Engine.World = Engine.NewWorld()

var neonBlueFace text.Face
var neonBlueColor = color.RGBA{R: 0, G: 200, B: 255, A: 255}

func (g *Game) Update() error {
	dt := 1.0 / float64(ebiten.TPS())

	if Project.CurrentGameState == Project.StateStartMenu {
		return Project.MenuUpdate()
	}

	if Project.CurrentGameState == Project.StateGameComplete {
		err := Project.VictoryUpdate()
		if err != nil {
			if errors.Is(err, Project.ErrRestartGame) {
				Project.PlaySoundEffect("Assets/Sfx/Menuselect.wav")
				initGameWorld()
				Project.CurrentGameState = Project.StateStartMenu
			}
		}
		return nil
	}

	// if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
	// 	Project.Editor.Active = !Project.Editor.Active
	// }

	if Project.Editor.Active {
		Project.EditorUpdateFunc(world, dt)
	} else {
		for _, updateFunc := range world.EntityUpdateFuncs {
			updateFunc(world, dt)
		}

		if Project.CheckGameComplete(world) {
			Project.GameCompletionTime = time.Since(Project.GameStartTime)
			Project.CurrentGameState = Project.StateGameComplete
			return nil
		}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {

	if Project.CurrentGameState == Project.StateStartMenu {
		Project.MenuDraw(screen)
		return
	}

	if Project.CurrentGameState == Project.StateGameComplete {
		Project.VictoryDraw(screen)
		return
	}

	camera := world.Cameras[Project.MainCameraEntity]
	camera.RenderTexture.Clear()

	Engine.DrawParticleSystems(camera, Project.MainCameraEntity, world)

	world.Scene.RenderSceneHierarchy(Project.MainCameraEntity, world)

	cameraMatrix := camera.GetCameraTransformMatrix(Project.MainCameraEntity, world)

	for _, platform := range world.Platforms {

		collider := platform.AABB
		x, y := cameraMatrix.Apply(collider.X, collider.Y)

		vector.StrokeRect(
			camera.RenderTexture,
			float32(x),
			float32(y),
			float32(collider.Width),
			float32(collider.Height),
			1,
			color.RGBA{0, 255, 0, 255},
			false,
		)
	}

	for _, trigger := range world.Triggers {

		collider := trigger.AABB
		x, y := cameraMatrix.Apply(collider.X, collider.Y)

		triggerColor := color.RGBA{255, 0, 0, 255}
		if trigger.TriggerType == "DashPowerUpTrigger" || trigger.TriggerType == "WallClimbPowerUpTrigger" {
			triggerColor = color.RGBA{255, 255, 0, 255}

		}

		vector.StrokeRect(
			camera.RenderTexture,
			float32(x),
			float32(y),
			float32(collider.Width),
			float32(collider.Height),
			1,
			triggerColor,
			false,
		)
	}

	for _, cp := range world.Checkpoints {

		trigger := world.Triggers[cp.Entity]
		aabb := trigger.AABB
		x, y := cameraMatrix.Apply(aabb.X, aabb.Y)

		// Determine colors based on activation state
		poleColor := color.RGBA{R: 180, G: 140, B: 0, A: 255}
		foldColor := color.RGBA{R: 200, G: 200, B: 0, A: 180}

		var flagFillColor color.RGBA
		var flagOutlineColor color.RGBA
		if cp.Activated {
			flagFillColor = color.RGBA{R: 255, G: 255, B: 0, A: 255} // Bright solid yellow
			flagOutlineColor = color.RGBA{R: 255, G: 255, B: 150, A: 255}
		} else {
			flagFillColor = color.RGBA{R: 100, G: 100, B: 0, A: 150} // Dimmer/more transparent yellow
			flagOutlineColor = color.RGBA{R: 100, G: 100, B: 0, A: 200}
		}

		// --- Draw Fancy Rectangle Flag ---
		poleX := float32(x + 8)
		poleTop := float32(y + 2)
		poleBottom := float32(y + aabb.Height - 2) // Reaches the ground

		// 1. Pole
		vector.StrokeLine(camera.RenderTexture, poleX, poleTop, poleX, poleBottom, 2, poleColor, false)

		// 2. Flag body (Rectangle at the top of the pole)
		flagX := poleX
		flagY := poleTop
		flagW := float32(20)
		flagH := float32(10)

		vector.FillRect(camera.RenderTexture, flagX, flagY, flagW, flagH, flagFillColor, false)
		vector.StrokeRect(camera.RenderTexture, flagX, flagY, flagW, flagH, 2, flagOutlineColor, false)

		// 3. Fancy details (Fabric fold lines)
		vector.StrokeLine(camera.RenderTexture, flagX+flagW/2, flagY, flagX+flagW/2, flagY+flagH, 1, foldColor, false)
		vector.StrokeLine(camera.RenderTexture, flagX, flagY+flagH/2, flagX+flagW, flagY+flagH, 1, foldColor, false)
	}

	for e, cc := range world.CharacterControllers {

		x, y := cameraMatrix.Apply(
			world.Transforms[e].Position.X-cc.Width/2.0,
			world.Transforms[e].Position.Y-cc.Height/2.0,
		)

		vector.StrokeRect(
			camera.RenderTexture,
			float32(x),
			float32(y),
			float32(cc.Width),
			float32(cc.Height),
			1,
			color.RGBA{0, 0, 255, 255},
			false,
		)

		releaseText := Project.GetPlayerReleaseText()

		textWorldX := world.Transforms[e].Position.X
		textWorldY := world.Transforms[e].Position.Y - (cc.Height / 2.0) - 20.0

		textScreenX, textScreenY := cameraMatrix.Apply(textWorldX, textWorldY)

		adv, _ := text.Measure(releaseText, neonBlueFace, 0)
		drawX := textScreenX - float64(adv)/2.0
		drawY := textScreenY

		op := &text.DrawOptions{}
		op.GeoM.Translate(drawX, drawY)
		op.ColorScale.ScaleWithColor(neonBlueColor)
		text.Draw(camera.RenderTexture, releaseText, neonBlueFace, op)
	}

	if Project.Editor.Active {
		Project.EditorDrawFunc(camera, Project.MainCameraEntity, world)
	}

	cameras := make([]*Project.MainCameraEntityTransformPair, 0, len(world.Cameras))
	for e, camera := range world.Cameras {
		curEle := &Project.MainCameraEntityTransformPair{
			Entity:    e,
			Transform: world.Transforms[e],
			Camera:    camera,
		}
		cameras = append(cameras, curEle)
	}
	slices.SortFunc(cameras, func(a, b *Project.MainCameraEntityTransformPair) int {
		return cmp.Compare(a.Transform.Position.Z, b.Transform.Position.Z)
	})

	op := ebiten.DrawImageOptions{}
	for _, camera := range cameras {
		screen.DrawImage(camera.Camera.RenderTexture, &op)
	}

	Project.DrawMinimap(screen, world)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 640, 480
}

func initGameWorld() {
	world = Engine.NewWorld()

	world.EntityInitfuncs = append(world.EntityInitfuncs,
		Project.PlatformsInitFunc,
		Project.PlayerInitFunc,
		Project.MainCameraInitFunc,
		Project.TriggersInitFunc,
		Project.SFXInitFunc,
	)

	world.EntityUpdateFuncs = append(world.EntityUpdateFuncs,
		Project.PlatformsUpdateFunc,
		Project.PlayerUpdateFunc,
		Project.MainCameraUpdateFunc,
		Project.TriggersUpdateFunc,
		Engine.ParticlesUpdateFunc,
		Engine.AnimationEntityUpdateFunc,
	)

	for _, initFunc := range world.EntityInitfuncs {
		initFunc(world)
	}

	if err := Project.LoadLevel(Project.LevelSaveFile, world, false); err != nil {
		log.Panic("Load failed: " + err.Error())
	}
}

func main() {
	game := &Game{}
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Sora Engine")

	Project.AudioContext = audio.NewContext(Project.SampleRate)

	fontData, err := Project.Assets.ReadFile("Assets/Fonts/PressStart2P-Regular.ttf")
	if err != nil {
		log.Fatal(err)
	}

	source, err := text.NewGoTextFaceSource(bytes.NewReader(fontData))
	if err != nil {
		log.Fatal(err)
	}

	neonBlueFace = &text.GoTextFace{
		Source: source,
		Size:   8,
	}

	Project.MenuInit(source)

	initGameWorld()

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
