package html

import (
	"strconv"

	"github.com/neuroplastio/hotty-a2ui/view"
)

// scrollView is a HottyScrollView on a host (profile §2), a baseline: a
// focusable box as many terminal rows tall as its height (the host's
// --hotty-cell-h, SPEC §8), whose content overflows it and which the
// host scrolls itself, with its scrollbar, the wheel, a touch drag and
// the keys a browser scrolls with (SPEC §5.3; the storybook's surfaces
// ask to scroll), and with bubbles' letters too (scrollKeys). Lines are a
// row each, in the terminal's font, so that the box shows height of them;
// a log's (follow) is role=log, which a screen reader reads as lines
// arrive. A program cannot set where a host has scrolled, so the box
// starts at its top, follow or not, and hottyScrollTo does nothing there:
// vault KIT-07h.
func (m *markup) scrollView(e *view.Element) *node {
	keys := scrollKeys
	if len(e.Children) == 0 && !e.Wrap {
		keys += sideKeys
	}
	n := el("div", "id", domID(e.ID), "class", "k-scroll", "tabindex", "0",
		"style", "--k-rows: "+strconv.Itoa(e.Height), "data-keys", keys)
	if len(e.Children) > 0 {
		return n.add(m.all(e.Children)...)
	}
	if e.Active {
		n.set("role", "log")
	}
	class := "k-lines"
	if e.Wrap {
		class += " k-wrap"
	}
	lines := el("div", "id", partID(e.ID, partLines), "class", class)
	for _, l := range e.Lines {
		row := el("div", "class", "k-line")
		if l == "" {
			row.add(txt(" "))
		} else {
			row.add(texts(l)...)
		}
		lines.add(row)
	}
	return n.add(lines)
}

// scrollKeys bind the keys bubbles' viewport scrolls with, past the ones a
// browser does, to the host's scroll actions (SPEC §10.2, Scrolling keys),
// so that the box scrolls with them on a host as it does in cells (profile
// §3.7). Its buttons keep Space and its fields type the letters: a key an
// element uses stays its own, and a field leaves scroll actions out.
// Shift+Space is bound too, or it would page down as Space does.
const scrollKeys = "j=scroll-down k=scroll-up f=scroll-page-down b=scroll-page-up " +
	"Space=scroll-page-down Shift+Space=scroll-page-up " +
	"u=scroll-half-page-up d=scroll-half-page-down Control+u=scroll-half-page-up Control+d=scroll-half-page-down " +
	"g=scroll-start G=scroll-end"

// sideKeys scroll lines that do not wrap sideways (h and l), as in cells,
// where the two fall through when the lines wrap.
const sideKeys = " h=scroll-left l=scroll-right"
