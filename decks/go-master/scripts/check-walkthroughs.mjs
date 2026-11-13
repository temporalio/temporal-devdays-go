// Checks every code-walkthrough slide, click by click, that each bubble sits on
// the lines it is lit against. The static test (snippets/deck_test.go, TestWalkthroughs)
// proves the right lines are lit; this proves the layout put the bubble there.
//
// Walks the whole deck rather than hardcoding slide numbers, so it survives
// slides being added or removed.
//
// Run with the dev server up:  node scripts/check-walkthroughs.mjs
//   SLIDEV_URL=http://localhost:3031 node scripts/check-walkthroughs.mjs
import { chromium } from "playwright-chromium";

const BASE = process.env.SLIDEV_URL || "http://localhost:3030";
// How far a bubble's tail may point from the centre of its lit lines, as a
// fraction of one code line's height. The layout aims exactly, so anything
// past half a line means the bubble is visibly beside the wrong line. A
// bubble must also sit wholly inside the slide body.
const TOLERANCE_LINES = 0.5;

const browser = await chromium.launch();
const page = await browser.newPage({ viewport: { width: 1600, height: 900 } });

await page.goto(`${BASE}/1`, { waitUntil: "networkidle" });
await page.waitForTimeout(800);
const total = await page.evaluate(() => {
  const m = document.body.innerText.match(/\b\d{1,3}\s*\/\s*(\d{1,3})\b/);
  return m ? Number(m[1]) : 0;
});
if (!total) {
  console.error("Could not read the slide count. Is the dev server up?");
  process.exit(2);
}

let slides = 0;
let clicks = 0;
const failures = [];

for (let n = 1; n <= total; n++) {
  await page.goto(`${BASE}/${n}`, { waitUntil: "networkidle" });
  await page.waitForTimeout(500);
  // Slidev keeps the neighbouring slides mounted but hidden, so count bubbles
  // only in the walkthrough that is actually on screen.
  const bubbles = await page.evaluate(() => {
    const layout = [...document.querySelectorAll(".code-walkthrough")].find((e) => e.clientHeight > 0);
    return layout ? layout.querySelectorAll(".bubble").length : 0;
  });
  if (!bubbles) continue;
  slides++;

  for (let k = 1; k <= bubbles; k++) {
    await page.goto(`${BASE}/${n}?clicks=${k}`, { waitUntil: "networkidle" });
    // Fonts, the auto-fit pass and the bubble's 220ms slide all have to settle.
    await page.waitForTimeout(900);

    const r = await page.evaluate(() => {
      const layout = [...document.querySelectorAll(".code-walkthrough")].find((e) => e.clientHeight > 0);
      if (!layout) return null;
      const heading = layout.querySelector(".heading")?.textContent?.trim() ?? "(no heading)";
      const lines = [...layout.querySelectorAll(".code-pane .line")];
      const lit = lines.filter((l) => l.classList.contains("highlighted"));
      const shown = [...layout.querySelectorAll(".bubble")].filter(
        (b) => !b.classList.contains("slidev-vclick-hidden") && getComputedStyle(b).opacity !== "0",
      );
      const body = layout.querySelector(".cw-body").getBoundingClientRect();
      if (shown.length !== 1) return { heading, problem: `${shown.length} bubbles showing` };
      if (!lit.length || lit.length === lines.length) return { heading, problem: "no single range lit" };
      const el = shown[0];
      const b = el.getBoundingClientRect();
      const first = lit[0].getBoundingClientRect();
      const last = lit[lit.length - 1].getBoundingClientRect();
      const lineHeight = first.height || 1;
      // Measure where the tail points, not the bubble's centre: near the top
      // or bottom of a snippet the layout slides the bubble inside the slide
      // and moves the tail to compensate. The tail's `top` is in unscaled
      // canvas px, so scale it to screen px like the rects around it.
      const tailTop = parseFloat(getComputedStyle(el, "::before").top);
      const scale = el.offsetHeight ? b.height / el.offsetHeight : 1;
      const tailY = b.top + tailTop * scale;
      return {
        heading,
        text: el.textContent.trim().slice(0, 60),
        offsetLines: (tailY - (first.top + last.bottom) / 2) / lineHeight,
        clipped: b.top < body.top - 1 || b.bottom > body.bottom + 1,
      };
    });

    clicks++;
    if (!r) continue;
    const off = r.offsetLines ?? 0;
    const bad = r.problem || r.clipped || Math.abs(off) > TOLERANCE_LINES;
    const where = `${String(n).padStart(3)}.${k} ${r.heading}`;
    if (bad) {
      const why = r.problem || (r.clipped ? "bubble clipped by the slide" : `bubble ${off.toFixed(2)} lines off`);
      failures.push(`${where}: ${why}${r.text ? `  ("${r.text}")` : ""}`);
      console.log(`MISALIGNED ${where}: ${why}`);
    } else {
      console.log(`ok         ${where} (${off.toFixed(2)} lines)`);
    }
  }
}

console.log(
  failures.length === 0
    ? `\nAll ${clicks} bubbles on ${slides} walkthrough slides sit on their lines.`
    : `\n${failures.length} of ${clicks} bubbles misaligned:\n  ${failures.join("\n  ")}`,
);
await browser.close();
process.exit(failures.length === 0 && slides > 0 ? 0 : 1);
