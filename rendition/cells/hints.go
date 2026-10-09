package cells

import (
	"github.com/neuroplastio/hotty-a2ui/view"
)

// A HottyKeyHints in cells (profile §3.3, §3.4), as bubbles' help draws
// one: the short view a line of keys and what they do, " • " between
// them; the full view their groups in columns, four columns apart. Keys
// are muted, what they do fainter, and the separators faint border, the
// three steps of bubbles' help (the Terminal theme's muted and border are
// the same colour).

const (
	hintSep = " • "
	hintGap = "    "
)

var (
	hintKey  = style{role: Muted}
	hintDesc = style{role: Muted, attr: Faint}
	hintDim  = style{role: Border, attr: Faint}
)

// hintGlyphs is one hint: its key, a space, and what it does.
func hintGlyphs(h view.Hint) []glyph {
	return concat(line(h.Key, hintKey), line(" "+h.Desc, hintDesc))
}

// shortHints is the short view in w columns, cut as bubbles' help cuts
// it: the hints that do not fit go, and " …" ends the line where it fits.
func shortHints(hs []view.Hint, w int) []glyph {
	var out []glyph
	for i, h := range hs {
		item := hintGlyphs(h)
		if i > 0 {
			item = concat(line(hintSep, hintDim), item)
		}
		if width(out)+width(item) > w {
			if tail := line(" …", hintDim); width(out)+width(tail) <= w {
				out = concat(out, tail)
			}
			break
		}
		out = concat(out, item)
	}
	return out
}

// fullHints is the full view in w columns, a row of glyphs each: a column
// for each group, its keys padded to the widest; the groups that do not
// fit go, and " …" follows where it fits.
func fullHints(groups [][]view.Hint, w int) [][]glyph {
	var rows [][]glyph
	used := 0
	for i, g := range groups {
		kw, dw := 0, 0
		for _, h := range g {
			kw, dw = max(kw, Width(h.Key)), max(dw, Width(h.Desc))
		}
		cw := kw + 1 + dw
		gap := 0
		if i > 0 {
			gap = len(hintGap)
		}
		if used+gap+cw > w {
			if used > 0 && used+2 <= w && len(rows) > 0 {
				rows[0] = concat(rows[0], line(" …", hintDim))
			}
			break
		}
		for len(rows) < len(g) {
			rows = append(rows, nil)
		}
		for j := range rows {
			cell := repeat(" ", gap+cw, style{})
			if j < len(g) {
				h := g[j]
				cell = concat(repeat(" ", gap, style{}), line(h.Key, hintKey),
					repeat(" ", kw-Width(h.Key)+1, style{}), line(h.Desc, hintDesc),
					repeat(" ", dw-Width(h.Desc), style{}))
			}
			rows[j] = concat(rows[j], repeat(" ", used-width(rows[j]), style{}), cell)
		}
		used += gap + cw
	}
	return rows
}

// keyHints is what a HottyKeyHints shows at w columns: the full view's
// rows while it is open, else the short line.
func (r *Rendition) keyHints(e *view.Element, w int) [][]glyph {
	short, full := r.c.KeyHints(r.keys, false)
	if e.Open {
		if rows := fullHints(full, w); len(rows) > 0 {
			return rows
		}
		return [][]glyph{nil}
	}
	return [][]glyph{shortHints(short, w)}
}

// hintsWidth is a HottyKeyHints' natural width: its widest row, uncut.
func (r *Rendition) hintsWidth(e *view.Element) int {
	n := 0
	for _, row := range r.keyHints(e, 1<<20) {
		n = max(n, width(row))
	}
	return n
}

// paintKeyHints paints a HottyKeyHints at (x, y), w wide.
func (l *layout) paintKeyHints(cv *canvas, e *view.Element, x, y, w int) {
	for i, row := range l.r.keyHints(e, w) {
		cv.write(x, y+i, w, row)
	}
}
