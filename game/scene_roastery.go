package game

import (
	"fmt"
	"sort"
	"strconv"

	"go-coffee-log/models"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type roastSub int

const (
	roastSubMenu roastSub = iota
	roastSubLogBrew
	roastSubRecentBrews
	roastSubBrewers
	roastSubAddBrewer
)

const (
	brewFieldCoffee  = 0
	brewFieldRating  = 1
	brewFieldDripper = 2
	brewFieldDDMin   = 3
	brewFieldDDSec   = 4
	brewFieldNote0   = 5
	brewFieldNote4   = 9
	brewFieldTrait0  = 10
	brewFieldTrait11 = 21
	brewFieldSubmit  = 22
)

const brewTotalFields = brewFieldSubmit + 1

const (
	addBrewerFieldName   = 0
	addBrewerFieldBall   = 1
	addBrewerFieldSubmit = 2
)

var traitNames = [12]string{
	"Berry", "Stonefruit", "Roast", "Citrus",
	"Bitter", "Floral", "Spice", "Sweet",
	"Aromatic", "Savory", "Body", "Clean",
}

var pokeballTypes = []string{"poke-ball", "great-ball", "ultra-ball", "fast-ball"}
var subMenuItems = []string{"Log Brew", "Recent Brews", "Brewers"}

type RoasteryScene struct {
	svc *Services
	sub roastSub
	sel int

	// Brew form
	coffees   []models.Coffee
	coffeeSel int
	rating    int
	dripper   AutoComplete
	ddMin     int
	ddSec     int
	notes     [5]TextInput
	traits    [12]int // -1 = not scored
	brewFocus  int
	brewScroll int
	brewMsg   string

	// Recent brews
	recentBrews []models.BrewWithCoffee
	brewSel     int
	recentScroll int

	// Brewers list
	brewers   []models.Brewer
	brewerSel int

	// Add brewer form
	brewerName     TextInput
	brewerBallIdx  int
	addBrewerFocus int
	addBrewerMsg   string
}

func NewRoasteryScene() *RoasteryScene { return &RoasteryScene{} }

func (s *RoasteryScene) OnEnter(svc *Services) {
	s.svc = svc
	s.sub = roastSubMenu
	s.sel = 0
	coffees, err := svc.Coffee.ListCoffees()
	if err == nil {
		s.coffees = coffees
	}
	s.loadDrippers()
}

func (s *RoasteryScene) loadDrippers() {
	brews, err := s.svc.Brew.ListBrews()
	if err != nil {
		return
	}
	dripSet := map[string]bool{}
	for _, b := range brews {
		if b.Dripper != "" {
			dripSet[b.Dripper] = true
		}
	}
	drippers := make([]string, 0, len(dripSet))
	for k := range dripSet {
		drippers = append(drippers, k)
	}
	sort.Strings(drippers)
	s.dripper.SetOptions(drippers)
}

func (s *RoasteryScene) loadRecentBrews() {
	brews, err := s.svc.Brew.GetRecentBrewsWithCoffee(50)
	if err == nil {
		s.recentBrews = brews
	}
	s.brewSel = 0
	s.recentScroll = 0
}

func (s *RoasteryScene) loadBrewers() {
	brewers, err := s.svc.Brewer.GetAllBrewers()
	if err == nil {
		s.brewers = brewers
	}
	s.brewerSel = 0
}

func (s *RoasteryScene) resetBrewForm() {
	s.coffeeSel = 0
	s.rating = 7
	s.dripper.Clear()
	s.ddMin = 0
	s.ddSec = 0
	for i := range s.notes {
		s.notes[i].Clear()
	}
	for i := range s.traits {
		s.traits[i] = -1
	}
	s.brewFocus = 0
	s.brewScroll = 0
	s.brewMsg = ""
}

func (s *RoasteryScene) resetAddBrewerForm() {
	s.brewerName.Clear()
	s.brewerBallIdx = 0
	s.addBrewerFocus = 0
	s.addBrewerMsg = ""
}

func (s *RoasteryScene) scrollBrewToFocus() {
	visible := brewVisible()
	if s.brewFocus < s.brewScroll {
		s.brewScroll = s.brewFocus
	}
	if s.brewFocus >= s.brewScroll+visible {
		s.brewScroll = s.brewFocus - visible + 1
	}
}

func brewVisible() int {
	return (hintsY - contentY - 4) / rowStep
}

func (s *RoasteryScene) submitBrew() string {
	if len(s.coffees) == 0 {
		return "No coffees — add one in Bean Warehouse first"
	}
	coffee := s.coffees[s.coffeeSel]
	var notes [5]string
	for i := range s.notes {
		notes[i] = s.notes[i].Value
	}
	brew := models.Brew{
		CoffeeID:     coffee.ID,
		Rating:       s.rating,
		Dripper:      s.dripper.Value,
		EndTime:      models.DrawDownTime{Minutes: s.ddMin, Seconds: s.ddSec},
		TastingNotes: notes,
		TastingTraits: models.TastingTraits{
			BerryIntensity:        s.traits[0],
			StonefruitIntensity:   s.traits[1],
			RoastIntensity:        s.traits[2],
			CitrusFruitsIntensity: s.traits[3],
			Bitterness:            s.traits[4],
			Florality:             s.traits[5],
			Spice:                 s.traits[6],
			Sweetness:             s.traits[7],
			AromaticIntensity:     s.traits[8],
			Savory:                s.traits[9],
			Body:                  s.traits[10],
			Cleanliness:           s.traits[11],
		},
	}
	_, err := s.svc.Brew.CreateBrew(brew)
	if err != nil {
		return "Error: " + err.Error()
	}
	return "ok"
}

func (s *RoasteryScene) submitAddBrewer() string {
	if s.brewerName.Value == "" {
		return "Name is required"
	}
	_, err := s.svc.Brewer.CreateBrewer(s.brewerName.Value, pokeballTypes[s.brewerBallIdx])
	if err != nil {
		return "Error: " + err.Error()
	}
	return "ok"
}

func (s *RoasteryScene) Update() SceneID {
	switch s.sub {
	case roastSubMenu:
		return s.updateSubMenu()
	case roastSubLogBrew:
		return s.updateLogBrew()
	case roastSubRecentBrews:
		return s.updateRecentBrews()
	case roastSubBrewers:
		return s.updateBrewers()
	case roastSubAddBrewer:
		return s.updateAddBrewer()
	}
	return SceneRoastery
}

func (s *RoasteryScene) updateSubMenu() SceneID {
	if isKeyJustPressed(ebiten.KeyEscape) {
		return SceneMenu
	}
	if isKeyActive(ebiten.KeyArrowDown) && s.sel < len(subMenuItems)-1 {
		s.sel++
	}
	if isKeyActive(ebiten.KeyArrowUp) && s.sel > 0 {
		s.sel--
	}
	if isKeyJustPressed(ebiten.KeyEnter) || isKeyJustPressed(ebiten.KeyZ) {
		switch s.sel {
		case 0:
			s.resetBrewForm()
			s.coffees, _ = s.svc.Coffee.ListCoffees()
			s.sub = roastSubLogBrew
		case 1:
			s.loadRecentBrews()
			s.sub = roastSubRecentBrews
		case 2:
			s.loadBrewers()
			s.sub = roastSubBrewers
		}
	}
	return SceneRoastery
}

func (s *RoasteryScene) updateLogBrew() SceneID {
	navConsumed := false

	// Field-specific input (runs before navigation so autocomplete can consume keys)
	switch s.brewFocus {
	case brewFieldDripper:
		navConsumed = s.dripper.Update(true)
	case brewFieldNote0, brewFieldNote0 + 1, brewFieldNote0 + 2, brewFieldNote0 + 3, brewFieldNote0 + 4:
		s.notes[s.brewFocus-brewFieldNote0].Update(true)
	}

	if navConsumed {
		return SceneRoastery
	}

	if isKeyJustPressed(ebiten.KeyEscape) {
		s.sub = roastSubMenu
		return SceneRoastery
	}

	// Navigation
	if isKeyActive(ebiten.KeyArrowDown) && s.brewFocus < brewFieldSubmit {
		s.brewFocus++
		s.scrollBrewToFocus()
	}
	if isKeyActive(ebiten.KeyArrowUp) && s.brewFocus > 0 {
		s.brewFocus--
		s.scrollBrewToFocus()
	}
	if isKeyJustPressed(ebiten.KeyTab) {
		s.brewFocus = (s.brewFocus + 1) % brewTotalFields
		s.scrollBrewToFocus()
	}

	// Field value changes
	switch s.brewFocus {
	case brewFieldCoffee:
		if isKeyActive(ebiten.KeyArrowLeft) && s.coffeeSel > 0 {
			s.coffeeSel--
		}
		if isKeyActive(ebiten.KeyArrowRight) && s.coffeeSel < len(s.coffees)-1 {
			s.coffeeSel++
		}
	case brewFieldRating:
		if isKeyActive(ebiten.KeyArrowLeft) && s.rating > 0 {
			s.rating--
		}
		if isKeyActive(ebiten.KeyArrowRight) && s.rating < 10 {
			s.rating++
		}
		var chars []rune
		chars = ebiten.AppendInputChars(chars)
		for _, ch := range chars {
			if ch >= '0' && ch <= '9' {
				s.rating = int(ch - '0')
			}
		}
	case brewFieldDDMin:
		if isKeyActive(ebiten.KeyArrowLeft) && s.ddMin > 0 {
			s.ddMin--
		}
		if isKeyActive(ebiten.KeyArrowRight) && s.ddMin < 99 {
			s.ddMin++
		}
	case brewFieldDDSec:
		if isKeyActive(ebiten.KeyArrowLeft) && s.ddSec > 0 {
			s.ddSec--
		}
		if isKeyActive(ebiten.KeyArrowRight) && s.ddSec < 59 {
			s.ddSec++
		}
	default:
		if s.brewFocus >= brewFieldTrait0 && s.brewFocus <= brewFieldTrait11 {
			ti := s.brewFocus - brewFieldTrait0
			if isKeyActive(ebiten.KeyArrowLeft) {
				if s.traits[ti] > -1 {
					s.traits[ti]--
				}
			}
			if isKeyActive(ebiten.KeyArrowRight) {
				if s.traits[ti] < 10 {
					s.traits[ti]++
				}
			}
			// Type digit directly
			var chars []rune
			chars = ebiten.AppendInputChars(chars)
			for _, ch := range chars {
				if ch >= '0' && ch <= '9' {
					s.traits[ti] = int(ch - '0')
				}
			}
			// Minus key or - to skip (set -1)
			if isKeyJustPressed(ebiten.KeyMinus) {
				s.traits[ti] = -1
			}
		}
	case brewFieldSubmit:
		if isKeyJustPressed(ebiten.KeyEnter) || isKeyJustPressed(ebiten.KeyZ) {
			msg := s.submitBrew()
			if msg == "ok" {
				s.resetBrewForm()
				s.brewMsg = "Brew logged!"
				s.sub = roastSubMenu
			} else {
				s.brewMsg = msg
			}
		}
	}

	if isKeyJustPressed(ebiten.KeyEnter) && s.brewFocus < brewFieldSubmit {
		s.brewFocus++
		s.scrollBrewToFocus()
	}

	return SceneRoastery
}

func (s *RoasteryScene) updateRecentBrews() SceneID {
	if isKeyJustPressed(ebiten.KeyEscape) {
		s.sub = roastSubMenu
		return SceneRoastery
	}
	visible := (hintsY - contentY) / (lineH + 2)
	if isKeyActive(ebiten.KeyArrowDown) && s.brewSel < len(s.recentBrews)-1 {
		s.brewSel++
		if s.brewSel >= s.recentScroll+visible {
			s.recentScroll++
		}
	}
	if isKeyActive(ebiten.KeyArrowUp) && s.brewSel > 0 {
		s.brewSel--
		if s.brewSel < s.recentScroll {
			s.recentScroll--
		}
	}
	return SceneRoastery
}

func (s *RoasteryScene) updateBrewers() SceneID {
	if isKeyJustPressed(ebiten.KeyEscape) {
		s.sub = roastSubMenu
		return SceneRoastery
	}
	total := len(s.brewers) + 1
	if isKeyActive(ebiten.KeyArrowDown) && s.brewerSel < total-1 {
		s.brewerSel++
	}
	if isKeyActive(ebiten.KeyArrowUp) && s.brewerSel > 0 {
		s.brewerSel--
	}
	if isKeyJustPressed(ebiten.KeyEnter) || isKeyJustPressed(ebiten.KeyZ) {
		if s.brewerSel == len(s.brewers) {
			s.resetAddBrewerForm()
			s.sub = roastSubAddBrewer
		}
	}
	return SceneRoastery
}

func (s *RoasteryScene) updateAddBrewer() SceneID {
	if isKeyJustPressed(ebiten.KeyEscape) {
		s.sub = roastSubBrewers
		return SceneRoastery
	}
	if isKeyActive(ebiten.KeyArrowDown) && s.addBrewerFocus < addBrewerFieldSubmit {
		s.addBrewerFocus++
	}
	if isKeyActive(ebiten.KeyArrowUp) && s.addBrewerFocus > 0 {
		s.addBrewerFocus--
	}
	if isKeyJustPressed(ebiten.KeyTab) {
		s.addBrewerFocus = (s.addBrewerFocus + 1) % (addBrewerFieldSubmit + 1)
	}

	switch s.addBrewerFocus {
	case addBrewerFieldName:
		s.brewerName.Update(true)
	case addBrewerFieldBall:
		if isKeyActive(ebiten.KeyArrowLeft) && s.brewerBallIdx > 0 {
			s.brewerBallIdx--
		}
		if isKeyActive(ebiten.KeyArrowRight) && s.brewerBallIdx < len(pokeballTypes)-1 {
			s.brewerBallIdx++
		}
	case addBrewerFieldSubmit:
		if isKeyJustPressed(ebiten.KeyEnter) || isKeyJustPressed(ebiten.KeyZ) {
			msg := s.submitAddBrewer()
			if msg == "ok" {
				s.loadBrewers()
				s.sub = roastSubBrewers
			} else {
				s.addBrewerMsg = msg
			}
		}
	}

	if isKeyJustPressed(ebiten.KeyEnter) && s.addBrewerFocus < addBrewerFieldSubmit {
		s.addBrewerFocus++
	}

	return SceneRoastery
}

// ── Draw ─────────────────────────────────────────────────────────────────────

func (s *RoasteryScene) Draw(screen *ebiten.Image) {
	drawBackground(screen)
	switch s.sub {
	case roastSubMenu:
		s.drawSubMenu(screen)
	case roastSubLogBrew:
		s.drawLogBrew(screen)
	case roastSubRecentBrews:
		s.drawRecentBrews(screen)
	case roastSubBrewers:
		s.drawBrewers(screen)
	case roastSubAddBrewer:
		s.drawAddBrewer(screen)
	}
}

func (s *RoasteryScene) drawSubMenu(screen *ebiten.Image) {
	drawHeader(screen, "Roastery")
	if s.brewMsg != "" {
		ebitenutil.DebugPrintAt(screen, s.brewMsg, 10, contentY+4)
	}
	for i, label := range subMenuItems {
		y := contentY + 20 + i*22
		drawListRow(screen, label, 40, y, InternalWidth-80, i == s.sel)
	}
	drawHints(screen, "[↑↓] Select   [Enter] Open   [Esc] Back")
}

func (s *RoasteryScene) drawLogBrew(screen *ebiten.Image) {
	drawHeader(screen, "Roastery — Log Brew")

	// Each field draws at its virtual y, shifted by scroll
	fieldY := func(idx int) int {
		return contentY + 4 + (idx-s.brewScroll)*rowStep
	}
	inView := func(idx int) bool {
		y := fieldY(idx)
		return y >= contentY && y+fieldH <= hintsY
	}

	// Coffee
	if inView(brewFieldCoffee) {
		coffeeName := "(no coffees)"
		if len(s.coffees) > 0 {
			coffeeName = fmt.Sprintf("◄ %s ►", truncate(s.coffees[s.coffeeSel].Name, 28))
		}
		drawFieldRow(screen, "Coffee:", coffeeName, fieldY(brewFieldCoffee), s.brewFocus == brewFieldCoffee)
	}

	// Rating
	if inView(brewFieldRating) {
		drawFieldRow(screen, "Rating:", fmt.Sprintf("◄ %d ►", s.rating), fieldY(brewFieldRating), s.brewFocus == brewFieldRating)
	}

	// Dripper (autocomplete)
	if inView(brewFieldDripper) {
		y := fieldY(brewFieldDripper)
		ebitenutil.DebugPrintAt(screen, "Dripper:", 10, y+2)
		s.dripper.Draw(screen, fieldX, y, fieldW, s.brewFocus == brewFieldDripper)
	}

	// Draw Down (combined row)
	if inView(brewFieldDDMin) {
		y := fieldY(brewFieldDDMin)
		ddFocused := s.brewFocus == brewFieldDDMin || s.brewFocus == brewFieldDDSec
		fillRect(screen, 8, y, InternalWidth-16, fieldH, colorInput)
		bc := colorBorder
		if ddFocused {
			bc = colorFocused
		}
		strokeRect(screen, 8, y, InternalWidth-16, fieldH, bc)
		ebitenutil.DebugPrintAt(screen, "Drawdown:", 10, y+2)
		if s.brewFocus == brewFieldDDMin {
			fillRect(screen, fieldX, y+1, 60, fieldH-2, colorSelected)
		}
		if s.brewFocus == brewFieldDDSec {
			fillRect(screen, fieldX+70, y+1, 60, fieldH-2, colorSelected)
		}
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("◄ %dm ►", s.ddMin), fieldX+2, y+2)
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("◄ %02ds ►", s.ddSec), fieldX+72, y+2)
	}
	// ddSec shares the same row — skip drawing a separate row for it
	// (inView check for ddSec would show a duplicate; we handle it above)

	// Notes
	for i := 0; i < 5; i++ {
		fi := brewFieldNote0 + i
		if inView(fi) {
			y := fieldY(fi)
			ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Note %d:", i+1), 10, y+2)
			s.notes[i].Draw(screen, fieldX, y, fieldW, s.brewFocus == fi)
		}
	}

	// Traits
	for i := 0; i < 12; i++ {
		fi := brewFieldTrait0 + i
		if inView(fi) {
			y := fieldY(fi)
			val := "-"
			if s.traits[i] >= 0 {
				val = strconv.Itoa(s.traits[i])
			}
			drawFieldRow(screen, traitNames[i]+":", fmt.Sprintf("◄ %2s ► (- to skip)", val), y, s.brewFocus == fi)
		}
	}

	// Submit
	if inView(brewFieldSubmit) {
		y := fieldY(brewFieldSubmit) + 2
		if s.brewFocus == brewFieldSubmit {
			fillRect(screen, InternalWidth/2-40, y-1, 80, lineH+4, colorSelected)
			strokeRect(screen, InternalWidth/2-40, y-1, 80, lineH+4, colorBorder)
		}
		ebitenutil.DebugPrintAt(screen, "[ Submit ]", InternalWidth/2-30, y)
	}

	// Scroll indicator
	visible := brewVisible()
	if brewTotalFields > visible {
		prog := fmt.Sprintf("%d/%d", s.brewFocus+1, brewTotalFields)
		ebitenutil.DebugPrintAt(screen, prog, InternalWidth-len(prog)*6-8, hintsY-12)
	}

	drawHints(screen, "[↑↓] Navigate   [←→] Change   [Tab] Next   [- ] Skip trait   [Esc] Back")

	// Draw dropdown last so it appears on top of all other content
	s.dripper.DrawDropdown(screen)
}

