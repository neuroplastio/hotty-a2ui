// Package basic is A2UI's basic catalog (v1) for this kit: the published
// catalog document, and the renderer's implementation of its functions.
//
// The functions behave as A2UI's other engines do (web_core and Dart's
// a2ui_core): the same coercions, the same messages, and Intl's output for
// en-US. Other locales format as en-US for now.
package basic

import (
	"fmt"
	"math"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf16"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	thirdparty "github.com/neuroplastio/hotty-a2ui/third_party"
)

// ID is the basic catalog's id.
const ID = "https://a2ui.org/specification/v1_0/catalogs/basic/catalog.json"

// Catalog returns a fresh copy of the basic catalog with every function
// implemented.
func Catalog() *a2ui.Catalog {
	c, err := a2ui.ParseCatalog(thirdparty.BasicCatalog)
	if err != nil {
		panic(err)
	}
	for name, impl := range Functions {
		if err := c.Implement(name, impl); err != nil {
			panic(err)
		}
	}
	for name, f := range c.Functions {
		if f.Impl == nil {
			panic("basic: no implementation of " + name)
		}
	}
	return c
}

// Functions are the basic catalog's functions.
var Functions = map[string]a2ui.FuncImpl{
	"required": validation(func(a map[string]any) a2ui.ValidationResult {
		v := a["value"]
		if v == nil || v == "" {
			return invalid("This field is required.")
		}
		if l, ok := v.([]any); ok && len(l) == 0 {
			return invalid("This field is required.")
		}
		return valid
	}),
	"regex": func(_ *a2ui.Context, a map[string]any) (any, error) {
		source := a2ui.ToString(a["pattern"])
		re, err := regexp.Compile(source)
		if err != nil {
			return nil, &a2ui.ExpressionError{Msg: fmt.Sprintf("Invalid regex pattern: %s", source)}
		}
		if re.MatchString(a2ui.ToString(a["value"])) {
			return valid, nil
		}
		return invalid("Value does not match required pattern."), nil
	},
	"length": validation(func(a map[string]any) a2ui.ValidationResult {
		n := 0
		switch v := a["value"].(type) {
		case string:
			n = len(utf16.Encode([]rune(v)))
		case []any:
			n = len(v)
		}
		if lo := a2ui.ToNumber(a["min"]); a["min"] != nil && !math.IsNaN(lo) && float64(n) < lo {
			return invalid("Minimum length is " + a2ui.NumberString(lo) + ".")
		}
		if hi := a2ui.ToNumber(a["max"]); a["max"] != nil && !math.IsNaN(hi) && float64(n) > hi {
			return invalid("Maximum length is " + a2ui.NumberString(hi) + ".")
		}
		return valid
	}),
	"numeric": validation(func(a map[string]any) a2ui.ValidationResult {
		n := a2ui.ToNumber(a["value"])
		if math.IsNaN(n) {
			return invalid("Value must be a valid number.")
		}
		if lo := a2ui.ToNumber(a["min"]); a["min"] != nil && !math.IsNaN(lo) && n < lo {
			return invalid("Minimum value is " + a2ui.NumberString(lo) + ".")
		}
		if hi := a2ui.ToNumber(a["max"]); a["max"] != nil && !math.IsNaN(hi) && n > hi {
			return invalid("Maximum value is " + a2ui.NumberString(hi) + ".")
		}
		return valid
	}),
	"email": validation(func(a map[string]any) a2ui.ValidationResult {
		if s, ok := a["value"].(string); ok && emailPattern.MatchString(s) {
			return valid
		}
		return a2ui.ValidationResult{Valid: false, Code: "INVALID_EMAIL", Message: "Please enter a valid email address (e.g. user@example.com).", Severity: "error"}
	}),
	"formatString": func(c *a2ui.Context, a map[string]any) (any, error) {
		return FormatString(c, a2ui.ToString(a["value"]))
	},
	"formatNumber": func(_ *a2ui.Context, a map[string]any) (any, error) {
		return FormatNumber(a["value"], a["decimals"], a["grouping"]), nil
	},
	"formatCurrency": func(_ *a2ui.Context, a map[string]any) (any, error) {
		return FormatCurrency(a["value"], a["currency"], a["decimals"], a["grouping"]), nil
	},
	"formatDate": func(_ *a2ui.Context, a map[string]any) (any, error) {
		return FormatDate(a["value"], a["format"]), nil
	},
	"pluralize": func(_ *a2ui.Context, a map[string]any) (any, error) {
		return Pluralize(a["value"], a), nil
	},
	"openUrl": func(c *a2ui.Context, a map[string]any) (any, error) {
		return nil, openURL(c, a["url"])
	},
	"and": func(_ *a2ui.Context, a map[string]any) (any, error) {
		vs, err := operands("and", a["values"])
		if err != nil {
			return nil, err
		}
		for _, v := range vs {
			if !a2ui.Truthy(v) {
				return false, nil
			}
		}
		return true, nil
	},
	"or": func(_ *a2ui.Context, a map[string]any) (any, error) {
		vs, err := operands("or", a["values"])
		if err != nil {
			return nil, err
		}
		for _, v := range vs {
			if a2ui.Truthy(v) {
				return true, nil
			}
		}
		return false, nil
	},
	"not": func(_ *a2ui.Context, a map[string]any) (any, error) {
		return !a2ui.Truthy(a["value"]), nil
	},
}

