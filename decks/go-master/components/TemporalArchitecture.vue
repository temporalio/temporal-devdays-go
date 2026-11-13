<!--
  ABOUTME: The three-zone Temporal architecture diagram — your app, the Server, your Workers.
  ABOUTME: Hand-built rather than Mermaid so the zone order and the arrow directions are exact.

  Usage, with no props:

    <TemporalArchitecture />

  Why not Mermaid: the whole point of the picture is that the two outer zones
  belong to you and the arrows into the middle both start outside it. Mermaid
  lays subgraphs out from edge order, so the `Workers -> Matching` back-edge
  pulled the Workers zone above the Server and told the opposite story. It also
  clipped the subgraph and edge labels at this size. A CSS grid holds the three
  zones where the argument needs them.

  The `is-yours` zones are mint, the Server is purple: colour carries the
  ownership split, so the slide reads before a word of it is spoken.

  The two outer arrows point at the Server zone rather than at a service inside
  it, which is true at this altitude — both of them really do land on Frontend,
  and the inside of the zone shows where traffic goes from there.
-->
<template>
  <div class="arch">
    <!-- Left: the caller -->
    <div class="zone is-yours">
      <p class="zone-label">YOURS</p>
      <div class="box">
        <p class="box-title">Temporal Client</p>
        <p class="box-sub">start · signal · query · cancel</p>
        <p class="box-note">a library in your app</p>
      </div>
    </div>

    <div class="link">
      <span class="link-label">gRPC</span>
      <span class="arrow arrow-right" />
    </div>

    <!-- Middle: the platform -->
    <div class="zone is-server">
      <p class="zone-label">TEMPORAL SERVER</p>

      <div class="box box-sm">
        <p class="box-title">Frontend</p>
        <p class="box-sub">the API door</p>
      </div>

      <!-- Frontend forks to both History and Matching. Drawn as SVG because a
           percentage width on a 1px-wide element resolves against that 1px,
           so a pseudo-element crossbar comes out 1.5px long. -->
      <svg
        class="wire wire-fork"
        viewBox="0 0 100 14"
        preserveAspectRatio="none"
        aria-hidden="true"
      >
        <path
          d="M50 0 V7 M25 7 H75 M25 7 V14 M75 7 V14"
          vector-effect="non-scaling-stroke"
        />
      </svg>

      <div class="pair">
        <div class="box box-sm">
          <p class="box-title">History</p>
          <p class="box-sub">Workflow state + Event History</p>
        </div>
        <div class="box box-sm">
          <p class="box-title">Matching</p>
          <p class="box-sub">Task Queues</p>
        </div>
      </div>

      <svg
        class="wire wire-drop"
        viewBox="0 0 100 10"
        preserveAspectRatio="none"
        aria-hidden="true"
      >
        <path d="M50 0 V10" vector-effect="non-scaling-stroke" />
      </svg>

      <div class="box box-sm box-db">
        <p class="box-title">Database</p>
        <p class="box-sub">the durable record — every step, kept</p>
      </div>

      <p class="zone-foot">runs none of your code</p>
    </div>

    <div class="link">
      <span class="arrow arrow-left" />
      <span class="link-label">long-poll<br />for Tasks</span>
    </div>

    <!-- Right: the executors -->
    <div class="zone is-yours">
      <p class="zone-label">YOURS</p>
      <div class="box">
        <p class="box-title">Workers</p>
        <p class="box-sub">your Workflow + Activity code</p>
        <p class="box-note">on your infrastructure</p>
      </div>
    </div>
  </div>
</template>

<style scoped>
.arch {
  display: grid;
  /* zone · link · zone · link · zone. The links are fixed so the three zones
     share the slack evenly and the middle one does not balloon. */
  grid-template-columns: 1fr 4.8rem 1.62fr 4.8rem 1fr;
  align-items: center;
  width: 100%;
}

.zone {
  border: 1px solid;
  border-radius: 0.6rem;
  padding: 0.65rem 0.8rem 0.7rem;
  background: rgb(255 255 255 / 2.5%);
}
.zone.is-yours {
  border-color: var(--temporal-green);
}
.zone.is-server {
  border-color: var(--temporal-purple);
  background: rgb(93 36 189 / 12%);
}

.zone-label {
  margin: 0 0 0.6rem;
  font-size: 0.74rem;
  font-weight: 500;
  letter-spacing: 0.1em;
  text-align: center;
}
.is-yours .zone-label {
  color: var(--temporal-green);
}
.is-server .zone-label {
  color: var(--temporal-lavender);
}
.zone-foot {
  margin: 0.6rem 0 0;
  font-size: 0.72rem;
  font-style: italic;
  text-align: center;
  color: var(--temporal-text-muted);
}

.box {
  border: 1px solid var(--temporal-border-hi);
  border-radius: 0.4rem;
  background: var(--temporal-bg-elev);
  padding: 0.62rem 0.65rem;
  text-align: center;
}
.box-sm {
  padding: 0.5rem 0.55rem;
}
.box-title {
  margin: 0;
  font-size: 0.96rem;
  font-weight: 500;
  color: var(--temporal-text-strong);
  line-height: 1.2;
}
.box-sm .box-title {
  font-size: 0.88rem;
}
.box-sub {
  margin: 0.16rem 0 0;
  font-size: 0.74rem;
  line-height: 1.25;
  color: var(--temporal-text);
}
.box-note {
  margin: 0.34rem 0 0;
  font-size: 0.7rem;
  font-style: italic;
  color: var(--temporal-text-muted);
}

.pair {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.5rem;
}
.box-db {
  /* Read as a cylinder without an image: elliptical caps on the side edges.
     Border stays neutral — mint is the "yours" colour and this is inside the
     Server. */
  border-radius: 1rem / 0.5rem;
  border-color: var(--temporal-lavender);
}

/* Connectors inside the Server zone. The paths stretch with the zone;
   non-scaling-stroke keeps them hairlines instead of smearing with the
   viewBox. */
.wire {
  display: block;
  width: 100%;
  stroke: var(--temporal-lavender);
  stroke-width: 1;
  fill: none;
}
.wire-fork {
  height: 1rem;
}
.wire-drop {
  height: 0.72rem;
}

/* Connectors between zones. Each is a rule with a solid head, so direction is
   unmistakable at the back of a room. */
.link {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.2rem;
  padding: 0 0.2rem;
}
.link-label {
  font-size: 0.68rem;
  line-height: 1.15;
  text-align: center;
  color: var(--temporal-text-muted);
}
.arrow {
  position: relative;
  display: block;
  width: 100%;
  height: 1px;
  background: var(--temporal-lavender);
}
.arrow::after {
  content: "";
  position: absolute;
  top: 50%;
  width: 0;
  height: 0;
  border-style: solid;
  transform: translateY(-50%);
}
.arrow-right::after {
  right: -1px;
  border-width: 0.24rem 0 0.24rem 0.36rem;
  border-color: transparent transparent transparent var(--temporal-lavender);
}
.arrow-left::after {
  left: -1px;
  border-width: 0.24rem 0.36rem 0.24rem 0;
  border-color: transparent var(--temporal-lavender) transparent transparent;
}
</style>
