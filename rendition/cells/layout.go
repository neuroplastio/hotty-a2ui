package cells

import (
	"github.com/neuroplastio/hotty-a2ui/view"
)

// layout lays one Draw's elements out (profile §3.3): what each is as wide
// as, how tall at a width, and where its children go. Results are kept for
// the Draw.
type layout struct {
	r    *Rendition
	nat  map[*view.Element]int
	min  map[*view.Element]int
	hgt  map[sized]int
	text map[sized][]tline
}

type sized struct {
	e *view.Element
	w int
}

func newLayout(r *Rendition) *layout {
	return &layout{r: r, nat: map[*view.Element]int{}, min: map[*view.Element]int{}, hgt: map[sized]int{}, text: map[sized][]tline{}}
}

// shown are the children that are drawn: not hidden.
func shown(es []*view.Element) []*view.Element {
	var out []*view.Element
	for _, e := range es {
		if e != nil && !e.A11y.Hidden {
			out = append(out, e)
		}
	}
	return out
}

// natural is an element's width when nothing wraps (its max-content).
func (l *layout) natural(e *view.Element) int {
	if n, ok := l.nat[e]; ok {
		return n
	}
	n := l.measure(e)
	l.nat[e] = n
	return n
}

func (l *layout) widest(es []*view.Element) int {
	n := 0
	for _, e := range es {
		n = max(n, l.natural(e))
	}
	return n
}

// minimum is the narrowest an element gets before a Row breaks its words
// or cuts it (its min-content): a Text's longest word; a container's
// children's; a control's natural width, but a field's label's longest
// word, a Slider's shortest track and its value, options' widest.
func (l *layout) minimum(e *view.Element) int {
	if n, ok := l.min[e]; ok {
		return n
	}
	n := 0
	kids := shown(e.Children)
	switch e.Kind {
	case view.Stack:
		for i, k := range kids {
			if e.Dir != view.Horizontal {
				n = max(n, l.minimum(k))
				continue
			}
			if i > 0 {
				n++
			}
			n += l.minimum(k)
		}
	case view.Card:
		for _, k := range kids {
			n = max(n, l.minimum(k))
		}
		n += 4
	case view.Form, view.Modal, view.Tabs:
		for _, k := range kids {
			n = max(n, l.minimum(k))
		}
	case view.Text:
		for _, t := range l.lines(e, noWrap) {
			n = max(n, longestWord(t.gs))
		}
	case view.TextField, view.DateTime:
		n = max(longestWord(line(e.Label, style{})), 1)
	case view.Slider:
		n = 4 + sliderValueWidth(e)
	case view.Choice:
		n = l.natural(e)
		if !isSelect(e) {
			n = longestWord(line(e.Label, style{}))
			for _, ow := range optionWidths(e) {
				n = max(n, ow)
			}
		}
	case view.Divider:
		n = 1
	default:
		n = l.natural(e)
	}
	l.min[e] = n
	return n
}

// longestWord is the widest run of glyphs without a space; a line's
// leading run (a list item's indent and marker) counts with its first
// word.
func longestWord(gs []glyph) int {
	n, cur, lead := 0, 0, true
	for _, g := range gs {
		if g.text == " " && !lead {
			cur = 0
			continue
		}
		if g.text != " " {
			lead = false
		}
		cur += g.width
		n = max(n, cur)
	}
	return n
}

func (l *layout) measure(e *view.Element) int {
	switch e.Kind {
	case view.Stack:
		kids := shown(e.Children)
		if e.Dir != view.Horizontal {
			return l.widest(kids)
		}
		n := 0
		for i, k := range kids {
			if i > 0 {
				n++
			}
			n += l.natural(k)
		}
		return n
	case view.Card:
		return l.widest(shown(e.Children)) + 4
	case view.Form, view.Modal:
		return l.widest(shown(e.Children))
	case view.Tabs:
		bar, content := tabParts(e)
		n := 0
		for i, t := range bar {
			if i > 0 {
				n += 2
			}
			n += Width(t.Label)
		}
		return max(n, l.widest(content))
	case view.Text:
		n := 0
		for _, t := range l.lines(e, noWrap) {
			n = max(n, width(t.gs))
		}
		return n
	case view.Divider:
		return 1
	}
	return controlWidth(e)
}

// height is an element's height at width w; 0 when w is not positive.
func (l *layout) height(e *view.Element, w int) int {
	if w <= 0 {
		return 0
	}
	k := sized{e, w}
	if h, ok := l.hgt[k]; ok {
		return h
	}
	h := l.measureHeight(e, w)
	l.hgt[k] = h
	return h
}

func (l *layout) measureHeight(e *view.Element, w int) int {
	switch e.Kind {
	case view.Stack:
		kids := shown(e.Children)
		if e.Dir != view.Horizontal {
			return l.columnHeight(kids, e.Align, w)
		}
		_, ws := l.row(kids, e.Justify, w)
		h := 0
		for i, k := range kids {
			h = max(h, l.height(k, ws[i]))
		}
		return h
	case view.Card:
		return l.columnHeight(shown(e.Children), "stretch", w-4) + 2
	case view.Form, view.Modal:
		return l.columnHeight(shown(e.Children), "stretch", w)
	case view.Tabs:
		bar, content := tabParts(e)
		return tabRows(bar, w) + l.columnHeight(content, "stretch", w)
	case view.Text:
		return len(l.lines(e, w))
	}
	return l.controlHeight(e, w)
}

// lines are a Text's lines at width w.
func (l *layout) lines(e *view.Element, w int) []tline {
	k := sized{e, w}
	if t, ok := l.text[k]; ok {
		return t
	}
	t := markdown(e, w)
	l.text[k] = t
	return t
}

