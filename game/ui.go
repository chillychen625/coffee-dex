package game

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	lineH   = 12
	padding = 4

	labelW  = 114
	fieldX  = 8 + labelW + 2
	fieldW  = InternalWidth - fieldX - 8
	fieldH  = lineH + 4
	rowStep = fieldH + 3

	contentY = 18
	hintsY   = InternalHeight - lineH - padding*2
)

var (
	colorBg       = color.RGBA{R: 15, G: 20, B: 35, A: 255}
	colorPanel    = color.RGBA{R: 28, G: 36, B: 60, A: 255}
	colorBorder   = color.RGBA{R: 80, G: 120, B: 180, A: 255}
	colorFocused  = color.RGBA{R: 200, G: 220, B: 255, A: 255}
	colorSelected = color.RGBA{R: 40, G: 80, B: 150, A: 255}
	colorInput    = color.RGBA{R: 20, G: 28, B: 50, A: 255}
	colorCursor   = color.RGBA{R: 200, G: 200, B: 255, A: 220}
	colorHeader   = color.RGBA{R: 20, G: 30, B: 55, A: 255}
	colorHints    = color.RGBA{R: 10, G: 14, B: 28, A: 255}
	colorSuccess  = color.RGBA{R: 60, G: 180, B: 100, A: 255}
	colorError    = color.RGBA{R: 200, G: 70, B: 70, A: 255}
	colorMuted    = color.RGBA{R: 100, G: 110, B: 140, A: 255}
)

// typeColors maps lowercase Pokemon type names to their Gen 1 palette colors.
var typeColors = map[string]color.RGBA{
	"normal":   {R: 168, G: 168, B: 120, A: 255},
	"fire":     {R: 240, G: 128, B: 48, A: 255},
	"water":    {R: 104, G: 144, B: 240, A: 255},
	"grass":    {R: 120, G: 200, B: 80, A: 255},
	"electric": {R: 248, G: 208, B: 48, A: 255},
	"ice":      {R: 152, G: 216, B: 216, A: 255},
	"fighting": {R: 192, G: 48, B: 40, A: 255},
	"poison":   {R: 160, G: 64, B: 160, A: 255},
	"ground":   {R: 224, G: 192, B: 104, A: 255},
	"flying":   {R: 168, G: 144, B: 240, A: 255},
	"psychic":  {R: 248, G: 88, B: 136, A: 255},
	"bug":      {R: 168, G: 184, B: 32, A: 255},
	"rock":     {R: 184, G: 160, B: 56, A: 255},
	"ghost":    {R: 112, G: 88, B: 152, A: 255},
	"dragon":   {R: 112, G: 56, B: 248, A: 255},
	"dark":     {R: 112, G: 88, B: 72, A: 255},
	"steel":    {R: 184, G: 184, B: 208, A: 255},
	"fairy":    {R: 240, G: 182, B: 188, A: 255},
}

// getTypeColor returns the color for a Pokemon type name (case-insensitive).
func getTypeColor(typeName string) color.RGBA {
	if c, ok := typeColors[strings.ToLower(typeName)]; ok {
		return c
	}
	return colorMuted
}

// splitTypes splits "Fire/Flying" → ["Fire", "Flying"].
func splitTypes(t string) []string {
	var parts []string
	for _, p := range strings.Split(t, "/") {
		if s := strings.TrimSpace(p); s != "" {
			parts = append(parts, s)
		}
	}
	return parts
}

// primaryTypeColor returns the color for the first type in a "Fire/Flying" string.
func primaryTypeColor(pokemonType string) color.RGBA {
	parts := splitTypes(pokemonType)
	if len(parts) == 0 {
		return colorMuted
	}
	return getTypeColor(parts[0])
}

// drawTypeBadge draws a colored type pill and returns the x position after it.
func drawTypeBadge(screen *ebiten.Image, typeName string, x, y int) int {
	if typeName == "" {
		return x
	}
	c := getTypeColor(typeName)
	w := len(typeName)*6 + 8
	fillRect(screen, x, y, w, lineH+2, c)
	ebitenutil.DebugPrintAt(screen, typeName, x+4, y+1)
	return x + w + 4
}

