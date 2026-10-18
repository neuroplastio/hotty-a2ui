package view

import (
	"io/fs"
	"reflect"
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
	// item of the page each HottyList with a height shows, the first
	// row each HottyScrollView shows, and the first suggestion each text
	// field's list shows, by id.
	Scroll map[string]int
	// Left is the first column each HottyScrollView of lines that do not
	// wrap shows, and each HottyCode that does not wrap; Tail whether it follows the tail, once the user or the
	// agent moved it (its follow prop until then).
	Left map[string]int
	Tail map[string]bool
	// Query is each HottyList's filter, by its id.
	Query map[string]Query
	// Open are the branches unfolded in each HottyTree whose expanded is
	// not bound, by its id: what the user made them.
	Open map[string][]string
	// Revealed is the node each HottyTree, the row each HottyTable, or the
	// entry each HottyFilePicker (by its path), with a height last brought
	// into view, by its id: the rows it shows
	// follow the selection when that moves, and the wheel moves them
	// otherwise (ScrollRows).
	Revealed map[string]string
	// FullHints: the surface's HottyKeyHints show their full view (?).
	FullHints bool
	// Drag is the drag the user is making, nil while none is (Lift).
	Drag *Drag
	// Moved are the items the user moved in each HottyList, HottyTable and
	// HottyTree whose items are not bound, by its id (Builder.List).
	Moved map[string]Moved
	// Suggest is each text field's suggestions as the user worked them
	// (io_neuroplast_hotty.suggestions), by its id: the one highlighted,
	// or the list closed. The first its list shows is in Scroll.
	Suggest map[string]Suggest
	// Pick is each HottyPicker as the user worked it, by its id: its
	// selection where selected is not bound, and its actions panel. The
	// first line it shows is in Scroll.
	Pick map[string]Pick
	// Toasts are the toasts shown over the surface, the oldest first
	// (hottyToast); ToastSeq numbers those whose caller named none.
	Toasts   []*ToastState
	ToastSeq int
	// Timers are the time each HottyTimer and HottyStopwatch has counted,
	// by its node's key (TickTimers).
	Timers map[string]*TimerState
	// Jump is the heading a HottyMarkdown's link in place last went to
	// (FollowLink).
	Jump Jump
	// Roots are the filesystems a HottyFilePicker may read, by the name
	// its `root` gives (profile §6.31). The program grants them; the
	// renderer reads only within one, and the agent never reads a
	// filesystem. A picker whose root is not granted shows its error.
	Roots map[string]fs.FS
	// File is each HottyFilePicker's own state, by its id: the directory
	// it shows where `path` is not bound, and each directory's highlight.
	File map[string]FilePick
	// Clicks is the run of clicks on an item the user is making
	// (ClickItem).
	Clicks Clicks
}

// Clicks is a run of clicks on an item (Controller.ClickItem): element
// ID's item Item, and whether a click of the run acted on it.
type Clicks struct {
	ID    string
	Item  int
	Acted bool
}

// FilePick is a HottyFilePicker's renderer state: the directory it shows
// where `path` is not bound ("" for the root), the one its last build
// showed, each directory it showed as the user left it, by its path, and
// the listing it read last.
type FilePick struct {
	Dir     string
	Shown   string
	Dirs    map[string]DirPick
	Listing *FileListing
}

// FileListing is a directory as a HottyFilePicker read it: its root's
// name, its path, and every entry, sorted as the picker lists them, or
// why it could not be read. A build reads a directory again only when
// the picker shows another one, or the program grants a root (Grant).
type FileListing struct {
	Root, Dir string
	Files     []FileEntry
	Failed    string
}

// DirPick is a directory as a HottyFilePicker left it: the entry it
// highlighted, by name, and the first row it showed. Coming back to it
// shows them again.
type DirPick struct {
	Name string
	Top  int
}

// NewState is a surface's state before the user does anything.
func NewState() *State {
	return &State{Tabs: map[string]int{}, Local: map[string]any{}, Touched: map[string]bool{}, Submitted: map[string]bool{}, Scroll: map[string]int{}, Query: map[string]Query{},
		Left: map[string]int{}, Tail: map[string]bool{}, Open: map[string][]string{}, Revealed: map[string]string{}, Moved: map[string]Moved{},
		Suggest: map[string]Suggest{}, Pick: map[string]Pick{}, Timers: map[string]*TimerState{}, File: map[string]FilePick{}}
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
	// page: the view is a page's (BuildPage).
	page bool
}

