package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/neuroplastio/hotty-go/hottytest"

	"github.com/neuroplastio/hotty-a2ui/rendition/theme"
)

// pictures writes small PNGs to a new directory.
func pictures(t *testing.T, names ...string) string {
	t.Helper()
	dir := t.TempDir()
	for i, name := range names {
		img := image.NewRGBA(image.Rect(0, 0, 4, 3))
		img.Set(0, 0, color.RGBA{R: uint8(i * 40), A: 255})
		var b bytes.Buffer
		if err := png.Encode(&b, img); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, name), b.Bytes(), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// runReview runs review on a test host until the test ends, with a
// feedback box for each of boxes (one without).
func runReview(t *testing.T, h *hottytest.Host, boxes []string, args ...string) {
	t.Helper()
	shots, dir, err := find(args)
	if err != nil {
		t.Fatal(err)
	}
	n, err := loadNotes(filepath.Join(dir, "feedback.md"))
	if err != nil {
		t.Fatal(err)
	}
	a, err := newApp(shots, n, filepath.Join(dir, "brief.md"), boxes, theme.Theme{})
	if err != nil {
		t.Fatal(err)
	}
	p := tea.NewProgram(a, tea.WithInput(h), tea.WithOutput(a.s.Watch(h)), tea.WithWindowSize(120, 40),
		tea.WithoutSignalHandler(), tea.WithEnvironment([]string{"TERM=xterm-256color"}))
	a.s.Attach(p.Send)
	done := make(chan struct{})
	go func() {
		defer close(done)
		if _, err := p.Run(); err != nil {
			t.Error(err)
		}
	}()
	t.Cleanup(func() {
		p.Quit()
		<-done
	})
}

// eventually waits for cond, a few seconds at most.
func eventually(t *testing.T, what string, cond func() bool) {
	t.Helper()
	for end := time.Now().Add(5 * time.Second); time.Now().Before(end); time.Sleep(5 * time.Millisecond) {
		if cond() {
			return
		}
	}
	t.Fatalf("never: %s", what)
}

// TestReview: the first picture shows, sent in band, with what to look for
// in it; Save writes its note to feedback.md; another picked in the list
// shows with its own brief and an empty box, and going back keeps what was
// typed about it, without Save.
func TestReview(t *testing.T) {
	dir := pictures(t, "a.png", "b.png")
	brief := "# Brief\n\n## a.png\n\nThe red dot, top left.\n\n## b.png\n\nNothing *red*.\n"
	if err := os.WriteFile(filepath.Join(dir, "brief.md"), []byte(brief), 0o644); err != nil {
		t.Fatal(err)
	}
	h := hottytest.New(t, hottytest.Size(120, 40))
	runReview(t, h, nil, dir)
	eventually(t, "the surface", func() bool { return h.Surface(surfaceID) != nil })
	s := h.Surface(surfaceID)
	if src, _ := s.Attr("shot", "src"); src != "cid:shot-0" {
		t.Fatalf("the picture is %q", src)
	}
	if got := s.TextOf("brief"); !strings.Contains(got, "What to look for") || !strings.Contains(got, "The red dot, top left.") {
		t.Fatalf("a.png's brief: %q", got)
	}
	want, _ := os.ReadFile(filepath.Join(dir, "a.png"))
	if mime, got, ok := h.Resource("shot-0"); !ok || mime != "image/png" || !bytes.Equal(got, want) {
		t.Fatalf("the host has shot-0 %v as %q, %d bytes", ok, mime, len(got))
	}
	if _, _, ok := h.Resource("shot-1"); ok {
		t.Fatal("b.png was sent before it was shown")
	}

	feedback := filepath.Join(dir, "feedback.md")
	read := func() string { b, _ := os.ReadFile(feedback); return string(b) }
	if err := h.Fill(surfaceID, "note0", "Too dark."); err != nil {
		t.Fatal(err)
	}
	if err := h.Click(surfaceID, "save"); err != nil {
		t.Fatal(err)
	}
	eventually(t, "the note saved", func() bool { return read() == "# Feedback\n\n## a.png\n\nToo dark.\n" })
	eventually(t, "the list says it", func() bool { return strings.Contains(s.TextOf("files"), "✎ Too dark.") })

	if err := h.Click(surfaceID, "files~i1"); err != nil {
		t.Fatal(err)
	}
	eventually(t, "b.png shown", func() bool {
		src, _ := s.Attr("shot", "src")
		_, _, sent := h.Resource("shot-1")
		return src == "cid:shot-1" && sent
	})
	if v, _ := s.Value("note0"); v != "" {
		t.Fatalf("b.png's box has %q", v)
	}
	if got := s.TextOf("brief"); !strings.Contains(got, "Nothing red.") {
		t.Fatalf("b.png's brief: %q", got)
	}
	if err := h.Fill(surfaceID, "note0", "Fine.\nThe gap is wide."); err != nil {
		t.Fatal(err)
	}
	if err := h.Click(surfaceID, "files~i0"); err != nil {
		t.Fatal(err)
	}
	eventually(t, "a.png back, its note in the box", func() bool {
		v, _ := s.Value("note0")
		return v == "Too dark."
	})
	if got := read(); got != "# Feedback\n\n## a.png\n\nToo dark.\n\n## b.png\n\nFine.\nThe gap is wide.\n" {
		t.Fatalf("feedback.md:\n%s", got)
	}
}

// TestNotes: the file read back is the notes written; a blank note goes.
func TestNotes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "feedback.md")
	n, _ := loadNotes(path)
	n.set("x.png", "", "One.\n\nTwo.")
	n.set("y.png", "A", "Three.")
	n.set("y.png", "B", "Four.")
	n.set("z.png", "", "  ")
	if err := n.write(); err != nil {
		t.Fatal(err)
	}
	m, err := loadNotes(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(m.order, " ") != "x.png y.png" || m.get("x.png", "") != "One.\n\nTwo." ||
		m.get("y.png", "A") != "Three." || m.get("y.png", "B") != "Four." {
		t.Fatalf("read back %q %q", m.order, m.text)
	}
	if got := m.summary("y.png"); got != "A: Three. · B: Four." {
		t.Fatalf("y.png's summary: %q", got)
	}
	m.set("x.png", "", "")
	if strings.Join(m.order, " ") != "y.png" {
		t.Fatalf("after a blank note: %q", m.order)
	}
}