func (s *RoasteryScene) drawRecentBrews(screen *ebiten.Image) {
	drawHeader(screen, "Roastery — Recent Brews")
	if len(s.recentBrews) == 0 {
		ebitenutil.DebugPrintAt(screen, "No brews logged yet.", 10, contentY+10)
		drawHints(screen, "[Esc] Back")
		return
	}

	visible := (hintsY - contentY) / (lineH + 2)
	for i := 0; i < visible && s.recentScroll+i < len(s.recentBrews); i++ {
		b := s.recentBrews[s.recentScroll+i]
		row := fmt.Sprintf("%-22s  ★%-2d  %-10s  %s",
			truncate(b.CoffeeName, 22), b.Rating,
			truncate(b.Dripper, 10), b.CreatedAt.Format("01/02/06"))
		drawListRow(screen, row, 0, contentY+i*(lineH+2), InternalWidth, s.recentScroll+i == s.brewSel)
	}

	if len(s.recentBrews) > visible {
		prog := strconv.Itoa(s.brewSel+1) + "/" + strconv.Itoa(len(s.recentBrews))
		ebitenutil.DebugPrintAt(screen, prog, InternalWidth-len(prog)*6-8, hintsY-12)
	}
	drawHints(screen, "[↑↓] Scroll   [Esc] Back")
}

