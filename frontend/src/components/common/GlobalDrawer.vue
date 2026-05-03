<script setup lang="ts">
defineProps<{
  open: boolean
  title: string
  description?: string
}>()

const emit = defineEmits<{
  (event: 'close'): void
}>()
</script>

<template>
  <Teleport to="body">
    <div v-if="open" class="drawer-layer">
      <button class="drawer-mask" type="button" aria-label="Close drawer" @click="emit('close')"></button>
      <aside class="drawer" :aria-label="title">
        <header class="drawer-header">
          <div>
            <strong>{{ title }}</strong>
            <small v-if="description">{{ description }}</small>
          </div>
          <button type="button" @click="emit('close')">Close</button>
        </header>
        <div class="drawer-body">
          <slot />
        </div>
      </aside>
    </div>
  </Teleport>
</template>

<style scoped>
.drawer-layer {
  position: fixed;
  inset: 0;
  z-index: 60;
  display: flex;
  justify-content: flex-end;
}

.drawer-mask {
  position: absolute;
  inset: 0;
  border: 0;
  background: rgba(0, 0, 0, 0.46);
  cursor: default;
}

.drawer {
  position: relative;
  z-index: 1;
  display: flex;
  width: min(500px, 94vw);
  height: 100%;
  min-height: 0;
  flex-direction: column;
  border-left: 1px solid rgba(240, 207, 132, 0.16);
  color: var(--text);
  background: rgba(12, 14, 15, 0.94);
  box-shadow: -30px 0 80px rgba(0, 0, 0, 0.45);
  backdrop-filter: blur(24px);
}

.drawer-header {
  display: flex;
  flex: 0 0 auto;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  border-bottom: 1px solid var(--border);
  padding: 16px;
}

.drawer-header strong,
.drawer-header small {
  display: block;
}

.drawer-header strong {
  color: #fff7e8;
  font-size: 17px;
}

.drawer-header small {
  margin-top: 4px;
  color: var(--text-muted);
  font-size: 12px;
}

.drawer-header button {
  height: 34px;
  border: 1px solid rgba(240, 207, 132, 0.16);
  border-radius: 9px;
  color: var(--text-soft);
  background: rgba(255, 255, 255, 0.07);
  cursor: pointer;
  font-size: 13px;
  font-weight: 760;
}

.drawer-header button:hover {
  border-color: rgba(240, 207, 132, 0.3);
  color: var(--text);
}

.drawer-body {
  min-height: 0;
  flex: 1 1 auto;
  overflow-y: auto;
}
</style>
