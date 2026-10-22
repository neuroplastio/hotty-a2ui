package cells

import (
	"slices"
	"strings"

	"github.com/neuroplastio/hotty-a2ui/view"
)

// A HottyPicker in cells (profile §3.3, §3.4, §3.7, §6.32), as a
// command palette draws one: the query on the first row, after its chip;
// a rule; the status, while it has one; the rows, a heading wherever their
// section changes, the selected one reversed in the accent across the
// picker, as a menu's item, a row's description in muted on the line under
// it; then a line of keys. While the actions panel is open, the selected
// row's actions take the rows' place.

// pickIndent is the columns before a row's text and after its keys.
const pickIndent = 1

// pickHead is how many rows a HottyPicker's lines above its rows take:
// the query and the rule, and the status while it has one.
func pickHead(e *view.Element) int {
	if e.Status != "" {
		return 3
	}
	return 2
}

// pickBody is how many rows a HottyPicker's rows take: as many lines as it
// has, at most its height, so that it keeps one height as it scrolls; a
// row for its empty text when it has none. The panel takes their place,
// and more rows only where its title and actions need them.
func pickBody(e *view.Element) int {
	n := max(min(len(e.PickLines()), e.Height), 1)
	if e.Panel {
		n = max(n, 1+len(e.SelectedPick().Actions))
	}
	return n
}

// pickHeight is a HottyPicker's height: its head, its rows, and the line
// of keys.
func pickHeight(e *view.Element) int { return pickHead(e) + pickBody(e) + 1 }

// pickWidth is a HottyPicker's natural width: its widest row, its keys
// as written (mod not yet read as ctrl or super).
func pickWidth(e *view.Element) int {
	n := max(Width(e.Chip)+1+max(Width(e.Placeholder), 20), Width(e.Status)+2, Width(e.NoRows))
	for _, row := range e.Rows {
		rw := Width(row.Title)
		if row.Group != "" {
			rw += Width(row.Group) + 2
		}
		if row.Detail != "" {
			rw += 2 + Width(row.Detail)
		}
		for i, a := range row.Actions {
			if i > 0 && a.Key != "" {
				rw += 2 + Width(view.KeyText(a.Key))
			}
		}
		n = max(n, rw, Width(row.Description))
	}
	return n + 2*pickIndent
}

// pickMinimum is a HottyPicker's minimum: a few columns of query.
func pickMinimum(e *view.Element) int { return Width(e.Chip) + 1 + 6 }

// pickRow is a row's text: its group, muted, and ": ", its title, its
// characters the query matched in the accent and bold, then its detail,
// muted, two columns on. In the selected row everything is reversed in the
// accent, and the matched characters underlined instead.
func (r *Rendition) pickRow(row view.PickRow, sel, focused bool) []glyph {
	base, dim, hit := style{}, style{role: Muted}, style{role: Accent, attr: Bold}
	if sel {
		base = r.pickSelected(focused)
		dim, hit = base, base
		hit.attr |= Bold | Underline
	}
	var gs []glyph
	if row.Group != "" {
		gs = line(row.Group+": ", dim)
	}
	at := 0
	for _, cl := range clusters(row.Title) {
		st := base
		n := len([]rune(cl))
		if slices.ContainsFunc(row.Match, func(m [2]int) bool { return at < m[1] && at+n > m[0] }) {
			st = hit
		}
		gs = append(gs, line(cl, st)...)
		at += n
	}
	if row.Detail != "" {
		gs = concat(gs, line("  "+row.Detail, dim))
	}
	return gs
}

// pickSelected is the selected row's style: reversed in the accent while
// the picker has the keyboard, reversed in muted otherwise.
func (r *Rendition) pickSelected(focused bool) style {
	if focused {
		return style{role: Accent, attr: Reverse}
	}
	return style{role: Muted, attr: Reverse}
}

// pickKeys are the keys of a row's actions past its first, as the key
// hints write them, two columns apart.
func (r *Rendition) pickKeys(row view.PickRow, st style) []glyph {
	var gs []glyph
	for i, a := range row.Actions {
		if i == 0 || a.Key == "" {
			continue
		}
		if gs != nil {
			gs = concat(gs, line("  ", st))
		}
		gs = concat(gs, line(r.keyText(a.Key), st))
	}
	return gs
}

