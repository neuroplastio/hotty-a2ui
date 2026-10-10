package view

import (
	"bytes"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	east "github.com/yuin/goldmark/extension/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"

	"github.com/neuroplastio/hotty-a2ui/a2ui"
	"github.com/neuroplastio/hotty-a2ui/catalog/hotty"
)

// A HottyMarkdown (profile §6.27) in the view: long-form Markdown, a docs
// page, read as GitHub reads it. Its headings get GitHub's IDs, numbered
// across the surface, the page; its alerts ("> [!NOTE]") are callouts; its
// HTML, comments and the docs module's markers among it, shows nothing.
// A link with no scheme goes in place: a #fragment scrolls to its heading,
// the renderer's own act, and any other is the agent's, written to where
// link is bound before onLink runs. A link with a scheme opens as a Text's
// does, the terminal's.

// The kinds a HottyMarkdown makes.
const (
	// Document is long-form Markdown (HottyMarkdown): Markdown, its text;
	// Doc, it read; Anchors, its headings' IDs; Value, the ID of the
	// heading a link in place last went to in it (Target). Its children are
	// its links in place, in order.
	Document Kind = "document"
	// Link is a HottyMarkdown's link in place, a Tab stop: Label, its text;
	// URL, its href; Selected, its index among its document's.
	Link Kind = "link"
)

// The blocks a HottyMarkdown has besides a Text's.
const (
	// TableBlock is a table: its Rows of cells, the header's first, and
	// its columns' Align.
	TableBlock BlockKind = "table"
	// AlertTitle is an alert's title row, of its Alert's kind: what the
	// alert's blocks follow, in its frame.
	AlertTitle BlockKind = "alert"
)

// The frames a document's block sits in.
const (
	// FrameItem is a list item: its Marker before its first block's first
	// line, as wide a margin before every other line.
	FrameItem = "item"
	// FrameQuote is a quote, a bar before each line.
	FrameQuote = "quote"
	// FrameAlert is an alert, a bar in its tone before each line.
	FrameAlert = "alert"
)

// An Alert is one of the five GitHub draws ("> [!NOTE]"): its title, the
// icon a host draws before it (a Material Symbols name, as GitHub's
// octicons), the mark cells and text draw instead (a character of width
// one), and its tone, the colour of its bar and its title: info, success,
// accent, warning or error.
type Alert struct {
	Title string `json:"title"`
	Icon  string `json:"icon"`
	Mark  string `json:"mark"`
	Tone  string `json:"tone"`
}

// Alerts are GitHub's alerts by their kind, the marker's word in lower
// case. IMPORTANT is GitHub's purple, which no role is: the accent, which
// a brand's theme makes its own colour.
var Alerts = map[string]Alert{
	"note":      {Title: "Note", Icon: "info", Mark: "ⓘ", Tone: "info"},
	"tip":       {Title: "Tip", Icon: "lightbulb", Mark: "✓", Tone: "success"},
	"important": {Title: "Important", Icon: "feedback", Mark: "★", Tone: "accent"},
	"warning":   {Title: "Warning", Icon: "warning", Mark: "!", Tone: "warning"},
	"caution":   {Title: "Caution", Icon: "report", Mark: "✗", Tone: "error"},
}

// Doc is a HottyMarkdown's Markdown read (ReadDoc), as the renditions
// draw it: its blocks in order, its headings, and its links in place.
type Doc struct {
	Blocks   []DocBlock   `json:"blocks,omitempty"`
	Headings []DocHeading `json:"headings,omitempty"`
	Links    []DocLink    `json:"links,omitempty"`
}

// DocHeading is one of a document's headings: its text, what its ID is
// made of (Slugger), and how many of the document's links in place come
// before it, where Tab goes on from it.
type DocHeading struct {
	Text  string `json:"text"`
	Links int    `json:"links,omitempty"`
}

// DocLink is one of a document's links in place: where it goes, and its
// text.
type DocLink struct {
	Href  string `json:"href"`
	Label string `json:"label,omitempty"`
}

