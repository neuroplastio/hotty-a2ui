package diff_test

import (
	"fmt"
	"math/rand"
	"slices"
	"strings"
	"testing"

	"github.com/neuroplastio/hotty-a2ui/diff"
	"github.com/neuroplastio/hotty-a2ui/highlight"
)

const patch = `From 1a2b3c Mon Sep 17 00:00:00 2001
Subject: [PATCH] api: HEAD too

diff --git a/api/handler.go b/api/handler.go
index 3b18e51..a9c2f4d 100644
--- a/api/handler.go
+++ b/api/handler.go
@@ -1,7 +1,7 @@
 package api

 import (
-	"errors"
+	"fmt"
 	"net/http"
 )

@@ -20,6 +20,7 @@ func Handle(w http.ResponseWriter, r *http.Request) error {
 	if r.Method != http.MethodGet {
 		return errors.New("only GET")
 	}
+	w.Header().Set("Content-Type", "text/plain")
 	w.WriteHeader(http.StatusOK)
 	_, err := w.Write([]byte("ok\n"))
 	return err
diff --git a/NOTES b/NOTES
new file mode 100644
--- /dev/null
+++ b/NOTES
@@ -0,0 +1,2 @@
+HEAD is answered as GET is.
+No body.
\ No newline at end of file
diff --git a/logo.png b/logo.png
index 0000000..1111111 100644
Binary files a/logo.png and b/logo.png differ
`

func TestParse(t *testing.T) {
	d := diff.Parse(patch, "")
	if len(d.Files) != 3 {
		t.Fatalf("%d files, want 3", len(d.Files))
	}
	f := d.Files[0]
	if f.Old != "api/handler.go" || f.New != "api/handler.go" || f.After != -1 {
		t.Errorf("file: %q → %q, after %d", f.Old, f.New, f.After)
	}
	if len(f.Hunks) != 2 {
		t.Fatalf("%d hunks, want 2", len(f.Hunks))
	}
	h := f.Hunks[1]
	if h.Header() != "@@ -20,6 +20,7 @@" || h.Section != "func Handle(w http.ResponseWriter, r *http.Request) error {" {
		t.Errorf("header %q, section %q", h.Header(), h.Section)
	}
	// Lines 8 to 19 are left out between the two.
	if f.Hunks[0].Skipped != 0 || h.Skipped != 12 {
		t.Errorf("skipped %d and %d, want 0 and 12", f.Hunks[0].Skipped, h.Skipped)
	}
	add := h.Lines[3]
	if add.Op != diff.Added || add.Old != 0 || add.New != 23 || !strings.Contains(add.Text(), `w.Header().Set(`) {
		t.Errorf("the added line: %+v %q", add, add.Text())
	}
	if l := h.Lines[6]; l.Op != diff.Context || l.Old != 25 || l.New != 26 {
		t.Errorf("the last line: %+v", l)
	}
	if a, r := f.Stat(); a != 2 || r != 1 {
		t.Errorf("stat +%d -%d, want +2 -1", a, r)
	}
	// Tabs are spaces, and the file's name says Go.
	if got := f.Hunks[0].Lines[3].Text(); got != `    "errors"` {
		t.Errorf("a tab: %q", got)
	}
	if !slices.ContainsFunc(h.Lines[0].Tokens, func(k highlight.Token) bool { return k.Kind == highlight.Keyword && k.Text == "if" }) {
		t.Errorf("not lexed as Go: %+v", h.Lines[0].Tokens)
	}

	nf := d.Files[1]
	if nf.Old != "" || nf.New != "NOTES" || nf.Name() != "NOTES" || len(nf.Hunks) != 1 || len(nf.Hunks[0].Lines) != 2 {
		t.Errorf("the new file: %+v", nf)
	}
	if nf.Hunks[0].Lines[1].New != 2 || nf.Hunks[0].Skipped != 0 {
		t.Errorf("the new file's lines: %+v", nf.Hunks[0])
	}
	if b := d.Files[2]; !b.Binary || b.Name() != "logo.png" || len(b.Hunks) != 0 {
		t.Errorf("the binary file: %+v", b)
	}
	if diff.Parse(patch, "") != d {
		t.Error("parsed twice, not shared")
	}
}

