package html

import (
	"encoding/json"
	"io"
	"io/fs"
	"strings"
	"testing"

	"github.com/neuroplastio/hotty-go"
	"github.com/neuroplastio/hotty-go/hottytest"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	hottycat "github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/story"
	thirdparty "github.com/neuroplastio/hotty-a2ui/third_party"
	"github.com/neuroplastio/hotty-a2ui/view"
)

func TestIDs(t *testing.T) {
	for _, id := range []string{"root", "row[/items/0]#2", "tabs/tab/0", "a b:c;d=e~f", "ü"} {
		dom := domID(id)
		if strings.ContainsAny(dom, " :;=[]#") {
			t.Errorf("%q: %q is not a control value", id, dom)
		}
		if got, part, ok := viewID(dom); !ok || got != id || part != "" {
			t.Errorf("%q: read back %q %q %v", id, got, part, ok)
		}
		if got, part, ok := viewID(partID(id, partError)); !ok || got != id || part != partError {
			t.Errorf("%q: part read back %q %q %v", id, got, part, ok)
		}
	}
	for _, own := range []string{surfaceID, layerID, backdropID, dialogID, "~zz", "~4", "x~2g"} {
		if id, part, ok := viewID(own); ok {
			t.Errorf("%q reads as an element's: %q %q", own, id, part)
		}
	}
}

// harness is a processor whose surfaces are on a test host.
type harness struct {
	t       *testing.T
	h       *hottytest.Host
	p       *a2ui.Processor
	rs      map[string]*Rendition
	order   []string
	seen    int
	actions []*a2ui.ActionMessage
	deltas  int
}

func newHarness(t *testing.T) *harness {
	x := &harness{t: t, h: hottytest.New(t), rs: map[string]*Rendition{}}
	x.p = a2ui.NewProcessor(basic.Catalog(), hottycat.Catalog())
	x.p.Send = func(o a2ui.Outbound) {
		if o.Action != nil {
			x.actions = append(x.actions, o.Action)
		}
	}
	return x
}

func (x *harness) send(cmds ...string) {
	x.t.Helper()
	for _, c := range cmds {
		if _, err := io.WriteString(x.h, c); err != nil {
			x.t.Fatal(err)
		}
	}
	if errs := x.h.Errors(); len(errs) > 0 {
		x.t.Fatalf("the host refused: %v", errs)
	}
}

// process hands the processor messages, then shows each surface: its
// document the first time, deltas after.
func (x *harness) process(msgs ...any) {
	x.t.Helper()
	if err := x.p.Process(msgs); err != nil {
		x.t.Fatal(err)
	}
	for _, s := range x.p.Surfaces() {
		r := x.rs[s.ID]
		if r == nil {
			name := "a2ui-" + strings.Map(func(c rune) rune {
				if hotty.ValidName(string(c)) {
					return c
				}
				return '-'
			}, s.ID)
			r = New(view.NewController(s), name)
			x.rs[s.ID] = r
			x.order = append(x.order, s.ID)
			x.send(hotty.Doc(name, r.Doc()), hotty.Place(name, hotty.Placement{Cols: 80, Rows: 24}))
			continue
		}
		r.C.Rebuild()
		d := r.Update()
		x.deltas += len(d)
		x.send(d...)
	}
}

// pump hands the host's events to the renditions, and what they answer
// to the host, until the host has nothing more to say.
func (x *harness) pump() {
	x.t.Helper()
	for {
		evs := x.h.Events()
		if x.seen == len(evs) {
			return
		}
		for _, ev := range evs[x.seen:] {
			x.seen++
			for _, r := range x.rs {
				if err := r.Event(ev); err != nil {
					x.t.Fatalf("%s %s: %v", ev.Kind, ev.Target, err)
				}
				x.send(r.Update()...)
			}
		}
	}
}

// fresh is the surface's document as the host has it, against one sent
// whole now.
func (x *harness) check(r *Rendition) {
	x.t.Helper()
	got := x.h.Surface(r.name).HTML()
	h2 := hottytest.New(x.t)
	_, _ = io.WriteString(h2, hotty.Doc(r.name, (&Rendition{C: r.C, name: r.name}).Doc()))
	if want := h2.Surface(r.name).HTML(); got != want {
		x.t.Errorf("%s: the deltas made\n%s\nthe document is\n%s", r.name, got, want)
	}
}

