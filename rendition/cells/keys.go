package cells

import (
	"strings"
	"unicode/utf8"

	"github.com/neuroplastio/hotty-a2ui/view"
)

// Key is a key the user pressed, as a W3C key value after its modifiers,
// joined by "+": "a", "A", " " or "Space", "Enter", "Shift+Tab",
// "Control+s" (profile §5, SPEC §10.2). While the surface has the
// keyboard, the focused element takes the keys its kind uses, unmodified
// or with Shift only; Tab and Shift+Tab move focus. Any other key, Escape
// among them, goes to the surface's Shortcuts; without the keyboard only
// they are tried. Then, as rendition/html does, Escape closes an open
// Modal. handled reports whether the surface took the key: when it did
// not, the key is the program's.
func (r *Rendition) Key(key string) (handled bool, err error) {
	mods, name := splitKey(key)
	c := r.c
	shiftOnly := len(mods) == 0 || len(mods) == 1 && mods[0] == "Shift"
	if c.St.Keyboard {
		e := c.V.Find(c.St.Focus)
		if e != nil && shiftOnly {
			if ok, err := r.elementKey(e, name); ok {
				return true, err
			}
		}
		if name == "Tab" && shiftOnly {
			r.list = ""
			if c.FocusNext(len(mods) == 1) {
				delete(r.cursor, c.St.Focus)
			}
			return true, nil
		}
		if isTextControl(e) && shiftOnly && printable(name) {
			return false, nil
		}
	}
	if ok, err := c.Shortcut(strings.Join(append(mods, name), "+")); ok {
		return true, err
	}
	if key == "Escape" && c.St.Modal != "" {
		r.list = ""
		c.CloseModal()
		return true, nil
	}
	return false, nil
}

// splitKey splits a key into its modifiers and its value; "Space" is " ",
// and a letter with Shift is a capital.
func splitKey(key string) (mods []string, name string) {
	name = key
	if key != "+" {
		if strings.HasSuffix(key, "++") {
			mods, name = strings.Split(strings.TrimSuffix(key, "++"), "+"), "+"
		} else if i := strings.LastIndex(key, "+"); i > 0 {
			mods, name = strings.Split(key[:i], "+"), key[i+1:]
		}
	}
	if name == "Space" {
		name = " "
	}
	if len(mods) == 1 && mods[0] == "Shift" && utf8.RuneCountInString(name) == 1 {
		name = strings.ToUpper(name)
	}
	return mods, name
}

// printable reports whether a key value is a character: one grapheme
// cluster, not a control.
func printable(name string) bool {
	cl := clusters(name)
	return len(cl) == 1 && !isControl(cl[0]) && !isBreak(cl[0])
}

// elementKey gives a key to the focused element; ok reports whether it
// took it (SPEC §10.2's table).
func (r *Rendition) elementKey(e *view.Element, name string) (ok bool, err error) {
	switch {
	case isTextControl(e):
		return r.editKey(e, name)
	case isSelect(e):
		return r.selectKey(e, name)
	}
	switch e.Kind {
	case view.Button, view.Tab, view.Option, view.CheckBox:
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
	if name == " " || name == "Enter" {
		return true, r.c.Activate(e.ID)
	}
	return false, nil
}

// editKey edits a text control: the data model is written on every
// change.
func (r *Rendition) editKey(e *view.Element, name string) (bool, error) {
	v, _ := e.Value.(string)
	cl := clusters(v)
	pos := r.cursorOf(e.ID, len(cl))
	long := isLongText(e)
	lines := splitClusters(cl)
	li, ci := locate(lines, pos)
	if !long {
		lines, li, ci = [][]string{cl}, 0, pos
	}
	write := func(before, insert, after string) error {
		r.cursor[e.ID] = len(clusters(before + insert))
		return r.c.SetValue(e.ID, before+insert+after)
	}
	join := func(cl []string) string { return strings.Join(cl, "") }
	switch name {
	case "Backspace":
		if pos > 0 {
			return true, write(join(cl[:pos-1]), "", join(cl[pos:]))
		}
	case "Delete":
		if pos < len(cl) {
			return true, write(join(cl[:pos]), "", join(cl[pos+1:]))
		}
	case "ArrowLeft":
		r.cursor[e.ID] = max(pos-1, 0)
	case "ArrowRight":
		r.cursor[e.ID] = min(pos+1, len(cl))
	case "Home":
		r.cursor[e.ID] = offset(lines, li, 0)
	case "End":
		r.cursor[e.ID] = offset(lines, li, len(lines[li]))
	case "ArrowUp", "ArrowDown", "PageUp", "PageDown":
		if !long {
			return false, nil
		}
		step := map[string]int{"ArrowUp": -1, "ArrowDown": 1, "PageUp": -r.pageRows(e), "PageDown": r.pageRows(e)}[name]
		t := li + step
		switch {
		case t < 0:
			r.cursor[e.ID] = 0
		case t >= len(lines):
			r.cursor[e.ID] = len(cl)
		default:
			col := colOf(lines[li], ci, false)
			r.cursor[e.ID] = offset(lines, t, indexAt(lines[t], col, false))
		}
	case "Enter":
		if long {
			return true, write(join(cl[:pos]), "\n", join(cl[pos:]))
		}
		r.cursor[e.ID] = pos
		return true, r.c.Enter(e.ID)
	default:
		if !printable(name) {
			return false, nil
		}
		if e.Kind == view.TextField && e.Variant == "number" && !strings.ContainsAny(name, "0123456789.,-+eE") {
			return true, nil
		}
		return true, write(join(cl[:pos]), name, join(cl[pos:]))
	}
	return true, nil
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
	case " ", "Enter":
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

// hasPrefixFold reports whether s starts with prefix, case aside.
func hasPrefixFold(s, prefix string) bool {
	return strings.HasPrefix(strings.ToLower(s), strings.ToLower(prefix))
}