func (s *RoasteryScene) drawBrewers(screen *ebiten.Image) {
	drawHeader(screen, "Roastery — Brewers")
	if len(s.brewers) == 0 {
		ebitenutil.DebugPrintAt(screen, "No brewers yet.", 10, contentY+10)
	}
	for i, b := range s.brewers {
		y := contentY + i*(lineH+2)
		row := fmt.Sprintf("%-24s  %s", truncate(b.Name, 24), b.PokeballType)
		drawListRow(screen, row, 0, y, InternalWidth, i == s.brewerSel)
	}
	addY := contentY + len(s.brewers)*(lineH+2)
	drawListRow(screen, "+ Add Brewer", 0, addY, InternalWidth, s.brewerSel == len(s.brewers))
	drawHints(screen, "[↑↓] Select   [Enter] Add   [Esc] Back")
}

func (s *RoasteryScene) drawAddBrewer(screen *ebiten.Image) {
	drawHeader(screen, "Roastery — Add Brewer")
	y := contentY + 8

	ebitenutil.DebugPrintAt(screen, "Name:", 10, y+2)
	s.brewerName.Draw(screen, fieldX, y, fieldW, s.addBrewerFocus == addBrewerFieldName)
	y += rowStep

	drawFieldRow(screen, "Pokeball:", fmt.Sprintf("◄ %s ►", pokeballTypes[s.brewerBallIdx]), y, s.addBrewerFocus == addBrewerFieldBall)
	y += rowStep + 4

	if s.addBrewerFocus == addBrewerFieldSubmit {
		fillRect(screen, InternalWidth/2-40, y-1, 80, lineH+4, colorSelected)
		strokeRect(screen, InternalWidth/2-40, y-1, 80, lineH+4, colorBorder)
	}
	ebitenutil.DebugPrintAt(screen, "[ Submit ]", InternalWidth/2-30, y)

	if s.addBrewerMsg != "" {
		ebitenutil.DebugPrintAt(screen, s.addBrewerMsg, 10, y+20)
	}

	drawHints(screen, "[↑↓] Navigate   [←→] Change   [Tab] Next   [Esc] Back")
}
