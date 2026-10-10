package cells

import (
	"strconv"

	"github.com/neuroplastio/hotty-a2ui/view"
)

// A HottyTable in cells (profile §3.3, §3.4, §3.7), as bubbles' table
// draws one: a header row and a rule, then a row of cells for each row,
// each cell padded a column on each side. One row is selected; the body
// scrolls under the header to keep it in view.

// cellPad is the columns of padding on each side of a cell: two between
// neighbours, as bubbles' table has it.
const cellPad = 1

// minColumn is the narrowest a column gets before the narrowest columns
// give way too: two characters and "…".
const minColumn = 3

// columnWidths are a Table's columns' natural widths: the width it sets,
// else its header's or its widest cell's, whichever is wider.
func columnWidths(e *view.Element) []int {
	ws := make([]int, len(e.Columns))
	for j, c := range e.Columns {
		if c.Width > 0 {
			ws[j] = c.Width
			continue
		}
		ws[j] = Width(c.Header)
		for _, row := range e.Cells {
			ws[j] = max(ws[j], Width(row[j]))
		}
	}
	return ws
}

// tableWidth is the columns a Table's cells take at the given widths.
func tableWidth(ws []int) int {
	n := 0
	for _, w := range ws {
		n += w + 2*cellPad
	}
	return n
}

// tableBar is the columns a Table's scrollbar takes after its last
// column's padding: while its body scrolls, a blank one and the bar, as a
// scroll view's (scrollBar), so that the selected row's fill ends apart
// from the thumb; none otherwise.
func tableBar(e *view.Element) int {
	if len(e.Cells) > bodyRows(e) {
		return scrollBar
	}
	return 0
}

// tableNatural is a Table's natural width: its columns at their natural
// widths, and its scrollbar.
func tableNatural(e *view.Element) int {
	return tableWidth(columnWidths(e)) + tableBar(e)
}

// tableMinimum is a Table's minimum: each column at most minColumn wide,
// and its scrollbar.
func tableMinimum(e *view.Element) int {
	ws := columnWidths(e)
	for j := range ws {
		ws[j] = min(ws[j], minColumn)
	}
	return tableWidth(ws) + tableBar(e)
}

// fitColumns are a Table's columns' widths in w columns (fitWidths).
func fitColumns(e *view.Element, w int) []int { return fitWidths(columnWidths(e), w) }

// fitWidths are columns' widths in w columns, from their natural widths
// ws: less a column at a time off the widest that is wider than minColumn
// (the first of equals) while the table is too wide, and once none is,
// off the widest. A column may end with no width; it is not drawn. A
// HottyMarkdown's tables are fitted so too (docTable).
func fitWidths(ws []int, w int) []int {
	ws = append([]int(nil), ws...)
	for tableWidth(ws) > w {
		j := -1
		for i := range ws {
			if ws[i] > minColumn && (j < 0 || ws[i] > ws[j]) {
				j = i
			}
		}
		if j < 0 {
			for i := range ws {
				if ws[i] > 0 && (j < 0 || ws[i] > ws[j]) {
					j = i
				}
			}
		}
		if j < 0 {
			break
		}
		ws[j]--
	}
	return ws
}

// bodyRows is how many rows a Table's body takes: its height, else one a
// row, and one for "No rows" when it has none.
func bodyRows(e *view.Element) int {
	if e.Height > 0 {
		return e.Height
	}
	return max(len(e.Cells), 1)
}

// aligned is a cell's text in w columns: cut with "…" when wider, else
// placed at the start, the centre (rounded down) or the end.
func aligned(gs []glyph, w int, align string, st style) []glyph {
	gs = fit(gs, w)
	spare := w - width(gs)
	left := 0
	switch align {
	case "center":
		left = spare / 2
	case "end":
		left = spare
	}
	return concat(repeat(" ", left, st), gs, repeat(" ", spare-left, st))
}

// tableRow is one row of a Table as drawn: each cell padded, in st.
func tableRow(cells []string, cols []view.Column, ws []int, st style) []glyph {
	var out []glyph
	for j, w := range ws {
		if w <= 0 {
			continue
		}
		pad := repeat(" ", cellPad, st)
		out = concat(out, pad, aligned(line(cells[j], st), w, cols[j].Align, st), pad)
	}
	return out
}

// paintTable paints a Table at (x, y), w wide: the header, bold; a rule
// in border, which says which rows show while the body scrolls; then the
// rows from the first shown, and while it scrolls a scroll view's bar
// beside them, past the last column's padding. The selected row is
// reversed across the table, in the accent while the table has the
// keyboard and in muted otherwise.
func (l *layout) paintTable(cv *canvas, e *view.Element, x, y, w int) {
	r := l.r
	bar := min(tableBar(e), w)
	ws := fitColumns(e, w-bar)
	tw := min(tableWidth(ws), w-bar)
	headers := make([]string, len(e.Columns))
	for j, c := range e.Columns {
		headers[j] = c.Header
	}
	cv.write(x, y, w, tableRow(headers, e.Columns, ws, style{attr: Bold}))
	h, n := bodyRows(e), len(e.Cells)
	top := e.Top
	rule := repeat("─", tw+bar, style{role: Border})
	if n > h {
		at := " " + strconv.Itoa(top+1) + "–" + strconv.Itoa(min(top+h, n)) + " of " + strconv.Itoa(n) + " "
		if aw := Width(at); aw+2 <= tw+bar {
			rule = concat(repeat("─", tw+bar-aw-1, style{role: Border}), line(at, style{role: Muted}), repeat("─", 1, style{role: Border}))
		}
	}
	cv.write(x, y+1, w, rule)
	if bar > 0 {
		l.paintScrollbar(cv, e, x+tw+bar-1, y+2, h, n, top)
	}
	r.hits = append(r.hits, hit{x: x, y: y, w: tw, h: 2, id: e.ID, opt: -1})
	if n == 0 {
		cv.write(x+cellPad, y+2, w-cellPad, fit(line("No rows", style{role: Muted}), w-cellPad))
		r.hits = append(r.hits, hit{x: x, y: y + 2, w: tw, h: 1, id: e.ID, opt: -1})
	}
	sel := e.SelectedRow()
	for i := top; i < min(top+h, n); i++ {
		st := style{}
		if i == sel {
			st = style{role: Muted, attr: Reverse}
			if r.focused(e.ID) {
				st.role = Accent
			}
		}
		row := tableRow(e.Cells[i], e.Columns, ws, st)
		cv.write(x, y+2+i-top, w, row)
		r.hits = append(r.hits, hit{x: x, y: y + 2 + i - top, w: tw, h: 1, id: e.ID, opt: i})
	}
	r.boxes[e.ID] = box{x, y, tw + bar, 2 + h}
}

// clickTable is a click on a Table, which has the keyboard now: on a row,
// it selects it, or acts on it when it was selected already; on the
// header, it does nothing more.
func (r *Rendition) clickTable(e *view.Element, row int) error {
	switch {
	case row < 0:
		return nil
	case row == e.SelectedRow():
		return r.c.Activate(e.ID)
	}
	return r.c.SelectRow(e.ID, row)
}
