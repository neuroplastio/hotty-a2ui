package cells

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// page is the Markdown of the documents here.
const page = "# Title\n\nSee [usage](#usage) and [the site](https://x.example).\n\n> [!NOTE]\n> Read **this**.\n\n" +
	"- one\n- two\n\n| Name | Size |\n| :-- | --: |\n| a.go | 10 |\n\n```go\nx := 1\n```\n\n---\n\n## Usage\n\nDone, [more](more.md).\n\nThe end.\n\nReally."

// docs is a HottyMarkdown of page, alone or (scrolled) in a HottyScrollView
// of 5 rows, its link bound and its onLink sent; and the actions it sent.
func docs(t *testing.T, scrolled bool) (*Rendition, *view.Controller, *[]*a2ui.ActionMessage) {
	t.Helper()
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	actions := new([]*a2ui.ActionMessage)
	p.Send = func(o a2ui.Outbound) {
		if o.Action != nil {
			*actions = append(*actions, o.Action)
		}
	}
	h := `"catalogId":"` + hotty.ID + `"`
	text, _ := json.Marshal(page)
	root := `{"id":"root","component":"Column","children":["doc"]}`
	if scrolled {
		root = `{"id":"root","component":"HottyScrollView",` + h + `,"child":"doc","height":5}`
	}
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `","dataModel":{"link":""}}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[` + root + `,
	 {"id":"doc","component":"HottyMarkdown",` + h + `,"text":` + string(text) + `,"link":{"@path":"/link"},
	  "onLink":{"event":{"name":"open","context":{"href":{"@path":"/link"}}}}}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	c := view.NewController(p.Surface("s"))
	return New(c), c, actions
}

// rowText is a frame's row as text, its trailing spaces cut.
func rowText(f *Frame, row int) string {
	var b strings.Builder
	for _, c := range f.Cells[row] {
		if c.Width > 0 {
			b.WriteString(c.Text)
		}
	}
	return strings.TrimRight(b.String(), " ")
}

// A HottyMarkdown is its blocks a blank row apart (a tight list's items
// together); an alert a bar and a title row in its tone; a table as a
// HottyTable's header and rule, its columns aligned; code two columns in;
// a rule across. A link in place is underlined with no OSC 8, a link with
// a scheme a hyperlink.
func TestDocumentDraws(t *testing.T) {
	r, _, _ := docs(t, false)
	f := r.Draw(30)
	want := []string{
		"Title",
		"",
		"See usage and the site.",
		"",
		"▎ ⓘ Note",
		"▎ Read this.",
		"",
		"• one",
		"• two",
		"",
		" Name  Size",
		"────────────",
		" a.go    10",
		"",
		"  x := 1",
		"",
		"──────────────────────────────",
		"",
		"Usage",
		"",
		"Done, more.",
		"",
		"The end.",
		"",
		"Really.",
	}
	var got []string
	for row := range f.Rows {
		got = append(got, rowText(f, row))
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("drawn:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
	for _, tc := range []struct {
		col, row int
		role     Role
		attr     Attr
		link     string
	}{
		{0, 0, Fg, Bold | Underline, ""},            // the h1
		{4, 2, Fg, Underline, ""},                   // usage, in place
		{14, 2, Fg, Underline, "https://x.example"}, // the site
		{0, 4, Info, 0, ""},                         // the note's bar
		{2, 4, Info, Bold, ""},                      // its mark
		{0, 5, Info, 0, ""},
		{1, 10, Fg, Bold, ""}, // the header
		{0, 11, Border, 0, ""},
	} {
		c := f.Cells[tc.row][tc.col]
		if c.Role != tc.role || c.Attr != tc.attr || c.Link != tc.link {
			t.Errorf("(%d, %d) %q: %v %v %q, want %v %v %q", tc.col, tc.row, c.Text, c.Role, c.Attr, c.Link, tc.role, tc.attr, tc.link)
		}
	}
	if col, row, w, h, ok := r.Box("doc/link/0"); !ok || col != 4 || row != 2 || w != 5 || h != 1 {
		t.Errorf("usage's box: %d %d %d %d %v", col, row, w, h, ok)
	}
}

// A link in place takes a click and the keyboard, reversed in the accent
// while it has it; Enter follows it, Space does not. A #fragment scrolls
// the HottyScrollView to its heading, at the top, and the user scrolls on
// from there; another href is written where link is bound and sends
// onLink.
func TestDocumentLinks(t *testing.T) {
	r, c, actions := docs(t, true)
	r.Draw(30)
	c.Focus("doc/link/0")
	f := r.Draw(30)
	if u := f.Cells[2][5]; u.Text != "u" || u.Role != Accent || u.Attr&Reverse == 0 {
		t.Errorf("focused link: %+v", u)
	}
	if ok, _ := r.Key("Space"); ok {
		t.Error("Space followed the link")
	}
	if ok, err := r.Key("Enter"); !ok || err != nil {
		t.Fatalf("Enter: %v %v", ok, err)
	}
	f = r.Draw(30)
	if c.St.Focus != "doc" || !strings.HasPrefix(rowText(f, 0), " Usage ") {
		t.Fatalf("after #usage: focus %q, top row %q", c.St.Focus, rowText(f, 0))
	}
	if col, row, _, h, ok := r.Sight("doc"); !ok || col != 1 || row != 0 || h != 1 {
		t.Errorf("sight: %d %d %d %v", col, row, h, ok)
	}
	// The wheel scrolls on from the heading, which no longer holds it.
	r.Wheel(5, 2, 0, -1)
	if f = r.Draw(30); strings.HasPrefix(rowText(f, 0), " Usage ") {
		t.Errorf("the wheel did not scroll away from the heading")
	}
	col, row, _, _, ok := r.Box("doc/link/1")
	if !ok {
		r.Wheel(5, 2, 0, 9)
		r.Draw(30)
		if col, row, _, _, ok = r.Box("doc/link/1"); !ok {
			t.Fatal("no box for more")
		}
	}
	if err := r.Click(col, row); err != nil {
		t.Fatal(err)
	}
	if err := r.Release(); err != nil {
		t.Fatal(err)
	}
	if v := c.S.Data.Value("/link"); v != "more.md" || len(*actions) != 1 || (*actions)[0].Name != "open" || c.St.Focus != "doc/link/1" {
		t.Errorf("after more.md: link %v, actions %d, focus %q", v, len(*actions), c.St.Focus)
	}
}
