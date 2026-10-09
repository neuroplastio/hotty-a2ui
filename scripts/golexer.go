//go:build ignore

// golexer writes chroma's Go lexer as XML (scripts/lexers.sh): chroma's is
// Go code in its lexers package, which would bring every lexer in with it.
// A raw string is a string here: chroma lexes Go's text templates inside
// one, with a lexer XML cannot name.
//
//	go run scripts/golexer.go > highlight/lexers/go.xml
package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
)

func main() {
	src, ok := lexers.Go.(*chroma.RegexLexer)
	if !ok {
		fail("lexers.Go is a %T", lexers.Go)
	}
	rules, err := src.Rules()
	if err != nil {
		fail("%v", err)
	}
	raw := 0
	for state, rs := range rules {
		for i, r := range rs {
			if r.Pattern == "(`)([^`]*)(`)" {
				rules[state][i] = chroma.Rule{Pattern: "`[^`]*`", Type: chroma.LiteralString}
				raw++
			}
		}
	}
	if raw != 1 {
		fail("found %d raw string rules, not 1: chroma's go.go changed", raw)
	}
	out, err := chroma.Marshal(chroma.MustNewLexer(src.Config(), func() chroma.Rules { return rules }))
	if err != nil {
		fail("%v", err)
	}
	fmt.Printf("%s\n", out)
}

func fail(f string, a ...any) {
	fmt.Fprintf(os.Stderr, "golexer: "+f+"\n", a...)
	os.Exit(1)
}
