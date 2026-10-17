package cells

import (
	"encoding/base64"
	"strings"

	"github.com/neuroplastio/hotty-go"
	"github.com/neuroplastio/hotty-go/hottyedit"

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

// The edit actions (SPEC §10.2) are a keymap's, as on a host: the default
// keymap binds each to the host's edit shortcut, mod+c copy, mod+x cut,
// mod+v paste, mod+z undo, mod+shift+z redo (and ctrl+y where mod is
// ctrl), mod being the controller's Platform's, ctrl where no host says. A program overrides them with its keymap (SetKeys, a component's
// keys), and binds one to program to keep its key.

// CopyKey copies what is selected to the clipboard on a key the keymap
// binds to copy, and reports whether it did; the selection stays, as a
// page's does. The selection is that of the text control that has the
// keyboard, by its keymap, else a Text's, by the keymap of the element it
// starts in (SPEC §11). With nothing selected, or the key bound to
// program, the key is not the copy's: ctrl+c then interrupts or quits,
// as the program has it. Key does this first; a program calls it where a
// selection may be in a rendition that doesn't have the keyboard.
func (r *Rendition) CopyKey(key string) bool {
	if k, ok := hotty.ParseKey(key); ok {
		key = k
	}
	if e := r.editing(); e != nil {
		if text := r.fieldSelection(e); text != "" {
			if r.fieldKeymap(e).Lookup(key) != hotty.Copy {
				return false
			}
			r.clip = &Clip{Text: text}
			return true
		}
	}
	if r.sel.empty() || r.keymapIn(r.sel.id, false).Lookup(key) != hotty.Copy {
		return false
	}
	r.clip = &Clip{Text: r.selText()}
	return true
}

// keymapIn is the keymap of the element id, as a host resolves a field's
// (SPEC §10.2): the default one for the Platform, then the rendition's
// (SetKeys), then those of the components id is in, its own last.
func (r *Rendition) keymapIn(id string, multiline bool) *hotty.Keymap {
	return hotty.Resolve(r.c.Platform, multiline, append([]string{r.keys}, r.c.V.KeyChain(id)...)...)
}

// fieldKeymap is a text control's keymap.
func (r *Rendition) fieldKeymap(e *view.Element) *hotty.Keymap {
	return r.keymapIn(e.ID, isLongText(e))
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

// cut cuts what is selected in a text control, on a key its keymap binds
// to cut (editKey): to the clipboard, as copy puts it there, and out of
// the field. ok is false with nothing selected there, or in a password
// field, and the key is then the program's: on a page, cut does nothing
// to a Text's selection either.
func (r *Rendition) cut(e *view.Element, f *hottyedit.Field) (ok bool, err error) {
	text := r.fieldSelection(e)
	if text == "" {
		return false, nil
	}
	before := stateOf(f)
	f.Do(hotty.DeleteCharBackward)
	r.cursor[e.ID] = f.Caret
	r.record(e, before, stateOf(f), noRun)
	r.clip = &Clip{Text: text}
	return true, r.c.SetValue(e.ID, f.Value)
}

// OSC52Query asks the terminal for its clipboard: it answers with OSC 52
// and the clipboard's text in base64, which Bubble Tea gives a program as
// a tea.ClipboardMsg (tea.ReadClipboard sends this). A terminal may ask
// the user first, or not answer.
const OSC52Query = "\x1b]52;c;?\x1b\\"

// PasteWanted reports, once, that a key bound to paste asked for the
// clipboard, in a text control (editKey). The program reads it
// (OSC52Query) and gives what comes back to Paste. A host that sends
// edit commands (Command) brings the text with the command instead.
func (r *Rendition) PasteWanted() bool {
	w := r.wantPaste
	r.wantPaste = false
	return w
}

// Paste types text into the text control that has the keyboard, in place
// of what is selected there, as a page pastes: the clipboard a key bound
// to paste asked for (PasteWanted), a paste command's text (Command), or
// the terminal's own paste (a bracketed paste). A
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

// Command does an edit command that a host sent in place of its key (SPEC
// §10.6, which a program asks for with hotty.Edit). A command with a key
// is that key, the keymap deciding, except a paste in a text control whose
// keymap binds the key to paste: Command types the text the paste brought,
// where the key would have asked for the clipboard (PasteWanted). One from
// the terminal's menu, with no key, does what it names where it can.
// handled is false for the rest: the program gives c.Key to Key, as it
// gives any key, and takes it as its own when Key does not, as ctrl+c
// interrupts with nothing selected.
func (r *Rendition) Command(c hotty.Command) (handled bool, err error) {
	if c.Key == "" {
		return r.menuCommand(c)
	}
	if hotty.Action(c.Name) == hotty.Paste {
		key := c.Key
		if k, ok := hotty.ParseKey(key); ok {
			key = k
		}
		if e := r.editing(); e != nil && r.fieldKeymap(e).Lookup(key) == hotty.Paste {
			if !c.HasText {
				// Nothing on the clipboard, or a host that cannot read it.
				return true, nil
			}
			return r.Paste(c.Text)
		}
	}
	return false, nil
}

// menuCommand does a command from the terminal's menu, which names no
// key: in the text control that has the keyboard, else copy in a Text.
func (r *Rendition) menuCommand(c hotty.Command) (handled bool, err error) {
	e := r.editing()
	if e == nil {
		if hotty.Action(c.Name) != hotty.Copy || r.sel.empty() {
			return false, nil
		}
		r.clip = &Clip{Text: r.selText()}
		return true, nil
	}
	f := r.field(e)
	switch a := hotty.Action(c.Name); a {
	case hotty.Copy:
		text := r.fieldSelection(e)
		if text == "" {
			return false, nil
		}
		r.clip = &Clip{Text: text}
		return true, nil
	case hotty.Cut:
		return r.cut(e, f)
	case hotty.Paste:
		if !c.HasText {
			return true, nil
		}
		return r.Paste(c.Text)
	case hotty.Undo, hotty.Redo:
		return true, r.undo(e, a == hotty.Undo)
	case hotty.SelectAll:
		f.Do(hotty.SelectAll)
		r.cursor[e.ID] = f.Caret
		return true, nil
	}
	return false, nil
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
