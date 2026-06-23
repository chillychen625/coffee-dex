package game

import (
	"fmt"
	"image/color"
	"sort"
	"time"

	"go-coffee-log/models"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type warehouseSub int

const (
	warehouseSubList warehouseSub = iota
	warehouseSubAdd
	warehouseSubActions    // sub-menu for a selected coffee
	warehouseSubConfirm    // confirm close-bag
	warehouseSubGenerating // pokemon generation in progress / result
)

const (
	coffeeFieldName       = 0
	coffeeFieldOrigin     = 1
	coffeeFieldRoaster    = 2
	coffeeFieldVariety    = 3
	coffeeFieldRoastLevel = 4
	coffeeFieldProcessing = 5
	coffeeFieldRoastDate  = 6
	coffeeFieldSubmit     = 7
)

var roastLevels = []string{"light", "medium", "dark", "light medium", "medium dark", "unclear", ""}
var processingMethods = []string{"washed", "natural", "honey", "coferment", "experimental", ""}

type genResult struct {
	pokemon *models.CoffeePokemon
	err     error
}

type WarehouseScene struct {
	svc *Services
	sub warehouseSub

	// List (sel=0 → Add Coffee; sel=i+1 → coffees[i])
	sel            int
	coffees        []models.Coffee // open bags only
	coffeeScroll   int
	lastBrewDates  map[string]time.Time

	// Add form
	name          TextInput
	origin        AutoComplete
	roaster       AutoComplete
	variety       AutoComplete
	roastLevelIdx int
	processingIdx int
	roastDate     DateInput
	coffeeFocus   int
	addMsg        string

	// Actions for selected coffee
	actionSel    int
	hasPokemon   bool
	brewProgress models.BrewProgress

	// Confirm close bag
	confirmSel int // 0=yes, 1=no

	// Pokemon generation
	generating     bool
	genChan        chan genResult
	genMsg         string
	genPokemon     *models.CoffeePokemon
	genCoffeeName  string

	// Encounter animation phases (active when !generating && genPokemon != nil)
	// 0=white-flash, 1=text-type, 2=sprite-slide, 3=result
	encTick  int
	encPhase int
}

func NewWarehouseScene() *WarehouseScene { return &WarehouseScene{} }

func (s *WarehouseScene) OnEnter(svc *Services) {
	s.svc = svc
	s.sub = warehouseSubList
	s.sel = 0
	s.loadCoffees()
}

func (s *WarehouseScene) loadCoffees() {
	all, err := s.svc.Coffee.ListCoffees()
	if err != nil {
		return
	}
	// Build set of coffees that already have a Pokemon assigned
	allPokemon, _ := s.svc.Pokemon.GetAllCoffeePokemon()
	withPokemon := make(map[string]bool, len(allPokemon))
	for _, p := range allPokemon {
		withPokemon[p.CoffeeID] = true
	}
	// Show all coffees without a Pokemon yet — open and closed bags both
	s.coffees = nil
	for _, c := range all {
		if !withPokemon[c.ID] {
			s.coffees = append(s.coffees, c)
		}
	}
	// Build autocomplete sets
	roasterSet := map[string]bool{}
	originSet := map[string]bool{}
	varietySet := map[string]bool{}
	for _, c := range s.coffees {
		if c.Roaster != "" {
			roasterSet[c.Roaster] = true
		}
		if c.Origin != "" {
			originSet[c.Origin] = true
		}
		if c.Variety != "" {
			varietySet[c.Variety] = true
		}
	}
	s.roaster.SetOptions(sortedKeys(roasterSet))
	s.origin.SetOptions(sortedKeys(originSet))
	s.variety.SetOptions(sortedKeys(varietySet))

	// Load last brew dates for recency display.
	if dates, err := s.svc.Brew.GetLastBrewDates(); err == nil {
		s.lastBrewDates = dates
	}
}

func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func (s *WarehouseScene) selectedCoffee() *models.Coffee {
	if s.sel == 0 || s.sel > len(s.coffees) {
		return nil
	}
	c := s.coffees[s.sel-1]
	return &c
}

func (s *WarehouseScene) loadCoffeeActions() {
	c := s.selectedCoffee()
	if c == nil {
		return
	}
	s.hasPokemon = s.svc.Pokemon.HasPokemon(c.ID)
	progress, err := s.svc.Brew.GetBrewProgress(c.ID, s.hasPokemon, c.IsFinished)
	if err == nil {
		s.brewProgress = progress
	}
	s.actionSel = 0
	s.genMsg = ""
	s.genPokemon = nil
}

func (s *WarehouseScene) actionItems() []string {
	var items []string
	if c := s.selectedCoffee(); c != nil && !c.IsFinished {
		items = append(items, "Close Bag")
	}
	if s.brewProgress.CanGeneratePokemon && !s.hasPokemon {
		items = append(items, "Generate Pokemon  [AI]")
	}
	items = append(items, "Back")
	return items
}

func (s *WarehouseScene) resetAddForm() {
	s.name.Clear()
	s.origin.Clear()
	s.roaster.Clear()
	s.variety.Clear()
	s.roastLevelIdx = 0
	s.processingIdx = 0
	s.roastDate.Clear()
	s.coffeeFocus = 0
	s.addMsg = ""
	s.loadCoffees()
}

func (s *WarehouseScene) submitCoffee() string {
	if s.name.Value == "" {
		return "Name is required"
	}
	c := models.Coffee{
		Name:             s.name.Value,
		Origin:           s.origin.Value,
		Roaster:          s.roaster.Value,
		Variety:          s.variety.Value,
		RoastLevel:       roastLevels[s.roastLevelIdx],
		ProcessingMethod: processingMethods[s.processingIdx],
	}
	if vs := s.roastDate.ValueString(); vs != "" {
		var d models.DateOnly
		if err := d.UnmarshalJSON([]byte(`"` + vs + `"`)); err != nil {
			return "Invalid date — use YYYY-MM-DD"
		}
		c.RoastDate = &d
	}
	_, err := s.svc.Coffee.CreateCoffee(c)
	if err != nil {
		return "Error: " + err.Error()
	}
	return "ok"
}

// ── Update ────────────────────────────────────────────────────────────────────

func (s *WarehouseScene) Update() SceneID {
	switch s.sub {
	case warehouseSubList:
		return s.updateList()
	case warehouseSubAdd:
		return s.updateAdd()
	case warehouseSubActions:
		return s.updateActions()
	case warehouseSubConfirm:
		return s.updateConfirm()
	case warehouseSubGenerating:
		return s.updateGenerating()
	}
	return SceneWarehouse
}

func (s *WarehouseScene) updateList() SceneID {
	if isKeyJustPressed(ebiten.KeyEscape) {
		return SceneMenu
	}
	total := len(s.coffees) + 1
	visible := (hintsY - contentY) / (lineH + 2)

	if isKeyActive(ebiten.KeyArrowDown) && s.sel < total-1 {
		s.sel++
		if s.sel > s.coffeeScroll+visible-1 {
			s.coffeeScroll++
		}
	}
	if isKeyActive(ebiten.KeyArrowUp) && s.sel > 0 {
		s.sel--
		if s.sel < s.coffeeScroll {
			s.coffeeScroll--
		}
	}
	if isKeyJustPressed(ebiten.KeyEnter) || isKeyJustPressed(ebiten.KeyZ) {
		if s.sel == 0 {
			s.resetAddForm()
			s.sub = warehouseSubAdd
		} else {
			s.loadCoffeeActions()
			s.sub = warehouseSubActions
		}
	}
	return SceneWarehouse
}

func (s *WarehouseScene) updateActions() SceneID {
	if isKeyJustPressed(ebiten.KeyEscape) {
		s.sub = warehouseSubList
		return SceneWarehouse
	}
	items := s.actionItems()
	if isKeyJustPressed(ebiten.KeyArrowDown) && s.actionSel < len(items)-1 {
		s.actionSel++
	}
	if isKeyJustPressed(ebiten.KeyArrowUp) && s.actionSel > 0 {
		s.actionSel--
	}
	if isKeyJustPressed(ebiten.KeyEnter) || isKeyJustPressed(ebiten.KeyZ) {
		chosen := items[s.actionSel]
		switch chosen {
		case "Close Bag":
			s.confirmSel = 1 // default to "No"
			s.sub = warehouseSubConfirm
		case "Back":
			s.sub = warehouseSubList
		default: // Generate Pokemon
			s.startGeneration()
		}
	}
	return SceneWarehouse
}

func (s *WarehouseScene) updateConfirm() SceneID {
	if isKeyJustPressed(ebiten.KeyEscape) || isKeyJustPressed(ebiten.KeyN) {
		s.sub = warehouseSubActions
		return SceneWarehouse
	}
	if isKeyJustPressed(ebiten.KeyArrowLeft) {
		s.confirmSel = 0
	}
	if isKeyJustPressed(ebiten.KeyArrowRight) {
		s.confirmSel = 1
	}
	if isKeyJustPressed(ebiten.KeyY) {
		s.doCloseBag()
		return SceneWarehouse
	}
	if isKeyJustPressed(ebiten.KeyEnter) || isKeyJustPressed(ebiten.KeyZ) {
		if s.confirmSel == 0 {
			s.doCloseBag()
		} else {
			s.sub = warehouseSubActions
		}
	}
	return SceneWarehouse
}

func (s *WarehouseScene) doCloseBag() {
	c := s.selectedCoffee()
	if c == nil {
		return
	}
	_, err := s.svc.Coffee.MarkAsFinished(c.ID)
	if err != nil {
		s.genMsg = "Error: " + err.Error()
		s.sub = warehouseSubGenerating // reuse result screen
		return
	}
	s.loadCoffees()
	s.sel = 0
	s.sub = warehouseSubList
}

func (s *WarehouseScene) startGeneration() {
	c := s.selectedCoffee()
	if c == nil {
		return
	}
	s.generating = true
	s.genMsg = ""
	s.genPokemon = nil
	s.genCoffeeName = c.Name
	s.encTick = 0
	s.encPhase = 0
	s.genChan = make(chan genResult, 1)
	coffeeID := c.ID
	go func() {
		p, err := s.svc.Pokemon.MapCoffeeToPokemon(coffeeID)
		s.genChan <- genResult{pokemon: p, err: err}
	}()
	s.sub = warehouseSubGenerating
}

func (s *WarehouseScene) updateGenerating() SceneID {
	s.encTick++

	// Poll goroutine while loading.
	if s.generating {
		select {
		case result := <-s.genChan:
			s.generating = false
			if result.err != nil {
				s.genMsg = result.err.Error()
			} else {
				s.genPokemon = result.pokemon
				s.genMsg = ""
				s.loadCoffees()
				// Reset tick to drive the encounter animation.
				s.encTick = 0
				s.encPhase = 0
			}
		default:
		}
		return SceneWarehouse
	}

	// Error state: any key exits.
	if s.genMsg != "" {
		if isKeyJustPressed(ebiten.KeyEnter) || isKeyJustPressed(ebiten.KeyEscape) ||
			isKeyJustPressed(ebiten.KeyZ) || isKeyJustPressed(ebiten.KeySpace) {
			s.sub = warehouseSubList
			s.sel = 0
		}
		return SceneWarehouse
	}

	// Encounter animation state machine.
	p := s.genPokemon
	const (
		flashDur  = 24
		slideLen  = 40
	)
	switch s.encPhase {
	case 0: // white flash fades out
		if s.encTick >= flashDur {
			s.encPhase = 1
			s.encTick = 0
		}
	case 1: // text types in
		fullText := "A wild " + p.PokemonName + " appeared!"
		if s.encTick >= len([]rune(fullText))*3 {
			s.encPhase = 2
			s.encTick = 0
		}
	case 2: // sprite slides in
		if s.encTick >= slideLen+10 {
			s.encPhase = 3
			s.encTick = 0
		}
	case 3: // result — wait for any key
		if isKeyJustPressed(ebiten.KeyEnter) || isKeyJustPressed(ebiten.KeyEscape) ||
			isKeyJustPressed(ebiten.KeyZ) || isKeyJustPressed(ebiten.KeySpace) {
			s.sub = warehouseSubList
			s.sel = 0
		}
	}
	return SceneWarehouse
}

func (s *WarehouseScene) updateAdd() SceneID {
	navConsumed := false

	switch s.coffeeFocus {
	case coffeeFieldName:
		s.name.Update(true)
	case coffeeFieldOrigin:
		navConsumed = s.origin.Update(true)
	case coffeeFieldRoaster:
		navConsumed = s.roaster.Update(true)
	case coffeeFieldVariety:
		navConsumed = s.variety.Update(true)
	case coffeeFieldRoastDate:
		s.roastDate.Update(true)
		if isKeyJustPressed(ebiten.KeyTab) && s.roastDate.seg < 2 {
			return SceneWarehouse
		}
	}

	if navConsumed {
		return SceneWarehouse
	}

	if isKeyJustPressed(ebiten.KeyEscape) {
		s.sub = warehouseSubList
		return SceneWarehouse
	}

	if isKeyJustPressed(ebiten.KeyArrowDown) && s.coffeeFocus < coffeeFieldSubmit {
		s.coffeeFocus++
	}
	if isKeyJustPressed(ebiten.KeyArrowUp) && s.coffeeFocus > 0 {
		s.coffeeFocus--
	}
	if isKeyJustPressed(ebiten.KeyTab) {
		s.coffeeFocus = (s.coffeeFocus + 1) % (coffeeFieldSubmit + 1)
	}

	switch s.coffeeFocus {
	case coffeeFieldRoastLevel:
		if isKeyJustPressed(ebiten.KeyArrowLeft) && s.roastLevelIdx > 0 {
			s.roastLevelIdx--
		}
		if isKeyJustPressed(ebiten.KeyArrowRight) && s.roastLevelIdx < len(roastLevels)-1 {
			s.roastLevelIdx++
		}
	case coffeeFieldProcessing:
		if isKeyJustPressed(ebiten.KeyArrowLeft) && s.processingIdx > 0 {
			s.processingIdx--
		}
		if isKeyJustPressed(ebiten.KeyArrowRight) && s.processingIdx < len(processingMethods)-1 {
			s.processingIdx++
		}
	case coffeeFieldSubmit:
		if isKeyJustPressed(ebiten.KeyEnter) || isKeyJustPressed(ebiten.KeyZ) {
			msg := s.submitCoffee()
			if msg == "ok" {
				s.loadCoffees()
				s.sub = warehouseSubList
				s.sel = len(s.coffees)
			} else {
				s.addMsg = msg
			}
		}
	}

	if isKeyJustPressed(ebiten.KeyEnter) && s.coffeeFocus < coffeeFieldSubmit {
		s.coffeeFocus++
	}

	return SceneWarehouse
}

// ── Draw ─────────────────────────────────────────────────────────────────────

func (s *WarehouseScene) Draw(screen *ebiten.Image) {
	drawBackground(screen)
	switch s.sub {
	case warehouseSubList:
		s.drawList(screen)
	case warehouseSubAdd:
		s.drawAdd(screen)
	case warehouseSubActions:
		s.drawActions(screen)
	case warehouseSubConfirm:
		s.drawConfirm(screen)
	case warehouseSubGenerating:
		s.drawGenerating(screen)
	}
}

func (s *WarehouseScene) drawList(screen *ebiten.Image) {
	drawHeader(screen, "Bean Warehouse — Unassigned")
	visible := (hintsY - contentY) / (lineH + 2)
	y := contentY

	if s.coffeeScroll == 0 {
		drawListRow(screen, "+ Add Coffee", 0, y, InternalWidth, s.sel == 0)
		y += lineH + 2
	}

	for i := 0; i < visible && s.coffeeScroll+i < len(s.coffees); i++ {
		c := s.coffees[s.coffeeScroll+i]
		origin := c.Origin
		if origin == "" {
			origin = "—"
		}
		roaster := c.Roaster
		if roaster == "" {
			roaster = "—"
		}
		// Brew recency label.
		recency := "no brews"
		if t, ok := s.lastBrewDates[c.ID]; ok && !t.IsZero() {
			days := int(time.Since(t).Hours() / 24)
			if days == 0 {
				recency = "today"
			} else {
				recency = fmt.Sprintf("%dd ago", days)
			}
		}
		closed := ""
		if c.IsFinished {
			closed = "[X]"
		}
		row := fmt.Sprintf("%-20s  %-12s  %-10s  %-8s %s",
			truncate(c.Name, 20), truncate(roaster, 12), truncate(origin, 10), recency, closed)
		drawListRow(screen, row, 0, y, InternalWidth, s.sel == i+1+s.coffeeScroll)
		y += lineH + 2
	}

	if len(s.coffees) == 0 {
		ebitenutil.DebugPrintAt(screen, "No open bags.", 10, contentY+18)
	}
	drawHints(screen, "[↑↓] Scroll   [Enter] Select   [Esc] Back")
}

func (s *WarehouseScene) drawActions(screen *ebiten.Image) {
	c := s.selectedCoffee()
	if c == nil {
		return
	}
	drawHeader(screen, "Bean Warehouse — "+truncate(c.Name, 36))
	y := contentY + 6

	// Coffee summary
	if c.Origin != "" || c.Roaster != "" {
		line := ""
		if c.Roaster != "" {
			line += c.Roaster
		}
		if c.Origin != "" {
			if line != "" {
				line += " · "
			}
			line += c.Origin
		}
		ebitenutil.DebugPrintAt(screen, line, 10, y)
		y += lineH + 2
	}
	if c.RoastLevel != "" || c.ProcessingMethod != "" {
		ebitenutil.DebugPrintAt(screen, c.RoastLevel+" · "+c.ProcessingMethod, 10, y)
		y += lineH + 2
	}

	// Bag lifecycle info
	daysOpenLine := fmt.Sprintf("Open %d days", c.DaysOpen())
	if c.FinishedAt != nil {
		daysOpenLine += "  ·  Closed: " + c.FinishedAt.Format("2006-01-02")
	}
	ebitenutil.DebugPrintAt(screen, daysOpenLine, 10, y)
	y += lineH + 2

	// Brew progress
	bp := s.brewProgress
	brewLine := fmt.Sprintf("Brews: %d", bp.Count)
	if !bp.IsFinished {
		brewLine += fmt.Sprintf("/%d", bp.Required)
	}
	if s.hasPokemon {
		brewLine += "  ·  Has Pokemon"
	} else if bp.CanGeneratePokemon {
		brewLine += "  ·  Ready to generate!"
	} else {
		brewLine += fmt.Sprintf("  ·  Need %d more", bp.Required-bp.Count)
	}
	ebitenutil.DebugPrintAt(screen, brewLine, 10, y)
	y += lineH + 8

	// Divider
	fillRect(screen, 8, y, InternalWidth-16, 1, colorBorder)
	y += 8

	// Action items
	items := s.actionItems()
	for i, item := range items {
		drawListRow(screen, item, 40, y, InternalWidth-80, i == s.actionSel)
		y += lineH + 6
	}

	drawHints(screen, "[↑↓] Select   [Enter] Confirm   [Esc] Back")
}

func (s *WarehouseScene) drawConfirm(screen *ebiten.Image) {
	c := s.selectedCoffee()
	if c == nil {
		return
	}
	drawHeader(screen, "Close Bag — Confirm")
	y := contentY + 30

	ebitenutil.DebugPrintAt(screen, "Mark this bag as finished?", InternalWidth/2-78, y)
	y += lineH + 4
	ebitenutil.DebugPrintAt(screen, truncate(c.Name, 50), InternalWidth/2-len(c.Name)*3, y)
	y += lineH + 4
	ebitenutil.DebugPrintAt(screen, "The bag moves out of the active list.", InternalWidth/2-110, y)
	y += 32

	// Yes / No buttons
	yesX := InternalWidth/2 - 70
	noX := InternalWidth/2 + 10
	if s.confirmSel == 0 {
		fillRect(screen, yesX-2, y-2, 56, lineH+6, colorSelected)
		strokeRect(screen, yesX-2, y-2, 56, lineH+6, colorFocused)
	}
	if s.confirmSel == 1 {
		fillRect(screen, noX-2, y-2, 56, lineH+6, colorSelected)
		strokeRect(screen, noX-2, y-2, 56, lineH+6, colorFocused)
	}
	ebitenutil.DebugPrintAt(screen, "[ Yes ]", yesX, y)
	ebitenutil.DebugPrintAt(screen, "[  No ]", noX, y)

	drawHints(screen, "[←→] Select   [Enter/Y] Confirm   [N/Esc] Cancel")
}

func (s *WarehouseScene) drawGenerating(screen *ebiten.Image) {
	// ── Still running ─────────────────────────────────────────────────────────
	if s.generating {
		drawBackground(screen)
		drawHeader(screen, "Bean Warehouse — Generate Pokemon")
		y := contentY + 20
		ebitenutil.DebugPrintAt(screen, "Generating Pokemon for:", 10, y)
		y += lineH + 4
		ebitenutil.DebugPrintAt(screen, truncate(s.genCoffeeName, 60), 10, y)
		y += lineH + 16
		dots := ""
		for i := 0; i < (s.encTick/15)%4; i++ {
			dots += "."
		}
		ebitenutil.DebugPrintAt(screen, "Consulting the AI"+dots, 10, y)
		y += lineH + 4
		ebitenutil.DebugPrintAt(screen, "(this may take a moment)", 10, y)
		return
	}

	// ── Error ─────────────────────────────────────────────────────────────────
	if s.genMsg != "" {
		drawBackground(screen)
		drawHeader(screen, "Bean Warehouse — Error")
		y := contentY + 20
		ebitenutil.DebugPrintAt(screen, "Could not generate Pokemon:", 10, y)
		y += lineH + 8
		wrapText(screen, s.genMsg, 10, y, (InternalWidth-20)/6)
		drawHints(screen, "[Enter/Esc] Back")
		return
	}

	// ── Encounter animation ───────────────────────────────────────────────────
	p := s.genPokemon
	const (
		spriteW  = 80
		spriteH  = 80
		flashDur = 24
		slideLen = 40
	)
	finalSpriteX := InternalWidth/2 - spriteW/2
	finalSpriteY := InternalHeight/2 - spriteH/2 - 20

	switch s.encPhase {
	case 0: // White flash fades to normal background.
		drawBackground(screen)
		alpha := float32(flashDur-s.encTick) / float32(flashDur)
		if alpha > 0 {
			fillRect(screen, 0, 0, InternalWidth, InternalHeight, whiteOverlay(alpha))
		}

	case 1: // Text types in character by character.
		drawBackground(screen)
		drawHeader(screen, "Bean Warehouse — Encounter!")
		fullText := "A wild " + p.PokemonName + " appeared!"
		charsShown := s.encTick / 3
		runes := []rune(fullText)
		if charsShown > len(runes) {
			charsShown = len(runes)
		}
		shown := string(runes[:charsShown])
		textX := InternalWidth/2 - len([]rune(fullText))*3
		ebitenutil.DebugPrintAt(screen, shown, textX, InternalHeight/2-6)

	case 2: // Sprite slides in from the right.
		drawBackground(screen)
		drawHeader(screen, "Bean Warehouse — Encounter!")
		fullText := "A wild " + p.PokemonName + " appeared!"
		ebitenutil.DebugPrintAt(screen, fullText, InternalWidth/2-len([]rune(fullText))*3, InternalHeight/2+50)
		progress := float32(s.encTick) / float32(slideLen)
		if progress > 1 {
			progress = 1
		}
		// Ease-out quad.
		progress = 1 - (1-progress)*(1-progress)
		startX := float32(InternalWidth + spriteW)
		sx := int(startX + (float32(finalSpriteX)-startX)*progress)
		DrawSprite(screen, p.PokemonID, sx, finalSpriteY, spriteW, spriteH)

	case 3: // Full result — layout mirrors Pokemon Lab detail.
		drawBackground(screen)
		drawHeader(screen, fmt.Sprintf("Pokédex — #%03d %s", p.PokemonID, p.PokemonName))

		const resultSpriteX = InternalWidth - spriteW - 12
		const resultSpriteY = contentY + 4
		DrawSprite(screen, p.PokemonID, resultSpriteX, resultSpriteY, spriteW, spriteH)

		// Type badges under sprite.
		bx := resultSpriteX
		for _, part := range splitTypes(p.PokemonType) {
			bx = drawTypeBadge(screen, part, bx, resultSpriteY+spriteH+4)
		}

		// Info on the left.
		y := contentY + 6
		const textW = resultSpriteX - 6
		const charsPerLine = textW / 6

		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Coffee: %s", truncate(s.genCoffeeName, charsPerLine-8)), 10, y)
		y += lineH + 4

		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Level:  %d   Conf: %.0f%%", p.Level, p.MappingConfidence*100), 10, y)
		y += lineH + 8

		fillRect(screen, 8, y, textW, 1, colorBorder)
		y += 6

		if p.LLMDescription != "" {
			wrapText(screen, p.LLMDescription, 10, y, charsPerLine)
		}

		drawHints(screen, "[Enter/Esc] Continue to list")
	}
}

// whiteOverlay returns a white color.RGBA with alpha scaled by [0,1].
func whiteOverlay(alpha float32) color.RGBA {
	a := alpha
	if a < 0 {
		a = 0
	}
	if a > 1 {
		a = 1
	}
	return color.RGBA{R: 255, G: 255, B: 255, A: uint8(255 * a)}
}

func (s *WarehouseScene) drawAdd(screen *ebiten.Image) {
	drawHeader(screen, "Bean Warehouse — Add Coffee")
	y := contentY + 4

	ebitenutil.DebugPrintAt(screen, "Name *:", 10, y+2)
	s.name.Draw(screen, fieldX, y, fieldW, s.coffeeFocus == coffeeFieldName)
	y += rowStep

	ebitenutil.DebugPrintAt(screen, "Origin:", 10, y+2)
	s.origin.Draw(screen, fieldX, y, fieldW, s.coffeeFocus == coffeeFieldOrigin)
	y += rowStep

	ebitenutil.DebugPrintAt(screen, "Roaster:", 10, y+2)
	s.roaster.Draw(screen, fieldX, y, fieldW, s.coffeeFocus == coffeeFieldRoaster)
	y += rowStep

	ebitenutil.DebugPrintAt(screen, "Variety:", 10, y+2)
	s.variety.Draw(screen, fieldX, y, fieldW, s.coffeeFocus == coffeeFieldVariety)
	y += rowStep

	rlVal := roastLevels[s.roastLevelIdx]
	if rlVal == "" {
		rlVal = "(none)"
	}
	drawFieldRow(screen, "Roast Level:", fmt.Sprintf("◄ %s ►", rlVal), y, s.coffeeFocus == coffeeFieldRoastLevel)
	y += rowStep

	pmVal := processingMethods[s.processingIdx]
	if pmVal == "" {
		pmVal = "(none)"
	}
	drawFieldRow(screen, "Processing:", fmt.Sprintf("◄ %s ►", pmVal), y, s.coffeeFocus == coffeeFieldProcessing)
	y += rowStep

	ebitenutil.DebugPrintAt(screen, "Roast Date:", 10, y+2)
	s.roastDate.Draw(screen, fieldX, y, 120, s.coffeeFocus == coffeeFieldRoastDate)
	if s.coffeeFocus == coffeeFieldRoastDate {
		ebitenutil.DebugPrintAt(screen, "[Tab] next segment", fieldX+130, y+2)
	}
	y += rowStep + 2

	if s.coffeeFocus == coffeeFieldSubmit {
		fillRect(screen, InternalWidth/2-40, y-1, 80, lineH+4, colorSelected)
		strokeRect(screen, InternalWidth/2-40, y-1, 80, lineH+4, colorBorder)
	}
	ebitenutil.DebugPrintAt(screen, "[ Submit ]", InternalWidth/2-30, y)

	if s.addMsg != "" {
		ebitenutil.DebugPrintAt(screen, s.addMsg, 10, y+18)
	}

	drawHints(screen, "[↑↓] Navigate   [←→] Change   [Tab] Next   [Esc] Back")

	// Dropdowns drawn last so they appear on top
	s.origin.DrawDropdown(screen)
	s.roaster.DrawDropdown(screen)
	s.variety.DrawDropdown(screen)
}
