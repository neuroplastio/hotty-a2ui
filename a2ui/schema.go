package a2ui

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"

	thirdparty "github.com/neuroplastio/hotty-a2ui/third_party"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"github.com/santhosh-tekuri/jsonschema/v6/kind"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// A2UI's schemas refer to each other by these URLs. common_types.json
// finds "the active catalog" at catalog.json beside it, so each catalog
// is compiled on its own, under that URL.
const (
	specBase       = "https://a2ui.org/specification/v1_0/"
	commonTypesURL = specBase + "common_types.json"
	activeURL      = specBase + "catalog.json"
	envelopeURL    = specBase + "agent_to_renderer.json"
	// Schemas made here, to compile one component or function.
	localBase = "https://hotty-a2ui.invalid/"
)

// The actions an agent's message may carry in v1.0, with their schemas'
// names in agent_to_renderer.json.
var envelopeDefs = map[string]string{
	"createSurface":         "CreateSurfaceMessage",
	"updateComponents":      "UpdateComponentsMessage",
	"updateDataModel":       "UpdateDataModelMessage",
	"deleteSurface":         "DeleteSurfaceMessage",
	"callRendererFunction":  "CallRendererFunctionMessage",
	"agentFunctionResponse": "AgentFunctionResponseMessage",
}

// xid compiles a schema's pattern. Go's regexp has no \p{XID_Start} or
// \p{XID_Continue}, which Extensions keys use; their general categories
// stand in, inside a character class as A2UI writes them.
func xid(pattern string) (jsonschema.Regexp, error) {
	pattern = strings.NewReplacer(
		`\p{XID_Start}`, `\p{L}\p{Nl}`,
		`\p{XID_Continue}`, `\p{L}\p{Nl}\p{Mn}\p{Mc}\p{Nd}\p{Pc}`,
	).Replace(pattern)
	return regexp.Compile(pattern)
}

func newCompiler() *jsonschema.Compiler {
	c := jsonschema.NewCompiler()
	c.DefaultDraft(jsonschema.Draft2020)
	c.UseRegexpEngine(xid)
	if err := c.AddResource(commonTypesURL, mustDoc(thirdparty.CommonTypes)); err != nil {
		panic(err)
	}
	return c
}

func mustDoc(b []byte) any {
	var v any
	if err := json.Unmarshal(b, &v); err != nil {
		panic(err)
	}
	return v
}

// envelopes are the v1.0 message schemas, by action, compiled against an
// open catalog: a component's own properties are checked against its
// surface's catalog when the message is applied.
var envelopes = sync.OnceValue(func() map[string]*jsonschema.Schema {
	c := newCompiler()
	if err := c.AddResource(envelopeURL, mustDoc(thirdparty.AgentToRenderer)); err != nil {
		panic(err)
	}
	if err := c.AddResource(activeURL, openCatalog()); err != nil {
		panic(err)
	}
	out := map[string]*jsonschema.Schema{}
	for action, def := range envelopeDefs {
		out[action] = c.MustCompile(envelopeURL + "#/$defs/" + def)
	}
	return out
})

// openCatalog stands in for the active catalog where none is known: any
// component, any call of a function with arguments, other keys refused.
func openCatalog() map[string]any {
	return map[string]any{"$defs": map[string]any{
		"anyComponent": map[string]any{
			"required":             []any{"component"},
			"properties":           map[string]any{"component": map[string]any{"type": "string"}},
			"additionalProperties": true,
		},
		"anyFunction": map[string]any{
			"not":        map[string]any{"properties": map[string]any{"@call": map[string]any{"const": "@index"}}},
			"properties": map[string]any{"args": map[string]any{"type": "object"}},
		},
	}}
}

// catalogSchemas are a catalog's components and functions, compiled.
type catalogSchemas struct {
	once       sync.Once
	err        error
	components map[string]*jsonschema.Schema
	functions  map[string]*jsonschema.Schema
}

// schemas compiles the catalog's schemas, once: its document's, and those
// of the functions it was given without one. An open catalog has none.
func (c *Catalog) schemas() (*catalogSchemas, error) {
	s := &c.compiled
	s.once.Do(func() { s.err = s.compile(c) })
	return s, s.err
}

