<script setup lang="ts">
import { computed } from 'vue'

import { useAuthStore } from '../../stores/auth'

defineProps<{
  connected: boolean
  pendingRequests: number
}>()

const emit = defineEmits<{
  (event: 'open-search'): void
  (event: 'open-friends'): void
  (event: 'open-groups'): void
  (event: 'open-profile'): void
  (event: 'logout'): void
}>()

const auth = useAuthStore()
const displayName = computed(() => auth.user?.nickname || auth.user?.username || 'User')
const avatarText = computed(() => displayName.value.slice(0, 1).toUpperCase() || '#')
</script>

<template>
  <header class="app-top-bar">
    <div class="brand">
      <span class="brand-mark">M</span>
      <div>
        <strong>Mini IM</strong>
        <small>Realtime workspace</small>
      </div>
    </div>

    <button class="global-search" type="button" @click="emit('open-search')">
      <span>Search messages and files</span>
      <kbd>/</kbd>
    </button>

    <nav class="top-actions" aria-label="Global tools">
      <span :class="['connection-pill', connected ? 'online' : 'offline']">
        {{ connected ? 'Online' : 'Offline' }}
      </span>
      <button type="button" @click="emit('open-friends')">
        Friends
        <span v-if="pendingRequests > 0" class="counter">{{ pendingRequests }}</span>
      </button>
      <button type="button" @click="emit('open-groups')">Groups</button>
      <button class="profile-button" type="button" @click="emit('open-profile')">
        <span>{{ avatarText }}</span>
        {{ displayName }}
      </button>
      <button class="logout-button" type="button" @click="emit('logout')">Logout</button>
    </nav>
  </header>
</template>

<style scoped>
.app-top-bar {
  display: grid;
  min-height: 64px;
  grid-template-columns: minmax(176px, 240px) minmax(220px, 520px) minmax(0, 1fr);
  gap: 16px;
  align-items: center;
  padding: 10px 18px;
  border-bottom: 1px solid var(--border);
  background: rgba(11, 13, 14, 0.78);
  backdrop-filter: blur(18px);
}

.brand {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 10px;
}

.brand-mark {
  display: grid;
  width: 36px;
  height: 36px;
  flex: 0 0 36px;
  place-items: center;
  border: 1px solid rgba(240, 207, 132, 0.3);
  border-radius: 10px;
  color: #15100a;
  background: linear-gradient(135deg, #f0cf84, #af8435);
  font-weight: 900;
}

.brand strong,
.brand small {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.brand strong {
  color: #fff7e8;
  font-size: 15px;
}

.brand small {
  margin-top: 2px;
  color: var(--text-muted);
  font-size: 12px;
}

.global-search {
  display: flex;
  min-width: 0;
  height: 40px;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  border: 1px solid rgba(240, 207, 132, 0.14);
  border-radius: 10px;
  padding: 0 12px 0 14px;
  color: rgba(243, 239, 229, 0.68);
  background: rgba(255, 255, 255, 0.06);
  cursor: pointer;
  text-align: left;
}

.global-search:hover {
  border-color: rgba(240, 207, 132, 0.32);
  color: var(--text);
  background: rgba(255, 255, 255, 0.09);
}

.global-search span {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

kbd {
  display: grid;
  width: 24px;
  height: 24px;
  flex: 0 0 24px;
  place-items: center;
  border: 1px solid rgba(240, 207, 132, 0.18);
  border-radius: 6px;
  color: var(--accent-strong);
  background: rgba(0, 0, 0, 0.24);
  font-size: 12px;
}

.top-actions {
  display: flex;
  min-width: 0;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
}

.top-actions button,
.connection-pill {
  display: inline-flex;
  height: 36px;
  align-items: center;
  gap: 6px;
  border: 1px solid rgba(240, 207, 132, 0.14);
  border-radius: 9px;
  padding: 0 11px;
  color: var(--text-soft);
  background: rgba(255, 255, 255, 0.055);
  font-size: 13px;
  font-weight: 720;
}

.top-actions button {
  cursor: pointer;
}

.top-actions button:hover {
  border-color: rgba(240, 207, 132, 0.28);
  color: var(--text);
  background: rgba(255, 255, 255, 0.09);
}

.connection-pill::before {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  content: "";
  background: var(--danger);
}

.connection-pill.online::before {
  background: var(--success);
  box-shadow: 0 0 12px rgba(94, 226, 160, 0.48);
}

.counter {
  display: grid;
  min-width: 18px;
  height: 18px;
  place-items: center;
  border-radius: 999px;
  padding: 0 5px;
  color: #1b1008;
  background: var(--accent-strong);
  font-size: 11px;
}

.profile-button span {
  display: grid;
  width: 22px;
  height: 22px;
  place-items: center;
  border-radius: 50%;
  color: #16120a;
  background: var(--accent-strong);
  font-size: 12px;
}

.logout-button {
  color: #ffd6d6 !important;
  border-color: rgba(248, 113, 113, 0.22) !important;
}

@media (max-width: 960px) {
  .app-top-bar {
    grid-template-columns: 1fr;
  }

  .top-actions {
    justify-content: flex-start;
    overflow-x: auto;
    padding-bottom: 2px;
  }
}
</style>
