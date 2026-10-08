package a2ui

import (
	"encoding/json"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

// MaxGlobalDepth is how deeply a payload may nest: deeper is a
// RecursionError, before any of it is applied. MaxCallDepth is how deeply
// function calls may nest in each other's arguments.
const (
	MaxGlobalDepth = 50
	MaxCallDepth   = 5
)

// The actions of a v1.0 message from the agent, in the spec's order, and
// those of other versions, which a v1.0 renderer knows only to refuse.
var (
	agentActions  = []string{"createSurface", "updateComponents", "updateDataModel", "deleteSurface", "callRendererFunction", "agentFunctionResponse"}
	otherActions  = []string{"beginRendering", "surfaceUpdate", "dataModelUpdate"}
	isAgentAction = func(k string) bool { return slices.Contains(agentActions, k) }
)

// Processor is a renderer's A2UI v1.0 message processor: it holds the
// catalogs the renderer supports and the surfaces agents create, applies
// the agent's messages, and sends the renderer's messages back.
//
// A Processor belongs to one goroutine, the renderer's, which also drives
// its surfaces (writes, actions). Only the agent calls it starts finish
// elsewhere, on their timers, and they touch nothing else.
type Processor struct {
	// Strict adds A2UI's topology checks to each component update (a
	// root, no orphans, no dangling references, no cycles) and refuses
	// component types the catalog lacks. Without it, a surface may be
	// built in any order, as streaming agents do.
	Strict bool
	// Locale and OpenURL are the surfaces' (Env).
	Locale  string
	OpenURL func(url string) error
	// Activated reports whether the user has just activated the renderer,
	// so that an agent's callRendererFunction of a function that requires
	// a user's activation may run.
	Activated func() bool
	// Send receives each message for the agent. It runs on the caller's
	// goroutine and must not call the processor back.
	Send func(Outbound)
	// Changed is called after the agent's messages change a surface: it
	// was created or updated, and resolved again, or it was deleted.
	Changed func(s *Surface, deleted bool)

	catalogs []*Catalog
	surfaces map[string]*Surface
	order    []string
	calls    agentCalls
}

// NewProcessor is a processor that supports catalogs, in order: a
// createSurface that names no catalog uses the first one of its version.
func NewProcessor(catalogs ...*Catalog) *Processor {
	p := &Processor{surfaces: map[string]*Surface{}}
	for _, c := range catalogs {
		p.Register(c)
	}
	return p
}

// Register adds a catalog, or replaces the one with its id.
func (p *Processor) Register(c *Catalog) {
	for i, o := range p.catalogs {
		if o.ID == c.ID {
			p.catalogs[i] = c
			return
		}
	}
	p.catalogs = append(p.catalogs, c)
}

// Catalogs lists the catalogs in the order they were registered.
func (p *Processor) Catalogs() []*Catalog { return slices.Clone(p.catalogs) }

// Catalog finds a catalog by id.
func (p *Processor) Catalog(id string) *Catalog {
	for _, c := range p.catalogs {
		if c.ID == id {
			return c
		}
	}
	return nil
}

// Surface finds a surface by id: nil when there is none.
func (p *Processor) Surface(id string) *Surface { return p.surfaces[id] }

// Surfaces lists the surfaces in the order they were created.
func (p *Processor) Surfaces() []*Surface {
	out := make([]*Surface, 0, len(p.order))
	for _, id := range p.order {
		out = append(out, p.surfaces[id])
	}
	return out
}

// Capabilities are the renderer's capabilities, as
// renderer_capabilities.json has them: for each version, the catalogs it
// supports. With no versions, v1.0's.
func (p *Processor) Capabilities(versions ...string) map[string]any {
	if len(versions) == 0 {
		versions = []string{Version}
	}
	out := map[string]any{}
	for _, v := range versions {
		ids := []any{}
		for _, c := range p.catalogs {
			if CompatibleVersions(c.ProtocolVersion, v) {
				ids = append(ids, c.ID)
			}
		}
		out[v] = map[string]any{"supportedCatalogIds": ids}
	}
	return out
}

// DataModels are the data models of the surfaces created with
// sendDataModel, as renderer_data_model.json has them: what the renderer
// sends along with each of its messages to the agent.
func (p *Processor) DataModels() map[string]any {
	surfaces := map[string]any{}
	for _, id := range p.order {
		if s := p.surfaces[id]; s.SendDataModel {
			surfaces[id] = s.Data.Root()
		}
	}
	return map[string]any{"version": Version, "surfaces": surfaces}
}

// ProcessJSON processes a JSON payload (see Process).
func (p *Processor) ProcessJSON(b []byte) error {
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		return p.fail(&ValidationError{Msg: "the payload is not JSON: " + err.Error()}, "")
	}
	return p.Process(v)
}

