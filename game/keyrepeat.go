package game

import "github.com/hajimehoshi/ebiten/v2"

const (
	keyRepeatDelay    = 18 // ~300ms at 60fps before repeat starts
	keyRepeatInterval = 4  // ~67ms between repeats
)

var keyRepeatCounters = map[ebiten.Key]int{}

// isKeyActive returns true on the first press and then repeatedly after a
// short delay. Use this for held navigation (arrow keys, etc.).
// For action keys (Enter, Esc) use isKeyJustPressed instead.
func isKeyActive(k ebiten.Key) bool {
	if !ebiten.IsKeyPressed(k) {
		keyRepeatCounters[k] = 0
		return false
	}
	keyRepeatCounters[k]++
	n := keyRepeatCounters[k]
	if n == 1 {
		return true
	}
	if n > keyRepeatDelay && (n-keyRepeatDelay)%keyRepeatInterval == 0 {
		return true
	}
	return false
}
