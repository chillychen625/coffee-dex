package game

import (
	"fmt"
	"sort"
	"strings"

	"go-coffee-log/models"
	"go-coffee-log/service"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

// Results returned by PokemonPickerView.Update.
const (
	pickerActive    = 0
	pickerConfirmed = 1
	pickerCanceled  = 2
)

const (
	pickerSpriteSize = 96
	pickerSpriteX    = 36
	pickerSpriteY    = 56
	pickerInfoX      = 170
	pickerNoteTicks  = 180
)

// PokemonPickerView is the trait-curated carousel of unassigned Pokemon.
type PokemonPickerView struct {
	cands     *service.PokemonCandidates
	showAll   bool
	idx       int
	note      string
	noteTicks int
}

func (v *PokemonPickerView) Load(cands *service.PokemonCandidates) {
	v.cands = cands
	v.idx = 0
	v.note = ""
	v.noteTicks = 0
	if len(cands.Curated) == 0 {
		v.showAll = true
		v.note = "no type matches - showing all"
		v.noteTicks = pickerNoteTicks
	} else {
		v.showAll = false
	}
}

func (v *PokemonPickerView) Current() models.Pokemon {
	list := v.active()
	if v.idx >= 0 && v.idx < len(list) {
		return list[v.idx]
	}
	return models.Pokemon{}
}

func (v *PokemonPickerView) active() []models.Pokemon {
	if v.showAll {
		return v.cands.All
	}
	return v.cands.Curated
}

func (v *PokemonPickerView) Update() int {
	if v.noteTicks > 0 {
		v.noteTicks--
		if v.noteTicks == 0 {
			v.note = ""
		}
	}

	list := v.active()
	if len(list) == 0 {
		return pickerActive
	}

	n := len(list)
	if isKeyActive(ebiten.KeyArrowLeft) {
		v.idx = (v.idx - 1 + n) % n
	}
	if isKeyActive(ebiten.KeyArrowRight) {
		v.idx = (v.idx + 1) % n
	}

	if isKeyJustPressed(ebiten.KeyTab) || isKeyJustPressed(ebiten.KeyA) {
		if len(v.cands.Curated) == 0 {
			v.note = "no type matches - showing all"
			v.noteTicks = pickerNoteTicks
		} else {
			v.showAll = !v.showAll
			v.idx = 0
		}
	}

	if isKeyJustPressed(ebiten.KeyEnter) || isKeyJustPressed(ebiten.KeyZ) {
		return pickerConfirmed
	}
	if isKeyJustPressed(ebiten.KeyEscape) {
		return pickerCanceled
	}
	return pickerActive
}

func (v *PokemonPickerView) Draw(screen *ebiten.Image) {
	drawHeader(screen, fmt.Sprintf("Choose - %s", truncate(v.cands.CoffeeName, 28)))
	drawHints(screen, "◄ ► browse   Tab toggle   Enter choose   Esc back")

	list := v.active()
	if len(list) == 0 {
		ebitenutil.DebugPrintAt(screen, "no unassigned Pokemon remain", padding*2, contentY+lineH)
		return
	}
	p := v.Current()

	pos := fmt.Sprintf("[%d/%d]", v.idx+1, len(list))
	ebitenutil.DebugPrintAt(screen, pos, InternalWidth-len(pos)*6-8, padding/2+1)

	mode := "TAB: show all unassigned"
	if v.showAll {
		mode = "TAB: show type matches"
	}
	ebitenutil.DebugPrintAt(screen, mode, padding*2, contentY+2)
	if v.note != "" {
		ebitenutil.DebugPrintAt(screen, v.note, padding*2, contentY+lineH+2)
	}

	DrawSprite(screen, p.ID, pickerSpriteX, pickerSpriteY, pickerSpriteSize, pickerSpriteSize)

	infoY := pickerSpriteY + 4
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("#%03d %s", p.ID, p.Name), pickerInfoX, infoY)
	infoY += lineH + 4

	badgeX := pickerInfoX
	for _, t := range splitTypes(p.Type) {
		w := len(t)*6 + 8
		if v.typeMatchesCoffee(t) {
			strokeRect(screen, badgeX-1, infoY-1, w+2, lineH+4, colorFocused)
		}
		badgeX = drawTypeBadge(screen, t, badgeX, infoY)
	}

	infoY += lineH + 6
	match := v.cands.Match(p)
	if match > 0 {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Match: %d%%", int(match*100)), pickerInfoX, infoY)
	} else {
		ebitenutil.DebugPrintAt(screen, "Match: -", pickerInfoX, infoY)
	}

	profileY := hintsY - lineH - 10
	x := padding * 2
	ebitenutil.DebugPrintAt(screen, "Coffee profile:", x, profileY+1)
	x += len("Coffee profile:")*6 + 8
	x = drawTypeBadge(screen, capitalize(v.cands.PrimaryType), x, profileY)
	if v.cands.SecondaryType != "" {
		drawTypeBadge(screen, capitalize(v.cands.SecondaryType), x, profileY)
	}
}

