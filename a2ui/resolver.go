package a2ui

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

// MaxDepth bounds how deep a surface's tree may nest.
const MaxDepth = 50

// MaxTemplateItems bounds how many children one template makes.
const MaxTemplateItems = 10_000

// NodeState is how far a node could be resolved.
type NodeState string

// The states of a node.
const (
	// Resolved: the node's component exists and its type is known.
	Resolved NodeState = "resolved"
	// Pending: a parent refers to a component that has not arrived.
	Pending NodeState = "pending"
	// Unknown: the component's type is in none of the renderer's
	// catalogs, or its catalog is unknown.
	Unknown NodeState = "unknown"
	// Cyclic: the component already appears above itself, in the same
	// scope, or the tree is too deep.
	Cyclic NodeState = "cyclic"
)

// Node is one mounted instance of a component: the component itself at the
// root scope, one per item of a template, one per parent that refers to
// it. A node keeps its identity while its edge (parent, property,
// component, scope) and its component do; its Props change in place.
type Node struct {
	// Key names the node uniquely within its surface: the component's id,
	// with the template item's path in brackets ("row[/items/0]"), and a
	// "#n" for a second node of the same.
	Key         string
	ComponentID string
	Type        string
	Catalog     string
	State       NodeState
	Scope       Scope
	Parent      *Node
	// Def is the component's type, when State is Resolved.
	Def *ComponentType
	// Props are the resolved properties: Bound for dynamic ones, *Node
	// for a child, []*Node for a child list, ActionRef for an action,
	// []ValidationResult for checks, and the literal value for the rest.
	Props map[string]any

	comp      *Component
	destroyed bool
}

// Bound is a resolved dynamic property: its value now, and the data path it
// is bound to when it is a binding (and so writable).
type Bound struct {
	Value any
	// Path is the absolute data path of a binding; "" for a literal or a
	// call.
	Path string
	// Err is why the value could not be resolved; Value is then nil.
	Err error
}

// Writable reports whether the property is a binding, which the user's
// input writes back.
func (b Bound) Writable() bool { return b.Path != "" }

// ActionRef is an action property, run by Resolver.Invoke.
type ActionRef struct {
	Raw map[string]any
}

// Destroyed reports whether the node has left the tree.
func (n *Node) Destroyed() bool { return n.destroyed }

// Valid reports whether every check of the node passes: none fails with
// severity error.
func (n *Node) Valid() bool {
	checks, _ := n.Props["checks"].([]ValidationResult)
	for _, c := range checks {
		if !c.Valid && (c.Severity == "" || c.Severity == "error") {
			return false
		}
	}
	return true
}

// Children lists the node's children: its child and child-list
// properties, in the order of the properties' names, then nested ones.
func (n *Node) Children() []*Node {
	var out []*Node
	var walk func(v any)
	walk = func(v any) {
		switch x := v.(type) {
		case *Node:
			out = append(out, x)
		case []*Node:
			out = append(out, x...)
		case []any:
			for _, e := range x {
				walk(e)
			}
		case map[string]any:
			for _, k := range SortedKeys(x) {
				walk(x[k])
			}
		}
	}
	for _, k := range SortedKeys(n.Props) {
		walk(n.Props[k])
	}
	return out
}

type edgeKey struct {
	parent *Node
	slot   string
	id     string
	path   string
}

// Resolver keeps a surface's node tree: Resolve rebuilds it from the
// components and the data, keeping the nodes whose edge and component are
// unchanged, and reports the retained nodes whose props changed and the
// nodes that left.
type Resolver struct {
	s *Surface
	// Root is the tree's root node: nil until a component "root" exists.
	Root *Node
	// OnEmit is called, after a Resolve, for each retained node whose props
	// changed.
	OnEmit func(*Node)
	// OnDestroy is called once for each node that leaves the tree.
	OnDestroy func(*Node)

	edges    map[edgeKey]*Node
	keys     map[string]*Node
	disposed bool
	reported map[string]bool

	// during a Resolve
	next    map[edgeKey]*Node
	fresh   []*Node
	emit    []*Node
	errs    []error
	errKeys map[string]bool
}

func newResolver(s *Surface) *Resolver {
	return &Resolver{s: s, edges: map[edgeKey]*Node{}, keys: map[string]*Node{}, reported: map[string]bool{}}
}

// Node finds a node by its Key.
func (r *Resolver) Node(key string) *Node { return r.keys[key] }

// Nodes lists every node in the tree, depth first.
func (r *Resolver) Nodes() []*Node {
	var out []*Node
	var walk func(n *Node)
	walk = func(n *Node) {
		out = append(out, n)
		for _, c := range n.Children() {
			walk(c)
		}
	}
	if r.Root != nil {
		walk(r.Root)
	}
	return out
}

