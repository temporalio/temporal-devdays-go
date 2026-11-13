---
layout: default
class: "!p-0"
---

<div class="absolute inset-0 flex flex-col items-center justify-center gap-6">

  <div class="flex items-center gap-2">
    <img
      v-click
      :src="'/mia.jpg'"
      alt="Mia, a black and tan rescue dog"
      class="mia-photo w-[248px] h-[248px] rounded-full object-cover"
      style="object-position: 50% 30%"
    />
    <img
      :src="'/nikolay.jpg'"
      alt="Nikolay Advolodkin"
      class="w-[248px] h-[248px] rounded-full object-cover"
    />
  </div>

  <h1 class="intro-name">Nikolay Advolodkin</h1>

  <div class="flex flex-col items-center gap-1">
    <p class="intro-line">Staff Developer Advocate @ Temporal</p>
    <p class="intro-line">Dog Dad &amp; Roller Skater</p>
  </div>

  <div class="flex items-center gap-10">
    <a class="intro-social" href="https://x.com/nikolay_a00">
      <img :src="'/icon-x.png'" alt="X" class="w-[34px] h-[34px] rounded-md" />
      <span>@nikolay_a00</span>
    </a>
    <a class="intro-social" href="https://www.linkedin.com/in/nikolayadvolodkin/">
      <img :src="'/icon-linkedin.png'" alt="LinkedIn" class="w-[34px] h-[34px] rounded-md" />
      <span>/in/nikolayadvolodkin/</span>
    </a>
  </div>

</div>

<style scoped>
/* v-click only fades opacity, which leaves Mia's 248px reserved and pushes the
   headshot off-centre before the click. Collapsing her instead keeps the first
   state centred and lets her pop in. */
.mia-photo.slidev-vclick-hidden {
  display: none;
}
.intro-name {
  font-size: 2.9rem;
  font-weight: 200;
  letter-spacing: -0.02em;
  line-height: 1;
  color: var(--temporal-text-strong);
  margin: 0;
}
.intro-line {
  font-size: 1.15rem;
  color: var(--temporal-text);
  margin: 0;
}
.intro-social {
  display: inline-flex;
  align-items: center;
  gap: 0.6rem;
  font-size: 1.1rem;
  color: var(--temporal-text);
  text-decoration: none;
  /* The theme underlines every link with a border-bottom; not wanted here. */
  border-bottom: none;
}
.intro-social:hover {
  color: var(--temporal-green);
}
</style>

<!--
Source: decks/_template/segments/00-open.md. Carried over from the original Essentials Deck.
Every Deck opens this way: cover, presenter, agenda.

Twenty seconds. This is a Foundations room, so a short intro is welcome, but it
is still not a bio. Mia is the click: she earns the "Dog Dad" line and buys a
laugh before the first diagram.

One question to the room before the agenda, because the answer sets your pacing:
"who has run a Workflow before today?" A room of mostly no is a room that needs
the concepts segment slowly.
-->

---
layout: toc
current: why
---

<style>
/* Every topic in the master, in delivery order: too many for the theme's
   one-column list, so this slide flows it into two columns. The list comes
   from themeConfig.toc in slides.md; keep it in step when segments are added
   or hidden. */
:global(.workshop-toc) {
  display: grid !important;
  grid-auto-flow: column;
  grid-template-rows: repeat(14, auto);
  column-gap: 3rem;
  row-gap: 0.45rem !important;
}
:global(.workshop-toc li) {
  font-size: 1.05rem !important;
  line-height: 1.25 !important;
  margin: 0 !important;
  padding: 0 !important;
}
:global(.workshop-toc li .number) {
  font-size: 0.8rem !important;
}
</style>

<!--
Source: PPT slide 3, expanded to every topic in this master, in delivery order.
The list lives in themeConfig.toc in slides.md: 27 entries, so the grid above
has 14 rows per column. The labs slide for Child Workflows, Heartbeats and
Continue-As-New, and the Saga lab, are left off; if you hide or add a segment,
update the list and the row count.
Read the rows, do not explain them.
-->
