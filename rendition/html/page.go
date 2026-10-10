package html

import (
	_ "embed"
	stdhtml "html"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/neuroplastio/hotty-a2ui/view"
)

// Page mode (profile §2.1): a surface in an ordinary web page, not a
// terminal, which a reader or a crawler reads with no program behind it.
// The markup is a host's, with what only a program makes work left out
// (a host's events and keys) and what a program would show one piece at a
// time shown whole: every tab, every node of a tree, every row.

//go:embed page.css
var pageSheet string

// PageCSS is the kit's stylesheet for a web page: the host's sheet, then
// page mode's rules. A page puts it in once, in its head, however many
// surfaces it shows (Page): in a style element, or a file it links.
//
// The colours are the page's: it sets the host's palette (SPEC §8's
// --hotty-bg, --hotty-fg, --hotty-accent, --hotty-ansi-8, and -9, -2, -3,
// -6 for error, success, warning and info) on any element around its
// surfaces, and the kit's own variables follow; what it leaves out is the
// kit's default, light or dark as the page's color-scheme says. The type
// is the page's too.
func PageCSS() string { return kitCSS + "\n" + pageSheet }

// Page is the surface's markup for an ordinary web page: a fragment for
// the page's body, under the stylesheet the page puts in once (PageCSS).
// It needs no program: nothing in it is read through a host's events or
// keys, which it leaves out. Every tab of a Tabs shows, each a section
// under its title; a HottyTree's every node is a row, its branches open or
// closed as given, which the reader opens and closes; links are links.
// Controls show their state and take no input (profile §2.1 lists what
// each shows). The surface's name is its top element's id and prefixes
// every id in it, so that a page's surfaces share none; a theme
// (SetTheme) is set on it as on a host. The view is built from the
// controller's surface and state (view.BuildPage), not taken from its
// view, which holds only the tab shown.
func (r *Rendition) Page() string {
	now := time.Now
	if r.Clock != nil {
		now = r.Clock
	}
	v := view.BuildPage(r.C.S, r.C.St)
	m := &markup{page: true, now: now()}
	n := el("div", "class", "k-surface k-page").add(m.element(v.Root))
	if css := themeCSS(r.theme); css != "" {
		// A surface in a theme's colours is a box of them on the page.
		n.set("class", "k-surface k-page k-themed").set("style", css)
	}
	prefix := ""
	if r.name != "" {
		prefix = domID(r.name) + "~~"
	}
	static(n, prefix)
	if r.name != "" {
		n.id = domID(r.name)
	}
	return n.html()
}

// onPage is an element as a page draws it where a page differs from a
// host by the element's fields alone: a Table's and a HottyList's every
// row, not a window of them, with no filter being typed; a Modal closed.
// Nil for what a page leaves out: a HottyKeyHints, whose keys reach no
// program there.
func onPage(e *view.Element) *view.Element {
	switch e.Kind {
	case view.KeyHints:
		return nil
	case view.Table, view.RichList:
		c := *e
		c.Height, c.Top, c.Query.Editing = 0, 0, false
		return &c
	case view.Modal:
		c := *e
		c.Open = false
		return &c
	}
	return e
}

// pageTabs is a Tabs on a page: every tab, each a section under its title,
// as GitHub shows the same Markdown, where a host shows one tab at a time
// and a program switches them. The title is drawn as a tab bar of one tab,
// the tab shown's look.
func (m *markup) pageTabs(e *view.Element) *node {
	n := el("div", "id", domID(e.ID), "class", "k-tabs k-stack k-col")
	for _, t := range e.Children {
		if t.Kind != view.Tab {
			continue
		}
		title := el("div", "id", domID(t.ID), "class", "k-tab k-tab-title").add(beside(t.Name)).add(texts(t.Label)...)
		n.add(el("section", "class", "k-tab-section k-stack k-col", "aria-labelledby", domID(t.ID)).add(
			el("div", "class", "k-tablist").add(title),
			el("div", "class", "k-tabpanel k-stack k-col").add(m.all(t.Children)...)))
	}
	return n
}