// keyText is a key a program wrote, read with the Platform, as the key
// hints write it: mod+k is ctrl+k, or super+k on a Mac.
func (r *Rendition) keyText(key string) string {
	if k, ok := r.c.Platform.ReadKey(key); ok {
		key = k
	}
	return view.KeyText(key)
}

// paintPicker paints a HottyPicker at (x, y), w wide.
func (l *layout) paintPicker(cv *canvas, e *view.Element, x, y, w int) {
	r := l.r
	focused := r.focused(e.ID)
	in, iw := x+pickIndent, max(w-2*pickIndent, 0)
	l.paintQuery(cv, e, in, y, iw)
	row := y + 1
	cv.write(x, row, w, fit(line(strings.Repeat("─", w), style{role: Border}), w))
	row++
	if e.Status != "" {
		st, text := style{}, e.Status
		if e.StatusError {
			st, text = style{role: Error}, "✗ "+text
		}
		cv.write(in, row, iw, fit(line(text, st), iw))
		row++
	}
	r.hits = append(r.hits, hit{x: x, y: y + 1, w: w, h: row - y - 1, id: e.ID, opt: -1})
	if e.Panel {
		l.paintPickPanel(cv, e, x, row, w, focused)
	} else {
		l.paintPickRows(cv, e, x, row, w, focused)
	}
	row += pickBody(e)
	cv.write(in, row, iw, shortHints(r.pickHints(e), iw))
	r.boxes[e.ID] = box{x, y, w, row + 1 - y}
}

// paintPickRows paints the lines that show from Top: a heading in muted
// and bold, or a row, the selected one across the picker, and a row's
// description in muted under it, across the picker with it. A heading on
// the last line, with none of its rows under it, is left out. A row's
// other actions' keys show at its right where they fit beside its whole
// text; a narrow picker shows them on the selected row alone, so that the
// others' details aren't cut for keys the panel lists anyway.
func (l *layout) paintPickRows(cv *canvas, e *view.Element, x, y, w int, focused bool) {
	r := l.r
	in, iw := x+pickIndent, max(w-2*pickIndent, 0)
	lines := e.PickLines()
	if len(lines) == 0 {
		text := e.NoRows
		if e.Busy {
			text = "Loading…"
		}
		cv.write(in, y, iw, fit(line(text, style{role: Muted}), iw))
		return
	}
	for k := 0; k < e.Height && e.Top+k < len(lines); k++ {
		ln, at := lines[e.Top+k], y+k
		if ln.Row < 0 {
			if k < e.Height-1 {
				cv.write(in, at, iw, fit(line(ln.Heading, style{role: Muted, attr: Bold}), iw))
			}
			continue
		}
		row := e.Rows[ln.Row]
		sel := ln.Row == e.Selected
		keys := style{role: Muted}
		if sel {
			keys = r.pickSelected(focused)
			cv.write(x, at, w, line(strings.Repeat(" ", w), keys))
		}
		r.hits = append(r.hits, hit{x: x, y: at, w: w, h: 1, id: e.ID, opt: ln.Row})
		if ln.Desc {
			cv.write(in, at, iw, fit(line(row.Description, keys), iw))
			continue
		}
		// The detail wins: a row's keys show where they fit beside its
		// whole text, else on the selected row alone, where they cut it.
		ks, text := r.pickKeys(row, keys), r.pickRow(row, sel, focused)
		tw := iw
		if kw := width(ks); kw > 0 && kw+2 < iw && (sel || width(text)+2+kw <= iw) {
			cv.write(in+iw-kw, at, kw, ks)
			tw = iw - kw - 2
		}
		cv.write(in, at, tw, fit(text, tw))
	}
}

// paintPickPanel paints the selected row's actions in the rows' place:
// a title, then each action and its key, the highlighted one reversed in
// the accent across the picker.
func (l *layout) paintPickPanel(cv *canvas, e *view.Element, x, y, w int, focused bool) {
	r := l.r
	in, iw := x+pickIndent, max(w-2*pickIndent, 0)
	row := e.SelectedPick()
	cv.write(in, y, iw, fit(line("Actions: "+row.Title, style{role: Muted, attr: Bold}), iw))
	for i, a := range row.Actions {
		at := y + 1 + i
		st, keys := style{}, style{role: Muted}
		if i == e.PanelAt {
			st = r.pickSelected(focused)
			keys = st
			cv.write(x, at, w, line(strings.Repeat(" ", w), st))
		}
		key := a.Key
		if i == 0 {
			key = "enter"
		}
		ks := line(r.keyText(key), keys)
		tw := iw
		if kw := width(ks); kw > 0 && kw+2 < iw {
			cv.write(in+iw-kw, at, kw, ks)
			tw = iw - kw - 2
		}
		cv.write(in, at, tw, fit(line(a.Title, st), tw))
	}
}

