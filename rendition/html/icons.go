package html

import (
	"strconv"

	"github.com/neuroplastio/hotty-a2ui/icons"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// icon is an Icon as a host draws it (gov R-4): its shape, one of the 59
// or an svgPath, as an inline <svg> 1em square, filled with the text's
// colour; else, for a name the kit has no shape for or a path it doesn't
// accept, the glyph cells draw. The svg is the kit's own, so an svgPath is
// only ever the d of its one path.
func icon(e *view.Element, id string) *node {
	ic, ok := icons.Basic(e.Name)
	label := e.Name
	if e.Path != "" {
		ic, ok = icons.Path(e.Path)
		label = "icon"
	}
	n := el("span", "id", id, "class", "k-icon", "role", "img", "aria-label", label)
	if !ok {
		return n.add(txt(icons.Glyph(e.Name)))
	}
	box := strconv.Itoa(ic.Box)
	return n.add(el("svg", "viewBox", "0 0 "+box+" "+box, "width", "1em", "height", "1em", "aria-hidden", "true").
		add(el("path", "fill", "currentColor", "d", ic.Path)))
}
