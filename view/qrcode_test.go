package view_test

import (
	"strings"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// A HottyQRCode is its value encoded at its level (M without one, or with
// one the catalog lacks), and its label; bound, the code follows the data
// model. A value no code holds is an error, an empty one no code. It takes
// no keyboard.
func TestQRCode(t *testing.T) {
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	h := `"catalogId":"` + hotty.ID + `"`
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `","dataModel":{"url":"https://hotty.neuroplast.io"}}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"Column","children":["site","wifi","odd","long","none"]},
	 {"id":"site","component":"HottyQRCode",` + h + `,"value":{"@path":"/url"},"label":"Open on your phone"},
	 {"id":"wifi","component":"HottyQRCode",` + h + `,"value":"WIFI:T:WPA;S:hotty;P:terminal;;","errorCorrection":"H"},
	 {"id":"odd","component":"HottyQRCode",` + h + `,"value":"12345","errorCorrection":"X"},
	 {"id":"long","component":"HottyQRCode",` + h + `,"value":"` + strings.Repeat("x", 3000) + `"},
	 {"id":"none","component":"HottyQRCode",` + h + `,"value":""}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	c := view.NewController(p.Surface("s"))
	for _, want := range []struct {
		id, value, label, level, err string
		size                         int
	}{
		{"site", "https://hotty.neuroplast.io", "Open on your phone", "M", "", 29},
		{"wifi", "WIFI:T:WPA;S:hotty;P:terminal;;", "", "H", "", 33},
		{"odd", "12345", "", "M", "", 21},
		{"long", strings.Repeat("x", 3000), "", "M", view.QRTooLong, 0},
		{"none", "", "", "M", "", 0},
	} {
		e := c.V.Find(want.id)
		size := 0
		if e.QR != nil {
			size = e.QR.Size
		}
		if e.Kind != view.QRCode || e.Value != want.value || e.Label != want.label || e.Variant != want.level || e.Error != want.err || size != want.size || e.Focusable() {
			t.Errorf("%s: %+v, %d modules a side", want.id, e, size)
		}
	}
	if err := c.S.Write("/url", "https://neuroplast.io"); err != nil {
		t.Fatal(err)
	}
	c.Rebuild()
	if e := c.V.Find("site"); e.Value != "https://neuroplast.io" || e.QR == nil || e.QR.Size != 25 {
		t.Errorf("once written: %+v", e)
	}
	if got := c.V.Focusables(); len(got) != 0 {
		t.Errorf("focusables %v", got)
	}
}