// pickHints are the picker's line of keys: the arrows, then what Enter
// does to the selected row and the panel's key; in the panel, its own.
func (r *Rendition) pickHints(e *view.Element) []view.Hint {
	if e.Panel {
		return []view.Hint{{Key: "↑/↓", Desc: "move"}, {Key: "enter", Desc: "run"}, {Key: "esc", Desc: "back"}}
	}
	hs := []view.Hint{{Key: "↑/↓", Desc: "move"}}
	if enter, panel := e.PickFooter(); enter != "" {
		hs = append(hs, view.Hint{Key: "enter", Desc: strings.ToLower(enter)})
		if panel {
			hs = append(hs, view.Hint{Key: r.keyText(e.PanelKey), Desc: "actions"})
		}
	}
	return hs
}

// pickCaret puts a HottyPicker's caret at its query's end once a query is
// written to it other than as typed (view.Element.Written: Tab's
// completion, the agent's), as a field's caret is when its whole value is
// set.
func (r *Rendition) pickCaret(e *view.Element) {
	if r.written[e.ID] != e.Written {
		r.written[e.ID] = e.Written
		delete(r.cursor, e.ID)
		delete(r.fields, e.ID)
	}
}

// paintQuery paints a HottyPicker's query on its row, after its chip: a
// one-line field, as paintField paints one's value, without its
// underline, the placeholder faint while it is empty, and the terminal's
// cursor at its caret, never a block. It scrolls to keep the cursor in it
// while the picker has the keyboard.
func (l *layout) paintQuery(cv *canvas, e *view.Element, x, y, w int) {
	r := l.r
	focused := r.focused(e.ID)
	vx, vw := x, w
	if e.Chip != "" {
		n := cv.write(x, y, w, fit(line(e.Chip+" ", style{role: Accent, attr: Bold}), w))
		vx, vw = x+n, max(w-n, 0)
	}
	if e.Busy && len(e.Rows) > 0 && vw > 2 {
		cv.write(vx+vw-1, y, 1, line("…", style{role: Muted}))
		vw -= 2
	}
	r.pickCaret(e)
	v, _ := e.Value.(string)
	cl := clusters(v)
	r.rows[e.ID] = 1
	area := &fieldArea{x: vx, y: y, w: vw, rows: 1}
	s0, s1, curCol := 0, 0, 0
	if focused {
		pos := r.cursorOf(e.ID, len(cl))
		if f := r.fields[e.ID]; f != nil && f.Value == v && f.Caret == pos {
			s0, s1 = f.Selection()
		}
		curCol = colOf(cl, pos, false)
		hoff := min(r.hscroll[e.ID], curCol)
		if curCol >= hoff+vw {
			hoff = curCol - vw + 1
		}
		if colOf(cl, len(cl), false) < vw {
			hoff = 0
		}
		r.hscroll[e.ID], area.hoff = hoff, hoff
	}
	c := 0
	for i, g := range cl {
		gl := shownGlyph(g, false, style{})
		if i >= s0 && i < s1 {
			gl.style = selected
		}
		if c >= area.hoff && c+gl.width <= area.hoff+vw {
			cv.set(vx+c-area.hoff, y, gl)
		}
		c += gl.width
	}
	if v == "" {
		cv.write(vx, y, vw, fit(line(e.Placeholder, style{role: Muted, attr: Faint}), vw))
	}
	// The caret is the terminal's own cursor, a text input's, whatever the
	// keymap: the terminal blinks it, so it is always there, never drawn.
	if cx := vx + curCol - area.hoff; focused && cx < cv.f.Cols && y < cv.f.Rows {
		cv.cursorAt(cx, y, false)
	}
	r.hits = append(r.hits, hit{x: x, y: y, w: w, h: 1, id: e.ID, opt: -1, field: area})
}