// Process applies a payload of the agent's messages: a list of them, the
// wrapper {"messages": [...]}, or one message. Every message is checked
// before any is applied, so a malformed one stops the whole payload; then
// they are applied in order, up to the first that fails. Each surface
// they changed is resolved again, once. An error also goes to the agent.
func (p *Processor) Process(payload any) error {
	msgs, err := messagesOf(payload)
	if err != nil {
		return p.fail(err, "")
	}
	if err := checkNesting(payload, 0, 0); err != nil {
		return p.fail(err, "")
	}
	ops, err := prepare(msgs)
	if err != nil {
		return p.fail(err, "")
	}
	var touched []string
	touch := func(id string) {
		if !slices.Contains(touched, id) {
			touched = append(touched, id)
		}
	}
	defer func() {
		for _, id := range touched {
			if s := p.surfaces[id]; s != nil {
				s.Tree.Resolve()
				if p.Changed != nil {
					p.Changed(s, false)
				}
			}
		}
	}()
	for _, o := range ops {
		if err := p.apply(o, touch); err != nil {
			return p.fail(err, str(o.body["surfaceId"]))
		}
	}
	return nil
}

// fail sends err to the agent, and returns it.
func (p *Processor) fail(err error, surfaceID string) error {
	m := &ErrorMessage{Code: Code(err), Message: strings.TrimPrefix(err.Error(), "a2ui: "), SurfaceID: surfaceID}
	if v, ok := err.(*ValidationError); ok {
		m.Path = v.Path
	}
	p.send(Outbound{Error: m})
	return err
}

func (p *Processor) send(o Outbound) {
	if p.Send != nil {
		p.Send(o)
	}
}

// op is one message's action, checked and ready to apply.
type op struct {
	action  string
	body    map[string]any
	version string
}

// messagesOf reads a payload as its list of messages.
func messagesOf(payload any) ([]map[string]any, error) {
	var list []any
	switch x := payload.(type) {
	case []any:
		list = x
	case map[string]any:
		if m, ok := x["messages"]; ok && x["version"] == nil {
			l, ok := m.([]any)
			if !ok {
				return nil, &ValidationError{Msg: "messages must be a list of messages", Path: "/messages"}
			}
			list = l
		} else {
			list = []any{x}
		}
	default:
		return nil, &ValidationError{Msg: fmt.Sprintf("a payload is a message, a list of messages or {\"messages\": [...]}, not %s", jsonType(payload))}
	}
	out := make([]map[string]any, len(list))
	for i, e := range list {
		m, ok := e.(map[string]any)
		if !ok {
			return nil, &ValidationError{Msg: fmt.Sprintf("message %d is %s, not an object", i, jsonType(e)), Path: "/" + strconv.Itoa(i)}
		}
		out[i] = m
	}
	return out, nil
}

// checkNesting refuses a payload nested deeper than MaxGlobalDepth, calls
// nested deeper than MaxCallDepth, and bindings whose path is not a
// pointer.
func checkNesting(v any, depth, calls int) error {
	if depth > MaxGlobalDepth {
		return &RecursionError{Msg: fmt.Sprintf("Global recursion limit exceeded: Depth > %d", MaxGlobalDepth)}
	}
	switch x := v.(type) {
	case map[string]any:
		if _, ok := x["@call"]; ok {
			if calls++; calls > MaxCallDepth {
				return &ValidationError{Msg: fmt.Sprintf("functionCall depth exceeds %d: calls nest in the arguments of calls too deeply", MaxCallDepth)}
			}
		}
		if p, ok := x["@path"].(string); ok {
			abs := p
			if !strings.HasPrefix(abs, "/") {
				abs = "/" + abs
			}
			if _, err := ParsePointer(abs); err != nil {
				return &ValidationError{Msg: fmt.Sprintf("Invalid path syntax: '%s'", p)}
			}
		}
		for _, k := range SortedKeys(x) {
			if err := checkNesting(x[k], depth+1, calls); err != nil {
				return err
			}
		}
	case []any:
		for _, e := range x {
			if err := checkNesting(e, depth+1, calls); err != nil {
				return err
			}
		}
	}
	return nil
}

