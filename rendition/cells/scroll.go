package cells

import (
	"slices"

	"github.com/neuroplastio/hotty-a2ui/view"
)

// A HottyScrollView in cells (profile §3.3, §3.4, §3.7), as bubbles'
// viewport: Height rows of its content from its Top, then a blank column
// and a scrollbar. Lines that do not wrap are cut at the width, from its
// Left; a child is laid out at the width and painted shifted up, through
// the box (canvas.clip). Where it is scrolled is the controller's state
// (view.Controller.Scrolled), clamped here to the content as last drawn.

const (
	scrollBar = 2 // the columns right of the content: a blank one and the bar
	wheelRows = 3 // the rows a notch of the wheel scrolls, as bubbles' viewport
	sideCols  = 6 // the columns ← and → scroll, as bubbles' viewport
)

// scrolled is a HottyScrollView as last drawn: its box, the rows it shows,
// its content's rows and the columns of it shown, its widest line (lines
// that do not wrap; 0 otherwise: wider than cols, they scroll sideways),
// and where it was scrolled to.
type scrolled struct {
	at               box
	page, rows, cols int
	wide             int
	top, left        int
}

// scrollChild is a HottyScrollView's child, nil for lines.
func scrollChild(e *view.Element) *view.Element {
	if kids := shown(e.Children); len(kids) > 0 {
		return kids[0]
	}
	return nil
}

// scrollLines are a HottyScrollView's lines as rows w columns wide: a row a
// line, or with Wrap as many as each takes, cut between clusters.
func scrollLines(e *view.Element, w int) [][]glyph {
	rows := make([][]glyph, 0, len(e.Lines))
	for _, s := range e.Lines {
		gs := line(s, style{})
		if !e.Wrap {
			rows = append(rows, gs)
			continue
		}
		for {
			n, i := 0, 0
			for ; i < len(gs) && n+gs[i].width <= max(w, 1); i++ {
				n += gs[i].width
			}
			rows = append(rows, gs[:i])
			if gs = gs[i:]; len(gs) == 0 {
				break
			}
		}
	}
	return rows
}

// scrollWidth is a HottyScrollView's natural width: its widest line
// unwrapped, wrap or not, or its child's, and the scrollbar's columns.
func (l *layout) scrollWidth(e *view.Element) int {
	if k := scrollChild(e); k != nil {
		return l.natural(k) + scrollBar
	}
	n := 0
	for _, s := range e.Lines {
		n = max(n, Width(s))
	}
	return n + scrollBar
}

// scrollMinimum is the narrowest a HottyScrollView gets: its child's, or
// a column of its lines, and the scrollbar's columns.
func (l *layout) scrollMinimum(e *view.Element) int {
	if k := scrollChild(e); k != nil {
		return l.minimum(k) + scrollBar
	}
	return 1 + scrollBar
}

// paintScroll paints a HottyScrollView at (x, y), w wide.
func (l *layout) paintScroll(cv *canvas, e *view.Element, x, y, w int) {
	r := l.r
	cw, page := max(w-scrollBar, 1), e.Height
	k := scrollChild(e)
	var rows [][]glyph
	total, wide := 0, 0
	if k != nil {
		total = l.height(k, cw)
	} else {
		rows = scrollLines(e, cw)
		total = len(rows)
		if !e.Wrap {
			for _, row := range rows {
				wide = max(wide, width(row))
			}
		}
	}
	last := max(total-page, 0)
	top := min(max(e.Top, 0), last)
	if e.Tail {
		top = last
	}
	if k != nil {
		top = l.revealFocus(e, k, cw, total, top, page)
	}
	left := min(max(e.Left, 0), max(wide-cw, 0))
	// At its end it follows what comes, with follow; and a hottyScrollTo
	// to its end without follow has done its work.
	tail := e.Active && top >= last
	if top != e.Top || left != e.Left || tail != e.Tail {
		r.c.Scrolled(e.ID, top, left, tail)
	}
	r.scrolls[e.ID] = scrolled{at: box{x, y, w, page}, page: page, rows: total, cols: cw, wide: wide, top: top, left: left}
	r.scrollOrder = append(r.scrollOrder, e.ID)
	// The box itself takes a click, under whatever the content has.
	r.hits = append(r.hits, hit{x: x, y: y, w: w, h: page, id: e.ID, opt: -1})
	if k != nil {
		was := cv.clipTo(box{x, y, cw, page})
		from := len(r.hits)
		l.paint(cv, k, x, y-top, cw, total)
		cv.clip = was
		r.clipHits(from, k, box{x, y, cw, page})
	} else {
		for i := range page {
			if top+i < len(rows) {
				cv.write(x, y+i, cw, skipCols(rows[top+i], left))
			}
		}
	}
	l.paintScrollbar(cv, e, x+w-1, y, page, total, top)
}

