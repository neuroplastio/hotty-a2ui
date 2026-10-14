package cells

import (
	"encoding/base64"
	"slices"
	"strings"

	"github.com/neuroplastio/hotty-go"

	"github.com/neuroplastio/hotty-a2ui/view"
)

// Selecting a Text's or a HottyMarkdown's text with the pointer, and
// copying it (profile §3.7, vault KIT-12c), as a page's text is selected:
// a press on a cell of one that no click acts on, then a drag, selects
// from there to the pointer, in the element's reading order, never across
// screen rows (a page's nav beside it stays out); the release puts it in
// the primary selection, and a copy key on the clipboard (Copied). Its
// cells are drawn in selection.

// textRow is a row of a Text as the last Draw painted it: where, the
// text each column copies as (a cluster at its first column, "" at a wide
// one's second and on its frame, glyph.deco), and how it goes on from the
// row before.
type textRow struct {
	x, y int
	cols []string
	cont cont
}

// textPos is a cell of a Text's rows: a row of them and a column of it.
type textPos struct{ row, col int }

func (p textPos) before(q textPos) bool {
	return p.row < q.row || p.row == q.row && p.col < q.col
}

// textSel is a selection in a Text: from its anchor, where the press was,
// to its head, where the pointer is; held while the button is down. It is
// a selection once the head has left the anchor's cell, or, made by a
// double or a triple click (whole), from the first: a word of one cell.
type textSel struct {
	id    string
	a, b  textPos
	held  bool
	whole bool
}

// empty reports whether a selection selects nothing: none, or a press that
// has not left its cell.
func (s *textSel) empty() bool { return s == nil || s.a == s.b && !s.whole }

// span is the selection's first and last cells, in reading order.
func (s *textSel) span() (lo, hi textPos) {
	if s.b.before(s.a) {
		return s.b, s.a
	}
	return s.a, s.b
}

// cols are the columns of a row a selection takes, from c0 to c1, and
// whether it takes any: none of its frame before its text (a quote's
// bar), as a page's selection leaves a quote's border.
func (s *textSel) cols(i int, row textRow) (c0, c1 int, ok bool) {
	if s.empty() {
		return 0, 0, false
	}
	lo, hi := s.span()
	if i < lo.row || i > hi.row {
		return 0, 0, false
	}
	first, _ := textExtent(row.cols)
	c0, c1 = first, len(row.cols)-1
	if i == lo.row {
		c0 = max(lo.col, first)
	}
	if i == hi.row {
		c1 = min(hi.col, c1)
	}
	return c0, c1, c0 <= c1
}

// paintText paints a row of a Text (or a HottyMarkdown) at (x, y), w wide,
// and keeps it for selecting: its cells in a selection are drawn in the
// selection's colour, reversed where that is not known, as a field's.
func (l *layout) paintText(cv *canvas, id string, x, y, w int, gs []glyph, c cont) {
	cv.write(x, y, w, gs)
	r := l.r
	var cols []string
	for _, g := range gs {
		if len(cols)+g.width > w {
			break
		}
		t := g.text
		if g.deco {
			t = ""
		}
		cols = append(cols, t)
		for range g.width - 1 {
			cols = append(cols, "")
		}
	}
	row := textRow{x: x, y: y, cols: cols, cont: c}
	if c0, c1, ok := r.sel.cols(len(r.texts[id]), row); ok && r.sel.id == id {
		cv.restyle(x+c0, y, c1-c0+1, 1, func(c *Cell) {
			c.Back, c.BackMix, c.BackAttr = selected.back, selected.backMix, selected.backAttr
		})
	}
	r.texts[id] = append(r.texts[id], row)
}

// textAt is the Text whose text is at a cell of the last frame, and the
// place there: on a column of a row of it, past its frame (not a quote's
// bar) and before the spaces it ends in.
func (r *Rendition) textAt(col, row int) (id string, at textPos, ok bool) {
	for id, rows := range r.texts {
		for i, t := range rows {
			if t.y != row || col < t.x {
				continue
			}
			c := col - t.x
			first, end := textExtent(t.cols)
			if c >= first && c < end {
				return id, textPos{i, c}, true
			}
		}
	}
	return "", textPos{}, false
}

