package a2ui

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
)

// The limits of the expression parser behind formatString. They are the
// same in every A2UI engine (web_core's expression_parser.ts), so that an
// expression one accepts the others do too.
const (
	// MaxExpressionDepth bounds nesting: interpolations within
	// interpolations, and calls within arguments.
	MaxExpressionDepth = 100
	// MaxTemplateLength is a template's length in UTF-16 code units.
	MaxTemplateLength = 10_000
	// MaxTemplateParts bounds the parts of one template.
	MaxTemplateParts = 1_000
)

// Binding is a reference to the data model in an expression: a JSON
// Pointer, absolute or relative to the scope it is evaluated in.
type Binding struct{ Path string }

// Call is a function call in an expression. Each argument is an
// expression: a literal (string, float64 or bool), a Binding or a Call.
type Call struct {
	Name string
	Args map[string]any
}

// ExpressionError is a template or an expression that does not parse.
type ExpressionError struct{ Msg string }

func (e *ExpressionError) Error() string { return "a2ui: expression: " + e.Msg }

func exprErr(format string, a ...any) error {
	return &ExpressionError{Msg: fmt.Sprintf(format, a...)}
}

// ParseTemplate parses a formatString template into its parts, in order:
// literal text (string), and the values of its ${…} expressions (string,
// float64, bool, Binding or Call). Empty parts are dropped, so "" parses to
// no parts and ${null} to nothing. \${ is a literal "${".
func ParseTemplate(s string) ([]any, error) {
	return parseTemplate(s, 0)
}

func parseTemplate(s string, depth int) ([]any, error) {
	if depth > MaxExpressionDepth {
		return nil, exprErr("Max recursion depth reached in parse")
	}
	if n := utf16Len(s); n > MaxTemplateLength {
		return nil, exprErr("Expression template length (%d) exceeds maximum limit (%d)", n, MaxTemplateLength)
	}
	if s == "" {
		return nil, nil
	}
	if !strings.Contains(s, "${") {
		return []any{s}, nil
	}
	var parts []any
	sc := &scanner{in: s}
	for !sc.end() {
		if len(parts) >= MaxTemplateParts {
			return nil, exprErr("Expression parts count exceeds maximum limit (%d)", MaxTemplateParts)
		}
		switch {
		case sc.has("${"):
			sc.pos += 2
			content, err := sc.interpolation()
			if err != nil {
				return nil, err
			}
			v, err := parseExpression(content, depth+1)
			if err != nil {
				return nil, err
			}
			parts = append(parts, v)
		case sc.has(`\${`):
			sc.pos += 3
			parts = append(parts, "${")
		default:
			start := sc.pos
			for !sc.end() && !sc.has("${") && !sc.has(`\${`) {
				sc.next()
			}
			parts = append(parts, s[start:sc.pos])
		}
	}
	out := parts[:0]
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return out, nil
}

// ParseExpression parses the inside of one ${…}.
func ParseExpression(s string) (any, error) { return parseExpression(s, 0) }

func parseExpression(s string, depth int) (any, error) {
	s = strings.TrimFunc(s, unicode.IsSpace)
	if s == "" {
		return "", nil
	}
	sc := &scanner{in: s}
	v, err := sc.expression(depth)
	if err != nil {
		return nil, err
	}
	if !sc.end() {
		return nil, exprErr("Unexpected characters at end of expression: '%s'", s[sc.pos:])
	}
	return v, nil
}

type scanner struct {
	in  string
	pos int
}

func (sc *scanner) end() bool { return sc.pos >= len(sc.in) }

func (sc *scanner) has(p string) bool { return strings.HasPrefix(sc.in[sc.pos:], p) }

// peek is the rune off runes ahead, or 0 past the end.
func (sc *scanner) peek(off int) rune {
	p := sc.pos
	for ; off > 0 && p < len(sc.in); off-- {
		_, n := utf8.DecodeRuneInString(sc.in[p:])
		p += n
	}
	if p >= len(sc.in) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(sc.in[p:])
	return r
}

func (sc *scanner) next() rune {
	if sc.end() {
		sc.pos++
		return 0
	}
	r, n := utf8.DecodeRuneInString(sc.in[sc.pos:])
	sc.pos += n
	return r
}

func (sc *scanner) skipSpace() {
	for !sc.end() && unicode.IsSpace(sc.peek(0)) {
		sc.next()
	}
}

// interpolation reads up to the } that closes the ${ just read, skipping
// quoted strings, and returns what is between them.
func (sc *scanner) interpolation() (string, error) {
	start := sc.pos
	balance := 1
	for !sc.end() && balance > 0 {
		switch c := sc.next(); c {
		case '{':
			balance++
		case '}':
			balance--
		case '\'', '"':
			for !sc.end() {
				d := sc.next()
				if d == '\\' {
					sc.next()
				} else if d == c {
					break
				}
			}
		}
	}
	if balance > 0 {
		return "", exprErr("Unclosed interpolation: missing '}'")
	}
	return sc.in[start : sc.pos-1], nil
}

