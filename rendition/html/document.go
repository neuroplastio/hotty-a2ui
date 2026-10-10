package html

import (
	"github.com/neuroplastio/hotty-a2ui/view"
)

// A HottyMarkdown on a host (profile §2, §6.27): a Text's HTML, in a div
// of class k-text k-doc. Its headings have ids (headingID) and tabindex -1,
// so that the program's a=focus takes the keyboard to one a link in place
// went to, which a scrolling surface scrolls into view (SPEC §5.3, §10.1).
// A link in place is a span, not a link: a Tab stop (tabindex 0) whose
// click is reported with its id (data-on=click, SPEC §9), Enter the
// program's (data-keys), which follows it as cells does. A link with a URL
// is a hyperlink, the terminal's (target=_blank). An alert is a div of
// role note, its title row its icon and its title, in its tone (kit.css).

// linkKeys is a link in place's keymap: Enter reaches the program, which
// follows the link. Space stays the surface's, which scrolls with it.
const linkKeys = "Enter=program"

// document is a HottyMarkdown's HTML.
func document(e *view.Element) string {
	return view.DocHTML(e.Markdown, view.DocMarkup{
		Heading: func(i int) []string {
			if i >= len(e.Anchors) {
				return nil
			}
			return []string{"id", headingID(e.ID, e.Anchors[i]), "tabindex", "-1"}
		},
		Link: func(i int) (string, string) {
			s := el("span", "id", domID(view.SubID(e.ID, "link", i)), "class", "k-link", "role", "link", "tabindex", "0",
				"data-on", "click", "data-keys", linkKeys).html()
			return s[:len(s)-len("</span>")], "</span>"
		},
		Alert: alert,
	})
}

// pageDocument is a HottyMarkdown's HTML on a page (Rendition.Page), where
// no program follows a link: a link in place is a link, to its href, which
// the page's links decide (pageLink: a path or a fragment in the same tab),
// and a heading's id is its anchor itself, GitHub's, so that a #fragment,
// the page's own or another's link to it, finds it.
func pageDocument(e *view.Element) string {
	return view.DocHTML(e.Markdown, view.DocMarkup{
		Heading: func(i int) []string {
			if i >= len(e.Anchors) {
				return nil
			}
			return []string{"id", e.Anchors[i]}
		},
		Link: func(i int) (string, string) {
			href := ""
			if i < len(e.Children) {
				href = e.Children[i].URL
			}
			a := el("a", "href", href).html()
			return a[:len(a)-len("</a>")], "</a>"
		},
		Alert: alert,
	})
}

// alert is the tags around an alert of its kind: a div of role note, its
// title row (its icon and its title) first.
func alert(kind string) (string, string) {
	a := view.Alerts[kind]
	title := el("p", "class", "k-alert-title").add(beside(a.Icon)).add(texts(a.Title)...)
	box := el("div", "class", "k-alert k-alert-"+kind, "role", "note").html()
	return box[:len(box)-len("</div>")] + title.html(), "</div>"
}