// TestBoxes: with boxes A and B, a picture has a box for each, and its
// section in feedback.md a part for each that has a note.
func TestBoxes(t *testing.T) {
	dir := pictures(t, "a.png", "b.png")
	h := hottytest.New(t, hottytest.Size(120, 40))
	runReview(t, h, []string{"A", "B"}, dir)
	eventually(t, "the surface", func() bool { return h.Surface(surfaceID) != nil })
	s := h.Surface(surfaceID)
	for i, want := range []string{"Feedback on A", "Feedback on B"} {
		if got := s.TextOf(fmt.Sprintf("note%d~l", i)); got != want {
			t.Fatalf("box %d is %q", i, got)
		}
	}
	feedback := filepath.Join(dir, "feedback.md")
	read := func() string { b, _ := os.ReadFile(feedback); return string(b) }
	if err := h.Fill(surfaceID, "note0", "Too tight."); err != nil {
		t.Fatal(err)
	}
	if err := h.Fill(surfaceID, "note1", "Better."); err != nil {
		t.Fatal(err)
	}
	if err := h.Click(surfaceID, "save"); err != nil {
		t.Fatal(err)
	}
	eventually(t, "both saved", func() bool {
		return read() == "# Feedback\n\n## a.png\n\n### A\n\nToo tight.\n\n### B\n\nBetter.\n"
	})
	if err := h.Click(surfaceID, "files~i1"); err != nil {
		t.Fatal(err)
	}
	eventually(t, "b.png's empty boxes", func() bool {
		a, _ := s.Value("note0")
		b, _ := s.Value("note1")
		src, _ := s.Attr("shot", "src")
		return src == "cid:shot-1" && a == "" && b == ""
	})
	if err := h.Fill(surfaceID, "note1", "Only B."); err != nil {
		t.Fatal(err)
	}
	if err := h.Click(surfaceID, "files~i0"); err != nil {
		t.Fatal(err)
	}
	eventually(t, "a.png's notes back", func() bool {
		b, _ := s.Value("note1")
		return b == "Better."
	})
	if got := read(); !strings.HasSuffix(got, "## b.png\n\n### B\n\nOnly B.\n") {
		t.Fatalf("feedback.md:\n%s", got)
	}
}
