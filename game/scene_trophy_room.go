package game

import (
	"fmt"
	"image/color"
	"sort"

	"go-coffee-log/service"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const trophyPageCount = 6

var trophyPageTitles = [trophyPageCount]string{
	"Overview",
	"Flavor Profile",
	"Type Distribution",
	"Origins",
	"Brewers & Processing",
	"Roasters",
}

// Sorted display slices built in OnEnter to avoid per-frame allocation.
type typeStat struct {
	name  string
	count int
}

type brewerEntry struct {
	name      string
	count     int
	avgRating float64
	avgTime   float64
}

type methodEntry struct {
	name      string
	count     int
	avgRating float64
}

type roasterEntry struct {
	name         string
	coffeeCount  int
	brewCount    int
	avgRating    float64
	pokemonCount int
}

type TrophyRoomScene struct {
	svc   *Services
	stats *service.Statistics
	err   string
	page  int

	sortedTypes    []typeStat
	sortedBrewers  []brewerEntry
	sortedMethods  []methodEntry
	sortedRoasters []roasterEntry
}

func NewTrophyRoomScene() *TrophyRoomScene { return &TrophyRoomScene{} }

func (s *TrophyRoomScene) OnEnter(svc *Services) {
	s.svc = svc
	s.page = 0
	stats, err := svc.Statistics.CalculateStatistics()
	if err != nil {
		s.err = err.Error()
		s.stats = nil
		return
	}
	s.stats = stats
	s.err = ""
	s.buildSortedSlices()
}

func (s *TrophyRoomScene) buildSortedSlices() {
	st := s.stats

	// Type distribution sorted by count desc.
	s.sortedTypes = nil
	for name, count := range st.TypeDistribution {
		s.sortedTypes = append(s.sortedTypes, typeStat{name, count})
	}
	sort.Slice(s.sortedTypes, func(i, j int) bool {
		return s.sortedTypes[i].count > s.sortedTypes[j].count
	})

	// Brewers sorted by count desc.
	s.sortedBrewers = nil
	for name, bs := range st.BrewerStats {
		s.sortedBrewers = append(s.sortedBrewers, brewerEntry{name, bs.Count, bs.AverageRating, bs.AvgBrewTime})
	}
	sort.Slice(s.sortedBrewers, func(i, j int) bool {
		return s.sortedBrewers[i].count > s.sortedBrewers[j].count
	})

	// Processing methods sorted by count desc.
	s.sortedMethods = nil
	for name, ps := range st.ProcessingStats {
		s.sortedMethods = append(s.sortedMethods, methodEntry{name, ps.Count, ps.AverageRating})
	}
	sort.Slice(s.sortedMethods, func(i, j int) bool {
		return s.sortedMethods[i].count > s.sortedMethods[j].count
	})

	// Roasters sorted by coffeeCount desc.
	s.sortedRoasters = nil
	for name, rs := range st.RoasterStats {
		s.sortedRoasters = append(s.sortedRoasters, roasterEntry{name, rs.CoffeeCount, rs.BrewCount, rs.AverageRating, rs.PokemonCount})
	}
	sort.Slice(s.sortedRoasters, func(i, j int) bool {
		return s.sortedRoasters[i].coffeeCount > s.sortedRoasters[j].coffeeCount
	})
}

func (s *TrophyRoomScene) Update() SceneID {
	if isKeyJustPressed(ebiten.KeyEscape) {
		return SceneMenu
	}
	if s.stats != nil {
		if isKeyJustPressed(ebiten.KeyTab) || isKeyJustPressed(ebiten.KeyArrowRight) || isKeyJustPressed(ebiten.KeyX) {
			s.page = (s.page + 1) % trophyPageCount
		}
		if isKeyJustPressed(ebiten.KeyArrowLeft) || isKeyJustPressed(ebiten.KeyZ) {
			s.page = (s.page + trophyPageCount - 1) % trophyPageCount
		}
	}
	return SceneTrophyRoom
}

func (s *TrophyRoomScene) Draw(screen *ebiten.Image) {
	drawBackground(screen)
	title := fmt.Sprintf("Trophy Room — %s  [%d/%d]", trophyPageTitles[s.page], s.page+1, trophyPageCount)
	drawHeader(screen, title)

	if s.err != "" {
		ebitenutil.DebugPrintAt(screen, "Error: "+s.err, 10, contentY+10)
		drawHints(screen, "[Esc] Back")
		return
	}
	if s.stats == nil {
		ebitenutil.DebugPrintAt(screen, "No data yet.", 10, contentY+10)
		drawHints(screen, "[Esc] Back")
		return
	}

	switch s.page {
	case 0:
		s.drawOverview(screen)
	case 1:
		s.drawFlavorProfile(screen)
	case 2:
		s.drawTypeDistribution(screen)
	case 3:
		s.drawOrigins(screen)
	case 4:
		s.drawBrewersProcessing(screen)
	case 5:
		s.drawRoasters(screen)
	}
	drawHints(screen, "[←/Z] Prev   [Tab/→/X] Next   [Esc] Back")
}

// ── Page 0: Overview ─────────────────────────────────────────────────────────

func (s *TrophyRoomScene) drawOverview(screen *ebiten.Image) {
	st := s.stats
	y := contentY + 6
	col2 := 240

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Coffees:        %d", st.TotalCoffees), 10, y)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Common Type:    %s", st.MostCommonType), col2, y)
	y += lineH + 2

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Pokemon:        %d", st.TotalPokemon), 10, y)
	if len(st.TopOrigins) > 0 {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Top Origin:     %s", st.TopOrigins[0].Origin), col2, y)
	}
	y += lineH + 2

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Completion:     %.1f%%", st.CompletionPercent), 10, y)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("High Conf.:     %d", st.HighConfidencePairings), col2, y)
	y += lineH + 2

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Avg Rating:     %.1f/10", st.AverageRating), 10, y)
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Avg Confidence: %.0f%%", st.AverageConfidence*100), col2, y)
	y += lineH + 6

	fillRect(screen, 8, y, InternalWidth-16, 1, colorBorder)
	y += 6

	if st.HighestRated != nil {
		ebitenutil.DebugPrintAt(screen, "Best:", 10, y)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%-24s  ★%d  %s",
			truncate(st.HighestRated.Name, 24), st.HighestRated.Rating,
			truncate(st.HighestRated.PokemonName, 14)), 46, y)
		y += lineH + 2
	}
	if st.LowestRated != nil {
		ebitenutil.DebugPrintAt(screen, "Worst:", 10, y)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%-24s  ★%d  %s",
			truncate(st.LowestRated.Name, 24), st.LowestRated.Rating,
			truncate(st.LowestRated.PokemonName, 14)), 46, y)
		y += lineH + 2
	}

	fillRect(screen, 8, y+2, InternalWidth-16, 1, colorMuted)
	y += 10

	// Quick 6-trait summary.
	tr := st.TraitAverages
	traits := []struct {
		name string
		val  int
	}{
		{"Sweet", tr.Sweetness}, {"Body", tr.Body},
		{"Roast", tr.RoastIntensity}, {"Bitter", tr.Bitterness},
		{"Floral", tr.Florality}, {"Citrus", tr.CitrusFruitsIntensity},
	}
	ebitenutil.DebugPrintAt(screen, "Trait Averages:", 10, y)
	y += lineH + 2
	barColor := color.RGBA{R: 80, G: 160, B: 220, A: 255}
	for i, t := range traits {
		x := 10 + (i%3)*156
		ry := y + (i/3)*(lineH+4)
		drawHBar(screen, x, ry, 6, 80, t.name, float64(t.val), 10, barColor)
	}
}

