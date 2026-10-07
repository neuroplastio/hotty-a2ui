package a2ui

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

// MaxListIndex is the largest list index a write may use. A write past a
// list's end pads it with nulls, so the bound keeps a payload from making
// the renderer allocate without limit.
const MaxListIndex = 10_000

// DataError is a data model read or write that cannot be done: through a
// value that is not a container, with a segment a list cannot take, or with
// a forbidden segment.
type DataError struct{ Msg string }

func (e *DataError) Error() string { return "a2ui: data model: " + e.Msg }

// forbidden are the segments that would reach an object's prototype in
// JavaScript. Every A2UI engine refuses them, so that a payload does not
// work in one and not another.
var forbidden = map[string]bool{"__proto__": true, "constructor": true, "prototype": true}

// DataModel is a surface's data: one JSON value (maps, slices, float64,
// string, bool, nil), read and written by JSON Pointer, with observers on
// paths. A write copies the containers along its path rather than changing
// them, so a value read before it stays as it was.
type DataModel struct {
	root     any
	watchers []*watcher
}

type watcher struct {
	path Pointer
	fn   func()
}

// NewDataModel returns a model holding initial, or an empty object when
// initial is nil. The model keeps its own copy.
func NewDataModel(initial any) *DataModel {
	if initial == nil {
		initial = map[string]any{}
	}
	return &DataModel{root: Clone(initial)}
}

// Root is the whole model. It must not be changed.
func (m *DataModel) Root() any { return m.root }

// Get reads the value at path: ok is false where there is none.
func (m *DataModel) Get(path string) (v any, ok bool, err error) {
	p, err := dataPath(path)
	if err != nil {
		return nil, false, err
	}
	v, ok = lookup(m.root, p)
	return v, ok, nil
}

// Value is Get without the error: nil where there is nothing to read.
func (m *DataModel) Value(path string) any {
	v, _, _ := m.Get(path)
	return v
}

// Set writes v at path, creating the containers on the way: a list where
// the next segment is an index, an object otherwise. Observers whose value
// changed are notified.
func (m *DataModel) Set(path string, v any) error {
	p, err := dataPath(path)
	if err != nil {
		return err
	}
	root, err := set(m.root, p, Clone(v), path)
	if err != nil {
		return err
	}
	m.commit(p, root)
	return nil
}

// Delete removes the value at path from its container: a key from an
// object, an element from a list. Deleting what is not there does nothing.
func (m *DataModel) Delete(path string) error {
	p, err := dataPath(path)
	if err != nil {
		return err
	}
	if len(p) == 0 {
		m.commit(p, map[string]any{})
		return nil
	}
	if _, ok := lookup(m.root, p); !ok {
		return nil
	}
	m.commit(p, del(m.root, p))
	return nil
}

// Watch calls fn after every write that changes the value at path, and
// returns the function that stops it.
func (m *DataModel) Watch(path string, fn func()) (stop func()) {
	p, err := dataPath(path)
	if err != nil {
		return func() {}
	}
	w := &watcher{path: p, fn: fn}
	m.watchers = append(m.watchers, w)
	return func() {
		for i, x := range m.watchers {
			if x == w {
				m.watchers = append(m.watchers[:i:i], m.watchers[i+1:]...)
				return
			}
		}
	}
}

// Dispose stops every observer.
func (m *DataModel) Dispose() { m.watchers = nil }

// commit installs a new root after a write at p, and notifies the
// observers whose value changed: those at p and above it when the value at
// p changed, and those below it whose own value did.
func (m *DataModel) commit(p Pointer, root any) {
	old := m.root
	m.root = root
	before, hadBefore := lookup(old, p)
	after, hasAfter := lookup(root, p)
	if hadBefore == hasAfter && reflect.DeepEqual(before, after) {
		return
	}
	var fire []*watcher
	for _, w := range m.watchers {
		switch {
		case p.HasPrefix(w.path):
			fire = append(fire, w)
		case w.path.HasPrefix(p):
			b, okb := lookup(old, w.path)
			a, oka := lookup(root, w.path)
			if okb != oka || !reflect.DeepEqual(a, b) {
				fire = append(fire, w)
			}
		}
	}
	for _, w := range fire {
		w.fn()
	}
}