func (s *catalogSchemas) compile(c *Catalog) error {
	s.components = map[string]*jsonschema.Schema{}
	s.functions = map[string]*jsonschema.Schema{}
	if c.Open {
		return nil
	}
	doc, _ := deepCopy(c.Doc).(map[string]any)
	if doc == nil {
		doc = map[string]any{}
	}
	defs, _ := doc["$defs"].(map[string]any)
	if defs == nil {
		defs = map[string]any{}
	}
	doc = rewriteRefs(doc, defs).(map[string]any)
	doc["$id"] = activeURL
	funcs, _ := doc["functions"].(map[string]any)
	if funcs == nil {
		funcs = map[string]any{}
	}
	// Functions the renderer was given without a document (Implement on a
	// synthetic catalog) join the document's.
	for name, f := range c.Functions {
		if _, ok := funcs[name]; !ok && f.Schema != nil {
			funcs[name] = rewriteRefs(deepCopy(f.Schema), defs)
		}
	}
	doc["functions"] = funcs
	// A call is one of this catalog's functions, or names another catalog,
	// whose own schemas check it when it runs.
	alts := []any{map[string]any{
		"required": []any{"catalogId"},
		"properties": map[string]any{
			"catalogId": map[string]any{"not": map[string]any{"const": c.ID}},
			"args":      true,
		},
	}}
	for _, name := range SortedKeys(funcs) {
		alts = append(alts, map[string]any{"$ref": "#/functions/" + EscapeToken(name)})
	}
	defs["anyFunction"] = map[string]any{"anyOf": alts}
	doc["$defs"] = defs

	comp := newCompiler()
	if err := comp.AddResource(activeURL, doc); err != nil {
		return fmt.Errorf("a2ui: catalog %s: %w", c.ID, err)
	}
	for _, name := range c.ComponentNames() {
		url := localBase + "components/" + EscapeToken(name)
		err := comp.AddResource(url, map[string]any{
			"allOf": []any{
				map[string]any{"$ref": commonTypesURL + "#/$defs/ComponentCommon"},
				map[string]any{"$ref": activeURL + "#/components/" + EscapeToken(name)},
				map[string]any{"properties": map[string]any{"component": map[string]any{"type": "string"}}},
			},
			"unevaluatedProperties": false,
		})
		if err != nil {
			return err
		}
		sch, err := comp.Compile(url)
		if err != nil {
			return fmt.Errorf("a2ui: catalog %s: component %s: %w", c.ID, name, err)
		}
		s.components[name] = sch
	}
	for _, name := range SortedKeys(funcs) {
		url := localBase + "functions/" + EscapeToken(name)
		if err := comp.AddResource(url, map[string]any{"$ref": activeURL + "#/functions/" + EscapeToken(name)}); err != nil {
			return err
		}
		sch, err := comp.Compile(url)
		if err != nil {
			return fmt.Errorf("a2ui: catalog %s: function %s: %w", c.ID, name, err)
		}
		s.functions[name] = sch
	}
	return nil
}

// rewriteRefs points a catalog's references to common types at
// common_types.json's URL: "common_types.json#/$defs/X", any path that
// ends in common_types.json, and a bare "#/$defs/X" the catalog does not
// define itself.
func rewriteRefs(v any, defs map[string]any) any {
	switch x := v.(type) {
	case map[string]any:
		for k, e := range x {
			if ref, ok := e.(string); ok && k == "$ref" {
				x[k] = rewriteRef(ref, defs)
				continue
			}
			x[k] = rewriteRefs(e, defs)
		}
	case []any:
		for i, e := range x {
			x[i] = rewriteRefs(e, defs)
		}
	}
	return v
}

func rewriteRef(ref string, defs map[string]any) string {
	file, frag, _ := strings.Cut(ref, "#")
	switch {
	case file == "common_types.json" || strings.HasSuffix(file, "/common_types.json"):
		return commonTypesURL + "#" + frag
	case file == "" && strings.HasPrefix(frag, "/$defs/"):
		name := strings.TrimPrefix(frag, "/$defs/")
		if _, ok := defs[name]; !ok {
			return commonTypesURL + "#" + frag
		}
	}
	return ref
}

