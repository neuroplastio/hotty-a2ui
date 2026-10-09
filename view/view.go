// Package view is what the HOTTY renditions draw: a surface's resolved
// A2UI nodes, with the renderer's own state (the tab shown, the modal
// open, a control's unbound value), made into a tree of a few kinds of
// element. Each rendition (html, cells, text) draws those kinds, so a
// catalog maps its components onto them once.
package view

import "github.com/neuroplastio/hotty-a2ui/a2ui"

// Kind is what an element is, for the renditions.
type Kind string

// The kinds of element.
const (
	// Stack lays its children out in a row or a column: Row, Column,
	// List.
	Stack Kind = "stack"
	// Card is a box around its child.
	Card Kind = "card"
	// Text is Markdown.
	Text Kind = "text"
	// Image is a picture: URL, Alt, Fit, Variant.
	Image Kind = "image"
	// Icon is a named icon.
	Icon Kind = "icon"
	// Media is a video or a sound, which a terminal shows as a labelled
	// link (NEIO-11's fallback).
	Media Kind = "media"
	// Divider is a rule along Dir.
	Divider Kind = "divider"
	// Button runs its action; its children are its content.
	Button Kind = "button"
	// TextField edits a string: Variant is shortText, longText, number
	// or obscured.
	TextField Kind = "textfield"
	// CheckBox edits a boolean.
	CheckBox Kind = "checkbox"
	// Choice picks one or more of Options. One of a single choice shown
	// as checkboxes is a list to pick from (a select); otherwise each
	// option is a child, which toggles.
	Choice Kind = "choice"
	// Option is one of a Choice's options as a child: Label, Active when
	// it is picked.
	Option Kind = "option"
	// Slider edits a number from Min to Max.
	Slider Kind = "slider"
	// DateTime edits a date, a time, or both (ISO 8601).
	DateTime Kind = "datetime"
	// Tabs shows its Tab children as a bar, and after them the content
	// of the tab Selected.
	Tabs Kind = "tabs"
	// Tab is one of a Tabs' titles: Label, Active when it is shown.
	Tab Kind = "tab"
	// Modal shows its first child, the trigger; the second, its content,
	// over the surface while Open.
	Modal Kind = "modal"
	// Form submits the fields inside it (hotty catalog).
	Form Kind = "form"
	// Progress is how far a task has come (HottyProgress): Value, a
	// float64 from 0 to Max, or nil while that is not known
	// (indeterminate); Label.
	Progress Kind = "progress"
	// Spinner shows that something is under way (HottySpinner): Variant
	// names its frame set (SpinnerFrames), Active whether it spins;
	// Label.
	Spinner Kind = "spinner"
	// Table is rows of data under a header (HottyTable): Columns, Cells
	// (each row's text, a column at a time), RowIDs (what identifies each
	// row), Value (the selected row's id, a string; "" for none), Height
	// (the body's rows, 0 for all of them) and Top (the first shown).
	Table Kind = "table"
	// Placeholder stands for a node that cannot be drawn: one still to
	// come (Pending), of a type no catalog here has (Unknown), or one
	// that contains itself (Cyclic). A component never fails its
	// surface.
	Placeholder Kind = "placeholder"
)

// Axis is a direction.
type Axis string

// The axes.
const (
	Horizontal Axis = "horizontal"
	Vertical   Axis = "vertical"
)

