package game

import (
	"slices"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const (
	maxLines  = 4
	maxChars  = 41
	charWidth = 6
)

type MultiLineInput struct {
	Lines     []string
	CursorRow int
	CursorCol int
	blink     int
}

func (m *MultiLineInput) Clear() {
	m.Lines = []string{""}
	m.CursorRow = 0
	m.CursorCol = 0
	m.blink = 0
}

func (m *MultiLineInput) Text() string {
	return strings.Join(m.Lines, "\n")
}

func (m *MultiLineInput) Update(focused bool) {
	if !focused {
		return
	}
	if len(m.Lines) == 0 {
		m.Lines = []string{""}
	}
	m.blink++

	var chars []rune
	chars = ebiten.AppendInputChars(chars)
	for _, ch := range chars {
		if ch < 32 {
			continue
		}
		lineRunes := []rune(m.Lines[m.CursorRow])
		if len(lineRunes) >= maxChars {
			continue
		}
		lineRunes = append(lineRunes[:m.CursorCol:m.CursorCol], append([]rune{ch}, lineRunes[m.CursorCol:]...)...)
		m.Lines[m.CursorRow] = string(lineRunes)
		m.CursorCol++
		m.blink = 0
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) && (m.CursorRow > 0 || m.CursorCol > 0) {
		if m.CursorCol == 0 {
			m.Lines[m.CursorRow-1] = m.Lines[m.CursorRow-1] + m.Lines[m.CursorRow]
			m.Lines = append(m.Lines[:m.CursorRow], m.Lines[m.CursorRow+1:]...)
			m.CursorRow--
			m.CursorCol = len([]rune(m.Lines[m.CursorRow]))
		} else {
			runes := []rune(m.Lines[m.CursorRow])
			runes = append(runes[:m.CursorCol-1], runes[m.CursorCol:]...)
			m.Lines[m.CursorRow] = string(runes)
			m.CursorCol--
		}
		m.blink = 0
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyDelete) {
		lineRunes := []rune(m.Lines[m.CursorRow])
		if m.CursorCol == len(lineRunes) && m.CursorRow < len(m.Lines)-1 {
			m.Lines[m.CursorRow] = m.Lines[m.CursorRow] + m.Lines[m.CursorRow+1]
			m.Lines = append(m.Lines[:m.CursorRow+1], m.Lines[m.CursorRow+2:]...)
		} else if m.CursorCol < len(lineRunes) {
			lineRunes = append(lineRunes[:m.CursorCol], lineRunes[m.CursorCol+1:]...)
			m.Lines[m.CursorRow] = string(lineRunes)
		}
		m.blink = 0
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) && (m.CursorRow > 0 || m.CursorCol > 0) {
		if m.CursorCol == 0 {
			m.CursorRow--
			m.CursorCol = len([]rune(m.Lines[m.CursorRow]))
		} else {
			m.CursorCol--
		}
		m.blink = 0
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) {
		atLineEnd := m.CursorCol >= len([]rune(m.Lines[m.CursorRow]))
		hasNext := m.CursorRow < len(m.Lines)-1
		if atLineEnd && hasNext {
			m.CursorRow++
			m.CursorCol = 0
		} else if !atLineEnd {
			m.CursorCol++
		}
		m.blink = 0
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) && m.CursorRow > 0 {
		m.CursorRow--
		if m.CursorCol > len([]rune(m.Lines[m.CursorRow])) {
			m.CursorCol = len([]rune(m.Lines[m.CursorRow]))
		}
		m.blink = 0
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) && m.CursorRow < len(m.Lines)-1 {
		m.CursorRow++
		if m.CursorCol > len([]rune(m.Lines[m.CursorRow])) {
			m.CursorCol = len([]rune(m.Lines[m.CursorRow]))
		}
		m.blink = 0
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) && len(m.Lines) < maxLines {
		runes := []rune(m.Lines[m.CursorRow])
		tail := string(runes[m.CursorCol:])
		m.Lines[m.CursorRow] = string(runes[:m.CursorCol])
		m.Lines = slices.Insert(m.Lines, m.CursorRow+1, tail)
		m.CursorRow++
		m.CursorCol = 0
		m.blink = 0
	}
}

func (m *MultiLineInput) Draw(screen *ebiten.Image, x, y, w, h int, focused bool) {
	fillRect(screen, x, y, w, h, colorInput)
	bc := colorBorder
	if focused {
		bc = colorFocused
	}
	strokeRect(screen, x, y, w, h, bc)

	maxCols := (w - 6) / charWidth
	for i, line := range m.Lines {
		runes := []rune(line)
		offset := 0
		if focused && i == m.CursorRow && len(runes) > maxCols {
			offset = min(max(m.CursorCol-maxCols+1, 0), len(runes)-maxCols)
		}
		end := min(offset+maxCols, len(runes))
		ebitenutil.DebugPrintAt(screen, string(runes[offset:end]), x+3, y+2+i*(lineH+2))
	}

	if focused && m.blink%60 < 40 {
		runes := []rune(m.Lines[m.CursorRow])
		offset := 0
		if len(runes) > maxCols {
			offset = min(max(m.CursorCol-maxCols+1, 0), len(runes)-maxCols)
		}
		cx := x + 3 + (m.CursorCol-offset)*charWidth
		cy := y + 2 + m.CursorRow*(lineH+2)
		fillRect(screen, cx, cy, 1, lineH, colorCursor)
	}
}