// Resolve rebuilds the tree from the surface's components and data.
func (r *Resolver) Resolve() {
	if r.disposed {
		return
	}
	r.next = map[edgeKey]*Node{}
	r.fresh, r.emit, r.errs = nil, nil, nil
	r.errKeys = map[string]bool{}
	var root *Node
	if _, ok := r.s.Components["root"]; ok {
		root = r.build(nil, "root", "root", RootScope, nil)
	}
	r.Root = root
	// Nodes left behind leave the tree.
	var gone []*Node
	for k, n := range r.edges {
		if r.next[k] != n {
			gone = append(gone, n)
		}
	}
	r.edges = r.next
	r.next = nil
	r.keys = map[string]*Node{}
	for _, n := range r.edges {
		if n.Key != "" {
			r.keys[n.Key] = n
		}
	}
	for _, n := range r.fresh {
		n.Key = r.uniqueKey(n)
		r.keys[n.Key] = n
	}
	r.destroy(gone)
	for _, n := range r.emit {
		if r.OnEmit != nil && !n.destroyed {
			r.OnEmit(n)
		}
	}
	for _, err := range r.errs {
		r.report(err)
	}
}

// Dispose destroys every node; the resolver resolves nothing after it.
func (r *Resolver) Dispose() {
	var all []*Node
	for _, n := range r.edges {
		all = append(all, n)
	}
	r.edges = map[edgeKey]*Node{}
	r.keys = map[string]*Node{}
	r.Root = nil
	r.disposed = true
	r.destroy(all)
}

func (r *Resolver) destroy(nodes []*Node) {
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].Key < nodes[j].Key })
	for _, n := range nodes {
		if n.destroyed {
			continue
		}
		n.destroyed = true
		if r.OnDestroy != nil {
			r.OnDestroy(n)
		}
	}
}

func (r *Resolver) uniqueKey(n *Node) string {
	base := n.ComponentID
	if n.Scope.InTemplate {
		base += "[" + n.Scope.Path + "]"
	}
	key := base
	for i := 2; r.keys[key] != nil; i++ {
		key = base + "#" + strconv.Itoa(i)
	}
	return key
}

// report sends an error to the agent once per distinct message.
func (r *Resolver) report(err error) {
	msg := err.Error()
	if r.reported[msg] {
		return
	}
	r.reported[msg] = true
	if r.s.env.Error != nil {
		r.s.env.Error(r.s, &ErrorMessage{Code: Code(err), Message: msg, SurfaceID: r.s.ID})
	}
}

func (r *Resolver) fail(err error) {
	if !r.errKeys[err.Error()] {
		r.errKeys[err.Error()] = true
		r.errs = append(r.errs, err)
	}
}

type frame struct {
	id   string
	path string
	up   *frame
	n    int
}

func (f *frame) has(id, path string) bool {
	for ; f != nil; f = f.up {
		if f.id == id && f.path == path {
			return true
		}
	}
	return false
}

func (r *Resolver) build(parent *Node, slot, id string, scope Scope, up *frame) *Node {
	ek := edgeKey{parent: parent, slot: slot, id: id, path: scope.Path}
	comp := r.s.Components[id]
	state, typ, catalog := Resolved, "", ""
	var def *ComponentType
	depth := 0
	if up != nil {
		depth = up.n + 1
	}
	switch {
	case comp == nil:
		state = Pending
	case up.has(id, scope.Path) || depth >= MaxDepth:
		state, typ, catalog = Cyclic, comp.Type, r.s.CatalogOf(comp)
		r.fail(&IntegrityError{Msg: fmt.Sprintf("Component '%s' refers to itself through its children", id)})
	default:
		typ, catalog = comp.Type, r.s.CatalogOf(comp)
		if def = r.s.ComponentType(comp); def == nil {
			state = Unknown
			r.fail(&CatalogError{Msg: fmt.Sprintf("Component '%s' has type '%s', which catalog '%s' does not have", id, comp.Type, catalog)})
		}
	}
	n := r.edges[ek]
	retained := n != nil && !n.destroyed && n.comp == comp && n.State == state && n.Type == typ
	if !retained {
		n = &Node{ComponentID: id, Type: typ, Catalog: catalog, State: state, Scope: scope, Parent: parent, Def: def, comp: comp}
		r.fresh = append(r.fresh, n)
	}
	n.Def = def
	r.next[ek] = n
	props := map[string]any{}
	if state == Resolved {
		ctx := r.s.Context(scope)
		ctx.Caller = id
		f := &frame{id: id, path: scope.Path, up: up, n: depth}
		for _, name := range SortedKeys(comp.Props) {
			p := def.Props[name]
			if p == nil {
				p = &Prop{Kind: Static}
			}
			props[name] = r.prop(n, name, p, comp.Props[name], ctx, f)
		}
	}
	if retained && !propsEqual(n.Props, props) {
		r.emit = append(r.emit, n)
	}
	n.Props = props
	return n
}

