package view

import (
	"math"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/basic"
	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
)

// The basic catalog and the hotty catalog, onto the view's kinds.
func init() {
	for typ, m := range map[string]Mapper{
		"Text": func(b *Builder, n *a2ui.Node) *Element {
			return &Element{Kind: Text, Markdown: b.String(n, "text"), Variant: b.Enum(n, "variant", "body")}
		},
		"Image": func(b *Builder, n *a2ui.Node) *Element {
			return &Element{Kind: Image, URL: b.String(n, "url"), Alt: b.String(n, "description"),
				Fit: b.Enum(n, "fit", "fill"), Variant: b.Enum(n, "variant", "mediumFeature")}
		},
		"Icon": func(b *Builder, n *a2ui.Node) *Element {
			return &Element{Kind: Icon, Name: b.String(n, "name")}
		},
		"Video": func(b *Builder, n *a2ui.Node) *Element {
			return &Element{Kind: Media, Variant: "video", URL: b.String(n, "url"), Alt: "Video"}
		},
		"AudioPlayer": func(b *Builder, n *a2ui.Node) *Element {
			return &Element{Kind: Media, Variant: "audio", URL: b.String(n, "url"), Alt: fallback(b.String(n, "description"), "Audio")}
		},
		"Row": func(b *Builder, n *a2ui.Node) *Element {
			return &Element{Kind: Stack, Dir: Horizontal, Justify: b.Enum(n, "justify", "start"), Align: b.Enum(n, "align", "stretch"),
				Children: b.Children(n.Props["children"])}
		},
		"Column": func(b *Builder, n *a2ui.Node) *Element {
			return &Element{Kind: Stack, Dir: Vertical, Justify: b.Enum(n, "justify", "start"), Align: b.Enum(n, "align", "stretch"),
				Children: b.Children(n.Props["children"])}
		},
		"List": func(b *Builder, n *a2ui.Node) *Element {
			e := &Element{Kind: Stack, Dir: Axis(b.Enum(n, "direction", "vertical")), Justify: "start", Align: b.Enum(n, "align", "stretch"),
				Scroll: true, Children: b.Children(n.Props["children"])}
			for _, c := range e.Children {
				if c != nil && c.Kind == Button && e.Dir != Horizontal {
					c.Item = true
				}
			}
			return e
		},
		"Card": func(b *Builder, n *a2ui.Node) *Element {
			return &Element{Kind: Card, Children: b.Children(n.Props["child"])}
		},
		"Divider": func(b *Builder, n *a2ui.Node) *Element {
			return &Element{Kind: Divider, Dir: Axis(b.Enum(n, "axis", "horizontal"))}
		},
		"Tabs": mapTabs,
		"Modal": func(b *Builder, n *a2ui.Node) *Element {
			e := &Element{Kind: Modal, Children: b.Children(n.Props["trigger"])}
			e.Open = b.St.Modal == n.Key
			if e.Open {
				if c := b.Children(n.Props["content"]); len(c) > 0 {
					b.SetOverlay(c[0])
				}
			}
			e.Clickable = !anyFocusable(e.Children)
			return e
		},
		"Button": func(b *Builder, n *a2ui.Node) *Element {
			valid, _ := b.Checks(n)
			return &Element{Kind: Button, Variant: b.Enum(n, "variant", "default"), Disabled: !valid,
				Children: b.Children(n.Props["child"])}
		},
		"TextField": func(b *Builder, n *a2ui.Node) *Element {
			return &Element{Kind: TextField, Label: b.String(n, "label"), Value: a2ui.ToString(nilString(b.Value(n, "value"))),
				Placeholder: b.String(n, "placeholder"), Variant: b.Enum(n, "variant", "shortText"), Error: b.FieldError(n)}
		},
		"CheckBox": func(b *Builder, n *a2ui.Node) *Element {
			return &Element{Kind: CheckBox, Label: b.String(n, "label"), Value: a2ui.Truthy(b.Value(n, "value")), Error: b.FieldError(n)}
		},
		"ChoicePicker": mapChoice,
		"Slider": func(b *Builder, n *a2ui.Node) *Element {
			lo, hi := a2ui.ToNumber(b.Raw(n, "min")), a2ui.ToNumber(b.Raw(n, "max"))
			if math.IsNaN(lo) {
				lo = 0
			}
			if math.IsNaN(hi) {
				hi = 100
			}
			v := a2ui.ToNumber(b.Value(n, "value"))
			if math.IsNaN(v) {
				v = lo
			}
			e := &Element{Kind: Slider, Label: b.String(n, "label"), Min: lo, Max: hi, Value: v, Error: b.FieldError(n)}
			if steps := a2ui.ToNumber(b.Raw(n, "steps")); steps >= 1 && hi > lo {
				e.Step = (hi - lo) / math.Floor(steps)
			}
			return e
		},
		"DateTimeInput": func(b *Builder, n *a2ui.Node) *Element {
			e := &Element{Kind: DateTime, Label: b.String(n, "label"), Value: a2ui.ToString(nilString(b.Value(n, "value"))),
				Date: b.Bool(n, "enableDate"), Time: b.Bool(n, "enableTime"),
				MinISO: b.String(n, "min"), MaxISO: b.String(n, "max"), Error: b.FieldError(n)}
			if !e.Date && !e.Time {
				e.Date = true
			}
			return e
		},
	} {
		Register(basic.ID, typ, m)
	}

	Register(hotty.ID, "Form", func(b *Builder, n *a2ui.Node) *Element {
		valid, _ := b.Checks(n)
		e := &Element{Kind: Form, Disabled: !valid}
		b.InForm(n.Key, func() { e.Children = b.Children(n.Props["child"]) })
		return e
	})
	Register(hotty.ID, "Shortcut", func(b *Builder, n *a2ui.Node) *Element {
		b.AddShortcut(Shortcut{ID: n.Key, Key: b.String(n, "key"), Press: b.String(n, "press"), Label: b.String(n, "label")})
		b.out.nodes[n.Key] = n
		return nil
	})
}

