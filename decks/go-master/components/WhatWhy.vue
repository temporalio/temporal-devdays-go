<!--
  ABOUTME: A feature's case in two cards: what it is, and why it is worth having.
  ABOUTME: Used on the More Features slides ahead of each feature's code.

    <WhatWhy>
      <template #what>One or two sentences.</template>
      <template #why>
        <li>A benefit</li>
      </template>
    </WhatWhy>
-->
<template>
  <div class="ww">
    <section class="card what">
      <h3>What it is</h3>
      <div class="body"><slot name="what" /></div>
    </section>
    <section class="card why">
      <h3>Why it matters</h3>
      <ul class="body"><slot name="why" /></ul>
    </section>
  </div>
</template>

<style scoped>
.ww {
  display: grid;
  grid-template-columns: minmax(0, 0.85fr) minmax(0, 1.15fr);
  gap: 1.2rem;
  margin-top: 1rem;
}
.card {
  border-radius: 0.7rem;
  padding: 1rem 1.2rem;
  background: var(--temporal-bg-elev);
  border: 1px solid var(--temporal-border-hi);
}
.what { border-top: 3px solid var(--temporal-lavender); }
.why { border-top: 3px solid var(--temporal-green); }
h3 {
  margin: 0 0 0.6rem !important;
  font-size: 0.7rem !important;
  font-weight: 500;
  letter-spacing: 0.12em;
  text-transform: uppercase;
}
.what h3 { color: var(--temporal-lavender); }
.why h3 { color: var(--temporal-green); }
/* The theme sizes every p and li for prose slides; these cards need their own. */
.body,
.body :slotted(p),
.body :slotted(li) {
  font-size: 0.95rem !important;
  line-height: 1.5 !important;
  color: var(--temporal-text-strong);
}
.body :slotted(p) {
  margin: 0 !important;
}
ul.body {
  margin: 0;
  padding-left: 1.1rem;
}
ul.body :slotted(li) {
  margin: 0 0 0.45rem;
}
</style>