func deepCopy(v any) any {
	switch x := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(x))
		for k, e := range x {
			out[k] = deepCopy(e)
		}
		return out
	case []any:
		out := make([]any, len(x))
		for i, e := range x {
			out[i] = deepCopy(e)
		}
		return out
	}
	return v
}

// checkSchema validates v, returning a ValidationError that says where
// and what failed. what names the value: "component 'root' (Text)".
func checkSchema(sch *jsonschema.Schema, v any, what string) error {
	err := sch.Validate(v)
	if err == nil {
		return nil
	}
	var ve *jsonschema.ValidationError
	if !errors.As(err, &ve) {
		return &ValidationError{Msg: what + ": " + err.Error()}
	}
	at, msg := describe(ve)
	return &ValidationError{Msg: what + ": " + msg, Path: at}
}

// describe writes a schema failure as one line. An alternative (oneOf)
// fails in every form it was tried as, so it reports the failures at the
// deepest place in the value, where the value went wrong; a property no
// schema declares counts only when nothing else failed, since a failed
// branch leaves its own properties undeclared.
func describe(ve *jsonschema.ValidationError) (at, msg string) {
	type leaf struct {
		at      []string
		msg     string
		unknown bool
	}
	var leaves []leaf
	var walk func(e *jsonschema.ValidationError)
	walk = func(e *jsonschema.ValidationError) {
		if len(e.Causes) > 0 {
			for _, c := range e.Causes {
				walk(c)
			}
			return
		}
		l := leaf{at: e.InstanceLocation, msg: kindString(e)}
		if _, ok := e.ErrorKind.(*kind.FalseSchema); ok && len(e.InstanceLocation) > 0 {
			l.unknown = true
			l.msg = fmt.Sprintf("unknown property '%s'", e.InstanceLocation[len(e.InstanceLocation)-1])
		}
		leaves = append(leaves, l)
	}
	walk(ve)
	known := leaves[:0:0]
	for _, l := range leaves {
		if !l.unknown {
			known = append(known, l)
		}
	}
	if len(known) > 0 {
		leaves = known
	}
	deepest := 0
	for _, l := range leaves {
		deepest = max(deepest, len(l.at))
	}
	var msgs []string
	seen := map[string]bool{}
	for _, l := range leaves {
		if len(l.at) != deepest {
			continue
		}
		at = pointer(l.at)
		if l.unknown && deepest > 0 {
			at = pointer(l.at[:deepest-1])
		}
		if !seen[l.msg] {
			seen[l.msg] = true
			msgs = append(msgs, l.msg)
		}
	}
	sort.Strings(msgs)
	if len(msgs) > 4 {
		msgs = append(msgs[:4], "…")
	}
	return at, "at " + at + ": " + strings.Join(msgs, "; or ")
}

func kindString(e *jsonschema.ValidationError) string {
	return e.ErrorKind.LocalizedString(english)
}

var english = message.NewPrinter(language.English)

func pointer(segs []string) string {
	if len(segs) == 0 {
		return "/"
	}
	var b strings.Builder
	for _, s := range segs {
		b.WriteString("/")
		b.WriteString(EscapeToken(s))
	}
	return b.String()
}

// commonTypes are common_types.json's definitions, compiled on first use.
var commonTypes = struct {
	sync.Mutex
	c       *jsonschema.Compiler
	schemas map[string]*jsonschema.Schema
}{}

// CheckCommonType validates v as one of common_types.json's definitions
// ("Extensions", "DynamicValue"); one that refers to the active catalog
// sees an open one.
func CheckCommonType(name string, v any) error {
	commonTypes.Lock()
	defer commonTypes.Unlock()
	if commonTypes.c == nil {
		commonTypes.c = newCompiler()
		if err := commonTypes.c.AddResource(activeURL, openCatalog()); err != nil {
			return err
		}
		commonTypes.schemas = map[string]*jsonschema.Schema{}
	}
	sch := commonTypes.schemas[name]
	if sch == nil {
		var err error
		if sch, err = commonTypes.c.Compile(commonTypesURL + "#/$defs/" + EscapeToken(name)); err != nil {
			return fmt.Errorf("a2ui: common type %s: %w", name, err)
		}
		commonTypes.schemas[name] = sch
	}
	return checkSchema(sch, v, name)
}
