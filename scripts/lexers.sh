#!/bin/sh
# lexers.sh — highlight/lexers: chroma's lexer definitions for the
# languages the kit highlights with no import (vault KIT-05), copied from
# the chroma that go.mod requires, with chroma's licence. The rest of
# chroma's lexers are a program's to add: import highlight/all.
#
# They are chroma's XML, which highlight embeds as it is, and Go's,
# which chroma has as Go code, written as XML (scripts/golexer.go).
set -eu
cd "$(dirname "$0")/.."

dir=$(mise x -- go list -m -f '{{.Dir}}' github.com/alecthomas/chroma/v2)
ver=$(mise x -- go list -m -f '{{.Version}}' github.com/alecthomas/chroma/v2)
[ -d "$dir/lexers/embedded" ] || { echo "lexers: no chroma at $dir" >&2; exit 2; }

# What coding agents write most, and the languages those embed (HTML's
# CSS and JavaScript, a Makefile's and a Dockerfile's Bash).
langs='bash c c++ c# css dart diff docker elixir graphql haskell html ini java
javascript json kotlin lua makefile nix php powershell protocol_buffer python
ruby rust scala scss sql swift terraform toml typescript xml yaml zig'

out=highlight/lexers
rm -rf "$out"
mkdir -p "$out"
for l in $langs; do
	cp "$dir/lexers/embedded/$l.xml" "$out/$l.xml"
done
cp "$dir/COPYING" "$out/COPYING"
chmod -R u+w "$out"

# Go's lexer is Go code in chroma's lexers package: scripts/golexer.go
# writes it as XML.
mise x -- go run scripts/golexer.go >"$out/go.xml"
echo "highlight/lexers: chroma $ver, $(ls "$out" | grep -c xml) languages, Go among them"
