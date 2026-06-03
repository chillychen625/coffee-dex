package game

import (
	"bytes"
	"embed"
	"fmt"
	"image"
	_ "image/png"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed assets/sprites/*.png
var spriteFS embed.FS

var spriteCache = map[int]*ebiten.Image{}

// GetSprite returns the cached sprite image for the given Pokemon ID (1–151).
// Returns nil if the sprite file is missing or unreadable.
func GetSprite(id int) *ebiten.Image {
	if img, ok := spriteCache[id]; ok {
		return img
	}
	data, err := spriteFS.ReadFile(fmt.Sprintf("assets/sprites/%03d.png", id))
	if err != nil {
		spriteCache[id] = nil
		return nil
	}
	raw, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		spriteCache[id] = nil
		return nil
	}
	eimg := ebiten.NewImageFromImage(raw)
	spriteCache[id] = eimg
	return eimg
}

// DrawSprite renders the Pokemon sprite for the given ID, scaled to (w×h) at (x, y).
func DrawSprite(screen *ebiten.Image, id, x, y, w, h int) {
	img := GetSprite(id)
	if img == nil {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(w)/float64(img.Bounds().Dx()), float64(h)/float64(img.Bounds().Dy()))
	op.GeoM.Translate(float64(x), float64(y))
	screen.DrawImage(img, op)
}

// DrawSpriteAlpha renders a sprite with a custom alpha (0.0–1.0).
func DrawSpriteAlpha(screen *ebiten.Image, id, x, y, w, h int, alpha float32) {
	img := GetSprite(id)
	if img == nil {
		return
	}
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(w)/float64(img.Bounds().Dx()), float64(h)/float64(img.Bounds().Dy()))
	op.GeoM.Translate(float64(x), float64(y))
	op.ColorScale.ScaleAlpha(alpha)
	screen.DrawImage(img, op)
}
