package cells

import (
	"strings"
	"unicode"

	"github.com/neuroplastio/hotty-a2ui/view"
)

// A HottyBigText in cells (profile §6.28): its text in block letters, a
// pixel font's glyphs (bigfont.go) drawn in block elements, the terminal's
// own drawing where it has one, so that the pixels meet with no seam. At
// the small and medium sizes a pixel is a column and half a row, two
// pixels a cell (▀ ▄ █), about square in a terminal's cell; at the large
// size a pixel is two columns and a row (██). Letters are a column apart,
// words a space's glyph and a column each side, and lines a blank row. A
// line too long for its box wraps between words, and a word too long for
// a line of its own between letters, so that nothing is cut.

// bigFont is a size's font: its glyphs, their rows of pixels, whether a
// pixel is two columns and a row (wide) rather than half a cell, and
// whether its text casts a shadow, a column and half a row over.
type bigFont struct {
	glyphs map[rune][]string
	px     int
	wide   bool
	shadow bool
}

var bigFonts = map[string]bigFont{
	view.BigSmall:  {glyphs: font3x5, px: 5},
	view.BigMedium: {glyphs: font5x7, px: 7},
	view.BigLarge:  {glyphs: font3x5, px: 5, wide: true},
}

// bigFontOf is a HottyBigText's font, by its size, with its shadow.
func bigFontOf(e *view.Element) bigFont {
	f, ok := bigFonts[e.Variant]
	if !ok {
		f = bigFonts[view.BigMedium]
	}
	f.shadow = e.Shadow
	return f
}

// rows is a line's rows of cells, two half rows each: a shadow's half row
// more fits in the half row the small and medium fonts leave blank, and
// is a row more at large.
func (f bigFont) rows() int {
	h := f.px
	if f.wide {
		h *= 2
	}
	if f.shadow {
		h++
	}
	return (h + 1) / 2
}

// gap is the columns between two glyphs: one, and the shadow's.
func (f bigFont) gap() int {
	if f.shadow {
		return 2
	}
	return 1
}

// glyph is the glyph r is drawn as: its own, its capital's, its letter's
// without a mark (latinBase), else the question mark's.
func (f bigFont) glyph(r rune) []string {
	if g, ok := f.glyphs[r]; ok {
		return g
	}
	if g, ok := f.glyphs[unicode.ToUpper(r)]; ok {
		return g
	}
	if b, ok := latinBase[r]; ok {
		return f.glyphs[b]
	}
	return f.glyphs['?']
}

// width is the columns a run of runes takes: its glyphs, a column a pixel
// (two when wide), the gap between each two, and its shadow's column.
func (f bigFont) width(rs []rune) int {
	n := 0
	if f.shadow && len(rs) > 0 {
		n = 1
	}
	for i, r := range rs {
		if i > 0 {
			n += f.gap()
		}
		px := len(f.glyph(r)[0])
		if f.wide {
			px *= 2
		}
		n += px
	}
	return n
}

// bigLines are a HottyBigText's lines at w columns, each the runes drawn
// on it: its own lines (at "\n"), each wrapped at the last space that
// fits, and a word too long for a line of its own at its last letter that
// does; unwrapped at noWrap. Runs of spaces are one.
func bigLines(e *view.Element, w int) [][]rune {
	if e.Label == "" {
		return nil
	}
	f := bigFontOf(e)
	var out [][]rune
	for _, para := range strings.Split(e.Label, "\n") {
		var cur []rune
		for _, word := range strings.Fields(para) {
			wd := []rune(word)
			if len(cur) > 0 {
				if next := append(append(cur[:len(cur):len(cur)], ' '), wd...); f.width(next) <= w {
					cur = next
					continue
				}
				out = append(out, cur)
			}
			for len(wd) > 1 && f.width(wd) > w {
				n := len(wd) - 1
				for n > 1 && f.width(wd[:n]) > w {
					n--
				}
				out = append(out, wd[:n])
				wd = wd[n:]
			}
			cur = wd
		}
		out = append(out, cur)
	}
	return out
}