func dataPath(path string) (Pointer, error) {
	p, err := ParsePointer(path)
	if err != nil {
		return nil, &DataError{Msg: err.Error()}
	}
	for _, t := range p {
		if forbidden[t] {
			return nil, &DataError{Msg: fmt.Sprintf("Forbidden path segment '%s' in '%s'", t, path)}
		}
	}
	return p, nil
}

// listIndex reads a segment as a list index: digits, no leading zero.
func listIndex(t string) (int, bool) {
	if t == "" || len(t) > 1 && t[0] == '0' {
		return 0, false
	}
	for _, c := range t {
		if c < '0' || c > '9' {
			return 0, false
		}
	}
	n, err := strconv.Atoi(t)
	return n, err == nil
}

func lookup(v any, p Pointer) (any, bool) {
	for _, t := range p {
		switch c := v.(type) {
		case map[string]any:
			x, ok := c[t]
			if !ok {
				return nil, false
			}
			v = x
		case []any:
			i, ok := listIndex(t)
			if !ok || i >= len(c) {
				return nil, false
			}
			v = c[i]
		default:
			return nil, false
		}
	}
	// null is no value: the nulls a sparse write pads a list with read as
	// absent, as JavaScript's holes do.
	return v, v != nil
}

// set returns a copy of v with x at p, copying only the containers on p's
// way.
func set(v any, p Pointer, x any, path string) (any, error) {
	if len(p) == 0 {
		return x, nil
	}
	t, rest := p[0], p[1:]
	switch c := v.(type) {
	case nil:
		// Vivify: a list when the segment is an index, else an object.
		if _, ok := listIndex(t); ok {
			return set([]any{}, p, x, path)
		}
		return set(map[string]any{}, p, x, path)
	case map[string]any:
		child, err := set(c[t], rest, x, path)
		if err != nil {
			return nil, err
		}
		n := make(map[string]any, len(c)+1)
		for k, e := range c {
			n[k] = e
		}
		n[t] = child
		return n, nil
	case []any:
		i, ok := listIndex(t)
		if !ok {
			return nil, &DataError{Msg: fmt.Sprintf("Cannot set path '%s': non-numeric segment '%s' in a list", path, t)}
		}
		if i > MaxListIndex {
			return nil, &DataError{Msg: fmt.Sprintf("Cannot set path '%s': index %d is over %d", path, i, MaxListIndex)}
		}
		var cur any
		if i < len(c) {
			cur = c[i]
		}
		child, err := set(cur, rest, x, path)
		if err != nil {
			return nil, err
		}
		n := make([]any, max(len(c), i+1))
		copy(n, c)
		n[i] = child
		return n, nil
	default:
		return nil, &DataError{Msg: fmt.Sprintf("Cannot set path '%s': '%s' is not an object or a list", path, t)}
	}
}

// del returns a copy of v without the value at p, which exists.
func del(v any, p Pointer) any {
	t := p[0]
	switch c := v.(type) {
	case map[string]any:
		n := make(map[string]any, len(c))
		for k, e := range c {
			n[k] = e
		}
		if len(p) == 1 {
			delete(n, t)
		} else {
			n[t] = del(c[t], p[1:])
		}
		return n
	case []any:
		i, _ := listIndex(t)
		if len(p) == 1 {
			n := make([]any, 0, len(c)-1)
			return append(append(n, c[:i]...), c[i+1:]...)
		}
		n := make([]any, len(c))
		copy(n, c)
		n[i] = del(c[i], p[1:])
		return n
	}
	return v
}

// Clone deep-copies a JSON value, normalising numbers to float64 so that
// equal values compare equal whatever produced them.
func Clone(v any) any {
	switch c := v.(type) {
	case map[string]any:
		n := make(map[string]any, len(c))
		for k, e := range c {
			n[k] = Clone(e)
		}
		return n
	case []any:
		n := make([]any, len(c))
		for i, e := range c {
			n[i] = Clone(e)
		}
		return n
	case int:
		return float64(c)
	case int64:
		return float64(c)
	case float32:
		return float64(c)
	}
	return v
}

// IsDataError reports whether err is a DataError.
func IsDataError(err error) bool {
	var d *DataError
	return errors.As(err, &d)
}
