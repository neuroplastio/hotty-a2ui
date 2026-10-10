package cells

import (
	"strings"

	"github.com/neuroplastio/hotty-go"
	"github.com/neuroplastio/hotty-go/hottyedit"

	"github.com/neuroplastio/hotty-a2ui/view"
)

// Key is a key the user pressed, named as SPEC §10.4 has it: "a", "A",
// "Space", "Enter", "Shift+Tab", "Control+s" (profile §5). While the
// surface has the keyboard, the focused element takes the keys it uses: a
// text control those its keymap names, and the characters it types (SPEC
// §10.2, editKey); another kind those of SPEC §10.2's table, unmodified or
// with Shift only. Tab and Shift+Tab move focus. Any other key, Escape
// among them, goes to the surface's Shortcuts; without the keyboard only
// they are tried. Then ? switches a HottyKeyHints' views, and, as
// rendition/html does, Escape closes an open Modal. handled reports whether the surface took the key: when it did
// not, the key is the program's.
func (r *Rendition) Key(key string) (handled bool, err error) {
	if k, ok := hotty.ParseKey(key); ok {
		key = k
	}
	mods, name := splitKey(key)
	c := r.c
	shiftOnly := len(mods) == 0 || len(mods) == 1 && mods[0] == "Shift"
	if key == "Escape" && r.cancelDrag() {
		return true, nil
	}
	if c.St.Keyboard {
		switch e := c.V.Find(c.St.Focus); {
		case isTextControl(e):
			if ok, err := r.editKey(e, key); ok {
				return true, err
			}
		case e != nil && e.Kind == view.ScrollView:
			if ok, err := r.scrollKey(e, key); ok {
				return true, err
			}
		case e != nil && shiftOnly:
			if ok, err := r.elementKey(e, name); ok {
				return true, err
			}
		}
		// Alt with the arrows moves the item the keyboard is on (§6.21).
		if ok, err := c.MoveKey(c.St.Focus, key); ok {
			return true, err
		}
		if name == "Tab" && shiftOnly {
			r.list = ""
			if c.FocusNext(len(mods) == 1) {
				delete(r.cursor, c.St.Focus)
			}
			return true, nil
		}
	}
	if ok, err := c.Shortcut(key); ok {
		return true, err
	}
	// ? that nothing took switches the key hints' views (HottyKeyHints).
	if name == "?" && shiftOnly && c.ToggleHints() {
		return true, nil
	}
	if key == "Escape" && c.St.Modal != "" {
		r.list = ""
		c.CloseModal()
		return true, nil
	}
	return false, nil
}

// splitKey splits a key's name into its modifiers and its value.
func splitKey(key string) (mods []string, name string) {
	name = key
	if key != "+" {
		if strings.HasSuffix(key, "++") {
			mods, name = strings.Split(strings.TrimSuffix(key, "++"), "+"), "+"
		} else if i := strings.LastIndex(key, "+"); i > 0 {
			mods, name = strings.Split(key[:i], "+"), key[i+1:]
		}
	}
	return mods, name
}

// printable reports whether a key value is a character: one grapheme
// cluster, not a control.
func printable(name string) bool {
	cl := clusters(name)
	return len(cl) == 1 && !isControl(cl[0]) && !isBreak(cl[0])
}

// elementKey gives a key to the focused element that is not a text
// control; ok reports whether it took it (SPEC §10.2's table).
func (r *Rendition) elementKey(e *view.Element, name string) (ok bool, err error) {
	if isSelect(e) {
		return r.selectKey(e, name)
	}
	if e.Kind == view.Table {
		return r.c.TableKey(e.ID, name)
	}
	if e.Kind == view.DiffView {
		return r.c.DiffKey(e.ID, name)
	}
	if e.Kind == view.RichList {
		return r.c.ListKey(e.ID, name)
	}
	if e.Kind == view.Tree {
		return r.c.TreeKey(e.ID, name)
	}
	// A HottyRangeSlider's knob steps as a Slider does, and stops where it
	// meets the other.
	if e.Kind == view.Slider || e.Kind == view.Knob {
		switch name {
		case "ArrowLeft", "ArrowDown":
			return true, r.c.StepSlider(e.ID, -1, "")
		case "ArrowRight", "ArrowUp":
			return true, r.c.StepSlider(e.ID, 1, "")
		case "Home", "End":
			return true, r.c.StepSlider(e.ID, 0, name)
		}
		return false, nil
	}
	if r.isBox(e) {
		// As a host takes them (SPEC §10.2, a checkbox): Space checks it,
		// and Enter submits its HottyForm, or does nothing outside one.
		switch name {
		case "Space":
			return true, r.c.Activate(e.ID)
		case "Enter":
			return true, r.c.Enter(e.ID)
		}
		return false, nil
	}
	switch e.Kind {
	// A HottySwitch is a button on a host: Enter flips it too, in a
	// HottyForm as anywhere.
	case view.Button, view.Tab, view.Option, view.Switch:
	case view.Media:
		if e.URL == "" {
			return false, nil
		}
	case view.Modal:
		if !e.Clickable {
			return false, nil
		}
	default:
		return false, nil
	}
	if name == "Space" || name == "Enter" {
		return true, r.c.Activate(e.ID)
	}
	return false, nil
}