// revealFocus is the top that keeps the element with the keyboard in
// sight, when it is in the scroll view's child: the top as it was, moved
// as little as shows it, its start where it is taller than the page. The
// child is laid out once more, off the frame, to find it.
func (l *layout) revealFocus(e, k *view.Element, cw, total, top, page int) int {
	r := l.r
	c := r.c
	if !c.St.Keyboard || c.St.Focus == e.ID || !slices.ContainsFunc(c.Ancestors(c.St.Focus), func(a *view.Element) bool { return a.ID == e.ID }) {
		return top
	}
	hits := len(r.hits)
	l.paint(&canvas{f: newFrame(cw, total)}, k, 0, 0, cw, total)
	r.hits = r.hits[:hits]
	b, ok := r.reveal[c.St.Focus]
	if !ok {
		b, ok = r.boxes[c.St.Focus]
	}
	if !ok {
		return top
	}
	if b.y+b.h > top+page {
		top = b.y + b.h - page
	}
	if b.y < top {
		top = b.y
	}
	return min(max(top, 0), max(total-page, 0))
}

// clipHits clips to a scroll view's window the hits painted since
// hits[from], and the boxes of its child k and what k holds: those it
// hides take no click, and are not where an element is.
func (r *Rendition) clipHits(from int, k *view.Element, win box) {
	clip := func(x, y, w, h int) (int, int, int, int, bool) {
		x0, y0 := max(x, win.x), max(y, win.y)
		x1, y1 := min(x+w, win.x+win.w), min(y+h, win.y+win.h)
		return x0, y0, x1 - x0, y1 - y0, x1 > x0 && y1 > y0
	}
	kept := r.hits[:from]
	for _, h := range r.hits[from:] {
		var ok bool
		if h.x, h.y, h.w, h.h, ok = clip(h.x, h.y, h.w, h.h); ok {
			kept = append(kept, h)
		}
	}
	r.hits = kept
	var walk func(e *view.Element)
	walk = func(e *view.Element) {
		if e == nil {
			return
		}
		if b, ok := r.boxes[e.ID]; ok {
			if x, y, w, h, ok := clip(b.x, b.y, b.w, b.h); ok {
				r.boxes[e.ID] = box{x, y, w, h}
			} else {
				delete(r.boxes, e.ID)
			}
		}
		for _, c := range e.Children {
			walk(c)
		}
	}
	walk(k)
}

// skipCols is a row from column n on: a wide glyph cut in half is a space.
func skipCols(gs []glyph, n int) []glyph {
	at := 0
	for i, g := range gs {
		if at >= n {
			return gs[i:]
		}
		if at+g.width > n {
			return append([]glyph{{text: " ", width: 1, style: g.style}}, gs[i+1:]...)
		}
		at += g.width
	}
	return nil
}

// paintScrollbar paints a scroll view's bar, page rows from (x, y): a
// track of │ in border, faint, and a thumb of ┃ as long as the share of
// the content that shows, where it shows; muted, and accent while the
// scroll view has the keyboard. Nothing while all of it shows.
func (l *layout) paintScrollbar(cv *canvas, e *view.Element, x, y, page, total, top int) {
	if total <= page {
		return
	}
	size := max(1, (page*page+total/2)/total)
	at := 0
	if last := total - page; last > 0 {
		at = (top*(page-size) + last/2) / last
	}
	thumb := style{role: Muted}
	if l.r.focused(e.ID) {
		thumb.role = Accent
	}
	for i := range page {
		g := glyph{text: "│", width: 1, style: style{role: Border, attr: Faint}}
		if i >= at && i < at+size {
			g = glyph{text: "┃", width: 1, style: thumb}
		}
		cv.set(x, y+i, g)
	}
}