// A DocBlock is one block of a document: a paragraph, a heading, an
// alert's title, a code block, a rule or a table, with its inline runs, in
// its frames.
type DocBlock struct {
	Kind BlockKind `json:"kind"`
	// Level is a heading's level, 1 to 6; Heading its index in
	// Doc.Headings.
	Level   int   `json:"level,omitempty"`
	Heading int   `json:"heading,omitempty"`
	Runs    []Run `json:"runs,omitempty"`
	// Lang is a fenced code block's language, its info string's first
	// word.
	Lang string `json:"lang,omitempty"`
	// Alert is an alert title's kind (Alerts).
	Alert string `json:"alert,omitempty"`
	// Rows are a table's rows, the header's first, each cell's runs; Align
	// is each column's alignment: "" (start), "center" or "end".
	Rows  [][][]Run `json:"rows,omitempty"`
	Align []string  `json:"align,omitempty"`
	// In are the frames it sits in, the outermost first.
	In []Frame `json:"in,omitempty"`
	// Gap: a blank row comes before it, which shows the first GapIn of
	// its frames (a quote's bar). None comes between the items of a tight
	// list, or between an alert's title and its first block.
	Gap   bool `json:"gap,omitempty"`
	GapIn int  `json:"gapIn,omitempty"`
}

// A Frame is what a document's block sits in: a list item, a quote or an
// alert (FrameItem, FrameQuote, FrameAlert).
type Frame struct {
	Kind string `json:"kind"`
	// Marker is an item's bullet, number or task box, with its space: "• ",
	// "3. ", "☐ ", "✓ ". First: the block is the item's first, whose first
	// line shows it.
	Marker string `json:"marker,omitempty"`
	First  bool   `json:"first,omitempty"`
	// Alert is an alert's kind.
	Alert string `json:"alert,omitempty"`
	item  int
}

// Jump is the heading a link in place last went to (FollowLink): its
// document, its ID, and how many jumps there have been, so that a
// rendition scrolls to it once a jump, and the user scrolls on from there.
type Jump struct {
	Doc     string
	Heading string
	Seq     int
}

func init() { Register(hotty.ID, "HottyMarkdown", mapDocument) }

// mapDocument makes a HottyMarkdown's element: its Markdown read, and a
// Link for each of its links in place. Its headings' IDs come once the
// surface is built (anchor).
func mapDocument(b *Builder, n *a2ui.Node) *Element {
	src := b.String(n, "text")
	d := ReadDoc(src)
	e := &Element{Kind: Document, Markdown: src, Doc: d}
	for i, l := range d.Links {
		e.Children = append(e.Children, &Element{ID: SubID(n.Key, "link", i), Kind: Link, Type: "HottyMarkdown.link",
			Label: l.Label, URL: l.Href, Selected: i})
	}
	if b.St.Jump.Doc == n.Key {
		e.Value = b.St.Jump.Heading
	}
	return e
}

// anchor gives the headings of the surface's documents their IDs, in tree
// order, as GitHub numbers a page's: a docs page comes as several
// HottyMarkdowns with components between them, and the surface is the
// page.
func anchor(v *Surface) {
	var s Slugger
	v.Walk(func(e *Element) bool {
		if e.Kind == Document && e.Doc != nil {
			e.Anchors = make([]string, len(e.Doc.Headings))
			for i, h := range e.Doc.Headings {
				e.Anchors[i] = s.Slug(h.Text)
			}
		}
		return true
	})
}

// Target is the index of the heading a link in place last went to in a
// Document, which its renditions scroll to while it has the keyboard; -1
// when there is none.
func (e *Element) Target() int {
	id, _ := e.Value.(string)
	if e.Kind != Document || id == "" {
		return -1
	}
	return slices.Index(e.Anchors, id)
}

