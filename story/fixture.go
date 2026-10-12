package story

import (
	"io/fs"
	"testing/fstest"
)

// FixtureFS is a filesystem fixture with directories and files, for the
// storybook and tests of HottyFilePicker (profile §6.31).
var FixtureFS = fstest.MapFS{
	// Root directories
	"docs":                 &fstest.MapFile{Mode: fs.ModeDir | 0755},
	"docs/api.md":          &fstest.MapFile{Data: []byte("# API Reference\n\nEndpoints and specifications.\n"), Mode: 0644},
	"docs/architecture.md": &fstest.MapFile{Data: []byte("# Architecture\n\nSystem design and overview.\n"), Mode: 0644},
	"docs/changelog.md":    &fstest.MapFile{Data: []byte("# Changelog\n\n- Initial release\n"), Mode: 0644},
	"docs/contributing.md": &fstest.MapFile{Data: []byte("# Contributing\n\nPull requests are welcome.\n"), Mode: 0644},
	"docs/faq.md":          &fstest.MapFile{Data: []byte("# FAQ\n\nFrequently asked questions.\n"), Mode: 0644},
	"docs/guide.md":        &fstest.MapFile{Data: []byte("# Guide\n\nHow to use the project.\n"), Mode: 0644},
	"docs/intro.md":        &fstest.MapFile{Data: []byte("# Introduction\n\nWelcome to the documentation.\n"), Mode: 0644},
	"docs/roadmap.md":      &fstest.MapFile{Data: []byte("# Roadmap\n\nUpcoming milestones.\n"), Mode: 0644},

	"src":                  &fstest.MapFile{Mode: fs.ModeDir | 0755},
	"src/api":              &fstest.MapFile{Mode: fs.ModeDir | 0755},
	"src/api/api.go":       &fstest.MapFile{Data: []byte("package api\n"), Mode: 0644},
	"src/auth":             &fstest.MapFile{Mode: fs.ModeDir | 0755},
	"src/auth/auth.go":     &fstest.MapFile{Data: []byte("package auth\n"), Mode: 0644},
	"src/cli":              &fstest.MapFile{Mode: fs.ModeDir | 0755},
	"src/cli/cli.go":       &fstest.MapFile{Data: []byte("package cli\n"), Mode: 0644},
	"src/config":           &fstest.MapFile{Mode: fs.ModeDir | 0755},
	"src/config/config.go": &fstest.MapFile{Data: []byte("package config\n"), Mode: 0644},
	"src/db":               &fstest.MapFile{Mode: fs.ModeDir | 0755},
	"src/db/db.go":         &fstest.MapFile{Data: []byte("package db\n"), Mode: 0644},
	"src/routes":           &fstest.MapFile{Mode: fs.ModeDir | 0755},
	"src/routes/routes.go": &fstest.MapFile{Data: []byte("package routes\n"), Mode: 0644},
	"src/view":             &fstest.MapFile{Mode: fs.ModeDir | 0755},
	"src/view/view.go":     &fstest.MapFile{Data: []byte("package view\n\nfunc Render() {}\n"), Mode: 0644},
	"src/app.go":           &fstest.MapFile{Data: []byte("package src\n\nfunc App() {}\n"), Mode: 0644},
	"src/server.go":        &fstest.MapFile{Data: []byte("package src\n\nfunc Serve() {}\n"), Mode: 0644},
	"src/util.go":          &fstest.MapFile{Data: []byte("package src\n\nfunc Helper() {}\n"), Mode: 0644},

	// Root files (ordered alphabetically after go.mod)
	"go.mod":       &fstest.MapFile{Data: []byte("module example.com/project\n\ngo 1.26.0\n"), Mode: 0644},
	"license":      &fstest.MapFile{Data: []byte("MIT License\n\nCopyright (c) 2026\n"), Mode: 0644},
	"main.go":      &fstest.MapFile{Data: []byte("package main\n\nimport \"fmt\"\n\nfunc main() {\n\tfmt.Println(\"Hello, world!\")\n}\n"), Mode: 0644},
	"Makefile":     &fstest.MapFile{Data: []byte("all:\n\tgo build ./...\n\ntest:\n\tgo test ./...\n"), Mode: 0644},
	"package.json": &fstest.MapFile{Data: []byte("{\n  \"name\": \"project\",\n  \"version\": \"1.0.0\"\n}\n"), Mode: 0644},
	"README.md":    &fstest.MapFile{Data: []byte("# Project\n\nA demonstration repository for the HottyFilePicker component.\n"), Mode: 0644},
	"server.go":    &fstest.MapFile{Data: []byte("package main\n\nfunc startServer() error {\n\treturn nil\n}\n"), Mode: 0644},
	"todo.txt":     &fstest.MapFile{Data: []byte("1. Implement features\n2. Add tests\n3. Release\n"), Mode: 0644},
	"util.go":      &fstest.MapFile{Data: []byte("package main\n\nfunc helper() string {\n\treturn \"ok\"\n}\n"), Mode: 0644},
	"version.txt":  &fstest.MapFile{Data: []byte("1.0.0-beta.1\n"), Mode: 0644},
}