func examples(t *testing.T) map[string][]any {
	names, _ := fs.Glob(thirdparty.BasicExamples, "a2ui/catalogs/basic/v1/examples/*.json")
	out := map[string][]any{}
	for _, name := range names {
		b, _ := fs.ReadFile(thirdparty.BasicExamples, name)
		var ex struct{ Messages []any }
		if err := json.Unmarshal(b, &ex); err != nil {
			t.Fatal(err)
		}
		out[name[strings.LastIndex(name, "/")+1:]] = ex.Messages
	}
	if len(out) == 0 {
		t.Fatal("no examples")
	}
	return out
}

// TestExamplesStream streams every basic example a message at a time: the
// first shows a document, the rest change it by deltas only, and what
// the host then has is the document the view makes.
func TestExamplesStream(t *testing.T) {
	for name, msgs := range examples(t) {
		t.Run(name, func(t *testing.T) {
			x := newHarness(t)
			for _, m := range msgs {
				x.process(m)
			}
			docs := 0
			for _, m := range x.h.Commands() {
				if m.Get("a") == "doc" {
					docs++
				}
			}
			if docs != len(x.rs) {
				t.Errorf("%d documents for %d surfaces", docs, len(x.rs))
			}
			for _, id := range x.order {
				r := x.rs[id]
				x.check(r)
				for _, f := range r.C.V.Focusables() {
					if _, ok := x.h.Surface(r.name).Element(domID(f)); !ok {
						t.Errorf("focusable %s has no element", f)
					}
				}
			}
		})
	}
}

