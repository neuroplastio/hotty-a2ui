package view

import (
	"github.com/neuroplastio/hotty-a2ui/a2ui"
)

// HottyScrollView (profile §6.11) in the view: a box of a fixed height
// whose content scrolls, as bubbles' viewport: a child, or lines as data.
// Where it is scrolled is the surface's state, which cells works (rendition
// /cells) and hottyScrollTo sets; a host scrolls it itself (SPEC §5.3).

// scrollHeight is a HottyScrollView's rows without a height.
const scrollHeight = 10

// mapScrollView makes a HottyScrollView's element: its child, or its lines
// (a line each, as text); Height, the rows it shows; Top and Left, where
// it is scrolled to; Active, its follow prop; Tail, whether it follows the
// tail, as follow says until the user scrolls; Wrap, whether long lines
// wrap.
func mapScrollView(b *Builder, n *a2ui.Node) *Element {
	e := &Element{Kind: ScrollView, Children: b.Children(n.Props["child"]), Wrap: b.Bool(n, "wrap"), Height: scrollHeight,
		Active: b.Bool(n, "follow")}
	if len(e.Children) == 0 {
		lines, _ := b.Raw(n, "lines").([]any)
		e.Lines = make([]string, 0, len(lines))
		for _, l := range lines {
			if l == nil {
				// An index past the end pads with nulls (vault a2ui-limits L4).
				e.Lines = append(e.Lines, "")
				continue
			}
			e.Lines = append(e.Lines, a2ui.ToString(l))
		}
	}
	if h := a2ui.ToNumber(b.Raw(n, "height")); h >= 1 {
		e.Height = int(h)
	}
	e.Top, e.Left = b.St.Scroll[n.Key], b.St.Left[n.Key]
	tail, set := b.St.Tail[n.Key]
	if !set {
		tail = e.Active
	}
	e.Tail = tail
	return e
}

// ScrollTo scrolls a HottyScrollView to its start or its end (to "start" or
// "end"), as hottyScrollTo does. At its end, one that follows (Active)
// follows the tail again, as it does once the user scrolls there; one
// that does not is at its end once, and stays where it is as lines arrive.
func (c *Controller) ScrollTo(id, to string) {
	switch to {
	case "start":
		c.St.Scroll[id], c.St.Left[id], c.St.Tail[id] = 0, 0, false
	default:
		c.St.Tail[id] = true
	}
	c.Rebuild()
}

// Scrolled records where a rendition scrolled a HottyScrollView: top, its
// first row shown; left, its first column; tail, whether that is the end,
// so that it follows what comes. The cells rendition clamps them to its
// content, and calls this when they change.
func (c *Controller) Scrolled(id string, top, left int, tail bool) {
	c.St.Scroll[id], c.St.Left[id], c.St.Tail[id] = top, left, tail
	if e := c.V.Find(id); e != nil {
		e.Top, e.Left, e.Tail = top, left, tail
	}
}