// ── Page 1: Flavor Profile ────────────────────────────────────────────────────

func (s *TrophyRoomScene) drawFlavorProfile(screen *ebiten.Image) {
	tr := s.stats.TraitAverages
	y := contentY + 6

	ebitenutil.DebugPrintAt(screen, "All-time averages from every brew  (0–10 scale)", 10, y)
	y += lineH + 6

	barColor := color.RGBA{R: 80, G: 160, B: 220, A: 255}
	const labelCols = 9
	const barW = 120
	const rowH = lineH + 4
	const col2 = 248

	left := []struct {
		name string
		val  int
	}{
		{"Berry", tr.BerryIntensity},
		{"Stonefrt", tr.StonefruitIntensity},
		{"Roast", tr.RoastIntensity},
		{"Citrus", tr.CitrusFruitsIntensity},
		{"Bitter", tr.Bitterness},
		{"Floral", tr.Florality},
	}
	right := []struct {
		name string
		val  int
	}{
		{"Spice", tr.Spice},
		{"Sweet", tr.Sweetness},
		{"Aroma", tr.AromaticIntensity},
		{"Savory", tr.Savory},
		{"Body", tr.Body},
		{"Clean", tr.Cleanliness},
	}

	for i, t := range left {
		drawHBar(screen, 8, y+i*rowH, labelCols, barW, t.name, float64(t.val), 10, barColor)
	}
	for i, t := range right {
		drawHBar(screen, col2, y+i*rowH, labelCols, barW, t.name, float64(t.val), 10, barColor)
	}
}

