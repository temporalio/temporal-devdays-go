<!--
  ABOUTME: A section divider's title beside an animated Ziggy acting out the section.
  ABOUTME: One scene per divider; the art is decorative, so it is hidden from screen readers.

    ---
    layout: section
    ---

    <ZiggyHero scene="heartbeat">

    # Heartbeats

    One line of subtitle.

    </ZiggyHero>

  Ziggy is Temporal's tardigrade mascot: the official character art from the
  "Temporal in Space" blog post (temporal.io/blog/temporal-in-space) with its
  background removed. The props around it are emoji and CSS, so a new scene
  needs no new art.
-->
<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(defineProps<{ scene: string; src?: string }>(), {
  src: '/images/ziggy.png',
})

// Scenes that draw their own Ziggys instead of the single centred one.
const ownZiggys = ['children', 'parallel']
const showMain = computed(() => !ownZiggys.includes(props.scene))
</script>

<template>
  <div class="zh">
    <div class="zh-title"><slot /></div>
    <div class="stage" :class="`scene-${scene}`" aria-hidden="true">
      <div class="glow"></div>

      <!-- Business logic is complex: juggling everything around the happy path. -->
      <template v-if="scene === 'juggle'">
        <div class="juggle">
          <span class="ball b1">📄</span>
          <span class="ball b2">⚠️</span>
          <span class="ball b3">⏰</span>
        </div>
        <span class="prop sweat">💧</span>
      </template>

      <!-- Bank transfer: coins arc from one account to the other. -->
      <template v-else-if="scene === 'transfer'">
        <span class="prop acct-from">👛</span>
        <span class="prop acct-to">🏦</span>
        <span class="prop coin c1">🪙</span>
        <span class="prop coin c2">🪙</span>
        <span class="prop coin c3">🪙</span>
      </template>

      <!-- Demo: a live terminal. -->
      <template v-else-if="scene === 'demo'">
        <div class="term">
          <div class="term-bar"><i></i><i></i><i></i></div>
          <div class="term-line"><span class="typed">$ go run ./worker</span><span class="cursor">▌</span></div>
        </div>
        <span class="live"><i></i>LIVE</span>
      </template>

      <!-- Temporal concepts: replaying the Event History. -->
      <template v-else-if="scene === 'replay'">
        <span class="prop rewind">↺</span>
        <div class="events">
          <span class="ev e1">1 · Started</span>
          <span class="ev e2">2 · Scheduled</span>
          <span class="ev e3">3 · Completed</span>
        </div>
      </template>

      <!-- Workflows: Ziggy conducts the orchestra. -->
      <template v-else-if="scene === 'conduct'">
        <span class="baton"></span>
        <span class="prop note n1">♪</span>
        <span class="prop note n2">♫</span>
        <span class="prop note n3">♬</span>
      </template>

      <!-- Activities: the real work, with sparks. -->
      <template v-else-if="scene === 'work'">
        <span class="prop hammer">🔨</span>
        <span class="prop spark">💥</span>
        <span class="prop retry">↻ retry</span>
      </template>

      <!-- Workers: polling Tasks off a Task Queue. -->
      <template v-else-if="scene === 'poll'">
        <div class="queue">
          <span class="queue-label">Task Queue</span>
          <span class="task t1"></span>
          <span class="task t2"></span>
          <span class="task t3"></span>
        </div>
      </template>

      <!-- Signals: messages flying in. -->
      <template v-else-if="scene === 'signal'">
        <span class="prop env v1">✉️</span>
        <span class="prop env v2">✉️</span>
        <span class="prop env v3">✉️</span>
      </template>

      <!-- Queries: inspecting state. -->
      <template v-else-if="scene === 'query'">
        <div class="state-card"><code>status: "underwriting"</code><code>amount: 250000</code></div>
        <span class="prop lens">🔍</span>
      </template>

      <!-- Updates: a request goes in, an answer comes back. -->
      <template v-else-if="scene === 'update'">
        <span class="prop req">✉️</span>
        <span class="prop ack">✅</span>
      </template>

      <!-- Quiz: question marks pop. -->
      <template v-else-if="scene === 'quiz'">
        <span class="prop qm q1">❓</span>
        <span class="prop qm q2">❔</span>
        <span class="prop qm q3">❓</span>
      </template>

      <!-- Advanced topics: the wizard casts sparkles from the wand. -->
      <template v-else-if="scene === 'magic'">
        <span class="prop sparkle s1">✦</span>
        <span class="prop sparkle s2">✦</span>
        <span class="prop sparkle s3">✧</span>
        <span class="prop sparkle s4">✦</span>
      </template>

      <!-- Child Workflows: a parent with little ones following. -->
      <template v-else-if="scene === 'children'">
        <img :src="src" alt="" class="ziggy parent" />
        <img :src="src" alt="" class="ziggy kid k1" />
        <img :src="src" alt="" class="ziggy kid k2" />
      </template>

      <!-- Parallel Activities: two Ziggys, two progress bars, finishing together. -->
      <template v-else-if="scene === 'parallel'">
        <div v-for="lane in 2" :key="lane" class="lane" :class="`lane-${lane}`">
          <img :src="src" alt="" class="ziggy runner" />
          <div class="bar"><span></span></div>
          <span class="done">✓</span>
        </div>
      </template>

      <!-- Continue-as-New: an endless loop, shedding history pages. -->
      <template v-else-if="scene === 'forever'">
        <svg class="infinity" viewBox="0 0 200 90">
          <path class="inf-track" d="M100 45 C 70 0, 15 0, 15 45 C 15 90, 70 90, 100 45 C 130 0, 185 0, 185 45 C 185 90, 130 90, 100 45 Z" />
          <path class="inf-comet" pathLength="100" d="M100 45 C 70 0, 15 0, 15 45 C 15 90, 70 90, 100 45 C 130 0, 185 0, 185 45 C 185 90, 130 90, 100 45 Z" />
        </svg>
        <span class="prop page p1">📜</span>
        <span class="prop page p2">📜</span>
      </template>

      <!-- Heartbeats: a beating heart and an ECG trace. -->
      <template v-else-if="scene === 'heartbeat'">
        <span class="prop heart">❤️</span>
        <svg class="ecg" viewBox="0 0 300 50">
          <polyline pathLength="100" points="0,30 70,30 85,30 95,8 108,48 120,30 175,30 190,30 200,8 213,48 225,30 300,30" />
        </svg>
      </template>

      <!-- Saga: a failed step, and undo running back through the ones before it. -->
      <template v-else-if="scene === 'undo'">
        <div class="steps">
          <span class="step st1">✓ reserve</span>
          <span class="step st2">✓ charge</span>
          <span class="step st3">✕ ship</span>
        </div>
        <span class="prop undo-arrow">↶</span>
      </template>

      <!-- Timeouts and retries: a ticking stopwatch and a retry loop. -->
      <template v-else-if="scene === 'timeout'">
        <div class="watch"><span class="hand"></span></div>
        <span class="prop spin-retry">↻</span>
      </template>

      <!-- Error handling: failures bounce off a shield. -->
      <template v-else-if="scene === 'shield'">
        <span class="prop shield">🛡️</span>
        <span class="prop bolt f1">⚠️</span>
        <span class="prop bolt f2">💥</span>
      </template>

      <!-- Observability: a telescope and a live chart. -->
      <template v-else-if="scene === 'observe'">
        <span class="prop scope">🔭</span>
        <svg class="sparkline" viewBox="0 0 160 60">
          <polyline pathLength="100" points="0,50 25,40 45,44 70,22 95,30 120,10 160,18" />
        </svg>
      </template>

      <!-- Key SDK metrics: a gauge needle. -->
      <template v-else-if="scene === 'gauge'">
        <div class="gauge"><span class="needle"></span><span class="pivot"></span></div>
      </template>

      <!-- Cloud-side observability: floating among clouds. -->
      <template v-else-if="scene === 'cloud'">
        <span class="prop cl c1">☁️</span>
        <span class="prop cl c2">☁️</span>
        <span class="prop cl c3">☁️</span>
        <span class="prop chart">📈</span>
      </template>

      <!-- Observability integrations: pieces snapping together. -->
      <template v-else-if="scene === 'plug'">
        <span class="prop piece pc1">🧩</span>
        <span class="prop piece pc2">🧩</span>
        <span class="prop snap">✦</span>
      </template>

      <!-- Worker tuning: gears and a wrench. -->
      <template v-else-if="scene === 'tune'">
        <span class="prop gear g1">⚙️</span>
        <span class="prop gear g2">⚙️</span>
        <span class="prop wrench">🔧</span>
      </template>

      <!-- Worker resources: lifting a heavy load. -->
      <template v-else-if="scene === 'lift'">
        <div class="barbell"><i></i><b></b><i></i></div>
      </template>

      <!-- Memory, IO, Compute: the three resources orbit Ziggy. -->
      <template v-else-if="scene === 'trio'">
        <div class="trio-ring">
          <span class="chip r1">🧠 Memory</span>
          <span class="chip r2">⇅ IO</span>
          <span class="chip r3">⚡ Compute</span>
        </div>
      </template>

      <!-- More Features: a tilted orbit of little moons. -->
      <template v-else-if="scene === 'orbit'">
        <div class="orbit-ring">
          <span class="moon m1"></span>
          <span class="moon m2"></span>
          <span class="moon m3"></span>
        </div>
        <span class="prop twinkle tw1">✦</span>
        <span class="prop twinkle tw2">✦</span>
        <span class="prop twinkle tw3">✦</span>
      </template>

      <!-- Design patterns: drawing up the blueprint. -->
      <template v-else-if="scene === 'blueprint'">
        <div class="bp-sheet">
          <svg viewBox="0 0 150 100">
            <path pathLength="100" d="M15 85 L15 40 L75 12 L135 40 L135 85 Z M62 85 L62 58 L88 58 L88 85" />
          </svg>
        </div>
        <span class="prop pencil">✏️</span>
      </template>

      <!-- Start vs. Execute: one launches and moves on, the other waits for the answer. -->
      <template v-else-if="scene === 'launch'">
        <span class="prop rocket">🚀</span>
        <span class="prop smoke">💨</span>
        <span class="prop hourglass">⏳</span>
      </template>

      <!-- Versioning: v1 makes way for v2. -->
      <template v-else-if="scene === 'versions'">
        <span class="prop ver v-old">v1</span>
        <span class="prop ver v-new">v2</span>
        <span class="prop branch">🔀</span>
      </template>

      <!-- Serialization: data packed into a box. -->
      <template v-else-if="scene === 'pack'">
        <span class="prop brace br1">{ }</span>
        <span class="prop brace br2">[ ]</span>
        <span class="prop brace br3">"…"</span>
        <span class="prop box">📦</span>
      </template>

      <!-- Encryption: plain text in, ciphertext out, and a key that turns. -->
      <template v-else-if="scene === 'lock'">
        <div class="cipher">
          <code class="plain">amount: 500</code>
          <code class="scrambled">x9#Qa!7Lp</code>
        </div>
        <span class="prop padlock">🔒</span>
        <span class="prop key">🔑</span>
      </template>

      <!-- Claim-Check: the parcel stays parked, only the ticket travels. -->
      <template v-else-if="scene === 'claim'">
        <span class="prop parcel">🗄️</span>
        <span class="prop ticket">🎟️</span>
      </template>

      <!-- Nexus: two services, packets crossing the bridge between them. -->
      <template v-else-if="scene === 'bridge'">
        <span class="prop svc svc-a">🏢</span>
        <span class="prop svc svc-b">🏬</span>
        <span class="link"></span>
        <span class="packet pk1"></span>
        <span class="packet pk2"></span>
      </template>

      <!-- Patterns and anti-patterns: a verdict, then the better way. -->
      <template v-else-if="scene === 'dodont'">
        <span class="prop verdict bad">✕ anti-pattern</span>
        <span class="prop verdict good">✓ pattern</span>
        <span class="prop thumb">👍</span>
      </template>

      <img v-if="showMain" :src="src" alt="" class="ziggy main" />
    </div>
  </div>