// Element is one element of a surface's view. Which fields mean
// something depends on its Kind.
type Element struct {
	// ID is the element's handle, stable while its node stays: the
	// node's key.
	ID   string `json:"id"`
	Kind Kind   `json:"kind"`
	// Type is the A2UI component type it was made from.
	Type     string     `json:"type"`
	Children []*Element `json:"children,omitempty"`
	// Weight is A2UI's weight in a Row or a Column (CSS flex, as A2UI's
	// renderers have it); 0 is none. The renditions say what it takes.
	Weight float64 `json:"weight,omitempty"`

	// Stack and Divider: the direction. Stack: how the children are
	// placed along it (Justify: start, center, end, spaceBetween,
	// spaceAround, spaceEvenly, stretch) and across it (Align: start,
	// center, end, stretch); a List scrolls.
	Dir     Axis   `json:"dir,omitempty"`
	Justify string `json:"justify,omitempty"`
	Align   string `json:"align,omitempty"`
	Scroll  bool   `json:"scroll,omitempty"`

	// Text: the Markdown. Variant: a Text's (body, caption), a Button's
	// (default, primary, borderless), an Image's size (icon, avatar,
	// smallFeature, mediumFeature, largeFeature, header), a TextField's,
	// a Choice's display (checkbox, chips).
	Markdown string `json:"markdown,omitempty"`
	Variant  string `json:"variant,omitempty"`

	// Image and Media: where it is, what it shows. Icon: Name.
	URL  string `json:"url,omitempty"`
	Alt  string `json:"alt,omitempty"`
	Fit  string `json:"fit,omitempty"`
	Name string `json:"name,omitempty"`

	// Controls.
	Label       string `json:"label,omitempty"`
	Placeholder string `json:"placeholder,omitempty"`
	// Value is a TextField's or a DateTime's string, a CheckBox's bool,
	// a Choice's []string, a Slider's float64.
	Value any `json:"value,omitempty"`
	// Error is the message of the first check that fails, once the user
	// has touched the control or tried to submit.
	Error string `json:"error,omitempty"`
	// Disabled: a Button or a Form whose checks fail does nothing.
	Disabled bool           `json:"disabled,omitempty"`
	Options  []ChoiceOption `json:"options,omitempty"`
	Multiple bool           `json:"multiple,omitempty"`
	Filter   bool           `json:"filter,omitempty"`
	// Slider: the range, and the step (0: any value). Progress: Max.
	Min  float64 `json:"min,omitempty"`
	Max  float64 `json:"max,omitempty"`
	Step float64 `json:"step,omitempty"`
	// DateTime: which parts it edits, and its bounds (DateHint).
	Date   bool   `json:"date,omitempty"`
	Time   bool   `json:"time,omitempty"`
	MinISO string `json:"minIso,omitempty"`
	MaxISO string `json:"maxIso,omitempty"`

	// Tabs: which tab is shown. Tab and Option: whether it is the one
	// shown, or picked. Modal: whether its content is shown. Spinner:
	// whether it spins.
	Selected int  `json:"selected,omitempty"`
	Active   bool `json:"active,omitempty"`
	Open     bool `json:"open,omitempty"`
	// Clickable: a Modal whose trigger has no control takes the keyboard
	// and a click itself.
	Clickable bool `json:"clickable,omitempty"`
	// Item: a Button that is an item of a vertical List is a row of it, as
	// a menu's are: as wide as the List, its variant marking the row (the
	// one picked, say) rather than shaping it.
	Item bool `json:"item,omitempty"`

	// Table: its columns, each row's cells as text, each row's id, the
	// rows of its body (0: as many as it has), and the first row the body
	// shows: where it was scrolled, moved as little as brings the selected
	// row into view (State.Scroll).
	Columns []Column   `json:"columns,omitempty"`
	Cells   [][]string `json:"cells,omitempty"`
	RowIDs  []string   `json:"rowIds,omitempty"`
	Height  int        `json:"height,omitempty"`
	Top     int        `json:"top,omitempty"`

	// Placeholder: State is the node's (pending, unknown, cyclic).
	State a2ui.NodeState `json:"state,omitempty"`

	// Autofocus: the element takes the keyboard when the surface is
	// first shown (io_neuroplast_hotty.autofocus).
	Autofocus bool `json:"autofocus,omitempty"`
	// Keys is the element's keymap for the text fields in it, itself if
	// it is one (io_neuroplast_hotty.keys): SPEC §10.2's data-keys, which
	// override the keymaps above it key by key (Surface.KeyChain).
	Keys string `json:"keys,omitempty"`
	A11y A11y   `json:"a11y,omitzero"`
}

// ChoiceOption is one of a Choice's options.
type ChoiceOption struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

// Column is one of a Table's columns: the row field it shows (Key), its
// header, its width in cells (0: its content's) and how its cells are
// aligned (start, center, end).
type Column struct {
	Key    string `json:"key"`
	Header string `json:"header,omitempty"`
	Width  int    `json:"width,omitempty"`
	Align  string `json:"align,omitempty"`
}

// A11y is a component's accessibility attributes.
type A11y struct {
	Label       string `json:"label,omitempty"`
	Description string `json:"description,omitempty"`
	// Live is polite or assertive: changes are announced.
	Live   string `json:"live,omitempty"`
	Hidden bool   `json:"hidden,omitempty"`
}