// ── Page 2: Type Distribution ─────────────────────────────────────────────────

func (s *TrophyRoomScene) drawTypeDistribution(screen *ebiten.Image) {
	if len(s.sortedTypes) == 0 {
		ebitenutil.DebugPrintAt(screen, "No type data yet.", 10, contentY+10)
		return
	}

	maxCount := s.sortedTypes[0].count

	y := contentY + 4
	const labelCols = 10
	const barW = 200
	const rowH = lineH + 4

	for i, t := range s.sortedTypes {
		if y+rowH > hintsY-lineH {
			break
		}
		// Draw label with type color badge.
		tc := getTypeColor(t.name)
		fillRect(screen, 8, y, 3, lineH, tc)
		ebitenutil.DebugPrintAt(screen, truncate(t.name, labelCols), 14, y+1)
		// Bar.
		bx := 8 + labelCols*6 + 4
		fillRect(screen, bx, y, barW, lineH, colorInput)
		strokeRect(screen, bx, y, barW, lineH, colorBorder)
		fill := int(float64(barW-2) * float64(t.count) / float64(maxCount))
		if fill > 0 {
			fillRect(screen, bx+1, y+1, fill, lineH-2, tc)
		}
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d", t.count), bx+barW+4, y+1)
		y += rowH
		_ = i
	}
}

// ── Page 3: Origins ───────────────────────────────────────────────────────────

func (s *TrophyRoomScene) drawOrigins(screen *ebiten.Image) {
	origins := s.stats.TopOrigins
	if len(origins) == 0 {
		ebitenutil.DebugPrintAt(screen, "No origin data yet.", 10, contentY+10)
		return
	}

	maxCount := 0
	for _, o := range origins {
		if o.Count > maxCount {
			maxCount = o.Count
		}
	}

	y := contentY + 4
	const labelCols = 14
	const barW = 160
	const rowH = lineH + 6
	barColor := color.RGBA{R: 80, G: 185, B: 120, A: 255}

	ebitenutil.DebugPrintAt(screen, "Origin            Coffees                  Avg Rating", 10, y)
	y += lineH + 4
	fillRect(screen, 8, y, InternalWidth-16, 1, colorBorder)
	y += 4

	for _, o := range origins {
		if y+rowH > hintsY-lineH {
			break
		}
		ebitenutil.DebugPrintAt(screen, truncate(o.Origin, labelCols), 10, y+1)
		bx := 10 + labelCols*6 + 4
		fillRect(screen, bx, y, barW, lineH, colorInput)
		strokeRect(screen, bx, y, barW, lineH, colorBorder)
		if maxCount > 0 {
			fill := int(float64(barW-2) * float64(o.Count) / float64(maxCount))
			if fill > 0 {
				fillRect(screen, bx+1, y+1, fill, lineH-2, barColor)
			}
		}
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d", o.Count), bx+barW+4, y+1)
		if o.AverageRating > 0 {
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("★%.1f", o.AverageRating), bx+barW+24, y+1)
		}
		y += rowH
	}
}

// ── Page 4: Brewers & Processing ─────────────────────────────────────────────

