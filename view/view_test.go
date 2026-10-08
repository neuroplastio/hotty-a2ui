package view_test

import (
	"encoding/json"
	"io/fs"
	"testing"

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
	 {"id":"root","component":"Form","catalogId":"` + hotty.ID + `","child":"col","onSubmit":{"event":{"name":"send","context":{"name":{"@path":"/name"}}}}},
	 {"id":"col","component":"Column","children":["name","agree","go","tabs","key"]},
	 {"id":"name","component":"TextField","label":"Name","value":{"@path":"/name"},"checks":[{"condition":{"@call":"required","args":{"value":{"@path":"/name"}}},"message":"Name, please"}]},
	 {"id":"agree","component":"CheckBox","label":"Agree","value":{"@path":"/agree"}},
	 {"id":"go","component":"Button","child":"go_t","action":{"event":{"name":"go"}}},
	 {"id":"go_t","component":"Text","text":"Go"},
	 {"id":"tabs","component":"Tabs","tabs":[{"title":"A","child":"ta"},{"title":"B","child":"tb"}]},
	 {"id":"ta","component":"Text","text":"in A"},{"id":"tb","component":"Text","text":"in B"},
	 {"id":"key","component":"Shortcut","catalogId":"` + hotty.ID + `","key":"Control+Enter","press":"go"}]}}]`
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
