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
Twenty seconds. This room already runs Temporal in production — they did not
come for a bio, and a long intro reads as filler to engineers who have on-call
rotations.

Mia is the click. She earns the "Dog Dad" line and buys a laugh before the first
diagram.

Then one question to the room before the TOC, because the answer sets your
pacing for the next twenty minutes: "who here has a Workflow in production that
calls two systems that both change state?" Every hand goes up. Say: then you
already own a Saga, whether or not it is written down as one.
-->

---
layout: toc
current: problem
---

<!--
Read the six rows, do not explain them. Twenty minutes of teaching, then the
rest of the session is the lab.

Shape of the session worth naming out loud: the teaching blocks map onto the
five TODOs in the exercise, grouped into the lab's three Challenges — unwind,
register first, make the undo safe to repeat. Nothing on a slide is background
material — every slide is a line they are about to type.
-->
