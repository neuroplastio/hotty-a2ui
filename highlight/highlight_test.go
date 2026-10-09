package highlight_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/highlight"
)

// text is a line's tokens as "kind:text" pairs, Plain spaces left out.
func text(line []highlight.Token) []string {
	var out []string
	for _, t := range line {
		if t.Kind == highlight.Plain && strings.TrimSpace(t.Text) == "" {
			continue
		}
		out = append(out, t.Kind.String()+":"+strings.TrimSpace(t.Text))
	}
	return out
}

// Go is lexed into kinds, a line at a time; tabs are spaces to the next
// stop of four.
func TestGo(t *testing.T) {
	ls := highlight.Lines("package main\n\n// Add adds.\nfunc Add(a, b int) int {\n\treturn a + b + 0x10 // sum\n}\n", "go")
	if len(ls) != 6 {
		t.Fatalf("%d lines: %q", len(ls), ls)
	}
	for i, want := range [][]string{
		{"keyword:package", "plain:main"},
		nil,
		{"comment:// Add adds."},
		{"keyword:func", "function:Add", "plain:(a, b", "type:int", "plain:)", "type:int", "plain:{"},
		{"keyword:return", "plain:a + b +", "literal:0x10", "comment:// sum"},
		{"plain:}"},
	} {
		if got := text(ls[i]); !slices.Equal(got, want) {
			t.Errorf("line %d: %q, want %q", i+1, got, want)
		}
	}
	if got := ls[4][0].Text; got != "    " {
		t.Errorf("the tab is %q", got)
	}
}

// A language goes by its name, an alias, an extension or a file name,
// whatever its case; one no lexer knows is plain, a token a line.
func TestLanguages(t *testing.T) {
	for _, lang := range []string{"go", "Go", "golang", "main.go", "python", "py", "ts", "TypeScript", "rust", "rs", "sh", "bash", "json", "yaml", "yml",
		"diff", "Dockerfile", "Makefile", "c++", "cpp", "c#", "csharp", "html", "css", "sql", "toml", "lua", "zig", "terraform", "tf"} {
		if highlight.Lexer(lang) == nil {
			t.Errorf("no lexer for %q", lang)
		}
	}
	if highlight.Lexer("cobol") != nil || highlight.Lexer("") != nil {
		t.Error("a lexer for cobol, or for nothing")
	}
	if got := highlight.Lines("a\tb\nc", "cobol"); len(got) != 2 || !slices.Equal(text(got[0]), []string{"plain:a   b"}) {
		t.Errorf("plain: %q", got)
	}
	if n := len(highlight.Languages()); n != 37 {
		t.Errorf("%d languages: %q", n, highlight.Languages())
	}
}

// Each of the kit's lexers reads (scripts/lexers.sh copies chroma's), and
// lexes something without an error token: a parse that failed would be
// plain, a lexer that failed in its rules a panic.
func TestLexers(t *testing.T) {
	for _, name := range highlight.Languages() {
		ls := highlight.Lines("x = 1 # 'a' \"b\" // c\n", name)
		if len(ls) != 1 {
			t.Errorf("%s: %d lines", name, len(ls))
		}
	}
}

// HTML's style and script are CSS and JavaScript, lexed by the kit's own.
func TestEmbedded(t *testing.T) {
	ls := highlight.Lines("<style>p { color: red }</style><script>let x = 1</script>", "html")
	kinds := map[highlight.Kind]bool{}
	for _, tok := range ls[0] {
		kinds[tok.Kind] = true
	}
	if !kinds[highlight.Type] || !kinds[highlight.Keyword] || !kinds[highlight.Literal] {
		t.Errorf("html with css and js: %q", text(ls[0]))
	}
}

// A diff's lines are inserted, deleted and headings; its file headers,
// --- and +++, are deleted and inserted, as chroma (and Pygments) has it.
func TestDiff(t *testing.T) {
	ls := highlight.Lines("--- a/x\n+++ b/x\n@@ -1 +1 @@\n-old\n+new\n same\n", "diff")
	var got []string
	for _, l := range ls {
		got = append(got, text(l)[0])
	}
	want := []string{"deleted:--- a/x", "inserted:+++ b/x", "heading:@@ -1 +1 @@", "deleted:-old", "inserted:+new", "plain:same"}
	if !slices.Equal(got, want) {
		t.Errorf("got %q", got)
	}
}

// Newlines: a final one ends the last line, a blank line stays, CRLF is
// a newline, and empty code is one empty line.
func TestNewlines(t *testing.T) {
	for code, n := range map[string]int{"a": 1, "a\n": 1, "a\n\n": 2, "a\r\nb\r\n": 2, "": 1, "\n\nb": 3} {
		// With a lexer and without one (plain text).
		for _, lang := range []string{"go", ""} {
			if got := len(highlight.Lines(code, lang)); got != n {
				t.Errorf("%q as %q: %d lines, want %d", code, lang, got, n)
			}
		}
	}
}
