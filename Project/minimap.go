package Project

import (
	"Disconnect/Engine"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

func DrawMinimap(screen *ebiten.Image, world *Engine.World) {
	if !world.Alive[player.Entity] {
		return
	}

	isIdle := !player.LeftHeld && !player.RightHeld && !player.UpHeld && !player.DownHeld
	cc := player.CC()

	if !isIdle || !cc.Grounded {
		return
	}

	roomStatus := make(map[Engine.Vector2]int)

	for _, cp := range world.Checkpoints {
		rx := math.Floor(cp.Position.X / 640.0)
		ry := math.Floor(cp.Position.Y / 480.0)
		roomKey := Engine.Vector2{X: rx, Y: ry}

		if cp.Activated {
			if roomStatus[roomKey] != 1 {
				roomStatus[roomKey] = 2
			}
		} else {
			roomStatus[roomKey] = 1
		}
	}

	if len(roomStatus) == 0 {
		return
	}

	minX, minY := math.Inf(1), math.Inf(1)
	maxX, maxY := math.Inf(-1), math.Inf(-1)
	for key := range roomStatus {
		if key.X < minX {
			minX = key.X
		}
		if key.Y < minY {
			minY = key.Y
		}
		if key.X > maxX {
			maxX = key.X
		}
		if key.Y > maxY {
			maxY = key.Y
		}
	}

	roomPixelSize := 12.0
	padding := 20.0

	bgColor := color.RGBA{R: 0, G: 0, B: 0, A: 180}               // Darker, more opaque background
	emptyColor := color.RGBA{R: 30, G: 30, B: 30, A: 200}         // Empty rooms
	redColor := color.RGBA{R: 200, G: 50, B: 50, A: 230}          // Missing checkpoints
	greyColor := color.RGBA{R: 150, G: 150, B: 150, A: 230}       // All collected
	borderColor := color.RGBA{R: 80, G: 80, B: 80, A: 220}        // Grid lines
	playerRoomColor := color.RGBA{R: 255, G: 255, B: 255, A: 255} // Current room

	mapWidth := (maxX - minX + 1) * roomPixelSize
	mapHeight := (maxY - minY + 1) * roomPixelSize

	vector.FillRect(screen,
		float32(padding-6), float32(padding-6),
		float32(mapWidth+12), float32(mapHeight+12),
		bgColor, false)

	for rx := minX; rx <= maxX; rx++ {
		for ry := minY; ry <= maxY; ry++ {
			key := Engine.Vector2{X: rx, Y: ry}

			drawX := padding + (rx-minX)*roomPixelSize
			drawY := padding + (ry-minY)*roomPixelSize

			var fillColor color.RGBA
			if status, exists := roomStatus[key]; exists {
				if status == 2 {
					fillColor = greyColor
				} else {
					fillColor = redColor
				}
			} else {
				fillColor = emptyColor
			}

			vector.FillRect(screen, float32(drawX), float32(drawY), float32(roomPixelSize), float32(roomPixelSize), fillColor, false)
			vector.StrokeRect(screen, float32(drawX), float32(drawY), float32(roomPixelSize), float32(roomPixelSize), 1, borderColor, false)
		}
	}

	playerRoomX := math.Floor(world.Transforms[player.Entity].Position.X / 640.0)
	playerRoomY := math.Floor(world.Transforms[player.Entity].Position.Y / 480.0)

	if playerRoomX >= minX && playerRoomX <= maxX && playerRoomY >= minY && playerRoomY <= maxY {
		indX := padding + (playerRoomX-minX)*roomPixelSize
		indY := padding + (playerRoomY-minY)*roomPixelSize

		vector.StrokeRect(screen,
			float32(indX)-2, float32(indY)-2,
			float32(roomPixelSize)+4, float32(roomPixelSize)+4,
			2, playerRoomColor, false)
	}
}
