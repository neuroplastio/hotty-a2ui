package view_test

import (
	"slices"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// srcTree is a repository's files as HottyTree nodes, each a path.
const srcTree = `[
 {"label":"src","icon":"folder","value":"src","children":[
  {"label":"main.go","value":"src/main.go"},
  {"label":"view","icon":"folder","value":"src/view","children":[
   {"label":"tree.go","value":"src/view/tree.go"},
   {"label":"list.go","value":"src/view/list.go"}]},
  {"label":"util.go","value":"src/util.go"}]},
 {"label":"docs","icon":"folder","value":"docs","children":[
  {"label":"profile.md","value":"docs/profile.md","href":"docs/profile.md#627-hottymarkdown"}]},
 {"label":"README.md","value":"README.md"}]`

// files is a surface with srcTree, selected and expanded bound, src
// open, and a tree beside it whose nodes have no values and whose folding
// is the renderer's; and the actions it sent.
func files(t *testing.T, filter string) (*view.Controller, *[]*a2ui.ActionMessage) {
	t.Helper()
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	actions := new([]*a2ui.ActionMessage)
	p.Send = func(o a2ui.Outbound) {
		if o.Action != nil {
			*actions = append(*actions, o.Action)
		}
	}
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `","dataModel":{"sel":"README.md","open":["src"],"q":"` + filter + `","files":` + srcTree + `}}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"Column","children":["t","plain"]},
	 {"id":"t","component":"HottyTree","catalogId":"` + hotty.ID + `","items":{"@path":"/files"},
	  "selected":{"@path":"/sel"},"expanded":{"@path":"/open"},"filter":{"@path":"/q"},
	  "onActivate":{"event":{"name":"open","context":{"file":{"@path":"/sel"}}}}},
	 {"id":"plain","component":"HottyTree","catalogId":"` + hotty.ID + `","expanded":["1"],
	  "items":[{"label":"One","children":[{"label":"A"}]},{"label":"Two","children":[{"label":"B"},{"label":"C"}]}]}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	return view.NewController(p.Surface("s")), actions
}

// shownLabels are the labels of the nodes a tree shows, in order.
func shownLabels(e *view.Element) []string {
	var out []string
	for _, i := range e.Shown {
		out = append(out, e.Nodes[i].Label)
	}
	return out
}

// A HottyTree's nodes are in pre-order, an id each: its value, else its
// place. It shows the roots and what is in the open branches.
func TestTreeBuilds(t *testing.T) {
	c, _ := files(t, "")
	e := c.V.Find("t")
	if e == nil || e.Kind != view.Tree || !e.Focusable() {
		t.Fatalf("t: %+v, want a focusable HottyTree", e)
	}
	if len(e.Nodes) != 9 || e.RowIDs[3] != "src/view/tree.go" || e.Value != "README.md" || !slices.Equal(e.Expanded, []string{"src"}) {
		t.Fatalf("nodes %v, ids %v, selected %q, expanded %v", e.Nodes, e.RowIDs, e.Value, e.Expanded)
	}
	if n := e.Nodes[2]; n.Label != "view" || n.Icon != "folder" || n.Level != 1 || n.Parent != 0 || n.Kids != 2 || n.Last || n.Open {
		t.Errorf("view: %+v", n)
	}
	if n := e.Nodes[5]; n.Label != "util.go" || !n.Last || n.Branch() {
		t.Errorf("util.go: %+v", n)
	}
	// A node may lead somewhere, a docs site's nav (profile §6.14).
	if e.Nodes[7].Href != "docs/profile.md#627-hottymarkdown" || e.Nodes[8].Href != "" {
		t.Errorf("hrefs %q, %q", e.Nodes[7].Href, e.Nodes[8].Href)
	}
	if got, want := shownLabels(e), []string{"src", "main.go", "view", "util.go", "docs", "README.md"}; !slices.Equal(got, want) {
		t.Errorf("shown %v, want %v", got, want)
	}
	if through, last := e.Guides(5); len(through) != 0 || !last { // util.go, src's last
		t.Errorf("util.go's guides: %v %v", through, last)
	}
	plain := c.V.Find("plain")
	if plain.RowIDs[2] != "1" || plain.RowIDs[4] != "1.1" || !slices.Equal(shownLabels(plain), []string{"One", "Two", "B", "C"}) {
		t.Errorf("plain: ids %v, shown %v", plain.RowIDs, shownLabels(plain))
	}
}

// The branches the selection is in show open, whatever expanded says.
func TestTreeRevealsTheSelection(t *testing.T) {
	c, _ := files(t, "")
	if err := c.S.Write("/sel", "src/view/list.go"); err != nil {
		t.Fatal(err)
	}
	if err := c.S.Write("/open", []any{}); err != nil {
		t.Fatal(err)
	}
	c.Rebuild()
	e := c.V.Find("t")
	if got, want := shownLabels(e), []string{"src", "main.go", "view", "tree.go", "list.go", "util.go", "docs", "README.md"}; !slices.Equal(got, want) {
		t.Errorf("shown %v, want %v", got, want)
	}
	through, last := e.Guides(4) // list.go: in view, which util.go follows
	if !slices.Equal(through, []bool{true}) || !last {
		t.Errorf("list.go's guides: %v %v", through, last)
	}
}

// The keys move the selection among the nodes that show, open and close
// branches, and act on a leaf; what is open is written where expanded is
// bound.
func TestTreeKeys(t *testing.T) {
	c, actions := files(t, "")
	c.Focus("t")
	key := func(k string) {
		t.Helper()
		if ok, err := c.TreeKey("t", k); !ok || err != nil {
			t.Fatalf("%s: %v %v", k, ok, err)
		}
	}
	data := func(path string) any { return c.S.Data.Value(path) }
	for _, step := range []struct {
		key, sel string
		open     []any
	}{
		{"g", "src", []any{"src"}},
		{"j", "src/main.go", []any{"src"}},
		{"j", "src/view", []any{"src"}},
		{"l", "src/view", []any{"src", "src/view"}},         // opens it
		{"l", "src/view/tree.go", []any{"src", "src/view"}}, // its first child
		{"h", "src/view", []any{"src", "src/view"}},         // to the parent
		{"h", "src/view", []any{"src"}},                     // closes it
		{"G", "README.md", []any{"src"}},
		{"k", "docs", []any{"src"}},
		{"Enter", "docs", []any{"src", "docs"}}, // a branch opens
		{"j", "docs/profile.md", []any{"src", "docs"}},
		{"ArrowUp", "docs", []any{"src", "docs"}},
		{"Space", "docs", []any{"src"}}, // and closes
		{"Home", "src", []any{"src"}},
		{"ArrowLeft", "src", []any{}},
	} {
		key(step.key)
		if got := data("/sel"); got != step.sel {
			t.Errorf("after %s: selected %v, want %s", step.key, got, step.sel)
		}
		if got, _ := data("/open").([]any); !slices.Equal(got, step.open) {
			t.Errorf("after %s: open %v, want %v", step.key, got, step.open)
		}
	}
	key("End")
	key("Enter")
	if len(*actions) != 1 || (*actions)[0].Name != "open" || (*actions)[0].Context["file"] != "README.md" {
		t.Errorf("actions %+v, want open README.md", *actions)
	}
	if ok, _ := c.TreeKey("t", "x"); ok {
		t.Error("the tree took x")
	}
}

// Closing the branch the selection is in selects the branch.
func TestTreeClosesAroundTheSelection(t *testing.T) {
	c, _ := files(t, "")
	if err := c.S.Write("/sel", "src/view/tree.go"); err != nil {
		t.Fatal(err)
	}
	c.Rebuild()
	if err := c.ToggleNode("t", 0); err != nil {
		t.Fatal(err)
	}
	e := c.V.Find("t")
	if e.Value != "src" || !slices.Equal(shownLabels(e), []string{"src", "docs", "README.md"}) {
		t.Errorf("selected %q, shown %v", e.Value, shownLabels(e))
	}
}

// A filter shows the nodes it matches, the branches that lead to them and
// what they hold, every branch open.
func TestTreeFilter(t *testing.T) {
	c, _ := files(t, "tree")
	e := c.V.Find("t")
	if got, want := shownLabels(e), []string{"src", "view", "tree.go"}; !slices.Equal(got, want) {
		t.Errorf("tree: shown %v, want %v", got, want)
	}
	if !e.Filtering() || !slices.Equal(e.Matched[2], []int{0, 1, 2, 3}) || e.Matched[0] != nil {
		t.Errorf("matched %v", e.Matched)
	}
	if err := c.S.Write("/q", "docs"); err != nil {
		t.Fatal(err)
	}
	c.Rebuild()
	if got, want := shownLabels(c.V.Find("t")), []string{"docs", "profile.md"}; !slices.Equal(got, want) {
		t.Errorf("docs: shown %v, want %v", got, want)
	}
}

// A tree whose expanded is not bound keeps its folding itself.
func TestTreeLocalFolding(t *testing.T) {
	c, _ := files(t, "")
	if err := c.ToggleNode("plain", 0); err != nil {
		t.Fatal(err)
	}
	if got := shownLabels(c.V.Find("plain")); !slices.Equal(got, []string{"One", "A", "Two", "B", "C"}) {
		t.Errorf("shown %v", got)
	}
	if !slices.Equal(c.St.Open["plain"], []string{"0", "1"}) {
		t.Errorf("open %v", c.St.Open["plain"])
	}
}

// FoldAll opens every branch, or closes them all and selects the root the
// selection was in.
func TestTreeFoldAll(t *testing.T) {
	c, _ := files(t, "")
	if err := c.S.Write("/sel", "src/view/tree.go"); err != nil {
		t.Fatal(err)
	}
	c.Rebuild()
	if err := c.FoldAll("t", true); err != nil {
		t.Fatal(err)
	}
	if got, _ := c.S.Data.Value("/open").([]any); !slices.Equal(got, []any{"src", "src/view", "docs"}) {
		t.Errorf("expanded all: %v", got)
	}
	if err := c.FoldAll("t", false); err != nil {
		t.Fatal(err)
	}
	if got, _ := c.S.Data.Value("/open").([]any); len(got) != 0 || c.S.Data.Value("/sel") != "src" {
		t.Errorf("collapsed all: open %v, selected %v", got, c.S.Data.Value("/sel"))
	}
	if got := shownLabels(c.V.Find("t")); !slices.Equal(got, []string{"src", "docs", "README.md"}) {
		t.Errorf("collapsed all, shown %v", got)
	}
	if err := c.FoldAll("plain", true); err != nil || !slices.Equal(c.St.Open["plain"], []string{"0", "1"}) {
		t.Errorf("plain, expanded all: %v %v", err, c.St.Open["plain"])
	}
}
