package a2ui

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
)

// Scope is where a dynamic value is evaluated: the data path relative
// paths continue (A2UI's collection scope), and, inside a template, the
// item's index.
type Scope struct {
	// Path is absolute: "/" at the root, the item's path in a template.
	Path string
	// Index is the item's index in its template, when InTemplate.
	Index      int
	InTemplate bool
}

// RootScope is the scope of a surface's root.
var RootScope = Scope{Path: "/"}

// Context evaluates dynamic values: data bindings, function calls, and the
// literals around them.
type Context struct {
	Data  *DataModel
	Scope Scope
	// Funcs finds a function by catalog and name.
	Funcs FunctionFinder
	// Catalog is the surface's default catalog, for calls without one.
	Catalog string
	// Activation is set while a user's gesture runs an action: what
	// requires it (openUrl) refuses to run otherwise.
	Activation bool
	// Agent runs a function the renderer does not implement, as a
	// callAgentFunction. Without it such a call is an error.
	Agent func(f *Function, catalog string, args map[string]any) (any, error)
	// Locale formats numbers, currencies, dates and plurals: a BCP 47 tag,
	// en-US when empty.
	Locale string
	// OpenURL opens a URL for openUrl.
	OpenURL func(url string) error
	// Caller is the component that the value belongs to, if any.
	Caller string
}

// FunctionFinder finds a function: the catalog is the call's own, or the
// surface's default.
type FunctionFinder interface {
	Function(catalog, name string) (*Function, error)
}

// In is the context for another scope.
func (c *Context) In(s Scope) *Context {
	n := *c
	n.Scope = s
	return &n
}

// Path makes path absolute in the context's scope.
func (c *Context) Path(path string) string {
	scope := c.Scope.Path
	if scope == "" {
		scope = "/"
	}
	if path == "." {
		return scope
	}
	return Join(scope, path)
}

// ReservedKeyError is an object key that starts with a single @ and is
// not one of A2UI's directives (@path, @call).
type ReservedKeyError struct{ Key string }

func (e *ReservedKeyError) Error() string {
	return fmt.Sprintf("a2ui: Unrecognized reserved protocol directive '%s': keys starting with a single '@' are reserved; write '@%s' for a literal key", e.Key, e.Key)
}

// IsBinding reports whether v is a data binding, {"@path": …}, and
// returns its path.
func IsBinding(v any) (string, bool) {
	m, ok := v.(map[string]any)
	if !ok {
		return "", false
	}
	p, ok := m["@path"].(string)
	return p, ok
}

// IsCall reports whether v is a function call, {"@call": …}.
func IsCall(v any) bool {
	m, ok := v.(map[string]any)
	if !ok {
		return false
	}
	_, ok = m["@call"]
	return ok
}

// Resolve evaluates a dynamic value: a binding reads the data model (nil
// where there is nothing), a call runs its function, and objects and lists
// are resolved element by element, with "@@" keys unescaped to "@".
func (c *Context) Resolve(v any) (any, error) {
	switch x := v.(type) {
	case map[string]any:
		if p, ok := x["@path"]; ok {
			path, ok := p.(string)
			if !ok {
				return nil, fmt.Errorf("a2ui: @path must be a string, not %T", p)
			}
			return c.Data.Value(c.Path(path)), nil
		}
		if _, ok := x["@call"]; ok {
			return c.call(x)
		}
		out := make(map[string]any, len(x))
		for k, e := range x {
			if reservedKey(k) {
				return nil, &ReservedKeyError{Key: k}
			}
			r, err := c.Resolve(e)
			if err != nil {
				return nil, err
			}
			if strings.HasPrefix(k, "@@") {
				k = k[1:]
			}
			out[k] = r
		}
		return out, nil
	case []any:
		out := make([]any, len(x))
		for i, e := range x {
			r, err := c.Resolve(e)
			if err != nil {
				return nil, err
			}
			out[i] = r
		}
		return out, nil
	}
	return v, nil
}

// reservedKey: a key with a single leading @ (the directives aside).
func reservedKey(k string) bool {
	return strings.HasPrefix(k, "@") && !strings.HasPrefix(k, "@@") && k != "@path" && k != "@call"
}

// call runs {"@call": name, "args": {…}, "catalogId": …}.
func (c *Context) call(x map[string]any) (any, error) {
	name, ok := x["@call"].(string)
	if !ok {
		return nil, fmt.Errorf("a2ui: @call must be a string")
	}
	rawArgs, _ := x["args"].(map[string]any)
	args := make(map[string]any, len(rawArgs))
	for k, a := range rawArgs {
		r, err := c.Resolve(a)
		if err != nil {
			return nil, err
		}
		args[k] = r
	}
	if name == "@index" {
		return c.index(args)
	}
	catalog := str(x["catalogId"])
	if catalog == "" {
		catalog = c.Catalog
	}
	return c.Invoke(catalog, name, args)
}

