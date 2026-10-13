package cells

import (
	"encoding/base64"
	"strings"
)

// Selecting a Text's or a HottyMarkdown's text with the pointer, and
// copying it (profile §3.7, vault KIT-12c), as a page's text is selected:
// a press on a cell of one that no click acts on, then a drag, selects
// from there to the pointer, in the element's reading order, never across
// screen rows (a page's nav beside it stays out); the release copies it
// (Copied). Its cells are drawn in selection.

// textRow is a row of a Text as the last Draw painted it: where, the
// text each column copies as (a cluster at its first column, "" at a wide
// one's second and on its frame, glyph.deco), and how it goes on from the
// row before.
type textRow struct {
	x, y int
	cols []string
	cont cont
}

// textPos is a cell of a Text's rows: a row of them and a column of it.
type textPos struct{ row, col int }

func (p textPos) before(q textPos) bool {
	return p.row < q.row || p.row == q.row && p.col < q.col
}

// textSel is a selection in a Text: from its anchor, where the press was,
// to its head, where the pointer is; held while the button is down. It is
// a selection once the head has left the anchor's cell, or, made by a
// double or a triple click (whole), from the first: a word of one cell.
type textSel struct {
	id    string
	a, b  textPos
	held  bool
	whole bool
}

// empty reports whether a selection selects nothing: none, or a press that
// has not left its cell.
func (s *textSel) empty() bool { return s == nil || s.a == s.b && !s.whole }

// span is the selection's first and last cells, in reading order.
func (s *textSel) span() (lo, hi textPos) {
	if s.b.before(s.a) {
		return s.b, s.a
	}
	return s.a, s.b
}

// cols are the columns of a row a selection takes, from c0 to c1, and
// whether it takes any: none of its frame before its text (a quote's
// bar), as a page's selection leaves a quote's border.
func (s *textSel) cols(i int, row textRow) (c0, c1 int, ok bool) {
	if s.empty() {
		return 0, 0, false
	}
	lo, hi := s.span()
	if i < lo.row || i > hi.row {
		return 0, 0, false
	}
	first, _ := textExtent(row.cols)
	c0, c1 = first, len(row.cols)-1
	if i == lo.row {
		c0 = max(lo.col, first)
	}
	if i == hi.row {
		c1 = min(hi.col, c1)
	}
	return c0, c1, c0 <= c1
}

// paintText paints a row of a Text (or a HottyMarkdown) at (x, y), w wide,
// and keeps it for selecting: its cells in a selection are drawn in the
// selection's colour, reversed where that is not known, as a field's.
func (l *layout) paintText(cv *canvas, id string, x, y, w int, gs []glyph, c cont) {
	cv.write(x, y, w, gs)
	r := l.r
	var cols []string
	for _, g := range gs {
		if len(cols)+g.width > w {
			break
		}
		t := g.text
		if g.deco {
			t = ""
		}
		cols = append(cols, t)
		for range g.width - 1 {
			cols = append(cols, "")
		}
	}
	row := textRow{x: x, y: y, cols: cols, cont: c}
	if c0, c1, ok := r.sel.cols(len(r.texts[id]), row); ok && r.sel.id == id {
		cv.restyle(x+c0, y, c1-c0+1, 1, func(c *Cell) {
			c.Back, c.BackMix, c.BackAttr = selected.back, selected.backMix, selected.backAttr
		})
	}
	r.texts[id] = append(r.texts[id], row)
}

// textAt is the Text whose text is at a cell of the last frame, and the
// place there: on a column of a row of it, past its frame (not a quote's
// bar) and before the spaces it ends in.
func (r *Rendition) textAt(col, row int) (id string, at textPos, ok bool) {
	for id, rows := range r.texts {
		for i, t := range rows {
			if t.y != row || col < t.x {
				continue
			}
			c := col - t.x
			first, end := textExtent(t.cols)
			if c >= first && c < end {
				return id, textPos{i, c}, true
			}
		}
	}
	return "", textPos{}, false
}

// textExtent is the columns of a row with text: from its first that is
// not its frame to its last that is not a space, end exclusive.
func textExtent(cols []string) (first, end int) {
	first = len(cols)
	for i, t := range cols {
		if t != "" {
			first = i
			break
		}
	}
	end = len(cols)
	for end > first && (cols[end-1] == " " || cols[end-1] == "") {
		end--
	}
	return first, end
}

// textNear is the place in a Text's rows nearest a cell, as a drag that
// left them selects: above its first row its start, below its last its
// end; on a row of it, left of the row its start, right of it its end.
func (r *Rendition) textNear(id string, col, row int) textPos {
	rows := r.texts[id]
	if len(rows) == 0 {
		return textPos{}
	}
	if row < rows[0].y {
		return textPos{0, 0}
	}
	i := len(rows) - 1
	for k, t := range rows {
		if t.y >= row {
			i = k
			break
		}
	}
	t := rows[i]
	if row > t.y {
		return textPos{i, max(len(t.cols)-1, 0)}
	}
	return textPos{i, min(max(col-t.x, 0), max(len(t.cols)-1, 0))}
}