</template>

<style>
/* The section layout caps its content at 22ch for a centred one-liner; this
   puts text and art side by side, so it needs the full width. */
.section-inner:has(.zh) {
  max-width: none !important;
}
</style>

<style scoped>
.zh {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 2.5rem;
  text-align: left;
}
.zh-title {
  flex: 0 1 26rem;
}
.zh-title :deep(h1) {
  max-width: 12ch;
}
.stage {
  position: relative;
  width: 360px;
  height: 330px;
  flex-shrink: 0;
}
.glow {
  position: absolute;
  inset: 40px;
  border-radius: 50%;
  background: radial-gradient(circle, rgb(154 128 255 / 32%), rgb(89 253 160 / 8%) 55%, transparent 70%);
}
.ziggy {
  position: absolute;
  filter: drop-shadow(0 16px 26px rgb(0 0 0 / 45%));
}
.main {
  left: 80px;
  top: 90px;
  width: 230px;
  z-index: 1;
  animation: bob 3.2s ease-in-out infinite;
}
.prop {
  position: absolute;
  line-height: 1;
  z-index: 2;
}

/* ---------- Ziggy body moves ---------- */
@keyframes bob {
  0%, 100% { transform: translateY(0) rotate(-3deg); }
  50% { transform: translateY(-12px) rotate(3deg); }
}
@keyframes bounce {
  0%, 100% { transform: translateY(0); }
  40% { transform: translateY(-22px) rotate(-4deg); }
  60% { transform: translateY(-22px) rotate(2deg); }
}
@keyframes sway {
  0%, 100% { transform: rotate(-7deg); }
  50% { transform: rotate(7deg); }
}
@keyframes nod {
  0%, 60%, 100% { transform: rotate(0deg); }
  70% { transform: rotate(-8deg) translateY(4px); }
}
@keyframes tilt {
  0%, 100% { transform: rotate(0deg); }
  30%, 60% { transform: rotate(-12deg); }
}
@keyframes wobble {
  0%, 100% { transform: rotate(-2deg); }
  25% { transform: rotate(3deg) translateX(3px); }
  75% { transform: rotate(-4deg) translateX(-3px); }
}
@keyframes squat {
  0%, 100% { transform: scaleY(1) translateY(0); }
  50% { transform: scaleY(0.9) translateY(12px); }
}
@keyframes twinkle {
  0%, 100% { opacity: 0.2; transform: scale(0.7); }
  50% { opacity: 1; transform: scale(1.15); }
}

