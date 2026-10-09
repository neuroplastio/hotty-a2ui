// Package text is the rendition with no terminal: a surface's view as
// plain text, for a pipe. It says what the surface shows and holds, in
// tree order, with no layout but lines: a row's short parts share one.
package text

import (
	"fmt"
	"strings"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/rendition/cells"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// Render is a surface's view as text, ending in a newline; an open
// Modal's content follows the surface, after a rule.
func Render(v *view.Surface) string {
	lines := element(v.Root)
	if v.Overlay != nil {
		lines = append(lines, "───")
		lines = append(lines, element(v.Overlay)...)
	}
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}

func element(e *view.Element) []string {
	if e == nil || e.A11y.Hidden {
		return nil
	}
	errLine := func(lines []string) []string {
		if e.Error != "" {
			lines = append(lines, "✗ "+e.Error)
		}
		return lines
	}
	switch e.Kind {
	case view.Stack:
		var parts [][]string
		for _, c := range e.Children {
			if l := element(c); len(l) > 0 {
				parts = append(parts, l)
			}
		}
		if e.Dir == view.Horizontal && allOneLine(parts) {
			var cells []string
			for _, p := range parts {
				cells = append(cells, p[0])
			}
			if line := strings.Join(cells, "  "); line != "" {
				return []string{line}
			}
			return nil
		}
		var out []string
		for _, p := range parts {
			out = append(out, p...)
		}
		return out
	case view.Card, view.Form:
		return children(e)
	case view.Text:
		if e.Markdown == "" {
			return nil
		}
		return strings.Split(view.PlainText(view.Markdown(e.Markdown)), "\n")
	case view.Image:
		return []string{"[image: " + fallback(e.Alt, e.URL) + "]"}
	case view.Icon:
		return []string{view.IconGlyph(e.Name)}
	case view.Media:
		return []string{"▶ " + e.Alt + ": " + e.URL}
	case view.Divider:
		if e.Dir == view.Vertical {
			return []string{"│"}
		}
		return []string{"───"}
	case view.Button:
		label := oneLine(children(e))
		if e.Disabled {
			return []string{"[ " + label + " ] (disabled)"}
		}
		return []string{"[ " + label + " ]"}
	case view.TextField:
		v, _ := e.Value.(string)
		if e.Variant == "obscured" {
			v = strings.Repeat("•", len([]rune(v)))
		}
		if v == "" && e.Placeholder != "" {
			v = "(" + e.Placeholder + ")"
		}
		return errLine([]string{field(e.Label, v)})
	case view.DateTime:
		v, _ := e.Value.(string)
		return errLine([]string{field(e.Label, v)})
	case view.CheckBox:
		box := "[ ] "
		if on, _ := e.Value.(bool); on {
			box = "[x] "
		}
		return errLine([]string{box + e.Label})
	case view.Choice:
		picked, _ := e.Value.([]string)
		var labels []string
		for _, o := range e.Options {
			for _, p := range picked {
				if p == o.Value {
					labels = append(labels, o.Label)
				}
			}
		}
		return errLine([]string{field(e.Label, strings.Join(labels, ", "))})
	case view.Slider:
		f, _ := e.Value.(float64)
		return errLine([]string{field(e.Label, a2ui.NumberString(f)+" ("+a2ui.NumberString(e.Min)+"–"+a2ui.NumberString(e.Max)+")")})
	case view.Progress:
		if f, ok := e.Fraction(); ok {
			return []string{field(e.Label, fmt.Sprintf("%.0f%%", f*100))}
		}
		return []string{field(e.Label, "…")}
	case view.Table:
		return table(e)
	case view.Spinner:
		if e.Active {
			return []string{field(e.Label, "…")}
		}
		return []string{e.Label}
	case view.Tabs:
		var titles []string
		var rest []string
		for _, c := range e.Children {
			if c.Kind == view.Tab {
				if c.Active {
					titles = append(titles, "["+c.Label+"]")
				} else {
					titles = append(titles, c.Label)
				}
				continue
			}
			rest = append(rest, element(c)...)
		}
		return append([]string{strings.Join(titles, " | ")}, rest...)
	case view.Modal:
		return children(e)
	case view.Placeholder:
		if e.State == a2ui.Pending {
			return []string{"…"}
		}
		return []string{"! " + e.Type}
	}
	return children(e)
}

func children(e *view.Element) []string {
	var out []string
	for _, c := range e.Children {
		out = append(out, element(c)...)
	}
	return out
}

func allOneLine(parts [][]string) bool {
	for _, p := range parts {
		if len(p) != 1 {
			return false
		}
	}
	return true
}

func oneLine(lines []string) string { return strings.Join(lines, " ") }

func field(label, value string) string {
	if label == "" {
		return value
	}
	return label + ": " + value
}

func fallback(s, def string) string {
	if s != "" {
		return s
	}
	return def
}

// table is a Table as text: its header and every row, the columns two
// spaces apart and as wide as their widest cell, numbers and the like
// aligned as the column says; "> " marks the selected row.
func table(e *view.Element) []string {
	ws := make([]int, len(e.Columns))
	for j, c := range e.Columns {
		ws[j] = cells.Width(c.Header)
		for _, row := range e.Cells {
			ws[j] = max(ws[j], cells.Width(row[j]))
		}
	}
	format := func(row []string) string {
		parts := make([]string, len(row))
		for j, s := range row {
			pad := strings.Repeat(" ", ws[j]-cells.Width(s))
			switch e.Columns[j].Align {
			case "end":
				parts[j] = pad + s
			case "center":
				parts[j] = pad[:len(pad)/2] + s + pad[len(pad)/2:]
			default:
				parts[j] = s + pad
			}
		}
		return strings.TrimRight(strings.Join(parts, "  "), " ")
	}
	headers := make([]string, len(e.Columns))
	for j, c := range e.Columns {
		headers[j] = c.Header
	}
	out := []string{"  " + format(headers)}
	sel := e.SelectedRow()
	for i, row := range e.Cells {
		mark := "  "
		if i == sel {
			mark = "> "
		}
		out = append(out, mark+format(row))
	}
	if len(e.Cells) == 0 {
		out = append(out, "  No rows")
	}
	return out
}