// drawHBar draws a labeled horizontal bar chart row.
// labelCols is the label width in character columns (6px each).
// barW is the pixel width of the bar.
func drawHBar(screen *ebiten.Image, x, y, labelCols, barW int, label string, val, maxVal float64, c color.RGBA) {
	ebitenutil.DebugPrintAt(screen, truncate(label, labelCols), x, y+1)
	bx := x + labelCols*6 + 2
	fillRect(screen, bx, y, barW, lineH, colorInput)
	strokeRect(screen, bx, y, barW, lineH, colorBorder)
	if maxVal > 0 && val > 0 {
		fill := int(float64(barW-2) * val / maxVal)
		if fill > 0 {
			fillRect(screen, bx+1, y+1, fill, lineH-2, c)
		}
	}
	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%.0f", val), bx+barW+4, y+1)
}

func fillRect(screen *ebiten.Image, x, y, w, h int, c color.RGBA) {
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(w), float32(h), c, false)
}

func strokeRect(screen *ebiten.Image, x, y, w, h int, c color.RGBA) {
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(w), 1, c, false)
	vector.DrawFilledRect(screen, float32(x), float32(y+h-1), float32(w), 1, c, false)
	vector.DrawFilledRect(screen, float32(x), float32(y), 1, float32(h), c, false)
	vector.DrawFilledRect(screen, float32(x+w-1), float32(y), 1, float32(h), c, false)
}

func drawBackground(screen *ebiten.Image) {
	screen.Fill(colorBg)
}

func drawHeader(screen *ebiten.Image, title string) {
	fillRect(screen, 0, 0, InternalWidth, contentY, colorHeader)
	ebitenutil.DebugPrintAt(screen, title, padding*2, padding/2+1)
}

func drawHints(screen *ebiten.Image, hints string) {
	fillRect(screen, 0, hintsY, InternalWidth, InternalHeight-hintsY, colorHints)
	ebitenutil.DebugPrintAt(screen, hints, padding*2, hintsY+padding/2+1)
}

// drawFieldRow draws a label + value row. Returns the y bottom of the row.
func drawFieldRow(screen *ebiten.Image, label, value string, y int, focused bool) {
	bc := colorBorder
	if focused {
		bc = colorFocused
		fillRect(screen, 8, y, InternalWidth-16, fieldH, colorSelected)
	}
	ebitenutil.DebugPrintAt(screen, label, 8+2, y+2)
	fillRect(screen, fieldX, y, fieldW, fieldH, colorInput)
	strokeRect(screen, fieldX, y, fieldW, fieldH, bc)
	ebitenutil.DebugPrintAt(screen, value, fieldX+3, y+2)
}

// drawListRow draws a selectable list row.
func drawListRow(screen *ebiten.Image, text string, x, y, w int, selected bool) {
	if selected {
		fillRect(screen, x, y, w, lineH+2, colorSelected)
	}
	prefix := "  "
	if selected {
		prefix = "► "
	}
	ebitenutil.DebugPrintAt(screen, prefix+text, x+4, y+1)
}

// drawListRowTyped draws a selectable list row with a type-color accent bar on the left.
func drawListRowTyped(screen *ebiten.Image, text string, x, y, w int, selected bool, pokemonType string) {
	tc := primaryTypeColor(pokemonType)
	if selected {
		fillRect(screen, x, y, w, lineH+2, colorSelected)
	} else {
		fillRect(screen, x, y, 3, lineH+2, tc)
	}
	prefix := "  "
	if selected {
		prefix = "► "
	}
	ebitenutil.DebugPrintAt(screen, prefix+text, x+4, y+1)
}

// drawMsg draws a status message (success=true for green, false for red).
func drawMsg(screen *ebiten.Image, msg string, y int, success bool) {
	if msg == "" {
		return
	}
	c := colorSuccess
	if !success {
		c = colorError
	}
	fillRect(screen, 8, y, InternalWidth-16, lineH+2, color.RGBA{})
	_ = c
	ebitenutil.DebugPrintAt(screen, msg, 12, y+1)
}

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n-1]) + "…"
}

// wrapText draws word-wrapped text and returns the y position after the last line.
func wrapText(screen *ebiten.Image, text string, x, y, maxChars int) int {
	desc := []rune(text)
	for len(desc) > 0 && y < hintsY-lineH {
		end := maxChars
		if end > len(desc) {
			end = len(desc)
		}
		if end < len(desc) {
			for end > 0 && desc[end] != ' ' {
				end--
			}
			if end == 0 {
				end = maxChars
			}
		}
		ebitenutil.DebugPrintAt(screen, string(desc[:end]), x, y)
		desc = desc[end:]
		if len(desc) > 0 && desc[0] == ' ' {
			desc = desc[1:]
		}
		y += lineH
	}
	return y
}
