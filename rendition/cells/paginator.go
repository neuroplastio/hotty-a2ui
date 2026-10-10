package cells

import (
	"strconv"

	"github.com/neuroplastio/hotty-a2ui/view"
)

// A HottyPaginator in cells (profile §3.3, §3.4, §3.7), as bubbles'
// paginator draws its pages under what it pages: its child, past a field's
// gutter, then a blank row and its dots; without a child, the dots alone.
// The gutter's bar shows while it has the keyboard, as a field's does.

// pageDot is a page's dot, as bubbles' paginator and list draw it.
const pageDot = "•"

// pageDots are a paginator's pages as drawn, w wide at most, as bubbles'
// list draws its paginator: a dot a page, the page shown's in cur and the
// others in border and faint, bubbles' "very subdued" colour, as the help
// line's dots between keys are (and under NO_COLOR the page shown is
// still the one dot not faint); or the page shown of how many, "3/10", in
// num, with numbers or where the dots do not fit. A HottyList's dots are
// these.
func pageDots(pages, page, w int, numbers bool, cur, num style) []glyph {
	if numbers || pages > w {
		return line(strconv.Itoa(page+1)+"/"+strconv.Itoa(pages), num)
	}
	gs := make([]glyph, pages)
	for p := range gs {
		st := style{role: Border, attr: Faint}
		if p == page {
			st = cur
		}
		gs[p] = glyph{text: pageDot, width: 1, style: st}
	}
	return gs
}

// numbersWidth is the most columns a paginator's numbers take: its last
// page of its pages, so that they stay put as the pages turn.
func numbersWidth(pages int) int { return 2*len(strconv.Itoa(pages)) + 1 }

// pagerChild is a HottyPaginator's child, nil without one.
func pagerChild(e *view.Element) *view.Element {
	if kids := shown(e.Children); len(kids) > 0 {
		return kids[0]
	}
	return nil
}

// pagerPages are a HottyPaginator's child as it stands for each of its
// pages, for measuring: the child itself for the page shown, and for each
// other a copy holding that page's items. Nil without a child.
func (l *layout) pagerPages(e *view.Element) []*view.Element {
	c := pagerChild(e)
	if c == nil {
		return nil
	}
	if ps, ok := l.pages[e]; ok {
		return ps
	}
	ps := []*view.Element{c}
	if c.Kind == view.Stack {
		ps = make([]*view.Element, len(e.Paged))
		for i, items := range e.Paged {
			if i == e.Selected {
				ps[i] = c
				continue
			}
			cp := *c
			cp.Children = items
			ps[i] = &cp
		}
	}
	l.pages[e] = ps
	return ps
}

// pagerWidth is a HottyPaginator's natural width: the gutter, and its
// widest page's or its dots, whichever is wider.
func (l *layout) pagerWidth(e *view.Element) int {
	n := e.PageCount
	if e.Variant == "numbers" {
		n = numbersWidth(e.PageCount)
	}
	for _, p := range l.pagerPages(e) {
		n = max(n, l.natural(p))
	}
	return gutter + n
}

// pagerMinimum is the narrowest a HottyPaginator gets: the gutter, and its
// narrowest page's minimum or its numbers, which its dots become.
func (l *layout) pagerMinimum(e *view.Element) int {
	n := numbersWidth(e.PageCount)
	for _, p := range l.pagerPages(e) {
		n = max(n, l.minimum(p))
	}
	return gutter + n
}

// pagerBody is the rows a HottyPaginator's child takes at width w past
// the gutter: its tallest page's, so that it keeps its height as the pages
// turn, the last page's shorter too.
func (l *layout) pagerBody(e *view.Element, w int) int {
	h := 0
	for _, p := range l.pagerPages(e) {
		h = max(h, l.height(p, w))
	}
	return h
}

// pagerHeight is a HottyPaginator's height at width w: its child's rows
// and a blank row, then its dots.
func (l *layout) pagerHeight(e *view.Element, w int) int {
	if pagerChild(e) == nil {
		return 1
	}
	return l.pagerBody(e, max(w-gutter, 1)) + 2
}

// paintPaginator paints a HottyPaginator at (x, y), w wide and h tall: its
// child's page past the gutter, then its dots, which take clicks, a dot a
// page. While it has the keyboard, the gutter is "┃" in the accent down
// its rows, and the page shown is in the accent too. The whole box takes
// a click under its child's, so that a click on the page gives it the
// keyboard.
func (l *layout) paintPaginator(cv *canvas, e *view.Element, x, y, w, h int) {
	r := l.r
	if w <= gutter {
		return
	}
	focused := r.focused(e.ID)
	cx, cw, row := x+gutter, w-gutter, y
	r.hits = append(r.hits, hit{x: x, y: y, w: w, h: h, id: e.ID, opt: -1})
	if c := pagerChild(e); c != nil {
		l.paint(cv, c, cx, y, cw, l.height(c, cw))
		row += l.pagerBody(e, cw) + 1
	}
	cur, num := style{}, style{}
	if focused {
		cur, num = style{role: Accent}, style{role: Accent}
	}
	dots := fit(pageDots(e.PageCount, e.Selected, cw, e.Variant == "numbers", cur, num), cw)
	n := cv.write(cx, row, cw, dots)
	r.boxes[e.ID] = box{x, y, w, row - y + 1}
	if e.Variant != "numbers" && e.PageCount <= cw {
		for p := range e.PageCount {
			r.hits = append(r.hits, hit{x: cx + p, y: row, w: 1, h: 1, id: e.ID, opt: p})
		}
	} else {
		r.hits = append(r.hits, hit{x: x, y: row, w: gutter + n, h: 1, id: e.ID, opt: -1})
	}
	if focused {
		for j := y; j <= row; j++ {
			cv.set(x, j, glyph{text: "┃", width: 1, style: style{role: Accent}})
		}
	}
}

// clickPager is a click on a HottyPaginator, which has the keyboard now: on
// a dot, it shows that page.
func (r *Rendition) clickPager(e *view.Element, page int) error {
	if page < 0 {
		return nil
	}
	return r.c.TurnPage(e.ID, page)
}

// dotCell is the cell of a HottyPaginator's dot for page (from 1) in the
// last frame drawn; ok is false when its dots were not drawn.
func (r *Rendition) dotCell(id string, page int) (col, row int, ok bool) {
	for _, h := range r.hits {
		if h.id == id && h.opt == page-1 {
			return h.x, h.y, true
		}
	}
	return 0, 0, false
}