// bigWidth is a HottyBigText's natural width, its widest line unwrapped;
// bigMinimum its widest word, under which a word breaks between letters.
func bigWidth(e *view.Element) int {
	f, n := bigFontOf(e), 0
	for _, l := range bigLines(e, noWrap) {
		n = max(n, f.width(l))
	}
	return n
}

func bigMinimum(e *view.Element) int {
	f, n := bigFontOf(e), 0
	for _, word := range strings.Fields(e.Label) {
		n = max(n, f.width([]rune(word)))
	}
	return n
}

// bigHeight is its rows at w columns: its lines' rows, a blank row apart.
func bigHeight(e *view.Element, w int) int {
	n := len(bigLines(e, w))
	if n == 0 {
		return 0
	}
	return n*bigFontOf(e).rows() + n - 1
}

// bigCells are a line's rows of cells: its glyphs a gap apart, on a grid
// of columns and half rows, a pixel each at the small and medium sizes
// and two by two at large, each cell drawn from its two halves (bigCell).
// A shadow is the ink a column and half a row down and to the right, where
// there is no ink: a pixel over at small and medium, half one at large,
// which a whole one would swamp.
func (f bigFont) bigCells(rs []rune) [][]glyph {
	w, h := f.width(rs), 2*f.rows()
	px := 1
	if f.wide {
		px = 2
	}
	ink := make([][]bool, h)
	for y := range ink {
		ink[y] = make([]bool, w)
	}
	x := 0
	for i, r := range rs {
		if i > 0 {
			x += f.gap()
		}
		g := f.glyph(r)
		for gy, row := range g {
			for gx, c := range row {
				if c != '#' {
					continue
				}
				for dy := range px {
					for dx := range px {
						ink[gy*px+dy][x+gx*px+dx] = true
					}
				}
			}
		}
		x += len(g[0]) * px
	}
	at := func(x, y int) bigPixel {
		switch {
		case ink[y][x]:
			return bigInk
		case f.shadow && x > 0 && y > 0 && ink[y-1][x-1]:
			return bigShadow
		}
		return bigBlank
	}
	rows := make([][]glyph, f.rows())
	for y := range rows {
		for x := range w {
			rows[y] = append(rows[y], bigCell(at(x, 2*y), at(x, 2*y+1)))
		}
	}
	return rows
}

// bigPixel is what a pixel of big text is: blank, ink or shadow.
type bigPixel uint8

const (
	bigBlank bigPixel = iota
	bigInk
	bigShadow
)

// The shadow is in the border's colour, faint under NO_COLOR; a cell of
// ink over shadow (or under it) is the ink's half block on the shadow's
// colour, where the terminal can show it, else the ink alone.
var (
	shadowStyle = style{role: Border, monoAttr: Faint}
	onShadow    = style{back: Border, backMix: 255, backShade: true}
)

// bigCell is the cell of two pixels, one over the other: a full block or
// a half one in the ink's colour or the shadow's, or the ink's half on the
// shadow's colour.
func bigCell(top, bottom bigPixel) glyph {
	if top != bottom && top != bigBlank && bottom != bigBlank {
		return glyph{text: halfBlock(top == bigInk, bottom == bigInk), width: 1, style: onShadow}
	}
	st := style{}
	if top == bigShadow || bottom == bigShadow {
		st = shadowStyle
	}
	return glyph{text: halfBlock(top != bigBlank, bottom != bigBlank), width: 1, style: st}
}

// halfBlock is the cell of two pixels, one over the other, inked or blank.
func halfBlock(top, bottom bool) string {
	switch {
	case top && bottom:
		return "█"
	case top:
		return "▀"
	case bottom:
		return "▄"
	}
	return " "
}

// paintBigText paints a HottyBigText's lines in its box, w wide, each at
// its align, in the text's colour.
func (l *layout) paintBigText(cv *canvas, e *view.Element, x, y, w, h int) {
	f := bigFontOf(e)
	row := y
	for _, ln := range bigLines(e, w) {
		dx := 0
		switch lw := f.width(ln); e.Align {
		case "center":
			dx = max(w-lw, 0) / 2
		case "end":
			dx = max(w-lw, 0)
		}
		for _, gs := range f.bigCells(ln) {
			if row >= y+h {
				return
			}
			cv.write(x+dx, row, w-dx, gs)
			row++
		}
		row++
	}
}
