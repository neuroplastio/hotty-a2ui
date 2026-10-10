package cells

import (
	"strconv"

	"github.com/neuroplastio/hotty-a2ui/highlight"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// tokenStyle is how a token of kind k is painted (profile §3.4); plain is
// the role a plain token takes.
func tokenStyle(k highlight.Kind, plain Role) style {
	switch k {
	case highlight.Comment:
		return style{role: Muted, attr: Italic}
	case highlight.Keyword:
		// Not accent, which marks only focus (profile §3.6): bold, as
		// types are info and regular.
		return style{role: Info, attr: Bold}
	case highlight.Type:
		return style{role: Info}
	case highlight.Function:
		return style{role: plain, attr: Bold}
	case highlight.String, highlight.Inserted:
		return style{role: Success}
	case highlight.Literal, highlight.Meta:
		return style{role: Warning}
	case highlight.Deleted, highlight.Invalid:
		return style{role: Error}
	case highlight.Heading:
		return style{role: Info, attr: Bold}
	}
	return style{role: plain}
}

// tokenGlyphs are a line of code's glyphs, each token styled by its kind.
func tokenGlyphs(ts []highlight.Token, plain Role) []glyph {
	var out []glyph
	for _, t := range ts {
		out = append(out, line(t.Text, tokenStyle(t.Kind, plain))...)
	}
	return out
}

// markTint is the role a marked line's background is tinted toward, and
// how far (Cell.BackMix): a highlighted line takes the selection colour
// itself, the others a sixth of their role's over the background.
func markTint(m view.Mark) (Role, uint8) {
	switch m {
	case view.MarkHighlight:
		return Selection, 255
	case view.MarkAdded:
		return Success, 42
	case view.MarkWarning:
		return Warning, 42
	}
	return Error, 42
}

// markSign is a marked line's sign in the gutter.
func markSign(m view.Mark) glyph {
	text, role := "▎", Info
	switch m {
	case view.MarkAdded:
		text, role = "+", Success
	case view.MarkRemoved:
		text, role = "-", Error
	case view.MarkError:
		text, role = "✗", Error
	case view.MarkWarning:
		text, role = "!", Warning
	}
	return glyph{text: text, width: 1, style: style{role: role}}
}

// codeGutter is a HottyCode's columns before its code: its line numbers,
// right-aligned as wide as the widest, and a space; then, when it has
// marks, a column for their signs, and a space.
func codeGutter(e *view.Element) (numbers, signs int) {
	if e.Numbers {
		last := e.FirstLine + max(len(e.Code), 1) - 1
		numbers = max(len(strconv.Itoa(e.FirstLine)), len(strconv.Itoa(last))) + 1
	}
	if len(e.Marks) > 0 {
		signs = 2
	}
	return numbers, signs
}

// codeRow is a row of a HottyCode: its line (an index into Code), whether
// it is the line's first row, and its code.
type codeRow struct {
	line  int
	first bool
	gs    []glyph
}

// codeRows are a HottyCode's rows at width w: each line's code in the
// columns past the gutter, broken into rows of them with wrap; without,
// the columns from its Left, cut with "…" at a side that hides more
// (codeWindow); at noWrap, a row a line.
func codeRows(e *view.Element, w int) []codeRow {
	n, s := codeGutter(e)
	cw := max(w-n-s, 1)
	if w == noWrap {
		cw = noWrap
	}
	left, _ := codeLeft(e, cw)
	var out []codeRow
	for i, l := range e.Code {
		gs := tokenGlyphs(l, Fg)
		if cw == noWrap {
			out = append(out, codeRow{line: i, first: true, gs: gs})
			continue
		}
		if !e.Wrap {
			out = append(out, codeRow{line: i, first: true, gs: codeWindow(gs, left, cw)})
			continue
		}
		for j, t := range chars(gs, cw) {
			out = append(out, codeRow{line: i, first: j == 0, gs: t.gs})
		}
	}
	return out
}

// codeWindow is a line of code shown from column left, cw columns of it:
// a "…" in its first column while columns before it are hidden, and in
// its last while more follow, as a cut line ends.
func codeWindow(gs []glyph, left, cw int) []glyph {
	if left > 0 {
		rest := skipCols(gs, left+1)
		if len(rest) == 0 {
			// The line ends before the window: nothing of it shows.
			return nil
		}
		gs = append([]glyph{{text: "…", width: 1, style: rest[0].style}}, rest...)
	}
	return fit(gs, cw)
}

// codeLeft is the first column of code a HottyCode without wrap shows in
// cw columns: its Left, at most what shows its widest line's end; and the
// widest line's width. Lines that wrap, or fit, show from the start.
func codeLeft(e *view.Element, cw int) (left, wide int) {
	if e.Wrap || cw == noWrap {
		return 0, 0
	}
	for _, l := range e.Code {
		wide = max(wide, width(tokenGlyphs(l, Fg)))
	}
	return min(max(e.Left, 0), max(wide-cw, 0)), wide
}

// codeWidth is a HottyCode's natural width: its gutter and its widest
// line.
func codeWidth(e *view.Element) int {
	n, s := codeGutter(e)
	widest := 0
	for _, l := range e.Code {
		widest = max(widest, width(tokenGlyphs(l, Fg)))
	}
	return n + s + widest
}

// codeMinimum is the narrowest a HottyCode gets: its gutter and eight
// columns of code, or its widest line when that is narrower.
func codeMinimum(e *view.Element) int {
	n, s := codeGutter(e)
	return n + s + max(min(codeWidth(e)-n-s, 8), 1)
}

// paintCode paints a HottyCode at (x, y), w wide (profile §3.4): each row
// its line's number, in muted, and its mark's sign on the line's first
// row, then its code; a marked line's rows tinted across the width. Lines
// cut at the width scroll sideways under the wheel (Wheel), as a scroll
// view's do, the gutter staying put: its box, where it shows, is one.
func (l *layout) paintCode(cv *canvas, e *view.Element, x, y, w int) {
	n, s := codeGutter(e)
	rows := codeRows(e, w)
	if cw := max(w-n-s, 1); !e.Wrap && w != noWrap {
		if left, wide := codeLeft(e, cw); wide > cw {
			if left != e.Left {
				l.r.c.Scrolled(e.ID, 0, left, false)
			}
			at := box{x, y, w, len(rows)}
			if c := cv.clip; c != nil {
				x0, y0 := max(at.x, c.x), max(at.y, c.y)
				x1, y1 := min(at.x+at.w, c.x+c.w), min(at.y+at.h, c.y+c.h)
				at = box{x0, y0, max(x1-x0, 0), max(y1-y0, 0)}
			}
			l.r.scrolls[e.ID] = scrolled{at: at, page: len(rows), rows: len(rows), cols: cw, wide: wide, left: left}
			l.r.scrollOrder = append(l.r.scrollOrder, e.ID)
		}
	}
	for i, r := range rows {
		num := e.FirstLine + r.line
		if r.first && n > 0 {
			digits := strconv.Itoa(num)
			cv.write(x+n-1-len(digits), y+i, len(digits), glyphs(digits, style{role: Muted}))
		}
		mark, marked := e.Marks[num]
		if r.first && marked {
			cv.set(x+n, y+i, markSign(mark))
		}
		cv.write(x+n+s, y+i, w-n-s, r.gs)
		if marked {
			back, mix := markTint(mark)
			cv.restyle(x, y+i, w, 1, func(c *Cell) { c.Back, c.BackMix = back, mix })
		}
	}
}
