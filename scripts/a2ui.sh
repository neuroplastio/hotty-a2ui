#!/bin/sh
# Copies A2UI, at one commit, into third_party/a2ui: the v1.0 schemas and
# docs, the basic catalog with its examples, and the conformance suites with
# the fixtures they load. Run it from the repository root (make a2ui REV=…).
set -eu

rev=${1:?usage: scripts/a2ui.sh <A2UI commit>}
repo=${A2UI_REPO:-https://github.com/google/A2UI}
dest=third_party/a2ui

paths="
LICENSE
specification/v1_0/json
specification/v1_0/docs
specification/v0_9_1/catalogs/basic/catalog.json
catalogs/basic/v1
conformance/README.md
conformance/conformance_schema.json
conformance/core
conformance/test_data/node
conformance/test_data/catalogs
"

tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT
git -C "$tmp" init -q
git -C "$tmp" fetch -q --depth 1 "$repo" "$rev"
git -C "$tmp" checkout -q FETCH_HEAD

rm -rf "$dest"
for p in $paths; do
	mkdir -p "$dest/$(dirname "$p")"
	cp -R "$tmp/$p" "$dest/$p"
done
git -C "$tmp" log -1 --format='%H %cs %s' > "$dest/REV"
echo "third_party/a2ui at $(cut -c1-12 "$dest/REV")"