// Invoke runs a function with resolved arguments: the renderer's own
// implementation, or the agent's when the renderer has none.
func (c *Context) Invoke(catalog, name string, args map[string]any) (any, error) {
	if c.Funcs == nil {
		return nil, fmt.Errorf("a2ui: no functions to call %s", name)
	}
	f, err := c.Funcs.Function(catalog, name)
	if err != nil {
		return nil, err
	}
	if f.AllowedCallers == AgentOnly {
		return nil, &ExpressionError{Msg: fmt.Sprintf("function '%s' is agentOnly: only the agent may call it", name)}
	}
	if f.Impl == nil {
		if c.Agent == nil {
			return nil, &ExpressionError{Msg: fmt.Sprintf("function '%s' has no implementation here", name)}
		}
		return c.Agent(f, catalog, args)
	}
	if f.RequiresUserActivation && !c.Activation {
		return nil, &ExpressionError{Msg: fmt.Sprintf("function '%s' requires a user's activation", name)}
	}
	return f.Impl(c, args)
}

func (c *Context) index(args map[string]any) (any, error) {
	if !c.Scope.InTemplate {
		return nil, &ValidationError{Msg: "@index function can only be evaluated inside a collection template iteration scope."}
	}
	off := 0.0
	if o, ok := args["offset"]; ok && o != nil {
		off = ToNumber(o)
		if math.IsNaN(off) {
			return nil, &ExpressionError{Msg: "@index: offset is not a number"}
		}
	}
	return float64(c.Scope.Index) + off, nil
}

// ValidationResult is what a check evaluates to.
type ValidationResult struct {
	Valid    bool   `json:"valid"`
	Code     string `json:"code,omitempty"`
	Message  string `json:"message,omitempty"`
	Severity string `json:"severity,omitempty"`
}

// ValidityOf reads a ValidationResult, or an object with a boolean
// "valid", as A2UI v1.0 does: ok is false for anything else.
func ValidityOf(v any) (valid, ok bool) {
	switch x := v.(type) {
	case ValidationResult:
		return x.Valid, true
	case *ValidationResult:
		return x.Valid, true
	case map[string]any:
		b, ok := x["valid"].(bool)
		return b, ok
	}
	return false, false
}

// Truthy is JavaScript's truthiness, which A2UI's logic functions use:
// false for nil, false, "", 0 and NaN. A ValidationResult is its
// validity.
func Truthy(v any) bool {
	if valid, ok := ValidityOf(v); ok {
		return valid
	}
	switch x := v.(type) {
	case nil:
		return false
	case bool:
		return x
	case string:
		return x != ""
	case float64:
		return x != 0 && !math.IsNaN(x)
	case int:
		return x != 0
	}
	return true
}

// ToNumber reads v as a number the way A2UI's functions do: booleans are 1
// and 0, a blank string 0, a numeric string its number, anything else NaN.
func ToNumber(v any) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case int:
		return float64(x)
	case bool:
		if x {
			return 1
		}
		return 0
	case string:
		s := strings.TrimSpace(x)
		if s == "" {
			return 0
		}
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return math.NaN()
		}
		return f
	}
	return math.NaN()
}

// ToString converts a value for interpolation, as A2UI specifies: nil is
// "", numbers and booleans their usual text (JavaScript's for numbers),
// objects and lists JSON.
func ToString(v any) string {
	switch x := v.(type) {
	case nil:
		return ""
	case string:
		return x
	case bool:
		return strconv.FormatBool(x)
	case float64:
		return NumberString(x)
	case int:
		return strconv.Itoa(x)
	case ValidationResult:
		b, _ := json.Marshal(x)
		return string(b)
	}
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprint(v)
	}
	return string(b)
}

// NumberString writes a number as JavaScript's Number.prototype.toString
// does: integers without a fraction, the shortest digits that read back,
// and an exponent from 1e21 up and below 1e-6.
func NumberString(f float64) string {
	switch {
	case math.IsNaN(f):
		return "NaN"
	case math.IsInf(f, 1):
		return "Infinity"
	case math.IsInf(f, -1):
		return "-Infinity"
	case f == 0:
		return "0"
	}
	abs := math.Abs(f)
	if abs >= 1e21 || abs < 1e-6 {
		s := strconv.FormatFloat(f, 'e', -1, 64)
		// Go writes e+21 and e-07; JavaScript e+21 and e-7.
		mant, exp, _ := strings.Cut(s, "e")
		sign := exp[0]
		exp = strings.TrimLeft(exp[1:], "0")
		return mant + "e" + string(sign) + exp
	}
	return strconv.FormatFloat(f, 'f', -1, 64)
}

// SortedKeys lists a map's keys in order.
func SortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