var (
	valid        = a2ui.ValidationResult{Valid: true}
	emailPattern = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
)

func invalid(msg string) a2ui.ValidationResult {
	return a2ui.ValidationResult{Valid: false, Message: msg}
}

func validation(rule func(map[string]any) a2ui.ValidationResult) a2ui.FuncImpl {
	return func(_ *a2ui.Context, a map[string]any) (any, error) { return rule(a), nil }
}

func operands(name string, v any) ([]any, error) {
	l, ok := v.([]any)
	if !ok || len(l) < 2 {
		return nil, &a2ui.ExpressionError{Msg: name + " requires at least 2 values"}
	}
	return l, nil
}

// FormatString interpolates a template in a context: each ${…} is
// evaluated and converted with a2ui.ToString.
func FormatString(c *a2ui.Context, template string) (string, error) {
	parts, err := a2ui.ParseTemplate(template)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	for _, p := range parts {
		v, err := evalPart(c, p)
		if err != nil {
			return "", err
		}
		b.WriteString(a2ui.ToString(v))
	}
	return b.String(), nil
}

// evalPart evaluates one parsed part of a template: a literal, a binding
// or a call, whose arguments are parts too.
func evalPart(c *a2ui.Context, p any) (any, error) {
	switch x := p.(type) {
	case a2ui.Binding:
		return c.Resolve(map[string]any{"@path": x.Path})
	case a2ui.Call:
		args := make(map[string]any, len(x.Args))
		for k, a := range x.Args {
			v, err := evalPart(c, a)
			if err != nil {
				return nil, err
			}
			args[k] = v
		}
		if x.Name == "@index" {
			return c.Resolve(map[string]any{"@call": "@index", "args": args})
		}
		return c.Invoke(c.Catalog, x.Name, args)
	}
	return p, nil
}

// allowedSchemes are the schemes openUrl opens, as web_core and Dart's
// core allow them.
var allowedSchemes = map[string]bool{"http": true, "https": true, "mailto": true, "tel": true}

func openURL(c *a2ui.Context, v any) error {
	s, ok := v.(string)
	if !ok || strings.TrimSpace(s) == "" {
		return nil
	}
	s = strings.TrimSpace(s)
	u, err := url.Parse(s)
	if err != nil {
		return &a2ui.ExpressionError{Msg: "Invalid URL specified: " + s}
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme == "" {
		return &a2ui.ExpressionError{Msg: "URL must be absolute: " + s}
	}
	if !allowedSchemes[scheme] {
		return &a2ui.ExpressionError{Msg: "Unsupported URL scheme: " + scheme + ":"}
	}
	if c == nil || c.OpenURL == nil {
		return &a2ui.ExpressionError{Msg: "openUrl has no handler"}
	}
	return c.OpenURL(s)
}
