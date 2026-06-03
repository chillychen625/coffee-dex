package game

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// DateInput is a structured YYYY-MM-DD entry widget. Typing digits auto-advances
// between segments; Tab moves forward, Backspace moves back.
type DateInput struct {
	parts [3]string // ["YYYY", "MM", "DD"]
	seg   int       // 0=year, 1=month, 2=day
}

var dateMaxLen = [3]int{4, 2, 2}
var datePlaceholder = [3]string{"YYYY", "MM", "DD"}

func (d *DateInput) Update(focused bool) {
	if !focused {
		return
	}

	var chars []rune
	chars = ebiten.AppendInputChars(chars)
	for _, ch := range chars {
		if ch >= '0' && ch <= '9' {
			ml := dateMaxLen[d.seg]
			if len(d.parts[d.seg]) < ml {
				d.parts[d.seg] += string(ch)
				// Auto-advance to next segment when full
				if len(d.parts[d.seg]) == ml && d.seg < 2 {
					d.seg++
				}
			}
		}
	}

	if isKeyJustPressed(ebiten.KeyBackspace) {
		if len(d.parts[d.seg]) > 0 {
			d.parts[d.seg] = d.parts[d.seg][:len(d.parts[d.seg])-1]
		} else if d.seg > 0 {
			d.seg--
		}
	}
	// Tab advances within the date widget (caller should not also advance field)
	if isKeyJustPressed(ebiten.KeyTab) && d.seg < 2 {
		d.seg++
	}
}

// ValueString returns "YYYY-MM-DD" if all segments are filled, otherwise "".
func (d *DateInput) ValueString() string {
	if len(d.parts[0]) < 4 || len(d.parts[1]) < 2 || len(d.parts[2]) < 2 {
		return ""
	}
	return fmt.Sprintf("%s-%s-%s", d.parts[0], d.parts[1], d.parts[2])
}

func (d *DateInput) IsEmpty() bool {
	return d.parts[0] == "" && d.parts[1] == "" && d.parts[2] == ""
}

func (d *DateInput) Clear() {
	d.parts = [3]string{}
	d.seg = 0
}

func (d *DateInput) Draw(screen *ebiten.Image, x, y, w int, focused bool) {
	fillRect(screen, x, y, w, fieldH, colorInput)
	bc := colorBorder
	if focused {
		bc = colorFocused
	}
	strokeRect(screen, x, y, w, fieldH, bc)

	cx := x + 3
	for i := 0; i < 3; i++ {
		part := d.parts[i]
		ph := datePlaceholder[i]
		display := part
		// Pad with underscores to show expected length
		for len(display) < dateMaxLen[i] {
			display += "_"
		}
		if part == "" {
			display = ph
		}

		// Highlight the active segment when focused
		if focused && i == d.seg {
			sw := len(display) * 6
			fillRect(screen, cx-1, y+1, sw+2, fieldH-2, colorSelected)
		}

		ebitenutil.DebugPrintAt(screen, display, cx, y+2)
		cx += len(display) * 6

		if i < 2 {
			ebitenutil.DebugPrintAt(screen, "-", cx, y+2)
			cx += 6
		}
	}
}
