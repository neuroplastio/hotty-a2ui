# hotty-a2ui

HOTTY's UI kit on [A2UI](https://github.com/google/A2UI): a program, or an
agent, describes its UI once as A2UI v1.0, and gets it in the three
renditions of HOTTY's SDK.md §2.5. Those are surfaces on a HOTTY host, cells
on a terminal that is not one, and plain text with no terminal.

This repository is the proof of concept that gov's NEIO-11 calls for: a
draft, with no release, whose API changes freely. hotty-demo shows its
storybook as an app, on https://hotty.neuroplast.io/storybook.

## What is here

| path | what |
| --- | --- |
| [`docs/profile.md`](docs/profile.md) | the HOTTY profile of A2UI, a draft: renditions, keys and focus, the hotty catalog, fallbacks |
| [`catalog/hotty/`](catalog/hotty) | the hotty catalog: `Shortcut`, `Form`, `focus`, `blur` |
| [`catalog/basic/`](catalog/basic) | A2UI's basic catalog, its functions implemented (en-US formatting) |
| [`a2ui/`](a2ui) | an A2UI v1.0 core in Go: the data model, expressions and functions, catalogs checked against the spec's JSON Schemas, node resolution, and the message processor with the agent's function calls |
| [`conformance/`](conformance) | A2UI's conformance suites, run against the core |
| [`view/`](view) | the renderer's model of a surface: its nodes made into a few kinds of element, with the renderer's own state (focus, the tab shown, the modal open), and the controller the user's acts go through |
| [`rendition/html/`](rendition/html) | the rendition on a HOTTY host: a document, then deltas, and the host's events back as the user's acts |
| [`rendition/cells/`](rendition/cells) | the rendition in a terminal that is not a host: laid out and painted in cells by the profile's rules (§3), with keys and clicks as SPEC §10 has them |
| [`rendition/text/`](rendition/text) | the rendition with no terminal: plain text for a pipe |
| [`rendition/theme/`](rendition/theme) | the themes both renditions paint with: colours by role, and on a host the shapes |
| [`vectors/`](vectors) | keys and focus, one set of vectors run against the cells rendition and against the HTML one on a host |
| [`story/`](story) | the stories (A2UI's basic examples, the hotty catalog's, the fallbacks) and a story as it runs |
| [`storybook/`](storybook) | the storybook: every story in the rendition picked, live, with its actions, data model and messages; or what an agent streams. A program shows it in a part of its screen, over its own HOTTY session (hotty-demo's storybook app) |
| [`cmd/storybook/`](cmd/storybook) | the storybook on the whole screen, as plain text in a pipe, or a story's HTML |
| [`third_party/a2ui/`](third_party/a2ui) | A2UI at one commit ([`REV`](third_party/a2ui/REV)): the v1.0 schemas, the basic catalog and its examples, the conformance suites |

## The storybook

```sh
go run ./cmd/storybook                       # the stories, in this terminal
go run ./cmd/storybook -text 36_modal        # one story as text
agent | go run ./cmd/storybook -stream - -out actions.jsonl
                                             # what an agent streams; the user's actions to a file
```

On a HOTTY host (hottyterm, xterm-addon-hotty) a story shows as surfaces,
in cells, as text, or as surfaces beside cells; elsewhere in cells or as
text. Side by side, both take input: the keyboard is in one of them at a
time, and the other shows the same state. Tab moves through the story,
then the panel and the list; F2 changes the rendition, F3 the theme, and
Ctrl+C quits. The storybook's own list and panel are A2UI surfaces as
well.

```sh
go run ./cmd/storybook -theme "material dark"     # start in a theme
go run ./cmd/storybook -html 06_music-player > p.html
hotty render p.html -o p.png --cols 56            # hotty-blitz: see a surface
```

## The gate

```sh
mise install
make check   # gofmt, go mod tidy, vet, staticcheck, and the tests with -race
```

The tests run A2UI's conformance suites from `third_party/a2ui`: every v1.0
case of the core's suites, a few of them translated from v0.9's vocabulary.
The cases not run, and the suites, each say why (`conformance/`, the
`*Skips` maps and `outOfScope`); a new pin that adds a suite fails the gate
until a test runs it or says why not.

## A2UI's pin

A2UI is pinned at one commit and refreshed by hand:

```sh
make a2ui REV=<full A2UI commit>
```

The script copies only what the kit and its tests read. A2UI is Apache-2.0
(`third_party/a2ui/LICENSE`).

## License

Apache-2.0 ([`LICENSE`](LICENSE)).
