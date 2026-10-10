package a2ui

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// Catalog is an A2UI catalog as a renderer holds it: the document, its
// component types with their properties classified, and its functions with
// what the renderer implements of them.
type Catalog struct {
	ID              string
	ProtocolVersion string
	Title           string
	Instructions    string
	// Doc is the catalog document, as published.
	Doc        map[string]any
	Components map[string]*ComponentType
	Functions  map[string]*Function
	// Open accepts any component type with any properties, unchecked: a
	// stand-in for a catalog the renderer knows only by its id.
	Open bool
}

// ComponentType is one component of a catalog.
type ComponentType struct {
	Name   string
	Schema map[string]any
	// Props classifies each property the schema declares: what the node
	// resolver does with it.
	Props map[string]*Prop
	// Required lists the properties the schema requires, component aside.
	Required []string
	// AllowedParents and AllowedChildren constrain which component types
	// may contain this one, and which it may contain: nil when the schema
	// does not say, empty when none may. A surface's root has the parent
	// type Surface.
	AllowedParents  []string
	AllowedChildren []string
}

// PropKind is what a property holds, read from its schema.
type PropKind int

// The kinds of property.
const (
	// Static: a literal the renderer reads as it is (an enum, a number).
	Static PropKind = iota
	// Dynamic: a DynamicString, -Number, -Boolean, -StringList or
	// -Value, a DataBinding or a FunctionCall: resolved in the node's
	// scope, and writable when it is a binding.
	Dynamic
	// ChildRef: a ComponentId or Child, which becomes a node.
	ChildRef
	// ChildList: a ChildList, static or a template, which becomes nodes.
	ChildList
	// ActionProp: an Action, run when the user does what the component
	// says.
	ActionProp
	// Checks: a list of CheckRules, evaluated into ValidationResults.
	Checks
	// Array: a list whose items have the kind of Items.
	Array
	// Object: an object whose fields have the kinds of Fields.
	Object
)

// Prop is one property's classification.
type Prop struct {
	Kind   PropKind
	Items  *Prop
	Fields map[string]*Prop
	// Default is the schema's default, if it has one.
	Default any
	// Enum lists the values a Static string may take, if the schema
	// says.
	Enum []string
}

// Function is one function of a catalog.
type Function struct {
	Name string
	// ReturnType is string, number, boolean, array, object,
	// validationResult, any or void.
	ReturnType string
	// AllowedCallers is rendererOnly (the default), agentOnly or
	// rendererOrAgent.
	AllowedCallers         string
	RequiresUserActivation bool
	Schema                 map[string]any
	// Impl runs it in the renderer. A function without one is the agent's
	// (a callAgentFunction).
	Impl FuncImpl
}

// FuncImpl implements a function. args are resolved; ctx is where the
// call is evaluated.
type FuncImpl func(ctx *Context, args map[string]any) (any, error)

// The allowed callers of a function.
const (
	RendererOnly    = "rendererOnly"
	AgentOnly       = "agentOnly"
	RendererOrAgent = "rendererOrAgent"
)

// ParseCatalog reads a catalog document. Its components are classified
// from their schemas; its functions have no implementations until the
// renderer gives them (Implement).
func ParseCatalog(doc []byte) (*Catalog, error) {
	var d map[string]any
	if err := json.Unmarshal(doc, &d); err != nil {
		return nil, fmt.Errorf("a2ui: catalog: %w", err)
	}
	c := &Catalog{
		ID:              str(d["catalogId"]),
		ProtocolVersion: str(d["protocolVersion"]),
		Title:           str(d["title"]),
		Instructions:    str(d["instructions"]),
		Doc:             d,
		Components:      map[string]*ComponentType{},
		Functions:       map[string]*Function{},
	}
	if c.ID == "" {
		c.ID = str(d["$id"])
	}
	if c.ID == "" {
		return nil, &CatalogError{Msg: "catalog has no catalogId"}
	}
	if major, _, ok := parseVersion(c.ProtocolVersion); ok && major >= 1 {
		if err := checkIdentifiers(d); err != nil {
			return nil, err
		}
	}
	comps, _ := d["components"].(map[string]any)
	for name, s := range comps {
		schema, _ := s.(map[string]any)
		c.Components[name] = classifyComponent(name, schema)
	}
	funcs, _ := d["functions"].(map[string]any)
	for name, s := range funcs {
		schema, _ := s.(map[string]any)
		f := &Function{
			Name:                   name,
			ReturnType:             str(schema["returnType"]),
			AllowedCallers:         str(schema["allowedCallers"]),
			RequiresUserActivation: schema["requiresUserActivation"] == true,
			Schema:                 schema,
		}
		if f.AllowedCallers == "" {
			f.AllowedCallers = RendererOnly
		}
		c.Functions[name] = f
	}
	return c, nil
}