// TestParseBare: hunks with no file around them, plain diff -u names, a
// hunk cut short, and a removed line that reads like a header.
func TestParseBare(t *testing.T) {
	d := diff.Parse("@@ -1,3 +1,3 @@\n a\n--- b\n+b\n c\n", "")
	if len(d.Files) != 1 || len(d.Files[0].Hunks) != 1 {
		t.Fatalf("%+v", d)
	}
	ls := d.Files[0].Hunks[0].Lines
	if len(ls) != 4 || ls[1].Op != diff.Removed || ls[1].Text() != "-- b" {
		t.Errorf("lines: %+v", ls)
	}
	d = diff.Parse("--- old.txt\t2026-10-09 10:00:00\n+++ new.txt\t2026-10-09 10:01:00\n@@ -1,4 +1,2 @@\n a\n-b\n", "")
	f := d.Files[0]
	if f.Old != "old.txt" || f.New != "new.txt" || len(f.Hunks[0].Lines) != 2 {
		t.Errorf("diff -u, cut short: %+v", f)
	}
	// diff -r's binary line is a file of its own.
	d = diff.Parse("--- a/x\n+++ b/x\n@@ -1 +1 @@\n-a\n+b\nBinary files a/i.png and b/i.png differ\n", "")
	if len(d.Files) != 2 || d.Files[0].Binary || !d.Files[1].Binary || d.Files[1].Old != "i.png" || d.Files[1].New != "i.png" {
		t.Errorf("diff -r, a binary file: %+v", d.Files)
	}
}

// TestCompare: two texts in hunks, which give back both texts.
func TestCompare(t *testing.T) {
	var a []string
	for i := 1; i <= 30; i++ {
		a = append(a, fmt.Sprintf("line %d", i))
	}
	b := slices.Clone(a)
	b[1] = "line two"                   // in the first hunk
	b = slices.Insert(b, 5, "inserted") // close enough to join it
	b = slices.Delete(b, 20, 22)        // a second hunk
	old, new := strings.Join(a, "\n")+"\n", strings.Join(b, "\n")+"\n"
	d := diff.Compare("x.txt", old, new, "", 3)
	f := d.Files[0]
	if len(f.Hunks) != 2 {
		t.Fatalf("%d hunks, want 2: %+v", len(f.Hunks), f.Hunks)
	}
	if h := f.Hunks[0]; h.Header() != "@@ -1,8 +1,9 @@" || h.Skipped != 0 {
		t.Errorf("first hunk %s, skipped %d", h.Header(), h.Skipped)
	}
	if h := f.Hunks[1]; h.Header() != "@@ -17,8 +18,6 @@" || h.Skipped != 8 || f.After != 6 {
		t.Errorf("second hunk %s skipped %d, after %d", h.Header(), h.Skipped, f.After)
	}
	gotOld, gotNew := rebuild(a, f)
	if gotOld != old || gotNew != new {
		t.Errorf("rebuilt:\n%s\nwant\n%s", gotNew, new)
	}
	if len(diff.Compare("x.txt", old, old, "", 3).Files[0].Hunks) != 0 {
		t.Error("no change has hunks")
	}
}

// TestCompareRandom: random edits, every context, give back both texts.
func TestCompareRandom(t *testing.T) {
	r := rand.New(rand.NewSource(1))
	for n := range 200 {
		var a []string
		for i := range r.Intn(40) {
			a = append(a, fmt.Sprintf("%d", i%7))
		}
		b := slices.Clone(a)
		for range r.Intn(6) {
			switch i := r.Intn(len(b) + 1); r.Intn(3) {
			case 0:
				b = slices.Insert(b, i, "new")
			case 1:
				if i < len(b) {
					b = slices.Delete(b, i, i+1)
				}
			default:
				if i < len(b) {
					b[i] = "changed"
				}
			}
		}
		old, new := join(a), join(b)
		ctx := r.Intn(5)
		f := diff.Compare("", old, new, "", ctx).Files[0]
		if gotOld, gotNew := rebuild(a, f); gotOld != old || gotNew != new {
			t.Fatalf("%d (context %d):\n%q → %q\nrebuilt %q → %q", n, ctx, old, new, gotOld, gotNew)
		}
	}
}

