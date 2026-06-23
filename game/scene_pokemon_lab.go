package game

import (
	"fmt"
	"image/color"
	"sort"
	"strings"

	"go-coffee-log/models"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type labSortMode int

const (
	labSortDate    labSortMode = iota
	labSortPokedex
	labSortName
	labSortRating
	labSortModeCount
)

var labSortLabels = [labSortModeCount]string{"Date Added", "Pokédex #", "Name", "Avg Rating"}

type labSubMode int

const (
	labSubList    labSubMode = iota
	labSubDetail
	labSubCompare
)

const (
	detailSpriteSize = 72
	detailSpriteX    = InternalWidth - detailSpriteSize - 10
	detailSpriteY    = contentY + 2
	detailTextMaxX   = detailSpriteX - 6
)

type PokemonLabScene struct {
	svc         *Services
	pokemon     []models.CoffeePokemon
	coffees     map[string]models.Coffee
	coffeeNames map[string]string
	avgRatings  map[string]float64
	aggData     map[string]*models.AggregatedBrewData
	sel         int
	labScroll   int
	sortMode    labSortMode
	labSub      labSubMode

	// Compare mode
	compareA    int
	compareB    int
	comparePick int // 0=picking B, 1=showing result

	// kept for backwards compat with existing detail flag uses
	detail bool
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
	s.aggData = make(map[string]*models.AggregatedBrewData, len(s.pokemon))
	for _, p := range s.pokemon {
		agg, err := svc.Brew.GetAggregatedData(p.CoffeeID)
		if err == nil && agg != nil {
			s.avgRatings[p.CoffeeID] = agg.AverageRating
			s.aggData[p.CoffeeID] = agg
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
		if isKeyJustPressed(ebiten.KeyEnter) || isKeyJustPressed(ebiten.KeyZ) || isKeyJustPressed(ebiten.KeyEscape) {
			s.labSub = labSubList
		}
		return ScenePokemonLab
	}
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

// ── List ─────────────────────────────────────────────────────────────────────

func (s *PokemonLabScene) drawList(screen *ebiten.Image) {
	title := fmt.Sprintf("Pokédex  [S: %s]", labSortLabels[s.sortMode])
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
		coffeeName := truncate(s.coffeeNames[p.CoffeeID], 26)
		nick := p.Nickname
		if nick == "" {
			nick = p.PokemonName
		}
		rating := "     "
		if r, ok := s.avgRatings[p.CoffeeID]; ok && r > 0 {
			rating = fmt.Sprintf("★%4.1f", r)
		}
		// Coffee name first so you know what you're looking at immediately
		row := fmt.Sprintf("%-26s  %-12s  %-14s  %s",
			coffeeName,
			truncate(nick, 12),
			truncate(p.PokemonType, 14),
			rating)
		rowY := contentY + i*(lineH+2)
		drawListRowTyped(screen, row, 0, rowY, InternalWidth, s.labScroll+i == s.sel, p.PokemonType)
	}

	if len(s.pokemon) > visible {
		prog := fmt.Sprintf("%d/%d", s.sel+1, len(s.pokemon))
		ebitenutil.DebugPrintAt(screen, prog, InternalWidth-len(prog)*6-8, hintsY-12)
	}
	drawHints(screen, "[↑↓] Scroll   [S] Sort   [C] Compare   [Enter] Details   [Esc] Back")
}

// ── Detail ───────────────────────────────────────────────────────────────────

func (s *PokemonLabScene) drawDetail(screen *ebiten.Image, p models.CoffeePokemon) {
	coffeeName := s.coffeeNames[p.CoffeeID]
	if coffeeName == "" {
		coffeeName = "(unknown)"
	}
	drawHeader(screen, fmt.Sprintf("#%03d %s  —  %s", p.PokemonID, p.PokemonName, truncate(coffeeName, 32)))

	// Sprite — top right
	DrawSprite(screen, p.PokemonID, detailSpriteX, detailSpriteY, detailSpriteSize, detailSpriteSize)

	// Type badges below sprite
	bx := detailSpriteX
	for _, part := range splitTypes(p.PokemonType) {
		bx = drawTypeBadge(screen, part, bx, detailSpriteY+detailSpriteSize+4)
	}

	const textW = detailSpriteX - 10
	const cpl = textW / 6 // chars per line
	y := contentY + 4

	// ── Coffee identity block ─────────────────────────────────────────────────
	c, hasCoffee := s.coffees[p.CoffeeID]
	if hasCoffee {
		if c.Roaster != "" {
			ebitenutil.DebugPrintAt(screen, truncate(c.Roaster, cpl), 10, y)
			y += lineH + 1
		}
		parts := []string{}
		if c.Origin != "" {
			parts = append(parts, c.Origin)
		}
		if c.Variety != "" {
			parts = append(parts, c.Variety)
		}
		if c.RoastLevel != "" {
			parts = append(parts, c.RoastLevel)
		}
		if c.ProcessingMethod != "" {
			parts = append(parts, c.ProcessingMethod)
		}
		if len(parts) > 0 {
			ebitenutil.DebugPrintAt(screen, truncate(strings.Join(parts, " · "), cpl), 10, y)
			y += lineH + 1
		}
		if c.RoastDate != nil && !c.RoastDate.IsZero() {
			daysOff := c.DaysOffRoast()
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Roasted %s  (%d days ago)", c.RoastDate.Time().Format("Jan 2 2006"), daysOff), 10, y)
			y += lineH + 1
		}
	}

	// ── Stats row ─────────────────────────────────────────────────────────────
	y += 2
	fillRect(screen, 8, y, textW, 1, colorBorder)
	y += 5

	agg := s.aggData[p.CoffeeID]
	statLine := fmt.Sprintf("Lv.%d  Conf:%.0f%%", p.Level, p.MappingConfidence*100)
	if agg != nil {
		statLine += fmt.Sprintf("  Brews:%d", agg.BrewCount)
		if agg.AverageRating > 0 {
			statLine += fmt.Sprintf("  ★%.1f/10", agg.AverageRating)
		}
	}
	ebitenutil.DebugPrintAt(screen, statLine, 10, y)
	y += lineH + 3

	// ── Trait bars ────────────────────────────────────────────────────────────
	if agg != nil {
		tr := agg.AverageTraits
		type traitVal struct {
			name string
			val  int
		}
		traits := []traitVal{
			{"Sweet", tr.Sweetness},
			{"Floral", tr.Florality},
			{"Citrus", tr.CitrusFruitsIntensity},
			{"Berry", tr.BerryIntensity},
			{"Stonefrt", tr.StonefruitIntensity},
			{"Body", tr.Body},
			{"Roast", tr.RoastIntensity},
			{"Bitter", tr.Bitterness},
			{"Aroma", tr.AromaticIntensity},
			{"Spice", tr.Spice},
			{"Savory", tr.Savory},
			{"Clean", tr.Cleanliness},
		}
		// only show traits with data (>= 0)
		var scored []traitVal
		for _, t := range traits {
			if t.val >= 0 {
				scored = append(scored, t)
			}
		}
		if len(scored) > 0 {
			ebitenutil.DebugPrintAt(screen, "Flavor profile:", 10, y)
			y += lineH + 1
			barColor := color.RGBA{R: 80, G: 160, B: 220, A: 255}
			const labelCols = 8
			const barW = 60
			cols := 2
			colWidth := textW / cols
			for i, t := range scored {
				col := i % cols
				row := i / cols
				rx := 10 + col*colWidth
				ry := y + row*(lineH+2)
				drawHBar(screen, rx, ry, labelCols, barW, t.name, float64(t.val), 10, barColor)
			}
			rows := (len(scored) + cols - 1) / cols
			y += rows*(lineH+2) + 3
		}
	}

	// ── Tasting notes ─────────────────────────────────────────────────────────
	if agg != nil && len(agg.CombinedNotes) > 0 {
		fillRect(screen, 8, y, textW, 1, colorBorder)
		y += 5
		// Collect unique non-empty notes (up to 6)
		seen := map[string]bool{}
		var notes []string
		for _, n := range agg.CombinedNotes {
			n = strings.TrimSpace(n)
			if n != "" && !seen[strings.ToLower(n)] {
				seen[strings.ToLower(n)] = true
				notes = append(notes, n)
			}
		}
		if len(notes) > 8 {
			notes = notes[:8]
		}
		if len(notes) > 0 {
			ebitenutil.DebugPrintAt(screen, "Notes: "+truncate(strings.Join(notes, "  ·  "), cpl-7), 10, y)
			y += lineH + 3
		}
	}

	// ── Pokedex description ───────────────────────────────────────────────────
	if p.LLMDescription != "" {
		fillRect(screen, 8, y, textW, 1, colorBorder)
		y += 5
		wrapText(screen, p.LLMDescription, 10, y, cpl)
	}

	drawHints(screen, "[Esc/Z] Back to list")
}

// ── Compare ──────────────────────────────────────────────────────────────────

func (s *PokemonLabScene) drawCompare(screen *ebiten.Image) {
	if s.comparePick == 0 {
		drawHeader(screen, "Compare — pick second Pokemon")
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
			coffee := truncate(s.coffeeNames[p.CoffeeID], 22)
			row := fmt.Sprintf("%s %-22s  %-12s  %s", prefix, coffee, truncate(nick, 12), truncate(p.PokemonType, 12))
			drawListRowTyped(screen, row, 0, contentY+i*(lineH+2), InternalWidth, idx == s.compareB, p.PokemonType)
		}
		drawHints(screen, "[↑↓] Select   [Enter/Z] Confirm   [Esc] Cancel")
		return
	}

	// ── Side-by-side result ───────────────────────────────────────────────────
	pA := s.pokemon[s.compareA]
	pB := s.pokemon[s.compareB]
	drawHeader(screen, fmt.Sprintf("%s  vs  %s", truncate(pA.PokemonName, 16), truncate(pB.PokemonName, 16)))

	const half = InternalWidth / 2
	const sprSize = 56
	const panelPad = 4

	// Sprites
	DrawSprite(screen, pA.PokemonID, half/2-sprSize/2, contentY+1, sprSize, sprSize)
	DrawSprite(screen, pB.PokemonID, half+half/2-sprSize/2, contentY+1, sprSize, sprSize)

	// Divider
	fillRect(screen, half-1, contentY, 2, hintsY-contentY, colorBorder)

	y := contentY + sprSize + 4
	const colW = half - panelPad*2
	const cpl = colW / 6

	aggA := s.aggData[pA.CoffeeID]
	aggB := s.aggData[pB.CoffeeID]

	// Draw one panel
	drawPanel := func(p models.CoffeePokemon, agg *models.AggregatedBrewData, xOff int) {
		ry := y

		// Pokemon + type
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("#%03d %s", p.PokemonID, truncate(p.PokemonName, cpl-5)), xOff+panelPad, ry)
		ry += lineH + 1
		bx := xOff + panelPad
		for _, part := range splitTypes(p.PokemonType) {
			bx = drawTypeBadge(screen, part, bx, ry)
		}
		ry += lineH + 3

		// Coffee identity
		coffeeName := s.coffeeNames[p.CoffeeID]
		ebitenutil.DebugPrintAt(screen, truncate(coffeeName, cpl), xOff+panelPad, ry)
		ry += lineH + 1
		if c, ok := s.coffees[p.CoffeeID]; ok {
			meta := ""
			if c.Origin != "" {
				meta = c.Origin
			}
			if c.RoastLevel != "" {
				if meta != "" {
					meta += " · "
				}
				meta += c.RoastLevel
			}
			if c.ProcessingMethod != "" {
				if meta != "" {
					meta += " · "
				}
				meta += c.ProcessingMethod
			}
			if meta != "" {
				ebitenutil.DebugPrintAt(screen, truncate(meta, cpl), xOff+panelPad, ry)
				ry += lineH + 1
			}
		}

		// Stats
		statLine := fmt.Sprintf("Lv.%d", p.Level)
		if agg != nil && agg.AverageRating > 0 {
			statLine += fmt.Sprintf("  ★%.1f", agg.AverageRating)
		}
		if agg != nil {
			statLine += fmt.Sprintf("  %d brews", agg.BrewCount)
		}
		ebitenutil.DebugPrintAt(screen, statLine, xOff+panelPad, ry)
		ry += lineH + 3

		// Trait bars
		if agg != nil {
			fillRect(screen, xOff+panelPad, ry, colW-panelPad, 1, colorBorder)
			ry += 4
			drawTraitMini(screen, agg.AverageTraits, xOff+panelPad, ry, colW-panelPad*2)
			ry += 8*(lineH+1) + 2
		}

		// Top notes
		if agg != nil && len(agg.CombinedNotes) > 0 {
			fillRect(screen, xOff+panelPad, ry, colW-panelPad, 1, colorBorder)
			ry += 4
			var notes []string
			seen := map[string]bool{}
			for _, n := range agg.CombinedNotes {
				n = strings.TrimSpace(n)
				if n != "" && !seen[strings.ToLower(n)] {
					seen[strings.ToLower(n)] = true
					notes = append(notes, n)
				}
			}
			for i, n := range notes {
				if i >= 4 || ry+lineH >= hintsY {
					break
				}
				ebitenutil.DebugPrintAt(screen, "· "+truncate(n, cpl-2), xOff+panelPad, ry)
				ry += lineH + 1
			}
		}
	}

	drawPanel(pA, aggA, 0)
	drawPanel(pB, aggB, half)

	// Highlight winner in avg rating
	if aggA != nil && aggB != nil && aggA.AverageRating != aggB.AverageRating {
		winX := 0
		if aggB.AverageRating > aggA.AverageRating {
			winX = half
		}
		fillRect(screen, winX+panelPad, y+sprSize-sprSize, colW-panelPad, 1, color.RGBA{R: 60, G: 200, B: 100, A: 200})
	}

	drawHints(screen, "[Enter/Z/Esc] Back")
}

// drawTraitMini renders a compact 2-column trait grid within a panel.
func drawTraitMini(screen *ebiten.Image, tr models.TastingTraits, x, y, width int) {
	type tv struct {
		name string
		val  int
	}
	traits := []tv{
		{"Sweet", tr.Sweetness},
		{"Floral", tr.Florality},
		{"Citrus", tr.CitrusFruitsIntensity},
		{"Berry", tr.BerryIntensity},
		{"Body", tr.Body},
		{"Roast", tr.RoastIntensity},
		{"Bitter", tr.Bitterness},
		{"Clean", tr.Cleanliness},
	}
	barColor := color.RGBA{R: 80, G: 160, B: 220, A: 255}
	const labelCols = 6
	barW := width/2 - labelCols*6 - 6
	if barW < 10 {
		barW = 10
	}
	colWidth := width / 2
	for i, t := range traits {
		if t.val < 0 {
			continue
		}
		col := i % 2
		row := i / 2
		rx := x + col*colWidth
		ry := y + row*(lineH+1)
		drawHBar(screen, rx, ry, labelCols, barW, t.name, float64(t.val), 10, barColor)
	}
}
