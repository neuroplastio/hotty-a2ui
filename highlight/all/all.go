// Package all adds all of chroma's lexers to highlight's, for a program
// that wants every language chroma knows, at the size they take: import
// it for its effect.
//
//	import _ "github.com/neuroplastio/hotty-a2ui/highlight/all"
package all

import (
	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"

	"github.com/neuroplastio/hotty-a2ui/highlight"
)

func init() {
	highlight.AddLookup(func(lang string) chroma.Lexer { return lexers.Get(lang) })
}
