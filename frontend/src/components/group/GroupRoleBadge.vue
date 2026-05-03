<script setup lang="ts">
import { computed } from 'vue'

import type { GroupRole } from '../../api/group'

const props = defineProps<{
  role?: GroupRole | ''
}>()

const normalizedRole = computed(() => (props.role === 'owner' || props.role === 'admin' ? props.role : 'member'))
const visible = computed(() => normalizedRole.value === 'owner' || normalizedRole.value === 'admin')
const label = computed(() => (normalizedRole.value === 'owner' ? 'Owner' : 'Admin'))
</script>

<template>
  <span v-if="visible" :class="['group-role-badge', normalizedRole]">{{ label }}</span>
</template>

<style scoped>
.group-role-badge {
  display: inline-flex;
  height: 19px;
  align-items: center;
  border-radius: 5px;
  padding: 0 6px;
  font-size: 11px;
  font-weight: 840;
  line-height: 19px;
  white-space: nowrap;
}

.group-role-badge.owner {
  color: #201304;
  background: #f0cf84;
}

.group-role-badge.admin {
  color: #041b10;
  background: #5ee2a0;
}
</style>
