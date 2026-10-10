package cells

import (
	"strings"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// switches is a HottyForm of switches, Wi-Fi on and Bluetooth off (both
// bound), Locked on and disabled, and a text field; and the actions it
// sent.
func switches(t *testing.T) (*Rendition, *view.Controller, *[]string) {
	t.Helper()
	p := a2ui.NewProcessor(basic.Catalog(), hotty.Catalog())
	actions := new([]string)
	p.Send = func(o a2ui.Outbound) {
		if o.Action != nil {
			*actions = append(*actions, o.Action.Name)
		}
	}
	h := `"catalogId":"` + hotty.ID + `"`
	msgs := `[{"version":"v1.0","createSurface":{"surfaceId":"s","catalogId":"` + basic.ID + `","dataModel":{"wifi":true,"bt":false}}},
	{"version":"v1.0","updateComponents":{"surfaceId":"s","components":[
	 {"id":"root","component":"HottyForm",` + h + `,"child":"col","onSubmit":{"event":{"name":"save"}}},
	 {"id":"col","component":"Column","children":["wifi","bt","lock","name"]},
	 {"id":"wifi","component":"HottySwitch",` + h + `,"label":"Wi-Fi","value":{"@path":"/wifi"}},
	 {"id":"bt","component":"HottySwitch",` + h + `,"label":"Bluetooth","value":{"@path":"/bt"}},
	 {"id":"lock","component":"HottySwitch",` + h + `,"label":"Locked","value":true,"disabled":true},
	 {"id":"name","component":"TextField","label":"Name","value":"Ada"}]}}]`
	if err := p.ProcessJSON([]byte(msgs)); err != nil {
		t.Fatal(err)
	}
	c := view.NewController(p.Surface("s"))
	return New(c), c, actions
}

// A HottySwitch is a field's gutter, then its track with the knob at its
// end, "━━●" in the accent when on and "●──" in muted when off, then its
// label; a disabled one is muted and faint, and a focused one has the
// gutter's bar and its label in the accent. A stack of them stays tight,
// as CheckBoxes do.
func TestSwitchDraws(t *testing.T) {
	r, c, _ := switches(t)
	want := "  ━━● Wi-Fi\n" +
		"  ●── Bluetooth\n" +
		"  ━━● Locked\n" +
		"\n" +
		"  Name\n" +
		"  > Ada"
	f := r.Draw(30)
	if got := f.Plain(); got != want {
		t.Fatalf("got\n%s\nwant\n%s", got, want)
	}
	for _, tc := range []struct {
		row, col int
		role     Role
		attr     Attr
	}{
		{0, 2, Accent, 0}, {0, 4, Accent, 0}, {0, 6, Fg, 0},
		{1, 2, Muted, 0}, {1, 3, Muted, 0}, {1, 6, Fg, 0},
		{2, 2, Muted, Faint}, {2, 6, Muted, Faint},
	} {
		if cell := f.Cells[tc.row][tc.col]; cell.Role != tc.role || cell.Attr != tc.attr {
			t.Errorf("row %d col %d %q: %v %v, want %v %v", tc.row, tc.col, cell.Text, cell.Role, cell.Attr, tc.role, tc.attr)
		}
	}
	c.Focus("bt")
	f = r.Draw(30)
	if bar, label := f.Cells[1][0], f.Cells[1][6]; bar.Text != "┃" || bar.Role != Accent || label.Role != Accent {
		t.Errorf("focused: bar %q %v, label %v", bar.Text, bar.Role, label.Role)
	}
	if got := r.Draw(10).Plain(); !strings.HasPrefix(got, "  ━━● Wi-…\n") {
		t.Errorf("10 wide:\n%s", got)
	}
}

// Space and Enter flip a focused switch, in a HottyForm too, where Enter
// does not submit; Tab passes over a disabled one.
func TestSwitchKeys(t *testing.T) {
	r, c, actions := switches(t)
	key := func(k string) {
		t.Helper()
		if _, err := r.Key(k); err != nil {
			t.Fatal(err)
		}
		r.Draw(30)
	}
	c.Focus("wifi")
	key("Space")
	if got := c.S.Data.Value("/wifi"); got != false {
		t.Errorf("Space: /wifi %v", got)
	}
	key("Enter")
	if got := c.S.Data.Value("/wifi"); got != true || len(*actions) != 0 {
		t.Errorf("Enter: /wifi %v, actions %v", got, *actions)
	}
	key("Tab")
	key("Tab")
	if c.St.Focus != "name" {
		t.Errorf("Tab from Bluetooth went to %q, past the disabled switch to the field", c.St.Focus)
	}
	key("Enter")
	if len(*actions) != 1 {
		t.Errorf("Enter in the field: actions %v", *actions)
	}
}

// A click on a switch or its label flips it and gives it the keyboard; a
// disabled one does not flip, and the keyboard goes back.
func TestSwitchClicks(t *testing.T) {
	r, c, _ := switches(t)
	r.Draw(30)
	if err := r.Click(8, 1); err != nil { // "Bluetooth"
		t.Fatal(err)
	}
	if got := c.S.Data.Value("/bt"); got != true || c.St.Focus != "bt" || !c.St.Keyboard {
		t.Errorf("a click on Bluetooth: /bt %v, keyboard %v on %q", got, c.St.Keyboard, c.St.Focus)
	}
	r.Draw(30)
	if err := r.Click(3, 2); err != nil { // Locked's track
		t.Fatal(err)
	}
	if e := c.V.Find("lock"); !e.On() || c.St.Keyboard {
		t.Errorf("a click on a disabled switch: on %v, keyboard %v", e.On(), c.St.Keyboard)
	}
}
