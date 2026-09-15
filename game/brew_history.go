package game

import (
	"fmt"
	"image/color"
	"strings"

	"go-coffee-log/models"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// brewMetric identifies which value to graph.
type brewMetric int

const (
	bmRating brewMetric = iota
	bmSweetness
	bmFlorality
	bmCitrus
	bmBerry
	bmStonefruit
	bmBody
	bmRoast
	bmBitterness
	bmAroma
	bmSpice
	bmSavory
	bmCleanliness
	bmCount
)

var brewMetricLabels = [bmCount]string{
	"Rating", "Sweet", "Floral", "Citrus", "Berry",
	"Stonefrt", "Body", "Roast", "Bitter", "Aroma",
	"Spice", "Savory", "Clean",
}

func (bm brewMetric) valueFor(b models.Brew) float64 {
	switch bm {
	case bmRating:
		return float64(b.Rating)
	case bmSweetness:
		return float64(b.TastingTraits.Sweetness)
	case bmFlorality:
		return float64(b.TastingTraits.Florality)
	case bmCitrus:
		return float64(b.TastingTraits.CitrusFruitsIntensity)
	case bmBerry:
		return float64(b.TastingTraits.BerryIntensity)
	case bmStonefruit:
		return float64(b.TastingTraits.StonefruitIntensity)
	case bmBody:
		return float64(b.TastingTraits.Body)
	case bmRoast:
		return float64(b.TastingTraits.RoastIntensity)
	case bmBitterness:
		return float64(b.TastingTraits.Bitterness)
	case bmAroma:
		return float64(b.TastingTraits.AromaticIntensity)
	case bmSpice:
		return float64(b.TastingTraits.Spice)
	case bmSavory:
		return float64(b.TastingTraits.Savory)
	case bmCleanliness:
		return float64(b.TastingTraits.Cleanliness)
	}
	return -1
}

// BrewHistoryView is a reusable sub-view showing per-brew scores and a graph.
// Embed it in a scene and delegate Update/Draw when active.
type BrewHistoryView struct {
	CoffeeName string
	brews      []models.Brew
	sel        int
	scroll     int
	metric     brewMetric
}

func (v *BrewHistoryView) Load(brews []models.Brew, coffeeName string) {
	v.brews = brews
	v.CoffeeName = coffeeName
	v.sel = 0
	v.scroll = 0
	v.metric = bmRating
}

const (
	bhListW    = 196 // left panel width
	bhDivider  = 200 // x position of divider
	bhRightX   = 204 // right panel start x
	bhRightW   = InternalWidth - bhRightX - 2
	bhListRows = 10  // visible brew rows in list
	bhGraphH   = 70  // pixel height of bar graph
)

// Update handles input; returns true if Esc/Z pressed (caller should exit).
func (v *BrewHistoryView) Update() bool {
	if isKeyJustPressed(ebiten.KeyEscape) || isKeyJustPressed(ebiten.KeyZ) {
		return true
	}
	if isKeyActive(ebiten.KeyArrowDown) && v.sel < len(v.brews)-1 {
		v.sel++
		if v.sel >= v.scroll+bhListRows {
			v.scroll = v.sel - bhListRows + 1
		}
	}
	if isKeyActive(ebiten.KeyArrowUp) && v.sel > 0 {
		v.sel--
		if v.sel < v.scroll {
			v.scroll = v.sel
		}
	}
	if isKeyJustPressed(ebiten.KeyArrowRight) {
		v.metric = (v.metric + 1) % bmCount
	}
	if isKeyJustPressed(ebiten.KeyArrowLeft) {
		v.metric = (v.metric + bmCount - 1) % bmCount
	}
	return false
}

func (v *BrewHistoryView) Draw(screen *ebiten.Image) {
	title := fmt.Sprintf("Brews — %s", truncate(v.CoffeeName, 38))
	drawHeader(screen, title)

	if len(v.brews) == 0 {
		ebitenutil.DebugPrintAt(screen, "No brews logged yet.", 10, contentY+20)
		drawHints(screen, "[Esc/Z] Back")
		return
	}

	// Vertical divider
	fillRect(screen, bhDivider, contentY, 2, hintsY-contentY, colorBorder)

	v.drawList(screen)
	v.drawGraph(screen)
	v.drawBrewDetail(screen)

	drawHints(screen, "[↑↓] Select brew   [←→] Cycle metric   [Esc/Z] Back")
}

func (v *BrewHistoryView) drawList(screen *ebiten.Image) {
	for i := 0; i < bhListRows && v.scroll+i < len(v.brews); i++ {
		b := v.brews[v.scroll+i]
		idx := v.scroll + i
		date := b.CreatedAt.Format("01/02")
		dripper := truncate(b.Dripper, 7)
		if dripper == "" {
			dripper = "—"
		}
		row := fmt.Sprintf("#%-2d %s ★%2d %-7s", idx+1, date, b.Rating, dripper)
		rowY := contentY + i*(lineH+2)
		drawListRow(screen, row, 0, rowY, bhListW, idx == v.sel)
	}
	// Scroll indicator
	if len(v.brews) > bhListRows {
		prog := fmt.Sprintf("%d/%d", v.sel+1, len(v.brews))
		ebitenutil.DebugPrintAt(screen, prog, bhListW-len(prog)*6-2, hintsY-lineH-2)
	}
}

func (v *BrewHistoryView) drawGraph(screen *ebiten.Image) {
	// Header: metric name with arrows
	label := fmt.Sprintf("◄ %s ►", brewMetricLabels[v.metric])
	lx := bhRightX + (bhRightW-len(label)*6)/2
	ebitenutil.DebugPrintAt(screen, label, lx, contentY+1)

	graphY := contentY + lineH + 3
	graphW := bhRightW
	graphX := bhRightX

	// Graph background
	fillRect(screen, graphX, graphY, graphW, bhGraphH, colorPanel)
	strokeRect(screen, graphX, graphY, graphW, bhGraphH, colorBorder)

	n := len(v.brews)
	if n == 0 {
		return
	}

	barW := graphW / n
	if barW < 2 {
		barW = 2
	}
	if barW > 20 {
		barW = 20
	}
	gap := 1
	if barW <= 3 {
		gap = 0
	}
	totalBarW := barW + gap
	startX := graphX + (graphW-n*totalBarW)/2
	if startX < graphX {
		startX = graphX
	}

	barColor := color.RGBA{R: 80, G: 160, B: 220, A: 255}
	selColor := color.RGBA{R: 255, G: 200, B: 60, A: 255}
	skipColor := colorMuted

	for i, b := range v.brews {
		val := v.metric.valueFor(b)
		bx := startX + i*totalBarW
		if bx+barW > graphX+graphW {
			break
		}
		c := barColor
		if i == v.sel {
			c = selColor
		}
		// Draw bar background (empty slot)
		fillRect(screen, bx, graphY+1, barW, bhGraphH-2, colorInput)
		if val < 0 {
			// Not scored — draw a dim tick at bottom
			fillRect(screen, bx, graphY+bhGraphH-3, barW, 2, skipColor)
			continue
		}
		h := int(val / 10.0 * float64(bhGraphH-2))
		if h > 0 {
			fillRect(screen, bx, graphY+bhGraphH-1-h, barW, h, c)
		}
	}

	// Value label for selected brew
	if v.sel < len(v.brews) {
		val := v.metric.valueFor(v.brews[v.sel])
		valStr := "—"
		if val >= 0 {
			valStr = fmt.Sprintf("%.0f/10", val)
		}
		vLabel := fmt.Sprintf("#%d: %s", v.sel+1, valStr)
		ebitenutil.DebugPrintAt(screen, vLabel, bhRightX+2, graphY+bhGraphH+2)
	}
}

func (v *BrewHistoryView) drawBrewDetail(screen *ebiten.Image) {
	if v.sel >= len(v.brews) {
		return
	}
	b := v.brews[v.sel]

	detailY := contentY + lineH + 3 + bhGraphH + lineH + 8
	fillRect(screen, bhRightX, detailY-2, bhRightW, 1, colorBorder)
	detailY += 2

	// Trait bars for selected brew — compact 2 columns
	tr := b.TastingTraits
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
	barColor := color.RGBA{R: 255, G: 200, B: 60, A: 255}
	const labelCols = 6
	colW := bhRightW / 2
	barW := colW - labelCols*6 - 10
	if barW < 8 {
		barW = 8
	}
	for i, t := range traits {
		if t.val < 0 {
			continue
		}
		col := i % 2
		row := i / 2
		if detailY+row*(lineH+1) >= hintsY-lineH {
			break
		}
		rx := bhRightX + col*colW
		ry := detailY + row*(lineH+1)
		drawHBar(screen, rx, ry, labelCols, barW, t.name, float64(t.val), 10, barColor)
	}

	// Tasting notes below the trait grid (left panel, below list)
	notesY := contentY + bhListRows*(lineH+2) + 4
	if notesY < hintsY-lineH*3 {
		fillRect(screen, 0, notesY, bhListW, 1, colorBorder)
		notesY += 4
		var notes []string
		for _, n := range b.TastingNotes {
			if s := strings.TrimSpace(n); s != "" {
				notes = append(notes, s)
			}
		}
		if len(notes) > 0 {
			const noteCPL = bhListW / 6
			ebitenutil.DebugPrintAt(screen, truncate(strings.Join(notes, "  ·  "), noteCPL), 4, notesY)
			notesY += lineH + 2
		}
		// Dripper + time
		dripper := b.Dripper
		if dripper == "" {
			dripper = "no dripper"
		}
		meta := fmt.Sprintf("%s  %d:%02d", dripper, b.EndTime.Minutes, b.EndTime.Seconds)
		if b.DaysOffRoast >= 0 {
			meta += fmt.Sprintf("  %dd off roast", b.DaysOffRoast)
		}
		ebitenutil.DebugPrintAt(screen, truncate(meta, bhListW/6), 4, notesY)
	}
}