func join(ls []string) string {
	if len(ls) == 0 {
		return ""
	}
	return strings.Join(ls, "\n") + "\n"
}

// rebuild makes both texts again from the old text's lines and the
// hunks: the old side checks the numbers and the skipped runs, the new
// side the changes.
func rebuild(a []string, f diff.File) (string, string) {
	var old, new []string
	at := 0 // the next old line, from 0
	for _, h := range f.Hunks {
		old = append(old, a[at:at+h.Skipped]...)
		new = append(new, a[at:at+h.Skipped]...)
		at += h.Skipped
		for _, l := range h.Lines {
			if l.Op != diff.Added {
				if l.Old != at+1 {
					return fmt.Sprintf("line %d numbered %d", at+1, l.Old), ""
				}
				old = append(old, l.Text())
				at++
			}
			if l.Op != diff.Removed {
				new = append(new, l.Text())
			}
		}
	}
	old = append(old, a[at:]...)
	new = append(new, a[at:]...)
	if f.After >= 0 && len(a)-at != f.After {
		return fmt.Sprintf("after %d, not %d", f.After, len(a)-at), ""
	}
	return join(old), join(new)
}

func TestWords(t *testing.T) {
	d := diff.Compare("h.go", "package api\n\nfunc f() error {\n\treturn errors.New(\"only GET\")\n}\n",
		"package api\n\nfunc f() error {\n\treturn fmt.Errorf(\"only GET, not %s\", m)\n}\n", "", 3)
	ls := d.Files[0].Hunks[0].Lines
	var rm, add diff.Line
	for _, l := range ls {
		switch l.Op {
		case diff.Removed:
			rm = l
		case diff.Added:
			add = l
		}
	}
	if got := marked(rm); !slices.Equal(got, []string{"errors", "New"}) {
		t.Errorf("removed marks %q in %q", got, rm.Text())
	}
	// The closing quote is the one after %s: the words around it are new.
	if got := marked(add); !slices.Equal(got, []string{"fmt", "Errorf", ", not %s", ", m"}) {
		t.Errorf("added marks %q in %q", got, add.Text())
	}

	// Of two equally short diffs, the one that marks one run.
	d = diff.Compare("", "raise E(f\"gave up on {url}\")\n", "raise E(f\"gave up on {url} after {tries} tries\")\n", "", 3)
	if got := marked(d.Files[0].Hunks[0].Lines[1]); !slices.Equal(got, []string{" after {tries} tries"}) {
		t.Errorf("an insertion marks %q", got)
	}
	// The story's: a shorter line beside a longer one still marks its
	// words (distance).
	d = diff.Compare("", "return errors.New(\"only GET\")\n", "return fmt.Errorf(\"only GET or HEAD, not %s\", r.Method)\n", "", 3)
	if got := marked(d.Files[0].Hunks[0].Lines[0]); !slices.Equal(got, []string{"errors", "New"}) {
		t.Errorf("errors.New marks %q", got)
	}

	// Too little in common: the whole line changed, no word marks.
	d = diff.Compare("", "alpha beta gamma\n", "one two three four\n", "", 3)
	for _, l := range d.Files[0].Hunks[0].Lines {
		if l.Changed != nil {
			t.Errorf("%q marks %v", l.Text(), l.Changed)
		}
	}
}

func marked(l diff.Line) []string {
	var out []string
	for _, r := range l.Changed {
		out = append(out, l.Text()[r[0]:r[1]])
	}
	return out
}

func TestSegments(t *testing.T) {
	l := diff.Line{
		Tokens:  []highlight.Token{{Text: "return ", Kind: highlight.Keyword}, {Text: "errors.New", Kind: highlight.Plain}},
		Changed: [][2]int{{4, 9}},
	}
	var got []string
	for _, s := range l.Segments() {
		mark := ""
		if s.Changed {
			mark = "*"
		}
		got = append(got, mark+s.Text)
	}
	if want := []string{"retu", "*rn ", "*er", "rors.New"}; !slices.Equal(got, want) {
		t.Errorf("segments %q, want %q", got, want)
	}
	if got := l.Segments()[1].Kind; got != highlight.Keyword {
		t.Errorf("a changed run keeps its kind: %v", got)
	}
}
