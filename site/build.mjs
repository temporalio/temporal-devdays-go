#!/usr/bin/env node
// ABOUTME: Builds the landing page (site/public/) from site/labs.json.
//
//   node site/build.mjs          # HTML only
//   node site/build.mjs --og     # also renders each lab's 1200x630 social card
//                                # with local Chrome (commit the PNGs)
//
// One page for every lab, plus one page per lab. A social post links to the
// lab's own page, so its preview card shows that lab, not the index.
import { execFileSync } from 'node:child_process';
import { existsSync, mkdirSync, readFileSync, writeFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const HERE = dirname(fileURLToPath(import.meta.url));
const OUT = join(HERE, 'public');
const { base, repo, labs } = JSON.parse(readFileSync(join(HERE, 'labs.json'), 'utf8'));
const css = readFileSync(join(HERE, 'style.css'), 'utf8');

const esc = (s) => String(s).replace(/[&<>"]/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;' })[c]);
const fonts = '<link rel="preconnect" href="https://fonts.googleapis.com"><link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>'
  + '<link href="https://fonts.googleapis.com/css2?family=Bricolage+Grotesque:opsz,wght@12..96,400;12..96,600;12..96,800&family=JetBrains+Mono:wght@400;600&display=swap" rel="stylesheet">';

const startButton = (lab, big = false) => lab.invite
  ? `<a class="start${big ? ' big' : ''}" href="${esc(lab.invite)}">Start the lab</a>`
  : `<span class="start soon${big ? ' big' : ''}" aria-disabled="true">Opens soon</span>`;

// An Event History, as the Temporal UI lists it. Event IDs are a real sequence.
const history = (events, { first = 1, animate = false, note = null } = {}) => {
  const rows = events.map(([type, detail, tone], i) => {
    const noteRow = note && note.after === i - 1
      ? `<li class="note" style="--i:${i}"><span></span><span>${esc(note.text)}</span></li>` : '';
    return `${noteRow}<li class="ev ${tone}" style="--i:${i + (note && i > note.after ? 1 : 0)}">`
      + `<span class="id">${first + i}</span><span class="type">${esc(type)}</span>`
      + `<span class="detail">${esc(detail)}</span></li>`;
  }).join('');
  return `<ol class="history${animate ? ' play' : ''}" aria-label="Event History">${rows}</ol>`;
};

const page = ({ title, description, url, image, body }) => `<!doctype html>
<html lang="en"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1">
<title>${esc(title)}</title>
<meta name="description" content="${esc(description)}">
<meta property="og:type" content="website"><meta property="og:url" content="${url}">
<meta property="og:title" content="${esc(title)}"><meta property="og:description" content="${esc(description)}">
<meta property="og:image" content="${image}"><meta property="og:image:width" content="1200"><meta property="og:image:height" content="630">
<meta name="twitter:card" content="summary_large_image"><meta name="twitter:image" content="${image}">
${fonts}<style>${css}</style></head>
<body>${body}
<footer class="foot"><p>Built from <a href="${repo}">temporal-devdays-go</a>. Every lab is open source: the code, the Checks and the slides.</p></footer>
</body></html>`;

const levels = [
  ['Foundations', 'New to Temporal. Start here.'],
  ['Advanced', 'Already running Workflows. Go deeper.'],
];

const heroEvents = [
  ['WorkflowExecutionStarted', 'transfer-1042', ''],
  ['ActivityTaskCompleted', 'Withdraw $2,500', 'ok'],
  ['ActivityTaskScheduled', 'Deposit $2,500', ''],
  ['ActivityTaskStarted', 'Deposit, attempt 4', 'warn'],
  ['ActivityTaskCompleted', 'Deposit $2,500', 'ok'],
  ['WorkflowExecutionCompleted', 'withdrawn once, deposited once', 'ok'],
];

const index = page({
  title: 'Temporal labs in Go',
  description: 'Six hands-on Temporal labs in Go that run in your browser: Workflows, Child Workflows, Continue-As-New, Heartbeats, Sagas and Nexus. No install, no account.',
  url: `${base}/`,
  image: `${base}/og.png`,
  body: `<main>
<header class="hero">
  <div class="pitch">
    <p class="kicker">Temporal DevDays, in Go</p>
    <h1>Crash the Worker. The transfer still finishes.</h1>
    <p class="lede">Six hands-on Temporal labs written in Go. Each one opens in your browser with an editor, a terminal and a Temporal Service already running. No install and no account.</p>
    <a class="start big" href="#labs">Pick a lab</a>
  </div>
  <figure class="panel">
    ${history(heroEvents, { animate: true, note: { after: 2, text: 'Worker crashed here. Nothing was lost.' } })}
    <figcaption>The Event History of a transfer from the first lab.</figcaption>
  </figure>
</header>
<section id="labs" class="labs">
${levels.map(([level, blurb]) => `<h2>${level}<span>${blurb}</span></h2>
<ul class="rows">${labs.filter((l) => l.level === level).map((l) => `
  <li class="row">
    <div><h3><a href="${l.slug}/">${esc(l.title)}</a></h3><p>${esc(l.pitch)}</p></div>
    <p class="mins">${l.minutes} min</p>
    ${startButton(l)}
  </li>`).join('')}
</ul>`).join('\n')}
</section>
</main>`,
});

const labPage = (l) => page({
  title: `${l.title}: a Temporal lab in Go`,
  description: `${l.pitch} A ${l.minutes}-minute hands-on lab that runs in your browser.`,
  url: `${base}/${l.slug}/`,
  image: `${base}/${l.slug}/og.png`,
  body: `<main class="lab">
<p class="back"><a href="../">All Temporal labs in Go</a></p>
<header class="hero">
  <div class="pitch">
    <p class="kicker">${l.level} lab, ${l.minutes} minutes, in Go</p>
    <h1>${esc(l.title)}</h1>
    <p class="lede">${esc(l.pitch)}</p>
    <ol class="steps">${l.steps.map((s) => `<li>${esc(s)}</li>`).join('')}</ol>
    ${startButton(l, true)}
    <p class="aside">Runs in your browser. Code: <a href="${repo}/tree/main/workshops/${l.workshop}">workshops/${l.workshop}</a></p>
  </div>
  <figure class="panel">${history(l.events, { first: 5, animate: true })}
    <figcaption>What you will see in the Event History.</figcaption></figure>
</header>
</main>`,
});

// The social card: the lab's title beside the lab's own Event History.
const card = (title, sub, events) => `<!doctype html><html data-theme="dark"><head><meta charset="utf-8">${fonts}<style>${css}
html,body{margin:0;width:1200px;height:630px;overflow:hidden}
.card{box-sizing:border-box;width:1200px;height:630px;padding:64px 64px 56px;display:grid;grid-template-columns:1fr 520px;gap:48px;align-items:center;background:var(--paper)}
.card h1{font-size:76px;margin:0 0 24px}
.card p{font-size:28px;color:var(--muted);margin:0}
.card .kicker{font-size:24px;margin-bottom:28px}
.card .history{font-size:19px}
</style></head><body><div class="card"><div><p class="kicker">A Temporal lab in Go</p><h1>${esc(title)}</h1><p>${esc(sub)}</p></div>
<div class="panel">${history(events, { first: 5 })}</div></div></body></html>`;

mkdirSync(OUT, { recursive: true });
writeFileSync(join(OUT, 'index.html'), index);
writeFileSync(join(OUT, '.nojekyll'), '');
for (const l of labs) {
  mkdirSync(join(OUT, l.slug), { recursive: true });
  writeFileSync(join(OUT, l.slug, 'index.html'), labPage(l));
}

if (process.argv.includes('--og')) {
  const chrome = '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome';
  if (!existsSync(chrome)) throw new Error(`--og needs Chrome at ${chrome}`);
  const shoot = (html, png) => {
    const tmp = join(OUT, '.card.html');
    writeFileSync(tmp, html);
    execFileSync(chrome, ['--headless=new', '--disable-gpu', '--hide-scrollbars', '--force-device-scale-factor=1',
      '--virtual-time-budget=4000', `--screenshot=${png}`, '--window-size=1200,630', `file://${tmp}`], { stdio: 'ignore' });
    execFileSync('rm', [tmp]);
  };
  shoot(card('Crash the Worker. It still finishes.', 'Six hands-on labs that run in your browser.', heroEvents.slice(1)), join(OUT, 'og.png'));
  for (const l of labs) shoot(card(l.title, `${l.minutes} minutes, in your browser. No install.`, l.events), join(OUT, l.slug, 'og.png'));
}
console.log(`site/public: index + ${labs.length} lab pages${process.argv.includes('--og') ? ' + social cards' : ''}`);