// prepare checks every message: the payload's version, one action each,
// and the action's v1.0 schema.
func prepare(msgs []map[string]any) ([]op, error) {
	if len(msgs) == 0 {
		return nil, nil
	}
	version, err := payloadVersion(msgs[0])
	if err != nil {
		return nil, err
	}
	ops := make([]op, 0, len(msgs))
	for i, m := range msgs {
		where := fmt.Sprintf("message %d", i)
		var actions, others []string
		for _, k := range SortedKeys(m) {
			switch {
			case isAgentAction(k):
				actions = append(actions, k)
			case slices.Contains(otherActions, k):
				others = append(others, k)
			}
		}
		for _, k := range append(slices.Clone(actions), others...) {
			if body, ok := m[k].(map[string]any); ok {
				if id, ok := body["surfaceId"]; ok {
					if _, ok := id.(string); !ok {
						return nil, &ValidationError{Msg: fmt.Sprintf("%s: surfaceId must be a string, not %s", where, jsonType(id)), Path: "/" + k + "/surfaceId"}
					}
				}
			}
		}
		switch {
		case len(actions) > 1:
			return nil, &ValidationError{Msg: fmt.Sprintf("Message contains multiple conflicting update actions: %s.", strings.Join(actions, ", "))}
		case len(actions) == 0 && len(others) > 0:
			return nil, &ValidationError{Msg: fmt.Sprintf("Invalid v1.0 message: action '%s' is not supported in protocol version v1.0. Allowed actions: %s.", others[0], strings.Join(agentActions, ", "))}
		case len(actions) == 0:
			return nil, &ValidationError{Msg: fmt.Sprintf("Invalid v1.0 message: %s carries no action. Allowed actions: %s.", where, strings.Join(agentActions, ", "))}
		}
		action := actions[0]
		if err := checkSchema(envelopes()[action], m, fmt.Sprintf("Invalid v1.0 message: %s (%s)", where, action)); err != nil {
			return nil, err
		}
		body, _ := m[action].(map[string]any)
		if p, ok := body["path"].(string); ok && action == "updateDataModel" {
			if _, err := ParsePointer(p); err != nil {
				return nil, &ValidationError{Msg: fmt.Sprintf("Invalid path syntax: '%s' (%s)", p, strings.TrimPrefix(err.Error(), "a2ui: ")), Path: "/updateDataModel/path"}
			}
		}
		ops = append(ops, op{action: action, body: body, version: version})
	}
	return ops, nil
}

// payloadVersion reads a payload's protocol version from its first
// message: v1.0 is the one this renderer speaks.
func payloadVersion(m map[string]any) (string, error) {
	v, ok := m["version"]
	if !ok {
		for _, k := range otherActions {
			if _, ok := m[k]; ok {
				return "", &ValidationError{Msg: fmt.Sprintf("Unsupported protocol version 'v0.8' (a message with %s and no version); this renderer speaks %s", k, Version)}
			}
		}
		return "", &ValidationError{Msg: fmt.Sprintf("Missing required version field: a message needs \"version\": %q", Version)}
	}
	s, ok := v.(string)
	if !ok {
		return "", &ValidationError{Msg: fmt.Sprintf("Missing required version field: version must be a string, not %s", jsonType(v))}
	}
	if s != Version {
		return "", &ValidationError{Msg: fmt.Sprintf("Unsupported protocol version '%s'; this renderer speaks %s", s, Version)}
	}
	return s, nil
}

func (p *Processor) apply(o op, touch func(string)) error {
	id := str(o.body["surfaceId"])
	switch o.action {
	case "createSurface":
		if err := p.createSurface(o); err != nil {
			return err
		}
		touch(id)
	case "updateComponents":
		s := p.surfaces[id]
		if s == nil {
			return &IntegrityError{Msg: fmt.Sprintf("Surface not found for message: %s", id)}
		}
		if err := p.updateComponents(s, defsOf(o.body["components"])); err != nil {
			return err
		}
		touch(id)
	case "updateDataModel":
		s := p.surfaces[id]
		if s == nil {
			return &IntegrityError{Msg: fmt.Sprintf("Surface not found for message: %s", id)}
		}
		path, ok := o.body["path"].(string)
		if !ok {
			path = "/"
		}
		var err error
		if v := o.body["value"]; v == nil {
			err = s.Data.Delete(path)
		} else {
			err = s.Data.Set(path, v)
		}
		if err != nil {
			return err
		}
		touch(id)
	case "deleteSurface":
		s := p.surfaces[id]
		if s == nil {
			return nil
		}
		s.Tree.Dispose()
		s.Data.Dispose()
		delete(p.surfaces, id)
		p.order = slices.DeleteFunc(p.order, func(x string) bool { return x == id })
		p.calls.forget(id)
		if p.Changed != nil {
			p.Changed(s, true)
		}
	case "callRendererFunction":
		p.callRendererFunction(o)
	case "agentFunctionResponse":
		if c := p.calls.answer(o.body); c != nil && c.cached {
			touch(c.SurfaceID)
		}
	}
	return nil
}

