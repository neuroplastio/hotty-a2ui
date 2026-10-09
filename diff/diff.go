// Package diff is a change to text as the kit shows it (HottyDiff, profile
// §6.13): files of hunks of lines, each line lexed as its file's language
// (package highlight) and, where a line replaces another, the words it
// changes marked. It reads a unified diff, as git and diff -u write it
// (Parse), or compares two texts (Compare).
package diff

import (
	"strconv"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"github.com/aymanbagabas/go-udiff/lcs"

	"github.com/neuroplastio/hotty-a2ui/highlight"
)

// Op is what a line of a hunk is.
type Op uint8

// The ops.
const (
	Context Op = iota // a line both sides have
	Removed           // a line only the old side has
	Added             // a line only the new side has
)

// Line is one line of a hunk.
type Line struct {
	Op Op
	// Old and New are its numbers on each side, from 1; 0 on the side that
	// does not have it.
	Old, New int
	// Tokens are its text, lexed with the rest of its side of the hunk;
	// tabs are spaces (highlight.TabWidth).
	Tokens []highlight.Token
	// Changed are the byte ranges of its text (Text) that a removed line
	// and the added line that replaces it do not share: the words the
	// change changes, in order. Nil on a context line, on a line that
	// replaces none, and on one too unlike its pair for its words to say
	// more than that the whole line changed.
	Changed [][2]int
}

// Text is the line's text, its tabs spaces.
func (l Line) Text() string {
	var b strings.Builder
	for _, t := range l.Tokens {
		b.WriteString(t.Text)
	}
	return b.String()
}

// Segment is a run of a line's text of one kind, changed or not.
type Segment struct {
	Text    string
	Kind    highlight.Kind
	Changed bool
}

// Segments are the line's tokens, split where its changed ranges start
// and end.
func (l Line) Segments() []Segment {
	var out []Segment
	at, r := 0, 0
	for _, t := range l.Tokens {
		s := t.Text
		for s != "" {
			for r < len(l.Changed) && l.Changed[r][1] <= at {
				r++
			}
			n, changed := len(s), false
			if r < len(l.Changed) {
				from, to := l.Changed[r][0], l.Changed[r][1]
				if at < from {
					n = min(n, from-at)
				} else {
					n, changed = min(n, to-at), true
				}
			}
			out = append(out, Segment{Text: s[:n], Kind: t.Kind, Changed: changed})
			s, at = s[n:], at+n
		}
	}
	return out
}

// Hunk is a run of changes with the unchanged lines around them.
type Hunk struct {
	// OldStart and NewStart are its first line's numbers on each side,
	// and OldLines and NewLines how many lines it has there, as its header
	// says (@@ -OldStart,OldLines +NewStart,NewLines @@). A side with no
	// lines starts at the line before it, as there.
	OldStart, OldLines, NewStart, NewLines int
	// Section is what follows the header: the function or the heading the
	// hunk is in, where the diff found one.
	Section string
	// Skipped is how many unchanged lines come between the previous hunk,
	// or the file's start, and this one: what the diff leaves out.
	Skipped int
	Lines   []Line
}

// Header is the hunk's header, @@ -1,7 +1,8 @@, without its Section.
func (h Hunk) Header() string {
	return "@@ -" + span(h.OldStart, h.OldLines) + " +" + span(h.NewStart, h.NewLines) + " @@"
}

func span(start, n int) string {
	if n == 1 {
		return strconv.Itoa(start)
	}
	return strconv.Itoa(start) + "," + strconv.Itoa(n)
}

// File is the change to one file.
type File struct {
	// Old and New are its names before and after, without git's a/ and b/;
	// "" where it did not exist (a file added, or deleted), and both "" for
	// texts compared with no name.
	Old, New string
	Hunks    []Hunk
	// After is how many unchanged lines follow the last hunk, where that is
	// known (Compare); -1 where it is not (a patch).
	After int
	// Binary: git says the file is binary, and shows it no hunks.
	Binary bool
}

// Name is the file's name: its new one, or its old one when it was
// deleted.
func (f File) Name() string {
	if f.New != "" {
		return f.New
	}
	return f.Old
}

// Stat is how many lines the change adds to the file and removes from it.
func (f File) Stat() (added, removed int) {
	for _, h := range f.Hunks {
		for _, l := range h.Lines {
			switch l.Op {
			case Added:
				added++
			case Removed:
				removed++
			}
		}
	}
	return added, removed
}

