package view

import (
	"math"
	"strconv"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
)

// A HottyPaginator (profile §6.26) in the view: the page that shows, of
// how many, as bubbles' paginator keeps them. With a child, a List, Column
// or Row, the view cuts the child's items into pages, as bubbles' program
// slices its items (GetSliceBounds), and the child holds the page shown,
// so that every rendition shows the same items and Tab goes through those
// alone. Without one the pages are the agent's: it hears of a turn by
// onChange and sends the page's data.

const (
	// perPage is the items a page shows without perPage, as bubbles'
	// paginator example has it.
	perPage = 10
	// maxPages caps pages as A2UI caps a list's index (a2ui.MaxListIndex).
	maxPages = a2ui.MaxListIndex
)

// mapPaginator makes a HottyPaginator's element: its child, holding the
// page shown, its items cut into pages (Paged); or, without one, its
// pages; the page shown, from page (from 1), held to them; and dots or
// numbers.
func mapPaginator(b *Builder, n *a2ui.Node) *Element {
	e := &Element{Kind: Paginator, Variant: "dots", Height: perPage}
	if b.Enum(n, "displayStyle", "dots") == "numbers" {
		e.Variant = "numbers"
	}
	if v := a2ui.ToNumber(b.Raw(n, "perPage")); v >= 1 {
		e.Height = int(min(v, maxPages))
	}
	pages := 1
	if kids := b.Children(n.Props["child"]); len(kids) > 0 {
		e.Children = kids[:1]
		items := kids[:1]
		if c := kids[0]; c.Kind == Stack {
			items = c.Children
		}
		for i := 0; i < len(items); i += e.Height {
			e.Paged = append(e.Paged, items[i:min(i+e.Height, len(items))])
		}
		pages = max(len(e.Paged), 1)
	} else if v := a2ui.ToNumber(b.Raw(n, "pages")); v >= 1 {
		pages = int(min(v, maxPages))
	}
	at := 0
	if v := a2ui.ToNumber(b.Value(n, "page")); !math.IsNaN(v) && v >= 1 {
		at = int(min(v, maxPages)) - 1
	}
	e.PageCount, e.Selected = pages, min(at, pages-1)
	if c := e.pagerChild(); c != nil && c.Kind == Stack {
		c.Children = nil
		if e.Selected < len(e.Paged) {
			c.Children = e.Paged[e.Selected]
		}
	}
	return e
}

// pagerChild is a HottyPaginator's child, nil without one.
func (e *Element) pagerChild() *Element {
	if len(e.Children) == 0 {
		return nil
	}
	return e.Children[0]
}

// PageText is the page a HottyPaginator shows, of how many, as bubbles'
// arabic paginator writes it: "3/10".
func (e *Element) PageText() string {
	return strconv.Itoa(e.Selected+1) + "/" + strconv.Itoa(e.PageCount)
}

// TurnPage shows page at (from 0) of a HottyPaginator, held to its pages:
// written to where page is bound (from 1), else to the renderer's state;
// then onChange runs, whose context reads it. A page it shows already
// changes nothing and sends nothing.
func (c *Controller) TurnPage(id string, at int) error {
	e := c.V.Find(id)
	if e == nil || e.Kind != Paginator {
		return nil
	}
	if at = min(max(at, 0), e.PageCount-1); at == e.Selected {
		return nil
	}
	err := c.setProp(e, "page", float64(at+1))
	c.Rebuild()
	if n := c.V.Node(id); err == nil && n != nil && n.Props["onChange"] != nil {
		err = c.S.Tree.Invoke(n, "onChange", true)
		c.Rebuild()
	}
	return err
}

// PageKey works a HottyPaginator by a key, named as SPEC §10.4 has it, as
// bubbles' paginator takes them: the arrow left, h and Page Up show the
// page before, the arrow right, l and Page Down the page after, none past
// the first or the last; Home and End the first and the last, as a
// Slider's ends. ok reports whether it took the key, at an end too.
func (c *Controller) PageKey(id, key string) (ok bool, err error) {
	e := c.V.Find(id)
	if e == nil || e.Kind != Paginator {
		return false, nil
	}
	to := e.Selected
	switch key {
	case "ArrowLeft", "h", "PageUp":
		to--
	case "ArrowRight", "l", "PageDown":
		to++
	case "Home":
		to = 0
	case "End":
		to = e.PageCount - 1
	default:
		return false, nil
	}
	return true, c.TurnPage(id, to)
}