// Focusable reports whether the element takes the keyboard (SPEC §10.1):
// a control, a Tabs' title, a Choice's option, a link, a Modal whose
// trigger is not a control. A disabled Button does not.
func (e *Element) Focusable() bool {
	switch e.Kind {
	case Button:
		return !e.Disabled
	case TextField, CheckBox, Slider, DateTime, Tab, Option, Table:
		return true
	case Choice:
		return len(e.Children) == 0
	case Media:
		return e.URL != ""
	}
	return e.Clickable
}

// Shortcut is a key a surface takes (hotty catalog).
type Shortcut struct {
	// ID is the Shortcut's own element id.
	ID string `json:"id"`
	// Key is a W3C key value after its modifiers: "Control+s".
	Key string `json:"key"`
	// Press is the id of the Button element it presses; empty when it
	// runs its own action.
	Press string `json:"press,omitempty"`
	Label string `json:"label,omitempty"`
}

// Surface is one surface's view.
type Surface struct {
	ID        string     `json:"id"`
	Root      *Element   `json:"root,omitempty"`
	Shortcuts []Shortcut `json:"shortcuts,omitempty"`
	// Overlay is the content of the open Modal, drawn over the surface;
	// nil when none is open.
	Overlay *Element `json:"overlay,omitempty"`

	nodes map[string]*a2ui.Node
}

// Node is the A2UI node an element was made from.
func (s *Surface) Node(id string) *a2ui.Node { return s.nodes[id] }

// Find finds an element by id, in the tree or the overlay.
func (s *Surface) Find(id string) *Element {
	var found *Element
	s.Walk(func(e *Element) bool {
		if e.ID == id {
			found = e
		}
		return found == nil
	})
	return found
}

// Walk visits the elements in tree order, the overlay's last, while fn
// returns true.
func (s *Surface) Walk(fn func(*Element) bool) {
	var walk func(e *Element) bool
	walk = func(e *Element) bool {
		if e == nil {
			return true
		}
		if !fn(e) {
			return false
		}
		for _, c := range e.Children {
			if !walk(c) {
				return false
			}
		}
		return true
	}
	if walk(s.Root) {
		walk(s.Overlay)
	}
}

// KeyChain is the keymaps (Element.Keys) of an element and those it is
// in, the outermost first, as SPEC §10.2 resolves a field's keymap from
// them: the field's own last. Empty ones are left out; nil when the
// element is not there.
func (s *Surface) KeyChain(id string) []string {
	var path []*Element
	var find func(e *Element) bool
	find = func(e *Element) bool {
		if e == nil {
			return false
		}
		path = append(path, e)
		if e.ID == id {
			return true
		}
		for _, c := range e.Children {
			if find(c) {
				return true
			}
		}
		path = path[:len(path)-1]
		return false
	}
	if !find(s.Root) && !find(s.Overlay) {
		return nil
	}
	var out []string
	for _, e := range path {
		if e.Keys != "" {
			out = append(out, e.Keys)
		}
	}
	return out
}

// Focusables lists the ids of the elements that take the keyboard, in
// tree order: what Tab moves through. While a Modal is open, only its
// content's.
func (s *Surface) Focusables() []string {
	var out []string
	add := func(e *Element) bool {
		if e.Focusable() {
			out = append(out, e.ID)
		}
		return true
	}
	if s.Overlay != nil {
		o := &Surface{Root: s.Overlay}
		o.Walk(add)
		return out
	}
	s.Walk(add)
	return out
}

// Fraction is how far a Progress has come, from 0 to 1; ok is false
// while that is not known (indeterminate).
func (e *Element) Fraction() (f float64, ok bool) {
	v, ok := e.Value.(float64)
	if !ok || e.Max <= 0 {
		return 0, false
	}
	return min(max(v/e.Max, 0), 1), true
}

// SelectedRow is the index of a Table's selected row, or -1.
func (e *Element) SelectedRow() int {
	v, _ := e.Value.(string)
	if v == "" {
		return -1
	}
	for i, id := range e.RowIDs {
		if id == v {
			return i
		}
	}
	return -1
}

// SliderStep is how far a Slider moves for one step of the keyboard or a
// button: its Step, else a twentieth of its range.
func (e *Element) SliderStep() float64 {
	if e.Step > 0 {
		return e.Step
	}
	return (e.Max - e.Min) / 20
}

// DateHint is a DateTime's placeholder: the form its value takes.
func (e *Element) DateHint() string {
	switch {
	case e.Date && e.Time:
		return "YYYY-MM-DDTHH:MM"
	case e.Time:
		return "HH:MM"
	}
	return "YYYY-MM-DD"
}