// editKey edits a text control as a host edits a text field (SPEC
// §10.2), with hottyedit: by its keymap, the surface's (SetKeys) and then
// those of the components it is in, its own last (view.KeyChain), and
// typing the characters it does not bind. A number field types only what
// a number has. ok is false for a key the keymap leaves to the program.
// The data model is written on every change.
func (r *Rendition) editKey(e *view.Element, key string) (ok bool, err error) {
	km := hotty.Resolve(isLongText(e), append([]string{r.keys}, r.c.V.KeyChain(e.ID)...)...)
	f := r.field(e)
	f.Rows = r.pageRows(e)
	switch km.Lookup(key) {
	case "":
		return false, nil
	case hotty.Insert:
		ch := key
		if strings.HasSuffix(key, "Space") {
			ch = " "
		}
		if e.Kind == view.TextField && e.Variant == "number" && !strings.ContainsAny(ch, "0123456789.,-+eE") {
			return true, nil
		}
	case hotty.Submit:
		return true, r.c.Enter(e.ID)
	}
	_, changed := f.Key(km, key)
	r.cursor[e.ID] = f.Caret
	if changed {
		return true, r.c.SetValue(e.ID, f.Value)
	}
	return true, nil
}

// field is the hottyedit.Field a text control edits by: the one kept while
// nothing else moved its caret or changed its value, so that a run of row
// moves keeps its place along the row and a selection stays; else a new
// one, with nothing selected.
func (r *Rendition) field(e *view.Element) *hottyedit.Field {
	v, _ := e.Value.(string)
	long := isLongText(e)
	pos := r.cursorOf(e.ID, len(clusters(v)))
	f := r.fields[e.ID]
	if f == nil || f.Value != v || f.Caret != pos || f.Multiline != long {
		f = &hottyedit.Field{Value: v, Caret: pos, Multiline: long, Password: e.Variant == "obscured"}
		r.fields[e.ID] = f
	}
	return f
}

func (r *Rendition) pageRows(e *view.Element) int {
	if n := r.rows[e.ID]; n > 0 {
		return n
	}
	return fieldRows(e)
}

// selectKey works a select. Closed, the arrows, Home, End, Page Up, Page
// Down and a letter change its value; Space or Enter opens its list, where
// they move the highlight instead, and Space or Enter picks it.
func (r *Rendition) selectKey(e *view.Element, name string) (bool, error) {
	n := len(e.Options)
	open := r.listOpen(e)
	cur := picked(e)
	if open {
		cur = r.hi
	}
	move := func(i int) error {
		if n == 0 {
			return nil
		}
		i = min(max(i, 0), n-1)
		if open {
			r.hi = i
			return nil
		}
		return r.c.SetValue(e.ID, e.Options[i].Value)
	}
	switch name {
	case "Space", "Enter":
		if open && n > 0 {
			r.list = ""
			return true, r.c.SetValue(e.ID, e.Options[r.hi].Value)
		}
		r.toggleList(e)
		return true, nil
	case "ArrowUp":
		return true, move(cur - 1)
	case "ArrowDown":
		return true, move(cur + 1)
	case "Home", "PageUp":
		return true, move(0)
	case "End", "PageDown":
		return true, move(n - 1)
	}
	if !printable(name) {
		return false, nil
	}
	for i := 1; i <= n; i++ {
		j := (max(cur, -1) + i + n) % n
		if hasPrefixFold(e.Options[j].Label, name) {
			return true, move(j)
		}
	}
	return true, nil
}

// isBox reports whether an element is a checkbox on a host: a CheckBox,
// or an option of a Choice shown as boxes (not chips, which are buttons).
func (r *Rendition) isBox(e *view.Element) bool {
	switch e.Kind {
	case view.CheckBox:
		return true
	case view.Option:
		anc := r.c.Ancestors(e.ID)
		return len(anc) > 0 && anc[len(anc)-1].Variant != "chips"
	}
	return false
}

// hasPrefixFold reports whether s starts with prefix, case aside.
func hasPrefixFold(s, prefix string) bool {
	return strings.HasPrefix(strings.ToLower(s), strings.ToLower(prefix))
}