// NewOpenCatalog is an open catalog: any component, no functions.
func NewOpenCatalog(id, protocolVersion string) *Catalog {
	return &Catalog{
		ID:              id,
		ProtocolVersion: protocolVersion,
		Components:      map[string]*ComponentType{},
		Functions:       map[string]*Function{},
		Open:            true,
	}
}

// Component finds a component type: the catalog's, or, in an open
// catalog, one that takes any properties (child and children are its
// children, as A2UI's containers name them).
func (c *Catalog) Component(name string) *ComponentType {
	if t, ok := c.Components[name]; ok {
		return t
	}
	if !c.Open {
		return nil
	}
	t := &ComponentType{Name: name, Props: map[string]*Prop{
		"child":    {Kind: ChildRef},
		"children": {Kind: ChildList},
	}}
	for k, p := range commonProps {
		t.Props[k] = p
	}
	return t
}

// checkIdentifiers checks that a v1 catalog names its components, their
// properties, its functions and their arguments as UAX #31 identifiers.
// "@index" is A2UI's own.
func checkIdentifiers(d map[string]any) error {
	bad := func(kind, name string) error {
		return &CatalogError{Msg: fmt.Sprintf("Invalid UAX #31 %s identifier: '%s'", kind, name)}
	}
	var props func(s map[string]any) error
	props = func(s map[string]any) error {
		if ps, ok := s["properties"].(map[string]any); ok {
			for _, k := range SortedKeys(ps) {
				if !IsIdentifier(k) && !strings.HasPrefix(k, "@") {
					return bad("property", k)
				}
			}
		}
		for _, key := range []string{"allOf", "anyOf", "oneOf"} {
			alts, _ := s[key].([]any)
			for _, a := range alts {
				if m, ok := a.(map[string]any); ok {
					if err := props(m); err != nil {
						return err
					}
				}
			}
		}
		return nil
	}
	comps, _ := d["components"].(map[string]any)
	for _, name := range SortedKeys(comps) {
		if !IsIdentifier(name) {
			return bad("component", name)
		}
		if s, ok := comps[name].(map[string]any); ok {
			if err := props(s); err != nil {
				return err
			}
		}
	}
	funcs, _ := d["functions"].(map[string]any)
	for _, name := range SortedKeys(funcs) {
		if !IsIdentifier(name) && name != "@index" {
			return bad("function", name)
		}
		f, _ := funcs[name].(map[string]any)
		args, _ := f["parameters"].(map[string]any)
		if p, ok := f["properties"].(map[string]any); ok {
			if a, ok := p["args"].(map[string]any); ok {
				args, _ = a["properties"].(map[string]any)
			}
		}
		for _, k := range SortedKeys(args) {
			if !IsIdentifier(k) {
				return bad("argument", k)
			}
		}
	}
	return nil
}

// Implement gives the renderer's implementation of a function the catalog
// declares.
func (c *Catalog) Implement(name string, impl FuncImpl) error {
	f, ok := c.Functions[name]
	if !ok {
		return fmt.Errorf("a2ui: catalog %s has no function %s", c.ID, name)
	}
	f.Impl = impl
	return nil
}