func (p *Processor) createSurface(o op) error {
	id := str(o.body["surfaceId"])
	catID, named := o.body["catalogId"].(string)
	var cat *Catalog
	if named {
		if cat = p.Catalog(catID); cat == nil {
			return &CatalogError{Msg: "Catalog not found: " + catID}
		}
	} else {
		for _, c := range p.catalogs {
			if CompatibleVersions(c.ProtocolVersion, o.version) {
				cat = c
				break
			}
		}
		if cat == nil {
			return &CatalogError{Msg: fmt.Sprintf("Catalog not found: surface '%s' names none, and no catalog of %s is supported", id, o.version)}
		}
	}
	if !CompatibleVersions(cat.ProtocolVersion, o.version) {
		return &ValidationError{Msg: fmt.Sprintf("Catalog '%s' specification version (%s) does not match message protocol version (%s).", cat.ID, canonicalVersion(cat.ProtocolVersion), o.version)}
	}
	if p.surfaces[id] != nil {
		return &IntegrityError{Msg: fmt.Sprintf("Surface %s already exists.", id)}
	}
	available := CatalogSet{}
	for _, c := range p.catalogs {
		if CompatibleVersions(c.ProtocolVersion, cat.ProtocolVersion) {
			available[c.ID] = c
		}
	}
	s := NewSurface(id, cat.ID, available, p.env())
	s.SendDataModel = o.body["sendDataModel"] == true
	s.Metadata, _ = o.body["metadata"].(map[string]any)
	if dm, ok := o.body["dataModel"]; ok {
		if err := s.Data.Set("/", dm); err != nil {
			return err
		}
	}
	if comps, ok := o.body["components"]; ok {
		if err := p.updateComponents(s, defsOf(comps)); err != nil {
			return err
		}
	}
	p.surfaces[id] = s
	p.order = append(p.order, id)
	return nil
}

func (p *Processor) env() Env {
	return Env{
		Locale:  p.Locale,
		OpenURL: p.OpenURL,
		Agent:   p.agentFunction,
		Action:  func(_ *Surface, a *ActionMessage) { p.send(Outbound{Action: a}) },
		Error:   func(_ *Surface, e *ErrorMessage) { p.send(Outbound{Error: e}) },
	}
}

func defsOf(v any) []map[string]any {
	l, _ := v.([]any)
	out := make([]map[string]any, 0, len(l))
	for _, e := range l {
		if m, ok := e.(map[string]any); ok {
			out = append(out, m)
		}
	}
	return out
}

// updateComponents checks a batch of component definitions against the
// surface, then applies it: nothing changes when any is wrong.
func (p *Processor) updateComponents(s *Surface, defs []map[string]any) error {
	seen := map[string]bool{}
	for _, d := range defs {
		id := str(d["id"])
		if seen[id] {
			return &IntegrityError{Msg: fmt.Sprintf("Duplicate component ID: '%s'", id)}
		}
		seen[id] = true
	}
	for _, d := range defs {
		if err := p.checkComponent(s, d); err != nil {
			return err
		}
	}
	g := graphOf(s, defs)
	if err := g.checkComposition(); err != nil {
		return err
	}
	if p.Strict {
		if err := g.checkTopology(); err != nil {
			return err
		}
	}
	s.Update(defs)
	return nil
}