// textExtent is the columns of a row with text: from its first that is
// not its frame to its last that is not a space, end exclusive.
func textExtent(cols []string) (first, end int) {
	first = len(cols)
	for i, t := range cols {
		if t != "" {
			first = i
			break
		}
	}
	end = len(cols)
	for end > first && (cols[end-1] == " " || cols[end-1] == "") {
		end--
	}
	return first, end
}

// textNear is the place in a Text's rows nearest a cell, as a drag that
// left them selects: above its first row its start, below its last its
// end; on a row of it, left of the row its start, right of it its end.
func (r *Rendition) textNear(id string, col, row int) textPos {
	rows := r.texts[id]
	if len(rows) == 0 {
		return textPos{}
	}
	if row < rows[0].y {
		return textPos{0, 0}
	}
	i := len(rows) - 1
	for k, t := range rows {
		if t.y >= row {
			i = k
			break
		}
	}
	t := rows[i]
	if row > t.y {
		return textPos{i, max(len(t.cols)-1, 0)}
	}
	return textPos{i, min(max(col-t.x, 0), max(len(t.cols)-1, 0))}
}

// arm starts a selection at a press on a cell of a Text that no click
// acts on (not a link in it, nor a field over it); any other press drops
// the selection there was. The second press of a run selects the word
// there instead, and a third its line (selectWords).
func (r *Rendition) arm(col, row int, h *hit) {
	r.sel = nil
	if h != nil && (h.field != nil || r.clickable(h)) {
		return
	}
	if id, at, ok := r.textAt(col, row); ok {
		r.sel = &textSel{id: id, a: at, b: at, held: true}
		if r.count >= 2 {
			r.selectWords(id, at, r.count)
		}
	}
}

// selectWords selects in a Text what the n-th press of a run (n ≥ 2) at
// at selects, as a page's text takes a double and a triple click, and
// copies it at once (Copied): the word there (wordAt, as a field's), or
// on a third press and after, its line, the rows a soft wrap made of one,
// a paragraph. The button held drags nothing.
func (r *Rendition) selectWords(id string, at textPos, n int) {
	rows := r.texts[id]
	first, last := at.row, at.row
	for first > 0 && rows[first].cont != newLine {
		first--
	}
	for last+1 < len(rows) && rows[last+1].cont != newLine {
		last++
	}
	// The line's clusters, as a copy joins its rows, and the cell of each:
	// a space a wrap dropped has none, and takes the next row's first.
	var line strings.Builder
	var cells []textPos
	p := -1
	for i := first; i <= last; i++ {
		t := rows[i]
		c0, end := textExtent(t.cols)
		if i > first && t.cont == wrapSpace {
			line.WriteString(" ")
			cells = append(cells, textPos{i, c0})
		}
		for c := c0; c < end; c++ {
			if t.cols[c] == "" {
				continue
			}
			if i == at.row && c <= at.col {
				p = len(cells)
			}
			line.WriteString(t.cols[c])
			cells = append(cells, textPos{i, c})
		}
	}
	if p < 0 || len(cells) == 0 {
		return
	}
	start, end := 0, len(cells)
	if n == 2 {
		start, end = wordAt(line.String(), p, false)
	}
	if end <= start {
		return
	}
	r.sel = &textSel{id: id, a: cells[start], b: cells[end-1], whole: true}
	r.clip = &Clip{Text: r.selText(), Primary: true}
}

// extend moves a held selection's head to the pointer, kept in its Text.
func (r *Rendition) extend(col, row int) bool {
	s := r.sel
	if s == nil || !s.held {
		return false
	}
	s.b = r.textNear(s.id, col, row)
	return true
}

// letGo ends a held selection: a press that did not leave its cell
// selects nothing; else its text goes to the primary selection (Copied).
func (r *Rendition) letGo() {
	s := r.sel
	if s == nil || !s.held {
		return
	}
	s.held = false
	if s.empty() {
		r.sel = nil
		return
	}
	r.clip = &Clip{Text: r.selText(), Primary: true}
}

// selText is a selection's text: each row's columns it takes, the spaces
// a row ends in dropped, a row that wraps the one before joined to it as
// the line it was (a space where the wrap dropped one), any other on a
// line of its own, so that a blank row between paragraphs is a blank line.
func (r *Rendition) selText() string {
	s := r.sel
	if s == nil {
		return ""
	}
	var b strings.Builder
	lo, hi := s.span()
	rows := r.texts[s.id]
	for i := lo.row; i <= hi.row && i < len(rows); i++ {
		t := rows[i]
		if i > lo.row {
			b.WriteString(t.cont.joiner())
		}
		if c0, c1, ok := s.cols(i, t); ok {
			b.WriteString(strings.TrimRight(strings.Join(t.cols[c0:c1+1], ""), " "))
		}
	}
	return b.String()
}