// ComponentNames lists the catalog's components, sorted.
func (c *Catalog) ComponentNames() []string {
	names := make([]string, 0, len(c.Components))
	for n := range c.Components {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// commonProps are the properties every component has, from
// ComponentCommon, which the envelope adds to each catalog's schema.
var commonProps = map[string]*Prop{
	"accessibility": {Kind: Object, Fields: map[string]*Prop{
		"label":       {Kind: Dynamic},
		"description": {Kind: Dynamic},
		"live":        {Kind: Static},
		"hidden":      {Kind: Dynamic},
	}},
	"metadata": {Kind: Static},
}

func classifyComponent(name string, schema map[string]any) *ComponentType {
	t := &ComponentType{Name: name, Schema: schema, Props: map[string]*Prop{}}
	for k, p := range commonProps {
		t.Props[k] = p
	}
	t.AllowedParents = constraint(schema, "allowedParents")
	t.AllowedChildren = constraint(schema, "allowedChildren")
	var walk func(s map[string]any)
	walk = func(s map[string]any) {
		if props, ok := s["properties"].(map[string]any); ok {
			for k, ps := range props {
				if k == "component" {
					continue
				}
				m, _ := ps.(map[string]any)
				t.Props[k] = classify(m)
			}
		}
		for _, r := range strs(s["required"]) {
			if r != "component" {
				t.Required = append(t.Required, r)
			}
		}
		if all, ok := s["allOf"].([]any); ok {
			for _, a := range all {
				if m, ok := a.(map[string]any); ok {
					if ref := str(m["$ref"]); ref != "" && refName(ref) == "Checkable" {
						t.Props["checks"] = &Prop{Kind: Checks}
						continue
					}
					walk(m)
				}
			}
		}
	}
	walk(schema)
	sort.Strings(t.Required)
	return t
}

// refName is the definition a common_types reference names:
// "common_types.json#/$defs/Child" is "Child".
func refName(ref string) string {
	if i := strings.LastIndex(ref, "/"); i >= 0 {
		return ref[i+1:]
	}
	return ref
}

func classify(s map[string]any) *Prop {
	p := &Prop{Kind: Static, Default: s["default"], Enum: strs(s["enum"])}
	if ref := str(s["$ref"]); ref != "" {
		switch refName(ref) {
		case "ComponentId", "Child":
			// A2UI's topology counts every component id as an edge: a
			// property that names a component without containing it is a
			// plain string.
			p.Kind = ChildRef
		case "ChildList":
			p.Kind = ChildList
		case "Action":
			p.Kind = ActionProp
		case "DynamicString", "DynamicNumber", "DynamicBoolean", "DynamicStringList",
			"DynamicValue", "DataBinding", "FunctionCall":
			p.Kind = Dynamic
		case "CheckRule":
			p.Kind = Checks
		case "AccessibilityAttributes":
			return commonProps["accessibility"]
		}
		return p
	}
	for _, key := range []string{"oneOf", "anyOf", "allOf"} {
		alts, ok := s[key].([]any)
		if !ok {
			continue
		}
		for _, a := range alts {
			if m, ok := a.(map[string]any); ok {
				if c := classify(m); c.Kind == Dynamic {
					p.Kind = Dynamic
					return p
				}
			}
		}
	}
	switch str(s["type"]) {
	case "array":
		items, _ := s["items"].(map[string]any)
		ip := classify(items)
		if ip.Kind == Checks {
			return &Prop{Kind: Checks}
		}
		if ip.Kind != Static {
			p.Kind, p.Items = Array, ip
		}
	case "object":
		props, _ := s["properties"].(map[string]any)
		fields := map[string]*Prop{}
		dynamic := false
		for k, fs := range props {
			m, _ := fs.(map[string]any)
			fields[k] = classify(m)
			dynamic = dynamic || fields[k].Kind != Static
		}
		if dynamic {
			p.Kind, p.Fields = Object, fields
		}
	}
	return p
}

// constraint reads a composition constraint: nil when absent.
func constraint(schema map[string]any, key string) []string {
	if _, ok := schema[key].([]any); !ok {
		return nil
	}
	return append([]string{}, strs(schema[key])...)
}

func str(v any) string {
	s, _ := v.(string)
	return s
}

func strs(v any) []string {
	l, _ := v.([]any)
	var out []string
	for _, x := range l {
		if s, ok := x.(string); ok {
			out = append(out, s)
		}
	}
	return out
}
