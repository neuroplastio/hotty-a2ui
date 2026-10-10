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
		// A Confirm's (instruction 24): a letter presses an answer, an arrow
		// moves the keyboard to one.
		{`{"id":"s","component":"HottyShortcut","key":"y","press":"yes","label":"yes"}`, true},
		{`{"id":"s","component":"HottyShortcut","key":"ArrowLeft","label":"move","action":{"functionCall":{"@call":"hottyFocus","args":{"id":"yes"}}}}`, true},
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
		{`{"id":"g","component":"HottyPaginator","child":"things","perPage":5,"page":{"@path":"/page"}}`, true},
		{`{"id":"g","component":"HottyPaginator","pages":{"@path":"/results/pages"},"page":{"@path":"/results/page"},"displayStyle":"numbers",
			"onChange":{"event":{"name":"search","context":{"page":{"@path":"/results/page"}}}}}`, true},
		{`{"id":"g","component":"HottyPaginator","page":2}`, false},
		{`{"id":"g","component":"HottyPaginator","child":"things","pages":3}`, false},
		{`{"id":"g","component":"HottyPaginator","pages":3,"displayStyle":"arabic"}`, false},
		{`{"id":"g","component":"HottyPaginator","child":"things","perPage":0}`, false},
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
		{`{"id":"t","component":"HottyTimer","label":"Tea","duration":180000,"running":{"@path":"/tea"},"interval":100,"format":"clock",
			"onTimeout":{"functionCall":{"@call":"hottyToast","args":{"message":"Tea is ready"}}}}`, true},
		{`{"id":"t","component":"HottyTimer","duration":{"@path":"/ms"},"onTimeout":{"event":{"name":"done"}}}`, true},
		{`{"id":"t","component":"HottyTimer","label":"Tea"}`, false},
		{`{"id":"t","component":"HottyTimer","duration":1000,"format":"digital"}`, false},
		{`{"id":"t","component":"HottyTimer","duration":1000,"interval":0}`, false},
		{`{"id":"w","component":"HottyStopwatch","label":"Lap","running":false,"interval":10}`, true},
		{`{"id":"w","component":"HottyStopwatch","onTimeout":{"event":{"name":"done"}}}`, false},
		{`{"id":"f","component":"HottyForm","child":"col","onSubmit":{"functionCall":{"@call":"hottyStartTimer","args":{"id":"tea"}}}}`, true},
		{`{"id":"f","component":"HottyForm","child":"col","onSubmit":{"functionCall":{"@call":"hottyStopTimer","args":{"id":"tea"}}}}`, true},
		{`{"id":"f","component":"HottyForm","child":"col","onSubmit":{"functionCall":{"@call":"hottyToggleTimer","args":{"id":"tea"}}}}`, true},
		{`{"id":"f","component":"HottyForm","child":"col","onSubmit":{"functionCall":{"@call":"hottyResetTimer","args":{"id":"tea"}}}}`, true},
		{`{"id":"f","component":"HottyForm","child":"col","onSubmit":{"functionCall":{"@call":"hottyResetTimer","args":{}}}}`, false},
		{`{"id":"f","component":"HottyForm","child":"col","onSubmit":{"functionCall":{"@call":"hottyStartTimer","args":{"id":"tea","at":0}}}}`, false},
		// A HottyMarkdown (instruction 27): its text, and where a link in place
		// goes before onLink runs.
		{`{"id":"d","component":"HottyMarkdown","text":"# Install\n\nSee [usage](#usage)."}`, true},
		{`{"id":"d","component":"HottyMarkdown","text":{"@path":"/page/body"},"link":{"@path":"/page/link"},
			"onLink":{"event":{"name":"open","context":{"href":{"@path":"/page/link"}}}},"weight":1}`, true},
		{`{"id":"d","component":"HottyMarkdown","link":{"@path":"/page/link"}}`, false},
		{`{"id":"d","component":"HottyMarkdown","text":"x","onLink":"open"}`, false},
		// A HottyTree's nodes may lead somewhere (href), a docs site's nav.
		{`{"id":"n","component":"HottyTree","items":[{"label":"Guide","href":"guide.md","children":[{"label":"Install","href":"guide.md#install"}]}]}`, true},
		{`{"id":"n","component":"HottyTree","items":[{"label":"Guide","href":3}]}`, false},
		{`{"id":"b","component":"HottyBigText","text":"HOTTY","size":"large","align":"center"}`, true},
		{`{"id":"b","component":"HottyBigText","text":{"@path":"/score"}}`, true},
		{`{"id":"b","component":"HottyBigText","size":"small"}`, false},
		{`{"id":"b","component":"HottyBigText","text":"HOTTY","size":"huge"}`, false},
		{`{"id":"b","component":"HottyBigText","text":"HOTTY","align":"middle"}`, false},
		{`{"id":"q","component":"HottyQRCode","value":"https://hotty.neuroplast.io","errorCorrection":"H","label":"Scan me"}`, true},
		{`{"id":"q","component":"HottyQRCode","value":{"@path":"/url"},"label":{"@path":"/hint"}}`, true},
		{`{"id":"q","component":"HottyQRCode","label":"Scan me"}`, false},
		{`{"id":"q","component":"HottyQRCode","value":"x","errorCorrection":"X"}`, false},
		{`{"id":"q","component":"HottyQRCode","value":"x","errorCorrection":"m"}`, false},
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