// checkComponent checks one definition: its catalog is the surface's to
// use, its type is the catalog's (when strict), it has the type's schema,
// and its values use no reserved key.
func (p *Processor) checkComponent(s *Surface, d map[string]any) error {
	id, typ := str(d["id"]), str(d["component"])
	catID := str(d["catalogId"])
	if catID == "" {
		catID = s.Catalog
	}
	cat := s.catalogs[catID]
	if cat == nil {
		if o := p.Catalog(catID); o != nil {
			return &CatalogError{Msg: fmt.Sprintf("Component '%s' catalog '%s' specification version (%s) does not match surface default catalog version (%s).",
				id, catID, canonicalVersion(o.ProtocolVersion), canonicalVersion(s.catalogs[s.Catalog].ProtocolVersion))}
		}
		return &CatalogError{Msg: fmt.Sprintf("Catalog '%s' is not supported by surface '%s'.", catID, s.ID)}
	}
	if typ == "" {
		if old := s.Components[id]; old != nil {
			typ = old.Type
		}
	}
	if typ == "Surface" {
		return &ValidationError{Msg: fmt.Sprintf("Component '%s': Surface is the surface's own container, not a component", id)}
	}
	if cat.Component(typ) == nil {
		if p.Strict {
			return &ValidationError{Msg: fmt.Sprintf("Component '%s' has type '%s', which catalog '%s' does not have", id, typ, catID)}
		}
		return nil
	}
	for _, k := range SortedKeys(d) {
		if k == "id" || k == "component" || k == "catalogId" {
			continue
		}
		if at, key, ok := findReservedKey(d[k], "/"+EscapeToken(k)); ok {
			return &ValidationError{Msg: fmt.Sprintf("Component '%s': Unrecognized reserved protocol directive '%s' at %s: keys starting with a single '@' are reserved; write '@%s' for a literal key", id, key, at, key), Path: at}
		}
	}
	return cat.CheckComponent(d)
}

func findReservedKey(v any, at string) (string, string, bool) {
	switch x := v.(type) {
	case map[string]any:
		for _, k := range SortedKeys(x) {
			if reservedKey(k) {
				return at, k, true
			}
			if a, key, ok := findReservedKey(x[k], at+"/"+EscapeToken(k)); ok {
				return a, key, true
			}
		}
	case []any:
		for i, e := range x {
			if a, key, ok := findReservedKey(e, at+"/"+strconv.Itoa(i)); ok {
				return a, key, true
			}
		}
	}
	return "", "", false
}

// graph is a surface's components as they would be after an update,
// with the components each refers to.
type graph struct {
	s     *Surface
	comps map[string]*Component
	ids   []string
	// refs are each component's children; templates the components its
	// templates repeat.
	refs, templates map[string][]string
}

func graphOf(s *Surface, defs []map[string]any) *graph {
	g := &graph{s: s, comps: map[string]*Component{}, refs: map[string][]string{}, templates: map[string][]string{}}
	for id, c := range s.Components {
		g.comps[id] = c
	}
	for _, d := range defs {
		id := str(d["id"])
		c := &Component{ID: id, Type: str(d["component"]), Catalog: str(d["catalogId"]), Props: d}
		if c.Type == "" && g.comps[id] != nil {
			c.Type = g.comps[id].Type
		}
		g.comps[id] = c
	}
	g.ids = SortedKeys(g.comps)
	for _, id := range g.ids {
		c := g.comps[id]
		var refs, templates []string
		if t := s.ComponentType(c); t != nil {
			for _, k := range SortedKeys(c.Props) {
				if p := t.Props[k]; p != nil {
					childRefs(p, c.Props[k], &refs, &templates)
				}
			}
		}
		g.refs[id], g.templates[id] = refs, templates
	}
	return g
}

// children are what a component contains: its children and its
// templates' components.
func (g *graph) children(id string) []string {
	return append(slices.Clone(g.refs[id]), g.templates[id]...)
}

