package cells

import (
	"github.com/neuroplastio/hotty-a2ui/highlight"
	"github.com/neuroplastio/hotty-a2ui/view"
)

// A HottyMarkdown in cells (profile §3.4, §6.27): a Text's blocks, styled
// as a Text's, a blank row between them but a tight list's items; list
// items, quotes and alerts as frames before their lines (a marker and its
// margin, a bar); an alert's title row, its mark and its title, and its
// bar in its tone; a table as a HottyTable draws its header, its cells
// wrapped in their columns. A link in place is underlined, with no OSC 8
// (it has no URL), takes a click, and is reversed in the accent while it
// has the keyboard. A heading a link in place went to is the document's
// sight (Sight), which an enclosing HottyScrollView scrolls to the top.

// alertRoles are the alerts' tones as roles.
var alertRoles = map[string]Role{"info": Info, "success": Success, "accent": Accent, "warning": Warning, "error": Error}

// alertRole is an alert's colour.
func alertRole(kind string) Role { return alertRoles[view.Alerts[kind].Tone] }

// document is a HottyMarkdown's lines at width w, unwrapped when w is
// noWrap. A rule's line is its frames' prefix, the rule painted after it
// across the width. Each row of a heading is marked with it (tline.head).
func (r *Rendition) document(e *view.Element, w int) []tline {
	if e.Doc == nil {
		return nil
	}
	var out []tline
	for _, b := range e.Doc.Blocks {
		if b.Gap {
			bar, _ := frames(b.In[:b.GapIn])
			out = append(out, tline{gs: trimSpaces(bar)})
		}
		first, rest := frames(b.In)
		if w != noWrap && (width(first) >= w || width(rest) >= w) {
			first, rest = nil, nil
		}
		inner := w
		if w != noWrap {
			inner = w - width(rest)
		}
		switch b.Kind {
		case view.CodeBlock:
			// Highlighted, two columns in from the prose, as a Text's.
			indent := codeIndent
			if inner != noWrap && inner-indent < 1 {
				indent = 0
			}
			pre := first
			for _, l := range highlight.Lines(b.Runs[0].Text, b.Lang) {
				cw := inner
				if cw != noWrap {
					cw = max(inner-indent, 1)
				}
				for _, t := range chars(tokenGlyphs(l, Fg), cw) {
					out = append(out, tline{gs: concat(pre, repeat(" ", indent, style{}), t.gs)})
					pre = rest
				}
			}
			continue
		case view.Rule:
			out = append(out, tline{gs: first, rule: true})
			continue
		case view.TableBlock:
			pre := first
			for _, row := range r.docTable(e, b, inner) {
				out = append(out, tline{gs: concat(pre, row)})
				pre = rest
			}
			continue
		}
		var gs []glyph
		head := 0
		switch b.Kind {
		case view.Heading:
			st := style{attr: Bold}
			if b.Level == 1 {
				st.attr |= Underline
			}
			gs = r.runGlyphs(e, b.Runs, st)
			head = b.Heading + 1
		case view.AlertTitle:
			a := view.Alerts[b.Alert]
			gs = glyphs(a.Mark+" "+a.Title, style{role: alertRole(b.Alert), attr: Bold})
		default:
			gs = r.runGlyphs(e, b.Runs, style{})
		}
		for _, l := range wrap(gs, w, first, rest) {
			out = append(out, tline{gs: l, head: head})
		}
	}
	return out
}

// frames are the prefixes of a block's first line and of its others, its
// frames' from the outermost: an item's marker on its first block's first
// line and as many spaces otherwise, a quote's bar in border, an alert's
// in its tone.
func frames(in []view.Frame) (first, rest []glyph) {
	for _, f := range in {
		switch f.Kind {
		case view.FrameItem:
			m := glyphs(f.Marker, style{})
			pad := repeat(" ", width(m), style{})
			if f.First {
				first = concat(first, m)
			} else {
				first = concat(first, pad)
			}
			rest = concat(rest, pad)
		case view.FrameQuote, view.FrameAlert:
			st := style{role: Border}
			if f.Kind == view.FrameAlert {
				st.role = alertRole(f.Alert)
			}
			bar := glyphs("▎ ", st)
			first, rest = concat(first, bar), concat(rest, bar)
		}
	}
	return first, rest
}

// trimSpaces is glyphs without the spaces they end in.
func trimSpaces(gs []glyph) []glyph {
	for len(gs) > 0 && gs[len(gs)-1].text == " " {
		gs = gs[:len(gs)-1]
	}
	return gs
}

