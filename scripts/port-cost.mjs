#!/usr/bin/env node
// Measures what a Port cost: agent tokens, dollars, and time.
//
//   node scripts/port-cost.mjs <project-slug> [--since 2026-11-13T15:35Z] [--until ...]
//                              [--label "..."] [--scope main|subagents|all]
//                              [--agent <id>[,<id>...]]
//
// <project-slug> is a directory under ~/.claude/projects, e.g.
// -Users-nikk-source-temporal-go-temporal-devdays-go.
//
// Several Ports in one session are told apart two ways (docs/ports/README.md):
//   --since/--until  a time window, for Ports done one after another;
//   --agent          only the subagent transcripts whose file name contains one
//                    of these ids, for Ports delegated to parallel subagents.
//                    Implies --scope subagents.
//
// Claude Code writes the main session to <project>/<session-id>.jsonl and each
// subagent to <project>/<session-id>/subagents/agent-*.jsonl. Main-session lines
// carry isSidechain:false, subagent lines isSidechain:true — which is how work
// delegated to an agent gets costed separately from work done in the main thread.
// We therefore walk the tree, not just the top level.
//
// Reports a FLOOR, not a total: only what the agent spent. Human review, sandbox
// debugging, and image rebuilds are invisible here and must be logged by hand.
import { readdirSync, readFileSync, statSync } from 'node:fs';
import { join } from 'node:path';
import { homedir } from 'node:os';

// $ per million tokens. Base rates are published; the cache multipliers are the
// standard Anthropic ratios (1h write 2x, 5m write 1.25x, read 0.1x) rather than
// numbers read off a rate card, so treat cache dollars as approximate.
// Same table as temporal-devdays-python's scripts/authoring-cost.mjs, so the two
// repos price a model the same way. `read` overrides the 0.1x cache-read ratio
// where the model's rate card says otherwise.
const RATES = {
  'claude-fable-5-1':  { in: 10, out: 50, read: 0.25 },
  'claude-fable-5':    { in: 10, out: 50 },
  'claude-opus-5-5':   { in: 4,  out: 20, read: 0.20 },
  'claude-opus-5':     { in: 5,  out: 25 },
  'claude-opus-4-8':   { in: 5,  out: 25 },
  'claude-sonnet-5-5': { in: 2,  out: 10, read: 0.20 },
  'claude-sonnet-5':   { in: 2,  out: 10 },
  'claude-haiku-4-5':  { in: 1,  out: 5  },
};
const CACHE_1H = 2.0, CACHE_5M = 1.25, CACHE_READ = 0.1;

// A gap longer than this is a break, not work. Wall-clock over a multi-day port
// is meaningless; summed sub-threshold gaps approximate time at the keyboard.
const IDLE_GAP_MS = 5 * 60 * 1000;

const [slug, ...rest] = process.argv.slice(2);
if (!slug) {
  console.error('usage: port-cost.mjs <project-slug> [--since YYYY-MM-DD] [--label "..."] [--scope main|subagents|all]');
  process.exit(1);
}
const arg = (n) => { const i = rest.indexOf(n); return i === -1 ? null : rest[i + 1]; };
const since = arg('--since') ? Date.parse(arg('--since')) : 0;
const until = arg('--until') ? Date.parse(arg('--until')) : Infinity;
const label = arg('--label') ?? slug;
const agents = arg('--agent') ? arg('--agent').split(',') : null;
const scope = agents ? 'subagents' : (arg('--scope') ?? 'all');
if (!['main', 'subagents', 'all'].includes(scope)) { console.error(`bad --scope: ${scope}`); process.exit(1); }

const walk = (d, acc = []) => {
  for (const e of readdirSync(d)) {
    const p = join(d, e);
    if (statSync(p).isDirectory()) walk(p, acc);
    else if (e.endsWith('.jsonl')) acc.push(p);
  }
  return acc;
};

const dir = join(homedir(), '.claude', 'projects', slug);
let files;
try { files = walk(dir); } catch { console.error(`no such project dir: ${dir}`); process.exit(1); }
if (files.length === 0) { console.error(`no transcripts under ${dir}`); process.exit(1); }

const blank = () => ({ models: new Map(), stamps: [], files: new Set() });
const buckets = { main: blank(), subagents: blank() };