// scrollKey works a focused scroll view by a key, as bubbles' viewport
// takes them: ArrowUp and ArrowDown (k and j) a row, PageUp (b) and
// PageDown (f, Space) a page, u and d (Control+u, Control+d) half a page,
// Home and End (g and G) to its ends, and ArrowLeft and ArrowRight (h and
// l) six columns, when its lines do not wrap. ok reports whether the key is
// one of those.
func (r *Rendition) scrollKey(e *view.Element, key string) (ok bool, err error) {
	s, drawn := r.scrolls[e.ID]
	if !drawn {
		return false, nil
	}
	dy, dx := 0, 0
	switch key {
	case "ArrowUp", "k":
		dy = -1
	case "ArrowDown", "j":
		dy = 1
	case "PageUp", "b":
		dy = -s.page
	case "PageDown", "f", "Space":
		dy = s.page
	case "u", "Control+u":
		dy = -s.page / 2
	case "d", "Control+d":
		dy = s.page / 2
	case "Home", "g":
		dy = -s.rows
	case "End", "G":
		dy = s.rows
	case "ArrowLeft", "h":
		dx = -sideCols
	case "ArrowRight", "l":
		dx = sideCols
	default:
		return false, nil
	}
	if dx != 0 && s.wide <= s.cols {
		return false, nil
	}
	r.scrollBy(e, s, dx, dy)
	return true, nil
}

// scrollBy moves a scroll view dy rows and dx columns, clamped to its
// content; at its end, with follow, it follows the tail again.
func (r *Rendition) scrollBy(e *view.Element, s scrolled, dx, dy int) {
	last := max(s.rows-s.page, 0)
	top := min(max(s.top+dy, 0), last)
	left := min(max(s.left+dx, 0), max(s.wide-s.cols, 0))
	s.top, s.left = top, left
	r.scrolls[e.ID] = s
	r.c.Scrolled(e.ID, top, left, e.Active && top >= last)
}

// Wheel is a notch of the wheel at (col, row) of the last frame: dy rows
// down (up when negative), dx columns right, as the terminal reported it.
// A HottyTree with a height under the pointer scrolls first, three rows a
// notch, while it can move that way: it is inside any scroll view there.
// Else the innermost scroll view under the pointer that can still move
// that way scrolls, three rows a notch, or six columns, as bubbles'
// viewport does; handled reports whether one did.
func (r *Rendition) Wheel(col, row, dx, dy int) (handled bool) {
	if dy != 0 {
		for _, h := range slices.Backward(r.hits) {
			if col < h.x || col >= h.x+h.w || row < h.y || row >= h.y+h.h {
				continue
			}
			if e := r.c.V.Find(h.id); e != nil && e.Kind == view.Tree && r.c.ScrollTree(e.ID, dy*wheelRows) {
				return true
			}
			break
		}
	}
	for _, id := range slices.Backward(r.scrollOrder) {
		s := r.scrolls[id]
		e := r.c.V.Find(id)
		if e == nil || col < s.at.x || col >= s.at.x+s.at.w || row < s.at.y || row >= s.at.y+s.at.h {
			continue
		}
		mx, my := dx*sideCols, dy*wheelRows
		if s.wide <= s.cols {
			mx = 0
		}
		last := max(s.rows-s.page, 0)
		if (my < 0 && s.top == 0 || my > 0 && s.top >= last || my == 0) &&
			(mx < 0 && s.left == 0 || mx > 0 && s.left >= max(s.wide-s.cols, 0) || mx == 0) {
			continue
		}
		r.scrollBy(e, s, mx, my)
		return true
	}
	return false
}
