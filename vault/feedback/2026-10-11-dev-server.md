# The maintainer on the site's local dev server (2026-10-11)

Reported while testing hotty.neuroplast.io's storybook on the local dev
server (web's `p3` checkout, the kit at 4b74875). Each item is deleted
in the leg that addresses it.

- **I-beam.** "I beam cursor only shown when hovering over the blinking
  cursor block, but it should be when hovering over any point of the
  text input field." The kit's Pointer gives "text" over the whole
  value; the site's app panes drop OSC 22 (web vterm/pane.go), asked of
  web.
- **Slider drags.** "Cell-only sliders lost their draggability (hotty
  version works)." A drag through the Book in Go moves the slider, so
  the cause is likely the page's mouse reports under 0.2.
- **Gradient.** "Gradient is not shown in cell only progress bar (web
  browser)". x/vt answers OSC 10/11/12 but not OSC 4, so the fill has
  no accent and magenta to blend; asked of web (the pane answering OSC 4
  from the page's palette).
