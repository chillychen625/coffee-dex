package game

import (
	"fmt"
	"sort"

	"go-coffee-log/models"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type labSortMode int

const (
	labSortDate    labSortMode = iota // CoffeePokemon.CreatedAt desc
	labSortPokedex                    // PokemonID asc
	labSortName                       // PokemonName asc
	labSortRating                     // avg rating desc
	labSortModeCount
)

var labSortLabels = [labSortModeCount]string{"Date Added", "Pokédex #", "Name", "Avg Rating"}

// Layout constants for the detail view.
const (
	detailSpriteSize = 80 // sprite display size in pixels
	detailSpriteX    = InternalWidth - detailSpriteSize - 12
	detailSpriteY    = contentY + 4
	detailTextMaxX   = detailSpriteX - 6 // text must not go past the sprite
)

type PokemonLabScene struct {
	svc         *Services
	pokemon     []models.CoffeePokemon
	coffees     map[string]models.Coffee
	coffeeNames map[string]string
	avgRatings  map[string]float64
	sel         int
	labScroll   int
	detail      bool
	sortMode    labSortMode
}

func NewPokemonLabScene() *PokemonLabScene { return &PokemonLabScene{} }

func (s *PokemonLabScene) OnEnter(svc *Services) {
	s.svc = svc
	s.sel = 0
	s.labScroll = 0
	s.detail = false

	pokemon, err := svc.Pokemon.GetAllCoffeePokemon()
	if err == nil {
		s.pokemon = pokemon
	}

	coffeeList, err := svc.Coffee.ListCoffees()
	s.coffees = make(map[string]models.Coffee, len(coffeeList))
	s.coffeeNames = make(map[string]string, len(coffeeList))
	if err == nil {
		for _, c := range coffeeList {
			s.coffees[c.ID] = c
			s.coffeeNames[c.ID] = c.Name
		}
	}

	s.avgRatings = make(map[string]float64, len(s.pokemon))
	for _, p := range s.pokemon {
		agg, err := svc.Brew.GetAggregatedData(p.CoffeeID)
		if err == nil && agg != nil {
			s.avgRatings[p.CoffeeID] = agg.AverageRating
		}
	}

	s.applySort()
}

func (s *PokemonLabScene) applySort() {
	switch s.sortMode {
	case labSortDate:
		sort.Slice(s.pokemon, func(i, j int) bool {
			return s.pokemon[i].CreatedAt.After(s.pokemon[j].CreatedAt)
		})
	case labSortPokedex:
		sort.Slice(s.pokemon, func(i, j int) bool {
			return s.pokemon[i].PokemonID < s.pokemon[j].PokemonID
		})
	case labSortName:
		sort.Slice(s.pokemon, func(i, j int) bool {
			return s.pokemon[i].PokemonName < s.pokemon[j].PokemonName
		})
	case labSortRating:
		sort.Slice(s.pokemon, func(i, j int) bool {
			return s.avgRatings[s.pokemon[i].CoffeeID] > s.avgRatings[s.pokemon[j].CoffeeID]
		})
	}
	s.sel = 0
	s.labScroll = 0
}

func (s *PokemonLabScene) visibleCount() int {
	return (hintsY - contentY) / (lineH + 2)
}

func (s *PokemonLabScene) scrollToSel() {
	v := s.visibleCount()
	if s.sel < s.labScroll {
		s.labScroll = s.sel
	}
	if s.sel >= s.labScroll+v {
		s.labScroll = s.sel - v + 1
	}
}

func (s *PokemonLabScene) Update() SceneID {
	if s.detail {
		if isKeyJustPressed(ebiten.KeyEscape) || isKeyJustPressed(ebiten.KeyZ) {
			s.detail = false
		}
		return ScenePokemonLab
	}

	if isKeyJustPressed(ebiten.KeyEscape) {
		return SceneMenu
	}

	if isKeyActive(ebiten.KeyArrowDown) && s.sel < len(s.pokemon)-1 {
		s.sel++
		s.scrollToSel()
	}
	if isKeyActive(ebiten.KeyArrowUp) && s.sel > 0 {
		s.sel--
		s.scrollToSel()
	}

	if isKeyJustPressed(ebiten.KeyS) {
		s.sortMode = (s.sortMode + 1) % labSortModeCount
		s.applySort()
	}

	if isKeyJustPressed(ebiten.KeyEnter) || isKeyJustPressed(ebiten.KeyZ) {
		if len(s.pokemon) > 0 {
			s.detail = true
		}
	}
	return ScenePokemonLab
}

func (s *PokemonLabScene) Draw(screen *ebiten.Image) {
	drawBackground(screen)
	if s.detail && len(s.pokemon) > 0 {
		s.drawDetail(screen, s.pokemon[s.sel])
		return
	}
	s.drawList(screen)
}

func (s *PokemonLabScene) drawList(screen *ebiten.Image) {
	title := fmt.Sprintf("Pokédex  [S: %s ↕]", labSortLabels[s.sortMode])
	drawHeader(screen, title)

	if len(s.pokemon) == 0 {
		ebitenutil.DebugPrintAt(screen, "No Pokemon yet.", 10, contentY+10)
		ebitenutil.DebugPrintAt(screen, "Log 5+ brews for a coffee to generate one.", 10, contentY+24)
		drawHints(screen, "[Esc] Back")
		return
	}

	visible := s.visibleCount()
	for i := 0; i < visible && s.labScroll+i < len(s.pokemon); i++ {
		p := s.pokemon[s.labScroll+i]
		nick := p.Nickname
		if nick == "" {
			nick = p.PokemonName
		}
		rating := ""
		if r, ok := s.avgRatings[p.CoffeeID]; ok && r > 0 {
			rating = fmt.Sprintf("★%.1f", r)
		}
		coffeeName := truncate(s.coffeeNames[p.CoffeeID], 18)
		row := fmt.Sprintf("#%-3d %-14s  %-14s  Lv.%-3d  %-7s  %s",
			p.PokemonID,
			truncate(nick, 14),
			truncate(p.PokemonType, 14),
			p.Level,
			rating,
			coffeeName)
		rowY := contentY + i*(lineH+2)
		drawListRowTyped(screen, row, 0, rowY, InternalWidth, s.labScroll+i == s.sel, p.PokemonType)
	}

	if len(s.pokemon) > visible {
		prog := fmt.Sprintf("%d/%d", s.sel+1, len(s.pokemon))
		ebitenutil.DebugPrintAt(screen, prog, InternalWidth-len(prog)*6-8, hintsY-12)
	}

	drawHints(screen, "[↑↓] Scroll   [S] Sort   [Enter] Details   [Esc] Back")
}

func (s *PokemonLabScene) drawDetail(screen *ebiten.Image, p models.CoffeePokemon) {
	drawHeader(screen, fmt.Sprintf("Pokédex — #%03d %s", p.PokemonID, p.PokemonName))

	// Sprite — right side
	DrawSprite(screen, p.PokemonID, detailSpriteX, detailSpriteY, detailSpriteSize, detailSpriteSize)

	// Type badges below sprite
	typeBadgeY := detailSpriteY + detailSpriteSize + 4
	bx := detailSpriteX
	for _, part := range splitTypes(p.PokemonType) {
		bx = drawTypeBadge(screen, part, bx, typeBadgeY)
	}

	// Text — left side
	y := contentY + 6
	const textW = detailSpriteX - 10
	const charsPerLine = textW / 6

	nick := p.Nickname
	if nick == "" {
		nick = "(no nickname)"
	}
	coffeeName := s.coffeeNames[p.CoffeeID]
	if coffeeName == "" {
		coffeeName = "(unknown)"
	}

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Nickname:    %s", nick), 10, y)
	y += lineH + 4
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Coffee:      %s", truncate(coffeeName, charsPerLine-13)), 10, y)
	y += lineH + 4
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Level:       %d", p.Level), 10, y)
	y += lineH + 4
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Confidence:  %.0f%%", p.MappingConfidence*100), 10, y)
	y += lineH + 4
	if r, ok := s.avgRatings[p.CoffeeID]; ok && r > 0 {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Avg Rating:  %.1f/10", r), 10, y)
	}
	y += lineH + 6

	if c, ok := s.coffees[p.CoffeeID]; ok {
		var meta []string
		if c.Roaster != "" {
			meta = append(meta, c.Roaster)
		}
		if c.Origin != "" {
			meta = append(meta, c.Origin)
		}
		if c.RoastLevel != "" {
			meta = append(meta, c.RoastLevel)
		}
		if c.ProcessingMethod != "" {
			meta = append(meta, c.ProcessingMethod)
		}
		if len(meta) > 0 {
			line := ""
			for i, m := range meta {
				if i > 0 {
					line += " · "
				}
				line += m
			}
			ebitenutil.DebugPrintAt(screen, truncate(line, charsPerLine), 10, y)
			y += lineH + 4
		}
	}

	if p.LLMDescription != "" {
		fillRect(screen, 8, y, textW, 1, colorBorder)
		y += 6
		wrapText(screen, p.LLMDescription, 10, y, charsPerLine)
	}

	drawHints(screen, "[Esc/Z] Back to list")
}