// pageTree is a HottyTree on a page: every node a row, as on a host, in
// lists that nest as the nodes do, so that a reader with no program sees
// every node and follows every link. A branch is a details element, open
// or closed as given (expanded, and the branches the selection is in),
// whose summary is its row: the browser opens and closes it, and the
// sheet draws its fold (▸ or ▾) and its count, which shows while it is
// closed. A filter is a program's, so a page draws none. The selected node
// is marked as on a host; its label says it is the current one.
func (m *markup) pageTree(e *view.Element) *node {
	box := el("div", "id", domID(e.ID), "class", "k-tree")
	if e.Height > 0 {
		box.set("class", "k-tree k-scrolls").set("style", "--k-rows: "+strconv.Itoa(e.Height))
	}
	if len(e.Nodes) == 0 {
		return box.add(el("div", "class", "k-tree-empty").add(texts(e.Placeholder)...))
	}
	sel := e.SelectedRow()
	icons := slices.ContainsFunc(e.Nodes, func(n view.TreeNode) bool { return n.Icon != "" })
	flat := e.Flat()
	top := el("ul", "class", "k-tree-list", "role", "list")
	// lists[l] is the list a node of level l goes in: the nodes are in
	// pre-order, so a node's parent opened the list before it.
	lists := []*node{top}
	for i, n := range e.Nodes {
		lists = lists[:min(n.Level, len(lists)-1)+1]
		row := el("div", "id", partID(e.ID, partNode+strconv.Itoa(i)), "class", "k-node", "style", "--k-level: "+strconv.Itoa(n.Level))
		if i == sel {
			row.set("class", "k-node k-sel")
		}
		if !flat {
			row.add(el("span", "class", "k-node-fold", "aria-hidden", "true"))
		}
		switch {
		case n.Icon != "":
			row.add(shape(n.Icon, "", 0, ""))
		case icons:
			row.add(el("span", "class", "k-icon k-icon-blank", "aria-hidden", "true"))
		}
		// KIT-26: a node's href, once view.TreeNode has one, goes here in
		// place of "".
		row.add(treeLabel(n.Label, "", i == sel))
		item := el("li")
		if n.Branch() {
			row.tag = "summary"
			row.add(el("span", "class", "k-node-count").add(txt(strconv.Itoa(n.Kids))))
			kids := el("ul", "class", "k-tree-list", "role", "list")
			item.add(el("details").flag("open", n.Open).add(row, kids))
			lists[len(lists)-1].add(item)
			lists = append(lists, kids)
			continue
		}
		lists[len(lists)-1].add(item.add(row))
	}
	return box.add(top)
}

// treeLabel is a HottyTree node's label on a page: a link (link) when the
// node has an href, the current page's when the node is the one selected;
// else its text, the current one's when selected.
func treeLabel(label, href string, selected bool) *node {
	if a := link(href); a != nil {
		a.set("class", "k-node-label")
		if selected {
			a.set("aria-current", "page")
		}
		return a.add(texts(label)...)
	}
	n := el("span", "class", "k-node-label").add(texts(label)...)
	if selected {
		n.set("aria-current", "true")
	}
	return n
}

// pagePaginator is a HottyPaginator on a page: its child holding every
// page's items, one after another, as a page reads them, and no dots,
// where a host shows a page at a time. A bare one's pages are the
// agent's, so it says which shows, as on a host.
func (m *markup) pagePaginator(e *view.Element) *node {
	if len(e.Children) == 0 {
		return m.paginator(e)
	}
	child := *e.Children[0]
	if child.Kind == view.Stack {
		child.Children = slices.Concat(e.Paged...)
	}
	return el("div", "id", domID(e.ID), "class", "k-pager").add(m.element(&child))
}

// link is an <a> to href as a page has it (pageLink); nil when href is
// "" or goes nowhere a page links to.
func link(href string) *node {
	if href == "" {
		return nil
	}
	a := el("a", "href", href)
	pageLink(a)
	if _, ok := a.attr("href"); !ok {
		return nil
	}
	return a
}

