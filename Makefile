# hotty-a2ui. Toolchain from mise.toml (Go): `mise install` first.
GO          ?= mise x -- go
GOFMT       ?= mise x -- gofmt
# staticcheck 2026.2.1, the first that knows Go 1.26.
STATICCHECK ?= honnef.co/go/tools/cmd/staticcheck@v0.8.1

.PHONY: check fmt tidy vet lint test ref shot a2ui lexers icons gif clean

check: fmt tidy vet lint test ref   ## the gate

fmt:   ## fails, listing them, when gofmt would change files
	@out="$$($(GOFMT) -l $$(git ls-files -co --exclude-standard '*.go'))"; \
	if [ -n "$$out" ]; then echo "gofmt would change:"; echo "$$out"; exit 1; fi

tidy:   ## fails when go.mod or go.sum is not what go mod tidy makes
	$(GO) mod tidy -diff

vet:
	$(GO) vet ./...

lint:   ## staticcheck, pinned
	$(GO) run $(STATICCHECK) ./...

test:   ## the tests (A2UI's conformance suites among them), with the race detector
	$(GO) test -race ./...

ref:   ## the references (ref/, a module of its own): tidy, vetted, built
	cd ref && $(GO) mod tidy -diff && $(GO) vet ./... && $(GO) build -o ../bin/ref .

# A reference shot: Bubble Tea's component (or OpenTUI's, ref/opentui)
# beside the kit's story, as one picture in .shots/NAME.png
# (scripts/ref-shot.sh; vhs and ImageMagick).
# Not in the gate.
shot:   ## .shots/NAME.png: ref NAME beside its story (NAME=form; KIT_KEYS, REF_KEYS, COLS, ROWS)
	sh scripts/ref-shot.sh $(NAME) $(STORY)

a2ui:   ## third_party/a2ui at REV, a full A2UI commit (scripts/a2ui.sh)
	sh scripts/a2ui.sh $(REV)

lexers:   ## highlight/lexers from the chroma go.mod requires (scripts/lexers.sh)
	sh scripts/lexers.sh

icons:   ## icons/basic.go, storybook/icons_gen.go and rendition/html/icons_gen.go from the Iconify package icons/REV pins (cmd/iconsgen); VERSION=… pins another
	$(GO) run ./cmd/iconsgen -set storybook -set rendition/html $(if $(VERSION),-version $(VERSION))

# The README's GIF: the storybook in xterm.js, driven with Playwright from
# a checkout of neuroplastio/xterm-addon-hotty (ADDON, built: npm run build),
# and made a GIF with ffmpeg (scripts/storybook-gif.mjs). Not in the gate.
ADDON ?= ../../xterm-addon-hotty/main

gif:   ## docs/storybook.gif, recorded again: needs ADDON and ffmpeg
	$(GO) build -o .gif/storybook ./cmd/storybook
	cd $(ADDON) && mise x -- node $(CURDIR)/scripts/storybook-gif.mjs $(CURDIR)/.gif/storybook $(CURDIR)/docs/storybook.gif

clean:
	rm -rf coverage.out bin .shots
