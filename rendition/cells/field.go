package cells

import (
	"github.com/neuroplastio/hotty-a2ui/view"
)

// isBreak reports whether a cluster is a line break.
func isBreak(g string) bool { return g == "\n" || g == "\r\n" || g == "\r" }

// splitClusters splits a value's clusters into lines at its breaks.
func splitClusters(cl []string) [][]string {
	out := [][]string{nil}
	for _, g := range cl {
		if isBreak(g) {
			out = append(out, nil)
			continue
		}
		out[len(out)-1] = append(out[len(out)-1], g)
	}
	return out
}

// locate is the line of a cursor (a cluster index into the value), and
// its index in that line.
func locate(lines [][]string, pos int) (li, ci int) {
	for li = range lines {
		if pos <= len(lines[li]) {
			return li, pos
		}
		pos -= len(lines[li]) + 1
	}
	return len(lines) - 1, len(lines[len(lines)-1])
}

// offset is the cursor of line li's cluster ci.
func offset(lines [][]string, li, ci int) int {
	pos := 0
	for i := 0; i < li; i++ {
		pos += len(lines[i]) + 1
	}
	return pos + ci
}

// shownGlyph is how a field shows one cluster of its value: "•" when
// obscured, a break or a control as a space.
func shownGlyph(g string, obscured bool, st style) glyph {
	switch {
	case obscured:
		return glyph{text: "•", width: 1, style: st}
	case isBreak(g) || g == "\t" || isControl(g):
		return glyph{text: " ", width: 1, style: st}
	}
	return glyph{text: g, width: clusterWidth(g), style: st}
}

// colOf is the column of a line's cluster ci.
func colOf(line []string, ci int, obscured bool) int {
	n := 0
	for _, g := range line[:ci] {
		n += shownGlyph(g, obscured, style{}).width
	}
	return n
}

// indexAt is the cursor in a line that a click at column col puts: before
// the cluster under it, or at the end.
func indexAt(line []string, col int, obscured bool) int {
	n := 0
	for i, g := range line {
		gw := shownGlyph(g, obscured, style{}).width
		if col < n+gw {
			return i
		}
		n += gw
	}
	return len(line)
}

// paintField paints a TextField or a DateTime (profile §3.5): a muted
// label line, then the value on an underlined field as wide as its box,
// scrolled to keep the cursor in it while it has the keyboard.
func (l *layout) paintField(cv *canvas, e *view.Element, x, y, w int) {
	r := l.r
	focused := r.focused(e.ID)
	row := y
	if e.Label != "" {
		st := style{role: Muted}
		if focused {
			st.role = Accent
		}
		cv.write(x, row, w, fit(line(e.Label, st), w))
		row++
	}
	v, _ := e.Value.(string)
	cl := clusters(v)
	obscured := e.Variant == "obscured"
	lines := [][]string{cl}
	if isLongText(e) {
		lines = splitClusters(cl)
	}
	rows := fieldRows(e)
	r.rows[e.ID] = rows
	under := style{attr: Underline}
	area := &fieldArea{x: x, y: row, rows: rows}
	li, ci, curCol := 0, 0, 0
	if focused {
		li, ci = locate(lines, r.cursorOf(e.ID, len(cl)))
		curCol = colOf(lines[li], ci, obscured)
		voff := min(r.vscroll[e.ID], li)
		if li >= voff+rows {
			voff = li - rows + 1
		}
		voff = max(min(voff, len(lines)-rows), 0)
		hoff := min(r.hscroll[e.ID], curCol)
		if curCol >= hoff+w {
			hoff = curCol - w + 1
		}
		if colOf(lines[li], len(lines[li]), obscured) < w {
			hoff = 0
		}
		r.vscroll[e.ID], r.hscroll[e.ID] = voff, hoff
		area.voff, area.hoff = voff, hoff
	}
	for j := 0; j < rows; j++ {
		cv.write(x, row+j, w, repeat(" ", w, under))
		k := area.voff + j
		if k >= len(lines) {
			continue
		}
		gs := make([]glyph, len(lines[k]))
		for i, g := range lines[k] {
			gs[i] = shownGlyph(g, obscured, under)
		}
		if !focused {
			cv.write(x, row+j, w, fit(gs, w))
			continue
		}
		c := 0
		for _, g := range gs {
			if c >= area.hoff && c+g.width <= area.hoff+w {
				cv.set(x+c-area.hoff, row+j, g)
			}
			c += g.width
		}
	}
	if v == "" {
		hint := e.Placeholder
		if hint == "" && e.Kind == view.DateTime {
			hint = dateHint(e)
		}
		cv.write(x, row, w, fit(line(hint, style{role: Muted, attr: Faint | Underline}), w))
	}
	if focused {
		cx, cy := x+curCol-area.hoff, row+li-area.voff
		if cx >= 0 && cx < cv.f.Cols && cy >= 0 && cy < cv.f.Rows {
			cv.f.Cells[cy][cx].Attr |= Reverse
			cv.f.cursorCol, cv.f.cursorRow, cv.f.cursor = cx, cy, true
		}
	}
	r.hits = append(r.hits, hit{x: x, y: y, w: w, h: row - y + rows, id: e.ID, opt: -1, field: area})
	l.paintError(cv, e, x, row+rows, w)
}