func (v *PokemonPickerView) typeMatchesCoffee(typeName string) bool {
	return strings.EqualFold(typeName, v.cands.PrimaryType) ||
		strings.EqualFold(typeName, v.cands.SecondaryType)
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// Results returned by DescribeView.Update.
const (
	describeActive = 0
	describeSubmit = 1
	describeBack   = 2
)

const (
	descSpriteSize = 32
	descEditorX    = 216
	descEditorY    = 58
	descEditorW    = 256
	descEditorH    = 4*(lineH+2) + 6
	descConfirmY   = descEditorY + descEditorH + 4
	descNotesTop   = 150
	descNoteRows   = (hintsY - descNotesTop - lineH - 6) / (lineH + 2)
)

type traitBar struct {
	name string
	val  int
}

// DescribeView is the description-writing screen
type DescribeView struct {
	coffeeID   string
	coffeeName string
	pokemon    models.Pokemon
	avgRating  float64
	traits     []traitBar
	noteRows   []string
	noteScroll int
	focus      int // 0 editor, 1 notes, 2 confirm
	editor     MultiLineInput
	msg        string
	msgTicks   int
}

// Load fetches the brew data the screen displays and resets all state.
func (v *DescribeView) Load(svc *Services, coffeeID, coffeeName string, pokemon models.Pokemon) error {
	agg, err := svc.Brew.GetAggregatedData(coffeeID)
	if err != nil {
		return fmt.Errorf("failed to get aggregated brew data: %w", err)
	}
	brews, err := svc.Brew.GetBrewsForCoffee(coffeeID)
	if err != nil {
		return fmt.Errorf("failed to get brews: %w", err)
	}

	v.coffeeID = coffeeID
	v.coffeeName = coffeeName
	v.pokemon = pokemon
	v.avgRating = agg.AverageRating
	v.focus = 0
	v.noteScroll = 0
	v.msg = ""
	v.msgTicks = 0
	v.editor.Clear()

	tr := agg.AverageTraits
	v.traits = []traitBar{
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

	sort.SliceStable(brews, func(i, j int) bool { return brews[i].CreatedAt.After(brews[j].CreatedAt) })
	v.noteRows = make([]string, 0, len(brews))
	for i, b := range brews {
		notes := make([]string, 0, len(b.TastingNotes))
		for _, n := range b.TastingNotes {
			if s := strings.TrimSpace(n); s != "" {
				notes = append(notes, s)
			}
		}
		row := fmt.Sprintf("#%-2d %s ★%-2d %s", len(brews)-i, b.CreatedAt.Format("01/02"), b.Rating, strings.Join(notes, " · "))
		v.noteRows = append(v.noteRows, truncate(row, 76))
	}
	return nil
}

// SetMsg shows a transient message (e.g. a mapping error from the scene).
func (v *DescribeView) SetMsg(msg string) {
	v.msg = msg
	v.msgTicks = 240
}

// Description returns the editor contents for persistence.
func (v *DescribeView) Description() string {
	return v.editor.Text()
}

func (v *DescribeView) Update() int {
	if v.msgTicks > 0 {
		v.msgTicks--
		if v.msgTicks == 0 {
			v.msg = ""
		}
	}

	if isKeyJustPressed(ebiten.KeyEscape) {
		return describeBack
	}
	if isKeyJustPressed(ebiten.KeyTab) {
		v.focus = (v.focus + 1) % 3
	}

	switch v.focus {
	case 0:
		v.editor.Update(true)
	case 1:
		maxScroll := max(0, len(v.noteRows)-descNoteRows)
		if isKeyActive(ebiten.KeyArrowDown) && v.noteScroll < maxScroll {
			v.noteScroll++
		}
		if isKeyActive(ebiten.KeyArrowUp) && v.noteScroll > 0 {
			v.noteScroll--
		}
	case 2:
		if isKeyJustPressed(ebiten.KeyEnter) || isKeyJustPressed(ebiten.KeyZ) {
			return describeSubmit
		}
	}
	return describeActive
}

func (v *DescribeView) Draw(screen *ebiten.Image) {
	drawHeader(screen, fmt.Sprintf("Describe - %s -> %s", truncate(v.coffeeName, 20), truncate(v.pokemon.Name, 12)))
	drawHints(screen, "Tab next field   Enter newline/confirm   Esc back")

	DrawSprite(screen, v.pokemon.ID, 8, contentY+2, descSpriteSize, descSpriteSize)
	infoX := 8 + descSpriteSize + 6
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("#%03d %s", v.pokemon.ID, v.pokemon.Name), infoX, contentY+2)
	badgeY := contentY + lineH + 4
	badgeX := infoX
	for _, t := range splitTypes(v.pokemon.Type) {
		badgeX = drawTypeBadge(screen, t, badgeX, badgeY)
	}
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Avg rating %.1f", v.avgRating), infoX, badgeY+lineH+4)

	barColor := primaryTypeColor(v.pokemon.Type)
	const labelCols = 7
	colW := (descEditorX - 12) / 2
	barW := colW - labelCols*6 - 20
	shown := 0
	for _, t := range v.traits {
		if t.val < 0 {
			continue
		}
		col := shown % 2
		row := shown / 2
		if row >= 6 {
			break
		}
		drawHBar(screen, 8+col*colW, descEditorY+row*(lineH+2), labelCols, barW, t.name, float64(t.val), 10, barColor)
		shown++
	}

	v.editor.Draw(screen, descEditorX, descEditorY, descEditorW, descEditorH, v.focus == 0)

	const btnW = 96
	confirmFocused := v.focus == 2
	if confirmFocused {
		fillRect(screen, descEditorX, descConfirmY, btnW, fieldH, colorSelected)
	} else {
		fillRect(screen, descEditorX, descConfirmY, btnW, fieldH, colorInput)
	}
	btnBorder := colorBorder
	if confirmFocused {
		btnBorder = colorFocused
	}
	strokeRect(screen, descEditorX, descConfirmY, btnW, fieldH, btnBorder)
	ebitenutil.DebugPrintAt(screen, "[ Confirm ]", descEditorX+16, descConfirmY+2)

	if v.msg != "" {
		ebitenutil.DebugPrintAt(screen, truncate(v.msg, 76), descEditorX, descConfirmY+fieldH+4)
	}

	title := fmt.Sprintf("Tasting notes (%d)", len(v.noteRows))
	if v.focus == 1 {
		title = "► " + title
		strokeRect(screen, 0, descNotesTop-2, InternalWidth, hintsY-descNotesTop+2, colorFocused)
	}
	ebitenutil.DebugPrintAt(screen, title, padding*2, descNotesTop)
	for i := 0; i < descNoteRows && v.noteScroll+i < len(v.noteRows); i++ {
		drawListRow(screen, v.noteRows[v.noteScroll+i], 0, descNotesTop+lineH+4+i*(lineH+2), InternalWidth, false)
	}
	if len(v.noteRows) > descNoteRows {
		prog := fmt.Sprintf("%d/%d", min(v.noteScroll+descNoteRows, len(v.noteRows)), len(v.noteRows))
		ebitenutil.DebugPrintAt(screen, prog, InternalWidth-len(prog)*6-8, hintsY-lineH-2)
	}
}
