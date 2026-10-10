// story: hotty/toast
//
// The reference for hottyToast (vault KIT-13): toasts as OpenTUI's
// programs draw them. OpenTUI has no toast of its own (its renderer's
// triggerNotification is the desktop's, by OSC); opencode, built on it,
// draws one as a box of OpenTUI's, absolute in the top right corner, over
// the panel's colour, between bars in the kind's colour (┃), padded a row
// and two columns. opencode shows one at a time; here the story's four
// stack, so that the shot sets like beside like (scripts/ref-shot.sh).
// Nothing reads keys: the kind is the bars' colour alone, and there is no
// action.
//
//	bun ref/opentui/toast.ts      Control+C quits
import { BoxRenderable, TextAttributes, TextRenderable, createCliRenderer } from "@opentui/core"

// The kinds' colours, in Catppuccin Mocha, the shot's theme, as the kit's
// roles are (profile §3.4): info sky, success green, warning yellow, error
// red; the panel a tone above the base.
const kinds = {
  info: "#89dceb",
  success: "#a6e3a1",
  warning: "#f9e2af",
  error: "#f38ba8",
}
const panel = "#313244"

const renderer = await createCliRenderer({ exitOnCtrlC: true })

// What is under the toasts: the story's first lines.
const page = new BoxRenderable(renderer, { flexDirection: "column", width: "100%", height: "100%" })
for (const line of [
  "Toasts and tooltips",
  "The buttons show toasts. A click on a toast, or Escape, dismisses it.",
  "[ Info ] [ Success ] [ Warning ] [ Error ]",
  "[ Delete ] [ Upload ] [ Finish ] [ Pin one ] [ Unpin ]",
]) {
  page.add(new TextRenderable(renderer, { content: line }))
}
renderer.root.add(page)

// The toasts, the newest at the top, as the agent showed them.
const stack = new BoxRenderable(renderer, { position: "absolute", top: 1, right: 2, flexDirection: "column", rowGap: 1, zIndex: 10 })
const bars = { topLeft: " ", topRight: " ", bottomLeft: " ", bottomRight: " ", horizontal: " ", vertical: "┃", topT: " ", bottomT: " ", leftT: " ", rightT: " ", cross: " " }
for (const [kind, message] of [
  ["info", "Indexing 1,204 files"],
  ["success", "Draft saved"],
  ["warning", "Disk almost full: 2 GB left"],
  ["error", "Could not reach the server"],
] as const) {
  const toast = new BoxRenderable(renderer, {
    maxWidth: Math.min(60, renderer.width - 6),
    paddingLeft: 2,
    paddingRight: 2,
    paddingTop: 1,
    paddingBottom: 1,
    backgroundColor: panel,
    border: ["left", "right"],
    customBorderChars: bars,
    borderColor: kinds[kind],
  })
  if (kind === "error") {
    toast.add(new TextRenderable(renderer, { content: "Offline", attributes: TextAttributes.BOLD, marginBottom: 1 }))
  }
  toast.add(new TextRenderable(renderer, { content: message, wrapMode: "word" }))
  stack.add(toast)
}
renderer.root.add(stack)
