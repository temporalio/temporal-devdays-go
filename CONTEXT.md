# Temporal DevDays — Go

Vocabulary for the Go DevDays content program: the Workshops we deliver to
customers, and the work of recreating them across Temporal's supported SDKs.
Shared with `temporal-devdays-ts`, `-python` and `-java`; a term means the same
thing in every one of them.

## Language

**DevDays**:
The content program itself — the Workshops, their Ports across languages, and
the Decks they are delivered from. Names the ongoing body of work, never a
single event.
_Avoid_: Dev Days (the two-word form names the events), the program, DevDay

**Dev Day**:
A delivery event run for one customer, assembled from one or more Workshops.
_Avoid_: workshop (a Dev Day is the event, not the material), training, session

**Workshop**:
A reusable teaching unit on one topic — its Deck, its Track, and its Exercise
and Solution code. The thing that gets assembled into a Dev Day and Ported
between languages.
_Avoid_: course, lab, content, module

**Level**:
The experience a Workshop is written for — **Foundations** (new to Temporal) or
**Advanced** (already building with it). Every Workshop has exactly one.
_Avoid_: difficulty, tier, 101/201

**Port**:
Recreating an existing Workshop for a Temporal SDK language it does not yet
support, preserving the narrative while rewriting the code idiomatically. Covers
every artifact that recreation touches. One Port, one cost record, however many
artifacts it spans.
_Avoid_: translation, rewrite, conversion

**Track**:
The Instruqt artifact for a Workshop — a `track.yml`, a `config.yml`, and its
Challenges. One Track per Workshop per language.
_Avoid_: lab, workshop, course

**Challenge**:
One numbered step within a Track, with an assignment and its lifecycle scripts.
_Avoid_: exercise (that word names the code tree), lesson, step

**Check**:
The script behind a Challenge's Check button. It decides whether the learner
may advance, and when they may not, says what is missing and where.
_Avoid_: test (the Check runs the tests; it is not one), validation, grader

**Skip**:
The action that completes one Challenge's TODOs from the Solution, keeping the
learner's other work, so the next Challenge is always runnable.
_Avoid_: solve, cheat, auto-complete

**Invite**:
A public Instruqt link that lets anyone play one Track without an account.
A **Social Invite** is the one Invite per Track kept for social media posts,
separate from Invites made for a Dev Day so their numbers do not mix.
_Avoid_: share link, public link, lab link

**Deck**:
The Slidev presentation delivered alongside a Workshop.
_Avoid_: slides, presentation

**Code Walkthrough**:
A Deck slide that steps through one code block: one highlight and one short note
per click, in the order the code runs. The fuller explanation lives in the
presenter notes.
_Avoid_: annotated code, code slide

**Exercise**:
The code tree a learner starts from, with the work left undone as numbered TODOs.
_Avoid_: starter, skeleton, template

**Solution**:
The completed counterpart of an Exercise, structurally identical to it.
_Avoid_: answer, final, complete

**Sandbox Image**:
The container image Instruqt boots for a Track, holding the toolchain, the
editor, and the pre-warmed dependency cache.
_Avoid_: box, environment, VM

## Temporal UI

**Temporal UI**:
The browser app served by the local dev server, open in a Challenge's
**Temporal UI** tab.
_Avoid_: Web UI, dashboard, console

**Event History**:
The Service's record of every step one Workflow Execution took, as events in
order.
_Avoid_: history log, event log

**Attempt**:
One run of an Activity. The first run is attempt 1, and each retry adds one.
_Avoid_: try, retry number

**Crash**:
A Worker stopping without telling the Temporal Service, as Ctrl+C does in a
Track's Worker tab.
_Avoid_: shutdown, stop

**Shutdown**:
A Worker stopping after telling the Temporal Service, finishing or handing back
the work it held.
_Avoid_: crash, kill
