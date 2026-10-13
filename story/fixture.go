package story

import (
	_ "embed"
	"encoding/json"
	"testing/fstest"
)

// fixture is a small project's files, by path: the text of each.
//
//go:embed fixture.json
var fixture []byte

// FixtureFS is the files of fixture.json, a small project's, as a
// filesystem: the root the storybook grants as "project" (NewRun), for
// HottyFilePicker (profile §6.31). Bubbles' filepicker (ref) lists the
// same files.
var FixtureFS = func() fstest.MapFS {
	var files map[string]string
	if err := json.Unmarshal(fixture, &files); err != nil {
		panic("story: fixture.json: " + err.Error())
	}
	fsys := fstest.MapFS{}
	for name, text := range files {
		fsys[name] = &fstest.MapFile{Data: []byte(text), Mode: 0o644}
	}
	return fsys
}()
