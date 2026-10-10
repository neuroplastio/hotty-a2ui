package view_test

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// docPage is a docs page as the docs module sends one: two HottyMarkdowns
// with a Button between them, the first's link bound and its onLink sent;
// and the actions it sent.
func docPage(t *testing.T) (*view.Controller, *[]*a2ui.ActionMessage) {
	t.Helper()
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	actions := new([]*a2ui.ActionMessage)
	p.Send = func(o a2ui.Outbound) {
		if o.Action != nil {
			*actions = append(*actions, o.Action)
		}
	}
	h := `"catalogId":"` + hotty.ID + `"`
	top, _ := json.Marshal("# Guide\n\nSee [install](install.md), [usage](#usage), [later](#usage-1), [nowhere](#nowhere) and [the site](https://neuroplast.io).\n\n## Usage\n\nRun it.")
	bottom, _ := json.Marshal("## Usage\n\nAgain, [back](#guide).")
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `","dataModel":{"link":""}}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"Column","children":["top","ok","bottom"]},
	 {"id":"top","component":"HottyMarkdown",` + h + `,"text":` + string(top) + `,"link":{"@path":"/link"},
	  "onLink":{"event":{"name":"open","context":{"href":{"@path":"/link"}}}}},
	 {"id":"ok","component":"Button","child":"ok_label","action":{"event":{"name":"ok"}}},
	 {"id":"ok_label","component":"Text","text":"OK"},
	 {"id":"bottom","component":"HottyMarkdown",` + h + `,"text":` + string(bottom) + `}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	return view.NewController(p.Surface("s")), actions
}

// GitHub's IDs (github-slugger): lower case, what is not a letter, a
// mark, a number, "_", a space or "-" dropped, spaces "-", and an ID the
// page has already numbered.
func TestSlugger(t *testing.T) {
	var s view.Slugger
	for _, tc := range []struct{ text, id string }{
		{"Install hotty", "install-hotty"},
		{"Usage", "usage"},
		{"Usage", "usage-1"},
		{"Usage-1", "usage-1-1"},
		{"Usage", "usage-2"},
		{"C++ & Rust: 100% done!", "c--rust-100-done"},
		{"snake_case and kebab-case", "snake_case-and-kebab-case"},
		{"Привет, мир", "привет-мир"},
		{"🚀 Launch", "-launch"},
		{"Café crème", "café-crème"},
		{"API v2.0 (beta)", "api-v20-beta"},
		{"", ""},
		{"!!", "-1"},
	} {
		if got := s.Slug(tc.text); got != tc.id {
			t.Errorf("%q: %q, want %q", tc.text, got, tc.id)
		}
	}
}

// A link goes in place when its href names no scheme.
func TestInPlace(t *testing.T) {
	for href, want := range map[string]bool{
		"install.md": true, "#usage": true, "../api/errors.md#codes": true, "./a:b": true, "a/b:c": true, "?tab=2": true,
		"https://neuroplast.io": false, "HTTP://X": false, "mailto:ada@example.com": false, "x-man-page://ls": false,
		"//cdn.example.com/a.js": false, "": false,
	} {
		if got := view.InPlace(href); got != want {
			t.Errorf("%q: %v, want %v", href, got, want)
		}
	}
}

// A docs page reads into blocks as GitHub shows it: frames for lists,
// quotes and alerts, a blank row between blocks but a tight list's items,
// HTML and comments nothing, an image its alt text, and a table its cells.
func TestReadDoc(t *testing.T) {
	d := view.ReadDoc(strings.Join([]string{
		"# Title",
		"",
		"> [!WARNING]",
		"> Mind **this**.",
		"> And that.",
		"",
		"> plain",
		">",
		"> two",
		"",
		"1. one",
		"2. two",
		"   - nested [x](#title)",
		"",
		"<!-- docs:figure caption=\"Logo\" -->",
		"![The logo](logo.png)",
		"",
		"| A | B |",
		"| - | -: |",
		"| a | [b](b.md) |",
		"",
		"Text <!-- inline --> \\_escaped\\_ &amp; done.",
	}, "\n"))
	type row struct {
		kind  view.BlockKind
		text  string
		in    string
		gap   bool
		gapIn int
	}
	var got []row
	for _, b := range d.Blocks {
		var s strings.Builder
		for _, r := range b.Runs {
			s.WriteString(r.Text)
		}
		if b.Kind == view.AlertTitle {
			s.WriteString(b.Alert)
		}
		var in []string
		for _, f := range b.In {
			k := f.Kind
			if f.Kind == view.FrameItem && f.First {
				k += "(" + f.Marker + ")"
			}
			in = append(in, k)
		}
		got = append(got, row{b.Kind, s.String(), strings.Join(in, "/"), b.Gap, b.GapIn})
	}
	want := []row{
		{view.Heading, "Title", "", false, 0},
		{view.AlertTitle, "warning", "alert", true, 0},
		{view.Paragraph, "Mind this. And that.", "alert", false, 0},
		{view.Paragraph, "plain", "quote", true, 0},
		{view.Paragraph, "two", "quote", true, 1},
		{view.Paragraph, "one", "item(1. )", true, 0},
		{view.Paragraph, "two", "item(2. )", false, 0},
		{view.Paragraph, "nested x", "item/item(• )", false, 0},
		{view.Paragraph, "[image: The logo]", "", true, 0},
		{view.TableBlock, "", "", true, 0},
		{view.Paragraph, "Text  _escaped_ & done.", "", true, 0},
	}
	if !slices.Equal(got, want) {
		t.Errorf("blocks:\n%v\nwant\n%v", got, want)
	}
	tb := d.Blocks[9]
	if len(tb.Rows) != 2 || len(tb.Rows[1]) != 2 || tb.Rows[1][1][0].Text != "b" || tb.Rows[1][1][0].Link != 2 ||
		!slices.Equal(tb.Align, []string{"", "end"}) {
		t.Errorf("table: %+v", tb)
	}
	if len(d.Links) != 2 || d.Links[0] != (view.DocLink{Href: "#title", Label: "x"}) || d.Links[1].Href != "b.md" {
		t.Errorf("links: %+v", d.Links)
	}
	if len(d.Headings) != 1 || d.Headings[0].Text != "Title" {
		t.Errorf("headings: %+v", d.Headings)
	}
}

// The surface is the page: its documents' headings are numbered across
// it, in tree order. Links in place are Tab stops, those with a scheme
// not.
func TestDocBuilds(t *testing.T) {
	c, _ := docPage(t)
	top, bottom := c.V.Find("top"), c.V.Find("bottom")
	if top.Kind != view.Document || !slices.Equal(top.Anchors, []string{"guide", "usage"}) || !slices.Equal(bottom.Anchors, []string{"usage-1"}) {
		t.Fatalf("anchors %q, %q", top.Anchors, bottom.Anchors)
	}
	want := []string{"top/link/0", "top/link/1", "top/link/2", "top/link/3", "ok", "bottom/link/0"}
	if ids := c.V.Focusables(); !slices.Equal(ids, want) {
		t.Errorf("focusables %q, want %q", ids, want)
	}
	if l := c.V.Find("top/link/2"); l.Kind != view.Link || l.Label != "later" || l.URL != "#usage-1" {
		t.Errorf("link 2: %+v", l)
	}
}

// A #fragment that names a heading takes the keyboard to its document
// (the jump), sending nothing; any other href in place is written where
// link is bound before onLink runs; a fragment no heading has is the
// agent's too.
func TestFollowLink(t *testing.T) {
	c, actions := docPage(t)
	c.Focus("top/link/2")
	if err := c.Activate("top/link/2"); err != nil {
		t.Fatal(err)
	}
	if c.St.Focus != "bottom" || !c.St.Keyboard || c.St.Jump.Heading != "usage-1" || c.St.Jump.Seq != 1 ||
		c.V.Find("bottom").Target() != 0 || len(*actions) != 0 {
		t.Fatalf("after #usage-1: focus %q, jump %+v, %d actions", c.St.Focus, c.St.Jump, len(*actions))
	}
	if err := c.Activate("bottom/link/0"); err != nil {
		t.Fatal(err)
	}
	if c.St.Focus != "top" || c.V.Find("top").Target() != 0 || c.St.Jump.Seq != 2 {
		t.Errorf("after #guide: focus %q, jump %+v", c.St.Focus, c.St.Jump)
	}
	for i, href := range []string{"install.md", "#nowhere"} {
		if err := c.Activate("top/link/" + []string{"0", "3"}[i]); err != nil {
			t.Fatal(err)
		}
		if got := c.S.Data.Value("/link"); got != href {
			t.Errorf("link %v, want %q", got, href)
		}
		if n := len(*actions); n != i+1 || (*actions)[n-1].Name != "open" || (*actions)[n-1].Context["href"] != href {
			t.Errorf("actions after %s: %+v", href, *actions)
		}
	}
}

// Tab from a heading a link in place went to goes on from there: to the
// next link after it in its document, else to what follows the document;
// Shift+Tab to the one before.
func TestFocusAfterJump(t *testing.T) {
	for _, tc := range []struct {
		link string
		back bool
		want string
	}{
		{"top/link/1", false, "ok"},
		{"top/link/1", true, "top/link/3"},
		{"top/link/2", false, "bottom/link/0"},
		{"top/link/2", true, "ok"},
		{"bottom/link/0", false, "top/link/0"},
		{"bottom/link/0", true, ""},
	} {
		c, _ := docPage(t)
		c.Focus(tc.link)
		if err := c.Activate(tc.link); err != nil {
			t.Fatal(err)
		}
		c.FocusNext(tc.back)
		got := c.St.Focus
		if !c.St.Keyboard {
			got = ""
		}
		if got != tc.want {
			t.Errorf("%s then Tab (back %v): %q, want %q", tc.link, tc.back, got, tc.want)
		}
	}
}

// DocHTML gives the headings and the links in place the attributes and
// tags its caller makes, leaves HTML out, and makes a link with a scheme a
// hyperlink.
func TestDocHTML(t *testing.T) {
	h := view.DocHTML("# A\n\n<!-- docs:x -->\nSee [b](b.md), [c](https://c.example), <span>raw</span>.\n\n> [!NOTE]\n> Hi.", view.DocMarkup{
		Heading: func(i int) []string { return []string{"id", "h" + string(rune('0'+i))} },
		Link:    func(i int) (string, string) { return "<span id=\"l" + string(rune('0'+i)) + "\">", "</span>" },
		Alert:   func(kind string) (string, string) { return "<div class=\"" + kind + "\">", "</div>" },
	})
	for _, want := range []string{`<h1 id="h0">A</h1>`, `<span id="l0">b</span>`, `<a href="https://c.example" target="_blank">c</a>`,
		`<div class="note"><p>Hi.</p>`} {
		if !strings.Contains(h, want) {
			t.Errorf("no %s in\n%s", want, h)
		}
	}
	for _, not := range []string{"docs:x", "<!--", "<span>raw", "[!NOTE]"} {
		if strings.Contains(h, not) {
			t.Errorf("%s in\n%s", not, h)
		}
	}
}
