#!/bin/sh
# ref-shot.sh NAME [STORY] — a reference shot (vault KIT-REF): Bubble Tea's
# component (ref NAME) beside the kit's story (storybook -bare STORY), both
# in the same terminal, at the same size, with the same font and theme, as
# one picture in .shots/NAME.png. STORY defaults to the one ref -list names.
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
(cd ref && mise exec -- go build -o ../bin/ref .)
mise exec -- go build -o bin/storybook ./cmd/storybook
story=${2:-$(bin/ref -list | awk -v n="$name" '$1 == n { print $2 }')}
[ -n "$story" ] || { echo "ref-shot: no ref $name (bin/ref -list)" >&2; exit 2; }

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

tape ref "bin/ref $name" "${REF_KEYS:-}"
tape kit "bin/storybook -bare $story" "${KIT_KEYS-Tab}"

montage -background '#1e1e2e' -fill '#cdd6f4' -font DejaVu-Sans -pointsize 18 \
	-label "Bubble Tea · ref $name" "$tmp/ref.png" \
	-label "the kit · $story" "$tmp/kit.png" \
	-tile 2x1 -geometry +12+12 ".shots/$name.png"
echo ".shots/$name.png"
