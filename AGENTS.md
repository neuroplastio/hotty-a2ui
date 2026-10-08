# For agents

This repository is the proof of concept of HOTTY's UI kit on A2UI (gov
NEIO-11): an A2UI v1.0 core in Go, a HOTTY renderer and a storybook.

- **NEIO-11 is the brief.** Its exit criteria say when this is done. Its
  open questions are the maintainer's; don't settle them in code.
- **Only A2UI's extension points.** Custom catalogs mixed component by
  component, `Action` props, renderer functions with `allowedCallers`, and
  extensions under `io_neuroplast_hotty`. Never fork a basic component.
- **Nothing goes upstream to A2UI** (no issue, PR, proposal or discussion)
  until the maintainer says so.
- **HOTTY's SPEC.md doesn't change for this.** If the kit needs the wire to
  change, that lands in `hotty` first, as any wire change does.
- **hotty-go stays protocol-only.** This repository depends on it, never the
  other way round.
- **A2UI is pinned** in `third_party/a2ui` (`make a2ui REV=…`). Don't edit
  the files there; the conformance suites run from them as they are.
- **The gate is `make check`.** Push to main once it passes, fast-forward
  only. The repository is public (since 2026-10-08, for hotty-demo's
  storybook app); a release, a tag or a package is the maintainer's act.
