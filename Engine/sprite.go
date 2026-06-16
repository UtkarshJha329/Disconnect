package Engine

import (
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Sprite struct {
	Tex *ebiten.Image
}

func (spr *Sprite) SetImageFromFile(src string) {
	var err error
	spr.Tex, _, err = ebitenutil.NewImageFromFile(src)
	if err != nil {
		log.Fatal(err)
	}
}
