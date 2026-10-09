package html

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"strconv"
	"strings"
	"testing"
	"time"

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
	pops    map[string]bool // the open lists' surfaces the host has
	now     time.Time       // the renditions' clock, which stands still
}

func newHarness(t *testing.T) *harness {
	x := &harness{t: t, h: hottytest.New(t), rs: map[string]*Rendition{}, pops: map[string]bool{}, now: time.Unix(0, 0)}
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

// update sends the host a rendition's deltas, and plays the program's
// part for its open list: the list's surface while one is open
// (Rendition.Popover), none after.
func (x *harness) update(r *Rendition) {
	x.t.Helper()
	x.send(r.Update()...)
	_, open := r.Popover()
	name := r.PopoverName()
	switch {
	case open && !x.pops[name]:
		x.send(hotty.Doc(name, r.PopoverDoc()), hotty.Place(name, hotty.Placement{Cols: 20, Rows: 8}))
	case !open && x.pops[name]:
		x.send(hotty.Del(name))
	}
	x.pops[name] = open
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
			r.Clock = func() time.Time { return x.now }
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
				x.update(r)
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
	_, _ = io.WriteString(h2, hotty.Doc(r.name, (&Rendition{C: r.C, name: r.name, keys: r.keys, fit: r.fit, Clock: r.Clock}).Doc()))
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
 {"id":"root","component":"HottyForm","catalogId":"` + hottycat.ID + `","child":"col","onSubmit":{"event":{"name":"send","context":{"name":{"@path":"/name"}}}}},
 {"id":"col","component":"Column","children":["name","agree","go","tabs","key","more"]},
 {"id":"name","component":"TextField","label":"Name","value":{"@path":"/name"},"checks":[{"condition":{"@call":"required","args":{"value":{"@path":"/name"}}},"message":"Name, please"}]},
 {"id":"agree","component":"CheckBox","label":"Agree","value":{"@path":"/agree"}},
 {"id":"go","component":"Button","child":"go_t","action":{"event":{"name":"go","context":{"name":{"@path":"/name"}}}}},
 {"id":"go_t","component":"Text","text":"Go"},
 {"id":"tabs","component":"Tabs","tabs":[{"title":"A","child":"ta"},{"title":"B","child":"tb"}]},
 {"id":"ta","component":"Text","text":"in A"},{"id":"tb","component":"Text","text":"in B"},
 {"id":"key","component":"HottyShortcut","catalogId":"` + hottycat.ID + `","key":"Control+Enter","press":"go"},
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

// TestListRows: a List's Buttons are its rows (k-item), whatever their
// variant; a Row's are not.
func TestListRows(t *testing.T) {
	x := newHarness(t)
	var msgs []any
	if err := json.Unmarshal([]byte(`[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"`+basic.ID+`"}},
{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
 {"id":"root","component":"Column","children":["list","row"]},
 {"id":"list","component":"List","children":["a","b"]},
 {"id":"row","component":"Row","children":["c"]},
 {"id":"a","component":"Button","variant":"borderless","child":"t","action":{"event":{"name":"a"}}},
 {"id":"b","component":"Button","child":"t","action":{"event":{"name":"b"}}},
 {"id":"c","component":"Button","child":"t","action":{"event":{"name":"c"}}},
 {"id":"t","component":"Text","text":"Go"}]}}]`), &msgs); err != nil {
		t.Fatal(err)
	}
	x.process(msgs...)
	s := x.h.Surface(x.rs["s"].name)
	for id, want := range map[string]string{"a": "k-btn k-borderless k-item", "b": "k-btn k-default k-item", "c": "k-btn k-default"} {
		if got, _ := s.Attr(id, "class"); got != want {
			t.Errorf("%s's class is %q, want %q", id, got, want)
		}
	}
}

// TestNoEcho: what the user types is not sent back while they edit: the
// host has it. An echo comes a key late, and a host that set a focused
// field's value from it lost keys and the caret. Once the field is left
// the value goes out, and a value the agent sets goes out then too.
func TestNoEcho(t *testing.T) {
	x := newHarness(t)
	var msgs []any
	if err := json.Unmarshal([]byte(`[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"`+basic.ID+`","dataModel":{"text":""}}},
{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
 {"id":"root","component":"Column","children":["f","notes","out"]},
 {"id":"f","component":"TextField","label":"Type","value":{"@path":"/text"}},
 {"id":"notes","component":"TextField","label":"Notes","variant":"longText","value":{"@path":"/notes"}},
 {"id":"out","component":"Text","text":{"@call":"formatString","args":{"value":"You typed: ${/text}"}}}]}}]`), &msgs); err != nil {
		t.Fatal(err)
	}
	x.process(msgs...)
	r := x.rs["s"]
	s := x.h.Surface(r.name)
	for _, c := range []struct{ id, v string }{{"f", "a"}, {"f", "ab"}, {"notes", "line"}} {
		must(t, x.h.Fill(r.name, c.id, c.v))
		for _, ev := range x.h.Events()[x.seen:] {
			x.seen++
			must(t, r.Event(ev))
			for _, cmd := range r.Update() {
				if strings.Contains(cmd, ":t="+c.id+":") {
					t.Errorf("typing %q in %s sent back %q", c.v, c.id, cmd)
				}
				x.send(cmd)
			}
		}
	}
	if got := s.TextOf("out"); got != "You typed: ab" {
		t.Errorf("the output says %q", got)
	}
	x.send(hotty.Blur(r.name))
	x.pump()
	x.check(r)

	var set []any
	must(t, json.Unmarshal([]byte(`[{"version":"v1.0","updateDataModel":{"surfaceId":"s","path":"/text","value":"from the agent"}}]`), &set))
	x.process(set...)
	if v, _ := s.Value("f"); v != "from the agent" {
		t.Errorf("the agent's value did not reach the field: %q", v)
	}
	x.check(r)
}

// TestWeights: a weighted child says its weight, which kit.css makes its
// share of a Row (flex: weight, as A2UI's Lit renderer has it).
func TestWeights(t *testing.T) {
	x := newHarness(t)
	var msgs []any
	if err := json.Unmarshal([]byte(`[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"`+basic.ID+`"}},
{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
 {"id":"root","component":"Row","children":["a","b"]},
 {"id":"a","component":"Text","text":"Asset","weight":2},
 {"id":"b","component":"Text","text":"Price","weight":1.5}]}}]`), &msgs); err != nil {
		t.Fatal(err)
	}
	x.process(msgs...)
	s := x.h.Surface(x.rs["s"].name)
	for id, w := range map[string]string{"a": "2", "b": "1.5"} {
		class, _ := s.Attr(id, "class")
		style, _ := s.Attr(id, "style")
		if !strings.Contains(" "+class+" ", " k-weighted ") || style != "--k-w: "+w {
			t.Errorf("%s: class %q, style %q", id, class, style)
		}
	}
}

// TestSelect: a select is a button, and a click opens its list in the
// layer with the picked option highlighted. While it is open the program
// has the keyboard: a host would scroll with the arrows (SPEC §5.3). The
// arrows and letters move the highlight, Enter picks it, Escape closes
// the list without picking, and a click picks an option or, outside the
// list, closes it; the select has the keyboard again. Closed, the arrows
// and letters pick.
func TestSelect(t *testing.T) {
	x := newHarness(t)
	var msgs []any
	if err := json.Unmarshal([]byte(`[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"`+basic.ID+`","dataModel":{"size":["m"]}}},
{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
 {"id":"root","component":"Column","children":["size"]},
 {"id":"size","component":"ChoicePicker","label":"Size","value":{"@path":"/size"},
  "options":[{"label":"Small","value":"s"},{"label":"Medium","value":"m"},{"label":"Large","value":"l"}]}]}}]`), &msgs); err != nil {
		t.Fatal(err)
	}
	x.process(msgs...)
	r := x.rs["s"]
	s := x.h.Surface(r.name)
	// The open list is a surface of its own.
	list := func() *hottytest.Surface { return x.h.Surface(r.PopoverName()) }
	open := func() bool {
		if l := list(); l != nil {
			_, ok := l.Element(partID("size", partList))
			return ok
		}
		return false
	}
	hi := func() string {
		if l := list(); l != nil {
			v, _ := l.Attr(partID("size", partList), "aria-activedescendant")
			return v
		}
		return ""
	}
	size := func() string {
		v, _ := r.C.S.Data.Value("/size").([]any)
		if len(v) != 1 {
			return fmt.Sprint(v)
		}
		return fmt.Sprint(v[0])
	}
	key := func(k string) {
		t.Helper()
		_, ok, err := r.Key(k)
		must(t, err)
		if !ok {
			t.Fatalf("%s: not taken", k)
		}
		x.update(r)
		x.pump()
	}
	closed := func(when string) {
		t.Helper()
		if open() || s.Focused() != "size" {
			t.Errorf("%s: open %v, the keyboard on %q", when, open(), s.Focused())
		}
	}

	must(t, x.h.Click(r.name, "size"))
	x.pump()
	if !open() || hi() != partID("size", partOption+"1") {
		t.Fatalf("a click: open %v, highlight %q", open(), hi())
	}
	if s.Focused() != "" || !r.C.St.Keyboard {
		t.Errorf("open, the host's keyboard is on %q, the controller's %v", s.Focused(), r.C.St.Keyboard)
	}
	key("ArrowDown")
	if got, want := hi(), partID("size", partOption+"2"); got != want {
		t.Errorf("ArrowDown: highlight %q, want %q", got, want)
	}
	key("s")
	key("Enter")
	if size() != "s" {
		t.Errorf("s and Enter picked %v", size())
	}
	closed("after Enter")
	x.check(r)

	key("ArrowDown")
	if size() != "m" {
		t.Errorf("ArrowDown on the closed select picked %v", size())
	}
	key("l")
	if size() != "l" {
		t.Errorf("l on the closed select picked %v", size())
	}

	must(t, x.h.Click(r.name, "size"))
	x.pump()
	key("Home")
	key("Escape")
	closed("after Escape")
	if size() != "l" {
		t.Errorf("Escape picked %v", size())
	}
	must(t, x.h.Click(r.name, "size"))
	x.pump()
	if _, ok := s.Element(partID("size", partList)); ok {
		t.Error("the open list is in the select's surface too")
	}
	must(t, x.h.Click(r.PopoverName(), partID("size", partOption+"1")))
	x.pump()
	if size() != "m" {
		t.Errorf("a click on Medium picked %v", size())
	}
	closed("after a click on an option")
	must(t, x.h.Click(r.name, "size"))
	x.pump()
	must(t, x.h.Click(r.name, dismissID))
	x.pump()
	closed("after a click outside")
	x.check(r)
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

	// A drag: pressed on a notch near the start, moved across the track
	// to its last notch, past its end (nothing under the pointer), and let
	// go there. 0..10 has 21 notches (a twentieth, 0.5, each).
	if err := x.h.DragStart(r.name, partID("v", "k2"), 4, 2); err != nil {
		t.Fatal(err)
	}
	x.pump()
	if got := r.C.S.Data.Value("/v"); got != 1.0 {
		t.Errorf("pressed on notch 2: %v", got)
	}
	for _, step := range []struct {
		id   string
		want float64
	}{{partID("v", "k9"), 4.5}, {partID("v", "k20"), 10}, {"", 10}} {
		if err := x.h.DragMove(step.id, 30, 2); err != nil {
			t.Fatal(err)
		}
		x.pump()
		if got := r.C.S.Data.Value("/v"); got != step.want {
			t.Errorf("dragged onto %q: %v, want %v", step.id, got, step.want)
		}
	}
	if err := x.h.DragEnd("", 40, 2); err != nil {
		t.Fatal(err)
	}
	x.pump()
	if got := r.C.S.Data.Value("/v"); got != 10.0 {
		t.Errorf("let go past the end: %v", got)
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

// TestProgressAndSpinner: a HottyProgress is a progressbar, its value for
// assistive tech and a fill as wide as its fraction, k-done once full; one
// without a value has no aria-valuenow. A HottySpinner is a status: its
// first frame while it spins, and its label. Both change by deltas.
func TestProgressAndSpinner(t *testing.T) {
	x := newHarness(t)
	var msgs []any
	must(t, json.Unmarshal([]byte(`[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"`+basic.ID+`","dataModel":{"v":3,"busy":true}}},
{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
 {"id":"root","component":"Column","children":["p","i","s"]},
 {"id":"p","component":"HottyProgress","catalogId":"`+hottycat.ID+`","label":"Download","value":{"@path":"/v"},"max":8},
 {"id":"i","component":"HottyProgress","catalogId":"`+hottycat.ID+`","label":"Indexing"},
 {"id":"s","component":"HottySpinner","catalogId":"`+hottycat.ID+`","spinner":"line","label":"Working","active":{"@path":"/busy"}}]}}]`), &msgs))
	x.process(msgs...)
	r := x.rs["s"]
	s := x.h.Surface(r.name)
	for _, want := range [][3]string{{"p", "role", "progressbar"}, {"p", "aria-valuenow", "3"}, {"p", "aria-valuemax", "8"}, {"p", "aria-label", "Download"}, {"s", "role", "status"}} {
		if got, _ := s.Attr(want[0], want[1]); got != want[2] {
			t.Errorf("%s's %s is %q, want %q", want[0], want[1], got, want[2])
		}
	}
	if !strings.Contains(s.HTML(), `style="width: 37.50%"`) || !strings.Contains(s.TextOf("p"), "38%") {
		t.Errorf("the bar at 3 of 8:\n%s", s.HTML())
	}
	if v, ok := s.Attr("i", "aria-valuenow"); ok {
		t.Errorf("an indeterminate bar's value is %q", v)
	}
	if got := s.TextOf("s"); !strings.Contains(got, "|") || !strings.Contains(got, "Working") {
		t.Errorf("the spinner reads %q", got)
	}

	var set []any
	must(t, json.Unmarshal([]byte(`[{"version":"v1.0","updateDataModel":{"surfaceId":"s","path":"/v","value":8}},
{"version":"v1.0","updateDataModel":{"surfaceId":"s","path":"/busy","value":false}}]`), &set))
	x.process(set...)
	x.check(r)
	if !strings.Contains(s.HTML(), "k-progress-fill k-done") {
		t.Errorf("a full bar is not done:\n%s", s.HTML())
	}
	if got := s.TextOf("s"); strings.Contains(got, "|") || !strings.Contains(got, "Working") {
		t.Errorf("the stopped spinner reads %q", got)
	}
}

// TestProgressAndSpinnerMove: on the host as in cells, an indeterminate
// Progress sweeps and a Spinner turns on the view's clock, a delta each
// tick: the Progress's --k-at, the Spinner's frame's text. Once nothing
// moves, the rendition asks for no more ticks, and a tick sends nothing.
func TestProgressAndSpinnerMove(t *testing.T) {
	x := newHarness(t)
	var msgs []any
	must(t, json.Unmarshal([]byte(`[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"`+basic.ID+`","dataModel":{"busy":true}}},
{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
 {"id":"root","component":"Column","children":["i","s"]},
 {"id":"i","component":"HottyProgress","catalogId":"`+hottycat.ID+`","label":"Indexing","value":{"@path":"/done"}},
 {"id":"s","component":"HottySpinner","catalogId":"`+hottycat.ID+`","spinner":"line","label":"Working","active":{"@path":"/busy"}}]}}]`), &msgs))
	x.process(msgs...)
	r := x.rs["s"]
	s := x.h.Surface(r.name)
	frame := partID("s", partFrame)
	if got, _ := s.Attr("i", "style"); got != "--k-at: -25.00%" {
		t.Errorf("at tick 0 the bar starts at %q", got)
	}
	if got := s.TextOf(frame); got != "|" {
		t.Errorf("at tick 0 the frame is %q", got)
	}
	// The segment starts at step × 125% / 50 − 25%; the line's frames
	// change at the same rate, four of them.
	for _, at := range []struct {
		tick        int64
		left, frame string
	}{
		{1, "-22.50%", "/"},
		{2, "-20.00%", "-"},
		{24, "35.00%", "|"},
		{49, "97.50%", "/"},
		{50, "-25.00%", "-"},
	} {
		if d := r.Animating(); d != view.ProgressInterval {
			t.Errorf("tick %d: animates every %v, want %v", at.tick, d, view.ProgressInterval)
		}
		x.now = time.Unix(0, at.tick*int64(view.ProgressInterval))
		d := r.Update()
		want := []string{hotty.SetAttr(r.name, "i", "style", "--k-at: "+at.left), hotty.SetText(r.name, frame, at.frame)}
		if strings.Join(d, "") != strings.Join(want, "") {
			t.Errorf("tick %d: deltas %q, want %q", at.tick, d, want)
		}
		x.send(d...)
	}
	x.check(r)

	var set []any
	must(t, json.Unmarshal([]byte(`[{"version":"v1.0","updateDataModel":{"surfaceId":"s","path":"/done","value":1}},
{"version":"v1.0","updateDataModel":{"surfaceId":"s","path":"/busy","value":false}}]`), &set))
	x.process(set...)
	x.check(r)
	if d := r.Animating(); d != 0 {
		t.Errorf("a full bar and a stopped spinner animate every %v", d)
	}
	if _, ok := s.Attr("i", "style"); ok {
		t.Error("a bar with a value keeps --k-at")
	}
	x.now = x.now.Add(time.Second)
	if d := r.Update(); len(d) != 0 {
		t.Errorf("a tick with nothing moving sent %q", d)
	}
}

// TestTableOnHost: a HottyTable is a focusable box holding a table, its
// body the rows the view shows. A click on a row selects it, and a second
// acts on it; the program's arrows move the selection, and the body
// follows it, by deltas.
func TestTableOnHost(t *testing.T) {
	x := newHarness(t)
	var msgs []any
	must(t, json.Unmarshal([]byte(`[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"`+basic.ID+`","dataModel":{"sel":"",
 "rows":[{"id":"a","n":1},{"id":"b","n":2},{"id":"c","n":3}]}}},
{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
 {"id":"root","component":"HottyTable","catalogId":"`+hottycat.ID+`","columns":[{"key":"id","header":"Id"},{"key":"n","header":"N","align":"end"}],
  "rows":{"@path":"/rows"},"rowKey":"id","selected":{"@path":"/sel"},"height":2,
  "onActivate":{"event":{"name":"open","context":{"row":{"@path":"/sel"}}}}}]}}]`), &msgs))
	x.process(msgs...)
	r := x.rs["s"]
	s := x.h.Surface(r.name)
	for _, want := range [][3]string{{"root", "tabindex", "0"}, {"root", "role", "grid"}, {"root", "aria-rowcount", "3"}, {"root", "data-keys", tableKeys}} {
		if got, _ := s.Attr(want[0], want[1]); got != want[2] {
			t.Errorf("%s's %s is %q, want %q", want[0], want[1], got, want[2])
		}
	}
	row := func(i int) string { return partID("root", partRow+strconv.Itoa(i)) }
	if _, ok := s.Element(row(2)); ok {
		t.Error("the third row is in a body two rows high")
	}
	if !strings.Contains(s.TextOf("root"), "1–2 of 3") {
		t.Errorf("no position under a scrolling table: %q", s.TextOf("root"))
	}
	must(t, x.h.Click(r.name, row(1)))
	x.pump()
	if got := r.C.S.Data.Value("/sel"); got != "b" || !r.C.St.Keyboard || r.C.St.Focus != "root" {
		t.Fatalf("a click on b's row: selected %v, keyboard %v on %q", got, r.C.St.Keyboard, r.C.St.Focus)
	}
	if v, _ := s.Attr(row(1), "aria-selected"); v != "true" {
		t.Errorf("b's row is not selected on the host")
	}
	must(t, x.h.Click(r.name, row(1)))
	x.pump()
	if len(x.actions) != 1 || x.actions[0].Name != "open" || x.actions[0].Context["row"] != "b" {
		t.Fatalf("a second click: %+v", x.actions)
	}
	if _, ok, err := r.Key("ArrowDown"); !ok || err != nil {
		t.Fatalf("ArrowDown: %v %v", ok, err)
	}
	x.update(r)
	x.check(r)
	if _, ok := s.Element(row(0)); ok {
		t.Error("the first row is still in the body after it scrolled")
	}
	if v, _ := s.Attr(row(2), "aria-selected"); v != "true" {
		t.Errorf("c's row is not selected after ArrowDown")
	}
}

// TestListOnHost: a HottyList is a focusable box holding its title, its
// status line, a page of options and the page's dots. A click on an item
// selects it, and a second acts on it. The program's keys move it, turn
// its pages and type its filter, whose line takes the title's place; the
// host follows by deltas.
func TestListOnHost(t *testing.T) {
	x := newHarness(t)
	var msgs []any
	must(t, json.Unmarshal([]byte(`[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"`+basic.ID+`","dataModel":{"sel":""}}},
{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
 {"id":"root","component":"HottyList","catalogId":"`+hottycat.ID+`","title":"Snacks","selected":{"@path":"/sel"},"filterable":true,"height":2,
  "items":[{"label":"Nutella","description":"It's good on toast","value":"nutella"},{"label":"Bitter melon","value":"melon"},{"label":"Nuts","value":"nuts"}],
  "onActivate":{"event":{"name":"eat","context":{"snack":{"@path":"/sel"}}}}}]}}]`), &msgs))
	x.process(msgs...)
	r := x.rs["s"]
	s := x.h.Surface(r.name)
	for _, want := range [][3]string{{"root", "tabindex", "0"}, {"root", "role", "listbox"}, {"root", "aria-label", "Snacks"}, {"root", "data-keys", listKeys}} {
		if got, _ := s.Attr(want[0], want[1]); got != want[2] {
			t.Errorf("%s's %s is %q, want %q", want[0], want[1], got, want[2])
		}
	}
	item := func(i int) string { return partID("root", partItem+strconv.Itoa(i)) }
	if _, ok := s.Element(item(2)); ok {
		t.Error("the third item is on a page of two")
	}
	if got := s.TextOf(partID("root", partStatus)); got != "3 items" {
		t.Errorf("the status line reads %q", got)
	}
	if v, _ := s.Attr(partID("root", partDots), "aria-label"); v != "Page 1 of 2" {
		t.Errorf("the dots say %q", v)
	}
	must(t, x.h.Click(r.name, item(1)))
	x.pump()
	if got := r.C.S.Data.Value("/sel"); got != "melon" || !r.C.St.Keyboard || r.C.St.Focus != "root" {
		t.Fatalf("a click on Bitter melon: selected %v, keyboard %v on %q", got, r.C.St.Keyboard, r.C.St.Focus)
	}
	if v, _ := s.Attr(item(1), "aria-selected"); v != "true" {
		t.Errorf("Bitter melon is not selected on the host")
	}
	must(t, x.h.Click(r.name, item(1)))
	x.pump()
	if len(x.actions) != 1 || x.actions[0].Name != "eat" || x.actions[0].Context["snack"] != "melon" {
		t.Fatalf("a second click: %+v", x.actions)
	}
	key := func(k string) {
		t.Helper()
		if _, ok, err := r.Key(k); !ok || err != nil {
			t.Fatalf("%s: %v %v", k, ok, err)
		}
		x.update(r)
		x.check(r)
	}
	key("ArrowDown")
	if _, ok := s.Element(item(0)); ok || r.C.S.Data.Value("/sel") != "nuts" {
		t.Errorf("ArrowDown: the first page still shows, or %v is selected", r.C.S.Data.Value("/sel"))
	}
	key("/")
	key("n")
	key("u")
	if got := s.TextOf(partID("root", partTitle)); got != "Filter:nu" {
		t.Errorf("the filter line reads %q", got)
	}
	if got := s.TextOf(partID("root", partStatus)); got != "2 items • 1 filtered" {
		t.Errorf("filtered, the status line reads %q", got)
	}
	if !strings.Contains(s.HTML(), `<span class="k-match">Nu</span>`) {
		t.Errorf("the matched characters are not marked:\n%s", s.HTML())
	}
	key("Escape")
	if got := s.TextOf(partID("root", partTitle)); got != "Snacks" {
		t.Errorf("Escape left the title as %q", got)
	}
}
