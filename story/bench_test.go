package story

import "testing"

// A story switch in the storybook (Book.open): find the story, a new run,
// play it. The stories, the catalogs and their schemas are read once, so
// that a switch costs only its own story's messages.
func BenchmarkSwitch(b *testing.B) {
	for b.Loop() {
		if _, err := Start(Find("00_simple-login-form")); err != nil {
			b.Fatal(err)
		}
	}
}
