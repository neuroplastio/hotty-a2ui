package html

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	hottycat "github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// A heading's DOM id is its document's, "~#" and its ID, encoded: it reads
// back as its document's, and no other id has "~#".
func TestHeadingIDs(t *testing.T) {
	for _, tc := range []struct{ doc, anchor string }{{"page", "install"}, {"row[/items/0]#2", "usage-1"}, {"ü", "café-crème"}} {
		dom := headingID(tc.doc, tc.anchor)
		if strings.ContainsAny(dom, " :;=[]") || strings.Count(dom, "#") != 1 {
			t.Errorf("%q: %q is not a control value", tc.anchor, dom)
		}
		if id, part, ok := viewID(dom); !ok || id != tc.doc || part != "#"+domID(tc.anchor) {
			t.Errorf("%q: read back %q %q %v", dom, id, part, ok)
		}
		if strings.Contains(domID(tc.doc), "~#") {
			t.Errorf("%q's DOM id has ~#", tc.doc)
		}
	}
	if _, _, ok := viewID("~#usage"); ok {
		t.Error("a heading with no document reads as one")
	}
}

// A HottyMarkdown on a host: its headings have ids and take the program's
// focus; a link in place is a Tab stop whose click and Enter reach the
// program, a link with a scheme a hyperlink; an alert a note with its icon
// and title; HTML and comments nothing. A #fragment gives its heading the
// host's focus, which scrolls it into view; a relative link is written
// where link is bound and sends onLink.
func TestDocumentOnHost(t *testing.T) {
	x := newHarness(t)
	h := `"catalogId":"` + hottycat.ID + `"`
	text, _ := json.Marshal("# Guide\n\nSee [usage](#usage), [more](more.md) and [the site](https://x.example).\n\n" +
		"> [!CAUTION]\n> Hot.\n\n<!-- docs:figure -->\n![Logo](https://x.example/logo.png)\n\n## Usage\n\nDone.")
	var msgs []any
	must(t, json.Unmarshal([]byte(`[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"`+basic.ID+`","dataModel":{"link":""}}},
{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
 {"id":"root","component":"Column","children":["doc","ok"]},
 {"id":"doc","component":"HottyMarkdown",`+h+`,"text":`+string(text)+`,"link":{"@path":"/link"},
  "onLink":{"event":{"name":"open","context":{"href":{"@path":"/link"}}}}},
 {"id":"ok","component":"Button","child":"ok_label","action":{"event":{"name":"ok"}}},
 {"id":"ok_label","component":"Text","text":"OK"}]}}]`), &msgs))
	x.process(msgs...)
	r := x.rs["s"]
	s := x.h.Surface(r.name)
	link := func(i int) string { return domID(view.SubID("doc", "link", i)) }
	for _, want := range [][3]string{{"doc", "class", "k-text k-doc"},
		{headingID("doc", "guide"), "tabindex", "-1"}, {headingID("doc", "usage"), "tabindex", "-1"},
		{link(0), "class", "k-link"}, {link(0), "tabindex", "0"}, {link(0), "role", "link"},
		{link(0), "data-on", "click"}, {link(0), "data-keys", linkKeys}, {link(1), "tabindex", "0"}} {
		if got, _ := s.Attr(want[0], want[1]); got != want[2] {
			t.Errorf("%s's %s is %q, want %q", want[0], want[1], got, want[2])
		}
	}
	doc := s.HTML()
	for _, want := range []string{`<a href="https://x.example" target="_blank">the site</a>`, `<div class="k-alert k-alert-caution" role="note">`,
		`<p class="k-alert-title"><span class="k-icon" role="img" aria-label="report" aria-hidden="true"><svg`, `Caution</p>`,
		`<img src="https://x.example/logo.png" alt="Logo"`} {
		if !strings.Contains(doc, want) {
			t.Errorf("no %s in\n%s", want, doc)
		}
	}
	for _, not := range []string{"docs:figure", "[!CAUTION]", "<!--"} {
		if strings.Contains(doc, not) {
			t.Errorf("%s shows", not)
		}
	}

	must(t, x.h.Click(r.name, link(0)))
	x.pump()
	if r.C.St.Focus != "doc" || !r.C.St.Keyboard || s.Focused() != headingID("doc", "usage") || len(x.actions) != 0 {
		t.Fatalf("a click on #usage: the program's focus %q (%v), the host's %q, %d actions", r.C.St.Focus, r.C.St.Keyboard, s.Focused(), len(x.actions))
	}
	x.check(r)

	must(t, x.h.Click(r.name, link(1)))
	x.pump()
	if got := r.C.S.Data.Value("/link"); got != "more.md" || len(x.actions) != 1 || x.actions[0].Context["href"] != "more.md" {
		t.Fatalf("a click on more.md: link %v, actions %+v", got, x.actions)
	}
	if s.Focused() != link(1) {
		t.Errorf("the host's focus is on %q", s.Focused())
	}
	// Enter is the program's, which follows the link again.
	if x.h.Key("Enter") {
		t.Fatal("the host took Enter")
	}
	if _, ok, err := r.Key("Enter"); !ok || err != nil {
		t.Fatalf("Enter: %v %v", ok, err)
	}
	x.update(r)
	if len(x.actions) != 2 {
		t.Errorf("Enter on more.md sent %d actions", len(x.actions))
	}
	x.check(r)
}
