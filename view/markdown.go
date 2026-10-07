package view

import (
	"bytes"
	"strconv"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	east "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// A Text's content is Markdown, which goldmark reads (GitHub's flavour).
// A host is given goldmark's HTML; cells and plain text walk its tree
// into Blocks of styled Runs, which they wrap themselves.

// md is the Markdown every rendition reads: GFM, raw HTML left out (a
// Text has none), and links opened by the terminal (target=_blank makes a
// link a hyperlink, SPEC §9).
var md = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
	goldmark.WithParserOptions(parser.WithASTTransformers(util.Prioritized(hyperlinks{}, 100))),
)

type hyperlinks struct{}

func (hyperlinks) Transform(doc *ast.Document, _ text.Reader, _ parser.Context) {
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			if _, ok := n.(*ast.Link); ok {
				n.SetAttributeString("target", []byte("_blank"))
			}
		}
		return ast.WalkContinue, nil
	})
}

// MarkdownHTML is a Text's Markdown as HTML.
func MarkdownHTML(src string) string {
	var b bytes.Buffer
	if err := md.Convert([]byte(src), &b); err != nil {
		return "<p>" + escapeText(src) + "</p>"
	}
	return strings.TrimRight(b.String(), "\n")
}

func escapeText(s string) string {
	return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;").Replace(s)
}

// A Block is one block of a Text: a paragraph, a heading, a list item, a
// quote, a code block or a rule, with its inline runs.
type Block struct {
	Kind BlockKind `json:"kind"`
	// Level is a heading's level (1 to 6), or a list item's or a quote's
	// depth (0 for the outermost).
	Level int `json:"level,omitempty"`
	// Number is an ordered list item's number; 0 for a bullet.
	Number int `json:"number,omitempty"`
	// Task is a task list item's box: 0 none, 1 open, 2 ticked.
	Task int   `json:"task,omitempty"`
	Runs []Run `json:"runs,omitempty"`
}

// BlockKind is what a Block is.
type BlockKind string

// The kinds of block.
const (
	Paragraph BlockKind = "p"
	Heading   BlockKind = "h"
	ListItem  BlockKind = "li"
	Quote     BlockKind = "quote"
	CodeBlock BlockKind = "code"
	Rule      BlockKind = "hr"
)

// A Run is text in one style. A link's Href is set. Text may hold "\n",
// a hard line break.
type Run struct {
	Text  string `json:"text"`
	Style Style  `json:"style,omitempty"`
	Href  string `json:"href,omitempty"`
}

// Style is a run's inline style, a set of flags.
type Style uint8

// The inline styles.
const (
	Bold Style = 1 << iota
	Italic
	Code
	Strike
)

// Markdown reads a Text's Markdown into blocks.
func Markdown(src string) []Block {
	s := []byte(src)
	doc := md.Parser().Parse(text.NewReader(s))
	w := &walker{src: s}
	for n := doc.FirstChild(); n != nil; n = n.NextSibling() {
		w.block(n, 0)
	}
	return w.out
}

type walker struct {
	src []byte
	out []Block
}

func (w *walker) block(n ast.Node, depth int) {
	switch n := n.(type) {
	case *ast.Heading:
		w.out = append(w.out, Block{Kind: Heading, Level: n.Level, Runs: w.inline(n, 0, "")})
	case *ast.Paragraph, *ast.TextBlock:
		w.out = append(w.out, Block{Kind: Paragraph, Runs: w.inline(n, 0, "")})
	case *ast.List:
		num := n.Start
		for li := n.FirstChild(); li != nil; li = li.NextSibling() {
			item := Block{Kind: ListItem, Level: depth}
			if n.IsOrdered() {
				item.Number = max(num, 1)
				num++
			}
			first := true
			for c := li.FirstChild(); c != nil; c = c.NextSibling() {
				switch c.(type) {
				case *ast.Paragraph, *ast.TextBlock:
					if first {
						if box, ok := c.FirstChild().(*east.TaskCheckBox); ok {
							item.Task = 1
							if box.IsChecked {
								item.Task = 2
							}
						}
						item.Runs = w.inline(c, 0, "")
						w.out = append(w.out, item)
						first = false
						continue
					}
				}
				if first {
					w.out = append(w.out, item)
					first = false
				}
				w.block(c, depth+1)
			}
			if first {
				w.out = append(w.out, item)
			}
		}
	case *ast.Blockquote:
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			start := len(w.out)
			w.block(c, depth)
			for i := start; i < len(w.out); i++ {
				if w.out[i].Kind == Paragraph {
					w.out[i].Kind = Quote
					w.out[i].Level = depth
				}
			}
		}
	case *ast.FencedCodeBlock, *ast.CodeBlock:
		var b strings.Builder
		lines := n.Lines()
		for i := 0; i < lines.Len(); i++ {
			seg := lines.At(i)
			b.Write(seg.Value(w.src))
		}
		w.out = append(w.out, Block{Kind: CodeBlock, Runs: []Run{{Text: strings.TrimRight(b.String(), "\n"), Style: Code}}})
	case *ast.ThematicBreak:
		w.out = append(w.out, Block{Kind: Rule})
	case *east.Table:
		for r := n.FirstChild(); r != nil; r = r.NextSibling() {
			var runs []Run
			for c := r.FirstChild(); c != nil; c = c.NextSibling() {
				if c != r.FirstChild() {
					runs = appendRun(runs, Run{Text: "  │  "})
				}
				style := Style(0)
				if _, head := r.(*east.TableHeader); head {
					style = Bold
				}
				runs = append(runs, w.inline(c, style, "")...)
			}
			w.out = append(w.out, Block{Kind: Paragraph, Runs: runs})
		}
	case *ast.HTMLBlock:
		// A Text has no HTML.
	default:
		for c := n.FirstChild(); c != nil; c = c.NextSibling() {
			w.block(c, depth)
		}
	}
}

