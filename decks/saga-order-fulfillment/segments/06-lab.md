---
layout: exercise
minutes: 25
heading: Hands-on lab
---

**[play.instruqt.com/temporal/tracks/saga-go](https://play.instruqt.com/temporal/tracks/saga-go)**

<!--
Leave this slide up while people get into the sandbox. It is the only URL they
have to type, so give them a moment on it.

Twenty-five minutes is the floor. If the session is running to time, this block
gets whatever is left of the forty-five minutes after the slides; the lab is the
point of the session and the slides are the setup.

Three Challenges, and every Check is gated: it runs tests against a dev server
of its own and fails with the file, the function and the TODO still to do. A
learner who is stuck can Skip one Challenge; Skip fills in only that
Challenge's TODOs and keeps the rest of their work.

Walk the room. The failures you will see, in order of frequency:
1. Editing the code and re-running the order without restarting the Worker.
   `go run ./worker` builds the code once, when it starts; a running Worker
   still has the old code. Ctrl+C and run it again. Always check this first.
2. Reading the Client terminal instead of the Worker's, and concluding
   nothing ran.
3. A forward `range` loop over the slice. The ledger says `release -> refund`
   and the room needs help seeing that it is wrong, because it looks cleaned
   up. The Check catches it.
4. Undos run on `ctx` instead of the disconnected context. Every scenario the
   learner runs by hand passes; only the Check's cancellation test fails. Use
   that message as the teaching moment.
5. TODO 3a/3b done with a guard on the wrong ledger check — `isCharged` in the
   release, `isReserved` in the refund. Symptom is an undo that no-ops when it
   should not.

Bring the room back on the Temporal UI, not the terminal. Open the failed
Workflow's history and show the compensating Activities scheduled after the
failure, in order. That picture is what they will remember.
-->

---
layout: checklist
eyebrow: "RECAP"
title: Seven rules for a Saga you can trust
subtitle: "All seven are in twenty lines of the file you just edited."
items:
  - "Distributed transactions have no rollback — the undo is business logic you write"
  - "Pair every forward step with a compensating action"
  - "No Saga helper: a slice of funcs and a deferred backwards loop is the whole pattern"
  - "Register each undo BEFORE its step — a step can land its effect and still report failure"
  - "Make every undo idempotent: it can run when nothing happened, and Temporal retries it"
  - "Run the undos on a disconnected context, and log each failure and keep going"
  - "Leave err alone after compensating — compensating is not succeeding"
---

<!--
Run this in the debrief, after the lab, not before it. Each row should now be a
thing they watched happen rather than a rule they were told.

Ask for the ledger from row four specifically: who ran `-charge-times-out`
before fixing 2b, and what did they see? Get someone to say "the charge was
there and the refund was not" out loud. That is the sentence that leaves the
room with them.

The closing frame for a team already running Temporal in production: none of
this needed a framework, a saga table, or a compensation queue. It needed the
compensation slice to survive a crash — and that came from where the code runs,
not from what it imports.
-->

---
layout: end
---

# Thank you

The exercise, the solution, and this deck are in the workshop repository.

<!--
TODO FOR THE DELIVERING INSTRUCTOR: one link still to fill in.

The Track is at play.instruqt.com/temporal/tracks/saga-go and is on the
hands-on lab slide already. What is still missing is the feedback form for this
Dev Day, which is event-specific and has no stable URL. Put it on this slide or
say it out loud.

Last thing to say: the Saga in the lab is twenty lines and it is not a toy. The
same twenty lines scale to a twelve-step Saga; the slice just gets longer. What
does not scale is remembering to register the undo before the step, which is why
it is a review checklist item and not a thing you rely on remembering.
-->
