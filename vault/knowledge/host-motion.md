# Motion on a HOTTY host: CSS animations in hotty-blitz

As of 2026-10-09: hotty-blitz main and its Blitz fork (patches 0001–0023),
in hottyterm 1.3.2-hottyterm+78a44228. It was checked live in a headless
hottyterm, with a probe surface holding one `@keyframes` animation and one
`transition` on `left`.

**Verdict.** The kit can't hand motion to the host yet. A Spinner or an
indeterminate Progress moves by the program's deltas on the shared clock
(profile §3.4).

What the probe showed:
- **Nothing schedules frames for CSS motion.** Stylo computes animations
  and transitions, and Blitz reports them (`BaseDocument::is_animating`).
  But hotty-blitz's `Surface::next_frame` asks for frames only for animated
  images, fading scrollbars and the caret. A surface redraws when a delta
  or something else changes it, not because CSS is moving.
- **Transitions:** a `transition: left 3s` started by a class change stays
  at its start. With an unrelated delta every 400 ms, it advances once per
  delta, to where the clock says it should be.
- **`@keyframes`:** the first keyframe applies (its colour and place
  replace the element's own), but the animation never advances, even with
  forced redraws. It's stuck at its start time. Why isn't known.

What would change it, in hotty-blitz, not in the kit:
- `next_frame` counts `doc.is_animating()`, so a host renders while CSS
  moves;
- `@keyframes` advance, whatever holds them at their first frame;
- whether SPEC says anything about motion a host runs by itself (for
  instance `prefers-reduced-motion`) is the maintainer's call.

When those land, the kit could drop the deltas for an indeterminate bar
(`left` animated in CSS) and keep the clock only for cells. That leaves
cells and the host out of step, which the shared clock avoids today.

## A hidden page replays its backlog (2026-10-09)

On hotty.neuroplast.io, a page hidden for a while (a phone locked, a
laptop asleep, a background tab) fast-forwards its animations when it
shows again: the progress bar in the storybook, and the docs' screencast
(2026-10-11).

**Cause.** Chrome runs a hidden page's timers about once a minute, but
messages from the app's Worker still arrive. The app keeps sending
frames (the progress bar's ten a second), and xterm.js, which parses
what it is written on those throttled timers, falls behind: the deltas
queue, and play all at once on wake. Reproduced in Chromium by holding
the page's timers for 30 s: 12 changes reached the surface meanwhile,
then 897 in the first 250 ms after.

**Not the kit's to fix.** The kit already works out each frame from the
clock (view.UntilStep; a frame is a function of the time, not of the
frames before it), so one frame drawn now is the right one. What piles
up is the sending.

**Proposed fix (not landed; it changes the site's code, the
maintainer's call).** When the page is hidden (`visibilitychange`,
`document.hidden`), the site treats the open app as off screen, as it
already does when the user switches apps (web 743cf8b: off screen,
an app sends nothing). The app stops drawing; when the page shows again
it draws the current frame once. Nothing piles up, and a hidden page
spends no battery on animation. In web/hotty: a listener in
`web/src/main.ts`, and the demo taking focus from the app's pane. Check
it with the same timer-holding test. The rule for anything that
animates on the site, the docs' screencast too: frames are picked by
the clock, and while hidden none are sent; on showing, the current one,
never the backlog.
