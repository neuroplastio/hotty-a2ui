#!/bin/sh
# blind-shots.sh [DIR] — blind copies of the side-by-side shots in DIR
# (.shots by default), for a review that mustn't know which side is which:
# each picture's two halves in a random order, their captions painted
# over, "A" and "B" under them, in DIR/blind (cmd/review shows them).
#
# DIR/blind/.key says which was which, a line a picture:
#   form.png A=left B=right
# where left is the original's left half: the reference in a ref-shot.sh
# picture, "before" in a before-and-after one.
#
# The captions are montage's, as ref-shot.sh writes them: the bottom 38
# rows of the picture, under the screenshots. Needs ImageMagick 7.
set -eu
cd "$(dirname "$0")/.."

dir=${1:-.shots}
out=$dir/blind
mkdir -p "$out"
bg='#1e1e2e'
key=$out/.key
: >"$key"
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT

for f in "$dir"/*.png; do
	name=$(basename "$f")
	w=$(magick identify -format %w "$f")
	h=$(magick identify -format %h "$f")
	half=$((w / 2))
	for side in left right; do
		x=0
		[ "$side" = right ] && x=$half
		magick "$f" -crop "${half}x${h}+${x}+0" +repage \
			-fill "$bg" -draw "rectangle 0,$((h - 38)) $half,$h" "$tmp/$side.png"
	done
	if [ "$(od -An -N1 -tu1 /dev/urandom | tr -d ' ')" -lt 128 ]; then
		a=left b=right
	else
		a=right b=left
	fi
	label() {
		magick "$tmp/$1.png" -gravity south -font DejaVu-Sans -pointsize 22 \
			-fill '#cdd6f4' -annotate +0+8 "$2" "$tmp/$2.png"
	}
	label "$a" A
	label "$b" B
	magick "$tmp/A.png" "$tmp/B.png" +append "$out/$name"
	echo "$name A=$a B=$b" >>"$key"
done
echo "$out ($(wc -l <"$key") pictures; the key is $key)"