for (const f of files) {
  if (agents && !agents.some((id) => f.includes(id))) continue;
  for (const line of readFileSync(f, 'utf8').split('\n')) {
    if (!line.trim()) continue;
    let e; try { e = JSON.parse(line); } catch { continue; }
    const u = e.message?.usage;
    if (e.type !== 'assistant' || !u) continue;
    const t = Date.parse(e.timestamp ?? '');
    if (!Number.isFinite(t) || t < since || t > until) continue;

    const which = e.isSidechain === true ? 'subagents' : 'main';
    if (scope !== 'all' && scope !== which) continue;
    const b = buckets[which];
    b.stamps.push(t);
    b.files.add(f);

    const model = e.message.model ?? 'unknown';
    const c = b.models.get(model) ?? { input: 0, output: 0, thinking: 0, w1h: 0, w5m: 0, read: 0, msgs: 0 };
    c.msgs++;
    c.input    += u.input_tokens ?? 0;
    c.output   += u.output_tokens ?? 0;
    c.thinking += u.output_tokens_details?.thinking_tokens ?? 0;
    c.read     += u.cache_read_input_tokens ?? 0;
    c.w1h      += u.cache_creation?.ephemeral_1h_input_tokens ?? 0;
    c.w5m      += u.cache_creation?.ephemeral_5m_input_tokens ?? 0;
    b.models.set(model, c);
  }
}

const costOf = (model, c) => {
  const r = RATES[model];
  if (!r) return null;
  return (c.input / 1e6) * r.in + (c.output / 1e6) * r.out
    + (c.w1h / 1e6) * r.in * CACHE_1H + (c.w5m / 1e6) * r.in * CACHE_5M
    + (c.read / 1e6) * (r.read ?? r.in * CACHE_READ);
};
const activeMs = (stamps) => {
  const s = [...stamps].sort((a, b) => a - b);
  let a = 0;
  for (let i = 1; i < s.length; i++) { const d = s[i] - s[i - 1]; if (d < IDLE_GAP_MS) a += d; }
  return a;
};
const hrs = (ms) => (ms / 3_600_000).toFixed(1);
const fmt = (n) => n.toLocaleString('en-US');

const summarize = (b) => {
  if (b.stamps.length === 0) return null;
  let cost = 0;
  for (const [m, c] of b.models) { const x = costOf(m, c); if (x !== null) cost += x; }
  const s = [...b.stamps].sort((a, z) => a - z);
  return { cost, active: activeMs(b.stamps), msgs: b.stamps.length, first: s[0], last: s.at(-1), files: b.files.size };
};
const mainS = summarize(buckets.main);
const subS  = summarize(buckets.subagents);
if (!mainS && !subS) { console.error('no assistant messages matched'); process.exit(1); }

const all = [...buckets.main.stamps, ...buckets.subagents.stamps].sort((a, b) => a - b);
const out = [];
out.push(`# Port cost: ${label}`, '');
out.push(`Measured ${new Date(all[0]).toISOString().slice(0, 10)} to ${new Date(all.at(-1)).toISOString().slice(0, 10)}.`, '');
out.push('| Scope | Cost | Active | Messages | Transcripts |');
out.push('|---|--:|--:|--:|--:|');
const row = (n, s) => s && out.push(`| ${n} | $${s.cost.toFixed(2)} | ${hrs(s.active)} h | ${fmt(s.msgs)} | ${s.files} |`);
row('Main session', mainS);
row('Subagents', subS);
if (mainS && subS) {
  out.push(`| **Total** | **$${(mainS.cost + subS.cost).toFixed(2)}** | — | ${fmt(mainS.msgs + subS.msgs)} | ${mainS.files + subS.files} |`);
}
out.push('', '> Active time is summed gaps under 5 minutes, so an idle session does not inflate it.',
  '> Main and subagent time overlap when they run concurrently — do not add the two.', '');

for (const [name, b] of [['Main session', buckets.main], ['Subagents', buckets.subagents]]) {
  if (b.models.size === 0) continue;
  out.push(`## ${name} — by model`, '');
  out.push('| Model | Msgs | Input | Output | Thinking | Cache write | Cache read | Cost |');
  out.push('|---|--:|--:|--:|--:|--:|--:|--:|');
  for (const [m, c] of [...b.models].sort((a, z) => z[1].output - a[1].output)) {
    const x = costOf(m, c);
    out.push(`| \`${m}\` | ${fmt(c.msgs)} | ${fmt(c.input)} | ${fmt(c.output)} | ${fmt(c.thinking)} `
      + `| ${fmt(c.w1h + c.w5m)} | ${fmt(c.read)} | ${x === null ? 'n/a' : '$' + x.toFixed(2)} |`);
  }
  out.push('');
}
out.push('## Human time', '', '_Not measured. Fill in by hand._', '',
  '| Activity | Hours |', '|---|--:|', '| Review and correction | |',
  '| Sandbox / Instruqt debugging | |', '| Deck rehearsal | |', '',
  '> The agent figures above are a **floor**, not the cost of the Port.');
console.log(out.join('\n'));
