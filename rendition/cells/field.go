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

// selected is how a field shows the selected part of its value: on the
// selection colour, or reversed where the theme can't tint, as a terminal
// shows its own selection.
var selected = style{back: Selection, backMix: 255, backAttr: Reverse}

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

// paintField paints a TextField or a DateTime (profile §3.5) in its box
// past the gutter: a bold title, then the value. A one-line field's value
// follows a "> " prompt, as bubbles' text input has it; a longText's rows
// carry the textarea's "┃" in the gutter. It scrolls to keep the cursor in
// it while it has the keyboard.
func (l *layout) paintField(cv *canvas, e *view.Element, x, y, w int) {
	r := l.r
	focused := r.focused(e.ID)
	row := y
	if e.Label != "" {
		cv.write(x, row, w, fit(line(e.Label, titleStyle(focused)), w))
		row++
	}
	v, _ := e.Value.(string)
	cl := clusters(v)
	obscured := e.Variant == "obscured"
	long := isLongText(e)
	lines := [][]string{cl}
	if long {
		lines = splitClusters(cl)
	}
	rows := fieldRows(e)
	r.rows[e.ID] = rows
	vx, vw := x, w
	if !long && w > prompt {
		p := style{role: Muted}
		if focused {
			p.role = Accent
		}
		cv.write(x, row, w, glyphs("> ", p))
		vx, vw = x+prompt, w-prompt
	}
	area := &fieldArea{x: vx, y: row, rows: rows}
	li, ci, curCol := 0, 0, 0
	// What Shift and select-all selected (SPEC §10.2), while the field
	// editKey keeps is the one shown.
	s0, s1 := 0, 0
	if focused {
		pos := r.cursorOf(e.ID, len(cl))
		if f := r.fields[e.ID]; f != nil && f.Value == v && f.Caret == pos {
			s0, s1 = f.Selection()
		}
		li, ci = locate(lines, pos)
		curCol = colOf(lines[li], ci, obscured)
		voff := min(r.vscroll[e.ID], li)
		if li >= voff+rows {
			voff = li - rows + 1
		}
		voff = max(min(voff, len(lines)-rows), 0)
		hoff := min(r.hscroll[e.ID], curCol)
		if curCol >= hoff+vw {
			hoff = curCol - vw + 1
		}
		if colOf(lines[li], len(lines[li]), obscured) < vw {
			hoff = 0
		}
		r.vscroll[e.ID], r.hscroll[e.ID] = voff, hoff
		area.voff, area.hoff = voff, hoff
	}
	for j := 0; j < rows; j++ {
		if long {
			cv.set(x-gutter, row+j, glyph{text: "┃", width: 1, style: style{role: Border}})
		}
		k := area.voff + j
		if k >= len(lines) {
			continue
		}
		gs := make([]glyph, len(lines[k]))
		base := offset(lines, k, 0)
		for i, g := range lines[k] {
			gs[i] = shownGlyph(g, obscured, style{})
			if base+i >= s0 && base+i < s1 {
				gs[i].style = selected
			}
		}
		if !focused {
			cv.write(vx, row+j, vw, fit(gs, vw))
			continue
		}
		c := 0
		for _, g := range gs {
			if c >= area.hoff && c+g.width <= area.hoff+vw {
				cv.set(vx+c-area.hoff, row+j, g)
			}
			c += g.width
		}
	}
	if v == "" {
		hint := e.Placeholder
		if hint == "" && e.Kind == view.DateTime {
			hint = e.DateHint()
		}
		cv.write(vx, row, vw, fit(line(hint, style{role: Muted, attr: Faint}), vw))
	}
	if focused {
		cx, cy := vx+curCol-area.hoff, row+li-area.voff
		if cx >= 0 && cx < cv.f.Cols && cy >= 0 && cy < cv.f.Rows && cv.in(cx, cy) {
			// A selection shows no block caret, as a GUI's field shows
			// none beside one: it would read as one more selected cell.
			if s0 == s1 {
				cv.f.Cells[cy][cx].Attr |= Reverse
			}
			cv.cursorAt(cx, cy)
		}
	}
	r.hits = append(r.hits, hit{x: x - gutter, y: y, w: w + gutter, h: row - y + rows, id: e.ID, opt: -1, field: area})
	l.paintError(cv, e, x, row+rows, w)
}
