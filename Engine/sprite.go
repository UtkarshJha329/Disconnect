package Engine

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Sprite struct {
	Tex           *ebiten.Image
	FramePosition Vector2
	FrameSize     Vector2
}

func GetImageFromFile(src string) *ebiten.Image {
	img, _, err := ebitenutil.NewImageFromFile(src)
	if err != nil {
		log.Fatal(err)
	}
	return img
}

func (spr *Sprite) SetImageFromFile(src string) {
	var err error
	spr.Tex, _, err = ebitenutil.NewImageFromFile(src)
	if err != nil {
		log.Fatal(err)
	}
}
