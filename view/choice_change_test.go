package view_test

import (
	"reflect"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// onChange (io_neuroplast_hotty.onChange, profile §6.41) runs after a pick
// changes a ChoicePicker's value, a select's, a checkbox's of several, a
// template's, its context read after the data model has the pick; a pick
// of what is already picked runs nothing, and a ChoicePicker without it
// sends nothing.
func TestChoiceChange(t *testing.T) {
	p := a2ui.NewProcessor(basic.Catalog())
	var sent []map[string]any
	p.Send = func(o a2ui.Outbound) {
		if o.Action != nil {
			sent = append(sent, o.Action.Context)
		}
	}
	onChange := func(ctx string) string {
		return `"metadata":{"extensions":{"io_neuroplast_hotty":{"onChange":{"event":{"name":"changed","context":` + ctx + `}}}}}`
	}
	opts := `"options":[{"label":"Dev","value":"dev"},{"label":"Prod","value":"prod"}]`
	if err := p.ProcessJSON([]byte(`[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `",
	  "dataModel":{"env":["dev"],"tags":[],"plain":["dev"],"tasks":[{"id":"t1","to":["dev"]},{"id":"t2","to":["dev"]}]}}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"Column","children":["env","tags","plain","tasks"]},
	 {"id":"env","component":"ChoicePicker","label":"Env","variant":"mutuallyExclusive","value":{"@path":"/env"},` + opts + `,` + onChange(`{"env":{"@path":"/env"}}`) + `},
	 {"id":"tags","component":"ChoicePicker","label":"Tags","variant":"multipleSelection","value":{"@path":"/tags"},` + opts + `,` + onChange(`{"tags":{"@path":"/tags"}}`) + `},
	 {"id":"plain","component":"ChoicePicker","label":"Plain","variant":"mutuallyExclusive","value":{"@path":"/plain"},` + opts + `},
	 {"id":"tasks","component":"Column","children":{"componentId":"move","path":"/tasks"}},
	 {"id":"move","component":"ChoicePicker","label":"Move","variant":"mutuallyExclusive","value":{"@path":"to"},` + opts + `,` + onChange(`{"task":{"@path":"id"},"to":{"@path":"to"}}`) + `}]}}]`)); err != nil {
		t.Fatal(err)
	}
	c := view.NewController(p.Surface("s"))
	step := func(what string, do func() error, want ...map[string]any) {
		t.Helper()
		sent = nil
		if err := do(); err != nil {
			t.Fatalf("%s: %v", what, err)
		}
		if len(sent) != len(want) || len(want) > 0 && !reflect.DeepEqual(sent, want) {
			t.Errorf("%s sent %v, want %v", what, sent, want)
		}
	}
	step("a select's pick", func() error { return c.SetValue("env", "prod") }, map[string]any{"env": []any{"prod"}})
	step("the same pick", func() error { return c.SetValue("env", "prod") })
	step("a checkbox of several", func() error { return c.Activate(view.SubID("tags", "option", 1)) }, map[string]any{"tags": []any{"prod"}})
	step("another", func() error { return c.Activate(view.SubID("tags", "option", 0)) }, map[string]any{"tags": []any{"prod", "dev"}})
	step("one without onChange", func() error { return c.SetValue("plain", "prod") })
	second := c.FindComponent("move", a2ui.Scope{Path: "/tasks/1"})
	step("a template's second", func() error { return c.SetValue(second, "prod") }, map[string]any{"task": "t2", "to": []any{"prod"}})
}
