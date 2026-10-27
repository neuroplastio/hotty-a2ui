package html

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	hottycat "github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/view"
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

// A blank Keycaps.Then is a gap between a sequence's caps, a
// span.k-cap-gap with no text, where ">" was; the screen reader's words
// are as they were.
func TestKeysGap(t *testing.T) {
	x := newHarness(t)
	var msgs []any
	must(t, json.Unmarshal([]byte(`[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"`+basic.ID+`"}},
{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
 {"id":"root","component":"HottyKeys","catalogId":"`+hottycat.ID+`","keys":"ctrl+b v"}]}}]`), &msgs))
	x.process(msgs...)
	r := x.rs["s"]
	s := x.h.Surface(r.name)
	r.C.Keycaps = view.Keycaps{Then: " "}
	x.update(r)
	gap := `</kbd><span class="k-cap-gap" aria-hidden="true"></span><kbd class="k-cap">`
	if html := s.HTML(); !strings.Contains(html, gap) || strings.Contains(html, `class="k-cap-sep"`) {
		t.Errorf("no gap %s in\n%s", gap, html)
	}
	if got, _ := s.Attr("root", "aria-label"); got != "Ctrl+B V" {
		t.Errorf("the label is %q", got)
	}
}

// A Row of Buttons wraps with a menu among them: a Modal whose trigger is
// a Button, or a one-item template of one, a Row of it, is inline as the
// Button is. A Row holding a Column does not wrap.
func TestRowWrapsMenus(t *testing.T) {
	x := newHarness(t)
	var msgs []any
	must(t, json.Unmarshal([]byte(`[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"`+basic.ID+`","dataModel":{"one":[{}]}}},
{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
 {"id":"root","component":"Column","children":["acts","cols"]},
 {"id":"acts","component":"Row","children":["done","menu","later"]},
 {"id":"done","component":"Button","child":"done_t","action":{"event":{"name":"done"}}},
 {"id":"done_t","component":"Text","text":"Done"},
 {"id":"menu","component":"Modal","trigger":"menu_b","content":"menu_t"},
 {"id":"menu_b","component":"Button","child":"menu_bt","action":{"event":{"name":"noop"}}},
 {"id":"menu_bt","component":"Text","text":"Project ▾"},
 {"id":"menu_t","component":"Text","text":"Inbox"},
 {"id":"later","component":"Row","children":{"componentId":"later_m","path":"/one"}},
 {"id":"later_m","component":"Modal","trigger":"later_b","content":"later_t"},
 {"id":"later_b","component":"Button","child":"later_bt","action":{"event":{"name":"noop"}}},
 {"id":"later_bt","component":"Text","text":"Postpone ▾"},
 {"id":"later_t","component":"Text","text":"Tomorrow"},
 {"id":"cols","component":"Row","children":["done2","col"]},
 {"id":"done2","component":"Button","child":"done_t","action":{"event":{"name":"done"}}},
 {"id":"col","component":"Column","children":["done_t"]}]}}]`), &msgs))
	x.process(msgs...)
	s := x.h.Surface(x.rs["s"].name)
	for id, want := range map[string]bool{"acts": true, "cols": false} {
		if cls, _ := s.Attr(id, "class"); strings.Contains(cls, "k-wrap") != want {
			t.Errorf("%s's class is %q: wraps %v, want %v", id, cls, !want, want)
		}
	}
}
