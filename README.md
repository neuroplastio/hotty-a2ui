# hotty-a2ui

HOTTY's UI kit on [A2UI](https://github.com/google/A2UI): a program, or an
agent, describes its UI once as A2UI v1.0, and gets it in the three
renditions of HOTTY's SDK.md §2.5. Those are surfaces on a HOTTY host, cells
on a terminal that is not one, and plain text with no terminal.

This repository is the proof of concept that gov's NEIO-11 calls for. It is
private, and nothing in it is published.

## What is here

| path | what |
| --- | --- |
| [`docs/profile.md`](docs/profile.md) | the HOTTY profile of A2UI, a draft: renditions, keys and focus, the hotty catalog, fallbacks |
| [`catalog/hotty/catalog.json`](catalog/hotty/catalog.json) | the hotty catalog: `Shortcut`, `Form`, `focus`, `blur` |
| [`a2ui/`](a2ui) | an A2UI v1.0 core in Go: JSON Pointer, the data model, expressions, and more to come |
| [`third_party/a2ui/`](third_party/a2ui) | A2UI at one commit ([`REV`](third_party/a2ui/REV)): the v1.0 schemas, the basic catalog and its examples, the conformance suites |

## The gate

```sh
mise install
make check   # gofmt, go mod tidy, vet, staticcheck, and the tests with -race
```

The tests run A2UI's conformance suites from `third_party/a2ui`.

## A2UI's pin

A2UI is pinned at one commit and refreshed by hand:

```sh
make a2ui REV=<full A2UI commit>
```

The script copies only what the kit and its tests read. A2UI is Apache-2.0
(`third_party/a2ui/LICENSE`).
