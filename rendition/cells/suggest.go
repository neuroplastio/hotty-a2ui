package cells

import (
	"strconv"

	"github.com/neuroplastio/hotty-a2ui/view"
)

// A text field's suggestions in cells (profile §3.4, §6.22): the rest of
// the suggestion Tab takes, faint after the value, as bubbles' text input
// shows it, and while the field has the keyboard, the suggestions its
// value leaves, a row each in a box over what is under the field, as a
// menu drops down over a GUI's form. With `list` false, the ghost alone.

// ghostStyle is a suggestion's rest after the value: faint and muted, as
// a placeholder is, and as bubbles draws it (colour 240).
var ghostStyle = style{role: Muted, attr: Faint}

// suggestRows is how many rows a text field's list of suggestions shows
// now: none unless it has the keyboard, shows a list, and its value
// leaves some; else a row each, at most its Height, and one more saying
// which show when more than that do.
func (r *Rendition) suggestRows(e *view.Element) int {
	n := len(e.Shown)
	if n == 0 || !r.focused(e.ID) || isLongText(e) || e.GhostOnly {
		return 0
	}
	if n > e.Height {
		return e.Height + 1
	}
	return n
}

// suggestPlace is where the focused text field's list of suggestions goes
// over the frame: from row y, under the input, its text in the value
// text's column x.
type suggestPlace struct {
	e    *view.Element
	x, y int
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

// paintSuggestions paints the focused text field's list of suggestions
// over the frame, as an open Modal's panel and a toast are: it moves no
// row, hides what it covers and takes its clicks. It is a rounded box in
// border right under the input, a column of padding inside it, so
// that its text stands under the value's: a row each, what the value
// typed of it plain and its rest in muted, as the ghost is. The
// highlighted one is reversed in the accent across the box, as a menu's
// item: a "> " as the select's open list has would read as a second
// prompt under the field's. Past its Height, the rows that show from Top,
// then "1–5 of 12" in muted. It is as wide as its widest row needs, at
// most the frame, moved left to fit, and the frame grows to hold it. Each
// row takes a click, which picks it.
func (l *layout) paintSuggestions(cv *canvas, p *suggestPlace) {
	e := p.e
	if l.r.suggestRows(e) == 0 {
		return
	}
	n := len(e.Shown)
	end := min(e.Top+e.Height, n)
	var faces [][]glyph
	for i := e.Top; i < end; i++ {
		typed, rest := e.SuggestionParts(i)
		faces = append(faces, concat(line(typed, style{}), line(rest, style{role: Muted})))
	}
	var count []glyph
	if n > e.Height {
		count = line(strconv.Itoa(e.Top+1)+"–"+strconv.Itoa(end)+" of "+strconv.Itoa(n), style{role: Muted})
	}
	tw := width(count)
	for _, gs := range faces {
		tw = max(tw, width(gs))
	}
	f := cv.f
	bw := min(tw+4, f.Cols)
	tw = bw - 4
	bh := len(faces) + 2
	if count != nil {
		bh++
	}
	bx := max(min(p.x-2, f.Cols-bw), 0)
	by := p.y
	f.grow(by + bh)
	cv.fill(bx, by, bw, bh)
	cv.box(bx, by, bw, bh)
	l.r.hits = append(l.r.hits, hit{x: bx, y: by, w: bw, h: bh, id: e.ID, opt: -1})
	for k, gs := range faces {
		i, y := e.Top+k, by+1+k
		if i == e.Selected {
			hi := style{role: Accent, attr: Reverse}
			cv.write(bx+1, y, bw-2, concat(glyphs(" ", hi), fit(line(e.Suggestions[e.Shown[i]], hi), tw), repeat(" ", bw-3, hi)))
		} else {
			cv.write(bx+2, y, tw, fit(gs, tw))
		}
		l.r.hits = append(l.r.hits, hit{x: bx + 1, y: y, w: bw - 2, h: 1, id: e.ID, opt: i})
	}
	if count != nil {
		cv.write(bx+2, by+1+len(faces), tw, fit(count, tw))
	}
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