// prop resolves one property value by its kind; slot names where it is in
// the node, for the children it makes.
func (r *Resolver) prop(n *Node, slot string, p *Prop, v any, ctx *Context, f *frame) any {
	switch p.Kind {
	case Dynamic:
		if path, ok := IsBinding(v); ok {
			abs := ctx.Path(path)
			if _, err := ParsePointer(abs); err != nil {
				r.fail(&ValidationError{Msg: err.Error()})
				return Bound{Err: err}
			}
			return Bound{Value: ctx.Data.Value(abs), Path: abs}
		}
		val, err := ctx.Resolve(v)
		if err != nil {
			r.fail(fmt.Errorf("%s.%s: %w", n.ComponentID, slot, err))
			return Bound{Err: err}
		}
		return Bound{Value: val}
	case ChildRef:
		id, ok := v.(string)
		if !ok {
			return nil
		}
		return r.build(n, slot, id, ctx.Scope, f)
	case ChildList:
		return r.childList(n, slot, v, ctx, f)
	case ActionProp:
		m, _ := v.(map[string]any)
		return ActionRef{Raw: m}
	case Checks:
		return r.checks(v, ctx)
	case Array:
		l, ok := v.([]any)
		if !ok {
			return v
		}
		out := make([]any, len(l))
		for i, e := range l {
			out[i] = r.prop(n, slot+"["+strconv.Itoa(i)+"]", p.Items, e, ctx, f)
		}
		return out
	case Object:
		m, ok := v.(map[string]any)
		if !ok {
			return v
		}
		out := make(map[string]any, len(m))
		for _, k := range SortedKeys(m) {
			fp := p.Fields[k]
			if fp == nil {
				fp = &Prop{Kind: Static}
			}
			out[k] = r.prop(n, slot+"."+k, fp, m[k], ctx, f)
		}
		return out
	}
	return v
}

func (r *Resolver) childList(n *Node, slot string, v any, ctx *Context, f *frame) []*Node {
	var out []*Node
	switch x := v.(type) {
	case []any:
		for i, e := range x {
			if id, ok := e.(string); ok {
				out = append(out, r.build(n, slot+"["+strconv.Itoa(i)+"]", id, ctx.Scope, f))
			}
		}
	case map[string]any:
		id, _ := x["componentId"].(string)
		path, _ := x["path"].(string)
		if id == "" {
			return nil
		}
		list := ctx.Path(path)
		items, _ := ctx.Data.Value(list).([]any)
		for i := range min(len(items), MaxTemplateItems) {
			scope := Scope{Path: Join(list, strconv.Itoa(i)), Index: i, InTemplate: true}
			out = append(out, r.build(n, slot+"["+strconv.Itoa(i)+"]", id, scope, f))
		}
	}
	return out
}

// checks evaluates a component's check rules into their results, in order.
func (r *Resolver) checks(v any, ctx *Context) []ValidationResult {
	rules, _ := v.([]any)
	out := make([]ValidationResult, 0, len(rules))
	for i, raw := range rules {
		rule, ok := raw.(map[string]any)
		if !ok {
			r.fail(&ValidationError{Msg: fmt.Sprintf("check %d is not an object", i), Path: fmt.Sprintf("/checks/%d", i)})
			continue
		}
		cond, ok := rule["condition"]
		if !ok {
			cond = rule
		}
		msg, _ := rule["message"].(string)
		val, err := ctx.Resolve(cond)
		if err != nil {
			r.fail(err)
			out = append(out, ValidationResult{Valid: false, Message: fallback(msg, err.Error())})
			continue
		}
		out = append(out, checkResult(val, msg))
	}
	return out
}

// checkResult reads what a check's condition evaluated to: a
// ValidationResult or an object with "valid" by its validity, anything
// else valid only when it is true. An invalid result's message is the
// rule's, else its own, else "Validation failed".
func checkResult(v any, msg string) ValidationResult {
	res := ValidationResult{}
	switch x := v.(type) {
	case ValidationResult:
		res = x
	case map[string]any:
		if b, ok := x["valid"].(bool); ok {
			res.Valid = b
			res.Code, _ = x["code"].(string)
			res.Message, _ = x["message"].(string)
			res.Severity, _ = x["severity"].(string)
		}
	default:
		res.Valid = v == true
	}
	if !res.Valid {
		res.Message = fallback(msg, fallback(res.Message, "Validation failed"))
	}
	return res
}