const controls = `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `","dataModel":{"name":"","agree":false}}},
{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
 {"id":"root","component":"Form","catalogId":"` + hottycat.ID + `","child":"col","onSubmit":{"event":{"name":"send","context":{"name":{"@path":"/name"}}}}},
 {"id":"col","component":"Column","children":["name","agree","go","tabs","key","more"]},
 {"id":"name","component":"TextField","label":"Name","value":{"@path":"/name"},"checks":[{"condition":{"@call":"required","args":{"value":{"@path":"/name"}}},"message":"Name, please"}]},
 {"id":"agree","component":"CheckBox","label":"Agree","value":{"@path":"/agree"}},
 {"id":"go","component":"Button","child":"go_t","action":{"event":{"name":"go","context":{"name":{"@path":"/name"}}}}},
 {"id":"go_t","component":"Text","text":"Go"},
 {"id":"tabs","component":"Tabs","tabs":[{"title":"A","child":"ta"},{"title":"B","child":"tb"}]},
 {"id":"ta","component":"Text","text":"in A"},{"id":"tb","component":"Text","text":"in B"},
 {"id":"key","component":"Shortcut","catalogId":"` + hottycat.ID + `","key":"Control+Enter","press":"go"},
 {"id":"more","component":"Modal","trigger":"more_b","content":"more_t"},
 {"id":"more_b","component":"Button","child":"more_bt","action":{"event":{"name":"noop"}}},
 {"id":"more_bt","component":"Text","text":"More"},
 {"id":"more_t","component":"Text","text":"*Inside*"}]}}]`

func TestControlsOnHost(t *testing.T) {
	x := newHarness(t)
	var msgs []any
	if err := json.Unmarshal([]byte(controls), &msgs); err != nil {
		t.Fatal(err)
	}
	x.process(msgs...)
	r := x.rs["s"]
	name := r.name
	data := func() map[string]any { return x.p.Surface("s").Data.Root().(map[string]any) }

	// Enter in the empty field: the form's checks fail, and the error
	// shows, by a delta.
	must(t, x.h.Submit(name, "name"))
	x.pump()
	if len(x.actions) != 0 {
		t.Fatalf("submitted an invalid form: %v", x.actions)
	}
	if got := x.h.Surface(name).TextOf(partID("name", partError)); got != "✗ Name, please" {
		t.Errorf("error %q", got)
	}

	// Typing reaches the data model at every key, and clears the error.
	must(t, x.h.Fill(name, "name", "Ada"))
	x.pump()
	if data()["name"] != "Ada" {
		t.Errorf("name %v", data()["name"])
	}
	if got := x.h.Surface(name).TextOf(partID("name", partError)); got != "" {
		t.Errorf("error stays: %q", got)
	}
	must(t, x.h.Check(name, "agree", true))
	x.pump()
	if data()["agree"] != true {
		t.Errorf("agree %v", data()["agree"])
	}

	// Enter submits, with the field current.
	must(t, x.h.Submit(name, "name"))
	x.pump()
	if len(x.actions) != 1 || x.actions[0].Name != "send" || x.actions[0].Context["name"] != "Ada" {
		t.Fatalf("submit: %+v", x.actions)
	}

	// A tab is shown by a click.
	must(t, x.h.Click(name, domID("tabs/tab/1")))
	x.pump()
	if got := x.h.Surface(name).TextOf(partID("tabs", partPanel)); got != "in B" {
		t.Errorf("panel %q", got)
	}
	if v, _ := x.h.Surface(name).Attr(domID("tabs/tab/1"), "aria-selected"); v != "true" {
		t.Errorf("tab 1 aria-selected %q", v)
	}

	// The Shortcut, while the field has the keyboard and an edit the
	// host has not committed: a=blur, the change, the blur, and then the
	// Button runs on the value typed; the keyboard comes back.
	must(t, x.h.Fill(name, "name", "Grace"))
	x.pump()
	cmds, ok, err := r.Key("Control+Enter")
	if !ok || err != nil || len(cmds) != 1 || !strings.Contains(cmds[0], "a=blur") {
		t.Fatalf("Key: %q %v %v", cmds, ok, err)
	}
	x.send(cmds...)
	x.pump()
	if n := len(x.actions); n != 2 || x.actions[1].Name != "go" || x.actions[1].Context["name"] != "Grace" {
		t.Fatalf("shortcut: %+v", x.actions)
	}
	if got := x.h.Surface(name).Focused(); got != "name" {
		t.Errorf("focus after the shortcut: %q", got)
	}

	// The Modal opens over the surface, which goes inert, and closes on
	// Escape.
	must(t, x.h.Click(name, "more_b"))
	x.pump()
	if got := x.h.Surface(name).TextOf(dialogID); got != "Inside" {
		t.Errorf("dialog %q", got)
	}
	if _, inert := x.h.Surface(name).Attr(surfaceID, "inert"); !inert {
		t.Error("the surface is not inert under the modal")
	}
	if _, ok, _ := r.Key("Escape"); !ok {
		t.Error("Escape did not close the modal")
	}
	x.send(r.Update()...)
	if got := x.h.Surface(name).TextOf(layerID); got != "" {
		t.Errorf("layer after Escape %q", got)
	}
	x.check(r)
	for _, m := range x.h.Commands()[1:] {
		if a := m.Get("a"); a == "doc" {
			t.Error("a document was sent again")
		}
	}
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

// TestFocusBlurOnHost: the hotty story of focus and blur, on a host. The
// field with autofocus takes the keyboard once the document is there; a
// Button's focus moves it, and blur gives it back to the terminal.
func TestFocusBlurOnHost(t *testing.T) {
	run, err := story.Start(story.Find("hotty/focus-blur"))
	if err != nil {
		t.Fatal(err)
	}
	h := hottytest.New(t)
	s := run.Surfaces()[0]
	r := New(s.C, "profile")
	send := func(cmds ...string) {
		t.Helper()
		for _, c := range cmds {
			_, _ = io.WriteString(h, c)
		}
		if errs := h.Errors(); len(errs) > 0 {
			t.Fatalf("the host refused: %v", errs)
		}
	}
	seen := 0
	pump := func() {
		t.Helper()
		for evs := h.Events(); seen < len(evs); evs = h.Events() {
			for _, ev := range evs[seen:] {
				seen++
				if err := r.Event(ev); err != nil {
					t.Fatal(err)
				}
				send(r.Update()...)
			}
		}
	}
	send(hotty.Doc("profile", r.Doc()), hotty.Place("profile", hotty.Placement{Cols: 60, Rows: 12}))
	send(r.Update()...)
	if got := h.Surface("profile").Focused(); got != "name" {
		t.Fatalf("autofocus: %q", got)
	}
	must(t, h.Click("profile", "to_city"))
	pump()
	if got := h.Surface("profile").Focused(); got != "city" {
		t.Errorf("after focus: %q", got)
	}
	must(t, h.Click("profile", "done"))
	pump()
	if got := h.Surface("profile").Focused(); got != "" || s.C.St.Keyboard {
		t.Errorf("after blur: %q, keyboard %v", got, s.C.St.Keyboard)
	}
	if n := len(run.Actions()); n != 0 {
		t.Errorf("%d actions for the agent", n)
	}
}

// TestEmoji: an emoji goes in a span that names a colour font, in a
// label's text and in a Text's HTML; a pictograph that defaults to text
// (☀ without VS16) stays text, and so does what is inside a tag.
func TestEmoji(t *testing.T) {
	var b strings.Builder
	for _, n := range texts("Sun ☀️, sun ☀, cloud ⛅ 👍🏽!") {
		b.WriteString(n.html())
	}
	want := `Sun <span class="k-emoji">☀️</span>, sun ☀, cloud <span class="k-emoji">⛅</span> <span class="k-emoji">👍🏽</span>!`
	if got := b.String(); got != want {
		t.Errorf("texts:\n got %s\nwant %s", got, want)
	}
	got := emojiHTML(`<p title="☀️">Hi ⛅ <a href="x">🇺🇦</a></p>`)
	want = `<p title="☀️">Hi <span class="k-emoji">⛅</span> <a href="x"><span class="k-emoji">🇺🇦</span></a></p>`
	if got != want {
		t.Errorf("emojiHTML:\n got %s\nwant %s", got, want)
	}
	if got := emojiHTML("<p>plain</p>"); got != "<p>plain</p>" {
		t.Errorf("emojiHTML changed plain text: %s", got)
	}
}

// TestSliderAndDate: a Slider is a track the host can draw, stepped by
// its − and + buttons; a DateTime is a text field with its form as the
// placeholder, which shows a value with an offset.
func TestSliderAndDate(t *testing.T) {
	x := newHarness(t)
	var msgs []any
	if err := json.Unmarshal([]byte(`[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"`+basic.ID+`","dataModel":{"v":3,"due":"2025-12-15T17:00:00Z"}}},
{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
 {"id":"root","component":"Column","children":["v","due"]},
 {"id":"v","component":"Slider","label":"Volume","min":0,"max":10,"value":{"@path":"/v"}},
 {"id":"due","component":"DateTimeInput","label":"Due","enableDate":true,"enableTime":true,"value":{"@path":"/due"}}]}}]`), &msgs); err != nil {
		t.Fatal(err)
	}
	x.process(msgs...)
	r := x.rs["s"]
	if err := x.h.Click(r.name, partID("v", partMore)); err != nil {
		t.Fatal(err)
	}
	x.pump()
	if err := x.h.Click(r.name, partID("v", partMore)); err != nil {
		t.Fatal(err)
	}
	x.pump()
	if got := r.C.S.Data.Value("/v"); got != 4.0 {
		t.Errorf("two steps up from 3 made %v", got)
	}
	if v, _ := x.h.Surface(r.name).Attr("v", "aria-valuenow"); v != "4" {
		t.Errorf("the track says %q", v)
	}
	x.check(r)
	if typ, _ := x.h.Surface(r.name).Attr("due", "type"); typ != "text" {
		t.Errorf("the date is an input of type %q", typ)
	}
	if v, _ := x.h.Surface(r.name).Value("due"); v != "2025-12-15T17:00:00Z" {
		t.Errorf("the date shows %q", v)
	}
	if p, _ := x.h.Surface(r.name).Attr("due", "placeholder"); p != "YYYY-MM-DDTHH:MM" {
		t.Errorf("the date's placeholder is %q", p)
	}
}
