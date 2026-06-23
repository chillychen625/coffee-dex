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

type labSubMode int

const (
	labSubList    labSubMode = iota // Normal list view
	labSubDetail                    // Detail view
	labSubCompare                   // Comparison mode
)

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

	// Comparison mode
	labSub       labSubMode
	compareA     int // index of Pokemon A
	compareB     int // index of Pokemon B (while picking)
	comparePick  int // 0=picking B, 1=showing result
}

func NewPokemonLabScene() *PokemonLabScene { return &PokemonLabScene{} }

func (s *PokemonLabScene) OnEnter(svc *Services) {
	s.svc = svc
	s.sel = 0
	s.labScroll = 0
	s.detail = false
	s.labSub = labSubList

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
	// Handle comparison mode.
	if s.labSub == labSubCompare {
		return s.updateCompare()
	}

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

	// C key enters compare mode — set A to current selection.
	if isKeyJustPressed(ebiten.KeyC) && len(s.pokemon) >= 2 {
		s.compareA = s.sel
		s.compareB = s.sel
		s.comparePick = 0
		s.labSub = labSubCompare
	}

	if isKeyJustPressed(ebiten.KeyEnter) || isKeyJustPressed(ebiten.KeyZ) {
		if len(s.pokemon) > 0 {
			s.detail = true
		}
	}
	return ScenePokemonLab
}

func (s *PokemonLabScene) updateCompare() SceneID {
	if isKeyJustPressed(ebiten.KeyEscape) {
		s.labSub = labSubList
		return ScenePokemonLab
	}

	if s.comparePick == 1 {
		// Showing result — any key goes back.
		if isKeyJustPressed(ebiten.KeyEnter) || isKeyJustPressed(ebiten.KeyZ) {
			s.labSub = labSubList
		}
		return ScenePokemonLab
	}

	// Picking B — navigate the list.
	if isKeyActive(ebiten.KeyArrowDown) && s.compareB < len(s.pokemon)-1 {
		s.compareB++
	}
	if isKeyActive(ebiten.KeyArrowUp) && s.compareB > 0 {
		s.compareB--
	}
	if isKeyJustPressed(ebiten.KeyEnter) || isKeyJustPressed(ebiten.KeyZ) {
		if s.compareB != s.compareA {
			s.comparePick = 1
		}
	}
	return ScenePokemonLab
}

func (s *PokemonLabScene) Draw(screen *ebiten.Image) {
	drawBackground(screen)
	if s.labSub == labSubCompare {
		s.drawCompare(screen)
		return
	}
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

	drawHints(screen, "[↑↓] Scroll   [S] Sort   [C] Compare   [Enter] Details   [Esc] Back")
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

// drawCompare renders the Pokemon comparison view.
func (s *PokemonLabScene) drawCompare(screen *ebiten.Image) {
	if s.comparePick == 0 {
		// Picking B: show list with A highlighted and current selection marked.
		drawHeader(screen, "Compare — Select second Pokemon")
		visible := s.visibleCount()
		offset := s.compareB - visible/2
		if offset < 0 {
			offset = 0
		}
		for i := 0; i < visible && offset+i < len(s.pokemon); i++ {
			idx := offset + i
			p := s.pokemon[idx]
			nick := p.Nickname
			if nick == "" {
				nick = p.PokemonName
			}
			prefix := "  "
			if idx == s.compareA {
				prefix = "A:"
			} else if idx == s.compareB {
				prefix = "B>"
			}
			row := fmt.Sprintf("%s #%-3d %-14s  %s", prefix, p.PokemonID, truncate(nick, 14), truncate(p.PokemonType, 14))
			drawListRowTyped(screen, row, 0, contentY+i*(lineH+2), InternalWidth, idx == s.compareB, p.PokemonType)
		}
		drawHints(screen, "[↑↓] Select B   [Enter/Z] Confirm   [Esc] Cancel")
		return
	}

	// Show side-by-side comparison.
	pA := s.pokemon[s.compareA]
	pB := s.pokemon[s.compareB]
	drawHeader(screen, fmt.Sprintf("Compare: %s vs %s", truncate(pA.PokemonName, 18), truncate(pB.PokemonName, 18)))

	const half = InternalWidth / 2
	const sprSize = 60

	// Sprites
	DrawSprite(screen, pA.PokemonID, half/2-sprSize/2, contentY+2, sprSize, sprSize)
	DrawSprite(screen, pB.PokemonID, half+half/2-sprSize/2, contentY+2, sprSize, sprSize)

	// Divider
	fillRect(screen, half-1, contentY, 2, hintsY-contentY, colorBorder)

	y := contentY + sprSize + 6
	const colW = half - 4
	const chars = colW / 6

	drawPokemonComparePanel := func(p models.CoffeePokemon, xOff int) {
		name := p.PokemonName
		coffee := truncate(s.coffeeNames[p.CoffeeID], chars-9)
		rating := ""
		if r, ok := s.avgRatings[p.CoffeeID]; ok && r > 0 {
			rating = fmt.Sprintf("★%.1f", r)
		}

		ry := y
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("#%-3d %s", p.PokemonID, truncate(name, chars-5)), xOff+2, ry)
		ry += lineH + 2
		ebitenutil.DebugPrintAt(screen, truncate(p.PokemonType, chars), xOff+2, ry)
		ry += lineH + 2
		ebitenutil.DebugPrintAt(screen, "Coffee: "+coffee, xOff+2, ry)
		ry += lineH + 2
		if rating != "" {
			ebitenutil.DebugPrintAt(screen, "Rating: "+rating, xOff+2, ry)
		}
		ry += lineH + 2
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Lv.%d  Conf:%.0f%%", p.Level, p.MappingConfidence*100), xOff+2, ry)
	}

	drawPokemonComparePanel(pA, 0)
	drawPokemonComparePanel(pB, half)

	drawHints(screen, "[Enter/Z] Back   [Esc] Back")
}
