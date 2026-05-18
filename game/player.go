package game

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Direction the player is facing.
type Direction int

const (
	DirDown  Direction = iota
	DirUp
	DirLeft
	DirRight
)

// Player holds position and movement state.
type Player struct {
	// Pixel position (top-left of sprite)
	X, Y float64
	Dir  Direction

	// movement cooldown — frames until next step is allowed
	moveCooldown int
	world        *World
}

const (
	playerW        = 12
	playerH        = 16
	moveSpeed      = 16 // pixels per step (one full tile)
	moveCooldownMs = 10 // frames between steps
)

// NewPlayer creates a player starting at the given tile position.
func NewPlayer(tileX, tileY int, w *World) *Player {
	return &Player{
		X:     float64(tileX * TileSize),
		Y:     float64(tileY * TileSize),
		Dir:   DirDown,
		world: w,
	}
}

// TileX returns the player's current tile column.
func (p *Player) TileX() int { return int(p.X+playerW/2) / TileSize }

// TileY returns the player's current tile row.
func (p *Player) TileY() int { return int(p.Y+playerH) / TileSize }

// Update handles input and moves the player.
// Returns the SceneID of a building the player just entered, or SceneOverworld.
func (p *Player) Update() SceneID {
	if p.moveCooldown > 0 {
		p.moveCooldown--
		return SceneOverworld
	}

	dx, dy := 0, 0
	if ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA) {
		dx = -1
		p.Dir = DirLeft
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD) {
		dx = 1
		p.Dir = DirRight
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowUp) || ebiten.IsKeyPressed(ebiten.KeyW) {
		dy = -1
		p.Dir = DirUp
	} else if ebiten.IsKeyPressed(ebiten.KeyArrowDown) || ebiten.IsKeyPressed(ebiten.KeyS) {
		dy = 1
		p.Dir = DirDown
	}

	if dx != 0 || dy != 0 {
		nextTX := p.TileX() + dx
		nextTY := p.TileY() + dy
		if p.world.IsWalkable(nextTX, nextTY) {
			p.X += float64(dx * moveSpeed)
			p.Y += float64(dy * moveSpeed)
			p.moveCooldown = moveCooldownMs
		}
	}

	// Interact with a building when pressing Enter/Z/Space
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
		inpututil.IsKeyJustPressed(ebiten.KeyZ) ||
		inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		if b := p.world.BuildingAt(p.TileX(), p.TileY()); b != nil {
			return b.SceneID
		}
	}

	return SceneOverworld
}

// Draw renders the player as a placeholder colored rectangle.
func (p *Player) Draw(screen *ebiten.Image) {
	bodyColor := color.RGBA{R: 240, G: 80, B: 60, A: 255}
	vector.DrawFilledRect(screen, float32(p.X+(TileSize-playerW)/2), float32(p.Y), playerW, playerH, bodyColor, false)
}
