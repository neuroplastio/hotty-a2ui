package cells

import (
	"strings"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// repo is a surface with a HottyTree of a repository's files, src open
// and README.md selected, and the actions it sent.
func repo(t *testing.T, extra string) (*Rendition, *view.Controller, *[]string) {
	t.Helper()
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	actions := new([]string)
	p.Send = func(o a2ui.Outbound) {
		if o.Action != nil {
			*actions = append(*actions, o.Action.Name+":"+a2ui.ToString(o.Action.Context["file"]))
		}
	}
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `","dataModel":{"sel":"README.md","open":["src","src/view"]}}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"HottyTree","catalogId":"` + hotty.ID + `","selected":{"@path":"/sel"},"expanded":{"@path":"/open"}` + extra + `,
	  "onActivate":{"event":{"name":"open","context":{"file":{"@path":"/sel"}}}},
	  "items":[
	   {"label":"src","icon":"folder","value":"src","children":[
	    {"label":"main.go","value":"src/main.go"},
	    {"label":"view","icon":"folder","value":"src/view","children":[
	     {"label":"tree.go","value":"src/view/tree.go"},
	     {"label":"list.go","value":"src/view/list.go"}]},
	    {"label":"util.go","value":"src/util.go"}]},
	   {"label":"docs","icon":"folder","value":"docs","children":[
	    {"label":"profile.md","value":"docs/profile.md"}]},
	   {"label":"README.md","value":"README.md"}]}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	c := view.NewController(p.Surface("s"))
	return New(c), c, actions
}

// A HottyTree is a row a node, with lipgloss's guides: a branch's fold
// before its label, and its count while closed; the roots lined up, every
// row two columns in, the selected node's columns a "│" bar.
func TestTreeDraws(t *testing.T) {
	r, c, _ := repo(t, "")
	want := "  ▼ src\n" +
		"  ├── main.go\n" +
		"  ├── ▼ view\n" +
		"  │   ├── tree.go\n" +
		"  │   └── list.go\n" +
		"  └── util.go\n" +
		"  ▶ docs 1\n" +
		"│   README.md"
	f := r.Draw(30)
	if got := f.Plain(); got != want {
		t.Fatalf("got\n%s\nwant\n%s", got, want)
	}
	if guide, fold, count := f.Cells[1][2], f.Cells[0][2], f.Cells[6][9]; guide.Role != Border || fold.Role != Muted || count.Role != Muted {
		t.Errorf("guide %v, fold %v, count %v", guide.Role, fold.Role, count.Role)
	}
	if bar, label := f.Cells[7][0], f.Cells[7][4]; bar.Role != Muted || label.Role != Fg {
		t.Errorf("unfocused: bar %v, label %v", bar.Role, label.Role)
	}
	// The selected node's row is what a pane keeps in sight.
	if _, row, _, h, ok := r.Sight("root"); !ok || row != 7 || h != 1 {
		t.Errorf("sight: row %d, %d high (%v), want the selected row", row, h, ok)
	}
	c.Focus("root")
	f = r.Draw(30)
	if bar, label := f.Cells[7][0], f.Cells[7][4]; bar.Role != Accent || label.Role != Accent || label.Attr&Bold == 0 {
		t.Errorf("focused: bar %v, label %v", bar.Role, label.Role)
	}
	if want := "  ▼ src\n  ├── main.…\n  ├── ▼ view\n  │   ├── t…"; !strings.HasPrefix(r.Draw(12).Plain(), want) {
		t.Errorf("12 wide:\n%s", r.Draw(12).Plain())
	}
}

// The keys move among the rows, open and close branches and act on a leaf;
// a click on a branch opens or closes it, and on a leaf selects it, then
// acts on it.
func TestTreeKeysAndClicks(t *testing.T) {
	r, c, actions := repo(t, "")
	c.Focus("root")
	keys(t, r, "k", "h") // docs, which is closed: up to its parent, none
	if got := c.S.Data.Value("/sel"); got != "docs" {
		t.Fatalf("selected %v, want docs", got)
	}
	keys(t, r, "l", "l")
	if got := c.S.Data.Value("/sel"); got != "docs/profile.md" {
		t.Fatalf("selected %v, want docs/profile.md", got)
	}
	keys(t, r, "Enter")
	if len(*actions) != 1 || (*actions)[0] != "open:docs/profile.md" {
		t.Fatalf("actions %v", *actions)
	}
	r.Draw(30)
	if err := r.Click(5, 2); err != nil { // view: closes it
		t.Fatal(err)
	}
	if got := r.Draw(30).Plain(); got[:len("  ▼ src\n  ├── main.go\n│ ├── ▶ view 2\n  └── util.go")] != "  ▼ src\n  ├── main.go\n│ ├── ▶ view 2\n  └── util.go" {
		t.Errorf("view closed:\n%s", got)
	}
	if err := r.Click(5, 3); err != nil || c.S.Data.Value("/sel") != "src/util.go" || len(*actions) != 1 {
		t.Fatalf("a click on util.go: %v %v %v", err, c.S.Data.Value("/sel"), *actions)
	}
	r.Draw(30)
	if err := r.Click(5, 3); err != nil || len(*actions) != 2 || (*actions)[1] != "open:src/util.go" {
		t.Fatalf("a second click: %v %v", err, *actions)
	}
}

// With a height the tree shows that many rows, scrolled as little as
// keeps the selected node in sight; with a filter, the nodes it leaves,
// their matched characters underlined, and the empty text when it leaves
// none.
func TestTreeHeightAndFilter(t *testing.T) {
	r, c, _ := repo(t, `,"height":3`)
	if got, want := r.Draw(30).Plain(), "  └── util.go\n  ▶ docs 1\n│   README.md"; got != want {
		t.Errorf("scrolled to the end:\ngot\n%s\nwant\n%s", got, want)
	}
	c.Focus("root")
	keys(t, r, "g")
	if got, want := r.Draw(30).Plain(), "│ ▼ src\n  ├── main.go\n  ├── ▼ view"; got != want {
		t.Errorf("at the start:\ngot\n%s\nwant\n%s", got, want)
	}
	r, c, _ = repo(t, `,"filter":{"@path":"/q"}`)
	if err := c.S.Write("/q", "lis"); err != nil {
		t.Fatal(err)
	}
	c.Rebuild()
	f := r.Draw(30)
	if got, want := f.Plain(), "  ▼ src\n  └── ▼ view\n      └── list.go"; got != want {
		t.Errorf("filtered:\ngot\n%s\nwant\n%s", got, want)
	}
	for col, under := range map[int]bool{10: true, 11: true, 12: true, 13: false} {
		if got := f.Cells[2][col].Attr&Underline != 0; got != under {
			t.Errorf("list.go, column %d underlined: %v", col, got)
		}
	}
	if err := c.S.Write("/q", "zzz"); err != nil {
		t.Fatal(err)
	}
	c.Rebuild()
	if got := r.Draw(30).Plain(); got != "  Nothing here." {
		t.Errorf("nothing left: %q", got)
	}
}

// The wheel over a tree with a height moves its rows, three a notch, and
// leaves the selection; a key that moves the selection brings it back
// into view.
func TestTreeWheel(t *testing.T) {
	r, c, _ := repo(t, `,"height":3`)
	r.Draw(30)
	for _, step := range []struct {
		dy    int
		moved bool
		want  string
	}{
		{-1, true, "  ├── ▼ view\n  │   ├── tree.go\n  │   └── list.go"},
		{-1, true, "  ▼ src\n  ├── main.go\n  ├── ▼ view"},
		{-1, false, "  ▼ src\n  ├── main.go\n  ├── ▼ view"},
		{1, true, "  │   ├── tree.go\n  │   └── list.go\n  └── util.go"},
	} {
		if moved := r.Wheel(4, 1, 0, step.dy); moved != step.moved {
			t.Errorf("the wheel %d: moved %v", step.dy, moved)
		}
		if got := r.Draw(30).Plain(); got != step.want || c.S.Data.Value("/sel") != "README.md" {
			t.Errorf("after the wheel %d, selected %v:\n%s\nwant\n%s", step.dy, c.S.Data.Value("/sel"), got, step.want)
		}
	}
	c.Focus("root")
	keys(t, r, "k")
	if got, want := r.Draw(30).Plain(), "  │   └── list.go\n  └── util.go\n│ ▶ docs 1"; got != want {
		t.Errorf("k: got\n%s\nwant\n%s", got, want)
	}
}
