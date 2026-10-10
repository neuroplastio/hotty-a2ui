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

// bigFont is a size's font: its glyphs, their rows of pixels, and whether
// a pixel is two columns and a row (wide) rather than half a cell.
type bigFont struct {
	glyphs map[rune][]string
	px     int
	wide   bool
}

var bigFonts = map[string]bigFont{
	view.BigSmall:  {glyphs: font3x5, px: 5},
	view.BigMedium: {glyphs: font5x7, px: 7},
	view.BigLarge:  {glyphs: font3x5, px: 5, wide: true},
}

// bigFontOf is a HottyBigText's font, by its size.
func bigFontOf(e *view.Element) bigFont {
	if f, ok := bigFonts[e.Variant]; ok {
		return f
	}
	return bigFonts[view.BigMedium]
}

// rows is a line's rows of cells.
func (f bigFont) rows() int {
	if f.wide {
		return f.px
	}
	return (f.px + 1) / 2
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
// (two when wide), and a column between each two.
func (f bigFont) width(rs []rune) int {
	n := 0
	for i, r := range rs {
		if i > 0 {
			n++
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

// bigRows are a line's rows of cells, as text: its glyphs' a column apart.
func (f bigFont) bigRows(rs []rune) []string {
	rows := make([]string, f.rows())
	for i, r := range rs {
		for y, s := range f.cells(f.glyph(r)) {
			if i > 0 {
				rows[y] += " "
			}
			rows[y] += s
		}
	}
	return rows
}

// cells are a glyph's rows of cells, as text.
func (f bigFont) cells(g []string) []string {
	out := make([]string, 0, f.rows())
	var b strings.Builder
	if f.wide {
		for _, row := range g {
			b.Reset()
			for _, c := range row {
				if c == '#' {
					b.WriteString("██")
				} else {
					b.WriteString("  ")
				}
			}
			out = append(out, b.String())
		}
		return out
	}
	for y := 0; y < f.px; y += 2 {
		b.Reset()
		for x := range len(g[y]) {
			b.WriteString(halfBlock(g[y][x] == '#', y+1 < f.px && g[y+1][x] == '#'))
		}
		out = append(out, b.String())
	}
	return out
}

// halfBlock is the cell of two pixels, one over the other: inked, the
// terminal's text colour; blank, its background.
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
		for _, s := range f.bigRows(ln) {
			if row >= y+h {
				return
			}
			cv.write(x+dx, row, w-dx, glyphs(s, style{}))
			row++
		}
		row++
	}
}
