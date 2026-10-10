#!/bin/sh
# pair-shot.sh OUT LEFT_LABEL LEFT_CMD LEFT_KEYS RIGHT_LABEL RIGHT_CMD RIGHT_KEYS
# Two commands side by side in VHS, as ref-shot.sh shoots them (the same
# terminal, size, font and theme), as one picture OUT with a caption under
# each side: a kit before and after a change, say, two storybook binaries
# built at two commits. The keys are VHS commands to send first, separated
# by ";" (Tab;Down, or Type "ger"). COLS and ROWS set the terminal (80 by
# 24). scripts/blind-shots.sh blinds a directory of them for cmd/review.
set -eu
out=$1
cols=${COLS:-80}
rows=${ROWS:-24}
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

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
$(IFS=";"; for k in $3; do printf "%s\nSleep 200ms\n" "$k"; done)
Sleep 1s
Screenshot "$tmp/$1.png"
Sleep 500ms
EOF
	vhs "$tmp/$1.tape" >/dev/null
}

tape left "$3" "$4"
tape right "$6" "$7"
montage -background '#1e1e2e' -fill '#cdd6f4' -font DejaVu-Sans -pointsize 18 \
	-label "$2" "$tmp/left.png" -label "$5" "$tmp/right.png" \
	-tile 2x1 -geometry +12+12 "$out"
echo "$out"
