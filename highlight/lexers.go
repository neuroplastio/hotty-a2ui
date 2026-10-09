package highlight

import (
	"embed"
	"path"
	"strings"
	"sync"

	"github.com/alecthomas/chroma/v2"
)

// The kit's own lexers: chroma's definitions, copied by scripts/lexers.sh.
//
//go:embed lexers/*.xml
var xmls embed.FS

var (
	registry     *chroma.LexerRegistry
	registryOnce sync.Once
	lookupsMu    sync.Mutex
	lookups      []func(lang string) chroma.Lexer
)

// kit is the registry of the kit's own lexers, read the first time code
// is lexed: each is a regular expression compiled on its first use.
func kit() *chroma.LexerRegistry {
	registryOnce.Do(func() {
		registry = chroma.NewLexerRegistry()
		entries, _ := xmls.ReadDir("lexers")
		for _, e := range entries {
			l, err := chroma.NewXMLLexer(xmls, path.Join("lexers", e.Name()))
			if err != nil {
				// scripts/lexers.sh copies only what chroma itself reads;
				// TestLexers says which one did not.
				continue
			}
			registry.Register(l)
		}
	})
	return registry
}

// Lexer is the lexer for lang, a language's name or alias, or a file name
// (Lines); nil when none knows it. The kit's own come first, then those
// AddLookup added, in order.
func Lexer(lang string) chroma.Lexer {
	lang = strings.TrimSpace(lang)
	if lang == "" {
		return nil
	}
	// By name, alias, extension ("ts"), then file name.
	if l := kit().Get(lang); l != nil {
		return l
	}
	lookupsMu.Lock()
	more := lookups
	lookupsMu.Unlock()
	for _, f := range more {
		if l := f(lang); l != nil {
			return l
		}
	}
	return nil
}

// AddLookup adds a way to find a lexer for a language the kit's own do
// not know, as highlight/all does with all of chroma's.
func AddLookup(f func(lang string) chroma.Lexer) {
	lookupsMu.Lock()
	defer lookupsMu.Unlock()
	lookups = append(lookups, f)
}

// Languages are the names of the kit's own lexers, sorted.
func Languages() []string { return kit().Names(false) }
