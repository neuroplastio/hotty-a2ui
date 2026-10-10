// story: hotty/qr
//
// The reference for HottyQRCode (vault KIT-20): OpenTUI's QR code
// (@opentui/qrcode's QRCodeRenderable), with the content of the kit's story
// hotty/qr, so that a shot of one can be set beside a shot of the other
// (scripts/ref-shot.sh): a URL's code beside its text, then a Wi-Fi
// network's at error correction H beside its own. QRCodeRenderable's
// defaults are the kit's choices too: half blocks, a module a column wide
// and half a row tall, a quiet zone of four modules, black on white
// whatever the terminal's colours, level M. It has no label, so the labels
// are text under the codes; nothing is bound, so the URL's field is text.
//
//	bun ref/opentui/qr.ts      Control+C quits
import { BoxRenderable, TextRenderable, createCliRenderer } from "@opentui/core"
import { ErrorCorrectionLevel, QRCodeRenderable } from "@opentui/qrcode"

const renderer = await createCliRenderer({ exitOnCtrlC: true })
const page = new BoxRenderable(renderer, { flexDirection: "column", width: "100%", height: "100%" })
const text = (content: string) => new TextRenderable(renderer, { content })

// A code with its label centred under it, and its text beside them.
const row = (content: string, errorCorrectionLevel: ErrorCorrectionLevel, label: string, beside: string[]) => {
  const r = new BoxRenderable(renderer, { flexDirection: "row", columnGap: 1 })
  const code = new BoxRenderable(renderer, { flexDirection: "column", alignItems: "center", flexShrink: 0 })
  code.add(new QRCodeRenderable(renderer, { content, errorCorrectionLevel }))
  code.add(text(label))
  r.add(code)
  const side = new BoxRenderable(renderer, { flexDirection: "column", flexGrow: 1, flexShrink: 1 })
  for (const line of beside) side.add(text(line))
  r.add(side)
  return r
}

page.add(text("QR code"))
page.add(row("https://hotty.neuroplast.io", ErrorCorrectionLevel.M, "Scan to open", [
  "Bound to the data model: edit the URL and",
  "the code follows. It is black on white",
  "whatever the terminal's colours, with its",
  "quiet zone, as a phone's camera reads it.",
  "  URL  https://hotty.neuroplast.io",
]))
page.add(row("WIFI:T:WPA;S:hotty;P:terminal;;", ErrorCorrectionLevel.H, "Join hotty's Wi-Fi", [
  "Error correction H: a third of the",
  "code may be lost and it still reads,",
  "for a larger code. A Wi-Fi network's",
  "name and password, which a phone",
  "offers to join.",
]))
renderer.root.add(page)