// checkComposition checks each type's allowedParents and allowedChildren
// where both ends of a reference exist; the root's parent is Surface.
func (g *graph) checkComposition() error {
	typeOf := func(id string) *ComponentType {
		if c := g.comps[id]; c != nil {
			return g.s.ComponentType(c)
		}
		return nil
	}
	if t := typeOf("root"); t != nil && t.AllowedParents != nil && !slices.Contains(t.AllowedParents, "Surface") {
		return &ValidationError{Code: "UNALLOWED_PARENT", Msg: fmt.Sprintf("Component 'root' (%s) cannot be placed under parent 'Surface' (Surface). Allowed parents: %s.", t.Name, pyList(t.AllowedParents))}
	}
	for _, id := range g.ids {
		pt := typeOf(id)
		if pt == nil {
			continue
		}
		for _, child := range g.children(id) {
			ct := typeOf(child)
			if ct == nil {
				continue
			}
			if ct.AllowedParents != nil && !slices.Contains(ct.AllowedParents, pt.Name) {
				return &ValidationError{Code: "UNALLOWED_PARENT", Msg: fmt.Sprintf("Component '%s' (%s) cannot be placed under parent '%s' (%s). Allowed parents: %s.", child, ct.Name, id, pt.Name, pyList(ct.AllowedParents))}
			}
			if pt.AllowedChildren != nil && !slices.Contains(pt.AllowedChildren, ct.Name) {
				return &ValidationError{Code: "UNALLOWED_CHILD", Msg: fmt.Sprintf("Container '%s' (%s) cannot contain child '%s' (%s). Allowed children: %s.", id, pt.Name, child, ct.Name, pyList(pt.AllowedChildren))}
			}
		}
	}
	return nil
}

// pyList writes names as A2UI's messages list them: ['Text', 'Button'].
func pyList(names []string) string {
	q := make([]string, len(names))
	for i, n := range names {
		q[i] = "'" + n + "'"
	}
	return "[" + strings.Join(q, ", ") + "]"
}

// checkTopology checks the components as one tree: identifiers for ids, a
// root, every reference to a component that exists, no cycle, no deeper
// than MaxDepth, and every component reachable from the root. A
// template's component may contain its own template, since its data ends
// the recursion.
func (g *graph) checkTopology() error {
	for _, id := range g.ids {
		if !IsIdentifier(id) {
			return &ValidationError{Msg: fmt.Sprintf("Component id '%s' must be a valid identifier (UAX #31: a letter or '_', then letters, digits or '_')", id)}
		}
	}
	if g.comps["root"] == nil {
		return &IntegrityError{Msg: fmt.Sprintf("Missing root component: surface '%s' has no component 'root'", g.s.ID)}
	}
	for _, id := range g.ids {
		for _, r := range g.children(id) {
			if g.comps[r] == nil {
				return &IntegrityError{Msg: fmt.Sprintf("Dangling reference: component '%s' references non-existent component '%s'", id, r)}
			}
		}
	}
	const (
		unseen = iota
		open
		done
	)
	state := map[string]int{}
	depth := map[string]int{}
	var stack []string
	var visit func(id string) error
	visit = func(id string) error {
		switch state[id] {
		case open:
			if stack[len(stack)-1] == id {
				return &IntegrityError{Msg: fmt.Sprintf("Self-reference detected: component '%s' contains itself", id)}
			}
			i := slices.Index(stack, id)
			return &IntegrityError{Msg: fmt.Sprintf("Circular reference detected. Circular component reference: %s", strings.Join(append(stack[i:], id), " → "))}
		case done:
			return nil
		}
		state[id] = open
		stack = append(stack, id)
		for _, r := range g.refs[id] {
			if err := visit(r); err != nil {
				return err
			}
			depth[id] = max(depth[id], depth[r]+1)
		}
		stack = stack[:len(stack)-1]
		state[id] = done
		return nil
	}
	for _, id := range append([]string{"root"}, g.ids...) {
		if err := visit(id); err != nil {
			return err
		}
	}
	if depth["root"] >= MaxDepth {
		return &RecursionError{Msg: fmt.Sprintf("Global recursion limit exceeded: components nest %d deep from root, more than %d", depth["root"]+1, MaxDepth)}
	}
	reached := map[string]bool{}
	var reach func(id string)
	reach = func(id string) {
		if reached[id] {
			return
		}
		reached[id] = true
		for _, r := range g.children(id) {
			reach(r)
		}
	}
	reach("root")
	for _, id := range g.ids {
		if !reached[id] {
			return &IntegrityError{Msg: fmt.Sprintf("Component '%s' is not reachable from root", id)}
		}
	}
	return nil
}

// childRefs collects the component ids a property refers to: children
// in refs, templates' components in templates.
func childRefs(p *Prop, v any, refs, templates *[]string) {
	switch p.Kind {
	case ChildRef:
		if id, ok := v.(string); ok {
			*refs = append(*refs, id)
		}
	case ChildList:
		switch x := v.(type) {
		case []any:
			for _, e := range x {
				if id, ok := e.(string); ok {
					*refs = append(*refs, id)
				}
			}
		case map[string]any:
			if id, ok := x["componentId"].(string); ok {
				*templates = append(*templates, id)
			}
		}
	case Array:
		if l, ok := v.([]any); ok {
			for _, e := range l {
				childRefs(p.Items, e, refs, templates)
			}
		}
	case Object:
		if m, ok := v.(map[string]any); ok {
			for _, k := range SortedKeys(m) {
				if fp := p.Fields[k]; fp != nil {
					childRefs(fp, m[k], refs, templates)
				}
			}
		}
	}
}

