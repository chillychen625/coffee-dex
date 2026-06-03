package game

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const maxInputLen = 80

type TextInput struct {
	Value  string
	cursor int
	blink  int
}

func (t *TextInput) Update(focused bool) {
	if !focused {
		return
	}
	t.blink++

	var chars []rune
	chars = ebiten.AppendInputChars(chars)
	for _, ch := range chars {
		if ch >= 32 {
			runes := []rune(t.Value)
			if len(runes) >= maxInputLen {
				continue
			}
			runes = append(runes[:t.cursor:t.cursor], append([]rune{ch}, runes[t.cursor:]...)...)
			t.Value = string(runes)
			t.cursor++
			t.blink = 0
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) && t.cursor > 0 {
		runes := []rune(t.Value)
		runes = append(runes[:t.cursor-1], runes[t.cursor:]...)
		t.Value = string(runes)
		t.cursor--
		t.blink = 0
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyDelete) {
		runes := []rune(t.Value)
		if t.cursor < len(runes) {
			runes = append(runes[:t.cursor], runes[t.cursor+1:]...)
			t.Value = string(runes)
			t.blink = 0
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) && t.cursor > 0 {
		t.cursor--
		t.blink = 0
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) && t.cursor < len([]rune(t.Value)) {
		t.cursor++
		t.blink = 0
	}
}

func (t *TextInput) Draw(screen *ebiten.Image, x, y, w int, focused bool) {
	fillRect(screen, x, y, w, fieldH, colorInput)
	bc := colorBorder
	if focused {
		bc = colorFocused
	}
	strokeRect(screen, x, y, w, fieldH, bc)

	runes := []rune(t.Value)
	maxChars := (w - 6) / 6
	offset := 0
	if len(runes) > maxChars {
		offset = t.cursor - maxChars + 1
		if offset < 0 {
			offset = 0
		}
		if t.cursor < offset {
			offset = t.cursor
		}
	}
	end := offset + maxChars
	if end > len(runes) {
		end = len(runes)
	}
	display := string(runes[offset:end])
	ebitenutil.DebugPrintAt(screen, display, x+3, y+2)

	if focused && t.blink%60 < 40 {
		cx := x + 3 + (t.cursor-offset)*6
		if cx >= x+3 && cx <= x+w-3 {
			fillRect(screen, cx, y+2, 1, lineH, colorCursor)
		}
	}
}

func (t *TextInput) Clear() {
	t.Value = ""
	t.cursor = 0
	t.blink = 0
}