// columnWidth is a Column child's width: the column's when align is
// stretch, else its own, at most the column's.
func (l *layout) columnWidth(e *view.Element, align string, w int) int {
	if align == "stretch" || align == "" {
		return w
	}
	return min(l.natural(e), w)
}

func (l *layout) columnHeight(kids []*view.Element, align string, w int) int {
	h := 0
	for _, k := range kids {
		h += l.height(k, l.columnWidth(k, align, w))
	}
	return h
}

// row places a Row's children across w columns (profile §3.3): each as
// wide as its natural width; when they are too wide, a column at a time
// comes off the widest that is wider than its minimum (the first of
// equals), and once none is, off the widest; spare columns go to the
// weighted children by weight, else as justify says. With justify
// stretch, the columns are shared by weight (1 when unweighted) whatever
// the children's widths. One column separates neighbours. It returns each
// child's x and width.
func (l *layout) row(kids []*view.Element, justify string, w int) (xs, ws []int) {
	n := len(kids)
	xs, ws = make([]int, n), make([]int, n)
	if n == 0 {
		return
	}
	avail := max(w-(n-1), 0)
	weights := make([]float64, n)
	for i, k := range kids {
		weights[i] = k.Weight
	}
	if justify == "stretch" {
		// As flex: 1 1 0, a weight as its grow: the columns are shared
		// out whatever the children hold.
		for i := range weights {
			if weights[i] <= 0 {
				weights[i] = 1
			}
		}
		l.spread(ws, weights, "", avail)
		return place(ws, make([]int, n+1)), ws
	}
	sum := 0
	for i, k := range kids {
		ws[i] = min(l.natural(k), avail)
		sum += ws[i]
	}
	mins := make([]int, n)
	for i, k := range kids {
		mins[i] = min(l.minimum(k), ws[i])
	}
	for ; sum > avail; sum-- {
		j := -1
		for i := range ws {
			if ws[i] > mins[i] && (j < 0 || ws[i] > ws[j]) {
				j = i
			}
		}
		if j < 0 {
			for i := range ws {
				if j < 0 || ws[i] > ws[j] {
					j = i
				}
			}
		}
		ws[j]--
	}
	gaps := l.spread(ws, weights, justify, avail-sum)
	return place(ws, gaps), ws
}

// place is where a Row's children go: one column apart, plus the gaps.
func place(ws, gaps []int) []int {
	xs := make([]int, len(ws))
	x := gaps[0]
	for i := range ws {
		if i > 0 {
			x += 1 + gaps[i]
		}
		xs[i] = x
		x += ws[i]
	}
	return xs
}

// spread hands out spare room along an axis: to the weighted sizes, by
// weight (floors, then a unit each to the weighted in order); else by
// justify (a Column's stretch weighs every child 1). It grows sizes in
// place and returns the extra gap before each child, and after the last.
func (l *layout) spread(sizes []int, weights []float64, justify string, spare int) []int {
	n := len(sizes)
	gaps := make([]int, n+1)
	if spare <= 0 {
		return gaps
	}
	total := 0.0
	for _, wt := range weights {
		total += wt
	}
	if total == 0 && justify == "stretch" {
		for i := range weights {
			weights[i] = 1
		}
		total = float64(n)
	}
	if total > 0 {
		given := 0
		for i, wt := range weights {
			g := int(float64(spare) * wt / total)
			sizes[i] += g
			given += g
		}
		for i := 0; given < spare; i = (i + 1) % n {
			if weights[i] > 0 {
				sizes[i]++
				given++
			}
		}
		return gaps
	}
	switch justify {
	case "center":
		gaps[0] = spare / 2
	case "end":
		gaps[0] = spare
	case "spaceBetween":
		if n > 1 {
			for i, s := range split(spare, n-1) {
				gaps[i+1] = s
			}
		}
	case "spaceAround":
		s := split(spare, 2*n)
		gaps[0] = s[0]
		for i := 1; i < n; i++ {
			gaps[i] = s[2*i-1] + s[2*i]
		}
	case "spaceEvenly":
		s := split(spare, n+1)
		for i := 0; i < n; i++ {
			gaps[i] = s[i]
		}
	}
	return gaps
}

// split divides total into n near-equal parts, the larger first.
func split(total, n int) []int {
	out := make([]int, n)
	for i := range out {
		out[i] = total / n
		if i < total%n {
			out[i]++
		}
	}
	return out
}

// tabParts are a Tabs' titles and the content of the tab shown.
func tabParts(e *view.Element) (bar, content []*view.Element) {
	for _, c := range shown(e.Children) {
		if c.Kind == view.Tab {
			bar = append(bar, c)
		} else {
			content = append(content, c)
		}
	}
	return
}

// flow places items of the given widths in rows w wide, gap columns
// apart, as many to a row as fit; an item wider than a row is cut to it.
// It returns each item's row and x.
func flow(widths []int, gap, w int) (rows, xs []int) {
	rows, xs = make([]int, len(widths)), make([]int, len(widths))
	r, x := 0, 0
	for i, iw := range widths {
		iw = min(iw, w)
		if x > 0 && x+gap+iw > w {
			r, x = r+1, 0
		} else if x > 0 {
			x += gap
		}
		rows[i], xs[i] = r, x
		x += iw
	}
	return
}

// tabRows is how many rows a Tabs' bar takes: its titles, two columns
// apart, and a rule under them.
func tabRows(bar []*view.Element, w int) int {
	if len(bar) == 0 {
		return 0
	}
	ws := make([]int, len(bar))
	for i, t := range bar {
		ws[i] = Width(t.Label)
	}
	rows, _ := flow(ws, 2, w)
	return rows[len(rows)-1] + 2
}

// optionGap is the columns between a Choice's options.
const optionGap = 2
