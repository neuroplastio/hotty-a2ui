package cells

import (
	"strconv"

	"github.com/neuroplastio/hotty-a2ui/view"
)

// A text field's suggestions in cells (profile §3.4, §6.22): the rest of
// the suggestion Tab takes, faint after the value, as bubbles' text input
// shows it, and while the field has the keyboard, the suggestions its
// value leaves, a row each under it, as a select's open list is.

// ghostStyle is a suggestion's rest after the value: faint and muted, as
// a placeholder is, and as bubbles draws it (colour 240).
var ghostStyle = style{role: Muted, attr: Faint}

// suggestRows is how many rows a text field's list of suggestions takes
// now: none unless it has the keyboard and its value leaves some; else a
// row each, at most its Height, and one more saying which show when more
// than that do.
func (r *Rendition) suggestRows(e *view.Element) int {
	n := len(e.Shown)
	if n == 0 || !r.focused(e.ID) || isLongText(e) {
		return 0
	}
	if n > e.Height {
		return e.Height + 1
	}
	return n
}

// paintGhost paints the rest of a text field's suggestion after its
// value, faint, from the value's end on its row, cut at the field's edge.
func paintGhost(cv *canvas, e *view.Element, a *fieldArea, vw int, end int) {
	ghost := e.Ghost()
	gx := a.x + end - a.hoff
	if ghost == "" || gx < a.x || gx >= a.x+vw {
		return
	}
	cv.write(gx, a.y, a.x+vw-gx, line(ghost, ghostStyle))
}

// paintSuggestions paints a text field's list of suggestions from row y,
// past its gutter, as wide as its value's prompt and value: a row each,
// its text under the value's, what the value typed of it plain and its
// rest in muted, as the ghost is. The highlighted one is reversed in the
// accent, a column of padding either side, as a menu's item: a "> " as
// the select's open list has would read as a second prompt under the
// field's. Past its Height, the rows that show from Top, then "  1–5 of
// 12" in muted. Each row takes a click, which picks it. It returns the
// rows it painted.
func (l *layout) paintSuggestions(cv *canvas, e *view.Element, x, y, w int) int {
	rows := l.r.suggestRows(e)
	if rows == 0 {
		return 0
	}
	n := len(e.Shown)
	end := min(e.Top+e.Height, n)
	row := y
	for i := e.Top; i < end; i++ {
		typed, rest := e.SuggestionParts(i)
		gs := concat(glyphs("  ", style{}), line(typed, style{}), line(rest, style{role: Muted}))
		if i == e.Selected {
			hi := style{role: Accent, attr: Reverse}
			gs = concat(glyphs(" ", style{}), line(" "+typed+rest+" ", hi))
		}
		cv.write(x, row, w, fit(gs, w))
		l.r.hits = append(l.r.hits, hit{x: x - gutter, y: row, w: w + gutter, h: 1, id: e.ID, opt: i})
		row++
	}
	if n > e.Height {
		count := "  " + strconv.Itoa(e.Top+1) + "–" + strconv.Itoa(end) + " of " + strconv.Itoa(n)
		cv.write(x, row, w, fit(line(count, style{role: Muted}), w))
		row++
	}
	return row - y
}

// suggestKey gives a key to the focused text field's suggestions first
// (view.Controller.SuggestKey): ArrowRight takes one only at the value's
// end with nothing selected, where it would not move. A taken suggestion
// puts the caret at the end of the new value.
func (r *Rendition) suggestKey(e *view.Element, key string) (bool, error) {
	if e.Suggestions == nil {
		return false, nil
	}
	v, _ := e.Value.(string)
	n := len(clusters(v))
	f := r.field(e)
	s0, s1 := f.Selection()
	ok, err := r.c.SuggestKey(e.ID, key, f.Caret == n && s0 == s1)
	if now := r.c.V.Find(e.ID); ok && now != nil && now.Value != e.Value {
		delete(r.cursor, e.ID)
	}
	return ok, err
}
