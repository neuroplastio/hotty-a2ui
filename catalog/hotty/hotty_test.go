package hotty_test

import (
	"encoding/json"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/a2ui/schema"
	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
)

func TestCatalog(t *testing.T) {
	c := hotty.Catalog()
	if c.ID != hotty.ID {
		t.Errorf("id %q", c.ID)
	}
	if k := c.Components["HottyShortcut"].Props["press"].Kind; k != a2ui.Static {
		t.Errorf("Shortcut.press is %v, want a plain id", k)
	}
	if k := c.Components["HottyForm"].Props["child"].Kind; k != a2ui.ChildRef {
		t.Errorf("Form.child is %v, want a child", k)
	}
	for _, tc := range []struct {
		def string
		ok  bool
	}{
		{`{"id":"s","component":"HottyShortcut","key":"Control+s","press":"save"}`, true},
		{`{"id":"s","component":"HottyShortcut","key":"Escape","action":{"event":{"name":"close"}},"label":"Close"}`, true},
		{`{"id":"s","component":"HottyShortcut","key":"Control++","press":"zoom"}`, true},
		{`{"id":"s","component":"HottyShortcut","key":"Control+s"}`, false},
		{`{"id":"s","component":"HottyShortcut","key":"Control+s","press":"save","action":{"event":{"name":"x"}}}`, false},
		{`{"id":"s","component":"HottyShortcut","key":"Hyper+s","press":"save"}`, false},
		{`{"id":"f","component":"HottyForm","child":"col","onSubmit":{"event":{"name":"send"}}}`, true},
		{`{"id":"f","component":"HottyForm","child":"col","onSubmit":{"functionCall":{"@call":"hottyFocus","args":{"id":"name"}}}}`, true},
		{`{"id":"f","component":"HottyForm","child":"col","onSubmit":{"functionCall":{"@call":"hottyBlur"}}}`, true},
		{`{"id":"f","component":"HottyForm","child":"col","onSubmit":{"functionCall":{"@call":"hottyFocus"}}}`, false},
		{`{"id":"f","component":"HottyForm","child":"col"}`, false},
		{`{"id":"w","component":"HottySwitch","label":"Wi-Fi","value":{"@path":"/wifi"}}`, true},
		{`{"id":"w","component":"HottySwitch","value":true,"disabled":{"@path":"/airplane"},"checks":[{"condition":{"@path":"/wifi"},"message":"On, please"}]}`, true},
		{`{"id":"w","component":"HottySwitch","label":"Wi-Fi"}`, false},
		{`{"id":"w","component":"HottySwitch","value":"yes"}`, false},
		{`{"id":"p","component":"HottyRangeSlider","label":"Price","start":{"@path":"/price/start"},"end":{"@path":"/price/end"},"max":100,"steps":20}`, true},
		{`{"id":"p","component":"HottyRangeSlider","start":10,"end":90,"min":0,"max":100,"disabled":{"@path":"/off"},"checks":[{"condition":{"@path":"/ok"},"message":"Pick a range"}]}`, true},
		{`{"id":"p","component":"HottyRangeSlider","start":10,"max":100}`, false},
		{`{"id":"p","component":"HottyRangeSlider","start":10,"end":90}`, false},
		{`{"id":"p","component":"HottyRangeSlider","start":10,"end":90,"max":100,"steps":0}`, false},
		{`{"id":"f","component":"HottyForm","child":"col","onSubmit":{"functionCall":{"@call":"hottyToast","args":{"message":"Saved","kind":"success"}}}}`, true},
		{`{"id":"f","component":"HottyForm","child":"col","onSubmit":{"functionCall":{"@call":"hottyToast","args":{"message":{"@path":"/msg"},"id":"save","timeout":0,
			"actionLabel":"Undo","action":{"event":{"name":"undo","context":{"id":{"@path":"/id"}}}}}}}}`, true},
		{`{"id":"f","component":"HottyForm","child":"col","onSubmit":{"functionCall":{"@call":"hottyToast","args":{"message":"Saved","action":{"event":{"name":"undo"}}}}}}`, false},
		{`{"id":"f","component":"HottyForm","child":"col","onSubmit":{"functionCall":{"@call":"hottyToast","args":{"message":"Saved","actionLabel":"Undo"}}}}`, false},
		{`{"id":"f","component":"HottyForm","child":"col","onSubmit":{"functionCall":{"@call":"hottyToast","args":{"message":"Saved","actionLabel":"Go",
			"action":{"functionCall":{"@call":"hottyBlur"}}}}}}`, false},
		{`{"id":"f","component":"HottyForm","child":"col","onSubmit":{"functionCall":{"@call":"hottyToast","args":{"message":"Saved","kind":"fine"}}}}`, false},
		{`{"id":"f","component":"HottyForm","child":"col","onSubmit":{"functionCall":{"@call":"hottyToast","args":{"message":"Saved","timeout":-1}}}}`, false},
		{`{"id":"f","component":"HottyForm","child":"col","onSubmit":{"functionCall":{"@call":"hottyDismissToast","args":{"id":"save"}}}}`, true},
		{`{"id":"f","component":"HottyForm","child":"col","onSubmit":{"functionCall":{"@call":"hottyDismissToast"}}}`, false},
	} {
		var d map[string]any
		if err := json.Unmarshal([]byte(tc.def), &d); err != nil {
			t.Fatal(err)
		}
		if err := (&schema.Validator{}).Component(c, d); (err == nil) != tc.ok {
			t.Errorf("%s: %v, want ok=%v", tc.def, err, tc.ok)
		}
	}
}