// InPlace reports whether a link's href goes in place (profile §6.27): it
// names no scheme, so it is a path relative to the page or a #fragment. A
// scheme is a letter, then letters, digits, "+", "-" and ".", before the
// first ":" and before any "/", "?" or "#" (RFC 3986); "//host" is a URL
// too. An empty href is no link.
func InPlace(href string) bool {
	if href == "" || strings.HasPrefix(href, "//") {
		return false
	}
	i := strings.IndexAny(href, ":/?#")
	if i <= 0 || href[i] != ':' {
		return true
	}
	for k, c := range href[:i] {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || k > 0 && (c >= '0' && c <= '9' || c == '+' || c == '-' || c == '.')) {
			return true
		}
	}
	return false
}

// Heading finds a heading by its ID, a #fragment without its "#" (percent
// decoded here), among the surface's documents: the document, and the ID
// as its Anchors have it; nil when none has it. A fragment that matches
// no ID is looked for again in lower case, which GitHub's IDs are.
func (s *Surface) Heading(frag string) (*Element, string) {
	if u, err := url.PathUnescape(frag); err == nil {
		frag = u
	}
	if frag == "" {
		return nil, ""
	}
	for _, want := range []string{frag, strings.ToLower(frag)} {
		var found *Element
		s.Walk(func(e *Element) bool {
			if e.Kind == Document && slices.Contains(e.Anchors, want) {
				found = e
			}
			return found == nil
		})
		if found != nil {
			return found, want
		}
	}
	return nil, ""
}

// FollowLink follows a HottyMarkdown's link in place (profile §6.27). A
// #fragment that names one of the surface's headings is the renderer's to
// go to: the heading's document takes the keyboard and the jump (Jump),
// which a rendition scrolls to, and nothing is sent. Any other href, a
// fragment no heading has among them, is the agent's: it is written to
// where the document's link is bound, else to the renderer's state, and
// then onLink runs, whose context reads it (an action carries no payload).
func (c *Controller) FollowLink(id string) error {
	e := c.V.Find(id)
	if e == nil || e.Kind != Link {
		return nil
	}
	if frag, ok := strings.CutPrefix(e.URL, "#"); ok && c.GoTo(frag) {
		return nil
	}
	docID, _ := parentAndIndex(id)
	doc := c.V.Find(docID)
	if doc == nil {
		return nil
	}
	err := c.setProp(doc, "link", e.URL)
	c.Rebuild()
	if n := c.V.Node(docID); err == nil && n != nil && n.Props["onLink"] != nil {
		err = c.invoke(n, "onLink", true)
		c.Rebuild()
	}
	return err
}

// GoTo goes to the heading a #fragment names, as a link in place to it
// does (FollowLink): the heading's document takes the keyboard, a
// rendition brings the heading into sight once (State.Jump), and Tab goes
// on from it. frag is the fragment, its "#" optional, matched as a link's
// is (Surface.Heading). It is for a program's own contents ("On this
// page") and a #fragment in the address a page opens at; false when no
// heading on the surface has that anchor.
func (c *Controller) GoTo(frag string) bool {
	d, h := c.V.Heading(strings.TrimPrefix(frag, "#"))
	if d == nil {
		return false
	}
	c.St.Focus, c.St.Keyboard = d.ID, true
	c.St.Jump = Jump{Doc: d.ID, Heading: h, Seq: c.St.Jump.Seq + 1}
	c.Rebuild()
	return true
}

// HeadingID is the id of a HottyMarkdown's heading, its document's id,
// "~#" and the heading's anchor (Element.Anchors): where cells keeps its
// rows (Rendition.Box) and, encoded as every id is, the heading's element
// id on a host.
func HeadingID(doc, anchor string) string { return doc + "~#" + anchor }

// afterJump is where in ids, the surface's Focusables, the first element
// after the heading the keyboard is on comes, while it is on one a link in
// place went to: the document's next link in place, else the first
// focusable after the document. ok is false otherwise.
func (c *Controller) afterJump(ids []string) (at int, ok bool) {
	e := c.V.Find(c.St.Focus)
	if !c.St.Keyboard || e == nil || e.Kind != Document || e.Target() < 0 {
		return 0, false
	}
	// Its links come right after it in tree order, and each is a Tab stop.
	if k := e.Doc.Headings[e.Target()].Links; k < len(e.Children) {
		i := slices.Index(ids, e.Children[k].ID)
		return i, i >= 0
	}
	i := c.V.placeOf(e.ID)
	return i + len(e.Children), i >= 0
}

