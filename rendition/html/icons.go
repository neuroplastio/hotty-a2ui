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
func icon(e *view.Element, id string) *node { return shape(e.Name, e.Path, id) }

// shape is an icon by its name, or its path when it has one, as icon
// draws it; id is its DOM id, if it has one.
func shape(name, path, id string) *node {
	ic, ok := icons.Basic(name)
	label := name
	if path != "" {
		ic, ok = icons.Path(path)
		label = "icon"
	}
	n := el("span", "id", id, "class", "k-icon", "role", "img", "aria-label", label)
	if !ok {
		return n.add(txt(icons.Glyph(name)))
	}
	box := strconv.Itoa(ic.Box)
	return n.add(el("svg", "viewBox", "0 0 "+box+" "+box, "width", "1em", "height", "1em", "aria-hidden", "true").
		add(el("path", "fill", "currentColor", "d", ic.Path)))
}
