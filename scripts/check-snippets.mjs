#!/usr/bin/env node
// ABOUTME: Fails when an assignment's code snippet has drifted from the Exercise.
//
// Assignments show the learner the exact code to change, so they never scroll to
// find it (ADR 0008). That only helps while the snippet and its line range are
// true. Mark each snippet in an assignment.md like this:
//
//     <!-- snippet: exercise/src/workflows/pack.ts TODO 1 -->
//     Lines 20-25 are the end of `packWorkflow`:
//
//     ```typescript
//     ...
//     ```
//
// The snippet is the region of the file from the `// TODO 1 start` line to the
// `// TODO 1 end` line, inclusive. Add `before=N` to the marker to include N
// lines above the start marker, e.g. the function signature. Paths are relative
// to the Workshop directory (the parent of `instruqt/`).
//
// The prose between the marker and the fence must say `lines A-B` with the
// region's real line numbers.
//
//     node scripts/check-snippets.mjs [workshops/<slug> ...] [--fix]
//
// With no path, checks every Workshop that has a sandbox/challenges.tsv (the
// gated ones). --fix rewrites the fences and line ranges in place instead of
// failing.

import { existsSync, readdirSync, readFileSync, writeFileSync } from 'node:fs';
import { dirname, join, relative, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const REPO = resolve(dirname(fileURLToPath(import.meta.url)), '..');
const MARKER = /<!--\s*snippet:\s*(\S+)\s+(TODO \w+)((?:\s+\w+=\d+)*)\s*-->/g;
const LINES = /[Ll]ines? (\d+)-(\d+)/;
const FENCE = /^```[^\n]*\n([\s\S]*?)^```/m;

function region(path, todo, before) {
  const lines = readFileSync(path, 'utf8').split('\n');
  let start = -1;
  let end = -1;
  lines.forEach((line, i) => {
    if (new RegExp(`//\\s*${todo} start\\b`).test(line)) start = i;
    else if (new RegExp(`//\\s*${todo} end\\b`).test(line)) end = i;
  });
  if (start < 0 || end < 0 || end < start) {
    throw new Error(`${relative(REPO, path)} has no '// ${todo} start' / '// ${todo} end' pair`);
  }
  start = Math.max(0, start - before);
  return { first: start + 1, last: end + 1, code: lines.slice(start, end + 1).join('\n') + '\n' };
}

function check(assignment, fix) {
  const workshop = resolve(dirname(assignment), '..', '..');
  let text = readFileSync(assignment, 'utf8');
  const problems = [];
  const out = [];
  let pos = 0;
  for (const m of text.matchAll(MARKER)) {
    const [, rel, todo, attrs] = m;
    const before = Number(/before=(\d+)/.exec(attrs)?.[1] ?? 0);
    const where = `${relative(REPO, assignment)}: ${rel} ${todo}`;
    const afterMarker = m.index + m[0].length;
    const fence = FENCE.exec(text.slice(afterMarker));
    if (!fence) {
      problems.push(`${where}: no code fence after the marker`);
      continue;
    }
    const fenceStart = afterMarker + fence.index;
    const codeStart = fenceStart + fence[0].indexOf('\n') + 1;
    let found;
    try {
      found = region(join(workshop, rel), todo, before);
    } catch (e) {
      problems.push(`${where}: ${e.message}`);
      continue;
    }
    const { first, last, code } = found;
    let prose = text.slice(afterMarker, fenceStart);
    const rng = LINES.exec(prose);
    if (!rng) {
      problems.push(`${where}: the text before the snippet must say 'lines ${first}-${last}'`);
    } else if (Number(rng[1]) !== first || Number(rng[2]) !== last) {
      problems.push(`${where}: says lines ${rng[1]}-${rng[2]}, the code is at lines ${first}-${last}`);
    }
    if (fence[1] !== code) problems.push(`${where}: snippet does not match the file`);
    if (fix) {
      if (rng) prose = prose.slice(0, rng.index) + rng[0].replace(/\d+-\d+/, `${first}-${last}`) + prose.slice(rng.index + rng[0].length);
      out.push(text.slice(pos, afterMarker) + prose + text.slice(fenceStart, codeStart) + code + '```');
      pos = fenceStart + fence[0].length;
    }
  }
  if (fix && out.length) {
    text = out.join('') + text.slice(pos);
    writeFileSync(assignment, text);
  }
  return problems;
}

const fix = process.argv.includes('--fix');
const args = process.argv.slice(2).filter((a) => !a.startsWith('--'));
const roots = args.length
  ? args.map((a) => resolve(a))
  : readdirSync(join(REPO, 'workshops'))
      .map((d) => join(REPO, 'workshops', d))
      .filter((d) => existsSync(join(d, 'sandbox', 'challenges.tsv')));
const assignments = roots
  .flatMap((r) => (existsSync(join(r, 'instruqt')) ? readdirSync(join(r, 'instruqt')).map((c) => join(r, 'instruqt', c, 'assignment.md')) : []))
  .filter((a) => existsSync(a))
  .sort();
const problems = assignments.flatMap((a) => check(a, fix));
const snippets = assignments.reduce((n, a) => n + [...readFileSync(a, 'utf8').matchAll(MARKER)].length, 0);
if (problems.length && !fix) {
  console.log(problems.map((p) => `[FAIL] ${p}`).join('\n'));
  console.log(`${problems.length} drifted snippet(s). Re-run with --fix to rewrite them from the code.`);
  process.exit(1);
}
console.log(`[ OK ] ${snippets} snippet(s) in ${assignments.length} assignment(s) match the code${fix && problems.length ? ' (fixed)' : ''}`);
