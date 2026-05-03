<script setup lang="ts">
import MagneticParticleField from '../components/visual/MagneticParticleField.vue'
</script>

<template>
  <main class="auth-layout">
    <MagneticParticleField />
    <div class="auth-vignette"></div>
    <section class="auth-copy" aria-label="Mini IM identity">
      <p>Mini IM</p>
      <h1>Private conversations with a magnetic pulse.</h1>
      <span>Secure access, realtime presence, and focused messaging in one controlled workspace.</span>
    </section>
    <slot />
  </main>
</template>

<style scoped>
.auth-layout {
  position: relative;
  display: grid;
  min-height: 100vh;
  min-height: 100dvh;
  grid-template-columns: minmax(0, 1fr) minmax(360px, 460px);
  gap: clamp(28px, 5vw, 72px);
  align-items: center;
  overflow: hidden;
  padding: clamp(24px, 6vw, 72px);
  background:
    linear-gradient(135deg, rgba(8, 10, 12, 0.92), rgba(17, 19, 20, 0.82)),
    #080a0c;
}

.auth-vignette {
  position: absolute;
  inset: 0;
  pointer-events: none;
  background:
    radial-gradient(circle at 72% 42%, rgba(240, 207, 132, 0.14), transparent 24rem),
    linear-gradient(90deg, rgba(0, 0, 0, 0.12), rgba(0, 0, 0, 0.55));
}

.auth-copy {
  position: relative;
  z-index: 1;
  max-width: 720px;
  color: var(--text);
}

.auth-copy p {
  margin: 0 0 22px;
  color: var(--accent-strong);
  font-size: 13px;
  font-weight: 800;
  letter-spacing: 0.18em;
  text-transform: uppercase;
}

.auth-copy h1 {
  max-width: 12ch;
  margin: 0;
  color: #fff7e8;
  font-size: clamp(54px, 8vw, 116px);
  font-weight: 760;
  line-height: 0.9;
}

.auth-copy span {
  display: block;
  max-width: 560px;
  margin-top: 28px;
  color: rgba(243, 239, 229, 0.72);
  font-size: clamp(16px, 2vw, 20px);
  line-height: 1.65;
}

@media (max-width: 860px) {
  .auth-layout {
    grid-template-columns: 1fr;
    align-items: end;
    padding: 24px;
  }

  .auth-copy {
    max-width: 100%;
  }

  .auth-copy h1 {
    max-width: 10ch;
    font-size: clamp(42px, 14vw, 72px);
  }
}

@media (max-width: 560px) {
  .auth-layout {
    gap: 22px;
    padding: 18px;
  }

  .auth-copy span {
    display: none;
  }
}
</style>
