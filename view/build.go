package view

import (
	"strconv"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
)

// State is the renderer's own state for a surface: what A2UI leaves to the
// renderer. It belongs to the surface, not to a rendition, so a story
// keeps it when the rendition changes.
type State struct {
	// Focus is the id of the element with the keyboard, or of the one
	// that last had it; Keyboard says whether the surface has it now.
	Focus    string
	Keyboard bool
	// Tabs are the tab each Tabs shows, by its id; 0 when absent.
	Tabs map[string]int
	// Modal is the id of the open Modal: one at a time.
	Modal string
	// Local are the values of controls whose value is not bound to the
	// data model, by id: what the user made them.
	Local map[string]any
	// Touched are the controls the user has changed; Submitted the forms
	// (and buttons) the user tried while their checks failed. A control's
	// error shows once either is so.
	Touched   map[string]bool
	Submitted map[string]bool
	// Scroll is the first row each Table with a height shows, the first
	// item of the page each HottyList with a height shows, and the first
	// row each HottyScrollView shows, by id.
	Scroll map[string]int
	// Left is the first column each HottyScrollView of lines that do not
	// wrap shows; Tail whether it follows the tail, once the user or the
	// agent moved it (its follow prop until then).
	Left map[string]int
	Tail map[string]bool
	// Query is each HottyList's filter, by its id.
	Query map[string]Query
	// FullHints: the surface's HottyKeyHints show their full view (?).
	FullHints bool
}

// NewState is a surface's state before the user does anything.
func NewState() *State {
	return &State{Tabs: map[string]int{}, Local: map[string]any{}, Touched: map[string]bool{}, Submitted: map[string]bool{}, Scroll: map[string]int{}, Query: map[string]Query{},
		Left: map[string]int{}, Tail: map[string]bool{}}
}

// Mapper makes an element of a node, children included; nil when the
// node draws nothing (a Shortcut).
type Mapper func(b *Builder, n *a2ui.Node) *Element

// mappers are the components the renderer knows, by catalog and type.
var mappers = map[string]map[string]Mapper{}

// Register says how to draw a catalog's component type.
func Register(catalog, typ string, m Mapper) {
	if mappers[catalog] == nil {
		mappers[catalog] = map[string]Mapper{}
	}
	mappers[catalog][typ] = m
}

// Builder makes a surface's view from its resolved nodes.
type Builder struct {
	S     *a2ui.Surface
	St    *State
	out   *Surface
	forms []string
}

// Build makes the view of a surface as it is resolved now.
func Build(s *a2ui.Surface, st *State) *Surface {
	b := &Builder{S: s, St: st, out: &Surface{ID: s.ID, nodes: map[string]*a2ui.Node{}}}
	if s.Tree.Root != nil {
		b.out.Root = b.Node(s.Tree.Root)
	}
	if st.Modal != "" && b.out.Overlay == nil {
		st.Modal = ""
	}
	for i, sc := range b.out.Shortcuts {
		b.out.Shortcuts[i].Press = b.pressTarget(sc.Press)
	}
	return b.out
}

// Node makes the element of a node, by its catalog's mapper; a node that
// cannot be drawn is a Placeholder.
func (b *Builder) Node(n *a2ui.Node) *Element {
	if n == nil {
		return nil
	}
	var e *Element
	if n.State != a2ui.Resolved {
		e = &Element{Kind: Placeholder, State: n.State}
	} else if m := mappers[n.Catalog][n.Type]; m != nil {
		if e = m(b, n); e == nil {
			return nil
		}
	} else {
		e = &Element{Kind: Placeholder, State: a2ui.Unknown}
	}
	e.ID, e.Type = n.Key, n.Type
	b.out.nodes[e.ID] = n
	if w, ok := a2ui.ToNumber(b.Raw(n, "weight")), n.Props["weight"] != nil; ok && w > 0 {
		e.Weight = w
	}
	e.A11y = b.a11y(n)
	e.Autofocus = autofocus(n)
	e.Keys = keys(n)
	return e
}

// Children makes the elements of a prop that holds nodes, in order.
func (b *Builder) Children(v any) []*Element {
	var out []*Element
	switch x := v.(type) {
	case *a2ui.Node:
		if e := b.Node(x); e != nil {
			out = append(out, e)
		}
	case []*a2ui.Node:
		for _, n := range x {
			if e := b.Node(n); e != nil {
				out = append(out, e)
			}
		}
	}
	return out
}

// Raw is a prop's resolved value: a binding's value, a literal as it is.
func (b *Builder) Raw(n *a2ui.Node, prop string) any {
	v := n.Props[prop]
	if bd, ok := v.(a2ui.Bound); ok {
		return bd.Value
	}
	return v
}