// Clip is text for the terminal to keep: on its clipboard, or, Primary,
// in its primary selection, which a middle click pastes.
type Clip struct {
	Text    string
	Primary bool
}

// OSC52 is the sequence that puts a clip where it goes (OSC 52's "c" or
// "p"). A terminal may ask the user first, or refuse; macOS has no primary
// selection, and its terminals drop one.
func (c Clip) OSC52() string {
	to := "c"
	if c.Primary {
		to = "p"
	}
	return "\x1b]52;" + to + ";" + base64.StdEncoding.EncodeToString([]byte(c.Text)) + "\x1b\\"
}

// Copied is the clip the last copy made, once: ok is false until the
// next. A program puts it where it goes (Clip.OSC52). As on a page, a
// selection made with the pointer goes to the primary selection when the
// button is let go (or at once, by a double or a triple click), and a copy
// key puts what is selected on the clipboard (CopyKey).
func (r *Rendition) Copied() (c Clip, ok bool) {
	if r.clip == nil {
		return Clip{}, false
	}
	c, r.clip = *r.clip, nil
	return c, true
}

// DefaultCopyKeys are a page's keys that copy: Control+c, and Command+c
// on a Mac, which SPEC §10.4 names Meta+c.
var DefaultCopyKeys = []string{"Control+c", "Meta+c"}

// DefaultCutKeys are a page's keys that cut: Control+x, and Command+x on
// a Mac (Meta+x).
var DefaultCutKeys = []string{"Control+x", "Meta+x"}

// DefaultPasteKeys are a page's keys that paste: Control+v, and Command+v
// on a Mac (Meta+v).
var DefaultPasteKeys = []string{"Control+v", "Meta+v"}

// SetCopyKeys sets the keys that copy what is selected (CopyKey), in SPEC
// §10.4's names. New starts with DefaultCopyKeys; with none, a key copies
// nothing and the program has them all.
func (r *Rendition) SetCopyKeys(keys ...string) { r.copyKeys = keyNames(keys) }

// SetCutKeys sets the keys that cut what is selected in a text control
// (Key), as SetCopyKeys does the copy keys. New starts with
// DefaultCutKeys.
func (r *Rendition) SetCutKeys(keys ...string) { r.cutKeys = keyNames(keys) }

// SetPasteKeys sets the keys that paste into a text control (Key,
// PasteWanted), as SetCopyKeys does the copy keys. New starts with
// DefaultPasteKeys.
func (r *Rendition) SetPasteKeys(keys ...string) { r.pasteKeys = keyNames(keys) }

// keyNames are keys named as hotty.ParseKey names them, so that a key
// pressed matches however the program spelled it.
func keyNames(keys []string) []string {
	out := make([]string, 0, len(keys))
	for _, k := range keys {
		if p, ok := hotty.ParseKey(k); ok {
			k = p
		}
		out = append(out, k)
	}
	return out
}

// CopyKey copies what is selected (Selection) to the clipboard, when key
// is one of the copy keys (SetCopyKeys), and reports whether it did. The
// selection stays, as a page's does. With nothing selected the key is not
// the copy's: Control+c then interrupts or quits, as the program has it.
// Key does this first; a program calls it where a selection may be in a
// rendition that doesn't have the keyboard.
func (r *Rendition) CopyKey(key string) bool {
	if k, ok := hotty.ParseKey(key); ok {
		key = k
	}
	if !slices.Contains(r.copyKeys, key) {
		return false
	}
	text := r.Selection()
	if text == "" {
		return false
	}
	r.clip = &Clip{Text: text}
	return true
}

// Selection is the text selected now, "" for none: in the text control
// that has the keyboard, else in a Text. A password field's selection
// copies nothing, as on a page.
func (r *Rendition) Selection() string {
	if e := r.editing(); e != nil {
		if text := r.fieldSelection(e); text != "" {
			return text
		}
	}
	if r.sel.empty() {
		return ""
	}
	return r.selText()
}

