package html

import (
	"slices"
	"strconv"

	"github.com/clipperhouse/uax29/v2/graphemes"

	"github.com/neuroplastio/hotty-a2ui/view"
)

// listKeys are the keys a HottyList gives the program, which works its
// selection, its pages and its filter (view.Controller.ListKey): those a
// Table gives, the arrows left and right that turn pages, and Space, which
// a filter types; a host that scrolls would take them. A focused box uses
// no keys (SPEC §10.2), so the characters a filter types, Backspace,
// Enter and Escape reach the program without a binding.
const listKeys = tableKeys + " ArrowLeft=program ArrowRight=program Space=program"

// richList is a HottyList on a host (profile §2): a focusable box, as a
// Table is, holding its title (or its filter while typed, with a caret),
// its status line, the items the view shows (the same page as cells),
// each an option whose id is the list's, "~i" and the item's index, and
// the page's dots. The box's keys are the program's, so the filter is
// typed as in cells: / and then the text.
func richList(e *view.Element) *node {
	box := el("div", "id", domID(e.ID), "class", "k-rich", "tabindex", "0", "role", "listbox", "data-keys", listKeys)
	if e.Label != "" {
		box.set("aria-label", e.Label)
	}
	switch {
	case e.Query.Editing:
		box.add(el("div", "id", partID(e.ID, partTitle), "class", "k-rich-filter").add(
			el("span", "class", "k-rich-prompt").add(txt("Filter:")),
			el("span", "class", "k-rich-query").add(texts(e.Query.Text)...),
			el("span", "class", "k-caret", "aria-hidden", "true")))
	case e.Label != "":
		box.add(el("div", "id", partID(e.ID, partTitle), "class", "k-rich-title").add(texts(e.Label)...))
	}
	box.add(el("div", "id", partID(e.ID, partStatus), "class", "k-rich-status", "aria-live", "polite").add(txt(e.ListStatus())))
	end := len(e.Shown)
	if e.Height > 0 {
		end = min(e.Top+e.Height, end)
	}
	sel := e.SelectedRow()
	for k := e.Top; k < end; k++ {
		i := e.Shown[k]
		it := e.Items[i]
		var match []int
		if k < len(e.Matched) {
			match = e.Matched[k]
		}
		item := el("div", "id", partID(e.ID, partItem+strconv.Itoa(i)), "class", "k-rich-item", "role", "option",
			"aria-selected", strconv.FormatBool(i == sel), "data-on", "click")
		if i == sel {
			item.set("class", "k-rich-item k-sel")
		}
		item.add(el("div", "class", "k-rich-label").add(marked(it.Label, match)...))
		if it.Description != "" {
			item.add(el("div", "class", "k-rich-desc").add(texts(it.Description)...))
		}
		box.add(item)
	}
	if len(e.Items) == 0 {
		box.add(el("div", "class", "k-rich-empty").add(texts(e.Placeholder)...))
	}
	// A paged list keeps its height, as in cells: hidden rows stand in
	// for the items a filter leaves out of a page, and for the dots
	// while one page is left.
	paged := e.Height > 0 && len(e.Items) > e.Height
	if paged {
		desc := slices.ContainsFunc(e.Items, func(it view.Entry) bool { return it.Description != "" })
		for range e.Height - (end - e.Top) {
			gap := el("div", "class", "k-rich-item k-rich-gap", "aria-hidden", "true").add(el("div", "class", "k-rich-label").add(txt("\u00a0")))
			if desc {
				gap.add(el("div", "class", "k-rich-desc").add(txt("\u00a0")))
			}
			box.add(gap)
		}
	}
	if pages, page := e.Pages(); pages > 1 || paged {
		// A HottyPaginator's dots (pageDots), without its clicks.
		dots := pageDots(e.ID, "k-rich-pages", pages, page, false)
		if pages == 1 {
			dots.set("class", "k-rich-pages k-rich-gap")
			dots.set("aria-hidden", "true")
		}
		box.add(dots)
	}
	return box
}

// marked is a label as nodes, the runs of clusters that hold a matched
// byte (view.Element.Matched) in a span of class k-match.
func marked(s string, at []int) []*node {
	if len(at) == 0 {
		return texts(s)
	}
	var out []*node
	run, on, off := "", false, 0
	flush := func() {
		if run == "" {
			return
		}
		if on {
			out = append(out, el("span", "class", "k-match").add(texts(run)...))
		} else {
			out = append(out, texts(run)...)
		}
		run = ""
	}
	for g := graphemes.FromString(s); g.Next(); {
		v := g.Value()
		hit := slices.ContainsFunc(at, func(b int) bool { return b >= off && b < off+len(v) })
		if hit != on {
			flush()
			on = hit
		}
		run += v
		off += len(v)
	}
	flush()
	return out
}
