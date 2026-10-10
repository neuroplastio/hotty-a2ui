package html

import (
	"strconv"

	"github.com/neuroplastio/hotty-a2ui/icons"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// icon is an Icon or a HottyIcon as a host draws it (gov R-4): its shape,
// one of the 59, a Material Symbols name registered (icons.Named) or an
// svgPath, as an inline <svg> 1em square, filled with the text's colour or
// stroked with it; else, for a name the kit has no shape for or a path it
// doesn't accept, the glyph cells draw. The svg is the kit's own, so an
// svgPath is only ever the d of its one path.
func icon(e *view.Element, id string) *node { return shape(e.Name, e.Path, e.Stroke, id) }

// shape is an icon by its name, or its path when it has one, stroked
// stroke wide when that is more than 0, as icon draws it; id is its DOM id,
// if it has one.
func shape(name, path string, stroke float64, id string) *node {
	ic, ok := icons.Basic(name)
	if !ok {
		ic, ok = icons.Named(name)
	}
	label := name
	if path != "" {
		ic, ok = icons.Path(path)
		ic.Stroke = stroke
		label = "icon"
	}
	n := el("span", "id", id, "class", "k-icon", "role", "img", "aria-label", label)
	if !ok {
		return n.add(txt(icons.Glyph(name)))
	}
	box := strconv.Itoa(ic.Box)
	p := el("path", "fill", "currentColor", "d", ic.Path)
	if ic.Stroke > 0 {
		p = el("path", "fill", "none", "stroke", "currentColor", "stroke-width", strconv.FormatFloat(ic.Stroke, 'f', -1, 64),
			"stroke-linecap", "round", "stroke-linejoin", "round", "d", ic.Path)
	}
	return n.add(el("svg", "viewBox", "0 0 "+box+" "+box, "width", "1em", "height", "1em", "aria-hidden", "true").add(p))
}

// beside is the icon a Tabs' title or a ChoicePicker's option has beside
// its label (io_neuroplast_hotty.icons), by its name: hidden from a
// screen reader, as the label says it; nil for none.
func beside(name string) *node {
	if name == "" {
		return nil
	}
	return shape(name, "", 0, "").set("aria-hidden", "true")
}
