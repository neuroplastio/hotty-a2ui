package html

import (
	"github.com/neuroplastio/hotty-a2ui/view"
)

// keyHints is a HottyKeyHints on a host (profile §2), a baseline as cells
// draws it: the short view a line, each key a kbd and what it does a span,
// " • " between them; the full view its groups in columns, a grid of keys
// and what they do each. The host moves focus among the elements it works
// without telling the renderer, so their keys are left out; a box's, a
// Slider's and a select's show, as the program's keys go there
// (view.Controller.KeyHints). Keycaps are vault KIT-08h.
func (m *markup) keyHints(e *view.Element) *node {
	n := el("div", "id", domID(e.ID), "class", "k-hints")
	if !e.Open {
		line := el("div", "class", "k-hints-short")
		for i, h := range m.short {
			if i > 0 {
				line.add(el("span", "class", "k-hint-sep", "aria-hidden", "true").add(txt(" • ")))
			}
			line.add(el("span", "class", "k-hint").add(
				el("kbd", "class", "k-hint-key").add(texts(h.Key)...),
				el("span", "class", "k-hint-desc").add(texts(" "+h.Desc)...)))
		}
		return n.add(line)
	}
	full := el("div", "class", "k-hints-full")
	for _, g := range m.full {
		col := el("div", "class", "k-hints-col")
		for _, h := range g {
			col.add(el("kbd", "class", "k-hint-key").add(texts(h.Key)...),
				el("span", "class", "k-hint-desc").add(texts(h.Desc)...))
		}
		full.add(col)
	}
	return n.add(full)
}
