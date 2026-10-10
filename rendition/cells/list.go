package cells

import (
	"slices"

	"github.com/neuroplastio/hotty-a2ui/view"
)

// A HottyList in cells (profile §3.3, §3.4, §3.7), as bubbles' list draws
// one with its default delegate: the title, then a status line, then the
// items, a label and a description each, then a page's dots. Every line
// starts two columns in; the selected item's two columns are a "│" bar.

// listIndent is the columns before every line of a HottyList: the
// selected item's bar and a space.
const listIndent = 2

// listFilter is what a HottyList's filter line starts with.
const listFilter = "Filter: "

// itemRows is how many rows each of a HottyList's items takes: two when
// one has a description, else one; and the blank rows between them, one
// between items of two rows.
func itemRows(e *view.Element) (rows, gap int) {
	for _, it := range e.Items {
		if it.Description != "" {
			return 2, 1
		}
	}
	return 1, 0
}

// listHead is how many rows a HottyList's lines above its items take: its
// title (or its filter while typed) and a blank row, unless it has
// neither; then its status line and a blank row.
func listHead(e *view.Element) int {
	if e.Label != "" || e.Query.Editing {
		return 4
	}
	return 2
}

// listBody is how many rows a HottyList's items take: its height's worth,
// so that turning a page leaves it the same height, else as many as it
// shows; at least a row, for the empty text when it has none.
func listBody(e *view.Element) int {
	rows, gap := itemRows(e)
	n := len(e.Shown)
	if e.Height > 0 {
		n = e.Height
	}
	return max(n*rows+max(n-1, 0)*gap, 1)
}

// listPaged reports whether a HottyList keeps rows for its page's dots: a
// blank one and the dots. One with a height and more items than a page
// keeps them while a filter leaves fewer, so that it stays one height.
func listPaged(e *view.Element) bool { return e.Height > 0 && len(e.Items) > e.Height }

// listHeight is a HottyList's height.
func listHeight(e *view.Element) int {
	h := listHead(e) + listBody(e)
	if listPaged(e) {
		h += 2
	}
	return h
}

// listWidth is a HottyList's natural width: its widest line.
func listWidth(e *view.Element) int {
	n := max(Width(e.Label)+2, Width(e.ListStatus()), Width(e.Placeholder))
	if e.Filter {
		n = max(n, Width(listFilter)+1)
	}
	for _, it := range e.Items {
		n = max(n, Width(it.Label), Width(it.Description))
	}
	return listIndent + n
}

// listMinimum is a HottyList's minimum: its indent and a few columns.
func listMinimum(e *view.Element) int { return listIndent + 3 + 1 }

// marked is a label as glyphs, the clusters that hold a matched byte
// (view.Element.Matched) underlined.
func marked(s string, at []int, st style) []glyph {
	if len(at) == 0 {
		return line(s, st)
	}
	var out []glyph
	off := 0
	for _, cl := range clusters(s) {
		cst := st
		if slices.ContainsFunc(at, func(b int) bool { return b >= off && b < off+len(cl) }) {
			cst.attr |= Underline
		}
		out = append(out, line(cl, cst)...)
		off += len(cl)
	}
	return out
}

// paintList paints a HottyList at (x, y), w wide. The selected item's
// bar, label and description are in the accent while the list has the
// keyboard; otherwise its bar is in muted and its label bold, as a
// HottyTree's (the maintainer, round 6). A label's characters
// that matched the filter are underlined; the filter line, while typed,
// has the cursor at its end.
func (l *layout) paintList(cv *canvas, e *view.Element, x, y, w int) {
	r := l.r
	focused := r.focused(e.ID)
	in := x + listIndent
	iw := max(w-listIndent, 0)
	row := y
	switch {
	case e.Query.Editing:
		gs := concat(line(listFilter, style{role: Accent}), line(e.Query.Text, style{}))
		n := cv.write(in, row, iw, fit(gs, iw))
		if focused && n < iw && in+n < cv.f.Cols && row < cv.f.Rows && r.caretOn(e.ID) {
			cv.cursorAt(in+n, row, r.blockCursor())
		}
		row += 2
	case e.Label != "":
		cv.write(in, row, iw, fit(line(" "+e.Label+" ", style{role: Accent, attr: Reverse}), iw))
		row += 2
	}
	cv.write(in, row, iw, fit(line(e.ListStatus(), style{role: Muted}), iw))
	row += 2
	r.hits = append(r.hits, hit{x: x, y: y, w: w, h: row - y, id: e.ID, opt: -1})

	rows, gap := itemRows(e)
	end := len(e.Shown)
	if e.Height > 0 {
		end = min(e.Top+e.Height, end)
	}
	if len(e.Items) == 0 {
		cv.write(in, row, iw, fit(line(e.Placeholder, style{role: Muted}), iw))
		r.hits = append(r.hits, hit{x: x, y: row, w: w, h: 1, id: e.ID, opt: -1})
	}
	sel := e.SelectedRow()
	at := row
	for k := e.Top; k < end; k++ {
		i := e.Shown[k]
		it := e.Items[i]
		label, desc, bar := style{}, style{role: Muted}, style{}
		if i == sel {
			label, bar = style{attr: Bold}, style{role: Muted}
			if focused {
				label, desc, bar = style{role: Accent}, style{role: Accent}, style{role: Accent}
			}
		}
		var match []int
		if k < len(e.Matched) {
			match = e.Matched[k]
		}
		for j := range rows {
			if i == sel {
				cv.write(x, at+j, w, line("│", bar))
			}
		}
		cv.write(in, at, iw, fit(marked(it.Label, match, label), iw))
		if rows > 1 {
			cv.write(in, at+1, iw, fit(line(it.Description, desc), iw))
		}
		r.hits = append(r.hits, hit{x: x, y: at, w: w, h: rows, id: e.ID, opt: i})
		at += rows + gap
	}
	row += listBody(e)
	if listPaged(e) {
		// A HottyPaginator's dots (pageDots), as bubbles' list has its
		// paginator; numbers in muted where they do not fit.
		if pages, page := e.Pages(); pages > 1 {
			cv.write(in, row+1, iw, fit(pageDots(pages, page, iw, false, style{}, style{role: Muted}), iw))
		}
		row += 2
	}
	r.boxes[e.ID] = box{x, y, w, row - y}
}

// clickList is a click on a HottyList, which has the keyboard now: on an
// item, it selects it, or acts on it when it was selected already;
// elsewhere, it does nothing more.
func (r *Rendition) clickList(e *view.Element, item int) error {
	switch {
	case item < 0:
		return nil
	case item == e.SelectedRow():
		return r.c.Activate(e.ID)
	}
	return r.c.SelectRow(e.ID, item)
}
