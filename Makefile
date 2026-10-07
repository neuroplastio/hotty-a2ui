# hotty-a2ui. Toolchain from mise.toml (Go): `mise install` first.
GO          ?= mise x -- go
GOFMT       ?= mise x -- gofmt
# staticcheck 2026.2.1, the first that knows Go 1.26.
STATICCHECK ?= honnef.co/go/tools/cmd/staticcheck@v0.8.1

.PHONY: check fmt tidy vet lint test a2ui clean

check: fmt tidy vet lint test   ## the gate

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

a2ui:   ## third_party/a2ui at REV, a full A2UI commit (scripts/a2ui.sh)
	sh scripts/a2ui.sh $(REV)

clean:
	rm -f coverage.out
