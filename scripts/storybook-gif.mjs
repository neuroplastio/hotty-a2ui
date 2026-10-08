// The README's GIF: the storybook in xterm.js with the HOTTY addon
// (neuroplastio/xterm-addon-hotty, its serve.py running the storybook on a
// pty), driven with Playwright, recorded as video and made a GIF with
// ffmpeg. `make gif` builds the storybook and runs this from the addon's
// checkout, whose Playwright it uses.
//
//   node storybook-gif.mjs <storybook binary> <out.gif>
//
// What it shows: a few of A2UI's examples as surfaces; a login form typed
// into, its action back in the panel; the same surface beside its cells;
// the themes; and the storybook all in cells.
import { spawn, spawnSync } from "node:child_process";
import { mkdtempSync, readdirSync, rmSync, statSync } from "node:fs";
import { createRequire } from "node:module";
import { tmpdir } from "node:os";
import { join, resolve } from "node:path";

const [bin, out = "storybook.gif"] = process.argv.slice(2).map((p) => resolve(p));
if (!bin) {
  console.error("usage: node storybook-gif.mjs <storybook binary> <out.gif>");
  process.exit(2);
}
const addon = process.cwd();
const { chromium } = createRequire(join(addon, "package.json"))("@playwright/test");
const fps = process.env.GIF_FPS ?? "12";
const width = process.env.GIF_WIDTH ?? "960";
const port = 8890 + Math.floor(Math.random() * 100);
const videoDir = mkdtempSync(join(tmpdir(), "storybook-gif-"));
const size = { width: 1280, height: 760 };

const server = spawn("python3", ["serve.py", "--port", String(port), "--", bin, "basic/01_flight-status"], {
  cwd: addon,
  stdio: "ignore",
});
await new Promise((r) => setTimeout(r, 800));

const browser = await chromium.launch();
const context = await browser.newContext({ viewport: size, recordVideo: { dir: videoDir, size } });
let start = 0;
let crop = "";
try {
  const page = await context.newPage();
  // A cursor, since a video has no pointer: it shows where the user is
  // about to act, and fades when still.
  await page.addInitScript(() => {
    const put = () => {
      const c = document.createElement("div");
      c.id = "__cursor";
      c.style.cssText =
        "position:fixed;z-index:2147483647;width:16px;height:16px;margin:-8px 0 0 -8px;opacity:0;" +
        "border-radius:50%;background:rgba(157,144,255,.95);box-shadow:0 0 0 3px rgba(157,144,255,.3)," +
        "0 0 14px rgba(157,144,255,.7);pointer-events:none;" +
        "transition:left .25s ease,top .25s ease,opacity .4s ease";
      document.body.appendChild(c);
    };
    // The terminal fills the page, which the video is cropped to.
    const fill = () => {
      const st = document.createElement("style");
      st.textContent = "#term{inset:0!important}";
      document.head.appendChild(st);
    };
    if (document.readyState === "loading") addEventListener("DOMContentLoaded", () => (put(), fill()));
    else put(), fill();
  });
  const t0 = Date.now();
  await page.goto(`http://127.0.0.1:${port}/?size=15`);
  const beat = (ms) => page.waitForTimeout(ms);
  const frame = (name) => page.frameLocator(`.hotty-surface[data-surface="${name}"] iframe`);
  const nav = frame("nav");
  // The story's surface: its name changes with each story opened.
  const story = async () => {
    const s = page.locator('.hotty-surface[data-surface^="s"][data-surface$="-h"]').last();
    await s.waitFor({ state: "visible" });
    return frame(await s.getAttribute("data-surface"));
  };
  async function point(locator) {
    await locator.scrollIntoViewIfNeeded();
    const box = await locator.boundingBox();
    if (!box) return;
    const [x, y] = [box.x + Math.min(box.width / 2, 60), box.y + box.height / 2];
    await page.evaluate(([x, y]) => {
      const c = document.getElementById("__cursor");
      c.style.left = x + "px";
      c.style.top = y + "px";
      c.style.opacity = "1";
      clearTimeout(window.__idle);
      window.__idle = setTimeout(() => (c.style.opacity = "0"), 1600);
    }, [x, y]);
    await page.mouse.move(x, y, { steps: 6 });
    await beat(320);
  }
  async function click(locator) {
    await point(locator);
    await locator.click();
  }
  const open = (title) => click(nav.getByText(title, { exact: true }));
  async function pick(select, option) {
    await click(nav.locator("#" + select));
    await beat(500);
    await click(nav.locator(`[id="${select}~o${option}"]`));
  }

  await nav.getByText("Flight Status", { exact: true }).waitFor();
  await (await story()).locator("body").waitFor();
  start = (Date.now() - t0) / 1000 + 0.6;
  const box = await page.locator(".xterm-screen").boundingBox();
  const even = (n) => Math.floor(n / 2) * 2;
  crop = `crop=${even(box.width)}:${even(box.height)}:${Math.round(box.x)}:${Math.round(box.y)},`;
  await beat(1800);

  await open("Weather Current");
  await beat(1600);
  await open("Coffee Order");
  await beat(1600);

  // A form: typed into, sent, and its action in the panel, as the agent
  // gets it.
  await open("Login Form with Validation");
  await beat(700);
  let s = await story();
  await click(s.locator("#email_field"));
  await page.keyboard.type("ada@analytical.engine", { delay: 45 });
  await click(s.locator("#password_field"));
  await page.keyboard.type("bernoulli1843", { delay: 45 });
  await beat(300);
  await click(s.locator("#login_btn"));
  await beat(1800);

  // The same surface beside its cells, which show what was typed.
  await pick("rend", 3);
  await beat(2200);

  // Themes.
  await pick("theme_p", 4);
  await beat(1500);
  await pick("theme_p", 2);
  await beat(1500);
  await open("Calendar Day");
  await beat(1500);
  await pick("theme_p", 3);
  await beat(1500);

  // All in cells: the storybook's own list and panel stay surfaces.
  await pick("rend", 1);
  await beat(2200);
  await open("Flight Status");
  await beat(500);
  await pick("theme_p", 0);
  await pick("rend", 0);
  await beat(1200);
} finally {
  await context.close();
  await browser.close();
  server.kill();
}

const webm = readdirSync(videoDir).filter((f) => f.endsWith(".webm")).map((f) => join(videoDir, f)).pop();
if (!webm) {
  console.error("storybook-gif: no video was recorded");
  process.exit(1);
}
if (process.env.GIF_KEEP) spawnSync("cp", [webm, out.replace(/\.gif$/, "") + ".webm"]);
const filter = [
  `[0:v]${crop}fps=${fps},scale=${width}:-1:flags=lanczos,split[s0][s1]`,
  "[s0]palettegen=max_colors=128:stats_mode=diff[p]",
  "[s1][p]paletteuse=dither=bayer:bayer_scale=5:diff_mode=rectangle[out]",
].join(";");
const ff = spawnSync(
  "ffmpeg",
  ["-y", "-loglevel", "error", "-ss", String(start), "-i", webm, "-filter_complex", filter, "-map", "[out]", "-loop", "0", out],
  { stdio: "inherit" },
);
rmSync(videoDir, { recursive: true, force: true });
if (ff.status !== 0) process.exit(ff.status ?? 1);
console.log(`${out}: ${statSync(out).size} bytes`);
