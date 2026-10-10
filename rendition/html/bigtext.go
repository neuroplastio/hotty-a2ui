package html

import (
	"strings"

	"github.com/neuroplastio/hotty-a2ui/view"
)

// bigText is a HottyBigText (profile §2): its text as it is, in large
// display type, its size and its align classes (kit.css); its lines apart
// at a br each, as cells breaks them, and a line too long wraps as text
// does. A screen reader reads the text, as it reads any.
func bigText(e *view.Element) *node {
	class := "k-big k-big-" + e.Variant
	if e.Align == "center" || e.Align == "end" {
		class += " k-big-" + e.Align
	}
	n := el("div", "id", domID(e.ID), "class", class)
	for i, l := range strings.Split(e.Label, "\n") {
		if i > 0 {
			n.add(el("br"))
		}
		n.add(texts(l)...)
	}
	return n
}