/* ---------- juggle ---------- */
.scene-juggle .main { animation: wobble 1.2s ease-in-out infinite; }
.juggle {
  position: absolute;
  left: 10px;
  top: 0;
  width: 170px;
  height: 170px;
  z-index: 2;
  animation: spin 2.6s linear infinite;
}
.ball {
  position: absolute;
  font-size: 1.9rem;
  margin: -0.95rem 0 0 -0.95rem;
  animation: spin 2.6s linear infinite reverse;
}
.b1 { left: 50%; top: 0; }
.b2 { left: 93%; top: 75%; }
.b3 { left: 7%; top: 75%; }
.sweat {
  left: 230px;
  top: 100px;
  font-size: 1.3rem;
  animation: drip 1.6s ease-in infinite;
}
@keyframes spin { to { transform: rotate(360deg); } }
@keyframes drip {
  0% { opacity: 0; transform: translateY(-6px); }
  30% { opacity: 1; }
  100% { opacity: 0; transform: translateY(26px); }
}

/* ---------- transfer ---------- */
.scene-transfer .main { animation: bounce 1.8s ease-in-out infinite; top: 110px; }
.acct-from, .acct-to {
  top: 30px;
  font-size: 2.8rem;
}
.acct-from { left: 0; }
.acct-to { left: 290px; }
.coin {
  left: 20px;
  top: 30px;
  font-size: 1.6rem;
  animation: arc 2.4s ease-in-out infinite;
}
.c2 { animation-delay: 0.8s; }
.c3 { animation-delay: 1.6s; }
@keyframes arc {
  0% { opacity: 0; transform: translate(0, 0) rotate(0deg); }
  10% { opacity: 1; }
  50% { transform: translate(135px, -60px) rotate(180deg); }
  90% { opacity: 1; }
  100% { opacity: 0; transform: translate(275px, 0) rotate(360deg); }
}

