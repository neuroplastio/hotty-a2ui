// Package text is the rendition with no terminal: a surface's view as
// plain text, for a pipe. It says what the surface shows and holds, in
// tree order, with no layout but lines: a row's short parts share one.
package text

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/diff"
	"github.com/neuroplastio/hotty-a2ui/icons"
	"github.com/neuroplastio/hotty-a2ui/rendition/cells"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// Render is a surface's view as text, ending in a newline; an open
// Modal's content follows the surface, after a rule, and then the toasts,
// after another, the newest first: each its kind's mark, its message, and
// its action as a button (profile §4). A pipe has no time, so a toast is
// listed while the view has it.
func Render(v *view.Surface) string {
	lines := element(v.Root)
	if v.Overlay != nil {
		lines = append(lines, "───")
		lines = append(lines, element(v.Overlay)...)
	}
	if len(v.Toasts) > 0 {
		lines = append(lines, "───")
		for _, t := range v.Toasts {
			line := icons.Glyph(view.ToastIcons[t.Variant]) + " " + t.Label
			for _, a := range t.Children {
				line += "  [ " + a.Label + " ]"
			}
			lines = append(lines, line)
		}
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
		return []string{icons.Glyph(e.Name)}
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
	case view.Switch:
		// What it says, as a field's value is: "Wi-Fi: on".
		s := field(e.Label, map[bool]string{true: "on", false: "off"}[e.On()])
		if e.Disabled {
			s += " (disabled)"
		}
		return errLine([]string{s})
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
	case view.RangeSlider:
		// Its start and its end, as a field's value: "Price: 20–70".
		s := field(e.Label, e.RangeText())
		if e.Disabled {
			s += " (disabled)"
		}
		return errLine([]string{s})
	case view.Progress:
		if f, ok := e.Fraction(); ok {
			return []string{field(e.Label, fmt.Sprintf("%.0f%%", f*100))}
		}
		return []string{field(e.Label, "…")}
	case view.Table:
		return table(e)
	case view.RichList:
		return richList(e)
	case view.Tree:
		return tree(e)
	case view.Chart:
		return chartTable(e)
	case view.Sparkline:
		// Its values, oldest first, "–" where one is missing.
		if len(e.Series) == 0 || len(e.Series[0].Values) == 0 {
			return nil
		}
		var vs []string
		for _, v := range e.Series[0].Values {
			vs = append(vs, numberOr(v, "–"))
		}
		return []string{strings.Join(vs, " ")}
	case view.KeyHints:
		// A pipe takes no keys.
		return nil
	case view.ScrollView:
		// All of it, whatever its height, as a Table's rows.
		if len(e.Children) > 0 {
			return children(e)
		}
		return e.Lines
	case view.Paginator:
		return pager(e)
	case view.Listing:
		return listing(e)
	case view.DiffView:
		return patch(e)
	case view.Spinner:
		if e.Active {
			return []string{field(e.Label, "…")}
		}
		return []string{e.Label}
	case view.Timer:
		// The time it showed when the view was made, as a field's value:
		// "Tea: 2m57s". A pipe has no time to tick in.
		t, _ := e.Value.(string)
		return []string{field(e.Label, t)}
	case view.BigText:
		// The text as it is, a line a line: what it says, not its letters'
		// blocks.
		if e.Label == "" {
			return nil
		}
		return strings.Split(e.Label, "\n")
	case view.QRCode:
		// What it holds, as a field's value, "Scan to open: https://…":
		// text has no code to scan, and needs none, whether a code holds
		// the value or not.
		v, _ := e.Value.(string)
		if v == "" {
			return nil
		}
		return []string{field(e.Label, v)}
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

// chartTable is a HottyChart as text: its values as a table, a row a
// point, its label first (else its number, from 1), and a column a
// series, headed by its label, the numbers aligned to the end and a
// missing one blank; "No data" when it has no values.
func chartTable(e *view.Element) []string {
	if e.Empty() {
		return []string{"No data"}
	}
	rows := [][]string{{""}}
	for i, s := range e.Series {
		rows[0] = append(rows[0], fallback(s.Label, "series "+strconv.Itoa(i+1)))
	}
	if len(e.Series) == 1 && e.Series[0].Label == "" {
		rows[0][1] = "value"
	}
	for j := range e.Points() {
		row := []string{fallback(e.PointLabel(j), strconv.Itoa(j+1))}
		for i := range e.Series {
			row = append(row, numberOr(e.At(i, j), ""))
		}
		rows = append(rows, row)
	}
	ws := make([]int, len(rows[0]))
	for _, row := range rows {
		for k, s := range row {
			ws[k] = max(ws[k], cells.Width(s))
		}
	}
	out := make([]string, len(rows))
	for r, row := range rows {
		parts := make([]string, len(row))
		for k, s := range row {
			pad := strings.Repeat(" ", ws[k]-cells.Width(s))
			if k == 0 {
				parts[k] = s + pad
			} else {
				parts[k] = pad + s
			}
		}
		out[r] = strings.TrimRight(strings.Join(parts, "  "), " ")
	}
	return out
}

// numberOr is v as the data model writes it, or missing for NaN.
func numberOr(v float64, missing string) string {
	if math.IsNaN(v) {
		return missing
	}
	return a2ui.NumberString(v)
}

// richList is a HottyList as text: its title, its status line, then
// every item the filter leaves, its label and its description after a
// dash, "> " marking the selected one; its empty text when it has none.
func richList(e *view.Element) []string {
	var out []string
	if e.Label != "" {
		out = append(out, e.Label)
	}
	out = append(out, e.ListStatus())
	sel := e.SelectedRow()
	for _, i := range e.Shown {
		it := e.Items[i]
		mark := "  "
		if i == sel {
			mark = "> "
		}
		s := mark + it.Label
		if it.Description != "" {
			s += " — " + it.Description
		}
		out = append(out, s)
	}
	if len(e.Items) == 0 {
		out = append(out, "  "+e.Placeholder)
	}
	return out
}

// pager is a HottyPaginator as text: its child with every page's items,
// whatever the page, as a HottyScrollView's whole content, then the page
// the view shows, "Page 2 of 6"; that line alone for one without a child,
// whose pages the agent holds.
func pager(e *view.Element) []string {
	var out []string
	if len(e.Children) > 0 {
		all := *e.Children[0]
		all.Children = nil
		for _, p := range e.Paged {
			all.Children = append(all.Children, p...)
		}
		out = element(&all)
	}
	return append(out, "Page "+strconv.Itoa(e.Selected+1)+" of "+strconv.Itoa(e.PageCount))
}

// tree is a HottyTree as text: every node it shows, whatever its height,
// drawn as cells draws it (guides, a fold before a branch, a closed one's
// count, no fold column in a tree with no branches), "> " marking the
// selected one; its empty text when it shows none.
func tree(e *view.Element) []string {
	var out []string
	sel := e.SelectedRow()
	for _, i := range e.Shown {
		n := e.Nodes[i]
		s := "  "
		if i == sel {
			s = "> "
		}
		if n.Level > 0 {
			through, last := e.Guides(i)
			for _, on := range through {
				s += map[bool]string{true: "│   ", false: "    "}[on]
			}
			s += map[bool]string{true: "└── ", false: "├── "}[last]
		}
		switch {
		case n.Branch() && (n.Open || e.Filtering()):
			s += "▼ " + n.Label
		case n.Branch():
			s += "▶ " + n.Label + " " + strconv.Itoa(n.Kids)
		case n.Level == 0 && !e.Flat():
			s += "  " + n.Label
		default:
			s += n.Label
		}
		out = append(out, s)
	}
	if len(e.Shown) == 0 {
		out = append(out, "  "+e.Placeholder)
	}
	return out
}

// listing is a HottyCode as text: each line as it is, after its number
// when the numbers show, and its mark's sign when it has marks (+, -, ✗,
// ! and > for a highlighted line), as a pipe would want to read a diff.
// patch is a HottyDiff as text: a unified diff, which patch and git apply
// read: for each file with a name, its --- and +++ lines (git's a/ and b/,
// /dev/null for a side it does not have), then its hunks, each its header
// and its lines, signed. A binary file is git's line for one. Tabs are
// spaces, as the diff shows them.
func patch(e *view.Element) []string {
	var out []string
	for _, f := range e.Diff.Files {
		old, new := "a/"+f.Old, "b/"+f.New
		if f.Old == "" {
			old = "/dev/null"
		}
		if f.New == "" {
			new = "/dev/null"
		}
		switch {
		case f.Binary:
			out = append(out, "Binary files "+old+" and "+new+" differ")
			continue
		case f.Name() != "":
			out = append(out, "--- "+old, "+++ "+new)
		}
		for _, h := range f.Hunks {
			head := h.Header()
			if h.Section != "" {
				head += " " + h.Section
			}
			out = append(out, head)
			for _, l := range h.Lines {
				sign := " "
				switch l.Op {
				case diff.Removed:
					sign = "-"
				case diff.Added:
					sign = "+"
				}
				out = append(out, sign+l.Text())
			}
		}
	}
	return out
}

func listing(e *view.Element) []string {
	numbers := 0
	if e.Numbers {
		numbers = max(len(strconv.Itoa(e.FirstLine)), len(strconv.Itoa(e.FirstLine+len(e.Code)-1)))
	}
	signs := map[view.Mark]string{view.MarkHighlight: ">", view.MarkAdded: "+", view.MarkRemoved: "-", view.MarkError: "✗", view.MarkWarning: "!"}
	out := make([]string, 0, len(e.Code))
	for i, l := range e.Code {
		var b strings.Builder
		if numbers > 0 {
			fmt.Fprintf(&b, "%*d ", numbers, e.FirstLine+i)
		}
		if len(e.Marks) > 0 {
			sign := signs[e.Marks[e.FirstLine+i]]
			if sign == "" {
				sign = " "
			}
			b.WriteString(sign + " ")
		}
		for _, t := range l {
			b.WriteString(t.Text)
		}
		out = append(out, strings.TrimRight(b.String(), " "))
	}
	return out
}