func mapTabs(b *Builder, n *a2ui.Node) *Element {
	tabs, _ := n.Props["tabs"].([]any)
	sel := b.St.Tabs[n.Key]
	if sel >= len(tabs) || sel < 0 {
		sel = 0
	}
	e := &Element{Kind: Tabs, Selected: sel}
	for i, t := range tabs {
		m, _ := t.(map[string]any)
		title := ""
		if bd, ok := m["title"].(a2ui.Bound); ok {
			title = a2ui.ToString(nilString(bd.Value))
		}
		e.Children = append(e.Children, &Element{ID: SubID(n.Key, "tab", i), Kind: Tab, Type: "Tabs.tab", Label: title, Active: i == sel})
	}
	if sel < len(tabs) {
		m, _ := tabs[sel].(map[string]any)
		e.Children = append(e.Children, b.Children(m["child"])...)
	}
	return e
}

func mapChoice(b *Builder, n *a2ui.Node) *Element {
	e := &Element{Kind: Choice, Label: b.String(n, "label"), Multiple: b.Enum(n, "variant", "mutuallyExclusive") == "multipleSelection",
		Variant: b.Enum(n, "displayStyle", "checkbox"), Filter: b.Bool(n, "filterable"), Error: b.FieldError(n)}
	opts, _ := n.Props["options"].([]any)
	for _, o := range opts {
		m, _ := o.(map[string]any)
		label := ""
		if bd, ok := m["label"].(a2ui.Bound); ok {
			label = a2ui.ToString(nilString(bd.Value))
		}
		value, _ := m["value"].(string)
		e.Options = append(e.Options, ChoiceOption{Label: label, Value: value})
	}
	var picked []string
	switch v := b.Value(n, "value").(type) {
	case []any:
		for _, x := range v {
			if s, ok := x.(string); ok {
				picked = append(picked, s)
			}
		}
	case []string:
		picked = v
	case string:
		picked = []string{v}
	}
	e.Value = picked
	if e.Multiple || e.Variant == "chips" {
		for i, o := range e.Options {
			e.Children = append(e.Children, &Element{ID: SubID(n.Key, "option", i), Kind: Option, Type: "ChoicePicker.option",
				Label: o.Label, Value: o.Value, Active: contains(picked, o.Value)})
		}
	}
	return e
}

func anyFocusable(es []*Element) bool {
	for _, e := range es {
		if e.Focusable() || anyFocusable(e.Children) {
			return true
		}
	}
	return false
}

func contains(l []string, s string) bool {
	for _, x := range l {
		if x == s {
			return true
		}
	}
	return false
}

func nilString(v any) any {
	if v == nil {
		return ""
	}
	return v
}

func fallback(s, def string) string {
	if s != "" {
		return s
	}
	return def
}