// Diff is a change to one or more files. A Diff Parse or Compare made is
// shared: a caller does not change it.
type Diff struct {
	Files []File
}

// Parse reads a unified diff: what git diff, diff -u and most tools write,
// one file or several. Lines it does not know are skipped (git's index and
// mode lines, a mail's headers before the first file), and a hunk whose
// lines run out early ends there. Each file is lexed as lang says, a
// language's name or a file name, or by its own name when lang is "".
func Parse(patch, lang string) *Diff {
	return cached("p\x00"+lang+"\x00"+patch, func() *Diff { return build(parse(patch), lang) })
}

// Compare is the change from old to new, as a unified diff of them would
// show it: each run of changes with context unchanged lines around it,
// runs closer than twice that in one hunk (git's default context is 3; a
// negative one is that). name names the texts' file, which lexes them
// when lang is "".
func Compare(name, old, new, lang string, context int) *Diff {
	if context < 0 {
		context = 3
	}
	key := "c\x00" + name + "\x00" + lang + "\x00" + strconv.Itoa(context) + "\x00" + old + "\x00\x00" + new
	return cached(key, func() *Diff { return build([]rawFile{compare(name, old, new, context)}, lang) })
}

// raw is a line as parse and compare read it, before it is lexed.
type raw struct {
	op       Op
	old, new int
	text     string
}

// rawHunk is a hunk as read, its lines still raw.
type rawHunk struct {
	Hunk
	raw []raw
}

// rawFile is a file as read.
type rawFile struct {
	File
	hunks []rawHunk
}

func parse(patch string) []rawFile {
	var files []rawFile
	cur := func() *rawFile { return &files[len(files)-1] }
	start := func() { files = append(files, rawFile{File: File{After: -1}}) }
	inHunk := func() *rawHunk {
		if len(files) == 0 || len(cur().hunks) == 0 {
			return nil
		}
		return &cur().hunks[len(cur().hunks)-1]
	}
	oldLeft, newLeft, o, n := 0, 0, 0, 0
	minus := false // the current file has its --- line
	// A final newline ends the last line, not a blank one after it.
	for _, l := range strings.Split(strings.TrimSuffix(strings.ReplaceAll(patch, "\r\n", "\n"), "\n"), "\n") {
		if h := inHunk(); h != nil && (oldLeft > 0 || newLeft > 0) {
			op, ok := Context, true
			text := l
			switch {
			case l == "":
				// A blank context line whose space a tool dropped.
			case l[0] == ' ':
				text = l[1:]
			case l[0] == '-' && oldLeft > 0:
				op, text = Removed, l[1:]
			case l[0] == '+' && newLeft > 0:
				op, text = Added, l[1:]
			case l[0] == '\\':
				continue // \ No newline at end of file
			default:
				ok = false
			}
			if ok && (op != Context || oldLeft > 0 && newLeft > 0) {
				r := raw{op: op, text: text}
				if op != Added {
					o++
					r.old = o
					oldLeft--
				}
				if op != Removed {
					n++
					r.new = n
					newLeft--
				}
				h.raw = append(h.raw, r)
				continue
			}
			oldLeft, newLeft = 0, 0
		}
		switch {
		case strings.HasPrefix(l, "diff --git "):
			start()
			minus = false
			if a, b, ok := gitNames(l[len("diff --git "):]); ok {
				cur().Old, cur().New = a, b
			}
		case strings.HasPrefix(l, "--- "):
			if len(files) == 0 || len(cur().hunks) > 0 || minus {
				start()
			}
			minus = true
			cur().Old = name(l[4:], "a/")
		case strings.HasPrefix(l, "+++ ") && len(files) > 0:
			cur().New = name(l[4:], "b/")
		case strings.HasPrefix(l, "@@ "):
			h, ok := hunkHeader(l)
			if !ok {
				continue
			}
			if len(files) == 0 {
				start()
			}
			cur().hunks = append(cur().hunks, rawHunk{Hunk: h})
			oldLeft, newLeft = h.OldLines, h.NewLines
			o, n = h.OldStart-1, h.NewStart-1
			if h.OldLines == 0 {
				o = h.OldStart
			}
			if h.NewLines == 0 {
				n = h.NewStart
			}
		case strings.HasPrefix(l, "Binary files ") && strings.HasSuffix(l, " differ"):
			// git's comes in its file's header; diff -r's, and git's
			// after a file of hunks, is a file of its own.
			if len(files) == 0 || len(cur().hunks) > 0 || minus {
				start()
				minus = false
				names := l[len("Binary files ") : len(l)-len(" differ")]
				if a, b, ok := strings.Cut(names, " and "); ok {
					cur().Old, cur().New = name(a, "a/"), name(b, "b/")
				}
			}
			cur().Binary = true
		case len(files) == 0:
		case strings.HasPrefix(l, "new file mode"):
			cur().Old = ""
		case strings.HasPrefix(l, "deleted file mode"):
			cur().New = ""
		case strings.HasPrefix(l, "rename from "):
			cur().Old = l[len("rename from "):]
		case strings.HasPrefix(l, "rename to "):
			cur().New = l[len("rename to "):]
		case l == "GIT binary patch":
			cur().Binary = true
		}
	}
	// What each hunk leaves out before it.
	for i := range files {
		to := 1
		for j := range files[i].hunks {
			h := &files[i].hunks[j]
			from := h.OldStart
			if h.OldLines == 0 {
				from++
			}
			h.Skipped = max(from-to, 0)
			to = from + h.OldLines
		}
	}
	return files
}

