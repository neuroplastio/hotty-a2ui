package view_test

import (
	"slices"
	"testing"

	hottygo "github.com/neuroplastio/hotty-go"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// note is a surface with a HottyForm of a title, a body, a box and a
// Button, chips outside it, a HottyList, two HottyShortcuts (one with a
// label) and a HottyKeyHints, whose toggle is the given JSON ("" for
// none).
func note(t *testing.T, toggle string) *view.Controller {
	t.Helper()
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	hints := `{"id":"hints","component":"HottyKeyHints","catalogId":"` + hotty.ID + `"}`
	if toggle != "" {
		hints = `{"id":"hints","component":"HottyKeyHints","catalogId":"` + hotty.ID + `","toggle":` + toggle + `}`
	}
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `","dataModel":{"title":"","body":"","pin":false,"tags":[]}}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"Column","children":["form","tags","snacks","hints","save_key","quiet_key"]},
	 {"id":"form","component":"HottyForm","catalogId":"` + hotty.ID + `","child":"fields","onSubmit":{"event":{"name":"save"}}},
	 {"id":"fields","component":"Column","children":["title","body","pin","save"]},
	 {"id":"title","component":"TextField","label":"Title","value":{"@path":"/title"}},
	 {"id":"body","component":"TextField","label":"Body","value":{"@path":"/body"},"variant":"longText"},
	 {"id":"pin","component":"CheckBox","label":"Pin it","value":{"@path":"/pin"}},
	 {"id":"save","component":"Button","child":"save_t","action":{"event":{"name":"save"}}},
	 {"id":"save_t","component":"Text","text":"Save"},
	 {"id":"tags","component":"ChoicePicker","label":"Tags","value":{"@path":"/tags"},"variant":"multipleSelection","displayStyle":"chips",
	  "options":[{"label":"Work","value":"work"},{"label":"Home","value":"home"}]},
	 {"id":"snacks","component":"HottyList","catalogId":"` + hotty.ID + `","filterable":true,"height":1,
	  "items":[{"label":"Nutella"},{"label":"Nuts"}],"onActivate":{"event":{"name":"eat"}}},
	 ` + hints + `,
	 {"id":"save_key","component":"HottyShortcut","catalogId":"` + hotty.ID + `","key":"Control+s","press":"save","label":"save"},
	 {"id":"quiet_key","component":"HottyShortcut","catalogId":"` + hotty.ID + `","key":"Control+q","action":{"event":{"name":"quiet"}}}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	return view.NewController(p.Surface("s"))
}

// hintKeys are hints as "key desc" strings, for comparing.
func hintKeys(hs []view.Hint) []string {
	var out []string
	for _, h := range hs {
		out = append(out, h.Key+" "+h.Desc)
	}
	return out
}

// The short hints are those of the element with the keyboard, then the
// HottyShortcuts that have a label, then "? more"; the full ones the same
// in groups, with Tab, Shift+Tab and ?, and a text field's editing keys,
// from the surface's keymap. On a host, only the keys of an element the
// program works show.
func TestKeyHints(t *testing.T) {
	c := note(t, "")
	general := []string{"tab next", "shift+tab back", "? close help"}
	short, full := c.KeyHints("", false)
	if got := hintKeys(short); !slices.Equal(got, []string{"ctrl+s save", "? more"}) {
		t.Errorf("without the keyboard: %q", got)
	}
	if len(full) != 2 || !slices.Equal(hintKeys(full[0]), []string{"ctrl+s save"}) || !slices.Equal(hintKeys(full[1]), general) {
		t.Errorf("without the keyboard, the full view: %q", full)
	}
	for _, step := range []struct {
		focus string
		short []string
	}{
		{"title", []string{"enter submit"}},
		{"body", nil},
		{"pin", []string{"space toggle", "enter submit"}},
		{"save", []string{"enter press"}},
		{"tags_0", []string{"space pick"}},
		{"snacks", []string{"↑/k up", "↓/j down", "/ filter", "enter choose"}},
	} {
		id := step.focus
		if id == "tags_0" {
			id = c.V.Find("tags").Children[0].ID
		}
		c.Focus(id)
		short, _ := c.KeyHints("", false)
		want := append(step.short, "ctrl+s save", "? more")
		if got := hintKeys(short); !slices.Equal(got, want) {
			t.Errorf("%s focused: %q, want %q", step.focus, got, want)
		}
	}
	c.Focus("snacks")
	_, full = c.KeyHints("", false)
	if got := hintKeys(full[0]); !slices.Equal(got, []string{"↑/k up", "↓/j down", "→/l/pgdn next page", "←/h/pgup prev page",
		"g/home go to start", "G/end go to end"}) || !slices.Equal(hintKeys(full[1]), []string{"/ filter", "enter choose"}) {
		t.Errorf("the list's full hints: %q", full)
	}
	if _, err := c.ListKey("snacks", "/"); err != nil {
		t.Fatal(err)
	}
	if short, _ = c.KeyHints("", false); !slices.Equal(hintKeys(short)[:2], []string{"enter apply filter", "esc cancel"}) {
		t.Errorf("while the filter is typed: %q", hintKeys(short))
	}
	// A field's keys are the surface's keymap's, with the components'
	// over it.
	c.Focus("title")
	_, full = c.KeyHints(hottygo.TerminalKeys, false)
	if got := hintKeys(full[0]); !slices.Equal(got, []string{"←/ctrl+b character backward", "→/ctrl+f character forward",
		"ctrl+←/alt+← word backward", "ctrl+→/alt+→ word forward", "home line start", "end/ctrl+e line end",
		"ctrl+home/alt+< input begin", "ctrl+end/alt+> input end", "ctrl+a select all"}) {
		t.Errorf("a field's moves: %q", got)
	}
	if got := hintKeys(full[1]); !slices.Equal(got, []string{"backspace/ctrl+h delete character backward",
		"delete/ctrl+d delete character forward", "ctrl+backspace/alt+backspace delete word backward", "ctrl+delete/alt+delete delete word forward",
		"ctrl+u delete before cursor", "ctrl+k delete after cursor", "enter submit"}) {
		t.Errorf("a field's edits: %q", got)
	}
	c.Focus("body")
	if _, full = c.KeyHints(hottygo.TerminalKeys, false); !slices.Contains(hintKeys(full[0]), "↑/ctrl+p line up") ||
		!slices.Contains(hintKeys(full[1]), "enter/ctrl+m insert newline") || slices.Contains(hintKeys(full[1]), "enter submit") {
		t.Errorf("a long text's full hints: %q", full)
	}
	if got := hintKeys(full[len(full)-1]); !slices.Equal(got, general) {
		t.Errorf("the general group: %q", got)
	}
	// On a host, a field's keys are left out, a list's are not.
	if short, full = c.KeyHints("", true); !slices.Equal(hintKeys(short), []string{"ctrl+s save", "? more"}) || len(full) != 2 {
		t.Errorf("on a host, the body focused: %q, %q", hintKeys(short), full)
	}
	c.Focus("snacks")
	if short, _ = c.KeyHints("", true); hintKeys(short)[0] != "enter apply filter" {
		t.Errorf("on a host, the list focused: %q", hintKeys(short))
	}
}

// ? switches the full view on and off while a HottyKeyHints takes it, and
// the element shows which is on; with toggle off, ? is not the surface's,
// and nothing says "? more".
func TestToggleHints(t *testing.T) {
	c := note(t, "")
	if c.V.Find("hints").Open || !c.ToggleHints() || !c.V.Find("hints").Open {
		t.Fatal("? does not open the full view")
	}
	if !c.ToggleHints() || c.V.Find("hints").Open {
		t.Fatal("? again does not close it")
	}
	c = note(t, "false")
	if c.ToggleHints() || c.St.FullHints {
		t.Error("? switched hints whose toggle is off")
	}
	short, full := c.KeyHints("", false)
	if got := hintKeys(short); !slices.Equal(got, []string{"ctrl+s save"}) || len(full[len(full)-1]) != 2 {
		t.Errorf("toggle off: %q, %q", got, full)
	}
}

// KeyText writes keys as bubbles' help does.
func TestKeyText(t *testing.T) {
	for key, want := range map[string]string{
		"Control+s": "ctrl+s", "ArrowUp": "↑", "Alt+ArrowLeft": "alt+←", "Escape": "esc", "PageDown": "pgdn",
		"Shift+Tab": "shift+tab", "G": "G", "/": "/", "Control++": "ctrl++", "Enter": "enter", "F5": "f5",
	} {
		if got := view.KeyText(key); got != want {
			t.Errorf("%s: %q, want %q", key, got, want)
		}
	}
}
