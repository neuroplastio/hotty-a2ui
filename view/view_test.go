package view_test

import (
	"encoding/json"
	"io/fs"
	"testing"
	"time"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	thirdparty "github.com/neuroplastio/hotty-a2ui/third_party"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// TestExamplesBuild builds every basic example's view: no placeholder for
// a node that resolved.
func TestExamplesBuild(t *testing.T) {
	names, _ := fs.Glob(thirdparty.BasicExamples, "a2ui/catalogs/basic/v1/examples/*.json")
	for _, name := range names {
		b, _ := fs.ReadFile(thirdparty.BasicExamples, name)
		var ex struct{ Messages []any }
		_ = json.Unmarshal(b, &ex)
		p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
		if err := p.Process(ex.Messages); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		for _, s := range p.Surfaces() {
			c := view.NewController(s)
			if c.V.Root == nil {
				t.Errorf("%s: no root", name)
				continue
			}
			c.V.Walk(func(e *view.Element) bool {
				if e.Kind == view.Placeholder {
					t.Errorf("%s: %s is a placeholder (%s)", name, e.ID, e.State)
				}
				return true
			})
		}
	}
}

func TestControls(t *testing.T) {
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	var actions []*a2ui.ActionMessage
	p.Send = func(o a2ui.Outbound) {
		if o.Action != nil {
			actions = append(actions, o.Action)
		}
	}
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `","dataModel":{"name":"","agree":false}}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"HottyForm","catalogId":"` + hotty.ID + `","child":"col","onSubmit":{"event":{"name":"send","context":{"name":{"@path":"/name"}}}}},
	 {"id":"col","component":"Column","children":["name","agree","go","tabs","key"]},
	 {"id":"name","component":"TextField","label":"Name","value":{"@path":"/name"},"checks":[{"condition":{"@call":"required","args":{"value":{"@path":"/name"}}},"message":"Name, please"}]},
	 {"id":"agree","component":"CheckBox","label":"Agree","value":{"@path":"/agree"}},
	 {"id":"go","component":"Button","child":"go_t","action":{"event":{"name":"go"}}},
	 {"id":"go_t","component":"Text","text":"Go"},
	 {"id":"tabs","component":"Tabs","tabs":[{"title":"A","child":"ta"},{"title":"B","child":"tb"}]},
	 {"id":"ta","component":"Text","text":"in A"},{"id":"tb","component":"Text","text":"in B"},
	 {"id":"key","component":"HottyShortcut","catalogId":"` + hotty.ID + `","key":"Control+Enter","press":"go"}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	c := view.NewController(p.Surface("s"))
	if got := c.V.Focusables(); len(got) != 5 || got[0] != "name" || got[3] != "tabs/tab/0" {
		t.Fatalf("focusables %v", got)
	}
	if err := c.Submit("root"); err != nil || len(actions) != 0 {
		t.Fatalf("submitted an invalid form: %v %v", err, actions)
	}
	if c.V.Find("name").Error != "Name, please" {
		t.Errorf("error %q", c.V.Find("name").Error)
	}
	_ = c.SetValue("name", "Ada")
	_ = c.Activate("agree")
	if v := p.Surface("s").Data.Root(); v.(map[string]any)["name"] != "Ada" || v.(map[string]any)["agree"] != true {
		t.Errorf("data %v", v)
	}
	if err := c.Enter("name"); err != nil || len(actions) != 1 || actions[0].Name != "send" || actions[0].Context["name"] != "Ada" {
		t.Fatalf("enter: %v %+v", err, actions)
	}
	if ok, err := c.Shortcut("Control+Enter"); !ok || err != nil || actions[1].Name != "go" {
		t.Fatalf("shortcut %v %v %+v", ok, err, actions)
	}
	_ = c.Activate("tabs/tab/1")
	if tb := c.V.Find("tabs"); tb.Selected != 1 || tb.Children[len(tb.Children)-1].ID != "tb" {
		t.Errorf("tabs %+v", tb)
	}
	c.Focus("go")
	for range 2 {
		if !c.FocusNext(false) {
			t.Fatal("Tab lost the keyboard")
		}
	}
	if c.St.Focus != "tabs/tab/1" {
		t.Errorf("focus %s", c.St.Focus)
	}
	if c.FocusNext(false) || c.St.Keyboard {
		t.Error("Tab past the last element keeps the keyboard")
	}
}

// TestBuildPageTabs: a page's view (BuildPage) holds every tab of a Tabs,
// each title with its own content, where Build holds the shown one's after
// the titles.
func TestBuildPageTabs(t *testing.T) {
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	if err := p.ProcessJSON([]byte(`[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `"}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"Tabs","tabs":[{"title":"A","child":"ta"},{"title":"B","child":"tb"}]},
	 {"id":"ta","component":"Text","text":"in A"},{"id":"tb","component":"Text","text":"in B"}]}}]`)); err != nil {
		t.Fatal(err)
	}
	c := view.NewController(p.Surface("s"))
	if kids := c.V.Root.Children; len(kids) != 3 || kids[2].ID != "ta" || len(kids[1].Children) != 0 {
		t.Fatalf("Build: %d children, the last %q", len(kids), kids[len(kids)-1].ID)
	}
	v := view.BuildPage(c.S, c.St)
	kids := v.Root.Children
	if len(kids) != 2 {
		t.Fatalf("BuildPage: %d children, want the two titles", len(kids))
	}
	for i, want := range []string{"ta", "tb"} {
		if tab := kids[i]; tab.Kind != view.Tab || len(tab.Children) != 1 || tab.Children[0].ID != want {
			t.Errorf("tab %d: %s holding %v, want %s", i, tab.Kind, tab.Children, want)
		}
	}
}

