package a2ui

import (
	"fmt"
	"sort"
	"strings"
)

// Surface is one A2UI surface as the renderer holds it: its components (the
// adjacency list, as the agent sent it), its data model, and the node tree
// resolved from them.
type Surface struct {
	ID string
	// Catalog is the surface's default catalog id: components and calls
	// that name none use it.
	Catalog       string
	SendDataModel bool
	// Metadata is createSurface's metadata: {"extensions": {…}}.
	Metadata   map[string]any
	Data       *DataModel
	Components map[string]*Component
	// Tree is the surface's resolved node tree.
	Tree *Resolver

	catalogs CatalogSet
	env      Env
}

// Component is one component definition: its id, its type, the catalog it
// comes from when it names one, and its other properties as sent. An
// update of the same type changes Props in place, so the Component (and
// the nodes made from it) keeps its identity; a change of type makes a new
// Component.
type Component struct {
	ID      string
	Type    string
	Catalog string
	Props   map[string]any
}

// Env is what a surface's evaluation needs from the renderer around it.
type Env struct {
	Locale  string
	OpenURL func(url string) error
	// Agent runs a function the renderer does not implement.
	Agent func(s *Surface, f *Function, catalog string, args map[string]any) (any, error)
	// Action receives an action the user started, for the agent.
	Action func(s *Surface, a *ActionMessage)
	// Error receives a runtime error to report to the agent.
	Error func(s *Surface, e *ErrorMessage)
}

// CatalogSet is the catalogs a renderer supports, by id.
type CatalogSet map[string]*Catalog

// Function finds a function: in the named catalog, else an error.
func (cs CatalogSet) Function(catalog, name string) (*Function, error) {
	c, ok := cs[catalog]
	if !ok {
		if catalog == "" {
			return nil, &CatalogError{Msg: "function '" + name + "' names no catalog, and the surface has no default"}
		}
		return nil, &CatalogError{Msg: "catalog '" + catalog + "' is not supported here"}
	}
	f, ok := c.Functions[name]
	if !ok {
		// Not the renderer's: the agent's, by A2UI's fallback routing.
		return &Function{Name: name, ReturnType: "any", AllowedCallers: RendererOrAgent}, nil
	}
	return f, nil
}

// NewSurface makes a surface over catalogs, with an empty data model and no
// components.
func NewSurface(id, catalog string, catalogs CatalogSet, env Env) *Surface {
	s := &Surface{
		ID:         id,
		Catalog:    catalog,
		Data:       NewDataModel(nil),
		Components: map[string]*Component{},
		catalogs:   catalogs,
		env:        env,
	}
	s.Tree = newResolver(s)
	return s
}

// Context is the evaluation context for a value of this surface in a
// scope.
func (s *Surface) Context(scope Scope) *Context {
	c := &Context{
		Data:    s.Data,
		Scope:   scope,
		Funcs:   s.catalogs,
		Catalog: s.Catalog,
		Locale:  s.env.Locale,
		OpenURL: s.env.OpenURL,
	}
	if s.env.Agent != nil {
		c.Agent = func(f *Function, catalog string, args map[string]any) (any, error) {
			return s.env.Agent(s, f, catalog, args)
		}
	}
	return c
}

// ComponentType finds a component's type in its catalog: nil when the
// catalog or the type is unknown.
func (s *Surface) ComponentType(c *Component) *ComponentType {
	id := c.Catalog
	if id == "" {
		id = s.Catalog
	}
	cat, ok := s.catalogs[id]
	if !ok {
		return nil
	}
	return cat.Component(c.Type)
}

// CatalogOf is the catalog a component resolves against.
func (s *Surface) CatalogOf(c *Component) string {
	if c.Catalog != "" {
		return c.Catalog
	}
	return s.Catalog
}

// ComponentIDs lists the surface's component ids, sorted.
func (s *Surface) ComponentIDs() []string {
	ids := make([]string, 0, len(s.Components))
	for id := range s.Components {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// Write writes a value to the data model, as a user's input does, and
// resolves the tree again.
func (s *Surface) Write(path string, v any) error {
	var err error
	if v == nil {
		err = s.Data.Delete(path)
	} else {
		err = s.Data.Set(path, v)
	}
	s.Tree.Resolve()
	return err
}

// Update applies component definitions, in order, as updateComponents
// does: a new id adds a component; an existing one of the same type has
// its properties replaced (not merged), and keeps its identity; a change
// of type replaces the component. It does not resolve the tree.
func (s *Surface) Update(defs []map[string]any) {
	for _, d := range defs {
		id, _ := d["id"].(string)
		typ, _ := d["component"].(string)
		cat, _ := d["catalogId"].(string)
		props := make(map[string]any, len(d))
		for k, v := range d {
			if k != "id" && k != "component" && k != "catalogId" {
				props[k] = v
			}
		}
		old := s.Components[id]
		if old != nil && (typ == "" || typ == old.Type) && cat == old.Catalog {
			old.Props = props
			continue
		}
		if typ == "" && old != nil {
			typ = old.Type
		}
		s.Components[id] = &Component{ID: id, Type: typ, Catalog: cat, Props: props}
	}
}

// Remove removes a component definition. It does not resolve the tree.
func (s *Surface) Remove(id string) { delete(s.Components, id) }

// Definition is a component as the agent sent it: its properties with id,
// component and catalogId.
func (c *Component) Definition() map[string]any {
	d := make(map[string]any, len(c.Props)+3)
	for k, v := range c.Props {
		d[k] = v
	}
	d["id"] = c.ID
	d["component"] = c.Type
	if c.Catalog != "" {
		d["catalogId"] = c.Catalog
	}
	return d
}

// Dispatch runs an Action that source, in scope, started: an event goes to
// the agent as an action message, its context and userMessage resolved
// now; a functionCall runs (with the user's activation when activation is
// set), and the tree is resolved again. An error also goes to the agent.
func (s *Surface) Dispatch(action map[string]any, scope Scope, source string, activation bool) error {
	ctx := s.Context(scope)
	ctx.Activation = activation
	ctx.Caller = source
	if fc, ok := action["functionCall"].(map[string]any); ok {
		_, err := ctx.Resolve(fc)
		if err != nil {
			s.Tree.report(err)
		}
		s.Tree.Resolve()
		return err
	}
	ev, ok := action["event"].(map[string]any)
	if !ok {
		err := &ValidationError{Msg: fmt.Sprintf("component '%s': an action is an event or a functionCall", source)}
		s.Tree.report(err)
		return err
	}
	a, err := s.ActionFor(ev, ctx, source)
	if err != nil {
		s.Tree.report(err)
		return err
	}
	if s.env.Action != nil {
		s.env.Action(s, a)
	}
	return nil
}

// ActionFor builds the action message an event sends, resolving its
// context and userMessage in ctx.
func (s *Surface) ActionFor(ev map[string]any, ctx *Context, source string) (*ActionMessage, error) {
	name, _ := ev["name"].(string)
	if strings.TrimSpace(name) == "" {
		return nil, &ValidationError{Msg: "an event action needs a name"}
	}
	a := &ActionMessage{
		Name:              name,
		SurfaceID:         s.ID,
		SourceComponentID: source,
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

// Env is what the surface's evaluation was given by the renderer.
func (s *Surface) Env() Env { return s.env }
