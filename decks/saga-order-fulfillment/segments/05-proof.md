---
layout: code-stack
heading: Three runs, one ledger
proseMax: 30
---

::code::

```text {1-2|4-5|7-8|all}
go run ./starter
  reserve -> charge -> dispatch

go run ./starter -fail-dispatch
  reserve -> charge -> refund -> release

go run ./starter -charge-times-out
  reserve -> charge -> refund -> release
```

::default::

Measured output from the finished lab, read off the Worker terminal. `ledger.go` records
what the downstream systems actually did — so this is the effect, not the intent.

<!--
These three lines are real. They come from running the solution tree in
workshops/saga-order-fulfillment/solution with a Temporal dev server up, and
reading the `[ledger]` lines the Worker prints. Do not describe them as an
example — describe them as the result, because that is what they are, and this
room can tell the difference.

- **Build 1 -** the clean order. Three forward effects; the undos on the slice
  never run, because the defer returns when `err` is nil. Nothing surprising,
  and that is the point: the Saga costs nothing when nothing fails.
- **Build 2 -** `-fail-dispatch`. Dispatch never appears — it failed before
  recording anything — and both undos ran.
- **Build 3 -** `-charge-times-out`. Look at it next to build 2. Same four
  entries.

Build 3 is the slide's payload. `charge` is in that ledger even though
`ChargePayment` returned an error, because the money moved before the gateway
reported failure. And `refund` is in it because the compensation was registered
before the step. Those two facts on one line are the whole session.

Have the room say the next slide's observation before you show it.
-->

---
layout: default
---

# `refund` before `release`

Both failure runs unwind in the same direction, and it is not alphabetical:

- `refund` is the **newest** registered compensation, so it goes first
- `release` was registered first, so it goes last
- LIFO, produced by `append` plus a loop from the last index down to 0

Walk the slice forward instead — `for _, undo := range compensations` — and the ledger
reads `release -> refund`. Still cleaned up. **Wrong order.**

<!--
This is the check to teach them to run on their own Sagas next week: read the
undo entries and confirm they are the mirror image of the forward entries.
`reserve -> charge` forward, `refund -> release` back. A palindrome.

Why the order matters here rather than in the abstract: the refund is what the
customer is watching. Putting stock back first is invisible to them and delays
the only step they care about. In a system with real dependencies it is worse
than cosmetic — a later step can depend on an earlier one still being in place,
and undoing the foundation first leaves the undo above it operating on a
half-dismantled record.

What the deck cannot show you, because it was not measured on the finished
code: the un-fixed exercise on `-charge-times-out`. Reason it out with the room
instead of claiming a number — the refund was never appended when the defer
fires, so the ledger reads `reserve -> charge -> release`. The stock came back;
the money did not. They will produce that output themselves in Challenge 2,
which is a better demonstration than a slide.
-->