// placeOf is how many of the elements Focusables lists come before id in
// tree order; -1 when id is not among those it walks.
func (s *Surface) placeOf(id string) int {
	n, at := 0, -1
	count := func(e *Element) bool {
		if e.ID == id {
			at = n
			return false
		}
		if e.Focusable() {
			n++
		}
		return true
	}
	if s.Overlay != nil {
		(&Surface{Root: s.Overlay, Toasts: s.Toasts}).Walk(count)
	} else {
		s.Walk(count)
	}
	return at
}

// Slugger makes headings' IDs as GitHub does (github-slugger), one page's
// at a time: the heading's text in lower case, without what is not a
// letter, a mark, a number, a connector ("_"), a space or "-", its spaces
// "-"; an ID the page has already gets "-1", then "-2", after it.
type Slugger struct{ seen map[string]int }

// Slug is the ID of a heading whose text is text, the next on the page.
func (s *Slugger) Slug(text string) string {
	if s.seen == nil {
		s.seen = map[string]int{}
	}
	base := slug(text)
	id := base
	for {
		if _, taken := s.seen[id]; !taken {
			break
		}
		s.seen[base]++
		id = base + "-" + strconv.Itoa(s.seen[base])
	}
	s.seen[id] = 0
	return id
}

// slug is a heading's text as GitHub's ID of it, before it is numbered.
func slug(text string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(text) {
		switch {
		case r == ' ':
			b.WriteByte('-')
		case r == '-' || unicode.IsLetter(r) || unicode.IsMark(r) || unicode.IsNumber(r) || unicode.Is(unicode.Pc, r):
			b.WriteRune(r)
		}
	}
	return b.String()
}

// docMD reads a HottyMarkdown: GFM, and GitHub's alerts.
var docMD = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
	goldmark.WithParserOptions(parser.WithASTTransformers(util.Prioritized(alerts{}, 100))),
)

// kindAlert is the kind of an alertNode.
var kindAlert = ast.NewNodeKind("HottyAlert")

// An alertNode is a quote that is a GitHub alert, its marker taken out.
type alertNode struct {
	ast.BaseBlock
	kind string
}

func (n *alertNode) Kind() ast.NodeKind { return kindAlert }

func (n *alertNode) Dump(src []byte, level int) {
	ast.DumpHelper(n, src, level, map[string]string{"Alert": n.kind}, nil)
}

// alerts makes each quote of the document whose first line is an alert's
// marker ("[!NOTE]", case aside, alone on it) an alertNode, as GitHub
// does: a quote in a list or in another quote stays one.
type alerts struct{}

func (alerts) Transform(doc *ast.Document, r text.Reader, _ parser.Context) {
	src := r.Source()
	for c := doc.FirstChild(); c != nil; {
		next := c.NextSibling()
		if q, ok := c.(*ast.Blockquote); ok {
			if kind, ok := alertKind(q, src); ok {
				a := &alertNode{kind: kind}
				for k := q.FirstChild(); k != nil; {
					after := k.NextSibling()
					a.AppendChild(a, k)
					k = after
				}
				doc.ReplaceChild(doc, q, a)
			}
		}
		c = next
	}
}