// inline is a node's inline children as runs, in style, inside a link to
// href.
func (w *walker) inline(n ast.Node, style Style, href string) []Run {
	var out []Run
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		switch c := c.(type) {
		case *ast.Text:
			t := string(c.Segment.Value(w.src))
			switch {
			case c.HardLineBreak():
				t += "\n"
			case c.SoftLineBreak():
				t += " "
			}
			out = appendRun(out, Run{Text: t, Style: style, Href: href})
		case *ast.String:
			out = appendRun(out, Run{Text: string(c.Value), Style: style, Href: href})
		case *ast.CodeSpan:
			out = appendRun(out, Run{Text: plain(c, w.src), Style: style | Code, Href: href})
		case *ast.Emphasis:
			f := Italic
			if c.Level >= 2 {
				f = Bold
			}
			out = appendRuns(out, w.inline(c, style|f, href))
		case *east.Strikethrough:
			out = appendRuns(out, w.inline(c, style|Strike, href))
		case *ast.Link:
			out = appendRuns(out, w.inline(c, style, string(c.Destination)))
		case *ast.AutoLink:
			u := string(c.URL(w.src))
			out = appendRun(out, Run{Text: string(c.Label(w.src)), Style: style, Href: u})
		case *ast.Image:
			alt := plain(c, w.src)
			if alt == "" {
				alt = string(c.Destination)
			}
			out = appendRun(out, Run{Text: "[image: " + alt + "]", Style: style | Italic, Href: href})
		case *east.TaskCheckBox, *ast.RawHTML:
		default:
			out = appendRuns(out, w.inline(c, style, href))
		}
	}
	return out
}

// plain is a node's text, its styles dropped.
func plain(n ast.Node, src []byte) string {
	var b strings.Builder
	_ = ast.Walk(n, func(c ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch t := c.(type) {
		case *ast.Text:
			b.Write(t.Segment.Value(src))
		case *ast.String:
			b.Write(t.Value)
		}
		return ast.WalkContinue, nil
	})
	return b.String()
}

func appendRuns(runs []Run, more []Run) []Run {
	for _, r := range more {
		runs = appendRun(runs, r)
	}
	return runs
}

// appendRun adds a run, joined to the last when they look the same.
func appendRun(runs []Run, r Run) []Run {
	if r.Text == "" {
		return runs
	}
	if n := len(runs); n > 0 && runs[n-1].Style == r.Style && runs[n-1].Href == r.Href {
		runs[n-1].Text += r.Text
		return runs
	}
	return append(runs, r)
}

// PlainText is blocks as plain text: a line a block, list items marked
// "• ", "1. " or a task's box, indented two spaces a level, quotes "▎ ".
func PlainText(blocks []Block) string {
	var b strings.Builder
	for i, bl := range blocks {
		if i > 0 {
			b.WriteByte('\n')
		}
		switch bl.Kind {
		case ListItem:
			b.WriteString(strings.Repeat("  ", bl.Level))
			switch {
			case bl.Task == 1:
				b.WriteString("☐ ")
			case bl.Task == 2:
				b.WriteString("✓ ")
			case bl.Number > 0:
				b.WriteString(strconv.Itoa(bl.Number) + ". ")
			default:
				b.WriteString("• ")
			}
		case Quote:
			b.WriteString(strings.Repeat("▎ ", bl.Level+1))
		case Rule:
			b.WriteString("───")
		}
		for _, r := range bl.Runs {
			b.WriteString(r.Text)
		}
	}
	return b.String()
}