func (s *TrophyRoomScene) drawBrewersProcessing(screen *ebiten.Image) {
	y := contentY + 4
	const labelCols = 12
	const barW = 130
	const rowH = lineH + 4
	brewerColor := color.RGBA{R: 200, G: 140, B: 64, A: 255}
	methodColor := color.RGBA{R: 160, G: 100, B: 200, A: 255}

	// Brewers section.
	ebitenutil.DebugPrintAt(screen, "Brewers:", 10, y)
	y += lineH + 2

	if len(s.sortedBrewers) == 0 {
		ebitenutil.DebugPrintAt(screen, "  (none logged)", 10, y)
		y += lineH + 2
	} else {
		maxB := s.sortedBrewers[0].count
		for _, b := range s.sortedBrewers {
			if y+rowH > hintsY-lineH*5 {
				break
			}
			drawHBar(screen, 10, y, labelCols, barW, b.name, float64(b.count), float64(maxB), brewerColor)
			info := ""
			if b.avgRating > 0 {
				info += fmt.Sprintf("★%.1f", b.avgRating)
			}
			if b.avgTime > 0 {
				m := int(b.avgTime) / 60
				sc := int(b.avgTime) % 60
				if info != "" {
					info += "  "
				}
				info += fmt.Sprintf("%d:%02d avg", m, sc)
			}
			if info != "" {
				ebitenutil.DebugPrintAt(screen, info, 10+labelCols*6+barW+24, y+1)
			}
			y += rowH
		}
	}

	y += 4
	fillRect(screen, 8, y, InternalWidth-16, 1, colorBorder)
	y += 6

	// Processing methods section.
	ebitenutil.DebugPrintAt(screen, "Processing:", 10, y)
	y += lineH + 2

	if len(s.sortedMethods) == 0 {
		ebitenutil.DebugPrintAt(screen, "  (none logged)", 10, y)
	} else {
		maxM := s.sortedMethods[0].count
		for _, m := range s.sortedMethods {
			if y+rowH > hintsY-lineH {
				break
			}
			drawHBar(screen, 10, y, labelCols, barW, m.name, float64(m.count), float64(maxM), methodColor)
			if m.avgRating > 0 {
				ebitenutil.DebugPrintAt(screen, fmt.Sprintf("★%.1f", m.avgRating), 10+labelCols*6+barW+24, y+1)
			}
			y += rowH
		}
	}
}

// ── Page 5: Roasters ─────────────────────────────────────────────────────────

func (s *TrophyRoomScene) drawRoasters(screen *ebiten.Image) {
	if len(s.sortedRoasters) == 0 {
		ebitenutil.DebugPrintAt(screen, "No roaster data yet.", 10, contentY+10)
		return
	}

	y := contentY + 4
	const rowH = lineH + 4
	roasterColor := color.RGBA{R: 180, G: 120, B: 60, A: 255}

	ebitenutil.DebugPrintAt(screen, "Roaster               Coffees  Brews  ★Avg  Poke", 10, y)
	y += lineH + 4
	fillRect(screen, 8, y, InternalWidth-16, 1, colorBorder)
	y += 4

	maxCoffees := 1
	if len(s.sortedRoasters) > 0 {
		maxCoffees = s.sortedRoasters[0].coffeeCount
	}
	const barW = 60

	for _, r := range s.sortedRoasters {
		if y+rowH > hintsY-lineH {
			break
		}
		ebitenutil.DebugPrintAt(screen, truncate(r.name, 20), 10, y+1)

		// Small bar for coffee count
		bx := 130
		fillRect(screen, bx, y, barW, lineH, colorInput)
		strokeRect(screen, bx, y, barW, lineH, colorBorder)
		fill := int(float64(barW-2) * float64(r.coffeeCount) / float64(maxCoffees))
		if fill > 0 {
			fillRect(screen, bx+1, y+1, fill, lineH-2, roasterColor)
		}
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d", r.coffeeCount), bx+barW+4, y+1)

		col2 := bx + barW + 24
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d brews", r.brewCount), col2, y+1)
		col3 := col2 + 60
		if r.avgRating > 0 {
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("★%.1f", r.avgRating), col3, y+1)
		}
		if r.pokemonCount > 0 {
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%d", r.pokemonCount), col3+36, y+1)
		}
		y += rowH
	}
}
