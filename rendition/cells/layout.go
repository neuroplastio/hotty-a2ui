package cells

import (
	"slices"

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
	// labelW is the width a one-line field pads its label to on its
	// value's row: its run's widest (labelRuns).
	labelW map[*view.Element]int
	// indented are the Buttons, and the Rows of them, that a Column of
	// fields starts past its gutter (labelRuns, indent).
	indented map[*view.Element]bool
	// pages are each HottyPaginator's child as it stands for each page
	// (pagerPages).
	pages map[*view.Element][]*view.Element
}

type sized struct {
	e *view.Element
	w int
}

func newLayout(r *Rendition) *layout {
	return &layout{r: r, nat: map[*view.Element]int{}, min: map[*view.Element]int{}, hgt: map[sized]int{}, text: map[sized][]tline{},
		labelW: map[*view.Element]int{}, indented: map[*view.Element]bool{}, pages: map[*view.Element][]*view.Element{}}
}

// labelRuns notes, for each run of a Column's children whose labels are
// on their value's row (inlineRow), the widest label: each pads its own to
// it, so that their values start in one column, as a GUI form's do. In a
// Column that holds a field, it notes the Buttons, and the Rows of
// nothing but Buttons, which start past the gutter (indent), so that a
// form's buttons line up with its labels, as a GUI form's do with its
// inputs (the maintainer, 2026-10-11).
func (l *layout) labelRuns(kids []*view.Element) {
	if slices.ContainsFunc(kids, holdsField) {
		for _, k := range kids {
			if onlyButtons(k) {
				l.indented[k] = true
			}
		}
	}
	for i := 0; i < len(kids); {
		if !inlineRow(kids[i]) {
			i++
			continue
		}
		j, lw := i, 0
		for ; j < len(kids) && inlineRow(kids[j]); j++ {
			lw = max(lw, Width(kids[j].Label))
		}
		for _, k := range kids[i:j] {
			l.labelW[k] = lw
		}
		i = j
	}
}

// labelWidth is the width a field's label takes on its value's row: its
// run's widest, or its own.
func (l *layout) labelWidth(e *view.Element) int {
	if n, ok := l.labelW[e]; ok {
		return n
	}
	return Width(e.Label)
}

