// Records the demo GIF source video for the README (docs/assets/demo.gif).
//
// The tour runs against the hub in demo mode, so it only ever shows synthetic
// data. Policy results turn stale 90 seconds after the hub starts, so start the
// hub immediately before recording.
//
//   make build
//   KFLEET_DEMO_MODE=true KFLEET_LISTEN_ADDR=127.0.0.1:18080 ./bin/hub &
//   npm install --no-save playwright-core && npx playwright-core install ffmpeg
//   OUT_DIR=/tmp/kfleet-demo node hack/record-demo.mjs
//
// It prints the path of a .webm. Convert it with ffmpeg:
//
//   ffmpeg -ss 1 -i <video.webm> -vf "fps=8,scale=900:-1:flags=lanczos,split[a][b];\
//   [a]palettegen=max_colors=128:stats_mode=diff[p];\
//   [b][p]paletteuse=dither=bayer:bayer_scale=4:diff_mode=rectangle" docs/assets/demo.gif
//
// Requires Google Chrome, Node.js 22, and ffmpeg.
import { chromium } from 'playwright-core';
import { mkdir, writeFile } from 'node:fs/promises';

const base = process.env.KFLEET_URL ?? 'http://127.0.0.1:18080';
const outDir = process.env.OUT_DIR ?? '/tmp/kfleet-demo';
const size = { width: 1280, height: 800 };

await mkdir(outDir, { recursive: true });

const browser = await chromium.launch({ channel: 'chrome', headless: true });
const context = await browser.newContext({
  viewport: size,
  colorScheme: 'dark',
  recordVideo: { dir: outDir, size },
});

// Headless video has no pointer, so draw one that follows the real mouse.
await context.addInitScript(() => {
  const install = () => {
    if (document.getElementById('tour-cursor')) return;
    const cursor = document.createElement('div');
    cursor.id = 'tour-cursor';
    cursor.style.cssText = [
      'position:fixed', 'left:0', 'top:0', 'width:22px', 'height:22px', 'margin:-11px 0 0 -11px',
      'border-radius:50%', 'background:rgba(88,166,255,.35)', 'border:2px solid rgba(88,166,255,.95)',
      'z-index:2147483647', 'pointer-events:none', 'transition:transform .12s ease-out',
      'transform:translate(-100px,-100px)',
    ].join(';');
    document.body.appendChild(cursor);
    let x = -100;
    let y = -100;
    let scale = 1;
    const paint = () => { cursor.style.transform = `translate(${x}px,${y}px) scale(${scale})`; };
    addEventListener('mousemove', (e) => { x = e.clientX; y = e.clientY; paint(); }, true);
    addEventListener('mousedown', () => { scale = 0.6; paint(); }, true);
    addEventListener('mouseup', () => { scale = 1; paint(); }, true);
  };
  if (document.body) install();
  else addEventListener('DOMContentLoaded', install);
});

const page = await context.newPage();
const started = Date.now();
const pause = (ms) => page.waitForTimeout(ms);

async function glideTo(locator) {
  await locator.scrollIntoViewIfNeeded();
  const box = await locator.boundingBox();
  if (!box) throw new Error(`no bounding box for ${locator}`);
  await page.mouse.move(box.x + box.width / 2, box.y + box.height / 2, { steps: 24 });
  await pause(250);
}

async function click(locator) {
  await glideTo(locator);
  await locator.click();
}

async function scrollBy(total, stepMs = 16) {
  const steps = Math.max(1, Math.round(Math.abs(total) / 20));
  for (let i = 0; i < steps; i += 1) {
    await page.mouse.wheel(0, total / steps);
    await pause(stepMs);
  }
}

let tourStart = 0;
try {
  await page.goto(base, { waitUntil: 'networkidle' });
  await page.getByRole('heading', { name: 'Fleet dashboard' }).waitFor();
  await page.mouse.move(640, 300);
  await pause(400);
  tourStart = Date.now() - started;

  // 1. Fleet dashboard: three clusters, one degraded, one unreachable.
  await pause(1800);
  await scrollBy(330);
  await pause(1300);

  // 2. Open the degraded cluster.
  await click(page.getByRole('link', { name: /Open demo-eu-west cluster/ }));
  await page.getByRole('heading', { name: 'demo-eu-west' }).waitFor();
  await pause(1700);
  await scrollBy(520);
  await pause(1500);

  // 3. Stream logs from a pod through the agent channel.
  await click(page.getByRole('button', { name: /View logs for pod ledger-api/ }));
  await pause(3600);

  // 4. Fleet-wide policy and drift results.
  await click(page.getByRole('link', { name: /^Policy/ }));
  await page.getByRole('heading', { name: 'Policy and drift' }).waitFor();
  await pause(1600);
  await scrollBy(420);
  await pause(1700);

  // 5. Alert history.
  await click(page.getByRole('link', { name: /^Alerts/ }));
  await page.getByRole('heading', { name: 'Fleet alerts' }).waitFor();
  await pause(2400);
} catch (err) {
  console.error('tour failed:', err);
  process.exitCode = 1;
} finally {
  const video = page.video();
  await context.close();
  await browser.close();
  if (video) {
    const path = await video.path();
    await writeFile(`${outDir}/meta.json`, JSON.stringify({ path, tourStartMs: tourStart }));
    console.log(JSON.stringify({ path, tourStartMs: tourStart }));
  }
}
