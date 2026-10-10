package text

import (
	"strings"

	"github.com/neuroplastio/hotty-a2ui/rendition/cells"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// document is a HottyMarkdown as text (profile §4, §6.27): its blocks as a
// Text's are, a blank line between them but a tight list's items; each
// line after its frames (an item's marker or its margin, a quote's or an
// alert's "▎ "); an alert's title its mark and its title; code as it is; a
// table as a HottyTable's, its columns two spaces apart, aligned as the
// Markdown says. HTML shows nothing.
func document(e *view.Element) []string {
	if e.Doc == nil {
		return nil
	}
	var out []string
	for _, b := range e.Doc.Blocks {
		if b.Gap {
			bar, _ := prefixes(b.In[:b.GapIn])
			out = append(out, strings.TrimRight(bar, " "))
		}
		var lines []string
		switch b.Kind {
		case view.AlertTitle:
			a := view.Alerts[b.Alert]
			lines = []string{a.Mark + " " + a.Title}
		case view.Rule:
			lines = []string{"───"}
		case view.TableBlock:
			lines = docTable(b)
		default:
			var s strings.Builder
			for _, r := range b.Runs {
				s.WriteString(r.Text)
			}
			lines = strings.Split(s.String(), "\n")
		}
		first, rest := prefixes(b.In)
		for i, l := range lines {
			p := rest
			if i == 0 {
				p = first
			}
			out = append(out, strings.TrimRight(p+l, " "))
		}
	}
	return out
}

// prefixes are what a block's frames put before its first line and its
// others.
func prefixes(in []view.Frame) (first, rest string) {
	for _, f := range in {
		switch f.Kind {
		case view.FrameItem:
			pad := strings.Repeat(" ", cells.Width(f.Marker))
			if f.First {
				first += f.Marker
			} else {
				first += pad
			}
			rest += pad
		default:
			first, rest = first+"▎ ", rest+"▎ "
		}
	}
	return first, rest
}

// docTable is a document's table as text: its header and its rows, the
// columns two spaces apart and as wide as their widest cell.
func docTable(b view.DocBlock) []string {
	rows := make([][]string, len(b.Rows))
	var ws []int
	for i, row := range b.Rows {
		for j, c := range row {
			var s strings.Builder
			for _, r := range c {
				s.WriteString(r.Text)
			}
			rows[i] = append(rows[i], strings.ReplaceAll(s.String(), "\n", " "))
			if j >= len(ws) {
				ws = append(ws, 0)
			}
			ws[j] = max(ws[j], cells.Width(rows[i][j]))
		}
	}
	out := make([]string, len(rows))
	for i, row := range rows {
		parts := make([]string, len(row))
		for j, s := range row {
			pad := strings.Repeat(" ", ws[j]-cells.Width(s))
			align := ""
			if j < len(b.Align) {
				align = b.Align[j]
			}
			switch align {
			case "end":
				parts[j] = pad + s
			case "center":
				parts[j] = pad[:len(pad)/2] + s + pad[len(pad)/2:]
			default:
				parts[j] = s + pad
			}
		}
		out[i] = strings.TrimRight(strings.Join(parts, "  "), " ")
	}
	return out
}
