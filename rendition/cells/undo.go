package cells

import (
	"slices"

	"github.com/neuroplastio/hotty-go"
	"github.com/neuroplastio/hotty-go/hottyedit"

	"github.com/neuroplastio/hotty-a2ui/view"
)

// Undoing and redoing a text control's edits, as a page does (profile
// §3.7): each text control has a history of its own. A run of typing is
// one step, and so is a run of deleting a character at a time, while
// nothing moves the caret between; a cut, a paste and any other edit is a
// step of its own. An edit after an undo drops what could be redone, and
// a value set from outside (by the agent, or by another control on the
// same data) drops the history, as a script setting an input's value
// does.

// DefaultUndoKeys are a page's keys that undo: Control+z, and Command+z
// on a Mac (Meta+z).
var DefaultUndoKeys = []string{"Control+z", "Meta+z"}

// DefaultRedoKeys are a page's keys that redo: Control+Shift+z and
// Control+y, and Command+Shift+z on a Mac.
var DefaultRedoKeys = []string{"Shift+Control+z", "Control+y", "Shift+Meta+z"}

// SetUndoKeys sets the keys that undo a text control's last edit (Key),
// as SetCopyKeys does the copy keys. New starts with DefaultUndoKeys.
func (r *Rendition) SetUndoKeys(keys ...string) { r.undoKeys = keyNames(keys) }

// SetRedoKeys sets the keys that redo what an undo key undid, as
// SetCopyKeys does the copy keys. New starts with DefaultRedoKeys.
func (r *Rendition) SetRedoKeys(keys ...string) { r.redoKeys = keyNames(keys) }

// maxSteps bounds a text control's history: the oldest step goes.
const maxSteps = 100

// fieldState is a text control's value and selection.
type fieldState struct {
	value         string
	anchor, caret int
}

func stateOf(f *hottyedit.Field) fieldState {
	a, _ := f.Anchor()
	return fieldState{f.Value, a, f.Caret}
}

// step is one edit: the text control before it and after it.
type step struct{ before, after fieldState }

// editRun is what kind of run an edit can join: typing, or deleting a
// character backward or forward; noRun joins nothing.
type editRun int

const (
	noRun editRun = iota
	typing
	deletingBack
	deletingForward
)

func runOf(a hotty.Action) editRun {
	switch a {
	case hotty.Insert:
		return typing
	case hotty.DeleteCharBackward:
		return deletingBack
	case hotty.DeleteCharForward:
		return deletingForward
	}
	return noRun
}

// history is a text control's steps: those to undo, the last on top, and
// those an undo undid, to redo.
type history struct {
	undo, redo []step
	// run is the last step's, while the next edit may join it.
	run editRun
	// now is the value the history left the control with: another means
	// something else set it.
	now string
}

// historyOf is a text control's history while its value v is the one the
// history left; else a new one.
func (r *Rendition) historyOf(e *view.Element, v string) *history {
	h := r.history[e.ID]
	if h == nil || h.now != v {
		h = &history{now: v}
		r.history[e.ID] = h
	}
	return h
}

// record adds an edit to a text control's history. It joins the last step
// when the two are one run and nothing moved the caret between them.
func (r *Rendition) record(e *view.Element, before, after fieldState, run editRun) {
	h := r.historyOf(e, before.value)
	h.redo, h.now = nil, after.value
	if n := len(h.undo); run != noRun && run == h.run && n > 0 && h.undo[n-1].after == before {
		h.undo[n-1].after = after
		return
	}
	h.run = run
	h.undo = append(h.undo, step{before, after})
	if len(h.undo) > maxSteps {
		h.undo = slices.Delete(h.undo, 0, 1)
	}
}

// undoKey undoes the last step of the text control that has the
// keyboard, on an undo key, or redoes the last it undid, on a redo key,
// value and selection both. The control takes the key with nothing to
// undo or redo too, as a page's does; elsewhere the key is the program's.
func (r *Rendition) undoKey(key string) (ok bool, err error) {
	e := r.editing()
	if e == nil {
		return false, nil
	}
	undo := slices.Contains(r.undoKeys, key)
	if !undo && !slices.Contains(r.redoKeys, key) {
		return false, nil
	}
	v, _ := e.Value.(string)
	h := r.historyOf(e, v)
	from, to := &h.undo, &h.redo
	if !undo {
		from, to = to, from
	}
	n := len(*from)
	if n == 0 {
		return true, nil
	}
	s := (*from)[n-1]
	*from, *to = (*from)[:n-1], append(*to, s)
	st := s.before
	if !undo {
		st = s.after
	}
	h.run, h.now = noRun, st.value
	f := r.field(e)
	f.Value = st.value
	f.Select(st.anchor, st.caret)
	r.cursor[e.ID] = f.Caret
	return true, r.c.SetValue(e.ID, f.Value)
}
