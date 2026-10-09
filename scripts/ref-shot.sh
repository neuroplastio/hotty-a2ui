#!/bin/sh
# ref-shot.sh NAME [STORY] — a reference shot (vault KIT-REF): Bubble Tea's
# component (ref NAME) beside the kit's story (storybook -bare STORY), both
# in the same terminal, at the same size, with the same font and theme, as
# one picture in .shots/NAME.png. STORY defaults to the one ref -list names.
#
# Where Bubble Tea has no such component, the reference is OpenTUI's:
# ref/opentui/NAME.ts, run with Bun, whose "// story:" line names its
# story (bun install in ref/opentui first).
#
# Keys to send first, as VHS commands, so that both sides are in the same
# state (a field focused, a row moved to):
#   REF_KEYS='Down Down'    for Bubble Tea's side (empty by default)
#   KIT_KEYS='Tab Down'     for the kit's (Tab by default: it gives the
#                           surface the keyboard, as huh and bubbles start)
# COLS and ROWS set the terminal (80 by 24).
#
# Needs vhs (with ttyd and a Chromium) and ImageMagick.
set -eu
cd "$(dirname "$0")/.."

name=${1:?usage: ref-shot.sh NAME [STORY]}
mkdir -p bin .shots
mise exec -- go build -o bin/storybook ./cmd/storybook
if [ -f "ref/opentui/$name.ts" ]; then
	refcmd="mise exec bun@1.4.2 -- bun ref/opentui/$name.ts"
	reflabel="OpenTUI · ref $name"
	story=${2:-$(sed -n 's|^// story: *||p' "ref/opentui/$name.ts" | head -1)}
else
	(cd ref && mise exec -- go build -o ../bin/ref .)
	refcmd="bin/ref $name"
	reflabel="Bubble Tea · ref $name"
	story=${2:-$(bin/ref -list | awk -v n="$name" '$1 == n { print $2 }')}
fi
[ -n "$story" ] || { echo "ref-shot: no ref $name (bin/ref -list, ref/opentui)" >&2; exit 2; }

cols=${COLS:-80}
rows=${ROWS:-24}
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

# tape SIDE COMMAND KEYS: shoot one side into $tmp/SIDE.png.
tape() {
	cat >"$tmp/$1.tape" <<EOF
Output "$tmp/$1.gif"
Set Shell bash
Set FontSize 16
Set Width $((cols * 11 + 60))
Set Height $((rows * 22 + 60))
Set Padding 20
Set TypingSpeed 0
Set Theme "Catppuccin Mocha"
Hide
Type "unset PLX_PANE PLX_SOCKET PLX_WORKSPACE; clear; stty cols $cols rows $rows; $2"
Enter
Sleep 2s
Show
$(for k in $3; do printf '%s\nSleep 200ms\n' "$k"; done)
Sleep 1s
Screenshot "$tmp/$1.png"
Sleep 500ms
EOF
	vhs "$tmp/$1.tape" >/dev/null
}

tape ref "$refcmd" "${REF_KEYS:-}"
tape kit "bin/storybook -bare $story" "${KIT_KEYS-Tab}"

montage -background '#1e1e2e' -fill '#cdd6f4' -font DejaVu-Sans -pointsize 18 \
	-label "$reflabel" "$tmp/ref.png" \
	-label "the kit · $story" "$tmp/kit.png" \
	-tile 2x1 -geometry +12+12 ".shots/$name.png"
echo ".shots/$name.png"