// runGlyphs are a document's runs as glyphs, on top of st: a link in place
// underlined, its glyphs its hits (glyph.hit), reversed in the accent while
// it has the keyboard; a link with a URL as a Text's, an OSC 8 hyperlink.
func (r *Rendition) runGlyphs(e *view.Element, runs []view.Run, st style) []glyph {
	var out []glyph
	for _, run := range runs {
		rs := runStyle(st, run)
		id := ""
		if run.Link > 0 {
			id = view.SubID(e.ID, "link", run.Link-1)
			rs.link = ""
			if r.focused(id) {
				rs.role, rs.attr = Accent, rs.attr|Reverse
			}
		}
		for _, g := range glyphs(run.Text, rs) {
			g.hit = id
			out = append(out, g)
		}
	}
	return out
}

// docTable is a document's table in w columns, as a HottyTable draws one
// (cellPad a side, the header bold, a rule in border under it, each column
// aligned as its Align says), its columns fitted as a HottyTable's
// (fitWidths), and each cell wrapped in its column rather than cut.
func (r *Rendition) docTable(e *view.Element, b view.DocBlock, w int) [][]glyph {
	cols := len(b.Align)
	for _, row := range b.Rows {
		cols = max(cols, len(row))
	}
	cells := make([][][]glyph, len(b.Rows))
	nat := make([]int, cols)
	for i, row := range b.Rows {
		st := style{}
		if i == 0 {
			st.attr = Bold
		}
		cells[i] = make([][]glyph, cols)
		for j := range cols {
			if j < len(row) {
				cells[i][j] = r.runGlyphs(e, row[j], st)
			}
			nat[j] = max(nat[j], width(cells[i][j]))
		}
	}
	ws := fitWidths(nat, w)
	align := func(j int) string {
		if j < len(b.Align) {
			return b.Align[j]
		}
		return ""
	}
	var out [][]glyph
	for i := range cells {
		lines := make([][][]glyph, cols)
		h := 1
		for j, cw := range ws {
			if cw > 0 {
				lines[j] = wrap(cells[i][j], cw, nil, nil)
				h = max(h, len(lines[j]))
			}
		}
		for k := range h {
			var row []glyph
			for j, cw := range ws {
				if cw <= 0 {
					continue
				}
				var gs []glyph
				if k < len(lines[j]) {
					gs = lines[j][k]
				}
				pad := repeat(" ", cellPad, style{})
				row = concat(row, pad, aligned(gs, cw, align(j), style{}), pad)
			}
			out = append(out, trimSpaces(row))
		}
		if i == 0 {
			out = append(out, repeat("─", tableWidth(ws), style{role: Border}))
		}
	}
	return out
}

// paintDocument paints a HottyMarkdown's lines at (x, y), w wide and h
// tall: a rule across what its frames leave; each link in place's run of
// cells a hit, its box its first; each heading's rows its box, by
// view.HeadingID, for a program's own contents (which heading is in
// sight, scrolling one to the top); and, while it has the keyboard after
// a link in place went to one of its headings, that heading's rows its
// sight (reveal).
func (l *layout) paintDocument(cv *canvas, e *view.Element, x, y, w, h int) {
	r := l.r
	target := -1
	if r.focused(e.ID) {
		target = e.Target()
	}
	var sight *box
	heads := map[int]*box{}
	seen := map[string]bool{}
	for i, t := range l.lines(e, w) {
		if i >= h {
			break
		}
		gs := t.gs
		if t.rule {
			gs = concat(gs, repeat("─", w-width(gs), style{role: Border}))
		}
		cv.write(x, y+i, w, gs)
		at := x
		for j := 0; j < len(gs); {
			id := gs[j].hit
			n, k := 0, j
			for ; k < len(gs) && gs[k].hit == id; k++ {
				n += gs[k].width
			}
			if id != "" {
				r.hits = append(r.hits, hit{x: at, y: y + i, w: n, h: 1, id: id, opt: -1})
				if !seen[id] {
					seen[id] = true
					r.boxes[id] = box{at, y + i, n, 1}
				}
			}
			at, j = at+n, k
		}
		if t.head > 0 {
			if b := heads[t.head-1]; b == nil {
				heads[t.head-1] = &box{x, y + i, w, 1}
			} else {
				b.h = y + i - b.y + 1
			}
		}
		if target >= 0 && t.head == target+1 {
			if sight == nil {
				sight = &box{x, y + i, w, 1}
			} else {
				sight.h = y + i - sight.y + 1
			}
		}
	}
	for k, b := range heads {
		if k < len(e.Anchors) {
			r.boxes[view.HeadingID(e.ID, e.Anchors[k])] = *b
		}
	}
	if sight != nil {
		r.reveal[e.ID] = *sight
	}
}
