package view_test

import (
	"slices"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// suggesting is a surface of a TextField with suggestions bound to /opts
// and an onInput that sends the query, and a plain one beside it; and the
// contexts of the actions the agent had.
func suggesting(t *testing.T, value string, opts ...string) (*view.Controller, *[]map[string]any) {
	t.Helper()
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	sent := new([]map[string]any)
	p.Send = func(o a2ui.Outbound) {
		if o.Action != nil {
			*sent = append(*sent, o.Action.Context)
		}
	}
	if err := p.ProcessJSON([]byte(`[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `","dataModel":{"q":"","opts":[]}}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"Column","children":["f","plain","long"]},
	 {"id":"f","component":"TextField","label":"F","value":{"@path":"/q"},
	  "metadata":{"extensions":{"io_neuroplast_hotty":{"suggestions":{"options":{"@path":"/opts"},
	   "onInput":{"event":{"name":"suggest","context":{"query":{"@path":"/q"}}}}}}}}},
	 {"id":"plain","component":"TextField","label":"Plain","value":"x"},
	 {"id":"long","component":"TextField","label":"Long","variant":"longText","value":"x",
	  "metadata":{"extensions":{"io_neuroplast_hotty":{"suggestions":{"options":["xy"]}}}}}]}}]`)); err != nil {
		t.Fatal(err)
	}
	c := view.NewController(p.Surface("s"))
	l := make([]any, len(opts))
	for i, o := range opts {
		l[i] = o
	}
	if err := c.S.Write("/opts", l); err != nil {
		t.Fatal(err)
	}
	if err := c.S.Write("/q", value); err != nil {
		t.Fatal(err)
	}
	c.Rebuild()
	c.Focus("f")
	return c, sent
}

// shown are the suggestions a field shows, and the highlighted one.
func shown(c *view.Controller) (l []string, hi string) {
	e := c.V.Find("f")
	for i, j := range e.Shown {
		l = append(l, e.Suggestions[j])
		if i == e.Selected {
			hi = e.Suggestions[j]
		}
	}
	return l, hi
}

func key(t *testing.T, c *view.Controller, k string, atEnd bool) bool {
	t.Helper()
	ok, err := c.SuggestKey("f", k, atEnd)
	if err != nil {
		t.Fatal(err)
	}
	return ok
}

// The suggestions a value leaves start with it, case aside, in the
// agent's order, but for the value itself; none while it is empty. Only a
// one-line field that shows its value takes them.
func TestSuggestMatches(t *testing.T) {
	c, _ := suggesting(t, "be", "Berlin", "Bern", "Amsterdam", "bergen", "Berlin", "be", "")
	if l, hi := shown(c); !slices.Equal(l, []string{"Berlin", "Bern", "bergen"}) || hi != "" {
		t.Errorf("be leaves %q, %q highlighted", l, hi)
	}
	e := c.V.Find("f")
	if e.Suggestion() != "Berlin" || e.Ghost() != "rlin" {
		t.Errorf("suggestion %q, ghost %q", e.Suggestion(), e.Ghost())
	}
	if typed, rest := e.SuggestionParts(2); typed != "be" || rest != "rgen" {
		t.Errorf("bergen's parts %q %q", typed, rest)
	}
	c.S.Write("/q", "")
	c.Rebuild()
	if l, _ := shown(c); l != nil || c.V.Find("f").Ghost() != "" {
		t.Errorf("empty, it shows %q", l)
	}
	c.S.Write("/q", "zz")
	c.Rebuild()
	if l, _ := shown(c); l != nil {
		t.Errorf("zz shows %q", l)
	}
	if c.V.Find("plain").Suggestions != nil || c.V.Find("long").Suggestions != nil {
		t.Error("a field without the extension, or a longText, has suggestions")
	}
	// Case folds rune by rune, beyond ASCII.
	c.S.Write("/opts", []any{"Ärger", "ÆØÅ"})
	c.S.Write("/q", "äR")
	c.Rebuild()
	if e := c.V.Find("f"); e.Ghost() != "ger" {
		t.Errorf("äR's ghost %q", e.Ghost())
	}
}

// Tab takes the suggestion, as written, and → at the value's end; the
// list then stays shut until the next edit, so that Tab moves on. Tab with
// nothing to take, and → before the end, are not the suggestions'.
func TestSuggestTake(t *testing.T) {
	c, sent := suggesting(t, "ber", "Berlin", "Bern")
	if key(t, c, "ArrowRight", false) {
		t.Error("→ before the end took a suggestion")
	}
	if !key(t, c, "Tab", true) || c.S.Data.Value("/q") != "Berlin" {
		t.Fatalf("Tab: /q is %v", c.S.Data.Value("/q"))
	}
	if l, _ := shown(c); l != nil || key(t, c, "Tab", true) {
		t.Errorf("after Tab it shows %q, and takes Tab", l)
	}
	if len(*sent) != 1 || (*sent)[0]["query"] != "Berlin" {
		t.Errorf("onInput sent %v", *sent)
	}
	// An edit opens it again.
	if err := c.SetValue("f", "Ber"); err != nil {
		t.Fatal(err)
	}
	if l, _ := shown(c); !slices.Equal(l, []string{"Berlin", "Bern"}) {
		t.Errorf("after an edit it shows %q", l)
	}
	if !key(t, c, "ArrowRight", true) || c.S.Data.Value("/q") != "Berlin" {
		t.Errorf("→ at the end: /q is %v", c.S.Data.Value("/q"))
	}
	if len(*sent) != 3 {
		t.Errorf("onInput ran %d times, want 3", len(*sent))
	}
	// A value left as it was sends nothing.
	if err := c.SetValue("f", "Berlin"); err != nil {
		t.Fatal(err)
	}
	if len(*sent) != 3 {
		t.Errorf("an unchanged value ran onInput: %d", len(*sent))
	}
}

// The arrows and Control+n and Control+p move the highlight round the
// list from nothing highlighted; Enter picks it and Escape shuts the list,
// which ArrowDown opens again at its first. Enter with nothing highlighted
// and Escape with nothing shown are not the suggestions'.
func TestSuggestMove(t *testing.T) {
	c, _ := suggesting(t, "b", "Berlin", "Bern", "Bergen")
	if key(t, c, "Enter", true) {
		t.Error("Enter took a suggestion with none highlighted")
	}
	for _, step := range []struct{ key, hi string }{
		{"ArrowDown", "Berlin"}, {"Control+n", "Bern"}, {"ArrowDown", "Bergen"}, {"ArrowDown", "Berlin"},
		{"ArrowUp", "Bergen"}, {"Control+p", "Bern"},
	} {
		if !key(t, c, step.key, true) {
			t.Fatalf("%s not taken", step.key)
		}
		if _, hi := shown(c); hi != step.hi {
			t.Errorf("%s: %q highlighted, want %q", step.key, hi, step.hi)
		}
	}
	if e := c.V.Find("f"); e.Suggestion() != "Bern" || e.Ghost() != "ern" {
		t.Errorf("the highlighted one is not Tab's: %q %q", e.Suggestion(), e.Ghost())
	}
	// The agent rewrites the list: the highlight stays on its text.
	c.S.Write("/opts", []any{"Bergen", "Bern"})
	c.Rebuild()
	if _, hi := shown(c); hi != "Bern" {
		t.Errorf("rewritten, %q highlighted", hi)
	}
	if !key(t, c, "Escape", true) {
		t.Fatal("Escape not taken")
	}
	if l, _ := shown(c); l != nil || key(t, c, "Escape", true) || key(t, c, "ArrowUp", true) {
		t.Errorf("shut, it shows %q, or takes Escape or ArrowUp", l)
	}
	if !key(t, c, "ArrowDown", true) {
		t.Fatal("ArrowDown did not open it again")
	}
	if l, hi := shown(c); len(l) != 2 || hi != "Bergen" {
		t.Errorf("opened again: %q, %q highlighted", l, hi)
	}
	if !key(t, c, "Enter", true) || c.S.Data.Value("/q") != "Bergen" {
		t.Errorf("Enter: /q is %v", c.S.Data.Value("/q"))
	}
}

// The list shows SuggestRows at most, from the first, and scrolls to keep
// the highlight in view; PickSuggestion takes one by its place in Shown.
func TestSuggestWindow(t *testing.T) {
	c, _ := suggesting(t, "a", "a1", "a2", "a3", "a4", "a5", "a6", "a7")
	if e := c.V.Find("f"); e.Height != view.SuggestRows || e.Top != 0 || len(e.Shown) != 7 {
		t.Fatalf("height %d top %d of %d", e.Height, e.Top, len(e.Shown))
	}
	key(t, c, "ArrowUp", true)
	if e := c.V.Find("f"); e.Selected != 6 || e.Top != 2 {
		t.Errorf("at the last: selected %d top %d", e.Selected, e.Top)
	}
	key(t, c, "ArrowDown", true)
	if e := c.V.Find("f"); e.Selected != 0 || e.Top != 0 {
		t.Errorf("round to the first: selected %d top %d", e.Selected, e.Top)
	}
	if err := c.PickSuggestion("f", 3); err != nil {
		t.Fatal(err)
	}
	if c.S.Data.Value("/q") != "a4" {
		t.Errorf("picked %v", c.S.Data.Value("/q"))
	}
}
