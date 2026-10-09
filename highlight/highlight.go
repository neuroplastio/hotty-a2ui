// Package highlight is code as lines of tokens of a few kinds, for the
// kit's HottyCode and a Text's fenced code blocks (profile §6.12): chroma
// lexes it, and a rendition colours each kind with a role, so that every
// theme, and the terminal's own palette, colours code.
//
// The languages are chroma's lexers for what coding agents write most
// (lexers/, scripts/lexers.sh). A program that wants all of chroma's
// imports highlight/all.
package highlight

import (
	"strconv"
	"strings"
	"sync"

	"github.com/alecthomas/chroma/v2"
)

// Kind is what a token is, as far as its colour goes.
type Kind uint8

// The kinds, and the roles the renditions colour them with.
const (
	Plain    Kind = iota // fg: names, operators, punctuation
	Comment              // muted, italic
	Keyword              // info, bold
	Type                 // info: types, builtins, classes, tags
	Function             // fg, bold: a function's name
	String               // success
	Literal              // warning: numbers, constants, true, nil
	Meta                 // warning: preprocessor, decorators, attributes
	Inserted             // success: a diff's added line
	Deleted              // error: a diff's removed line
	Heading              // info, bold: a diff's hunk headers
	Invalid              // error: what the lexer could not lex
)

var kindNames = [...]string{"plain", "comment", "keyword", "type", "function", "string", "literal", "meta", "inserted", "deleted", "heading", "invalid"}

func (k Kind) String() string {
	if int(k) < len(kindNames) {
		return kindNames[k]
	}
	return "kind(" + strconv.Itoa(int(k)) + ")"
}

// MarshalText is the kind's name, as a view's JSON has it.
func (k Kind) MarshalText() ([]byte, error) { return []byte(k.String()), nil }

// Token is a run of code of one kind, within one line.
type Token struct {
	Text string `json:"text"`
	Kind Kind   `json:"kind"`
}

// TabWidth is the columns between tab stops: a tab is spaces to the next
// one, so that both renditions indent alike.
const TabWidth = 4

// Lines is code as lines of tokens, lexed as lang says: a language's name
// or alias ("go", "Python", "ts"), or a file name ("main.go"). With no
// lexer for lang, each line is one Plain token. A line holds no "\n", and
// tabs are spaces (TabWidth); a final newline ends the last line rather
// than starting another. The lines are shared: a caller does not change
// them.
func Lines(code, lang string) [][]Token {
	key := lang + "\x00" + code
	if ls, ok := cache.get(key); ok {
		return ls
	}
	ls := lex(code, lang)
	cache.put(key, ls)
	return ls
}

func lex(code, lang string) [][]Token {
	code = strings.TrimSuffix(strings.ReplaceAll(code, "\r\n", "\n"), "\n")
	var tokens []chroma.Token
	if l := Lexer(lang); l != nil {
		if it, err := chroma.Coalesce(l).Tokenise(nil, code+"\n"); err == nil {
			tokens = it.Tokens()
		}
	}
	if tokens == nil {
		tokens = []chroma.Token{{Type: chroma.Text, Value: code}}
	}
	out := [][]Token{nil}
	col := 0
	for _, t := range tokens {
		kind := kindOf(t.Type)
		for i, part := range strings.Split(t.Value, "\n") {
			if i > 0 {
				out = append(out, nil)
				col = 0
			}
			if part == "" {
				continue
			}
			part, col = expand(part, col)
			last := &out[len(out)-1]
			if n := len(*last); n > 0 && (*last)[n-1].Kind == kind {
				(*last)[n-1].Text += part
				continue
			}
			*last = append(*last, Token{Text: part, Kind: kind})
		}
	}
	// The newline the lexer was given ends the last line.
	if len(out) > 1 && len(out[len(out)-1]) == 0 {
		out = out[:len(out)-1]
	}
	return out
}

// expand turns s's tabs into spaces to the next tab stop, s starting at
// column col; it returns s and the column after it. A rune is a column:
// near enough for indentation, which is what tabs are in code.
func expand(s string, col int) (string, int) {
	if !strings.ContainsRune(s, '\t') {
		return s, col + len([]rune(s))
	}
	var b strings.Builder
	for _, r := range s {
		if r == '\t' {
			n := TabWidth - col%TabWidth
			b.WriteString(strings.Repeat(" ", n))
			col += n
			continue
		}
		b.WriteRune(r)
		col++
	}
	return b.String(), col
}

// kindOf is a chroma token type's kind.
func kindOf(t chroma.TokenType) Kind {
	switch {
	case t == chroma.Error || t == chroma.GenericError:
		return Invalid
	case t.InSubCategory(chroma.CommentPreproc):
		return Meta
	case t.InCategory(chroma.Comment):
		return Comment
	case t == chroma.KeywordType:
		return Type
	case t == chroma.KeywordConstant:
		return Literal
	case t.InCategory(chroma.Keyword):
		return Keyword
	case t.InSubCategory(chroma.NameBuiltin), t == chroma.NameClass, t == chroma.NameTag, t == chroma.NameException:
		return Type
	case t.InSubCategory(chroma.NameFunction):
		return Function
	case t == chroma.NameConstant:
		return Literal
	case t == chroma.NameDecorator, t == chroma.NameAttribute:
		return Meta
	case t.InSubCategory(chroma.LiteralString):
		return String
	case t.InSubCategory(chroma.LiteralNumber), t == chroma.LiteralDate:
		return Literal
	case t == chroma.GenericInserted:
		return Inserted
	case t == chroma.GenericDeleted:
		return Deleted
	case t == chroma.GenericHeading, t == chroma.GenericSubheading:
		return Heading
	}
	return Plain
}

// cache keeps what Lines made, since a view is built again at every key
// and lexing is the slow part: up to cacheSize codes, then it starts over.
var cache = &lines{m: map[string][][]Token{}}

const cacheSize = 64

type lines struct {
	mu sync.Mutex
	m  map[string][][]Token
}

func (c *lines) get(k string) ([][]Token, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	ls, ok := c.m[k]
	return ls, ok
}

func (c *lines) put(k string, ls [][]Token) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.m) >= cacheSize {
		c.m = map[string][][]Token{}
	}
	c.m[k] = ls
}