// String is a prop's value as text: "" when absent or null.
func (b *Builder) String(n *a2ui.Node, prop string) string {
	v := b.Raw(n, prop)
	if v == nil {
		return ""
	}
	if f, ok := v.(float64); ok {
		return a2ui.NumberString(f)
	}
	return a2ui.ToString(v)
}

// Enum is a static prop's value, or def.
func (b *Builder) Enum(n *a2ui.Node, prop, def string) string {
	if s, ok := b.Raw(n, prop).(string); ok && s != "" {
		return s
	}
	return def
}

// Bool is a prop's value as a boolean.
func (b *Builder) Bool(n *a2ui.Node, prop string) bool { return a2ui.Truthy(b.Raw(n, prop)) }

// Value is a control's value: the user's, when it is not bound and the
// user changed it; else the prop's.
func (b *Builder) Value(n *a2ui.Node, prop string) any {
	if bd, ok := n.Props[prop].(a2ui.Bound); !ok || !bd.Writable() {
		if v, ok := b.St.Local[n.Key]; ok {
			return v
		}
	}
	return b.Raw(n, prop)
}

// Checks reads a node's checks: whether they pass, and the first failing
// one's message.
func (b *Builder) Checks(n *a2ui.Node) (valid bool, msg string) {
	rs, _ := n.Props["checks"].([]a2ui.ValidationResult)
	for _, r := range rs {
		if !r.Valid && (r.Severity == "" || r.Severity == "error") {
			return false, r.Message
		}
	}
	return true, ""
}

// FieldError is the message a control shows: its first failing check's,
// once the user has touched it or tried its form.
func (b *Builder) FieldError(n *a2ui.Node) string {
	valid, msg := b.Checks(n)
	if valid {
		return ""
	}
	if b.St.Touched[n.Key] {
		return msg
	}
	for _, f := range b.forms {
		if b.St.Submitted[f] {
			return msg
		}
	}
	return ""
}

// InForm runs fn with the form id as the enclosing form.
func (b *Builder) InForm(id string, fn func()) {
	b.forms = append(b.forms, id)
	defer func() { b.forms = b.forms[:len(b.forms)-1] }()
	fn()
}

// SetOverlay makes e the surface's overlay: an open Modal's content.
func (b *Builder) SetOverlay(e *Element) { b.out.Overlay = e }

// AddShortcut records a key the surface takes; press is the component id
// of the Button it presses, if any.
func (b *Builder) AddShortcut(sc Shortcut) { b.out.Shortcuts = append(b.out.Shortcuts, sc) }

// pressTarget finds the element of the Button a Shortcut presses by its
// component id: the first in tree order.
func (b *Builder) pressTarget(componentID string) string {
	if componentID == "" {
		return ""
	}
	found := ""
	b.out.Walk(func(e *Element) bool {
		if n := b.out.nodes[e.ID]; n != nil && n.ComponentID == componentID && e.Kind == Button {
			found = e.ID
		}
		return found == ""
	})
	return found
}

func (b *Builder) a11y(n *a2ui.Node) A11y {
	m, _ := n.Props["accessibility"].(map[string]any)
	if m == nil {
		return A11y{}
	}
	get := func(k string) any {
		if bd, ok := m[k].(a2ui.Bound); ok {
			return bd.Value
		}
		return m[k]
	}
	a := A11y{Hidden: a2ui.Truthy(get("hidden"))}
	if s, ok := get("label").(string); ok {
		a.Label = s
	}
	if s, ok := get("description").(string); ok {
		a.Description = s
	}
	if s, ok := get("live").(string); ok {
		a.Live = s
	}
	return a
}

// autofocus reads the hotty extension metadata.extensions
// .io_neuroplast_hotty.autofocus.
func autofocus(n *a2ui.Node) bool { return hottyExt(n)["autofocus"] == true }

// keys reads the hotty extension metadata.extensions
// .io_neuroplast_hotty.keys: a keymap, as SPEC §10.2's data-keys.
func keys(n *a2ui.Node) string {
	s, _ := hottyExt(n)["keys"].(string)
	return s
}

// hottyExt is a node's metadata.extensions.io_neuroplast_hotty.
func hottyExt(n *a2ui.Node) map[string]any {
	md, _ := n.Props["metadata"].(map[string]any)
	ext, _ := md["extensions"].(map[string]any)
	h, _ := ext["io_neuroplast_hotty"].(map[string]any)
	return h
}

// SubID is the id of a part of an element that is not a node: a Tabs'
// title, a Choice's option.
func SubID(id, part string, i int) string { return id + "/" + part + "/" + strconv.Itoa(i) }