// TestSliderSnaps: a Slider's value is on its steps and reads as the step
// does: two steps of 0.05 down from 0.45 is 0.35, not 0.35000000000000003.
func TestSliderSnaps(t *testing.T) {
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	var msgs []any
	if err := json.Unmarshal([]byte(`[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"`+basic.ID+`","dataModel":{"v":0.45}}},
{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[{"id":"root","component":"Slider","min":0,"max":1,"value":{"@path":"/v"}}]}}]`), &msgs); err != nil {
		t.Fatal(err)
	}
	if err := p.Process(msgs); err != nil {
		t.Fatal(err)
	}
	c := view.NewController(p.Surfaces()[0])
	for _, step := range []struct {
		n    int
		want float64
	}{{-1, 0.4}, {-1, 0.35}, {1, 0.4}, {-20, 0}, {3, 0.15}} {
		if err := c.StepSlider("root", step.n, ""); err != nil {
			t.Fatal(err)
		}
		if got := c.S.Data.Value("/v"); got != step.want {
			t.Fatalf("after %+d steps: %v, want %v", step.n, got, step.want)
		}
	}
	if err := c.SetValue("root", 0.123456); err != nil {
		t.Fatal(err)
	}
	if got := c.S.Data.Value("/v"); got != 0.12 {
		t.Errorf("a free value set to %v, want 0.12 (a hundredth of the range)", got)
	}
}

// TestSliderWidth: a Slider's value has room for any value it snaps to,
// so whole numbers take as much as the widest.
func TestSliderWidth(t *testing.T) {
	for _, c := range []struct {
		e    view.Element
		want int
	}{
		{view.Element{Kind: view.Slider, Min: 0, Max: 1, Value: 1.0}, 4},      // 0.45
		{view.Element{Kind: view.Slider, Min: 0, Max: 100, Value: 5.0}, 3},    // 100
		{view.Element{Kind: view.Slider, Min: 0, Max: 1, Step: 0.05}, 4},      // 0.05
		{view.Element{Kind: view.Slider, Min: -5, Max: 5, Step: 0.5}, 4},      // -4.5
		{view.Element{Kind: view.Slider, Min: 0, Max: 1, Value: 0.123456}, 8}, // the value, off its steps
	} {
		if got := c.e.SliderWidth(); got != c.want {
			t.Errorf("%v..%v step %v value %v: %d, want %d", c.e.Min, c.e.Max, c.e.Step, c.e.Value, got, c.want)
		}
	}
}

