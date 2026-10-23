package view_test

import (
	"testing"

	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// A HottyKeys is its keys as written, bound or made by a function, as a
// cheat sheet's follow the user's keymap.
func TestKeys(t *testing.T) {
	c, _ := surface(t, `{"leader":["ctrl+a"]}`, `[{"id":"root","component":"Column","children":["k","b"]},
	 {"id":"k","component":"HottyKeys","catalogId":"`+hotty.ID+`","keys":"mod+shift+p"},
	 {"id":"b","component":"HottyKeys","catalogId":"`+hotty.ID+`","keys":{"@call":"formatString","args":{"value":"${/leader/0} v"}}}]`)
	if e := c.V.Find("k"); e == nil || e.Kind != view.Keys || e.Label != "mod+shift+p" {
		t.Fatalf("the keys: %+v", e)
	}
	if e := c.V.Find("b"); e == nil || e.Label != "ctrl+a v" {
		t.Errorf("the bound keys: %+v", e)
	}
}

// KeysText writes each chord of each alternative as text does, and keeps
// the spaces and the "/"s that part them; a "/" that is a key is a chord.
func TestKeysText(t *testing.T) {
	text := func(k string) string { return "<" + k + ">" }
	for keys, want := range map[string]string{
		"ctrl+b v":          "<ctrl+b> <v>",
		"ctrl+b w t":        "<ctrl+b> <w> <t>",
		"arrow_up/k":        "<arrow_up>/<k>",
		"ctrl+b %/ctrl+b v": "<ctrl+b> <%>/<ctrl+b> <v>",
		"ctrl+/":            "<ctrl+/>",
		"ctrl+b /":          "<ctrl+b> </>",
		"+/-":               "<+>/<->",
		"":                  "",
	} {
		if got := view.KeysText(keys, text); got != want {
			t.Errorf("%q: %q, want %q", keys, got, want)
		}
	}
}