func fallback(s, def string) string {
	if s != "" {
		return s
	}
	return def
}

// propsEqual compares two nodes' props: nodes by identity, the rest by
// value.
func propsEqual(a, b map[string]any) bool {
	if len(a) != len(b) {
		return false
	}
	for k, x := range a {
		y, ok := b[k]
		if !ok || !valueEqual(x, y) {
			return false
		}
	}
	return true
}

func valueEqual(a, b any) bool {
	switch x := a.(type) {
	case *Node:
		y, ok := b.(*Node)
		return ok && x == y
	case []*Node:
		y, ok := b.([]*Node)
		if !ok || len(x) != len(y) {
			return false
		}
		for i := range x {
			if x[i] != y[i] {
				return false
			}
		}
		return true
	case []any:
		y, ok := b.([]any)
		if !ok || len(x) != len(y) {
			return false
		}
		for i := range x {
			if !valueEqual(x[i], y[i]) {
				return false
			}
		}
		return true
	case map[string]any:
		y, ok := b.(map[string]any)
		return ok && propsEqual(x, y)
	case Bound:
		y, ok := b.(Bound)
		return ok && x.Path == y.Path && reflect.DeepEqual(x.Value, y.Value) && errText(x.Err) == errText(y.Err)
	}
	return reflect.DeepEqual(a, b)
}

func errText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// Write writes a value through a node's writable property: the data model
// changes at the property's path, and the tree is resolved again.
func (r *Resolver) Write(n *Node, prop string, v any) error {
	b, ok := n.Props[prop].(Bound)
	if !ok || !b.Writable() {
		return fmt.Errorf("a2ui: %s.%s is not writable", n.Key, prop)
	}
	return r.s.Write(b.Path, v)
}

// ErrDisabled is an action of a component whose checks fail.
var ErrDisabled = errors.New("a2ui: the component's checks fail")

// Invoke runs a node's action property, with its context resolved now: an
// event goes to the agent as an action, a function call runs (with the
// user's activation when activation is set). A component whose checks
// fail does nothing.
func (r *Resolver) Invoke(n *Node, prop string, activation bool) error {
	if !n.Valid() {
		return ErrDisabled
	}
	raw, _ := n.comp.Props[prop].(map[string]any)
	if raw == nil {
		return fmt.Errorf("a2ui: %s has no action %s", n.Key, prop)
	}
	ctx := r.s.Context(n.Scope)
	ctx.Activation = activation
	ctx.Caller = n.ComponentID
	if fc, ok := raw["functionCall"].(map[string]any); ok {
		_, err := ctx.Resolve(fc)
		if err != nil {
			r.report(err)
		}
		r.Resolve()
		return err
	}
	ev, ok := raw["event"].(map[string]any)
	if !ok {
		err := &ValidationError{Msg: fmt.Sprintf("%s.%s is neither an event nor a functionCall", n.ComponentID, prop)}
		r.report(err)
		return err
	}
	a, err := r.ActionFor(n, ev, ctx)
	if err != nil {
		r.report(err)
		return err
	}
	if r.s.env.Action != nil {
		r.s.env.Action(r.s, a)
	}
	return nil
}

// ActionFor builds the action message an event sends, resolving its
// context and userMessage in ctx.
func (r *Resolver) ActionFor(n *Node, ev map[string]any, ctx *Context) (*ActionMessage, error) {
	name, _ := ev["name"].(string)
	if strings.TrimSpace(name) == "" {
		return nil, &ValidationError{Msg: "an event action needs a name"}
	}
	a := &ActionMessage{
		Name:              name,
		SurfaceID:         r.s.ID,
		SourceComponentID: n.ComponentID,
		Timestamp:         Now().UTC().Format("2006-01-02T15:04:05.000Z"),
		Context:           map[string]any{},
	}
	if c, ok := ev["context"].(map[string]any); ok {
		for _, k := range SortedKeys(c) {
			v, err := ctx.Resolve(c[k])
			if err != nil {
				return nil, err
			}
			a.Context[k] = v
		}
	}
	if um, ok := ev["userMessage"]; ok {
		v, err := ctx.Resolve(um)
		if err != nil {
			return nil, err
		}
		if s, ok := v.(string); ok {
			a.UserMessage = s
		}
	}
	return a, nil
}

// Now is the clock actions are stamped with; tests replace it.
var Now = time.Now