// shown are the children that are drawn: not hidden.
func shown(es []*view.Element) []*view.Element {
	var out []*view.Element
	for _, e := range es {
		if e != nil && !e.Hidden {
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

// widest is the widest of a Column's children, each past its indent.
func (l *layout) widest(es []*view.Element) int {
	l.labelRuns(es)
	n := 0
	for _, e := range es {
		n = max(n, l.indent(e)+l.natural(e))
	}
	return n
}

// indent is the columns a Column's child starts past the Column's start:
// a field's gutter for a Button, or a Row of them, in a Column of fields
// (labelRuns); else none.
func (l *layout) indent(e *view.Element) int {
	if l.indented[e] {
		return gutter
	}
	return 0
}

// holdsField reports whether a Column's child is a field, or a Row that
// holds one.
func holdsField(e *view.Element) bool {
	if e.Kind == view.Stack && e.Dir == view.Horizontal {
		return slices.ContainsFunc(shown(e.Children), holdsField)
	}
	return isField(e)
}

// onlyButtons reports whether a Column's child is a Button, not a List's
// row, or a Row of nothing but those.
func onlyButtons(e *view.Element) bool {
	if e.Kind == view.Stack && e.Dir == view.Horizontal {
		kids := shown(e.Children)
		return len(kids) > 0 && !slices.ContainsFunc(kids, func(k *view.Element) bool { return !onlyButtons(k) })
	}
	return e.Kind == view.Button && !e.Item
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
	switch {
	case richItem(e):
		n = l.minimum(e.Children[0]) + 2
		l.min[e] = n
		return n
	}
	switch e.Kind {
	case view.Stack:
		if e.Dir != view.Horizontal {
			l.labelRuns(kids)
		}
		for i, k := range kids {
			if e.Dir != view.Horizontal {
				n = max(n, l.indent(k)+l.minimum(k))
				continue
			}
			if i > 0 {
				n += colGap(e)
			}
			n += l.minimum(k)
		}
	case view.Card:
		l.labelRuns(kids)
		for _, k := range kids {
			n = max(n, l.indent(k)+l.minimum(k))
		}
		n += 4
	case view.Form, view.Modal, view.Tabs:
		l.labelRuns(kids)
		for _, k := range kids {
			n = max(n, l.indent(k)+l.minimum(k))
		}
	case view.Text, view.Document:
		for _, t := range l.lines(e, noWrap) {
			n = max(n, longestWord(t.gs))
		}
	case view.TextField, view.DateTime:
		n = gutter + max(longestWord(line(e.Label, style{})), 1)
		if !isLongText(e) {
			// Its label and a space, the inset and 3 columns of input,
			// which scrolls.
			n = gutter + labelled(e, inset+3)
		}
	case view.Slider, view.RangeSlider:
		n = gutter + 4 + sliderValueWidth(e)
	case view.Choice:
		n = gutter + labelled(e, inset+inputWidth(e))
		if !isSelect(e) {
			n = longestWord(line(e.Label, style{}))
			for _, ow := range optionWidths(e) {
				n = max(n, ow)
			}
			n += gutter
		}
	case view.Divider:
		n = 1
	case view.Progress:
		n = max(longestWord(line(e.Label, style{})), 3+percentWidth)
	case view.Table:
		n = tableMinimum(e)
	case view.RichList:
		n = listMinimum(e)
	case view.FilePicker:
		n = fileMinimum(e)
	case view.Picker:
		n = pickMinimum(e)
	case view.Tree:
		n = treeMinimum(e)
	case view.Chart:
		n = chartMinimum(e)
	case view.Sparkline:
		// It shows the last values that fit.
		n = min(sparkWidth(e), 1)
	case view.BigText:
		// Narrower, a word breaks between letters.
		n = bigMinimum(e)
	case view.QRCode:
		// A code does not shrink; narrower, its value shows as text.
		n = qrMinimum(e)
	case view.KeyHints:
		// It cuts what does not fit (shortHints).
		n = 1
	case view.ScrollView:
		n = l.scrollMinimum(e)
	case view.Paginator:
		n = l.pagerMinimum(e)
	case view.Listing:
		n = codeMinimum(e)
	case view.DiffView:
		n = diffMinimum(e)
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
	if richItem(e) {
		return l.natural(e.Children[0]) + 2
	}
	switch e.Kind {
	case view.Stack:
		kids := shown(e.Children)
		if e.Dir != view.Horizontal {
			l.labelRuns(kids)
			return l.widest(kids)
		}
		n := 0
		for i, k := range kids {
			if i > 0 {
				n += colGap(e)
			}
			n += l.natural(k)
		}
		return n
	case view.Card:
		kids := shown(e.Children)
		l.labelRuns(kids)
		return l.widest(kids) + 4
	case view.Form, view.Modal:
		kids := shown(e.Children)
		l.labelRuns(kids)
		return l.widest(kids)
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
	case view.Text, view.Document:
		n := 0
		for _, t := range l.lines(e, noWrap) {
			n = max(n, width(t.gs))
		}
		return n
	case view.Divider:
		return 1
	case view.KeyHints:
		return l.r.hintsWidth(e)
	case view.ScrollView:
		return l.scrollWidth(e)
	case view.Paginator:
		return l.pagerWidth(e)
	case view.Listing:
		return codeWidth(e)
	case view.DiffView:
		return diffWidth(e)
	}
	if inlineRow(e) {
		// Its label padded to its run's widest, so that the inputs of a
		// run start in one column and none is cut for it.
		return gutter + l.labelWidth(e) + 1 + inset + inputWidth(e)
	}
	if s := l.r.sized[e.ID]; isLongText(e) && s.cols > 0 {
		// As wide as its grip made its value.
		return gutter + max(Width(e.Label), s.cols)
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
	if richItem(e) {
		return l.height(e.Children[0], w-2)
	}
	switch e.Kind {
	case view.Stack:
		kids := shown(e.Children)
		if len(kids) == 0 && e.Movable {
			// An empty List whose items move keeps a row, as a page keeps
			// it a card's height, so that a drag has somewhere to land and
			// its line shows.
			return 1
		}
		if e.Dir != view.Horizontal {
			return l.columnHeight(kids, e.Align, rowGap(e), w)
		}
		_, ws := l.row(kids, e.Justify, colGap(e), w)
		h := 0
		for i, k := range kids {
			h = max(h, l.height(k, ws[i]))
		}
		return h
	case view.Card:
		return l.columnHeight(shown(e.Children), "stretch", -1, w-4) + 2
	case view.Form, view.Modal:
		return l.columnHeight(shown(e.Children), "stretch", -1, w)
	case view.Tabs:
		bar, content := tabParts(e)
		return tabRows(bar, w) + l.columnHeight(content, "stretch", -1, w)
	case view.Text, view.Document:
		return len(l.lines(e, w))
	case view.ScrollView:
		return e.Height
	case view.Paginator:
		return l.pagerHeight(e, w)
	case view.Listing:
		return len(codeRows(e, w))
	case view.DiffView:
		return len(l.diffRows(e, w))
	}
	return l.controlHeight(e, w)
}

// lines are a Text's lines at width w, or a HottyMarkdown's.
func (l *layout) lines(e *view.Element, w int) []tline {
	k := sized{e, w}
	if t, ok := l.text[k]; ok {
		return t
	}
	var t []tline
	if e.Kind == view.Document {
		t = l.r.document(e, w)
	} else {
		t = markdown(e, w)
	}
	l.text[k] = t
	return t
}

// columnWidth is a Column child's width: the column's when align is
// stretch, else its own, at most the column's.
func (l *layout) columnWidth(e *view.Element, align string, w int) int {
	w = max(w-l.indent(e), 1)
	if align == "stretch" || align == "" {
		return w
	}
	return min(l.natural(e), w)
}

func (l *layout) columnHeight(kids []*view.Element, align string, gap, w int) int {
	// Before a child's natural width is asked for, which its run's labels
	// widen.
	l.labelRuns(kids)
	hs := make([]int, len(kids))
	h := 0
	for i, k := range kids {
		hs[i] = l.height(k, l.columnWidth(k, align, w))
		h += hs[i]
	}
	for _, s := range separators(kids, hs, gap) {
		h += s
	}
	return h
}

// separators are the blank rows a Column puts before each of its
// children, as tall as hs: its gap (profile §6.36), else as separator
// says where gap is -1; none before one that draws no rows, such as an
// empty List, so that nothing is set apart from nothing, and one that
// follows such a child is set apart from the child before.
func separators(kids []*view.Element, hs []int, gap int) []int {
	out := make([]int, len(kids))
	prev := -1
	for i, k := range kids {
		if hs[i] == 0 {
			continue
		}
		if prev >= 0 && gap >= 0 {
			out[i] = gap
		} else if prev >= 0 {
			out[i] = separator(kids[prev], k)
		}
		prev = i
	}
	return out
}

// separator is the blank rows a Column puts between two children: none
// between two fields, which their underlines part as a GUI form's
// outlines do; one between a field and another control, as bubbles sets a
// form's button apart, and between two other controls when either has a
// title row; one before a control after a heading, a form's title; before
// a HottyKeyHints, as bubbles' help sits a row under what it is for; after
// a HottyScrollView, whose box draws no edge but its scrollbar, a
// HottyCode, which draws none, a HottyChart, whose labels and legend end
// it, or a HottyPaginator, whose dots end its pages, so that what follows
// does not read as its content; after a HottyBigText, as its own lines
// are a row apart, so that its letters do not meet what follows; before
// a Button (or a Row of them) after a Text, as huh sets a question's
// buttons apart (the maintainer, round 7); else none, so that a stack of
// Buttons or of HottySpinners stays tight.
func separator(before, e *view.Element) int {
	a, b := edge(before, false), edge(e, true)
	if b.Kind == view.KeyHints || a.Kind == view.ScrollView || a.Kind == view.Listing || a.Kind == view.DiffView || a.Kind == view.Chart ||
		a.Kind == view.Paginator || a.Kind == view.BigText && a.Label != "" {
		return 1
	}
	if b.Kind == view.Button && a.Kind == view.Text {
		return 1
	}
	if isControlElement(b) && endsInHeading(a) {
		// A heading over a form or a stack of bars, as huh and bubbles
		// set a form's title apart.
		return 1
	}
	if !isControlElement(a) || !isControlElement(b) || isField(a) && isField(b) {
		// Fields stack tight: an input's underline parts it from the next
		// field.
		return 0
	}
	if hasTitle(a) || hasTitle(b) || isField(a) != isField(b) {
		return 1
	}
	return 0
}

// rowGap is a Column's io_neuroplast_hotty.gap, the blank rows between
// its children, or -1 where it has none and separator decides them.
func rowGap(e *view.Element) int {
	if e.Gap == nil {
		return -1
	}
	return *e.Gap
}

// colGap is the columns between a Row's children: its
// io_neuroplast_hotty.gap, else one.
func colGap(e *view.Element) int {
	if e.Gap == nil {
		return 1
	}
	return *e.Gap
}

// endsInHeading reports whether an element is a Text whose last block is
// a heading.
func endsInHeading(e *view.Element) bool {
	if e == nil || e.Kind != view.Text {
		return false
	}
	bs := view.Markdown(e.Markdown)
	return len(bs) > 0 && bs[len(bs)-1].Kind == view.Heading
}

// edge is what a separator sees of a neighbour: through a Row, a Column or
// a HottyForm, its first child (first) or its last, so that a field just
// inside a form is set apart from one just outside it.
func edge(e *view.Element, first bool) *view.Element {
	for e != nil && (e.Kind == view.Stack || e.Kind == view.Form) {
		kids := shown(e.Children)
		if len(kids) == 0 {
			return e
		}
		if first {
			e = kids[0]
		} else {
			e = kids[len(kids)-1]
		}
	}
	return e
}

// row places a Row's children across w columns (profile §3.3), as a host
// lays out its flex row: an unweighted child as wide as its natural
// width, a weighted one from nothing (flex: weight 1 0%), so that the
// columns the unweighted leave are the weighted's, by weight, whatever
// they hold (share). When the children are too wide, a column at a time
// comes off the widest that is wider than its minimum (the first of
// equals), and once none is, off the widest. With no weighted child, the
// spare columns go as justify says; justify stretch weighs every child,
// 1 when it is not. gap columns separate neighbours (colGap). It returns
// each child's x and width.
func (l *layout) row(kids []*view.Element, justify string, gap, w int) (xs, ws []int) {
	n := len(kids)
	xs, ws = make([]int, n), make([]int, n)
	if n == 0 {
		return
	}
	avail := max(w-gap*(n-1), 0)
	weights := make([]float64, n)
	weighted := false
	for i, k := range kids {
		weights[i] = k.Weight
		if justify == "stretch" && weights[i] <= 0 {
			// As flex: 1 1 0, every child grows alike.
			weights[i] = 1
		}
		weighted = weighted || weights[i] > 0
	}
	sum := 0
	mins := make([]int, n)
	for i, k := range kids {
		ws[i] = min(l.natural(k), avail)
		mins[i] = min(l.minimum(k), ws[i])
		if weights[i] > 0 {
			ws[i] = mins[i]
		}
		sum += ws[i]
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
	if weighted {
		share(ws, mins, weights, avail-sum)
		return place(ws, make([]int, n+1), gap), ws
	}
	gaps := l.spread(ws, weights, justify, avail-sum)
	return place(ws, gaps, gap), ws
}

// share hands a Row's spare columns to its weighted children as a host's
// flex grow does from a basis of nothing: what they hold and the spare
// together, by weight. One whose part would be narrower than its minimum
// is held at it, and the others share what remains; the floors first,
// then a column each to those not held, in order.
func share(ws, mins []int, weights []float64, spare int) {
	if spare <= 0 {
		return
	}
	room := spare
	held := make([]bool, len(ws))
	for i, wt := range weights {
		if wt > 0 {
			room += ws[i]
		} else {
			held[i] = true
		}
	}
	total := 0.0
	for {
		total = 0
		for i, wt := range weights {
			if !held[i] {
				total += wt
			}
		}
		if total == 0 {
			return
		}
		var under []int
		for i, wt := range weights {
			if !held[i] && float64(room)*wt/total < float64(mins[i]) {
				under = append(under, i)
			}
		}
		if len(under) == 0 {
			break
		}
		for _, i := range under {
			held[i], ws[i] = true, mins[i]
			room -= mins[i]
		}
	}
	given := 0
	for i, wt := range weights {
		if !held[i] {
			ws[i] = int(float64(room) * wt / total)
			given += ws[i]
		}
	}
	for i := 0; given < room; i = (i + 1) % len(ws) {
		if !held[i] {
			ws[i]++
			given++
		}
	}
}

// place is where a Row's children go: gap columns apart, plus justify's
// gaps.
func place(ws, gaps []int, gap int) []int {
	xs := make([]int, len(ws))
	x := gaps[0]
	for i := range ws {
		if i > 0 {
			x += gap + gaps[i]
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
