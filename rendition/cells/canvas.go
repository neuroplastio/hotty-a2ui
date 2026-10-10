package cells

import (
	"strings"
)

// A glyph is one grapheme cluster with its style, to be painted.
type glyph struct {
	text  string
	width int
	style
}

// style is how a glyph is painted.
type style struct {
	role Role
	attr Attr
	link string
	// to and mix blend the colour toward another role's (Cell.To); back
	// and backMix tint the background (Cell.Back).
	to      Role
	mix     uint8
	back    Role
	backMix uint8
	// backFg and backAttr show back where it cannot tint (Cell.BackFg).
	backFg   bool
	backAttr Attr
}

// glyphs splits s into styled grapheme clusters: a tab is a space, a line
// break is a "\n" glyph, other control characters and clusters of width 0 are
// dropped.
func glyphs(s string, st style) []glyph {
	var out []glyph
	for _, g := range clusters(s) {
		switch {
		case g == "\n" || g == "\r\n" || g == "\r":
			out = append(out, glyph{text: "\n", style: st})
			continue
		case g == "\t":
			g = " "
		case len(g) == 1 && (g[0] < 0x20 || g[0] == 0x7f):
			continue
		}
		w := clusterWidth(g)
		if w == 0 || isControl(g) {
			continue
		}
		out = append(out, glyph{text: g, width: w, style: st})
	}
	return out
}

// isControl reports whether a cluster starts with a C0 or C1 control.
func isControl(g string) bool {
	r := []rune(g)[0]
	return r < 0x20 || r >= 0x7f && r < 0xa0
}

// line is s as glyphs on one line: line breaks become spaces.
func line(s string, st style) []glyph {
	gs := glyphs(s, st)
	for i, g := range gs {
		if g.text == "\n" {
			gs[i] = glyph{text: " ", width: 1, style: g.style}
		}
	}
	return gs
}

func width(gs []glyph) int {
	n := 0
	for _, g := range gs {
		n += g.width
	}
	return n
}

// fit cuts glyphs to w columns, ending in "…" when it cuts.
func fit(gs []glyph, w int) []glyph {
	if width(gs) <= w {
		return gs
	}
	if w <= 0 {
		return nil
	}
	n, i := 0, 0
	for ; i < len(gs) && n+gs[i].width <= w-1; i++ {
		n += gs[i].width
	}
	st := gs[min(i, len(gs)-1)].style
	return append(gs[:i:i], glyph{text: "…", width: 1, style: st})
}

func concat(parts ...[]glyph) []glyph {
	var out []glyph
	for _, p := range parts {
		out = append(out, p...)
	}
	return out
}

func repeat(s string, n int, st style) []glyph {
	if n <= 0 {
		return nil
	}
	return glyphs(strings.Repeat(s, n), st)
}

// canvas paints glyphs into a frame, within clip when it is set: a
// HottyScrollView's window, which its content is painted through.
type canvas struct {
	f    *Frame
	clip *box
}

// in reports whether (x, y) is inside the clip.
func (cv *canvas) in(x, y int) bool {
	c := cv.clip
	return c == nil || x >= c.x && x < c.x+c.w && y >= c.y && y < c.y+c.h
}

// clipTo narrows the clip to b, and returns the clip before, to restore.
func (cv *canvas) clipTo(b box) *box {
	was := cv.clip
	if was != nil {
		x0, y0 := max(b.x, was.x), max(b.y, was.y)
		x1, y1 := min(b.x+b.w, was.x+was.w), min(b.y+b.h, was.y+was.h)
		b = box{x0, y0, max(x1-x0, 0), max(y1-y0, 0)}
	}
	cv.clip = &b
	return was
}

// cursorAt puts the frame's cursor at (x, y), if that is inside the clip:
// a block, the cell reversed, or a line (Frame.BlockCursor).
func (cv *canvas) cursorAt(x, y int, block bool) {
	if cv.in(x, y) {
		cv.f.cursorCol, cv.f.cursorRow, cv.f.cursor, cv.f.cursorBlock = x, y, true, block
		if block {
			cv.f.Cells[y][x].Attr |= Reverse
		}
	}
}

// set paints one glyph at (x, y). A wide glyph that does not fit before
// the frame's edge is a space; a wide glyph it cuts in half is blanked.
func (cv *canvas) set(x, y int, g glyph) {
	f := cv.f
	if y < 0 || y >= f.Rows || x < 0 || x >= f.Cols || !cv.in(x, y) {
		return
	}
	row := f.Cells[y]
	if g.width == 2 && x+1 >= f.Cols {
		g = glyph{text: " ", width: 1, style: g.style}
	}
	cv.unwide(x, y)
	row[x] = Cell{Text: g.text, Width: g.width, Role: g.role, Attr: g.attr, Link: g.link, To: g.to, Mix: g.mix,
		Back: g.back, BackMix: g.backMix, BackFg: g.backFg, BackAttr: g.backAttr}
	if g.width == 2 {
		cv.unwide(x+1, y)
		row[x+1] = Cell{Width: 0, Role: g.role, Attr: g.attr, Link: g.link, To: g.to, Mix: g.mix,
			Back: g.back, BackMix: g.backMix, BackFg: g.backFg, BackAttr: g.backAttr}
	}
}

// unwide blanks the other half of a wide glyph that (x, y) is part of.
func (cv *canvas) unwide(x, y int) {
	row := cv.f.Cells[y]
	switch {
	case row[x].Width == 0 && row[x].Text == "" && x > 0:
		row[x-1] = blank
	case row[x].Width == 2 && x+1 < len(row):
		row[x+1] = blank
	}
}

// write paints glyphs from (x, y), within w columns: a wide glyph that
// would cross the edge is left out, and so is everything after it. It
// returns the columns painted.
func (cv *canvas) write(x, y, w int, gs []glyph) int {
	n := 0
	for _, g := range gs {
		if g.width == 0 {
			continue
		}
		if n+g.width > w {
			break
		}
		cv.set(x+n, y, g)
		n += g.width
	}
	return n
}

// fill paints a rectangle with blank cells.
func (cv *canvas) fill(x, y, w, h int) {
	for j := y; j < y+h; j++ {
		for i := x; i < x+w; i++ {
			cv.set(i, j, glyph{text: " ", width: 1})
		}
	}
}

// restyle changes the style of the cells in a rectangle.
func (cv *canvas) restyle(x, y, w, h int, fn func(c *Cell)) {
	for j := max(y, 0); j < min(y+h, cv.f.Rows); j++ {
		for i := max(x, 0); i < min(x+w, cv.f.Cols); i++ {
			if !cv.in(i, j) {
				continue
			}
			fn(&cv.f.Cells[j][i])
		}
	}
}

// box paints a rounded box in the border role.
func (cv *canvas) box(x, y, w, h int) {
	if w < 2 || h < 2 {
		return
	}
	b := style{role: Border}
	cv.write(x, y, w, concat(glyphs("╭", b), repeat("─", w-2, b), glyphs("╮", b)))
	for j := y + 1; j < y+h-1; j++ {
		cv.set(x, j, glyph{text: "│", width: 1, style: b})
		cv.set(x+w-1, j, glyph{text: "│", width: 1, style: b})
	}
	cv.write(x, y+h-1, w, concat(glyphs("╰", b), repeat("─", w-2, b), glyphs("╯", b)))
}