func (sc *scanner) expression(depth int) (any, error) {
	if depth > MaxExpressionDepth {
		return nil, exprErr("Max recursion depth reached in parse")
	}
	sc.skipSpace()
	if sc.end() {
		return "", nil
	}
	if sc.has("${") {
		sc.pos += 2
		content, err := sc.interpolation()
		if err != nil {
			return nil, err
		}
		return parseExpression(content, depth+1)
	}
	if c := sc.peek(0); c == '\'' || c == '"' {
		return sc.stringLiteral(), nil
	}
	if sc.numberStart() {
		return sc.number()
	}
	if sc.keyword("true") {
		return true, nil
	}
	if sc.keyword("false") {
		return false, nil
	}
	if sc.keyword("null") {
		return "", nil
	}
	tok := sc.pathOrIdent()
	sc.skipSpace()
	if sc.peek(0) == '(' {
		return sc.call(tok, depth)
	}
	if tok == "" {
		return "", nil
	}
	return Binding{Path: tok}, nil
}

func (sc *scanner) pathOrIdent() string {
	start := sc.pos
	for !sc.end() {
		c := sc.peek(0)
		if !isIDContinue(c) && c != '/' && c != '.' && c != '-' {
			break
		}
		sc.next()
	}
	return sc.in[start:sc.pos]
}

func (sc *scanner) ident() string {
	start := sc.pos
	for !sc.end() && isIDContinue(sc.peek(0)) {
		sc.next()
	}
	return sc.in[start:sc.pos]
}

func (sc *scanner) call(name string, depth int) (any, error) {
	sc.next() // (
	sc.skipSpace()
	args := map[string]any{}
	for !sc.end() && sc.peek(0) != ')' {
		arg := sc.ident()
		sc.skipSpace()
		if sc.peek(0) != ':' {
			return nil, exprErr("Expected ':' after argument name '%s' in function '%s'", arg, name)
		}
		sc.next()
		sc.skipSpace()
		v, err := sc.expression(depth + 1)
		if err != nil {
			return nil, err
		}
		args[arg] = v
		sc.skipSpace()
		if sc.peek(0) == ',' {
			sc.next()
			sc.skipSpace()
		}
	}
	if sc.peek(0) != ')' {
		return nil, exprErr("Expected ')' after function arguments for '%s'", name)
	}
	sc.next()
	return Call{Name: name, Args: args}, nil
}

func (sc *scanner) stringLiteral() string {
	quote := sc.next()
	var b strings.Builder
	for !sc.end() {
		c := sc.next()
		switch {
		case c == '\\':
			switch d := sc.next(); d {
			case 'n':
				b.WriteByte('\n')
			case 't':
				b.WriteByte('\t')
			case 'r':
				b.WriteByte('\r')
			case 0:
			default:
				b.WriteRune(d)
			}
		case c == quote:
			return b.String()
		default:
			b.WriteRune(c)
		}
	}
	return b.String()
}

func (sc *scanner) keyword(k string) bool {
	if sc.has(k) && !isIDContinue(sc.peek(utf8.RuneCountInString(k))) {
		sc.pos += len(k)
		return true
	}
	return false
}

func isDigit(c rune) bool { return c >= '0' && c <= '9' }

func (sc *scanner) numberStart() bool {
	off := 0
	if c := sc.peek(0); c == '-' || c == '+' {
		off = 1
	}
	if isDigit(sc.peek(off)) {
		return true
	}
	return sc.peek(off) == '.' && isDigit(sc.peek(off+1))
}

var numberLiteral = regexp.MustCompile(`^[+-]?(?:[0-9]+\.?[0-9]*|\.[0-9]+)(?:[eE][+-]?[0-9]+)?$`)

func (sc *scanner) number() (any, error) {
	start := sc.pos
	if c := sc.peek(0); c == '-' || c == '+' {
		sc.next()
	}
	for !sc.end() && (isDigit(sc.peek(0)) || sc.peek(0) == '.') {
		sc.next()
	}
	if c := sc.peek(0); c == 'e' || c == 'E' {
		sc.next()
		if c := sc.peek(0); c == '+' || c == '-' {
			sc.next()
		}
		for !sc.end() && isDigit(sc.peek(0)) {
			sc.next()
		}
	}
	text := sc.in[start:sc.pos]
	if !numberLiteral.MatchString(text) {
		return nil, exprErr("Invalid number literal: '%s'", text)
	}
	f, err := strconv.ParseFloat(text, 64)
	if err != nil && !isUnderflow(f) || math.IsInf(f, 0) {
		return nil, exprErr("Number literal is out of range: '%s'", text)
	}
	if f == 0 {
		f = 0 // -0 is 0
	}
	return f, nil
}

// isUnderflow tells a literal too small for a float64, which strconv
// reports as a range error with 0 and which JavaScript reads as 0, from one
// too large.
func isUnderflow(f float64) bool { return f == 0 }

func utf16Len(s string) int {
	n := 0
	for _, r := range s {
		n += utf16.RuneLen(r)
	}
	return n
}

// isIDContinue approximates Unicode's XID_Continue with Go's tables:
// ID_Start (letters, letter numbers, Other_ID_Start) and the marks, digits
// and connector punctuation ID_Continue adds, less Pattern_Syntax and
// Pattern_White_Space.
func isIDContinue(r rune) bool {
	if r == 0 {
		return false
	}
	if unicode.Is(unicode.Pattern_Syntax, r) || unicode.Is(unicode.Pattern_White_Space, r) {
		return false
	}
	return unicode.IsLetter(r) || unicode.Is(unicode.Nl, r) || unicode.Is(unicode.Other_ID_Start, r) ||
		unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Mc, r) || unicode.Is(unicode.Nd, r) ||
		unicode.Is(unicode.Pc, r) || unicode.Is(unicode.Other_ID_Continue, r)
}