// gitNames reads diff --git's "a/x b/y": the names after their a/ and b/.
// A name with a space is ambiguous there; the --- and +++ lines that follow
// say it plainly.
func gitNames(s string) (a, b string, ok bool) {
	if strings.HasPrefix(s, `"`) {
		q, err := strconv.QuotedPrefix(s)
		if err != nil {
			return "", "", false
		}
		a = name(q, "a/")
		return a, name(strings.TrimSpace(s[len(q):]), "b/"), true
	}
	i := strings.LastIndex(s, " b/")
	if i < 0 {
		return "", "", false
	}
	return name(s[:i], "a/"), name(s[i+1:], "b/"), true
}

// name is a file's name as a ---, a +++ or diff --git line gives it:
// without its timestamp, its prefix (a/ or b/) and quotes; "" for
// /dev/null.
func name(s, prefix string) string {
	if i := strings.IndexByte(s, '\t'); i >= 0 {
		s = s[:i]
	}
	if strings.HasPrefix(s, `"`) {
		if u, err := strconv.Unquote(s); err == nil {
			s = u
		}
	}
	if s == "/dev/null" {
		return ""
	}
	return strings.TrimPrefix(s, prefix)
}

// hunkHeader reads "@@ -a,b +c,d @@ section"; a count left out is 1.
func hunkHeader(l string) (Hunk, bool) {
	rest, ok := strings.CutPrefix(l, "@@ -")
	if !ok {
		return Hunk{}, false
	}
	ranges, section, ok := strings.Cut(rest, " @@")
	if !ok {
		return Hunk{}, false
	}
	oldR, newR, ok := strings.Cut(ranges, " +")
	if !ok {
		return Hunk{}, false
	}
	var h Hunk
	var okO, okN bool
	h.OldStart, h.OldLines, okO = startCount(oldR)
	h.NewStart, h.NewLines, okN = startCount(newR)
	h.Section = strings.TrimPrefix(section, " ")
	return h, okO && okN
}

func startCount(s string) (start, n int, ok bool) {
	a, b, comma := strings.Cut(s, ",")
	start, err := strconv.Atoi(a)
	if err != nil || start < 0 {
		return 0, 0, false
	}
	n = 1
	if comma {
		if n, err = strconv.Atoi(b); err != nil || n < 0 {
			return 0, 0, false
		}
	}
	return start, n, true
}

