package html

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	hottycat "github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	hotty "github.com/neuroplastio/hotty-go"
)

// TestKeys: a HottyKeys is its caps, as a picker row's keys are, one
// image to a screen reader labelled with the caps' words, for the user's
// system; a Row of it and a Text wraps as one of Texts does.
func TestKeys(t *testing.T) {
	x := newHarness(t)
	var msgs []any
	must(t, json.Unmarshal([]byte(`[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"`+basic.ID+`"}},
{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
 {"id":"root","component":"Column","children":["r","alt","p"]},
 {"id":"r","component":"Row","children":["k","t"]},
 {"id":"k","component":"HottyKeys","catalogId":"`+hottycat.ID+`","keys":"ctrl+b v","weight":1},
 {"id":"t","component":"Text","text":"Split right","weight":3},
 {"id":"alt","component":"HottyKeys","catalogId":"`+hottycat.ID+`","keys":"arrow_up/k"},
 {"id":"p","component":"HottyKeys","catalogId":"`+hottycat.ID+`","keys":"mod+shift+p"}]}}]`), &msgs))
	x.process(msgs...)
	r := x.rs["s"]
	s := x.h.Surface(r.name)
	caps := strings.TrimPrefix(capsOf("k-keys", []string{"\U000F0634", "B"}, []string{"V"}), `<kbd class="k-keys k-caps">`)
	if html := s.HTML(); !strings.Contains(html, caps) {
		t.Errorf("no caps %s in\n%s", caps, html)
	}
	for _, c := range []struct{ id, attr, want string }{
		{"k", "role", "img"},
		{"k", "aria-label", "Ctrl+B V"},
		{"alt", "aria-label", "↑ or K"},
		{"p", "aria-label", "Ctrl+Shift+P"},
	} {
		if got, _ := s.Attr(c.id, c.attr); got != c.want {
			t.Errorf("%s's %s is %q, want %q", c.id, c.attr, got, c.want)
		}
	}
	if cls, _ := s.Attr("r", "class"); !strings.Contains(cls, "k-wrap") {
		t.Errorf("the Row's class is %q, without k-wrap", cls)
	}
	// On a Mac, mod is ⌘, and the label a Mac's words.
	r.C.Platform = hotty.Platform{Mod: "super"}
	r.C.Rebuild()
	x.update(r)
	if got, _ := s.Attr("p", "aria-label"); got != "Command-Shift-P" {
		t.Errorf("on a Mac, mod+shift+p's label is %q", got)
	}
}
