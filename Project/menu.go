package Project

import (
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
)

var CurrentGameState GameState = StateStartMenu

var menuFaceSource *text.GoTextFaceSource

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