// callRendererFunction runs a function the agent called, and answers it.
func (p *Processor) callRendererFunction(o op) {
	id := str(o.body["functionCallId"])
	call, _ := o.body["callFunction"].(map[string]any)
	name, catID := str(call["@call"]), str(call["catalogId"])
	answer := func(v any, code, msg string) {
		r := &FunctionResponse{FunctionCallID: id, Value: v}
		if code != "" {
			r.Value, r.Error = nil, &ResponseError{Code: code, Message: msg}
		}
		p.send(Outbound{RendererFunctionResponse: r})
	}
	refuse := func(format string, a ...any) { answer(nil, "INVALID_FUNCTION_CALL", fmt.Sprintf(format, a...)) }
	cat := p.Catalog(catID)
	if cat == nil {
		refuse("Catalog not found: %s", catID)
		return
	}
	if !CompatibleVersions(cat.ProtocolVersion, o.version) {
		refuse("Catalog '%s' specification version (%s) does not match message protocol version (%s).", catID, canonicalVersion(cat.ProtocolVersion), o.version)
		return
	}
	f := cat.Functions[name]
	if f == nil {
		refuse("Function not found: %s", name)
		return
	}
	if f.AllowedCallers != AgentOnly && f.AllowedCallers != RendererOrAgent {
		refuse("Function '%s' cannot be called by agent (allowedCallers is %s).", name, f.AllowedCallers)
		return
	}
	if f.RequiresUserActivation && (p.Activated == nil || !p.Activated()) {
		refuse("Function '%s' requires user activation context to execute.", name)
		return
	}
	if err := cat.CheckCall(name, call); err != nil {
		refuse("%s", strings.TrimPrefix(err.Error(), "a2ui: "))
		return
	}
	if f.Impl == nil {
		answer(nil, "EXECUTION_ERROR", fmt.Sprintf("Function '%s' has no implementation in this renderer.", name))
		return
	}
	args, _ := Clone(call["args"]).(map[string]any)
	if args == nil {
		args = map[string]any{}
	}
	ctx := &Context{Data: NewDataModel(nil), Scope: RootScope, Funcs: CatalogSet{cat.ID: cat}, Catalog: cat.ID,
		Locale: p.Locale, OpenURL: p.OpenURL, Activation: f.RequiresUserActivation}
	v, err := f.Impl(ctx, args)
	if err != nil {
		msg := strings.TrimPrefix(err.Error(), "a2ui: ")
		if msg == "" {
			msg = "An error occurred during function execution."
		}
		answer(nil, "EXECUTION_ERROR", msg)
		return
	}
	answer(v, "", "")
}

// CompatibleVersions reports whether two protocol versions ("v1.0",
// "1.0", "v0.9") may be mixed: the same major version from 1 on, the
// same minor one before it.
func CompatibleVersions(a, b string) bool {
	am, an, ok1 := parseVersion(a)
	bm, bn, ok2 := parseVersion(b)
	if !ok1 || !ok2 || am != bm {
		return false
	}
	return am >= 1 || an == bn
}

func parseVersion(v string) (major, minor int, ok bool) {
	v = strings.TrimPrefix(v, "v")
	parts := strings.SplitN(v, ".", 3)
	if len(parts) < 2 {
		return 0, 0, false
	}
	var err1, err2 error
	major, err1 = strconv.Atoi(parts[0])
	minor, err2 = strconv.Atoi(parts[1])
	return major, minor, err1 == nil && err2 == nil
}

// canonicalVersion writes a version as messages do: "1.0" is "v1.0".
func canonicalVersion(v string) string {
	if v != "" && !strings.HasPrefix(v, "v") {
		return "v" + v
	}
	return v
}

func jsonType(v any) string {
	switch v.(type) {
	case nil:
		return "null"
	case bool:
		return "a boolean"
	case float64, int, json.Number:
		return "a number"
	case string:
		return "a string"
	case []any:
		return "a list"
	case map[string]any:
		return "an object"
	}
	return fmt.Sprintf("%T", v)
}
