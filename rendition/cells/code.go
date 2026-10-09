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
// columns past the gutter, broken into rows of them with wrap, cut at
// them with "…" without; at noWrap, a row a line.
func codeRows(e *view.Element, w int) []codeRow {
	n, s := codeGutter(e)
	cw := max(w-n-s, 1)
	if w == noWrap {
		cw = noWrap
	}
	var out []codeRow
	for i, l := range e.Code {
		gs := tokenGlyphs(l, Fg)
		if !e.Wrap || cw == noWrap {
			out = append(out, codeRow{line: i, first: true, gs: fit(gs, cw)})
			continue
		}
		for j, t := range chars(gs, cw) {
			out = append(out, codeRow{line: i, first: j == 0, gs: t.gs})
		}
	}
	return out
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
// row, then its code; a marked line's rows tinted across the width.
func (l *layout) paintCode(cv *canvas, e *view.Element, x, y, w int) {
	n, s := codeGutter(e)
	for i, r := range codeRows(e, w) {
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
