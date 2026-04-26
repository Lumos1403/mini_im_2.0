<script setup lang="ts">
import { computed } from 'vue'

import type { GroupRole } from '../../api/group'

const props = defineProps<{
  role?: GroupRole | ''
}>()

const normalizedRole = computed(() => (props.role === 'owner' || props.role === 'admin' ? props.role : 'member'))
const visible = computed(() => normalizedRole.value === 'owner' || normalizedRole.value === 'admin')
const label = computed(() => (normalizedRole.value === 'owner' ? '群主' : '管理员'))
</script>

<template>
  <span v-if="visible" :class="['group-role-badge', normalizedRole]">{{ label }}</span>
</template>

<style scoped>
.group-role-badge {
  display: inline-flex;
  height: 18px;
  align-items: center;
  border-radius: 4px;
  padding: 0 6px;
  font-size: 11px;
  font-weight: 700;
  line-height: 18px;
  white-space: nowrap;
}

.group-role-badge.owner {
  background: #fff4cc;
  color: #946200;
}

.group-role-badge.admin {
  background: #dcfae6;
  color: #027a48;
}
</style>