/* ---------- demo ---------- */
.scene-demo .main { animation: bounce 1.4s ease-in-out infinite; }
.term {
  position: absolute;
  left: 0;
  top: 10px;
  width: 215px;
  border-radius: 8px;
  background: #0b1020;
  border: 1px solid var(--temporal-border-hi);
  z-index: 2;
  box-shadow: 0 10px 24px rgb(0 0 0 / 40%);
}
.term-bar {
  display: flex;
  gap: 5px;
  padding: 6px 8px;
  border-bottom: 1px solid var(--temporal-border-hi);
}
.term-bar i {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #ff6b8a;
}
.term-bar i:nth-child(2) { background: #ffd250; }
.term-bar i:nth-child(3) { background: #59fda0; }
.term-line {
  display: flex;
  padding: 8px 10px 10px;
  font-family: var(--slidev-code-font-family, monospace);
  font-size: 0.8rem;
  color: #59fda0;
}
.typed {
  display: inline-block;
  overflow: hidden;
  white-space: nowrap;
  width: 0;
  animation: type 3.6s steps(21) infinite;
}
.cursor { animation: blink 0.8s step-end infinite; }
.live {
  position: absolute;
  right: 6px;
  top: 6px;
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 3px 10px;
  border-radius: 999px;
  font-size: 0.75rem;
  font-weight: 600;
  letter-spacing: 0.1em;
  color: #fff;
  background: rgb(255 107 138 / 25%);
  border: 1px solid #ff6b8a;
}
.live i {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #ff6b8a;
  animation: blink 1s step-end infinite;
}
@keyframes type {
  0% { width: 0; }
  60%, 100% { width: 21ch; }
}
@keyframes blink { 50% { opacity: 0; } }

/* ---------- replay ---------- */
.rewind {
  left: 40px;
  top: 10px;
  font-size: 290px;
  color: rgb(154 128 255 / 22%);
  z-index: 0;
  animation: spin 6s linear infinite reverse;
}
.events {
  position: absolute;
  left: 0;
  top: 8px;
  display: flex;
  flex-direction: column;
  gap: 6px;
  z-index: 2;
}
.ev {
  padding: 3px 10px;
  border-radius: 6px;
  font-size: 0.72rem;
  background: var(--temporal-bg-elev);
  border: 1px solid var(--temporal-border-hi);
  color: var(--temporal-text);
  animation: light 3s ease-in-out infinite;
}
.e2 { animation-delay: 0.6s; }
.e3 { animation-delay: 1.2s; }
@keyframes light {
  0%, 15%, 60%, 100% { border-color: var(--temporal-border-hi); background: var(--temporal-bg-elev); }
  25%, 45% { border-color: #59fda0; background: rgb(89 253 160 / 18%); }
}

/* ---------- conduct ---------- */
.scene-conduct .main { animation: sway 2.4s ease-in-out infinite; transform-origin: 50% 90%; }
.baton {
  position: absolute;
  left: 40px;
  top: 110px;
  width: 70px;
  height: 4px;
  border-radius: 2px;
  background: linear-gradient(90deg, #fff, #ffd250);
  transform-origin: 100% 50%;
  z-index: 2;
  animation: wave-baton 1.2s ease-in-out infinite;
}
.note {
  font-size: 1.8rem;
  color: var(--temporal-lavender);
  animation: rise 2.4s ease-out infinite;
}
.n1 { left: 30px; top: 90px; }
.n2 { left: 70px; top: 70px; color: #59fda0; animation-delay: 0.8s; }
.n3 { left: 10px; top: 60px; color: #ffd250; animation-delay: 1.6s; }
@keyframes wave-baton {
  0%, 100% { transform: rotate(-25deg); }
  50% { transform: rotate(20deg); }
}
@keyframes rise {
  0% { opacity: 0; transform: translate(0, 0) scale(0.6); }
  20% { opacity: 1; }
  100% { opacity: 0; transform: translate(-20px, -80px) scale(1.1); }
}

/* ---------- work ---------- */
.scene-work .main { animation: nod 0.9s ease-in-out infinite; }
.hammer {
  left: 20px;
  top: 60px;
  font-size: 3.8rem;
  transform-origin: 80% 80%;
  animation: whack 0.9s ease-in infinite;
}
.spark {
  left: 6px;
  top: 150px;
  font-size: 1.8rem;
  animation: pop 0.9s ease-out infinite;
}
.retry {
  left: 210px;
  top: 30px;
  font-size: 0.85rem;
  padding: 2px 10px;
  border-radius: 999px;
  color: #59fda0;
  border: 1px solid #59fda0;
  animation: fade-cycle 3.6s ease-in-out infinite;
}
@keyframes whack {
  0%, 100% { transform: rotate(-40deg); }
  45% { transform: rotate(-40deg); }
  60% { transform: rotate(25deg); }
}
@keyframes pop {
  0%, 55% { opacity: 0; transform: scale(0.3); }
  65% { opacity: 1; transform: scale(1.2); }
  100% { opacity: 0; transform: scale(1); }
}
@keyframes fade-cycle {
  0%, 50%, 100% { opacity: 0; transform: translateY(6px); }
  65%, 85% { opacity: 1; transform: translateY(0); }
}

/* ---------- poll ---------- */
.scene-poll .main { animation: nod 1.5s ease-in-out infinite; }
.queue {
  position: absolute;
  left: -10px;
  top: 40px;
  width: 140px;
  height: 46px;
  border-radius: 8px;
  border: 1px dashed var(--temporal-border-hi);
  background: rgb(11 16 32 / 55%);
  z-index: 2;
  overflow: hidden;
}
.queue-label {
  position: absolute;
  left: 8px;
  top: -1px;
  font-size: 0.55rem;
  letter-spacing: 0.12em;
  text-transform: uppercase;
  color: var(--temporal-text);
}
.task {
  position: absolute;
  top: 18px;
  left: -30px;
  width: 26px;
  height: 18px;
  border-radius: 4px;
  background: var(--temporal-lavender);
  animation: slide-in 4.5s linear infinite;
}
.t2 { background: #59fda0; animation-delay: 1.5s; }
.t3 { background: #ffd250; animation-delay: 3s; }
@keyframes slide-in {
  0% { transform: translateX(0); opacity: 1; }
  80% { transform: translateX(150px); opacity: 1; }
  100% { transform: translateX(170px); opacity: 0; }
}

/* ---------- signal ---------- */
.scene-signal .main { animation: nod 1.2s ease-in-out infinite; }
.env {
  left: -20px;
  top: 40px;
  font-size: 1.9rem;
  animation: fly-in 3.6s ease-in infinite;
}
.v2 { top: 110px; animation-delay: 1.2s; }
.v3 { top: 70px; animation-delay: 2.4s; }
@keyframes fly-in {
  0% { opacity: 0; transform: translate(0, 0) rotate(-15deg); }
  15% { opacity: 1; }
  70% { opacity: 1; transform: translate(120px, 40px) rotate(10deg) scale(0.9); }
  85%, 100% { opacity: 0; transform: translate(140px, 50px) scale(0.3); }
}

/* ---------- query ---------- */
.scene-query .main { animation: tilt 3s ease-in-out infinite; }
.state-card {
  position: absolute;
  left: 0;
  top: 20px;
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 8px 12px;
  border-radius: 8px;
  background: var(--temporal-bg-elev);
  border: 1px solid var(--temporal-border-hi);
  z-index: 2;
}
.state-card code {
  font-size: 0.72rem;
  background: none;
  color: var(--temporal-text-strong);
}
.lens {
  left: 0;
  top: 18px;
  font-size: 2.6rem;
  z-index: 3;
  animation: scan 3s ease-in-out infinite;
}
@keyframes scan {
  0%, 100% { transform: translate(0, 0); }
  50% { transform: translate(110px, 20px); }
}

/* ---------- update ---------- */
.scene-update .main { animation: nod 3s ease-in-out infinite; }
.req, .ack {
  left: -10px;
  top: 70px;
  font-size: 1.9rem;
}
.req { animation: go-in 3s ease-in-out infinite; }
.ack { animation: come-back 3s ease-in-out infinite; }
@keyframes go-in {
  0% { opacity: 0; transform: translateX(0); }
  10% { opacity: 1; }
  40% { opacity: 1; transform: translateX(110px) scale(0.9); }
  48%, 100% { opacity: 0; transform: translateX(120px) scale(0.4); }
}
@keyframes come-back {
  0%, 50% { opacity: 0; transform: translate(110px, 40px) scale(0.4); }
  60% { opacity: 1; transform: translate(100px, 40px) scale(1); }
  90% { opacity: 1; transform: translate(0, 40px) scale(1.1); }
  100% { opacity: 0; transform: translate(-10px, 40px); }
}

/* ---------- quiz ---------- */
.scene-quiz .main { animation: tilt 2.4s ease-in-out infinite; }
.qm {
  font-size: 2.2rem;
  animation: pop-up 2.4s ease-out infinite;
}
.q1 { left: 40px; top: 50px; }
.q2 { left: 110px; top: 10px; animation-delay: 0.8s; }
.q3 { left: 190px; top: 40px; animation-delay: 1.6s; }
@keyframes pop-up {
  0% { opacity: 0; transform: scale(0.2) translateY(20px); }
  25% { opacity: 1; transform: scale(1.2) translateY(0); }
  40% { transform: scale(1); }
  80% { opacity: 1; }
  100% { opacity: 0; transform: translateY(-20px); }
}

/* ---------- magic ---------- */
.scene-magic .main { width: 250px; left: 70px; top: 70px; }
.sparkle {
  font-size: 1.4rem;
  color: #ffd250;
  animation: burst 2s ease-out infinite;
}
.s1 { left: 82px; top: 178px; }
.s2 { left: 78px; top: 186px; color: #59fda0; animation-delay: 0.5s; }
.s3 { left: 86px; top: 180px; color: var(--temporal-lavender); animation-delay: 1s; }
.s4 { left: 80px; top: 182px; animation-delay: 1.5s; }
@keyframes burst {
  0% { opacity: 0; transform: translate(0, 0) scale(0.4); }
  20% { opacity: 1; }
  100% { opacity: 0; transform: translate(var(--dx, -60px), var(--dy, -70px)) scale(1.2) rotate(90deg); }
}
.s2 { --dx: -80px; --dy: -20px; }
.s3 { --dx: -30px; --dy: -100px; }
.s4 { --dx: -90px; --dy: -60px; }

/* ---------- children ---------- */
.parent {
  left: 130px;
  top: 40px;
  width: 210px;
  animation: bob 3.2s ease-in-out infinite;
}
.kid {
  width: 90px;
  animation: hop 1s ease-in-out infinite;
}
.k1 { left: 50px; top: 220px; }
.k2 { left: -10px; top: 240px; width: 75px; animation-delay: 0.3s; }
@keyframes hop {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-22px) rotate(-6deg); }
}

/* ---------- parallel ---------- */
.lane {
  position: absolute;
  left: 10px;
  width: 340px;
  height: 120px;
}
.lane-1 { top: 20px; }
.lane-2 { top: 175px; }
.runner {
  left: 0;
  top: 0;
  width: 110px;
  animation: hop 0.8s ease-in-out infinite;
}
.bar {
  position: absolute;
  left: 125px;
  top: 50px;
  width: 170px;
  height: 12px;
  border-radius: 6px;
  background: rgb(255 255 255 / 10%);
  overflow: hidden;
}
.bar span {
  display: block;
  height: 100%;
  width: 0;
  background: linear-gradient(90deg, var(--temporal-lavender), #59fda0);
  animation: fill 3s ease-in-out infinite;
}
.done {
  position: absolute;
  left: 302px;
  top: 40px;
  font-size: 1.4rem;
  color: #59fda0;
  animation: done 3s ease-in-out infinite;
}
@keyframes fill {
  0% { width: 0; }
  70%, 100% { width: 100%; }
}
@keyframes done {
  0%, 68% { opacity: 0; transform: scale(0.4); }
  78%, 100% { opacity: 1; transform: scale(1); }
}

/* ---------- forever ---------- */
.scene-forever .main { top: 110px; width: 210px; left: 90px; }
.infinity {
  position: absolute;
  left: 20px;
  top: 0;
  width: 320px;
  z-index: 2;
  fill: none;
}
.inf-track { stroke: rgb(154 128 255 / 35%); stroke-width: 4; }
.inf-comet {
  stroke: #59fda0;
  stroke-width: 6;
  stroke-linecap: round;
  stroke-dasharray: 14 86;
  filter: drop-shadow(0 0 6px #59fda0);
  animation: run 3s linear infinite;
}
.page {
  left: 300px;
  top: 150px;
  font-size: 1.5rem;
  animation: shed 3s ease-in infinite;
}
.p2 { animation-delay: 1.5s; }
@keyframes run { to { stroke-dashoffset: -100; } }
@keyframes shed {
  0% { opacity: 0; transform: translate(0, 0) rotate(0deg); }
  20% { opacity: 1; }
  100% { opacity: 0; transform: translate(30px, 140px) rotate(60deg); }
}

/* ---------- heartbeat ---------- */
.scene-heartbeat .main { animation: thump 1.2s ease-in-out infinite; }
.heart {
  left: 40px;
  top: 50px;
  font-size: 2.8rem;
  animation: beat 1.2s ease-in-out infinite;
}
.ecg {
  position: absolute;
  left: 0;
  bottom: 0;
  width: 360px;
  z-index: 2;
  fill: none;
  stroke: #ff6b8a;
  stroke-width: 3;
  stroke-linejoin: round;
  stroke-dasharray: 100;
  filter: drop-shadow(0 0 4px #ff6b8a);
  animation: trace 2.4s linear infinite;
}
@keyframes beat {
  0%, 40%, 100% { transform: scale(1); }
  10%, 30% { transform: scale(1.3); }
  20% { transform: scale(1.05); }
}
@keyframes thump {
  0%, 40%, 100% { transform: scale(1); }
  10%, 30% { transform: scale(1.03); }
}
@keyframes trace {
  0% { stroke-dashoffset: 100; }
  100% { stroke-dashoffset: -100; }
}

/* ---------- undo ---------- */
.steps {
  position: absolute;
  left: 0;
  top: 20px;
  display: flex;
  flex-direction: column;
  gap: 6px;
  z-index: 2;
}
.step {
  padding: 3px 10px;
  border-radius: 6px;
  font-size: 0.75rem;
  border: 1px solid #59fda0;
  color: #59fda0;
  background: var(--temporal-bg-elev);
}
.st3 {
  border-color: #ff6b8a;
  color: #ff6b8a;
  animation: shake 3s ease-in-out infinite;
}
.st2 { animation: undo 3s ease-in-out infinite; animation-delay: 0.6s; }
.st1 { animation: undo 3s ease-in-out infinite; animation-delay: 1.2s; }
.undo-arrow {
  left: 120px;
  top: 0;
  font-size: 3.4rem;
  color: #ffd250;
  animation: rewind 3s ease-in-out infinite;
}
@keyframes shake {
  0%, 20%, 100% { transform: translateX(0); }
  5%, 15% { transform: translateX(-4px); }
  10% { transform: translateX(4px); }
}
@keyframes undo {
  0%, 25%, 100% { border-color: #59fda0; color: #59fda0; text-decoration: none; }
  35%, 70% { border-color: #ffd250; color: #ffd250; text-decoration: line-through; }
}
@keyframes rewind {
  0%, 100% { transform: rotate(0deg); }
  60% { transform: rotate(-360deg); }
}

/* ---------- timeout ---------- */
.scene-timeout .main { animation: wobble 0.8s ease-in-out infinite; }
.watch {
  position: absolute;
  left: 20px;
  top: 30px;
  width: 80px;
  height: 80px;
  border-radius: 50%;
  border: 4px solid #ffd250;
  background: #0b1020;
  z-index: 2;
}
.watch::before {
  content: '';
  position: absolute;
  left: 50%;
  top: -14px;
  width: 14px;
  height: 8px;
  margin-left: -7px;
  border-radius: 2px;
  background: #ffd250;
}
.hand {
  position: absolute;
  left: 50%;
  bottom: 50%;
  width: 3px;
  height: 32px;
  margin-left: -1.5px;
  border-radius: 2px;
  background: #ff6b8a;
  transform-origin: 50% 100%;
  animation: spin 3s steps(12) infinite;
}
.spin-retry {
  left: 110px;
  top: 20px;
  font-size: 2.4rem;
  color: #59fda0;
  animation: spin 1.5s linear infinite;
}

/* ---------- shield ---------- */
.shield {
  left: 50px;
  top: 110px;
  font-size: 3.2rem;
  z-index: 3;
  animation: brace 1.8s ease-in-out infinite;
}
.bolt {
  left: -30px;
  top: 110px;
  font-size: 1.8rem;
  animation: deflect 1.8s ease-in infinite;
}
.f2 { top: 140px; animation-delay: 0.9s; }
@keyframes deflect {
  0% { opacity: 0; transform: translate(0, 0); }
  15% { opacity: 1; }
  45% { transform: translate(70px, 0); }
  100% { opacity: 0; transform: translate(-10px, -90px) rotate(-120deg); }
}
@keyframes brace {
  0%, 35%, 100% { transform: translateX(0); }
  45% { transform: translateX(6px) scale(1.08); }
}

/* ---------- observe ---------- */
.scope {
  left: 20px;
  top: 90px;
  font-size: 3rem;
  transform-origin: 70% 70%;
  animation: look 3s ease-in-out infinite;
}
.sparkline {
  position: absolute;
  left: 0;
  top: 10px;
  width: 190px;
  z-index: 2;
  fill: none;
  stroke: #59fda0;
  stroke-width: 3;
  stroke-linejoin: round;
  stroke-dasharray: 100;
  filter: drop-shadow(0 0 4px #59fda0);
  animation: draw 3s ease-in-out infinite;
}
@keyframes look {
  0%, 100% { transform: rotate(0deg); }
  50% { transform: rotate(-14deg); }
}
@keyframes draw {
  0% { stroke-dashoffset: 100; }
  60%, 100% { stroke-dashoffset: 0; }
}

/* ---------- gauge ---------- */
.gauge {
  position: absolute;
  left: 10px;
  top: 20px;
  width: 140px;
  height: 70px;
  border-radius: 70px 70px 0 0;
  background: conic-gradient(from 270deg at 50% 100%, #59fda0 0deg, #ffd250 110deg, #ff6b8a 180deg, transparent 180deg);
  z-index: 2;
}
.gauge::after {
  content: '';
  position: absolute;
  left: 16px;
  right: 16px;
  bottom: 0;
  height: 54px;
  border-radius: 54px 54px 0 0;
  background: #0b1020;
}
.needle {
  position: absolute;
  left: 50%;
  bottom: 0;
  width: 4px;
  height: 58px;
  margin-left: -2px;
  border-radius: 2px;
  background: #fff;
  transform-origin: 50% 100%;
  z-index: 1;
  animation: swing 2.4s ease-in-out infinite;
}
.pivot {
  position: absolute;
  left: 50%;
  bottom: -7px;
  width: 14px;
  height: 14px;
  margin-left: -7px;
  border-radius: 50%;
  background: #fff;
  z-index: 1;
}
@keyframes swing {
  0%, 100% { transform: rotate(-60deg); }
  40% { transform: rotate(45deg); }
  60% { transform: rotate(30deg); }
}

/* ---------- cloud ---------- */
.cl {
  font-size: 2.6rem;
  animation: drift 7s linear infinite;
}
.c1 { left: -20px; top: 20px; }
.c2 { left: -20px; top: 150px; font-size: 1.8rem; animation-delay: -3s; animation-duration: 9s; }
.c3 { left: -20px; top: 260px; animation-delay: -5s; }
.chart {
  left: 250px;
  top: 40px;
  font-size: 1.8rem;
  animation: twinkle 2s ease-in-out infinite;
}
@keyframes drift {
  0% { opacity: 0; transform: translateX(0); }
  15%, 85% { opacity: 1; }
  100% { opacity: 0; transform: translateX(340px); }
}

/* ---------- plug ---------- */
.scene-plug .main { animation: bounce 2.4s ease-in-out infinite; }
.piece {
  top: 40px;
  font-size: 2.4rem;
}
.pc1 { left: 0; animation: join-left 2.4s ease-in-out infinite; }
.pc2 { left: 120px; animation: join-right 2.4s ease-in-out infinite; }
.snap {
  left: 75px;
  top: 30px;
  font-size: 1.8rem;
  color: #ffd250;
  animation: flash 2.4s ease-out infinite;
}
@keyframes join-left {
  0%, 100% { transform: translateX(0) rotate(-10deg); }
  45%, 70% { transform: translateX(40px) rotate(0deg); }
}
@keyframes join-right {
  0%, 100% { transform: translateX(0) rotate(10deg) scaleX(-1); }
  45%, 70% { transform: translateX(-38px) rotate(0deg) scaleX(-1); }
}
@keyframes flash {
  0%, 40% { opacity: 0; transform: scale(0.3); }
  50% { opacity: 1; transform: scale(1.4); }
  70%, 100% { opacity: 0; transform: scale(1); }
}

/* ---------- tune ---------- */
.scene-tune .main { animation: nod 1.6s ease-in-out infinite; }
.gear { font-size: 2.6rem; animation: spin 4s linear infinite; }
.g1 { left: 20px; top: 30px; }
.g2 { left: 64px; top: 64px; font-size: 1.8rem; animation-direction: reverse; animation-duration: 2.8s; }
.wrench {
  left: 10px;
  top: 140px;
  font-size: 2.4rem;
  transform-origin: 80% 20%;
  animation: twist 1.6s ease-in-out infinite;
}
@keyframes twist {
  0%, 100% { transform: rotate(0deg); }
  50% { transform: rotate(-45deg); }
}

/* ---------- lift ---------- */
.scene-lift .main { animation: squat 1.6s ease-in-out infinite; transform-origin: 50% 100%; }
.barbell {
  position: absolute;
  left: 30px;
  top: 60px;
  display: flex;
  align-items: center;
  z-index: 2;
  animation: press 1.6s ease-in-out infinite;
}
.barbell b {
  width: 170px;
  height: 6px;
  border-radius: 3px;
  background: #cfd6ee;
}
.barbell i {
  width: 18px;
  height: 52px;
  border-radius: 4px;
  background: var(--temporal-lavender);
  box-shadow: 0 0 10px rgb(154 128 255 / 60%);
}
@keyframes press {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(18px); }
}

/* ---------- trio ---------- */
.trio-ring {
  position: absolute;
  inset: 20px;
  z-index: 2;
  animation: spin 12s linear infinite;
}
.chip {
  position: absolute;
  white-space: nowrap;
  padding: 3px 10px;
  border-radius: 999px;
  font-size: 0.8rem;
  background: var(--temporal-bg-elev);
  border: 1px solid var(--temporal-border-hi);
  color: var(--temporal-text-strong);
  translate: -50% -50%;
  animation: spin 12s linear infinite reverse;
}
.r1 { left: 50%; top: 0; }
.r2 { left: 93%; top: 75%; }
.r3 { left: 7%; top: 75%; }

/* ---------- orbit ---------- */
.scene-orbit .main { width: 260px; left: 50px; top: 50px; }
/* Not named `.ring`: that is a UnoCSS utility that draws a box-shadow. */
.orbit-ring {
  position: absolute;
  inset: 18px;
  border-radius: 50%;
  border: 1px dashed rgb(154 128 255 / 35%);
  transform: rotateX(68deg);
  animation: orbit 9s linear infinite;
}
.moon {
  position: absolute;
  width: 14px;
  height: 14px;
  margin: -7px 0 0 -7px;
  border-radius: 50%;
}
.m1 { top: 0; left: 50%; background: #59fda0; box-shadow: 0 0 12px #59fda0; }
.m2 { top: 75%; left: 93%; background: var(--temporal-lavender); box-shadow: 0 0 12px var(--temporal-lavender); }
.m3 { top: 75%; left: 7%; background: #ffd250; box-shadow: 0 0 12px #ffd250; }
.twinkle { color: #ffd250; animation: twinkle 1.8s ease-in-out infinite; }
.tw1 { top: 8%; left: 18%; font-size: 1.1rem; }
.tw2 { top: 14%; left: 82%; font-size: 0.75rem; animation-delay: 0.5s; color: #59fda0; }
.tw3 { top: 88%; left: 12%; font-size: 0.7rem; animation-delay: 1.4s; color: var(--temporal-lavender); }
@keyframes orbit { to { transform: rotateX(68deg) rotateZ(360deg); } }

/* ---------- blueprint ---------- */
.scene-blueprint .main { animation: nod 1.8s ease-in-out infinite; }
.bp-sheet {
  position: absolute;
  left: 0;
  top: 10px;
  width: 160px;
  height: 110px;
  border-radius: 6px;
  border: 1px solid #7fb2ff;
  background:
    repeating-linear-gradient(0deg, rgb(127 178 255 / 18%) 0 1px, transparent 1px 14px),
    repeating-linear-gradient(90deg, rgb(127 178 255 / 18%) 0 1px, transparent 1px 14px),
    #0e2a5c;
  z-index: 2;
}
.bp-sheet svg {
  width: 100%;
  height: 100%;
  fill: none;
  stroke: #e6f0ff;
  stroke-width: 2.5;
  stroke-linejoin: round;
  stroke-dasharray: 100;
  animation: draw 3.6s ease-in-out infinite;
}
.pencil {
  left: 10px;
  top: 80px;
  font-size: 1.8rem;
  z-index: 3;
  animation: sketch 3.6s ease-in-out infinite;
}
@keyframes sketch {
  0% { transform: translate(0, 0); }
  20% { transform: translate(55px, -60px); }
  40% { transform: translate(115px, -30px); }
  60% { transform: translate(115px, 20px); }
  80%, 100% { transform: translate(0, 0); }
}

/* ---------- launch ---------- */
.scene-launch .main { animation: tilt 3s ease-in-out infinite; }
.rocket {
  left: 20px;
  top: 200px;
  font-size: 2.8rem;
  animation: liftoff 3s ease-in infinite;
}
.smoke {
  left: 14px;
  top: 250px;
  font-size: 1.6rem;
  animation: puff 3s ease-out infinite;
}
.hourglass {
  left: 280px;
  top: 30px;
  font-size: 2.4rem;
  animation: flip 3s ease-in-out infinite;
}
@keyframes liftoff {
  0%, 15% { opacity: 1; transform: translate(0, 0); }
  70% { opacity: 1; transform: translate(40px, -210px) rotate(10deg); }
  80%, 100% { opacity: 0; transform: translate(50px, -240px); }
}
@keyframes puff {
  0%, 15% { opacity: 0; transform: scale(0.4); }
  30% { opacity: 0.9; transform: scale(1.2); }
  60%, 100% { opacity: 0; transform: scale(1.6); }
}
@keyframes flip {
  0%, 40% { transform: rotate(0deg); }
  55%, 90% { transform: rotate(180deg); }
  100% { transform: rotate(360deg); }
}

/* ---------- versions ---------- */
.scene-versions .main { animation: bob 3.2s ease-in-out infinite; }
.ver {
  left: 20px;
  top: 40px;
  padding: 4px 14px;
  border-radius: 999px;
  font-size: 1.3rem;
  font-weight: 600;
  background: var(--temporal-bg-elev);
}
.v-old {
  color: var(--temporal-lavender);
  border: 2px solid var(--temporal-lavender);
  animation: step-out 3.6s ease-in-out infinite;
}
.v-new {
  color: #59fda0;
  border: 2px solid #59fda0;
  animation: step-in 3.6s ease-in-out infinite;
}
.branch {
  left: 100px;
  top: 10px;
  font-size: 1.8rem;
  animation: twinkle 1.8s ease-in-out infinite;
}
@keyframes step-out {
  0%, 30% { opacity: 1; transform: translate(0, 0); }
  50%, 100% { opacity: 0; transform: translate(-30px, 40px) scale(0.7); }
}
@keyframes step-in {
  0%, 35% { opacity: 0; transform: translate(40px, -30px) scale(0.7); }
  55%, 90% { opacity: 1; transform: translate(0, 0) scale(1); }
  100% { opacity: 0; }
}

/* ---------- pack ---------- */
.scene-pack .main { animation: nod 1.2s ease-in-out infinite; }
.box {
  left: 20px;
  top: 170px;
  font-size: 3.4rem;
  animation: brace 1.2s ease-in-out infinite;
}
.brace {
  left: 30px;
  top: 0;
  font-family: var(--slidev-code-font-family, monospace);
  font-size: 1.3rem;
  font-weight: 600;
  color: var(--temporal-lavender);
  animation: drop-in 3.6s ease-in infinite;
}
.br2 { left: 70px; color: #59fda0; animation-delay: 1.2s; }
.br3 { left: 10px; color: #ffd250; animation-delay: 2.4s; }
@keyframes drop-in {
  0% { opacity: 0; transform: translateY(0) rotate(-20deg); }
  15% { opacity: 1; }
  70% { opacity: 1; transform: translateY(170px) rotate(10deg) scale(0.8); }
  85%, 100% { opacity: 0; transform: translateY(185px) scale(0.3); }
}

/* ---------- lock ---------- */
.scene-lock .main { animation: bob 3.2s ease-in-out infinite; }
.cipher {
  position: absolute;
  left: 0;
  top: 20px;
  width: 140px;
  height: 34px;
  border-radius: 8px;
  background: var(--temporal-bg-elev);
  border: 1px solid var(--temporal-border-hi);
  z-index: 2;
}
.cipher code {
  position: absolute;
  left: 12px;
  top: 7px;
  font-size: 0.8rem;
  background: none;
}
.plain { color: var(--temporal-text-strong); animation: show-first 3.6s ease-in-out infinite; }
.scrambled { color: #59fda0; animation: show-second 3.6s ease-in-out infinite; }
.padlock {
  left: 40px;
  top: 120px;
  font-size: 3rem;
  animation: brace 3.6s ease-in-out infinite;
}
.key {
  left: -20px;
  top: 135px;
  font-size: 2rem;
  transform-origin: 20% 50%;
  animation: turn-key 3.6s ease-in-out infinite;
}
@keyframes show-first {
  0%, 40% { opacity: 1; }
  50%, 90% { opacity: 0; }
  100% { opacity: 1; }
}
@keyframes show-second {
  0%, 40% { opacity: 0; }
  50%, 90% { opacity: 1; }
  100% { opacity: 0; }
}
@keyframes turn-key {
  0% { opacity: 0; transform: translateX(0); }
  15% { opacity: 1; }
  35% { transform: translateX(40px) rotate(0deg); }
  45%, 75% { transform: translateX(40px) rotate(90deg); }
  90%, 100% { opacity: 0; transform: translateX(0); }
}

/* ---------- claim ---------- */
.scene-claim .main { animation: nod 2.4s ease-in-out infinite; }
.parcel {
  left: 0;
  top: 190px;
  font-size: 3.6rem;
}
.ticket {
  left: 20px;
  top: 170px;
  font-size: 2rem;
  animation: hand-over 2.4s ease-in-out infinite;
}
@keyframes hand-over {
  0% { opacity: 0; transform: translate(0, 0) rotate(-10deg); }
  15% { opacity: 1; }
  60% { opacity: 1; transform: translate(130px, -100px) rotate(15deg); }
  75%, 100% { opacity: 0; transform: translate(140px, -110px) scale(0.5); }
}

/* ---------- bridge ---------- */
.scene-bridge .main { animation: bounce 2.4s ease-in-out infinite; top: 120px; }
.svc {
  top: 20px;
  font-size: 2.8rem;
}
.svc-a { left: 0; }
.svc-b { left: 300px; }
.link {
  position: absolute;
  left: 55px;
  right: 65px;
  top: 52px;
  border-top: 2px dashed rgb(154 128 255 / 55%);
  z-index: 1;
}
.packet {
  position: absolute;
  left: 55px;
  top: 46px;
  width: 12px;
  height: 12px;
  border-radius: 50%;
  background: #59fda0;
  box-shadow: 0 0 10px #59fda0;
  z-index: 2;
  animation: cross 2.4s linear infinite;
}
.pk2 { background: #ffd250; box-shadow: 0 0 10px #ffd250; animation-direction: reverse; }
@keyframes cross {
  0% { opacity: 0; transform: translateX(0); }
  10%, 90% { opacity: 1; }
  100% { opacity: 0; transform: translateX(230px); }
}

/* ---------- dodont ---------- */
.scene-dodont .main { animation: tilt 3.6s ease-in-out infinite; }
.verdict {
  left: 0;
  top: 40px;
  padding: 4px 12px;
  border-radius: 8px;
  font-size: 0.95rem;
  font-weight: 600;
  white-space: nowrap;
  background: var(--temporal-bg-elev);
}
.bad {
  color: #ff6b8a;
  border: 2px solid #ff6b8a;
  animation: show-first 3.6s ease-in-out infinite, shake 3.6s ease-in-out infinite;
}
.good {
  color: #59fda0;
  border: 2px solid #59fda0;
  animation: show-second 3.6s ease-in-out infinite;
}
.thumb {
  left: 250px;
  top: 40px;
  font-size: 2.2rem;
  animation: thumbs 3.6s ease-out infinite;
}
@keyframes thumbs {
  0%, 45% { opacity: 0; transform: scale(0.3) rotate(0deg); }
  55% { opacity: 1; transform: scale(1.3) rotate(-15deg); }
  65%, 90% { opacity: 1; transform: scale(1) rotate(0deg); }
  100% { opacity: 0; }
}

@media (prefers-reduced-motion: reduce) {
  .stage *,
  .stage *::before,
  .stage *::after {
    animation: none !important;
  }
}
</style>
