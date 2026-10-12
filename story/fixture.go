package story

import (
	"io/fs"
	"testing/fstest"
)

// FixtureFS is a filesystem fixture with directories and files, for the
// storybook and tests of HottyFilePicker (profile §6.31).
var FixtureFS = fstest.MapFS{
	"README.md":        &fstest.MapFile{Data: []byte("# Project\n\nA demonstration repository for the HottyFilePicker component.\n"), Mode: 0644},
	"go.mod":           &fstest.MapFile{Data: []byte("module example.com/project\n\ngo 1.26.0\n"), Mode: 0644},
	"main.go":          &fstest.MapFile{Data: []byte("package main\n\nfunc main() {}\n"), Mode: 0644},
	"license":          &fstest.MapFile{Data: []byte("MIT License\n"), Mode: 0644},
	"docs":             &fstest.MapFile{Mode: fs.ModeDir | 0755},
	"docs/intro.md":    &fstest.MapFile{Data: []byte("# Introduction\n\nWelcome to the documentation.\n"), Mode: 0644},
	"docs/guide.md":    &fstest.MapFile{Data: []byte("# Guide\n\nHow to use the project.\n"), Mode: 0644},
	"src":              &fstest.MapFile{Mode: fs.ModeDir | 0755},
	"src/app.go":       &fstest.MapFile{Data: []byte("package src\n"), Mode: 0644},
	"src/util.go":      &fstest.MapFile{Data: []byte("package src\n"), Mode: 0644},
	"src/view":         &fstest.MapFile{Mode: fs.ModeDir | 0755},
	"src/view/view.go": &fstest.MapFile{Data: []byte("package view\n"), Mode: 0644},
}
