package html

import (
	"strings"

	"github.com/neuroplastio/hotty-a2ui/view"
)

// keyHints is a HottyKeyHints on a host (profile §2), a baseline as cells
// draws it: the short view a line, each key a kbd and what it does a span,
// " • " between them; the full view its groups in columns, a grid of keys
// and what they do each. The host moves focus among the elements it works
// without telling the renderer, so their keys are left out; a box's, a
// Slider's and a select's show, as the program's keys go there
// (view.Controller.KeyHints). Each key is a cap (keycaps). On a surface with
// descriptions, a row comes first for the tooltip (profile §6.23): the
// description of the element the pointer is over, where the host says
// (SPEC §9.4), else of the one with the keyboard; empty, a line high,
// when neither has one, so that nothing moves.
func (m *markup) keyHints(e *view.Element) *node {
	n := el("div", "id", domID(e.ID), "class", "k-hints")
	if m.described {
		n.add(el("div", "id", partID(e.ID, partTip), "class", "k-hints-tip").add(texts(m.tip)...))
	}
	if !e.Open {
		line := el("div", "class", "k-hints-short")
		for i, h := range m.short {
			if i > 0 {
				line.add(el("span", "class", "k-hint-sep", "aria-hidden", "true").add(txt(" • ")))
			}
			line.add(el("span", "class", "k-hint").add(
				m.keycaps("k-hint-key", h.Key),
				el("span", "class", "k-hint-desc").add(texts(" "+h.Desc)...)))
		}
		return n.add(line)
	}
	full := el("div", "class", "k-hints-full")
	for _, g := range m.full {
		col := el("div", "class", "k-hints-col")
		for _, h := range g {
			col.add(m.keycaps("k-hint-key", h.Key),
				el("span", "class", "k-hint-desc").add(texts(h.Desc)...))
		}
		full.add(col)
	}
	return n.add(full)
}

// keycaps are a hint's keys as caps (view.Controller.Caps, KIT-08h), in a
// kbd of class cls: a kbd of class k-cap for each chord, its labels joined
// by its dim separator ("+", say) unless they sit tight (⌘⇧P), and a dim
// ">" between a sequence's chords (view.Keycaps.Between; a blank one a
// span.k-cap-gap, a gap alone) and "/" between alternatives.
func (m *markup) keycaps(cls, keys string) *node {
	n := el("kbd", "class", cls+" k-caps")
	var alts [][]view.Chord
	if m.caps != nil {
		alts = m.caps(keys)
	} else {
		alts = (&view.Controller{}).Caps(keys, true)
	}
	then := m.then
	if then == "" {
		then = view.Keycaps{}.Between()
	}
	sep := func(s string) *node {
		if strings.TrimSpace(s) == "" {
			return el("span", "class", "k-cap-gap", "aria-hidden", "true")
		}
		return el("span", "class", "k-cap-sep", "aria-hidden", "true").add(txt(s))
	}
	for i, seq := range alts {
		if i > 0 {
			n.add(sep("/"))
		}
		for j, ch := range seq {
			if j > 0 {
				n.add(sep(then))
			}
			cp := el("kbd", "class", "k-cap")
			for k, l := range ch.Labels {
				if ch.Joined(k) {
					cp.add(el("span", "class", "k-cap-join").add(txt(ch.Sep)))
				}
				cp.add(txt(l))
			}
			n.add(cp)
		}
	}
	return n
}