// pageLink makes an <a> a page's link, in place: the one place where page
// mode decides where links go and how they open. A link that leaves the
// site (http or https, or //host) opens in a new tab, and any other (a
// path, a fragment, a query, mailto: or tel:) in the same one. One to any
// other scheme (javascript:, data:) loses its href. What the host's
// markup said (Text's Markdown opens every link in the terminal, with
// target=_blank) and any event of a host's on it (data-*) go.
func pageLink(a *node) {
	href, _ := a.attr("href")
	a.attrs = slices.DeleteFunc(a.attrs, func(at attr) bool {
		return at.k == "target" || at.k == "rel" || at.k == "href" || strings.HasPrefix(at.k, "data-")
	})
	h, away, ok := pageHref(href)
	if !ok {
		return
	}
	a.set("href", h)
	if away {
		a.set("target", "_blank").set("rel", "noopener")
	}
}

// pageHref is where a page's link goes, and whether it leaves the site;
// ok is false for one a page doesn't link to.
func pageHref(href string) (h string, away, ok bool) {
	h = strings.TrimSpace(href)
	u, err := url.Parse(h)
	if h == "" || err != nil {
		return "", false, false
	}
	switch strings.ToLower(u.Scheme) {
	case "":
		return h, u.Host != "", true
	case "http", "https":
		return h, true, true
	case "mailto", "tel":
		return h, false, true
	}
	return "", false, false
}

var (
	// anchorTag is an <a> start tag in HTML goldmark wrote, which escapes
	// every attribute's value, so none holds ">".
	anchorTag = regexp.MustCompile(`<a(\s[^>]*)?>`)
	tagAttr   = regexp.MustCompile(`([^\s="]+)(?:="([^"]*)")?`)
)

// pageLinks makes the links in a Text's HTML a page's (pageLink).
func pageLinks(h string) string {
	return anchorTag.ReplaceAllStringFunc(h, func(tag string) string {
		a := el("a")
		for _, m := range tagAttr.FindAllStringSubmatch(tag[2:len(tag)-1], -1) {
			a.set(m[1], stdhtml.UnescapeString(m[2]))
		}
		pageLink(a)
		return strings.TrimSuffix(a.html(), "</a>")
	})
}

// static makes a page's markup work with no program behind it, in place:
//   - links are a page's (pageLink), in a Text's HTML too;
//   - a host's events and keys go (data-on, data-keys, data-steps), and so
//     does what only a program's keyboard makes true: tabindex on its boxes,
//     an open list's ids, a datalist of suggestions;
//   - a box of rows a program selects in (grid, listbox, option) is a box
//     of rows, its selected row the current one;
//   - buttons and checkboxes leave the Tab order (the sheet stops their
//     clicks), and a text field is readonly: its text can be selected and
//     copied, not changed;
//   - every id, and every attribute that names one, gets prefix.
func static(n *node, prefix string) {
	if n.tag == "" {
		return
	}
	if n.raw {
		n.text = pageLinks(n.text)
	}
	if n.id != "" {
		n.id = prefix + n.id
	}
	var attrs []attr
	for _, a := range n.attrs {
		switch {
		case strings.HasPrefix(a.k, "data-"), a.k == "tabindex", a.k == "list", a.k == "aria-controls",
			a.k == "aria-activedescendant", a.k == "aria-haspopup", a.k == "aria-expanded",
			a.k == "aria-rowcount", a.k == "aria-rowindex":
			continue
		case a.k == "role" && (a.v == "grid" || a.v == "listbox" || a.v == "option"):
			continue
		case a.k == "aria-selected":
			if a.v == "true" {
				attrs = append(attrs, attr{"aria-current", "true"})
			}
			continue
		case a.k == "for" || a.k == "aria-labelledby" || a.k == "aria-describedby":
			ids := strings.Fields(a.v)
			for i := range ids {
				ids[i] = prefix + ids[i]
			}
			a.v = strings.Join(ids, " ")
		}
		attrs = append(attrs, a)
	}
	n.attrs = attrs
	switch typ, _ := n.attr("type"); {
	case n.tag == "a":
		pageLink(n)
	case n.tag == "button", n.tag == "input" && typ == "checkbox":
		n.set("tabindex", "-1")
	case n.tag == "input", n.tag == "textarea":
		n.flag("readonly", true)
	}
	n.kids = slices.DeleteFunc(n.kids, func(k *node) bool { return k.tag == "datalist" })
	for _, k := range n.kids {
		static(k, prefix)
	}
}