// alertKind reads a quote's first line, an alert's marker or not, and
// takes the marker out of its first paragraph (all of it, when the marker
// is all it has).
func alertKind(q *ast.Blockquote, src []byte) (string, bool) {
	p, ok := q.FirstChild().(*ast.Paragraph)
	if !ok || p.Lines().Len() == 0 {
		return "", false
	}
	first := p.Lines().At(0)
	line := strings.TrimSpace(string(first.Value(src)))
	if len(line) < 4 || !strings.HasPrefix(line, "[!") || !strings.HasSuffix(line, "]") {
		return "", false
	}
	kind := strings.ToLower(line[2 : len(line)-1])
	if _, ok := Alerts[kind]; !ok {
		return "", false
	}
	if p.Lines().Len() == 1 {
		q.RemoveChild(q, p)
		return kind, true
	}
	// The marker's line is text alone.
	rest := p.Lines().At(1).Start
	for k := p.FirstChild(); k != nil; {
		t, ok := k.(*ast.Text)
		if !ok || t.Segment.Start >= rest {
			break
		}
		after := k.NextSibling()
		p.RemoveChild(p, k)
		k = after
	}
	return kind, true
}

// docIndex numbers a document's headings and its links in place, in tree
// order, the same for its blocks (ReadDoc) and its HTML (DocHTML). A link
// inside an image is the image's text, no link.
type docIndex struct {
	heads    map[*ast.Heading]int
	links    map[*ast.Link]int
	headings []DocHeading
	hrefs    []string
}

func indexDoc(doc ast.Node, src []byte) *docIndex {
	x := &docIndex{heads: map[*ast.Heading]int{}, links: map[*ast.Link]int{}}
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch n := n.(type) {
		case *ast.Image:
			return ast.WalkSkipChildren, nil
		case *ast.Heading:
			x.heads[n] = len(x.headings)
			x.headings = append(x.headings, DocHeading{Text: headingText(n, src), Links: len(x.hrefs)})
		case *ast.Link:
			if d := string(n.Destination); InPlace(d) {
				x.links[n] = len(x.hrefs) + 1
				x.hrefs = append(x.hrefs, d)
			}
		}
		return ast.WalkContinue, nil
	})
	return x
}

// headingText is a heading's text as a host has it (textContent): its
// code's and its links' too, an image's alt text not.
func headingText(n ast.Node, src []byte) string {
	var b strings.Builder
	_ = ast.Walk(n, func(c ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		switch t := c.(type) {
		case *ast.Image, *ast.RawHTML:
			return ast.WalkSkipChildren, nil
		case *ast.CodeSpan:
			b.WriteString(plain(t, src))
			return ast.WalkSkipChildren, nil
		case *ast.Text:
			b.WriteString(textOf(t, src))
			if t.SoftLineBreak() || t.HardLineBreak() {
				b.WriteByte('\n')
			}
		case *ast.String:
			b.Write(t.Value)
		case *ast.AutoLink:
			b.Write(t.Label(src))
		}
		return ast.WalkContinue, nil
	})
	return b.String()
}

// textOf is a text node's text as written out: its backslash escapes and
// its entities resolved.
func textOf(t *ast.Text, src []byte) string {
	v := t.Segment.Value(src)
	return string(util.ResolveEntityNames(util.ResolveNumericReferences(util.UnescapePunctuations(v))))
}

// docs are the documents read lately, by their Markdown: a document is
// built again with every key, and read once.
var docs struct {
	sync.Mutex
	m map[string]*Doc
}

// ReadDoc reads a HottyMarkdown's Markdown into blocks, its headings and
// its links in place. The Doc is shared: not to be changed.
func ReadDoc(src string) *Doc {
	docs.Lock()
	d := docs.m[src]
	docs.Unlock()
	if d != nil {
		return d
	}
	d = readDoc(src)
	docs.Lock()
	if docs.m == nil || len(docs.m) >= 64 {
		docs.m = map[string]*Doc{}
	}
	docs.m[src] = d
	docs.Unlock()
	return d
}

func readDoc(src string) *Doc {
	s := []byte(src)
	root := docMD.Parser().Parse(text.NewReader(s))
	x := indexDoc(root, s)
	w := &docWalker{src: s, x: x, doc: &Doc{Headings: x.headings}, shown: map[int]bool{}}
	for _, h := range x.hrefs {
		w.doc.Links = append(w.doc.Links, DocLink{Href: h})
	}
	w.blocks(root, nil, false)
	return w.doc
}