// Build makes the view of a surface as it is resolved now.
func Build(s *a2ui.Surface, st *State) *Surface { return build(s, st, false) }

// BuildPage makes the view of a surface for a web page with no program
// behind it (rendition/html's page mode, profile §2.1): as Build, but a
// Tabs holds every tab, each title with its own content as its Children,
// where Build gives a Tabs the shown tab's content alone, after the
// titles. It notes in st what Build notes (where a Table or a HottyTree
// is scrolled to), the tabs not shown included.
func BuildPage(s *a2ui.Surface, st *State) *Surface { return build(s, st, true) }

func build(s *a2ui.Surface, st *State, page bool) *Surface {
	b := &Builder{S: s, St: st, out: &Surface{ID: s.ID, nodes: map[string]*a2ui.Node{}}, page: page}
	if s.Tree.Root != nil {
		b.out.Root = b.Node(s.Tree.Root)
	}
	if st.Modal != "" && b.out.Overlay == nil {
		st.Modal = ""
	}
	for i, sc := range b.out.Shortcuts {
		b.out.Shortcuts[i].Press = b.pressTarget(sc.Press)
	}
	b.out.Toasts = mapToasts(st)
	anchor(b.out)
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
	if e.Kind == TextField && e.Variant == "longText" {
		e.Resize = resize(n)
	}
	if e.Kind != Placeholder {
		dragDrop(n, e)
	}
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

// markDefault marks a HottyForm's default (profile §6.2): the first
// Button in it whose action is the form's onSubmit, so that Enter in the
// form does what pressing it does.
func (b *Builder) markDefault(form *Element, n *a2ui.Node) {
	submit, ok := n.Props["onSubmit"].(a2ui.ActionRef)
	if !ok {
		return
	}
	var walk func(e *Element) bool
	walk = func(e *Element) bool {
		if bn := b.out.nodes[e.ID]; e.Kind == Button && bn != nil {
			if act, ok := bn.Props["action"].(a2ui.ActionRef); ok && reflect.DeepEqual(act.Raw, submit.Raw) {
				e.Default = true
				return true
			}
		}
		for _, ch := range e.Children {
			if walk(ch) {
				return true
			}
		}
		return false
	}
	walk(form)
}

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

// fill reads the hotty extension metadata.extensions
// .io_neuroplast_hotty.fill on a Slider: "end" fills it from its knob to
// its max; anything else, "start" the default, from its min to its knob.
func fill(n *a2ui.Node) string {
	if hottyExt(n)["fill"] == "end" {
		return "end"
	}
	return ""
}

// cycle reads the hotty extension metadata.extensions
// .io_neuroplast_hotty.cycle on a ChoicePicker: true steps through its
// options in place, where it would be a select.
func cycle(n *a2ui.Node) bool {
	v, _ := hottyExt(n)["cycle"].(bool)
	return v
}

// resize reads the hotty extension metadata.extensions
// .io_neuroplast_hotty.resize on a longText TextField: the axes its
// corner resizes it along, "vertical", "horizontal" or "none"; anything
// else is "both", as a web textarea's.
func resize(n *a2ui.Node) string {
	switch v := hottyExt(n)["resize"]; v {
	case "vertical", "horizontal", "none":
		return v.(string)
	}
	return "both"
}

// iconNames reads the hotty extension metadata.extensions
// .io_neuroplast_hotty.icons: a Tabs' icons, a name a tab in order
// (byIndex), or a ChoicePicker's, a name an option's value; "" where it
// gives none.
func iconNames(n *a2ui.Node) (byIndex []string, byValue map[string]string) {
	switch v := hottyExt(n)["icons"].(type) {
	case []any:
		for _, x := range v {
			s, _ := x.(string)
			byIndex = append(byIndex, s)
		}
	case map[string]any:
		byValue = map[string]string{}
		for k, x := range v {
			if s, ok := x.(string); ok {
				byValue[k] = s
			}
		}
	}
	return byIndex, byValue
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
