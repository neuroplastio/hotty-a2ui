package cells

import (
	"strconv"
	"strings"

	"github.com/neuroplastio/hotty-a2ui/view"
)

// A tline is one line of a Text: glyphs, or a rule across the width.
type tline struct {
	gs   []glyph
	rule bool
}

// markdown is a Text's lines at width w (profile §3.4): each block of its
// Markdown wrapped, a line a row; unwrapped when w is noWrap.
func markdown(e *view.Element, w int) []tline {
	base := style{}
	if e.Variant == "caption" {
		base.role = Muted
	}
	var out []tline
	for _, b := range view.Markdown(e.Markdown) {
		st := base
		var first, rest []glyph
		switch b.Kind {
		case view.Heading:
			st.attr |= Bold
			if b.Level == 1 {
				st.attr |= Underline
			}
		case view.ListItem:
			marker := "• "
			switch {
			case b.Task == 1:
				marker = "☐ "
			case b.Task == 2:
				marker = "✓ "
			case b.Number > 0:
				marker = strconv.Itoa(b.Number) + ". "
			}
			indent := strings.Repeat("  ", b.Level)
			first = glyphs(indent+marker, base)
			rest = repeat(" ", Width(indent+marker), base)
		case view.Quote:
			first = glyphs(strings.Repeat("▎ ", b.Level+1), style{role: Border})
			rest = first
		case view.CodeBlock:
			for _, l := range strings.Split(b.Runs[0].Text, "\n") {
				out = append(out, chars(glyphs(l, style{role: Muted}), w)...)
			}
			continue
		case view.Rule:
			out = append(out, tline{rule: true})
			continue
		}
		var gs []glyph
		for _, r := range b.Runs {
			gs = append(gs, glyphs(r.Text, runStyle(st, r))...)
		}
		for _, l := range wrap(gs, w, first, rest) {
			out = append(out, tline{gs: l})
		}
	}
	return out
}

// runStyle is a run's style on top of its block's.
func runStyle(st style, r view.Run) style {
	if r.Style&view.Bold != 0 {
		st.attr |= Bold
	}
	if r.Style&view.Italic != 0 {
		st.attr |= Italic
	}
	if r.Style&view.Strike != 0 {
		st.attr |= Strike
	}
	if r.Style&view.Code != 0 {
		st.role = Muted
	}
	if r.Href != "" {
		st.attr |= Underline
		st.link = r.Href
	}
	return st
}

// noWrap is a width wider than any line: lines are not wrapped.
const noWrap = 1 << 30

// wrap breaks glyphs into lines w columns wide (profile §3.3): at hard
// breaks, and at spaces; a word longer than a line is broken where the
// line ends. The spaces at a soft break are dropped, and so are trailing
// spaces. first starts the first line, rest the others; neither when they
// leave no room. Each hard line is at least one line.
func wrap(gs []glyph, w int, first, rest []glyph) [][]glyph {
	if w <= 0 {
		return nil
	}
	if width(first) >= w || width(rest) >= w {
		first, rest = nil, nil
	}
	var out [][]glyph
	var cur []glyph
	curW, empty := 0, true
	start := func(prefix []glyph) {
		cur = append([]glyph(nil), prefix...)
		curW, empty = width(prefix), true
	}
	breakLine := func() {
		out = append(out, cur)
		start(rest)
	}
	for n, hard := range splitLines(gs) {
		if n == 0 {
			start(first)
		} else {
			breakLine()
		}
		soft := false
		for i := 0; i < len(hard); {
			j := i
			for j < len(hard) && hard[j].text == " " {
				j++
			}
			k := j
			for k < len(hard) && hard[k].text != " " {
				k++
			}
			spaces, word := hard[i:j], hard[j:k]
			i = k
			if len(word) == 0 {
				break
			}
			if empty && soft || curW+width(spaces) > w {
				spaces = nil
			}
			sw, ww := width(spaces), width(word)
			switch {
			case curW+sw+ww <= w:
				cur = append(append(cur, spaces...), word...)
				curW += sw + ww
			case !empty && width(rest)+ww <= w:
				breakLine()
				soft = true
				cur = append(cur, word...)
				curW += ww
			default:
				if !empty {
					breakLine()
					soft = true
				} else {
					cur = append(cur, spaces...)
					curW += sw
				}
				for _, g := range word {
					if curW+g.width > w && curW > width(rest) {
						breakLine()
						soft = true
					}
					cur = append(cur, g)
					curW += g.width
				}
			}
			empty = false
		}
	}
	out = append(out, cur)
	return out
}

// splitLines splits glyphs at hard breaks.
func splitLines(gs []glyph) [][]glyph {
	out := [][]glyph{nil}
	for _, g := range gs {
		if g.text == "\n" {
			out = append(out, nil)
			continue
		}
		out[len(out)-1] = append(out[len(out)-1], g)
	}
	return out
}

// chars breaks one line of glyphs every w columns, as a code block's.
func chars(gs []glyph, w int) []tline {
	if w <= 0 {
		return nil
	}
	out := []tline{{}}
	n := 0
	for _, g := range gs {
		if n+g.width > w {
			out = append(out, tline{})
			n = 0
		}
		out[len(out)-1].gs = append(out[len(out)-1].gs, g)
		n += g.width
	}
	return out
}
