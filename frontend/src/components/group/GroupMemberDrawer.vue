<script setup lang="ts">
import type { GroupMember } from '../../api/group'
import GroupMemberList from './GroupMemberList.vue'

defineProps<{
  open: boolean
  title?: string
  members: GroupMember[]
  loading?: boolean
  selectedUserId?: string
}>()

const emit = defineEmits<{
  (event: 'close'): void
  (event: 'refresh'): void
  (event: 'select', member: GroupMember): void
}>()
</script>

<template>
  <Teleport to="body">
    <div v-if="open" class="drawer-layer">
      <button class="drawer-mask" type="button" aria-label="关闭群成员列表" @click="emit('close')"></button>
      <aside class="member-drawer" aria-label="群成员列表">
        <header class="drawer-header">
          <div class="drawer-title">
            <strong>群成员</strong>
            <small>{{ title || '当前群聊' }} · {{ members.length }} 人</small>
          </div>
          <div class="drawer-actions">
            <button type="button" :disabled="loading" @click="emit('refresh')">
              {{ loading ? '刷新中' : '刷新' }}
            </button>
            <button type="button" @click="emit('close')">关闭</button>
          </div>
        </header>

        <div class="drawer-body">
          <GroupMemberList
            :members="members"
            :loading="loading"
            :selected-user-id="selectedUserId"
            @select="emit('select', $event)"
          />
        </div>
      </aside>
    </div>
  </Teleport>
</template>

<style scoped>
.drawer-layer {
  position: fixed;
  inset: 0;
  z-index: 40;
  display: flex;
  justify-content: flex-end;
}

.drawer-mask {
  position: absolute;
  inset: 0;
  border: 0;
  background: rgba(16, 24, 40, 0.28);
  cursor: default;
}

.member-drawer {
  position: relative;
  z-index: 1;
  display: flex;
  width: min(440px, 92vw);
  height: 100%;
  min-height: 0;
  flex-direction: column;
  background: #ffffff;
  box-shadow: -12px 0 30px rgba(16, 24, 40, 0.16);
}

.drawer-header {
  display: flex;
  flex: 0 0 auto;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 16px;
  border-bottom: 1px solid #e4e7ec;
}

.drawer-title {
  min-width: 0;
}

.drawer-title strong,
.drawer-title small {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.drawer-title strong {
  font-size: 17px;
}

.drawer-title small {
  margin-top: 3px;
  color: #667085;
  font-size: 13px;
}

.drawer-actions {
  display: flex;
  flex: 0 0 auto;
  gap: 8px;
}

.drawer-actions button {
  height: 34px;
  border: 0;
  border-radius: 7px;
  padding: 0 10px;
  background: #eef4ff;
  color: #175cd3;
  font: inherit;
  font-size: 13px;
  font-weight: 700;
  cursor: pointer;
}

.drawer-actions button:disabled {
  background: #eaecf0;
  color: #98a2b3;
  cursor: not-allowed;
}

.drawer-body {
  min-height: 0;
  overflow-y: auto;
  padding: 14px;
}
</style>