type docWalker struct {
	src []byte
	x   *docIndex
	doc *Doc
	// gap: a blank row comes before the next block, showing gapIn frames.
	gap   bool
	gapIn int
	// items numbers the list items; shown are those whose marker is out.
	items int
	shown map[int]bool
}

// emit adds a block in its frames: an item's marker shows on its first.
func (w *docWalker) emit(b DocBlock, in []Frame) {
	b.In = slices.Clone(in)
	for i := range b.In {
		if f := &b.In[i]; f.Kind == FrameItem {
			f.First = !w.shown[f.item]
			w.shown[f.item] = true
		}
	}
	if w.gap && len(w.doc.Blocks) > 0 {
		b.Gap, b.GapIn = true, w.gapIn
	}
	w.gap = false
	w.doc.Blocks = append(w.doc.Blocks, b)
}

// marking reports whether a block in these frames would show an item's
// marker: an empty paragraph there still has to.
func (w *docWalker) marking(in []Frame) bool {
	return slices.ContainsFunc(in, func(f Frame) bool { return f.Kind == FrameItem && !w.shown[f.item] })
}

// blocks walks a container's blocks in its frames, a blank row between
// them unless they are a tight list's. HTML shows nothing, and takes no
// row.
func (w *docWalker) blocks(n ast.Node, in []Frame, tight bool) {
	first := true
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		if _, ok := c.(*ast.HTMLBlock); ok {
			continue
		}
		if !first && !tight {
			w.gap, w.gapIn = true, len(in)
		}
		first = false
		w.block(c, in, tight)
	}
}

func (w *docWalker) block(n ast.Node, in []Frame, tight bool) {
	switch n := n.(type) {
	case *ast.Heading:
		w.emit(DocBlock{Kind: Heading, Level: n.Level, Heading: w.x.heads[n], Runs: w.inline(n, 0, "", 0)}, in)
	case *ast.Paragraph, *ast.TextBlock:
		runs := w.inline(n, 0, "", 0)
		if len(runs) == 0 && !w.marking(in) {
			// Only HTML: nothing shows.
			return
		}
		w.emit(DocBlock{Kind: Paragraph, Runs: runs}, in)
	case *ast.List:
		num := n.Start
		for li := n.FirstChild(); li != nil; li = li.NextSibling() {
			if li != n.FirstChild() && !n.IsTight {
				w.gap, w.gapIn = true, len(in)
			}
			marker := "• "
			if n.IsOrdered() {
				marker = strconv.Itoa(num) + ". "
				num++
			}
			if p := li.FirstChild(); p != nil {
				if box, ok := p.FirstChild().(*east.TaskCheckBox); ok {
					marker = "☐ "
					if box.IsChecked {
						marker = "✓ "
					}
				}
			}
			w.items++
			kids := append(slices.Clone(in), Frame{Kind: FrameItem, Marker: marker, item: w.items})
			if li.FirstChild() == nil {
				w.emit(DocBlock{Kind: Paragraph}, kids)
				continue
			}
			w.blocks(li, kids, n.IsTight)
		}
	case *ast.Blockquote:
		w.blocks(n, append(slices.Clone(in), Frame{Kind: FrameQuote}), false)
	case *alertNode:
		kids := append(slices.Clone(in), Frame{Kind: FrameAlert, Alert: n.kind})
		w.emit(DocBlock{Kind: AlertTitle, Alert: n.kind}, kids)
		w.blocks(n, kids, false)
	case *ast.FencedCodeBlock, *ast.CodeBlock:
		var b strings.Builder
		lines := n.Lines()
		for i := 0; i < lines.Len(); i++ {
			seg := lines.At(i)
			b.Write(seg.Value(w.src))
		}
		blk := DocBlock{Kind: CodeBlock, Runs: []Run{{Text: strings.TrimRight(b.String(), "\n"), Style: Code}}}
		if f, ok := n.(*ast.FencedCodeBlock); ok {
			blk.Lang = string(f.Language(w.src))
		}
		w.emit(blk, in)
	case *ast.ThematicBreak:
		w.emit(DocBlock{Kind: Rule}, in)
	case *east.Table:
		blk := DocBlock{Kind: TableBlock}
		for _, a := range n.Alignments {
			blk.Align = append(blk.Align, map[east.Alignment]string{east.AlignCenter: "center", east.AlignRight: "end"}[a])
		}
		for r := n.FirstChild(); r != nil; r = r.NextSibling() {
			var row [][]Run
			for c := r.FirstChild(); c != nil; c = c.NextSibling() {
				row = append(row, w.inline(c, 0, "", 0))
			}
			blk.Rows = append(blk.Rows, row)
		}
		w.emit(blk, in)
	case *ast.HTMLBlock:
	default:
		w.blocks(n, in, tight)
	}
}