// compare is Compare's file: the lines of old and new, diffed, in hunks of
// context.
func compare(name, old, new string, context int) rawFile {
	a, b := lines(old), lines(new)
	f := rawFile{File: File{Old: name, New: name}}
	changes := lcs.DiffLines(a, b)
	to := 0 // the old line after the last hunk, from 0
	for i := 0; i < len(changes); {
		// A hunk: the changes that come within twice the context of each
		// other.
		j := i + 1
		for j < len(changes) && changes[j].Start-changes[j-1].End <= 2*context {
			j++
		}
		first, last := changes[i], changes[j-1]
		oldFrom, oldTo := max(first.Start-context, 0), min(last.End+context, len(a))
		newFrom, newTo := oldFrom+first.ReplStart-first.Start, oldTo+last.ReplEnd-last.End
		h := rawHunk{Hunk: Hunk{OldLines: oldTo - oldFrom, NewLines: newTo - newFrom, Skipped: oldFrom - to}}
		h.OldStart, h.NewStart = oldFrom+1, newFrom+1
		if h.OldLines == 0 {
			h.OldStart = oldFrom
		}
		if h.NewLines == 0 {
			h.NewStart = newFrom
		}
		o, n := oldFrom, newFrom
		for _, c := range changes[i:j] {
			for ; o < c.Start; o, n = o+1, n+1 {
				h.raw = append(h.raw, raw{op: Context, old: o + 1, new: n + 1, text: a[o]})
			}
			for ; o < c.End; o++ {
				h.raw = append(h.raw, raw{op: Removed, old: o + 1, text: a[o]})
			}
			for ; n < c.ReplEnd; n++ {
				h.raw = append(h.raw, raw{op: Added, new: n + 1, text: b[n]})
			}
		}
		for ; o < oldTo; o, n = o+1, n+1 {
			h.raw = append(h.raw, raw{op: Context, old: o + 1, new: n + 1, text: a[o]})
		}
		f.hunks = append(f.hunks, h)
		to = oldTo
		i = j
	}
	f.After = len(a) - to
	return f
}

// lines are a text's lines: a final newline ends the last one rather than
// starting another.
func lines(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	if s == "" {
		return nil
	}
	return strings.Split(strings.TrimSuffix(s, "\n"), "\n")
}

// build lexes what parse or compare read and marks the changed words.
func build(files []rawFile, lang string) *Diff {
	d := &Diff{}
	for _, rf := range files {
		f := rf.File
		l := lang
		if l == "" {
			l = f.Name()
		}
		for _, rh := range rf.hunks {
			h := rh.Hunk
			h.Lines = lex(rh.raw, l)
			markWords(h.Lines)
			f.Hunks = append(f.Hunks, h)
		}
		d.Files = append(d.Files, f)
	}
	return d
}

// lex makes a hunk's lines, each side's lexed as one text so that a string
// or a comment that spans lines is one; a context line takes the new
// side's tokens.
func lex(rs []raw, lang string) []Line {
	out := make([]Line, len(rs))
	for i, r := range rs {
		out[i] = Line{Op: r.op, Old: r.old, New: r.new}
	}
	for _, side := range []Op{Removed, Added} {
		var idx []int
		var text strings.Builder
		for i, r := range rs {
			if r.op == Context || r.op == side {
				idx = append(idx, i)
				text.WriteString(r.text)
				text.WriteByte('\n')
			}
		}
		if len(idx) == 0 {
			continue
		}
		toks := highlight.Lines(text.String(), lang)
		for k, i := range idx {
			if len(toks) == len(idx) {
				out[i].Tokens = toks[k]
			} else {
				out[i].Tokens = highlight.Lines(rs[i].text, "")[0]
			}
		}
	}
	return out
}

// distance is the most of a pair of lines, outside white space, that may
// change for their changed words to be marked, as delta has it: more, and
// the line is a new one, which marks would only speckle.
const distance = 0.6

// markWords marks the changed words of each removed line and the added
// line that replaces it: in each run of removed lines followed by added
// ones, the first of each, the second of each, and so on.
func markWords(ls []Line) {
	for i := 0; i < len(ls); {
		if ls[i].Op != Removed {
			i++
			continue
		}
		j := i
		for j < len(ls) && ls[j].Op == Removed {
			j++
		}
		k := j
		for k < len(ls) && ls[k].Op == Added {
			k++
		}
		for m := 0; i+m < j && j+m < k; m++ {
			ls[i+m].Changed, ls[j+m].Changed = changedWords(ls[i+m].Text(), ls[j+m].Text())
		}
		i = k
	}
}