// cutKey cuts what is selected in the text control that has the
// keyboard, on a cut key (SetCutKeys): it goes to the clipboard, as a
// copy key puts it there, and out of the field. ok is false when nothing
// is selected there, and the key is then the program's: on a page, a cut
// key does nothing to a Text's selection, or to a password field's.
func (r *Rendition) cutKey(key string) (ok bool, err error) {
	e := r.editing()
	if e == nil || !slices.Contains(r.cutKeys, key) {
		return false, nil
	}
	text := r.fieldSelection(e)
	if text == "" {
		return false, nil
	}
	f := r.field(e)
	before := stateOf(f)
	f.Do(hotty.DeleteCharBackward)
	r.cursor[e.ID] = f.Caret
	r.record(e, before, stateOf(f), noRun)
	r.clip = &Clip{Text: text}
	return true, r.c.SetValue(e.ID, f.Value)
}

// pasteKey asks for the clipboard on a paste key (SetPasteKeys) while a
// text control has the keyboard (PasteWanted); elsewhere the key is the
// program's.
func (r *Rendition) pasteKey(key string) bool {
	if r.editing() == nil || !slices.Contains(r.pasteKeys, key) {
		return false
	}
	r.wantPaste = true
	return true
}

// OSC52Query asks the terminal for its clipboard: it answers with OSC 52
// and the clipboard's text in base64, which Bubble Tea gives a program as
// a tea.ClipboardMsg (tea.ReadClipboard sends this). A terminal may ask
// the user first, or not answer.
const OSC52Query = "\x1b]52;c;?\x1b\\"

// PasteWanted reports, once, that a paste key asked for the clipboard.
// The program reads it (OSC52Query) and gives what comes back to Paste.
func (r *Rendition) PasteWanted() bool {
	w := r.wantPaste
	r.wantPaste = false
	return w
}

// Paste types text into the text control that has the keyboard, in place
// of what is selected there, as a page pastes: the clipboard a paste key
// asked for (PasteWanted), or the terminal's own paste (bracketed paste:
// Control+Shift+v, or Command+v on a Mac, by its default keys). A
// one-line field takes it without its line breaks, and a number field
// only what a number has, as each takes typing. handled is false when no
// text control has the keyboard: the paste is the program's.
func (r *Rendition) Paste(text string) (handled bool, err error) {
	e := r.editing()
	if e == nil {
		return false, nil
	}
	r.wake()
	r.hush()
	text = strings.NewReplacer("\r\n", "\n", "\r", "\n").Replace(text)
	if !isLongText(e) {
		text = strings.ReplaceAll(text, "\n", "")
	}
	if e.Kind == view.TextField && e.Variant == "number" {
		text = strings.Map(func(c rune) rune {
			if strings.ContainsRune(numberChars, c) {
				return c
			}
			return -1
		}, text)
	}
	f := r.field(e)
	before := stateOf(f)
	if !f.Type(text) {
		return true, nil
	}
	r.cursor[e.ID] = f.Caret
	r.record(e, before, stateOf(f), noRun)
	return true, r.c.SetValue(e.ID, f.Value)
}

// editing is the text control that has the keyboard; nil for none.
func (r *Rendition) editing() *view.Element {
	if e := r.c.V.Find(r.c.St.Focus); r.c.St.Keyboard && isTextControl(e) {
		return e
	}
	return nil
}

// Deselect drops the selection in a Text, as a press anywhere else does.
// A program with several renditions calls it on the others when one is
// pressed: a page has one selection.
func (r *Rendition) Deselect() { r.sel = nil }

// fieldSelection is the text selected in a text control, as editKey and
// the pointer left it; "" for none, and in a password field.
func (r *Rendition) fieldSelection(e *view.Element) string {
	if e.Variant == "obscured" {
		return ""
	}
	v, _ := e.Value.(string)
	cl := clusters(v)
	f := r.fields[e.ID]
	if f == nil || f.Value != v || f.Caret != r.cursorOf(e.ID, len(cl)) {
		return ""
	}
	s0, s1 := f.Selection()
	return strings.Join(cl[s0:s1], "")
}

// primary puts what a text control's pointer selected in the primary
// selection, as a page does with an input's.
func (r *Rendition) primary(e *view.Element) {
	if text := r.fieldSelection(e); text != "" {
		r.clip = &Clip{Text: text, Primary: true}
	}
}
