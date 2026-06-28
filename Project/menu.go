package Project

import (
	"Disconnect/Engine"
	"errors"
	"fmt"
	"image/color"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

type GameState int

const (
	StateStartMenu GameState = iota
	StatePlaying
	StateGameComplete
)

var CurrentGameState GameState = StateStartMenu

var menuFaceSource *text.GoTextFaceSource

var GameStartTime time.Time
var GameCompletionTime time.Duration

var ErrRestartGame = errors.New("restart game")

var titleColor = color.RGBA{R: 255, G: 255, B: 255, A: 255}
var promptColor = color.RGBA{R: 0, G: 200, B: 255, A: 255}

func MenuInit(faceSource *text.GoTextFaceSource) {
	menuFaceSource = faceSource
}

func MenuUpdate() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
		inpututil.IsKeyJustPressed(ebiten.KeySpace) ||
		inpututil.IsKeyJustPressed(ebiten.KeyZ) {
		CurrentGameState = StatePlaying
		GameStartTime = time.Now()
		PlaySoundEffect("Assets/Sfx/Menuselect.wav")
	}
	return nil
}

func CheckGameComplete(world *Engine.World) bool {
	if len(world.Checkpoints) == 0 {
		return false
	}
	for _, cp := range world.Checkpoints {
		if !cp.Activated {
			return false
		}
	}
	return true
}

func VictoryUpdate() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
		inpututil.IsKeyJustPressed(ebiten.KeySpace) ||
		inpututil.IsKeyJustPressed(ebiten.KeyZ) {
		return ErrRestartGame
	}
	return nil
}

func MenuDraw(screen *ebiten.Image) {
	if menuFaceSource == nil {
		return
	}

	titleFace := &text.GoTextFace{
		Source: menuFaceSource,
		Size:   18,
	}
	titleText := "DISCONNECT"
	titleAdv, _ := text.Measure(titleText, titleFace, 0)
	titleX := (640 - float64(titleAdv)) / 2.0
	titleY := 180.0

	bodyFace := &text.GoTextFace{
		Source: menuFaceSource,
		Size:   8,
	}

	titleTextDrawOption := text.DrawOptions{}
	titleTextDrawOption.GeoM.Translate(titleX, titleY)
	titleTextDrawOption.ColorScale.ScaleWithColor(titleColor)
	text.Draw(screen, titleText, titleFace, &titleTextDrawOption)

	blinkTime := time.Now().UnixMilli() % 1000
	currentPromptColor := promptColor
	if blinkTime < 500 {
		currentPromptColor.A = 0
	}

	promptText := "PRESS ENTER TO START"
	promptAdv, _ := text.Measure(promptText, bodyFace, 0)
	promptX := (640 - float64(promptAdv)) / 2.0
	promptY := 280.0

	promptTextDrawOption := text.DrawOptions{}
	promptTextDrawOption.GeoM.Translate(promptX, promptY)
	promptTextDrawOption.ColorScale.ScaleWithColor(currentPromptColor)
	text.Draw(screen, promptText, bodyFace, &promptTextDrawOption)

	hintText := "Arrow Keys: Move | C: Jump | X: Dash | Tab: Editor"
	hintAdv, _ := text.Measure(hintText, bodyFace, 0)
	hintX := (640 - float64(hintAdv)) / 2.0
	hintY := 420.0

	hintColor := color.RGBA{R: 150, G: 150, B: 150, A: 255}
	hintTextDrawOption := text.DrawOptions{}
	hintTextDrawOption.GeoM.Translate(hintX, hintY)
	hintTextDrawOption.ColorScale.ScaleWithColor(hintColor)
	text.Draw(screen, hintText, bodyFace, &hintTextDrawOption)
}

func VictoryDraw(screen *ebiten.Image) {
	if menuFaceSource == nil {
		return
	}

	titleFace := &text.GoTextFace{
		Source: menuFaceSource,
		Size:   14,
	}
	bodyFace := &text.GoTextFace{
		Source: menuFaceSource,
		Size:   10,
	}
	smallFace := &text.GoTextFace{
		Source: menuFaceSource,
		Size:   8,
	}

	titleText := "YOU PREVENTED THE PC FROM DISCONNECTING!!"
	titleAdv, _ := text.Measure(titleText, titleFace, 0)
	titleX := (640 - float64(titleAdv)) / 2.0
	titleY := 150.0

	op := text.DrawOptions{}
	op.GeoM.Translate(titleX, titleY)
	op.ColorScale.ScaleWithColor(color.RGBA{R: 100, G: 255, B: 100, A: 255}) // Green text
	text.Draw(screen, titleText, titleFace, &op)

	minutes := int(GameCompletionTime.Minutes())
	seconds := int(GameCompletionTime.Seconds()) % 60
	milliseconds := int(GameCompletionTime.Milliseconds()) % 1000
	timeStr := fmt.Sprintf("TIME: %02d:%02d.%03d", minutes, seconds, milliseconds)

	timeAdv, _ := text.Measure(timeStr, bodyFace, 0)
	timeX := (640 - float64(timeAdv)) / 2.0
	timeY := 220.0

	op2 := text.DrawOptions{}
	op2.GeoM.Translate(timeX, timeY)
	op2.ColorScale.ScaleWithColor(color.RGBA{R: 255, G: 255, B: 255, A: 255})
	text.Draw(screen, timeStr, bodyFace, &op2)

	blinkTime := time.Now().UnixMilli() % 1000
	currentPromptColor := promptColor
	if blinkTime < 500 {
		currentPromptColor.A = 0
	}

	promptText := "PRESS ENTER TO RESTART"
	promptAdv, _ := text.Measure(promptText, smallFace, 0)
	promptX := (640 - float64(promptAdv)) / 2.0
	promptY := 300.0

	op3 := text.DrawOptions{}
	op3.GeoM.Translate(promptX, promptY)
	op3.ColorScale.ScaleWithColor(currentPromptColor)
	text.Draw(screen, promptText, smallFace, &op3)
}
