package view_test

import (
	"slices"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// A HottyDiff is its patch parsed, or its old and new compared; unified
// unless its view is split, numbered and wrapped unless told not to; its
// hunks' ids are their file's name and new line, the line alone with no
// name; it takes the keyboard only when it has a hunk.
func TestDiff(t *testing.T) {
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `"}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"Column","children":["patch","texts","same"]},
	 {"id":"patch","component":"HottyDiff","catalogId":"` + hotty.ID + `","view":"split","lineNumbers":false,
	  "patch":"--- a/x.go\n+++ b/x.go\n@@ -1 +1 @@\n-a\n+b\n@@ -9 +9 @@\n-c\n+d\n--- a/y.go\n+++ b/y.go\n@@ -3 +3 @@\n-e\n+f\n"},
	 {"id":"texts","component":"HottyDiff","catalogId":"` + hotty.ID + `","old":"1\n2\n3\n4\n5\n","new":"1\n2\n3\n4\nfive\n","context":1,"wrap":false},
	 {"id":"same","component":"HottyDiff","catalogId":"` + hotty.ID + `","old":"a\n","new":"a\n"}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	c := view.NewController(p.Surface("s"))
	e := c.V.Find("patch")
	if e.Kind != view.DiffView || e.Variant != "split" || e.Numbers || !e.Wrap || len(e.Diff.Files) != 2 {
		t.Fatalf("patch: %+v", e)
	}
	if want := []string{"x.go:1", "x.go:9", "y.go:3"}; !slices.Equal(e.RowIDs, want) {
		t.Errorf("hunks %q, want %q", e.RowIDs, want)
	}
	tx := c.V.Find("texts")
	if tx.Variant != "unified" || !tx.Numbers || tx.Wrap || !slices.Equal(tx.RowIDs, []string{"4"}) {
		t.Errorf("texts: %+v", tx)
	}
	if h := tx.Diff.Files[0].Hunks[0]; h.Header() != "@@ -4,2 +4,2 @@" {
		t.Errorf("context 1: %s", h.Header())
	}
	if !e.Focusable() || c.V.Find("same").Focusable() {
		t.Error("a diff with hunks takes the keyboard, and one without does not")
	}
}

// The arrows (k and j) select the hunk before or after, Home and End (g
// and G) the first and the last, any of them the first from none; Enter
// acts on the selected hunk, and the agent reads its id.
func TestDiffKey(t *testing.T) {
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	var acts []string
	p.Send = func(o a2ui.Outbound) {
		if o.Action != nil {
			acts = append(acts, a2ui.ToString(o.Action.Context["hunk"]))
		}
	}
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `"}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"HottyDiff","catalogId":"` + hotty.ID + `","selected":{"@path":"/h"},
	  "patch":"@@ -1 +1 @@\n-a\n+b\n@@ -9 +9 @@\n-c\n+d\n@@ -20 +20 @@\n-e\n+f\n",
	  "onActivate":{"event":{"name":"stage","context":{"hunk":{"@path":"/h"}}}}}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	c := view.NewController(p.Surface("s"))
	for _, step := range []struct{ key, want string }{
		{"k", "1"}, {"ArrowDown", "9"}, {"j", "20"}, {"j", "20"}, {"Home", "1"}, {"G", "20"}, {"ArrowUp", "9"}, {"g", "1"}, {"End", "20"},
	} {
		if ok, err := c.DiffKey("root", step.key); !ok || err != nil {
			t.Fatalf("%s: %v %v", step.key, ok, err)
		}
		if got := c.S.Data.Value("/h"); got != step.want {
			t.Errorf("after %s: %v, want %s", step.key, got, step.want)
		}
	}
	if ok, _ := c.DiffKey("root", "PageDown"); ok {
		t.Error("PageDown is the diff's")
	}
	if ok, err := c.DiffKey("root", "Enter"); !ok || err != nil || !slices.Equal(acts, []string{"20"}) {
		t.Errorf("Enter: %v %v, acted on %q", ok, err, acts)
	}
}