// inline is a node's inline children as runs, in style, inside a link to
// href (link: a link in place's number, from 1). An image is its alt
// text, as a Text's; HTML is nothing.
func (w *docWalker) inline(n ast.Node, style Style, href string, link int) []Run {
	var out []Run
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		switch c := c.(type) {
		case *ast.Text:
			t := textOf(c, w.src)
			switch {
			case c.HardLineBreak():
				t += "\n"
			case c.SoftLineBreak():
				t += " "
			}
			out = appendRun(out, Run{Text: t, Style: style, Href: href, Link: link})
		case *ast.String:
			out = appendRun(out, Run{Text: string(c.Value), Style: style, Href: href, Link: link})
		case *ast.CodeSpan:
			out = appendRun(out, Run{Text: plain(c, w.src), Style: style | Code, Href: href, Link: link})
		case *ast.Emphasis:
			f := Italic
			if c.Level >= 2 {
				f = Bold
			}
			out = appendRuns(out, w.inline(c, style|f, href, link))
		case *east.Strikethrough:
			out = appendRuns(out, w.inline(c, style|Strike, href, link))
		case *ast.Link:
			dest := string(c.Destination)
			k := w.x.links[c]
			if dest == "" {
				out = appendRuns(out, w.inline(c, style, href, link))
				continue
			}
			runs := w.inline(c, style, dest, k)
			if k > 0 {
				var label strings.Builder
				for _, r := range runs {
					label.WriteString(r.Text)
				}
				w.doc.Links[k-1].Label = strings.TrimSpace(label.String())
			}
			out = appendRuns(out, runs)
		case *ast.AutoLink:
			u := string(c.URL(w.src))
			if c.AutoLinkType == ast.AutoLinkEmail && !strings.HasPrefix(strings.ToLower(u), "mailto:") {
				u = "mailto:" + u
			}
			out = appendRun(out, Run{Text: string(c.Label(w.src)), Style: style, Href: u})
		case *ast.Image:
			alt := plain(c, w.src)
			if alt == "" {
				alt = string(c.Destination)
			}
			out = appendRun(out, Run{Text: "[image: " + alt + "]", Style: style | Italic, Href: href, Link: link})
		case *east.TaskCheckBox, *ast.RawHTML:
		default:
			out = appendRuns(out, w.inline(c, style, href, link))
		}
	}
	return out
}

// DocMarkup is how a host's rendition names and draws a document's parts
// in its HTML (DocHTML).
type DocMarkup struct {
	// Heading is the attributes of heading i, its ID Element.Anchors[i]:
	// name, value, name, value.
	Heading func(i int) []string
	// Link is the tags around link in place i (Element.Children[i]).
	Link func(i int) (start, end string)
	// Alert is the tags around an alert of its kind (Alerts), its title
	// row in start.
	Alert func(kind string) (start, end string)
}

