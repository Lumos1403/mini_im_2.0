<script setup lang="ts">
import AmbientBackground from '../components/visual/AmbientBackground.vue'
</script>

<template>
  <main class="app-shell">
    <AmbientBackground />
    <div class="shell-surface">
      <slot />
    </div>
  </main>
</template>

<style scoped>
.app-shell {
  position: relative;
  min-height: 100vh;
  min-height: 100dvh;
  overflow: hidden;
  color: var(--text);
}

.shell-surface {
  position: relative;
  z-index: 1;
  height: 100vh;
  height: 100dvh;
  min-height: 0;
}
</style>