// changedWords are the byte ranges of a and of b that the other does not
// have, word by word; nil, nil when too much of the two changed
// (distance).
func changedWords(a, b string) (ra, rb [][2]int) {
	wa, wb := words(a), words(b)
	offA, offB := offsets(wa), offsets(wb)
	changed := 0
	for _, c := range tidy(lcs.DiffLines(wa, wb), wa, wb) {
		if c.End > c.Start {
			ra = appendRange(ra, offA[c.Start], offA[c.End])
			changed += solid(a[offA[c.Start]:offA[c.End]])
		}
		if c.ReplEnd > c.ReplStart {
			rb = appendRange(rb, offB[c.ReplStart], offB[c.ReplEnd])
			changed += solid(b[offB[c.ReplStart]:offB[c.ReplEnd]])
		}
	}
	whole := solid(a) + solid(b)
	if whole == 0 || float64(changed)/float64(whole) > distance {
		return nil, nil
	}
	return ra, rb
}

// tidy slides each change that only inserts, or only deletes, as far
// right as it goes (while the word it starts with is the word after it),
// and joins changes that then touch: of two equally short diffs it keeps
// the one with fewer, whole runs, so that "{url}" → "{url} after {tries}
// tries" marks " after {tries} tries", not "} after {tries" and " tries".
func tidy(cs []lcs.Diff, a, b []string) []lcs.Diff {
	out := make([]lcs.Diff, 0, len(cs))
	for i, c := range cs {
		// The most it may move: up to the next change, or the end.
		endA, endB := len(a), len(b)
		if i+1 < len(cs) {
			endA, endB = cs[i+1].Start, cs[i+1].ReplStart
		}
		switch {
		case c.Start == c.End:
			for c.ReplEnd < endB && c.End < endA && b[c.ReplStart] == b[c.ReplEnd] {
				c.Start, c.End, c.ReplStart, c.ReplEnd = c.Start+1, c.End+1, c.ReplStart+1, c.ReplEnd+1
			}
		case c.ReplStart == c.ReplEnd:
			for c.End < endA && c.ReplEnd < endB && a[c.Start] == a[c.End] {
				c.Start, c.End, c.ReplStart, c.ReplEnd = c.Start+1, c.End+1, c.ReplStart+1, c.ReplEnd+1
			}
		}
		if n := len(out); n > 0 && out[n-1].End == c.Start && out[n-1].ReplEnd == c.ReplStart {
			out[n-1].End, out[n-1].ReplEnd = c.End, c.ReplEnd
			continue
		}
		out = append(out, c)
	}
	return out
}

// words split a line into words (letters, digits and _), runs of white
// space, and every other character on its own.
func words(s string) []string {
	var out []string
	for i := 0; i < len(s); {
		r, n := utf8.DecodeRuneInString(s[i:])
		j := i + n
		if class := runeClass(r); class != 0 {
			for j < len(s) {
				r2, n2 := utf8.DecodeRuneInString(s[j:])
				if runeClass(r2) != class {
					break
				}
				j += n2
			}
		}
		out = append(out, s[i:j])
		i = j
	}
	return out
}

// runeClass is 1 for a word's character, 2 for white space, and 0 for any
// other, which stands alone.
func runeClass(r rune) int {
	switch {
	case r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r):
		return 1
	case unicode.IsSpace(r):
		return 2
	}
	return 0
}

// offsets are where each word starts, and after them the line's end.
func offsets(ws []string) []int {
	out := make([]int, len(ws)+1)
	for i, w := range ws {
		out[i+1] = out[i] + len(w)
	}
	return out
}

// appendRange adds [from, to) to rs, joining it to the last range when
// they touch.
func appendRange(rs [][2]int, from, to int) [][2]int {
	if n := len(rs); n > 0 && rs[n-1][1] == from {
		rs[n-1][1] = to
		return rs
	}
	return append(rs, [2]int{from, to})
}

// solid is how many bytes of s are not white space.
func solid(s string) int {
	n := 0
	for _, r := range s {
		if !unicode.IsSpace(r) {
			n += utf8.RuneLen(r)
		}
	}
	return n
}

// cache keeps what Parse and Compare made, since a view is built again at
// every key: up to cacheSize diffs, then it starts over.
var cache = struct {
	sync.Mutex
	m map[string]*Diff
}{m: map[string]*Diff{}}

const cacheSize = 32

func cached(key string, make func() *Diff) *Diff {
	cache.Lock()
	d, ok := cache.m[key]
	cache.Unlock()
	if ok {
		return d
	}
	d = make()
	cache.Lock()
	if len(cache.m) >= cacheSize {
		cache.m = map[string]*Diff{}
	}
	cache.m[key] = d
	cache.Unlock()
	return d
}
