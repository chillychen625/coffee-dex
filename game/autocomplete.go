package game

import (
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const maxDropItems = 6

type AutoComplete struct {
	TextInput
	options  []string
	matches  []string
	matchSel int
	DropOpen bool
	// last draw position — used by DrawDropdown to render on top of other content
	lastX, lastY, lastW int
}

func (a *AutoComplete) SetOptions(opts []string) {
	a.options = opts
	if a.Value != "" {
		a.filterMatches()
	}
}

func (a *AutoComplete) filterMatches() {
	lower := strings.ToLower(a.Value)
	a.matches = nil
	if lower == "" {
		a.DropOpen = false
		return
	}
	for _, opt := range a.options {
		if strings.Contains(strings.ToLower(opt), lower) {
			a.matches = append(a.matches, opt)
			if len(a.matches) >= maxDropItems {
				break
			}
		}
	}
	a.matchSel = 0
	a.DropOpen = len(a.matches) > 0
}

// Update handles input. Returns true if an arrow key or Escape was consumed by
// the dropdown (caller should skip field navigation for this frame).
func (a *AutoComplete) Update(focused bool) bool {
	if !focused {
		a.DropOpen = false
		return false
	}

	if a.DropOpen && len(a.matches) > 0 {
		if isKeyJustPressed(ebiten.KeyArrowDown) {
			if a.matchSel < len(a.matches)-1 {
				a.matchSel++
			}
			return true
		}
		if isKeyJustPressed(ebiten.KeyArrowUp) {
			if a.matchSel > 0 {
				a.matchSel--
				return true
			}
			// At top of list — collapse dropdown, let arrow-up navigate to prev field
			a.DropOpen = false
			return false
		}
		// Enter/Tab: accept suggestion, then let caller advance field normally
		if isKeyJustPressed(ebiten.KeyEnter) || isKeyJustPressed(ebiten.KeyTab) {
			a.Value = a.matches[a.matchSel]
			a.cursor = len([]rune(a.Value))
			a.DropOpen = false
			return false
		}
		if isKeyJustPressed(ebiten.KeyEscape) {
			a.DropOpen = false
			return true // suppress — don't propagate to scene navigation
		}
	}

	prev := a.Value
	a.TextInput.Update(focused)
	if a.Value != prev {
		a.filterMatches()
	}
	return false
}

// Draw renders the input box only. Call DrawDropdown after all other fields to
// ensure the suggestion list appears on top of everything else.
func (a *AutoComplete) Draw(screen *ebiten.Image, x, y, w int, focused bool) {
	a.lastX, a.lastY, a.lastW = x, y, w
	a.TextInput.Draw(screen, x, y, w, focused)
}

// DrawDropdown renders the suggestion list on top of all other content.
// Call this at the very end of your scene's Draw method.
func (a *AutoComplete) DrawDropdown(screen *ebiten.Image) {
	if !a.DropOpen || len(a.matches) == 0 {
		return
	}
	x, y, w := a.lastX, a.lastY, a.lastW
	dy := y + fieldH
	dh := len(a.matches) * (lineH + 2)
	fillRect(screen, x, dy, w, dh, colorPanel)
	strokeRect(screen, x, dy, w, dh, colorFocused)
	for i, m := range a.matches {
		iy := dy + i*(lineH+2)
		if i == a.matchSel {
			fillRect(screen, x, iy, w, lineH+2, colorSelected)
		}
		ebitenutil.DebugPrintAt(screen, truncate(m, w/6-1), x+3, iy+1)
	}
}

func (a *AutoComplete) Clear() {
	a.TextInput.Clear()
	a.DropOpen = false
	a.matches = nil
	a.matchSel = 0
}
