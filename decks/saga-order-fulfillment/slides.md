---
theme: ./theme
title: Patterns and anti-patterns
info: |
  ## Saga order fulfillment — Go
  45 minutes. One hands-on lab. Take an order Workflow that strands a charge
  when dispatch fails and turn it into a Saga that unwinds correctly, in the
  right direction, without a Saga helper.
author: Nikolay Advolodkin
keywords: temporal,saga,compensation,go,golang,workflows,anti-patterns
colorSchema: dark
fonts:
  sans: "Inter"
  mono: "Noto Sans Mono"
  weights: "200,300,400,500,600"
  italic: false
mdc: true
themeConfig:
  footer: "Patterns and anti-patterns | Go"
  toc:
    - id: problem
      label: The problem
    - id: saga
      label: The Saga pattern
    - id: shape
      label: The Go shape
    - id: antipatterns
      label: Three anti-patterns
    - id: proof
      label: Proof it unwound
    - id: lab
      label: Hands-on lab
layout: cover
variant: planet-teal
---

# Patterns and anti-patterns

## Saga compensations in Go

<!--
Title slide. Thirty seconds.

Say the promise before you move on: by the end of this session an order that
fails halfway through releases the stock and refunds the card, in that order,
every time — including the case where the charge lands and the gateway lies to
you about it, and the case where someone cancels the order mid-flight.

SCOPE NOTE FOR THE DELIVERING INSTRUCTOR: title and content agree — this is
patterns and anti-patterns, and compensation ordering is the whole teaching arc.
Compensations here unwind sequentially on purpose; LIFO is the point.

If the room asks about running compensations in parallel (`workflow.Go` over
independent undos, and what that does to the unwind order), take it in the
debrief rather than at the slides. There is no parallel step in the exercise
code, so anything said about it will not match what they just typed.
-->

---
src: ./segments/00-open.md
---

---
src: ./segments/01-problem.md
---

---
src: ./segments/02-saga.md
---

---
src: ./segments/03-shape.md
---

---
src: ./segments/04-antipatterns.md
---

---
src: ./segments/05-proof.md
---

---
src: ./segments/06-lab.md
---
