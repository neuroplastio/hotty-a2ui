package html

import (
	"strconv"

	"github.com/neuroplastio/hotty-a2ui/view"
)

// pagerKeys are the keys a HottyPaginator gives the program, which turns
// its pages (view.Controller.PageKey): the arrows left and right, Page Up
// and Page Down, Home and End, which a host that scrolls would take. A
// focused box uses no keys (SPEC §10.2), so h and l reach the program
// without a binding.
const pagerKeys = "ArrowLeft=program ArrowRight=program PageUp=program PageDown=program Home=program End=program"

// paginator is a HottyPaginator on a host (profile §2), a baseline: a
// focusable box, as a HottyList's, holding its child, which the view gave
// the page shown, and then its pages, a dot each that takes a click, or
// the page of how many. The box's keys are the program's, so the pages
// turn as in cells. Its name is its accessibility.label, else "Pages".
func (m *markup) paginator(e *view.Element) *node {
	n := el("div", "id", domID(e.ID), "class", "k-pager", "tabindex", "0", "role", "group", "aria-label", "Pages",
		"data-keys", pagerKeys).add(m.all(e.Children)...)
	if len(e.Children) == 0 {
		n.set("class", "k-pager k-bare")
	}
	if e.Variant == "numbers" {
		return n.add(el("div", "id", partID(e.ID, partDots), "class", "k-pages k-page-n", "role", "status").add(txt(e.PageText())))
	}
	return n.add(pageDots(e.ID, "k-pages", e.PageCount, e.Selected, true))
}

// pageDots are a paginator's pages, a HottyPaginator's or a HottyList's,
// as cells draws them: a dot a page ("~d0", …), the page shown's k-on, in
// a row (id "~d", of class) that says which page shows. A HottyPaginator's
// dots take a click, which shows their page; a HottyList's do not, as
// bubbles' list's do not. The row is a status: a screen reader says the
// page as it turns.
func pageDots(id, class string, pages, page int, click bool) *node {
	n := el("div", "id", partID(id, partDots), "class", class, "role", "status",
		"aria-label", "Page "+strconv.Itoa(page+1)+" of "+strconv.Itoa(pages))
	for p := range pages {
		class := "k-dot"
		if p == page {
			class += " k-on"
		}
		dot := el("span", "class", class)
		if click {
			dot.set("id", partID(id, partDots+strconv.Itoa(p))).set("data-on", "click")
		}
		n.add(dot.add(txt("•")))
	}
	return n
}
