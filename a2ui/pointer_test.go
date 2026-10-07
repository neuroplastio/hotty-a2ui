package a2ui

import (
	"reflect"
	"testing"
)

func TestParsePointer(t *testing.T) {
	for _, c := range []struct {
		in   string
		want Pointer
	}{
		{"", Pointer{}},
		{"/", Pointer{}},
		{"/a", Pointer{"a"}},
		{"/a~1b/c~0d", Pointer{"a/b", "c~d"}},
		{"/items/0", Pointer{"items", "0"}},
		{"//a///b//", Pointer{"a", "b"}},
	} {
		got, err := ParsePointer(c.in)
		if err != nil {
			t.Fatalf("%q: %v", c.in, err)
		}
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("%q: got %q, want %q", c.in, got, c.want)
		}
	}
	for _, bad := range []string{"a", "/a~", "/a~2"} {
		if _, err := ParsePointer(bad); err == nil {
			t.Errorf("%q parsed", bad)
		}
	}
}

func TestPointerString(t *testing.T) {
	if s := (Pointer{"a/b", "c~d"}).String(); s != "/a~1b/c~0d" {
		t.Errorf("got %q", s)
	}
	if s := (Pointer{}).String(); s != "/" {
		t.Errorf("root: got %q", s)
	}
}

func TestJoin(t *testing.T) {
	for _, c := range [][3]string{
		{"", "/a", "/a"},
		{"/items/1", "/a", "/a"},
		{"", "a", "/a"},
		{"/", "a", "/a"},
		{"/items/1", "title", "/items/1/title"},
		{"/items/1", "", "/items/1"},
		{"", "", "/"},
	} {
		if got := Join(c[0], c[1]); got != c[2] {
			t.Errorf("Join(%q, %q) = %q, want %q", c[0], c[1], got, c[2])
		}
	}
}
