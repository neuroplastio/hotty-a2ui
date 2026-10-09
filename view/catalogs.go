package view

import (
	"math"
	"strconv"

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

	Register(hotty.ID, "HottyForm", func(b *Builder, n *a2ui.Node) *Element {
		valid, _ := b.Checks(n)
		e := &Element{Kind: Form, Disabled: !valid}
		b.InForm(n.Key, func() { e.Children = b.Children(n.Props["child"]) })
		return e
	})
	Register(hotty.ID, "HottyProgress", func(b *Builder, n *a2ui.Node) *Element {
		hi := a2ui.ToNumber(b.Raw(n, "max"))
		if math.IsNaN(hi) || hi <= 0 {
			hi = 1
		}
		e := &Element{Kind: Progress, Label: b.String(n, "label"), Max: hi}
		if v := a2ui.ToNumber(b.Raw(n, "value")); b.Raw(n, "value") != nil && !math.IsNaN(v) {
			e.Value = math.Max(0, math.Min(v, hi))
		}
		return e
	})
	Register(hotty.ID, "HottySpinner", func(b *Builder, n *a2ui.Node) *Element {
		// active is true by default: absent, not bound to nothing.
		_, set := n.Props["active"]
		active := !set || b.Bool(n, "active")
		return &Element{Kind: Spinner, Label: b.String(n, "label"), Variant: b.Enum(n, "spinner", "dot"), Active: active}
	})
	Register(hotty.ID, "HottyTable", mapTable)
	Register(hotty.ID, "HottyShortcut", func(b *Builder, n *a2ui.Node) *Element {
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

// mapTable makes a HottyTable's element: its columns, and its rows as
// text, a cell for each column, from the row's field of the column's key.
// A row's id is its rowKey field, as text, else its index.
func mapTable(b *Builder, n *a2ui.Node) *Element {
	e := &Element{Kind: Table, Value: a2ui.ToString(b.Value(n, "selected"))}
	cols, _ := b.Raw(n, "columns").([]any)
	for _, c := range cols {
		m, _ := c.(map[string]any)
		key, _ := m["key"].(string)
		if key == "" {
			continue
		}
		col := Column{Key: key, Header: a2ui.ToString(m["header"]), Align: "start"}
		if w := a2ui.ToNumber(m["width"]); w >= 1 {
			col.Width = int(w)
		}
		switch a, _ := m["align"].(string); a {
		case "center", "end":
			col.Align = a
		}
		e.Columns = append(e.Columns, col)
	}
	rowKey := b.String(n, "rowKey")
	rows, _ := b.Raw(n, "rows").([]any)
	for i, r := range rows {
		m, _ := r.(map[string]any)
		cells := make([]string, len(e.Columns))
		for j, col := range e.Columns {
			cells[j] = a2ui.ToString(m[col.Key])
		}
		id := strconv.Itoa(i)
		if v, ok := m[rowKey]; rowKey != "" && ok && v != nil {
			id = a2ui.ToString(v)
		}
		e.Cells = append(e.Cells, cells)
		e.RowIDs = append(e.RowIDs, id)
	}
	if h := a2ui.ToNumber(b.Raw(n, "height")); h >= 1 {
		e.Height = int(h)
		top := b.St.Scroll[n.Key]
		if sel := e.SelectedRow(); sel >= 0 {
			top = min(top, sel)
			top = max(top, sel-e.Height+1)
		}
		e.Top = max(min(top, len(e.Cells)-e.Height), 0)
		b.St.Scroll[n.Key] = e.Top
	}
	return e
}
