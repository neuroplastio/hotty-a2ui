package theme

import "testing"

func TestByNameAndNext(t *testing.T) {
	if th, ok := ByName("nord"); !ok || th.Name != "Nord" {
		t.Fatalf("ByName(nord) = %v, %v", th.Name, ok)
	}
	if _, ok := ByName("no such"); ok {
		t.Fatal("ByName found a theme that is not there")
	}
	last := All[len(All)-1]
	if got := Next(last); got.Name != All[0].Name {
		t.Fatalf("Next(last) = %q, want the first", got.Name)
	}
	if got := Next(Default); got.Name != All[1].Name {
		t.Fatalf("Next(Default) = %q", got.Name)
	}
}

func TestColour(t *testing.T) {
	if got := All[1].Colour("error"); got != All[1].Error {
		t.Fatalf("Colour(error) = %q", got)
	}
	if got := All[1].Colour("nonsense"); got != "" {
		t.Fatalf("an unknown role has colour %q", got)
	}
}
