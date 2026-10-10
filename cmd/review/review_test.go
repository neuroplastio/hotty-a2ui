package main

import (
	"bytes"
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

// runReview runs review on a test host until the test ends.
func runReview(t *testing.T, h *hottytest.Host, args ...string) {
	t.Helper()
	shots, dir, err := find(args)
	if err != nil {
		t.Fatal(err)
	}
	n, err := loadNotes(filepath.Join(dir, "feedback.md"))
	if err != nil {
		t.Fatal(err)
	}
	a, err := newApp(shots, n, theme.Theme{})
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

// TestReview: the first picture shows, sent in band; Save writes its note
// to feedback.md; another picked in the list shows with an empty box, and
// going back keeps what was typed about it, without Save.
func TestReview(t *testing.T) {
	dir := pictures(t, "a.png", "b.png")
	h := hottytest.New(t, hottytest.Size(120, 40))
	runReview(t, h, dir)
	eventually(t, "the surface", func() bool { return h.Surface(surfaceID) != nil })
	s := h.Surface(surfaceID)
	if src, _ := s.Attr("shot", "src"); src != "cid:shot-0" {
		t.Fatalf("the picture is %q", src)
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
	if err := h.Fill(surfaceID, "note", "Too dark."); err != nil {
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
	if v, _ := s.Value("note"); v != "" {
		t.Fatalf("b.png's box has %q", v)
	}
	if err := h.Fill(surfaceID, "note", "Fine.\nThe gap is wide."); err != nil {
		t.Fatal(err)
	}
	if err := h.Click(surfaceID, "files~i0"); err != nil {
		t.Fatal(err)
	}
	eventually(t, "a.png back, its note in the box", func() bool {
		v, _ := s.Value("note")
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
	n.set("x.png", "One.\n\nTwo.")
	n.set("y.png", "Three.")
	n.set("z.png", "  ")
	if err := n.write(); err != nil {
		t.Fatal(err)
	}
	m, err := loadNotes(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Join(m.order, " ") != "x.png y.png" || m.text["x.png"] != "One.\n\nTwo." || m.text["y.png"] != "Three." {
		t.Fatalf("read back %q %q", m.order, m.text)
	}
	m.set("x.png", "")
	if strings.Join(m.order, " ") != "y.png" {
		t.Fatalf("after a blank note: %q", m.order)
	}
}
