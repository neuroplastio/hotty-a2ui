// Package view is what the HOTTY renditions draw: a surface's resolved
// A2UI nodes, with the renderer's own state (the tab shown, the modal
// open, a control's unbound value), made into a tree of a few kinds of
// element. Each rendition (html, cells, text) draws those kinds, so a
// catalog maps its components onto them once.
package view

import (
	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/diff"
	"github.com/neuroplastio/hotty-a2ui/highlight"
)

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
	// Switch edits a boolean as an on/off switch (HottySwitch): Label,
	// Value (a bool), Disabled while it cannot be flipped.
	Switch Kind = "switch"
	// Choice picks one or more of Options. One of a single choice shown
	// as checkboxes is a list to pick from (a select); otherwise each
	// option is a child, which toggles.
	Choice Kind = "choice"
	// Option is one of a Choice's options as a child: Label, Active when
	// it is picked.
	Option Kind = "option"
	// Slider edits a number from Min to Max; Fill says which side of its
	// knob is filled (io_neuroplast_hotty.fill).
	Slider Kind = "slider"
	// RangeSlider edits a range of numbers from Min to Max
	// (HottyRangeSlider): Label, Step and Error as a Slider's, Disabled
	// while it cannot be moved; its two Knob children, the start's and the
	// end's (Range).
	RangeSlider Kind = "rangeslider"
	// Knob is one of a RangeSlider's two knobs, each a Tab stop: Value, its
	// number; Selected, 0 for the start's and 1 for the end's; Label, its
	// name for a screen reader; Disabled as its RangeSlider.
	Knob Kind = "knob"
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
	// RichList is items to pick from, a label and a description each
	// (HottyList): Label (its title), Items, RowIDs (each item's id), Value
	// (the selected item's id), Filter (whether it filters), Query, Shown
	// (the items the query leaves, best first) and Matched, Height (the
	// items a page shows, 0 for all of them), Top (the page's first, in
	// Shown) and Placeholder (what shows when no item does).
	RichList Kind = "richlist"
	// KeyHints shows the keys its surface takes now (HottyKeyHints), as
	// Controller.KeyHints has them: Open while the full view shows, Active
	// when ? switches views.
	KeyHints Kind = "keyhints"
	// ScrollView is a box of Height rows whose content scrolls
	// (HottyScrollView): its child (Children), or Lines; Top and Left,
	// where it is scrolled to; Active, whether it follows the tail when
	// scrolled to its end, and Tail, whether it does now; Wrap, whether
	// long lines wrap.
	ScrollView Kind = "scrollview"
	// Listing is source code (HottyCode): Code, its lines as tokens;
	// Lang; Numbers, whether its line numbers show, from FirstLine; Marks,
	// by line number; Wrap, whether long lines wrap; Left, without it, the
	// first column of code shown.
	Listing Kind = "code"
	// DiffView is a change to files (HottyDiff): Diff, its files of hunks;
	// Variant, unified or split; Numbers, whether line numbers show; Wrap;
	// RowIDs, its hunks' ids, and Value, the selected one's, as a Table's.
	DiffView Kind = "diff"
	// Tree is nodes in branches that fold (HottyTree): Nodes, in
	// pre-order; RowIDs, each node's id; Value, the selected one's;
	// Expanded, the ids of the branches unfolded; Query, its filter's
	// text; Shown, the nodes that show, and Matched; Height and Top as a
	// Table's, Top in Shown; Placeholder, what shows when it has no node.
	Tree Kind = "tree"
	// Chart is series of numbers drawn on an axis (HottyChart): Variant,
	// line or bar; Series; Labels, a label a point; Height, the plot's
	// rows; Window, the points its x axis holds (0: as many as it has);
	// Min, Max and Step, its axis (YAxis).
	Chart Kind = "chart"
	// Sparkline is one series of numbers in a few cells, a column each
	// (HottySparkline): Series, Height, Window as a Chart's; Min and Max,
	// the values at its bottom and its top.
	Sparkline Kind = "sparkline"
	// Toast is a notice over the surface, in a corner, for a while
	// (hottyToast, Surface.Toasts): Label, its message; Variant, its kind
	// (info, success, warning, error); Name, its id; its ToastAction, when
	// it has one.
	Toast Kind = "toast"
	// ToastAction is a toast's action, a Tab stop: Label, what it does;
	// Name, the toast's id.
	ToastAction Kind = "toastaction"
	// Timer is a time that counts, down (HottyTimer) or up
	// (HottyStopwatch): Variant, TimerCountdown or TimerStopwatch; Label;
	// Value, the time it shows as text (FormatTimer); Active while it
	// counts.
	Timer Kind = "timer"
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

	// Image and Media: where it is, what it shows. Icon: Name, one of
	// the basic catalog's or (HottyIcon) a Material Symbols name, or else
	// Path, an svgPath's path data, stroked Stroke wide when that is more
	// than 0. Tab and Option: Name, an icon by its title, if the Tabs or
	// the ChoicePicker gives it one (io_neuroplast_hotty.icons).
	URL    string  `json:"url,omitempty"`
	Alt    string  `json:"alt,omitempty"`
	Fit    string  `json:"fit,omitempty"`
	Name   string  `json:"name,omitempty"`
	Path   string  `json:"path,omitempty"`
	Stroke float64 `json:"stroke,omitempty"`

	// Controls.
	Label       string `json:"label,omitempty"`
	Placeholder string `json:"placeholder,omitempty"`
	// Value is a TextField's or a DateTime's string, a CheckBox's or a
	// Switch's bool, a Choice's []string, a Slider's float64, a Timer's
	// time as text.
	Value any `json:"value,omitempty"`
	// Error is the message of the first check that fails, once the user
	// has touched the control or tried to submit.
	Error string `json:"error,omitempty"`
	// Disabled: a Button or a Form whose checks fail does nothing, and
	// neither does a Switch its disabled prop holds.
	Disabled bool           `json:"disabled,omitempty"`
	Options  []ChoiceOption `json:"options,omitempty"`
	Multiple bool           `json:"multiple,omitempty"`
	Filter   bool           `json:"filter,omitempty"`
	// Slider and RangeSlider: the range, and the step (0: any value).
	// Progress: Max.
	Min  float64 `json:"min,omitempty"`
	Max  float64 `json:"max,omitempty"`
	Step float64 `json:"step,omitempty"`
	// Fill: a Slider's filled side, "end" from its knob to Max
	// (io_neuroplast_hotty.fill); "" from Min to its knob.
	Fill string `json:"fill,omitempty"`
	// DateTime: which parts it edits, and its bounds (DateHint).
	Date   bool   `json:"date,omitempty"`
	Time   bool   `json:"time,omitempty"`
	MinISO string `json:"minIso,omitempty"`
	MaxISO string `json:"maxIso,omitempty"`

	// Tabs: which tab is shown. Tab and Option: whether it is the one
	// shown, or picked. Modal: whether its content is shown. Spinner:
	// whether it spins. Timer: whether it counts. A TextField with suggestions: Selected is the
	// highlighted one's place in Shown, -1 for none.
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

	// RichList: its items; the query that filters them; the items it
	// leaves, by index, best first; and where each of those matched it,
	// the byte offsets of its label's matched characters (nil without a
	// query). Its RowIDs, Height and Top are a Table's, Top in Shown.
	Items   []Entry `json:"items,omitempty"`
	Query   Query   `json:"query,omitzero"`
	Shown   []int   `json:"shown,omitempty"`
	Matched [][]int `json:"matched,omitempty"`

	// TextField: its suggestions (io_neuroplast_hotty.suggestions, profile
	// §6.22), the agent's options, non-nil when it has the extension. Its
	// Shown are those its value leaves, by index, while its list is open;
	// Selected, Height and Top the highlighted one, and the rows the list
	// shows at most from Top (Builder.suggestions). GhostOnly: the
	// extension's list is false, so only the ghost after the value shows.
	Suggestions []string `json:"suggestions,omitempty"`
	GhostOnly   bool     `json:"ghostOnly,omitempty"`

	// ScrollView: its lines, when it has no child; whether they wrap; the
	// first column shown, when they do not; and whether it follows the
	// tail. Its Height and Top are a Table's, Top a row.
	Lines []string `json:"lines,omitempty"`
	Wrap  bool     `json:"wrap,omitempty"`
	Left  int      `json:"left,omitempty"`
	Tail  bool     `json:"tail,omitempty"`

	// Code: its lines as tokens (highlight.Lines, shared: not to be
	// changed), the language they were lexed as, whether the line numbers
	// show, the first line's number, and the marked lines, by number
	// (Mark). Its Wrap is a ScrollView's.
	Code      [][]highlight.Token `json:"code,omitempty"`
	Lang      string              `json:"lang,omitempty"`
	Numbers   bool                `json:"numbers,omitempty"`
	FirstLine int                 `json:"firstLine,omitempty"`
	Marks     map[int]Mark        `json:"marks,omitempty"`

	// DiffView: its files of hunks (shared: not to be changed).
	Diff *diff.Diff `json:"diff,omitempty"`

	// Tree: its nodes, and the ids of the branches unfolded (TreeNode
	// Open adds those the selection is in).
	Nodes    []TreeNode `json:"nodes,omitempty"`
	Expanded []string   `json:"expanded,omitempty"`

	// Chart and Sparkline: the series, the points' labels, and the points
	// the x axis holds (0: as many as there are). Their Height is the
	// plot's rows; Min, Max and Step its axis.
	Series []Series `json:"series,omitempty"`
	Labels []string `json:"labels,omitempty"`
	Window int      `json:"window,omitempty"`

	// Drag and drop (profile §6.21). Movable: a HottyList, HottyTable or
	// HottyTree that is reorderable, or a List of a template with
	// io_neuroplast_hotty.reorder, whose items the user moves; ItemType, the
	// type its items carry in a drag (dragType, reorder's type). DragType:
	// the element is a drag source of that type (io_neuroplast_hotty.drag);
	// Accepts: the types it takes as a drop target (io_neuroplast_hotty.drop).
	Movable  bool     `json:"movable,omitempty"`
	ItemType string   `json:"itemType,omitempty"`
	DragType string   `json:"dragType,omitempty"`
	Accepts  []string `json:"accepts,omitempty"`

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
	// Icon is its icon's name, if the ChoicePicker gives it one
	// (io_neuroplast_hotty.icons).
	Icon string `json:"icon,omitempty"`
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
// trigger is not a control, a RangeSlider's knob. A disabled Button,
// Switch or knob does not.
func (e *Element) Focusable() bool {
	switch e.Kind {
	case Button, Switch, Knob:
		return !e.Disabled
	case TextField, CheckBox, Slider, DateTime, Tab, Option, Table, RichList, ScrollView, Tree, ToastAction:
		return true
	case DiffView:
		return len(e.RowIDs) > 0
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
	// Toasts are the toasts shown over it, the newest first (hottyToast).
	Toasts []*Element `json:"toasts,omitempty"`

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

// Walk visits the elements in tree order, then the overlay's, then the
// toasts', while fn returns true.
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
	if !walk(s.Root) || !walk(s.Overlay) {
		return
	}
	for _, t := range s.Toasts {
		if !walk(t) {
			return
		}
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
// content's. The toasts' actions come last, the newest first, as the
// toasts follow the surface in a host's document.
func (s *Surface) Focusables() []string {
	var out []string
	add := func(e *Element) bool {
		if e.Focusable() {
			out = append(out, e.ID)
		}
		return true
	}
	if s.Overlay != nil {
		o := &Surface{Root: s.Overlay, Toasts: s.Toasts}
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

// SelectedRow is the index of a Table's selected row, or a HottyList's
// selected item, or -1.
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
