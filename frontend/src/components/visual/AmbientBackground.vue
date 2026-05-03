<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'

const x = ref(50)
const y = ref(35)

function handlePointerMove(event: PointerEvent) {
  x.value = (event.clientX / Math.max(window.innerWidth, 1)) * 100
  y.value = (event.clientY / Math.max(window.innerHeight, 1)) * 100
}

onMounted(() => {
  window.addEventListener('pointermove', handlePointerMove, { passive: true })
})

onBeforeUnmount(() => {
  window.removeEventListener('pointermove', handlePointerMove)
})
</script>

<template>
  <div
    class="ambient-background"
    :style="{ '--ambient-x': `${x}%`, '--ambient-y': `${y}%` }"
    aria-hidden="true"
  ></div>
</template>

<style scoped>
.ambient-background {
  position: absolute;
  inset: 0;
  overflow: hidden;
  pointer-events: none;
  background:
    radial-gradient(circle at var(--ambient-x) var(--ambient-y), rgba(215, 180, 106, 0.1), transparent 24rem),
    radial-gradient(circle at 12% 20%, rgba(148, 163, 184, 0.09), transparent 22rem),
    linear-gradient(135deg, #0c0e10 0%, #141719 48%, #0a0d0e 100%);
}

.ambient-background::before {
  position: absolute;
  inset: 0;
  content: "";
  opacity: 0.28;
  background-image:
    linear-gradient(rgba(255, 255, 255, 0.03) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.025) 1px, transparent 1px);
  background-size: 42px 42px;
  mask-image: radial-gradient(circle at var(--ambient-x) var(--ambient-y), black, transparent 72%);
}
</style>
