package view

import "github.com/neuroplastio/hotty-a2ui/a2ui"

// mapSwitch makes a HottySwitch's element (profile §6.17): a boolean as a
// CheckBox's is, its label, value and checks, drawn as an on/off switch.
// disabled, a DynamicBoolean, keeps it from flipping and from the
// keyboard.
func mapSwitch(b *Builder, n *a2ui.Node) *Element {
	return &Element{Kind: Switch, Label: b.String(n, "label"), Value: a2ui.Truthy(b.Value(n, "value")),
		Disabled: b.Bool(n, "disabled"), Error: b.FieldError(n)}
}

// On reports whether a Switch (or a CheckBox) is on.
func (e *Element) On() bool {
	on, _ := e.Value.(bool)
	return on
}