// arm starts a selection at a press on a cell of a Text that no click
// acts on (not a link in it, nor a field over it); any other press drops
// the selection there was. The second press of a run selects the word
// there instead, and a third its line (selectWords).
func (r *Rendition) arm(col, row int, h *hit) {
	r.sel = nil
	if h != nil && (h.field != nil || r.clickable(h)) {
		return
	}
	if id, at, ok := r.textAt(col, row); ok {
		r.sel = &textSel{id: id, a: at, b: at, held: true}
		if r.count >= 2 {
			r.selectWords(id, at, r.count)
		}
	}
}

// selectWords selects in a Text what the n-th press of a run (n ≥ 2) at
// at selects, as a page's text takes a double and a triple click, and
// copies it at once (Copied): the word there (wordAt, as a field's), or
// on a third press and after, its line, the rows a soft wrap made of one,
// a paragraph. The button held drags nothing.
func (r *Rendition) selectWords(id string, at textPos, n int) {
	rows := r.texts[id]
	first, last := at.row, at.row
	for first > 0 && rows[first].cont != newLine {
		first--
	}
	for last+1 < len(rows) && rows[last+1].cont != newLine {
		last++
	}
	// The line's clusters, as a copy joins its rows, and the cell of each:
	// a space a wrap dropped has none, and takes the next row's first.
	var line strings.Builder
	var cells []textPos
	p := -1
	for i := first; i <= last; i++ {
		t := rows[i]
		c0, end := textExtent(t.cols)
		if i > first && t.cont == wrapSpace {
			line.WriteString(" ")
			cells = append(cells, textPos{i, c0})
		}
		for c := c0; c < end; c++ {
			if t.cols[c] == "" {
				continue
			}
			if i == at.row && c <= at.col {
				p = len(cells)
			}
			line.WriteString(t.cols[c])
			cells = append(cells, textPos{i, c})
		}
	}
	if p < 0 || len(cells) == 0 {
		return
	}
	start, end := 0, len(cells)
	if n == 2 {
		start, end = wordAt(line.String(), p, false)
	}
	if end <= start {
		return
	}
	r.sel = &textSel{id: id, a: cells[start], b: cells[end-1], whole: true}
	r.copied, r.copy = r.selText(), true
}

// extend moves a held selection's head to the pointer, kept in its Text.
func (r *Rendition) extend(col, row int) bool {
	s := r.sel
	if s == nil || !s.held {
		return false
	}
	s.b = r.textNear(s.id, col, row)
	return true
}

// letGo ends a held selection: a press that did not leave its cell
// selects nothing; else its text is copied (Copied).
func (r *Rendition) letGo() {
	s := r.sel
	if s == nil || !s.held {
		return
	}
	s.held = false
	if s.empty() {
		r.sel = nil
		return
	}
	r.copied, r.copy = r.selText(), true
}

// selText is a selection's text: each row's columns it takes, the spaces
// a row ends in dropped, a row that wraps the one before joined to it as
// the line it was (a space where the wrap dropped one), any other on a
// line of its own, so that a blank row between paragraphs is a blank line.
func (r *Rendition) selText() string {
	s := r.sel
	if s == nil {
		return ""
	}
	var b strings.Builder
	lo, hi := s.span()
	rows := r.texts[s.id]
	for i := lo.row; i <= hi.row && i < len(rows); i++ {
		t := rows[i]
		if i > lo.row {
			b.WriteString(t.cont.joiner())
		}
		if c0, c1, ok := s.cols(i, t); ok {
			b.WriteString(strings.TrimRight(strings.Join(t.cols[c0:c1+1], ""), " "))
		}
	}
	return b.String()
}

// Copied is the text a selection copied when the button was let go, once:
// ok is false until another selection ends. A program puts it on the
// clipboard, with OSC52 for the terminal's.
func (r *Rendition) Copied() (text string, ok bool) {
	text, ok = r.copied, r.copy
	r.copied, r.copy = "", false
	return text, ok
}

// Selection is the text selected now, "" for none.
func (r *Rendition) Selection() string {
	if r.sel.empty() {
		return ""
	}
	return r.selText()
}

// OSC52 is the sequence that puts text on the terminal's clipboard (OSC
// 52, "c", the clipboard), which a terminal may ask the user about or
// refuse.
func OSC52(text string) string {
	return "\x1b]52;c;" + base64.StdEncoding.EncodeToString([]byte(text)) + "\x1b\\"
}
