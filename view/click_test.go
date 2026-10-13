package view_test

import (
	"slices"
	"testing"
)

// A run of clicks on an item acts once (ClickItem). A double click on a
// HottyList's item not selected selects it, then acts on it, and a third
// click does nothing more. A double click on the selected item acts once.
// A click on its own on the selected item still acts.
func TestClickRunActsOnce(t *testing.T) {
	c, actions := faveList(t)
	click := func(i, count int) {
		t.Helper()
		if err := c.ClickItem("l", i, count); err != nil {
			t.Fatal(err)
		}
	}
	click(1, 1)
	click(1, 2)
	click(1, 3)
	if sel := c.S.Data.Value("/sel"); sel != "nutella" || len(*actions) != 1 {
		t.Fatalf("a triple click on Nutella: selected %v, %d actions, want nutella and 1", sel, len(*actions))
	}
	click(1, 1)
	click(1, 2)
	if len(*actions) != 2 {
		t.Errorf("a double click on the selected item: %d actions, want 2", len(*actions))
	}
	click(1, 1)
	if len(*actions) != 3 {
		t.Errorf("a click on the selected item: %d actions, want 3", len(*actions))
	}
	click(-1, 1)
	if len(*actions) != 3 {
		t.Errorf("a click beside the items acted")
	}
}

// A double click on a HottyTree's branch opens it, as its first click
// does, and its second leaves it open; on a leaf, it acts once.
func TestClickRunOnBranch(t *testing.T) {
	c, actions := files(t, "")
	tree := c.V.Find("t")
	at := -1
	for i, n := range tree.Nodes {
		if n.Label == "view" {
			at = i
		}
	}
	for count := 1; count <= 2; count++ {
		if err := c.ClickItem("t", at, count); err != nil {
			t.Fatal(err)
		}
	}
	if open, _ := c.S.Data.Value("/open").([]any); !slices.Contains(open, any("src/view")) {
		t.Errorf("after a double click on view, open: %v", open)
	}
	tree = c.V.Find("t")
	for i, n := range tree.Nodes {
		if n.Label == "main.go" {
			at = i
		}
	}
	for count := 1; count <= 3; count++ {
		if err := c.ClickItem("t", at, count); err != nil {
			t.Fatal(err)
		}
	}
	if len(*actions) != 1 {
		t.Errorf("a triple click on main.go: %d actions, want 1", len(*actions))
	}
}

// A double click on a HottyFilePicker's directory opens it, and a third
// click, on the entry now at its place in the directory opened, does
// nothing.
func TestClickRunOnFile(t *testing.T) {
	c, _ := filePickerHarness(t, "")
	for count := 1; count <= 3; count++ {
		if err := c.ClickItem("root", 1, count); err != nil {
			t.Fatal(err)
		}
	}
	if got := c.S.Data.Value("/dir"); got != "src" {
		t.Errorf("/dir = %v, want src", got)
	}
	if e := c.V.Find("root"); e.Selected != 0 {
		t.Errorf("the third click selected entry %d in src", e.Selected)
	}
}
