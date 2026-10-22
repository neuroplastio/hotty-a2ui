package cells

import (
	"github.com/neuroplastio/hotty-a2ui/view"
)

// A HottyKeyHints in cells (profile §3.3, §3.4), as bubbles' help draws
// one: the short view a line of keys and what they do, " • " between
// them; the full view their groups in columns, four columns apart. Keys
// are caps (Rendition.caps), what they do muted and faint, and the
// separators faint border, as bubbles' help steps them (the Terminal
// theme's muted and border are the same colour).

const (
	hintSep = " • "
	hintGap = "    "
)

var (
	hintDesc = style{role: Muted, attr: Faint}
	hintDim  = style{role: Border, attr: Faint}
)

// shortHints is the short view in w columns, each hint its key as key
// draws it, a space, and what it does; cut as bubbles' help cuts it: the
// hints that do not fit go, and " …" ends the line where it fits.
func shortHints(hs []view.Hint, w int, key func(string) []glyph) []glyph {
	var out []glyph
	for i, h := range hs {
		item := concat(key(h.Key), line(" "+h.Desc, hintDesc))
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
// for each group, its keys (as key draws them) padded to the widest; the
// groups that do not fit go, and " …" follows where it fits.
func fullHints(groups [][]view.Hint, w int, key func(string) []glyph) [][]glyph {
	var rows [][]glyph
	used := 0
	for i, g := range groups {
		kw, dw := 0, 0
		for _, h := range g {
			kw, dw = max(kw, width(key(h.Key))), max(dw, Width(h.Desc))
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
				k := key(h.Key)
				cell = concat(repeat(" ", gap, style{}), k,
					repeat(" ", kw-width(k)+1, style{}), line(h.Desc, hintDesc),
					repeat(" ", dw-Width(h.Desc), style{}))
			}
			rows[j] = concat(rows[j], repeat(" ", used-width(rows[j]), style{}), cell)
		}
		used += gap + cw
	}
	return rows
}

// tipStyle is a tooltip's: the text's own colour, italic, so that it reads
// over the hints' muted keys and apart from the surface's text.
var tipStyle = style{attr: Italic}

// keyHints is what a HottyKeyHints shows at w columns: on a surface with
// descriptions, a row for the tooltip (tipRow); then the full view's rows
// while it is open, else the short line.
func (r *Rendition) keyHints(e *view.Element, w int) [][]glyph {
	rows := r.hintRows(e, w)
	if r.c.V.Described() {
		rows = append([][]glyph{r.tipRow(w)}, rows...)
	}
	return rows
}

// tipRow is the tooltip, as a status line shows one (profile §6.23): the
// description of the element the pointer is on (Hover), else of the one
// with the keyboard, cut to w with "…"; blank when neither has one, so
// that nothing moves as the keyboard does.
func (r *Rendition) tipRow(w int) []glyph {
	return fit(line(r.c.Tooltip(r.hover), tipStyle), w)
}

// hintRows are a HottyKeyHints' keys: the full view's rows while it is
// open, else the short line.
func (r *Rendition) hintRows(e *view.Element, w int) [][]glyph {
	short, full := r.c.KeyHints(r.keys, false)
	if e.Open {
		if rows := fullHints(full, w, r.hintCaps); len(rows) > 0 {
			return rows
		}
		return [][]glyph{nil}
	}
	return [][]glyph{shortHints(short, w, r.hintCaps)}
}

// hintCaps is a hint's keys as caps.
func (r *Rendition) hintCaps(keys string) []glyph { return r.caps(keys, nil) }

// hintsWidth is a HottyKeyHints' natural width: its widest row of keys,
// uncut. The tooltip is cut to what they take, so that the layout does not
// move as it changes.
func (r *Rendition) hintsWidth(e *view.Element) int {
	n := 0
	for _, row := range r.hintRows(e, 1<<20) {
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
