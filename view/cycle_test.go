package view_test

import (
	"slices"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// cycles is a surface with ChoicePickers that ask for cycle: one that
// picks one (env, staging picked), one with nothing picked (tier), and
// two it does not apply to, chips and several, which keep their options
// as children; and its data model.
func cycles(t *testing.T) (*view.Controller, func(string) any) {
	t.Helper()
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	cyc := `"metadata":{"extensions":{"io_neuroplast_hotty":{"cycle":true}}}`
	opts := `"options":[{"label":"production","value":"prod"},{"label":"staging","value":"stage"},{"label":"preview","value":"prev"}]`
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `","dataModel":{"env":["stage"],"tier":[]}}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"Column","children":["env","tier","chips","many"]},
	 {"id":"env","component":"ChoicePicker","label":"Environment","value":{"@path":"/env"},` + opts + `,` + cyc + `},
	 {"id":"tier","component":"ChoicePicker","value":{"@path":"/tier"},` + opts + `,` + cyc + `},
	 {"id":"chips","component":"ChoicePicker","displayStyle":"chips","value":["prod"],` + opts + `,` + cyc + `},
	 {"id":"many","component":"ChoicePicker","variant":"multipleSelection","value":["prod"],` + opts + `,` + cyc + `}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	s := p.Surface("s")
	return view.NewController(s), func(path string) any {
		v, _ := s.Data.Value(path).([]any)
		if len(v) != 1 {
			return v
		}
		return v[0]
	}
}

// cycle applies to a ChoicePicker that would be a select: one that picks
// one, shown as checkboxes. Chips and several keep their options.
func TestCycleBuilds(t *testing.T) {
	c, _ := cycles(t)
	for id, want := range map[string]bool{"env": true, "tier": true, "chips": false, "many": false} {
		e := c.V.Find(id)
		if e.Cycle != want || e.Cycle && (len(e.Children) > 0 || !e.Focusable()) {
			t.Errorf("%s: cycle %v, %d children, focusable %v", id, e.Cycle, len(e.Children), e.Focusable())
		}
	}
	if i := c.V.Find("env").Picked(); i != 1 {
		t.Errorf("env picked %d, want 1", i)
	}
}

// Its keys step it round from either end, as its arrows do; Home and End
// go to the ends, and a letter to the next option that starts with it.
// With nothing picked, going on starts at the first, and going back at
// the last.
func TestCycleKey(t *testing.T) {
	c, value := cycles(t)
	for _, step := range []struct{ key, want string }{
		{"arrow_right", "prev"}, {"arrow_right", "prod"}, {"arrow_left", "prev"}, {"arrow_up", "stage"},
		{"arrow_down", "prev"}, {"space", "prod"}, {"enter", "stage"}, {"home", "prod"}, {"end", "prev"},
		{"p", "prod"}, {"p", "prev"}, {"S", "stage"},
	} {
		if ok, err := c.CycleKey("env", step.key); !ok || err != nil {
			t.Fatalf("%s: taken %v, %v", step.key, ok, err)
		}
		if got := value("/env"); got != step.want {
			t.Fatalf("%s: env %v, want %s", step.key, got, step.want)
		}
	}
	if ok, _ := c.CycleKey("env", "x"); !ok || value("/env") != "stage" {
		t.Errorf("a letter no option starts with: taken %v, env %v", ok, value("/env"))
	}
	for _, k := range []string{"tab", "escape", "ctrl+a", "page_down"} {
		if ok, _ := c.CycleKey("env", k); ok {
			t.Errorf("%s was taken", k)
		}
	}
	if ok, _ := c.CycleKey("chips", "arrow_right"); ok {
		t.Error("chips took a cycle's key")
	}
	if err := c.StepChoice("tier", -1); err != nil || value("/tier") != "prev" {
		t.Errorf("back with none picked: tier %v, %v", value("/tier"), err)
	}
	c.S.Data.Set("/tier", []any{})
	c.Rebuild()
	if err := c.StepChoice("tier", 1); err != nil || value("/tier") != "prod" {
		t.Errorf("on with none picked: tier %v, %v", value("/tier"), err)
	}
}

// Its hints are its arrows' keys.
func TestCycleHints(t *testing.T) {
	c, _ := cycles(t)
	c.Focus("env")
	short, full := c.KeyHints("", false)
	if got := hintKeys(short); !slices.Equal(got, []string{"←/→ change"}) {
		t.Errorf("short: %q", got)
	}
	if got := hintKeys(full[0]); !slices.Equal(got, []string{"←/→ change", "home/end first/last"}) {
		t.Errorf("full: %q", got)
	}
}
