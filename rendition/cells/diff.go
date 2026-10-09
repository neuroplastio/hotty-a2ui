package cells

import (
	"strconv"

	"github.com/neuroplastio/hotty-a2ui/diff"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// A HottyDiff in cells (profile §3.4): a row for each file's name, each
// hunk's header and each line, unified or split; a rail down the left that
// marks the selected hunk.

// How far a changed line's background is tinted toward its role (a sixth,
// as a marked line of code is), and its changed words' (a third).
const (
	lineTint = 42
	wordTint = 85
)

// splitSep is the columns between a split diff's two sides, " │ ";
// splitCode, the fewest columns of code a side needs, or the diff is
// unified.
const (
	splitSep  = 3
	splitCode = 16
)

// diffRow is a row of a HottyDiff past its rail: its glyphs, as wide as
// the row, and the hunk it is part of (an index into RowIDs), or -1.
type diffRow struct {
	hunk int
	gs   []glyph
}

// opRole is the role a removed or an added line is tinted toward; ok is
// false for a context line.
func opRole(op diff.Op) (r Role, ok bool) {
	switch op {
	case diff.Removed:
		return Error, true
	case diff.Added:
		return Success, true
	}
	return 0, false
}

// tint is a style tinted toward a line's role, as wide as the row: its
// background where the theme can tint it, its colour where not (BackFg).
func tint(st style, op diff.Op, word bool) style {
	r, ok := opRole(op)
	if !ok {
		return st
	}
	st.back, st.backMix, st.backFg = r, lineTint, true
	if word {
		st.backMix, st.backAttr = wordTint, Reverse
	}
	return st
}

// lineCode is a diff line's code: each token coloured by its kind, and a
// changed line tinted toward its role, its changed words more so.
func lineCode(l diff.Line) []glyph {
	var out []glyph
	for _, s := range l.Segments() {
		out = append(out, line(s.Text, tint(tokenStyle(s.Kind, Fg), l.Op, s.Changed))...)
	}
	return out
}

// diffSign is a line's sign: "-" in error, "+" in success, " ".
func diffSign(op diff.Op) glyph {
	switch op {
	case diff.Removed:
		return glyph{text: "-", width: 1, style: style{role: Error}}
	case diff.Added:
		return glyph{text: "+", width: 1, style: style{role: Success}}
	}
	return glyph{text: " ", width: 1}
}

// number is a line number, right-aligned in n columns, in muted; blank
// for 0 (the side that does not have the line).
func number(v, n int) []glyph {
	s := ""
	if v > 0 {
		s = strconv.Itoa(v)
	}
	return concat(repeat(" ", n-len(s), style{}), line(s, style{role: Muted}))
}

// digits are the columns of a HottyDiff's widest old and new line numbers.
func digits(d *diff.Diff) (old, new int) {
	for _, f := range d.Files {
		for _, h := range f.Hunks {
			for _, l := range h.Lines {
				old = max(old, len(strconv.Itoa(l.Old)))
				new = max(new, len(strconv.Itoa(l.New)))
			}
		}
	}
	return old, new
}

// diffGutter is a unified row's columns before its code: the old line's
// number and the new's, each and a space, then its sign and a space.
func diffGutter(e *view.Element) int {
	if !e.Numbers {
		return 2
	}
	o, n := digits(e.Diff)
	return o + n + 4
}

// sideGutter is a split side's: its line's number and a space, its sign
// and a space.
func sideGutter(e *view.Element) int {
	if !e.Numbers {
		return 2
	}
	o, n := digits(e.Diff)
	return max(o, n) + 3
}

// diffRail is the column left of the rows, which marks the selected hunk.
const diffRail = 1

// widestLine is the widest line of code a HottyDiff has.
func widestLine(e *view.Element) int {
	n := 0
	for _, f := range e.Diff.Files {
		for _, h := range f.Hunks {
			for _, l := range h.Lines {
				n = max(n, width(lineCode(l)))
			}
		}
	}
	return n
}

// diffWidth is a HottyDiff's natural width: its widest line after its
// gutter, twice that in a split, and its headers.
func diffWidth(e *view.Element) int {
	n := diffGutter(e) + widestLine(e)
	if e.Variant == "split" {
		n = 2*(sideGutter(e)+widestLine(e)) + splitSep
	}
	for _, f := range e.Diff.Files {
		n = max(n, width(fileHeader(f)))
		for _, h := range f.Hunks {
			n = max(n, width(hunkHeader(h, style{})))
			if h.Skipped > 0 {
				n = max(n, width(folded(h.Skipped)))
			}
		}
		if f.After > 0 {
			n = max(n, width(folded(f.After)))
		}
	}
	return diffRail + n
}

// diffMinimum is the narrowest a HottyDiff gets: its gutter and eight
// columns of code, unified.
func diffMinimum(e *view.Element) int {
	return diffRail + diffGutter(e) + 8
}

// splitSides reports whether a HottyDiff shows its sides side by side at width
// w: it asks to, and each side has room for splitCode columns of code.
func splitSides(e *view.Element, w int) bool {
	return e.Variant == "split" && (w-diffRail-splitSep)/2-sideGutter(e) >= splitCode
}

// fileHeader is a file's row: its name in bold (old → new when it was
// renamed), then what the change adds and removes, or that it is binary.
func fileHeader(f diff.File) []glyph {
	name := f.Name()
	if f.Old != "" && f.New != "" && f.Old != f.New {
		name = f.Old + " → " + f.New
	}
	out := line(name, style{attr: Bold})
	switch add, rm := f.Stat(); {
	case f.Binary:
		out = append(out, line("  binary", style{role: Muted})...)
	case f.Old == "" && f.New != "":
		out = append(out, line("  new", style{role: Success})...)
	case f.New == "" && f.Old != "":
		out = append(out, line("  deleted", style{role: Error})...)
	default:
		if add > 0 {
			out = append(out, line("  +"+strconv.Itoa(add), style{role: Success})...)
		}
		if rm > 0 {
			sep := "  "
			if add > 0 {
				sep = " "
			}
			out = append(out, line(sep+"-"+strconv.Itoa(rm), style{role: Error})...)
		}
	}
	return out
}

// hunkHeader is a hunk's row: its header in info, its section in muted.
// st overrides both (the selected hunk's, reversed).
func hunkHeader(h diff.Hunk, st style) []glyph {
	head, sec := style{role: Info}, style{role: Muted}
	if st != (style{}) {
		head, sec = st, st
	}
	out := line(h.Header(), head)
	if h.Section != "" {
		out = append(out, line(" "+h.Section, sec)...)
	}
	return out
}

// folded is the row for a run of n unchanged lines the diff leaves out,
// in muted.
func folded(n int) []glyph {
	s := "⋯ " + strconv.Itoa(n) + " unchanged lines"
	if n == 1 {
		s = "⋯ 1 unchanged line"
	}
	return line(s, style{role: Muted})
}

// pad fills gs out to w columns with spaces of style st, or cuts it there.
func pad(gs []glyph, w int, st style) []glyph {
	gs = fit(gs, w)
	return append(gs, repeat(" ", w-width(gs), st)...)
}

// diffRows are a HottyDiff's rows at width w, past its rail: for each file
// a blank row before all but the first, its name (when it has one), and
// for each hunk its header and its lines, unified or split. A run of
// unchanged lines left out is folded into a row of its own: before a
// hunk, and after the last where the diff knows (Compare). The selected
// hunk's header is reversed, in accent while the diff has the keyboard.
func (l *layout) diffRows(e *view.Element, w int) []diffRow {
	cw := max(w-diffRail, 1)
	sel := e.SelectedRow()
	selStyle := style{role: Muted, attr: Reverse}
	if l.r.focused(e.ID) {
		selStyle.role = Accent
	}
	var out []diffRow
	k := 0
	for i, f := range e.Diff.Files {
		if i > 0 {
			out = append(out, diffRow{hunk: -1, gs: pad(nil, cw, style{})})
		}
		if f.Name() != "" || len(e.Diff.Files) > 1 {
			out = append(out, diffRow{hunk: -1, gs: pad(fileHeader(f), cw, style{})})
		}
		for _, h := range f.Hunks {
			if h.Skipped > 0 {
				out = append(out, diffRow{hunk: -1, gs: pad(folded(h.Skipped), cw, style{})})
			}
			st, hs := style{}, style{}
			if k == sel {
				st, hs = selStyle, selStyle
			}
			out = append(out, diffRow{hunk: k, gs: pad(hunkHeader(h, hs), cw, st)})
			var rows [][]glyph
			if splitSides(e, w) {
				rows = l.splitRows(e, h, cw)
			} else {
				rows = l.unifiedRows(e, h, cw)
			}
			for _, r := range rows {
				out = append(out, diffRow{hunk: k, gs: r})
			}
			k++
		}
		if f.After > 0 && len(f.Hunks) > 0 {
			out = append(out, diffRow{hunk: -1, gs: pad(folded(f.After), cw, style{})})
		}
	}
	if k == 0 && len(out) == 0 {
		out = append(out, diffRow{hunk: -1, gs: pad(line("No changes", style{role: Muted}), cw, style{})})
	}
	return out
}

// codeLines are a line's code at width w: broken into rows of it, or cut
// with "…".
func codeLines(gs []glyph, w int, wrap bool) [][]glyph {
	if !wrap {
		return [][]glyph{fit(gs, w)}
	}
	var out [][]glyph
	for _, t := range chars(gs, w) {
		out = append(out, t.gs)
	}
	if len(out) == 0 {
		out = [][]glyph{nil}
	}
	return out
}

// unifiedRows are a hunk's lines in one column: each the old and the new
// line's numbers, its sign, and its code; a long line's further rows leave
// the gutter blank. A changed line's row is tinted to its end.
func (l *layout) unifiedRows(e *view.Element, h diff.Hunk, w int) [][]glyph {
	o, n := digits(e.Diff)
	g := diffGutter(e)
	var out [][]glyph
	for _, ln := range h.Lines {
		row := tint(style{}, ln.Op, false)
		gutter := []glyph{diffSign(ln.Op), {text: " ", width: 1}}
		if e.Numbers {
			gutter = concat(number(ln.Old, o), repeat(" ", 1, style{}), number(ln.New, n), repeat(" ", 1, style{}), gutter)
		}
		for j, code := range codeLines(lineCode(ln), max(w-g, 1), e.Wrap) {
			lead := gutter
			if j > 0 {
				lead = repeat(" ", g, style{})
			}
			out = append(out, pad(concat(retint(lead, ln.Op), code), w, row))
		}
	}
	return out
}

// retint tints a gutter's cells with its line's background, keeping their
// colours where the theme cannot tint (no BackFg): the numbers stay muted.
func retint(gs []glyph, op diff.Op) []glyph {
	out := make([]glyph, len(gs))
	for i, g := range gs {
		st := tint(g.style, op, false)
		st.backFg = false
		g.style = st
		out[i] = g
	}
	return out
}

// splitRows are a hunk's lines side by side, the old side left and the
// new right, " │ " between: a context line on both, and in a run of
// removed lines followed by added ones, each removed line beside the added
// line that replaces it. A row is as tall as its taller side.
func (l *layout) splitRows(e *view.Element, h diff.Hunk, w int) [][]glyph {
	sw := (w - splitSep) / 2
	o, n := digits(e.Diff)
	nw := max(o, n)
	g := sideGutter(e)
	side := func(ln *diff.Line, num int) ([][]glyph, style) {
		row := tint(style{}, ln.Op, false)
		gutter := []glyph{diffSign(ln.Op), {text: " ", width: 1}}
		if e.Numbers {
			gutter = concat(number(num, nw), repeat(" ", 1, style{}), gutter)
		}
		var out [][]glyph
		for j, code := range codeLines(lineCode(*ln), max(sw-g, 1), e.Wrap) {
			lead := gutter
			if j > 0 {
				lead = repeat(" ", g, style{})
			}
			out = append(out, pad(concat(retint(lead, ln.Op), code), sw, row))
		}
		return out, row
	}
	sep := line(" │ ", style{role: Border})
	var out [][]glyph
	// A side shorter than its pair is filled out with its line's tint;
	// a side with no line is blank.
	pair := func(a, b *diff.Line) {
		var left, right [][]glyph
		lt, rt := style{}, style{}
		if a != nil {
			left, lt = side(a, a.Old)
		}
		if b != nil {
			right, rt = side(b, b.New)
		}
		rw := w - sw - splitSep
		for j := range max(len(left), len(right)) {
			lg, rg := pad(nil, sw, lt), pad(nil, rw, rt)
			if j < len(left) {
				lg = left[j]
			}
			if j < len(right) {
				rg = pad(right[j], rw, rt)
			}
			out = append(out, concat(lg, sep, rg))
		}
	}
	ls := h.Lines
	for i := 0; i < len(ls); {
		if ls[i].Op == diff.Context {
			pair(&ls[i], &ls[i])
			i++
			continue
		}
		j := i
		for j < len(ls) && ls[j].Op == diff.Removed {
			j++
		}
		k := j
		for k < len(ls) && ls[k].Op == diff.Added {
			k++
		}
		for m := 0; m < max(j-i, k-j); m++ {
			var a, b *diff.Line
			if i+m < j {
				a = &ls[i+m]
			}
			if j+m < k {
				b = &ls[j+m]
			}
			pair(a, b)
		}
		i = k
	}
	return out
}

// paintDiff paints a HottyDiff at (x, y), w wide: its rail, the selected
// hunk's rows marked "▎" (accent while it has the keyboard), then its
// rows. A click on a hunk's row is on that hunk; the selected hunk, with
// the rows that lead to it, is what a scroll view keeps in sight.
func (l *layout) paintDiff(cv *canvas, e *view.Element, x, y, w int) {
	r := l.r
	rows := l.diffRows(e, w)
	sel := e.SelectedRow()
	bar := style{role: Muted}
	if r.focused(e.ID) {
		bar.role = Accent
	}
	first, last := -1, -1
	for i, row := range rows {
		if row.hunk == sel && sel >= 0 {
			cv.set(x, y+i, glyph{text: "▎", width: 1, style: bar})
			if first < 0 {
				first = i
			}
			last = i
		}
		cv.write(x+diffRail, y+i, w-diffRail, row.gs)
		r.hits = append(r.hits, hit{x: x, y: y + i, w: w, h: 1, id: e.ID, opt: row.hunk})
	}
	r.boxes[e.ID] = box{x, y, w, len(rows)}
	if first >= 0 {
		// What leads to the hunk shows with it: its file's name, and the
		// fold before it.
		for first > 0 && rows[first-1].hunk < 0 {
			first--
		}
		r.reveal[e.ID] = box{x, y + first, w, last - first + 1}
	}
}

// clickDiff is a click on a HottyDiff, which has the keyboard now: on a
// hunk, it selects it, or acts on it when it was selected already.
func (r *Rendition) clickDiff(e *view.Element, hunk int) error {
	switch {
	case hunk < 0:
		return nil
	case hunk == e.SelectedRow():
		return r.c.Activate(e.ID)
	}
	return r.c.SelectRow(e.ID, hunk)
}