// DocHTML is a HottyMarkdown's Markdown as HTML, as a Text's is, but its
// headings, its links in place and its alerts as m makes them; a link
// with a scheme a hyperlink, which the terminal opens (target=_blank, SPEC
// §9); and HTML, comments and the docs module's markers among it, left
// out.
func DocHTML(src string, m DocMarkup) string {
	s := []byte(src)
	root := docMD.Parser().Parse(text.NewReader(s))
	x := indexDoc(root, s)
	if m.Heading != nil {
		for h, i := range x.heads {
			attrs := m.Heading(i)
			for k := 0; k+1 < len(attrs); k += 2 {
				h.SetAttributeString(attrs[k], []byte(attrs[k+1]))
			}
		}
	}
	r := renderer.NewRenderer(renderer.WithNodeRenderers(
		util.Prioritized(html.NewRenderer(), 1000),
		util.Prioritized(extension.NewTableHTMLRenderer(), 500),
		util.Prioritized(extension.NewStrikethroughHTMLRenderer(), 500),
		util.Prioritized(extension.NewTaskCheckBoxHTMLRenderer(), 500),
		util.Prioritized(numbers{}, 100),
		util.Prioritized(fences{}, 100),
		util.Prioritized(&docHTML{m: m, x: x}, 100),
	))
	var b bytes.Buffer
	if err := r.Render(&b, s, root); err != nil {
		return "<p>" + escapeText(src) + "</p>"
	}
	return strings.TrimRight(b.String(), "\n")
}

// docHTML writes what DocHTML draws its own way.
type docHTML struct {
	m DocMarkup
	x *docIndex
}

func (d *docHTML) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(ast.KindLink, d.link)
	reg.Register(ast.KindAutoLink, d.autoLink)
	reg.Register(kindAlert, d.alert)
	none := func(util.BufWriter, []byte, ast.Node, bool) (ast.WalkStatus, error) { return ast.WalkSkipChildren, nil }
	reg.Register(ast.KindHTMLBlock, none)
	reg.Register(ast.KindRawHTML, none)
}

func (d *docHTML) link(w util.BufWriter, _ []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.Link)
	if len(n.Destination) == 0 {
		// No link: its text alone.
		return ast.WalkContinue, nil
	}
	if k, ok := d.x.links[n]; ok && d.m.Link != nil {
		start, end := d.m.Link(k - 1)
		if entering {
			_, _ = w.WriteString(start)
		} else {
			_, _ = w.WriteString(end)
		}
		return ast.WalkContinue, nil
	}
	if !entering {
		_, _ = w.WriteString("</a>")
		return ast.WalkContinue, nil
	}
	_, _ = w.WriteString(`<a href="`)
	if !html.IsDangerousURL(n.Destination) {
		_, _ = w.Write(util.EscapeHTML(util.URLEscape(n.Destination, true)))
	}
	_ = w.WriteByte('"')
	if n.Title != nil {
		_, _ = w.WriteString(` title="`)
		_, _ = w.Write(util.EscapeHTML(n.Title))
		_ = w.WriteByte('"')
	}
	_, _ = w.WriteString(` target="_blank">`)
	return ast.WalkContinue, nil
}

func (d *docHTML) autoLink(w util.BufWriter, src []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*ast.AutoLink)
	if !entering {
		return ast.WalkContinue, nil
	}
	u := n.URL(src)
	if n.AutoLinkType == ast.AutoLinkEmail && !bytes.HasPrefix(bytes.ToLower(u), []byte("mailto:")) {
		u = append([]byte("mailto:"), u...)
	}
	_, _ = w.WriteString(`<a href="`)
	if !html.IsDangerousURL(u) {
		_, _ = w.Write(util.EscapeHTML(util.URLEscape(u, false)))
	}
	_, _ = w.WriteString(`" target="_blank">`)
	_, _ = w.Write(util.EscapeHTML(n.Label(src)))
	_, _ = w.WriteString("</a>")
	return ast.WalkContinue, nil
}

func (d *docHTML) alert(w util.BufWriter, _ []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	n := node.(*alertNode)
	if d.m.Alert == nil {
		return ast.WalkContinue, nil
	}
	start, end := d.m.Alert(n.kind)
	if entering {
		_, _ = w.WriteString(start)
	} else {
		_, _ = w.WriteString(end)
	}
	return ast.WalkContinue, nil
}
