# The maintainer on the site's local dev server (2026-10-11)

Reported while testing hotty.neuroplast.io's storybook on the local dev
server (web's `p3` checkout, the kit at 4b74875). Each item is deleted
in the leg that addresses it.

- **Slider drags.** "Cell-only sliders lost their draggability (hotty
  version works)." A drag through the Book in Go moves the slider, so
  the cause is likely the page's mouse reports under 0.2. Web can't
  reproduce it with a mouse (Chromium, Firefox, local and live); a touch
  drag over cells scrolls instead. Asked the maintainer whether it was
  touch.