// TestProgressAndSpinner: a HottyProgress's value is a fraction of its
// max, clamped, and unknown while it is absent or bound to nothing; a
// HottySpinner spins unless its active says otherwise, bound to nothing
// included.
func TestProgressAndSpinner(t *testing.T) {
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `","dataModel":{"at":3,"busy":false}}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"Column","children":["half","bound","over","none","gone","spin","still","lost","odd"]},
	 {"id":"half","component":"HottyProgress","catalogId":"` + hotty.ID + `","label":"Half","value":0.5},
	 {"id":"bound","component":"HottyProgress","catalogId":"` + hotty.ID + `","value":{"@path":"/at"},"max":4},
	 {"id":"over","component":"HottyProgress","catalogId":"` + hotty.ID + `","value":7,"max":4},
	 {"id":"none","component":"HottyProgress","catalogId":"` + hotty.ID + `","label":"Indexing"},
	 {"id":"gone","component":"HottyProgress","catalogId":"` + hotty.ID + `","value":{"@path":"/missing"}},
	 {"id":"spin","component":"HottySpinner","catalogId":"` + hotty.ID + `","label":"Working"},
	 {"id":"still","component":"HottySpinner","catalogId":"` + hotty.ID + `","spinner":"line","active":{"@path":"/busy"}},
	 {"id":"lost","component":"HottySpinner","catalogId":"` + hotty.ID + `","active":{"@path":"/missing"}},
	 {"id":"odd","component":"HottySpinner","catalogId":"` + hotty.ID + `","spinner":"moon"}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	c := view.NewController(p.Surface("s"))
	for _, want := range []struct {
		id    string
		f     float64
		known bool
	}{{"half", 0.5, true}, {"bound", 0.75, true}, {"over", 1, true}, {"none", 0, false}, {"gone", 0, false}} {
		e := c.V.Find(want.id)
		if e == nil || e.Kind != view.Progress {
			t.Fatalf("%s: %+v, want a Progress", want.id, e)
		}
		if f, ok := e.Fraction(); f != want.f || ok != want.known {
			t.Errorf("%s: fraction %v (known %v), want %v (%v)", want.id, f, ok, want.f, want.known)
		}
	}
	if e := c.V.Find("none"); e.Label != "Indexing" || e.Max != 1 {
		t.Errorf("none: label %q, max %v; want Indexing, 1 by default", e.Label, e.Max)
	}
	for _, want := range []struct {
		id     string
		active bool
		frame  string
	}{{"spin", true, "⣾"}, {"still", false, "|"}, {"lost", false, "⣾"}, {"odd", true, "🌑"}} {
		e := c.V.Find(want.id)
		if e == nil || e.Kind != view.Spinner {
			t.Fatalf("%s: %+v, want a Spinner", want.id, e)
		}
		if e.Active != want.active || e.SpinnerFrames().Frames[0] != want.frame {
			t.Errorf("%s: active %v, first frame %q; want %v, %q", want.id, e.Active, e.SpinnerFrames().Frames[0], want.active, want.frame)
		}
	}
	if err := c.S.Write("/busy", true); err != nil {
		t.Fatal(err)
	}
	c.Rebuild()
	if !c.V.Find("still").Active {
		t.Error("a spinner bound to /busy does not spin once it is true")
	}
}

// TestTable: a HottyTable's rows are text, a cell a column; a row is its
// rowKey's value, else its index. Its keys move the selection to the bound
// path, and the body's window follows it. Enter acts on the selected row,
// whose context reads it from there.
func TestTable(t *testing.T) {
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	var actions []*a2ui.ActionMessage
	p.Send = func(o a2ui.Outbound) {
		if o.Action != nil {
			actions = append(actions, o.Action)
		}
	}
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `","dataModel":{"sel":"",
	 "rows":[{"id":"a","n":1,"ok":true},{"id":"b","n":2.5},{"id":"c","n":3},{"id":"d","n":4},{"id":"e","n":5}]}}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"Column","children":["t","plain"]},
	 {"id":"t","component":"HottyTable","catalogId":"` + hotty.ID + `","columns":[{"key":"id","header":"Id"},{"key":"n","header":"N","align":"end","width":4},{"key":"ok"}],
	  "rows":{"@path":"/rows"},"rowKey":"id","selected":{"@path":"/sel"},"height":2,
	  "onActivate":{"event":{"name":"open","context":{"row":{"@path":"/sel"}}}}},
	 {"id":"plain","component":"HottyTable","catalogId":"` + hotty.ID + `","columns":[{"key":"x"}],"rows":[{"x":"p"},{"x":"q"}]}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	c := view.NewController(p.Surface("s"))
	e := c.V.Find("t")
	if e == nil || e.Kind != view.Table || !e.Focusable() {
		t.Fatalf("t: %+v, want a focusable Table", e)
	}
	if got := e.Columns; len(got) != 3 || got[1] != (view.Column{Key: "n", Header: "N", Width: 4, Align: "end"}) || got[2].Align != "start" {
		t.Errorf("columns %+v", got)
	}
	if got := e.Cells[0]; got[0] != "a" || got[1] != "1" || got[2] != "true" || e.Cells[1][1] != "2.5" || e.Cells[1][2] != "" {
		t.Errorf("cells %q", e.Cells)
	}
	if got := c.V.Find("plain").RowIDs; len(got) != 2 || got[0] != "0" || got[1] != "1" {
		t.Errorf("rows without a rowKey are %q, want their indexes", got)
	}
	if e.SelectedRow() != -1 || e.Top != 0 {
		t.Fatalf("selected %d, top %d before anything", e.SelectedRow(), e.Top)
	}
	c.Focus("t")
	if ok, err := c.TableKey("t", "ArrowUp"); !ok || err != nil {
		t.Fatalf("ArrowUp: %v %v", ok, err)
	}
	for _, step := range []struct {
		key      string
		sel, top int
	}{{"", 0, 0}, {"ArrowDown", 1, 0}, {"ArrowDown", 2, 1}, {"PageDown", 4, 3}, {"ArrowDown", 4, 3}, {"Home", 0, 0}, {"End", 4, 3},
		{"g", 0, 0}, {"G", 4, 3}, {"k", 3, 3}, {"j", 4, 3}, {"PageUp", 2, 2}} {
		if step.key != "" {
			if ok, err := c.TableKey("t", step.key); !ok || err != nil {
				t.Fatalf("%s: %v %v", step.key, ok, err)
			}
		}
		e = c.V.Find("t")
		if e.SelectedRow() != step.sel || e.Top != step.top {
			t.Fatalf("after %q: row %d, top %d; want %d, %d", step.key, e.SelectedRow(), e.Top, step.sel, step.top)
		}
	}
	if got := c.S.Data.Value("/sel"); got != "c" {
		t.Errorf("the selection wrote %v, want c", got)
	}
	if ok, _ := c.TableKey("t", "Space"); ok {
		t.Error("Space is a Table's key")
	}
	if ok, err := c.TableKey("t", "Enter"); !ok || err != nil || len(actions) != 1 || actions[0].Name != "open" || actions[0].Context["row"] != "c" {
		t.Fatalf("Enter: %v %v %+v", ok, err, actions)
	}
	// The agent moves the selection: the window follows it.
	if err := c.S.Write("/sel", "e"); err != nil {
		t.Fatal(err)
	}
	c.Rebuild()
	if e = c.V.Find("t"); e.SelectedRow() != 4 || e.Top != 3 {
		t.Errorf("selected by the agent: row %d, top %d", e.SelectedRow(), e.Top)
	}
	// Without onActivate, or with nothing selected, Enter does nothing.
	if ok, err := c.TableKey("plain", "Enter"); !ok || err != nil || len(actions) != 1 {
		t.Errorf("Enter on a table without onActivate: %v %v %+v", ok, err, actions)
	}
}

// TestUntilStep: a frame clock that waits UntilStep wakes on the next
// step, whatever the interval and wherever in a step it starts: a
// twelfth of a second (miniDot) too, where Go's zero time falls mid-step.
func TestUntilStep(t *testing.T) {
	intervals := []time.Duration{view.ProgressInterval}
	for _, s := range view.Spinners {
		intervals = append(intervals, s.Interval)
	}
	base := time.Date(2026, 10, 9, 20, 0, 0, 0, time.UTC)
	for _, d := range intervals {
		for i := range 500 {
			at := base.Add(time.Duration(i) * 7919 * time.Microsecond)
			wait := view.UntilStep(at, d)
			if wait <= 0 || wait > d {
				t.Fatalf("%v at %v: wait %v", d, at, wait)
			}
			// Steps count from the Unix epoch.
			step := func(t time.Time) int64 { return t.UnixNano() / int64(d) }
			if a, b := step(at), step(at.Add(wait)); b != a+1 {
				t.Fatalf("%v at %v: after %v, step %d (was %d)", d, at, wait, b, a)
			}
		}
	}
}
