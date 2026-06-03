package game

import (
	"fmt"
	"image/color"
	"math"
	"sort"

	"go-coffee-log/models"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type MenuScene struct {
	svc      *Services
	sel      int
	tick     int
	pokemon  []models.CoffeePokemon
	pokIdx   int
	totalCof int
}

var menuItems = []struct {
	label   string
	desc    string
	sceneID SceneID
}{
	{"Roastery", "Log brews & manage gear", SceneRoastery},
	{"Bean Warehouse", "Add coffees, catch Pokemon", SceneWarehouse},
	{"Pokemon Lab", "Browse your Pokedex", ScenePokemonLab},
	{"Trophy Room", "Stats & analytics", SceneTrophyRoom},
}

func NewMenuScene() *MenuScene { return &MenuScene{} }

func (s *MenuScene) OnEnter(svc *Services) {
	s.svc = svc
	s.sel = 0

	if pokemon, err := svc.Pokemon.GetAllCoffeePokemon(); err == nil {
		s.pokemon = pokemon
		sort.Slice(s.pokemon, func(i, j int) bool {
			return s.pokemon[i].CreatedAt.After(s.pokemon[j].CreatedAt)
		})
	}
	if coffees, err := svc.Coffee.ListCoffees(); err == nil {
		s.totalCof = len(coffees)
	}
	if s.pokIdx >= len(s.pokemon) {
		s.pokIdx = 0
	}
}

func (s *MenuScene) Update() SceneID {
	s.tick++

	// Cycle through collected Pokemon every 5 seconds.
	if len(s.pokemon) > 1 && s.tick%300 == 0 {
		s.pokIdx = (s.pokIdx + 1) % len(s.pokemon)
	}

	if isKeyActive(ebiten.KeyArrowDown) && s.sel < len(menuItems)-1 {
		s.sel++
	}
	if isKeyActive(ebiten.KeyArrowUp) && s.sel > 0 {
		s.sel--
	}
	if isKeyJustPressed(ebiten.KeyEnter) || isKeyJustPressed(ebiten.KeyZ) || isKeyJustPressed(ebiten.KeySpace) {
		return menuItems[s.sel].sceneID
	}
	return SceneMenu
}

func (s *MenuScene) Draw(screen *ebiten.Image) {
	drawBackground(screen)

	// ── Background Pokeball decoration (top-right) ────────────────────────
	s.drawPokeball(screen, 388, 72, 58)

	// ── Banner ────────────────────────────────────────────────────────────
	const bannerH = 28
	fillRect(screen, 0, 0, InternalWidth, bannerH, colorHeader)
	fillRect(screen, 0, bannerH, InternalWidth, 1, colorBorder)

	// Double-draw title for a slight bold effect.
	ebitenutil.DebugPrintAt(screen, "Coffee Dex", 9, 5)
	ebitenutil.DebugPrintAt(screen, "Coffee Dex", 8, 4)
	ebitenutil.DebugPrintAt(screen, "Gotta Brew 'Em All!", 8, 17)

	caught := len(s.pokemon)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d/151 caught", caught), InternalWidth-84, 10)

	// ── Left panel: featured Pokemon ─────────────────────────────────────
	const spriteSize = 80
	const spritePanelW = 196
	const spriteCX = spritePanelW/2 - spriteSize/2 // x=58

	bounce := int(math.Sin(float64(s.tick)*0.08) * 3)
	spriteY := 42 + bounce

	if len(s.pokemon) > 0 {
		p := s.pokemon[s.pokIdx]
		tc := primaryTypeColor(p.PokemonType)

		// Type-colored accent strip on the far left.
		fillRect(screen, 0, bannerH+1, 3, InternalHeight-bannerH-1, tc)

		DrawSprite(screen, p.PokemonID, spriteCX, spriteY, spriteSize, spriteSize)

		// Name + dex number.
		nameY := spriteY + spriteSize + 5
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("#%03d %s", p.PokemonID, p.PokemonName), 8, nameY)
		nameY += lineH + 3

		// Type badge(s).
		bx := 8
		for _, part := range splitTypes(p.PokemonType) {
			bx = drawTypeBadge(screen, part, bx, nameY)
		}
		nameY += lineH + 8

		// Completion bar.
		ebitenutil.DebugPrintAt(screen, "DEX", 8, nameY+1)
		const barX = 32
		const barW = 130
		fillRect(screen, barX, nameY, barW, lineH, colorInput)
		strokeRect(screen, barX, nameY, barW, lineH, colorBorder)
		fill := int(float64(barW-2) * float64(caught) / 151.0)
		if fill > 0 {
			fillRect(screen, barX+1, nameY+1, fill, lineH-2,
				color.RGBA{R: 248, G: 208, B: 48, A: 255})
		}
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d/151", caught), barX+barW+4, nameY+1)
	} else {
		// No Pokemon yet — placeholder frame.
		strokeRect(screen, spriteCX, spriteY, spriteSize, spriteSize, colorBorder)
		ebitenutil.DebugPrintAt(screen, "?", spriteCX+spriteSize/2-3, spriteY+spriteSize/2-6)

		infoY := spriteY + spriteSize + 8
		ebitenutil.DebugPrintAt(screen, "No Pokemon yet.", 8, infoY)
		ebitenutil.DebugPrintAt(screen, "Log 5 brews for", 8, infoY+lineH+2)
		ebitenutil.DebugPrintAt(screen, "a coffee to start.", 8, infoY+lineH*2+4)
	}

	// ── Right panel: menu box (Game Boy dialogue style) ───────────────────
	const menuX = 200
	const menuY = bannerH + 4
	const menuW = InternalWidth - menuX - 6
	menuH := hintsY - menuY - 2

	fillRect(screen, menuX, menuY, menuW, menuH, colorPanel)
	// Outer border.
	strokeRect(screen, menuX, menuY, menuW, menuH, colorFocused)
	// Inner border for Game Boy double-frame feel.
	strokeRect(screen, menuX+3, menuY+3, menuW-6, menuH-6, colorBorder)

	for i, item := range menuItems {
		iy := menuY + 14 + i*((menuH-18)/len(menuItems))
		if i == s.sel {
			fillRect(screen, menuX+6, iy-1, menuW-12, 38, colorSelected)
			strokeRect(screen, menuX+6, iy-1, menuW-12, 38, colorFocused)
			ebitenutil.DebugPrintAt(screen, "►", menuX+10, iy+5)
			ebitenutil.DebugPrintAt(screen, item.label, menuX+22, iy+5)
			ebitenutil.DebugPrintAt(screen, item.desc, menuX+22, iy+19)
		} else {
			ebitenutil.DebugPrintAt(screen, "  "+item.label, menuX+10, iy+5)
		}
	}

	drawHints(screen, "[↑↓] Select   [Enter/Z] Open")
}

// drawPokeball draws a faint Pokeball ring decoration as a background element.
func (s *MenuScene) drawPokeball(screen *ebiten.Image, cx, cy int, r float32) {
	fcx, fcy := float32(cx), float32(cy)
	ring := color.RGBA{R: 45, G: 58, B: 88, A: 255}
	bg := colorBg
	btn := color.RGBA{R: 35, G: 46, B: 72, A: 255}

	// Outer filled circle.
	vector.DrawFilledCircle(screen, fcx, fcy, r, ring, true)
	// Inner cut-out to create ring effect.
	vector.DrawFilledCircle(screen, fcx, fcy, r-5, bg, true)
	// Horizontal divider band.
	fillRect(screen, cx-int(r), cy-2, int(2*r), 4, ring)
	// Re-fill the cut-out above/below divider so the band crosses cleanly.
	vector.DrawFilledCircle(screen, fcx, fcy-r/2, r/2-5, bg, true)
	vector.DrawFilledCircle(screen, fcx, fcy+r/2, r/2-5, bg, true)
	// Center button.
	vector.DrawFilledCircle(screen, fcx, fcy, 10, ring, true)
	vector.DrawFilledCircle(screen, fcx, fcy, 7, btn, true)
}
