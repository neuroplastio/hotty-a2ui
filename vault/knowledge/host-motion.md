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
