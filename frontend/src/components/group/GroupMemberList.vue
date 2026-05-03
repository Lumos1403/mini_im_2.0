<script setup lang="ts">
import { computed } from 'vue'

import type { GroupFriendshipStatus, GroupMember, GroupRole } from '../../api/group'
import GroupRoleBadge from './GroupRoleBadge.vue'

const props = defineProps<{
  members: GroupMember[]
  loading?: boolean
  selectedUserId?: string
}>()

const emit = defineEmits<{
  (event: 'select', member: GroupMember): void
}>()

const sortedMembers = computed(() =>
  [...props.members].sort((left, right) => {
    const roleDiff = roleRank(left.role) - roleRank(right.role)
    if (roleDiff !== 0) {
      return roleDiff
    }
    const joinedDiff = Date.parse(left.joined_at) - Date.parse(right.joined_at)
    if (Number.isFinite(joinedDiff) && joinedDiff !== 0) {
      return joinedDiff
    }
    return displayName(left).localeCompare(displayName(right))
  }),
)

function roleRank(role: GroupRole) {
  const ranks: Record<GroupRole, number> = {
    owner: 0,
    admin: 1,
    member: 2,
  }
  return ranks[role] ?? 2
}

function displayName(member: GroupMember) {
  return member.nickname || member.user_id
}

function avatarText(member: GroupMember) {
  return displayName(member).slice(0, 1).toUpperCase() || '#'
}

function roleText(role: GroupRole) {
  const labels: Record<GroupRole, string> = {
    owner: 'Owner',
    admin: 'Admin',
    member: 'Member',
  }
  return labels[role] || 'Member'
}

function friendshipText(status: GroupFriendshipStatus) {
  const labels: Record<GroupFriendshipStatus, string> = {
    self: 'You',
    friend: 'Friend',
    not_friend: 'Not friend',
    pending_sent: 'Request sent',
    pending_received: 'Request received',
  }
  return labels[status] || status
}

function muteText(value: string | null) {
  if (!value) {
    return 'Not muted'
  }
  const time = Date.parse(value)
  if (!Number.isFinite(time)) {
    return value
  }
  return time > Date.now() ? `Muted until ${new Date(time).toLocaleString()}` : 'Not muted'
}
</script>

<template>
  <div class="group-member-list">
    <div v-if="loading && sortedMembers.length === 0" class="empty-text">Loading members...</div>
    <div v-else-if="sortedMembers.length === 0" class="empty-text">No members loaded.</div>

    <article
      v-for="member in sortedMembers"
      :key="member.user_id"
      :class="['member-row', { selected: member.user_id === selectedUserId }]"
    >
      <button class="avatar-button" type="button" @click="emit('select', member)">
        <img v-if="member.avatar_url" class="member-avatar image" :src="member.avatar_url" alt="" />
        <span v-else class="member-avatar">{{ avatarText(member) }}</span>
      </button>

      <div class="member-meta">
        <button class="name-button" type="button" @click="emit('select', member)">
          <strong>{{ displayName(member) }}</strong>
          <GroupRoleBadge :role="member.role" />
        </button>
        <small>{{ member.user_id }}</small>
        <p>{{ member.bio?.trim() || 'No bio' }}</p>
        <div class="member-tags">
          <span>{{ roleText(member.role) }}</span>
          <span>{{ muteText(member.mute_until) }}</span>
          <span>{{ friendshipText(member.friendship_status) }}</span>
        </div>
      </div>
    </article>
  </div>
</template>

<style scoped>
.group-member-list {
  display: flex;
  min-height: 0;
  flex-direction: column;
  gap: 10px;
}

.member-row {
  display: flex;
  min-width: 0;
  gap: 10px;
  border: 1px solid rgba(240, 207, 132, 0.12);
  border-radius: 12px;
  padding: 10px;
  background: rgba(255, 255, 255, 0.055);
}

.member-row.selected {
  border-color: rgba(240, 207, 132, 0.35);
  background: rgba(240, 207, 132, 0.09);
}

.avatar-button,
.name-button {
  border: 0;
  padding: 0;
  background: transparent;
  color: inherit;
  cursor: pointer;
  text-align: left;
}

.member-avatar {
  display: grid;
  width: 42px;
  height: 42px;
  flex: 0 0 42px;
  place-items: center;
  border-radius: 12px;
  color: #171009;
  background: linear-gradient(135deg, #f0cf84, #9a7535);
  font-weight: 900;
  object-fit: cover;
}

.member-avatar.image {
  background: rgba(255, 255, 255, 0.08);
}

.member-meta {
  min-width: 0;
  flex: 1 1 auto;
}

.name-button {
  display: inline-flex;
  max-width: 100%;
  align-items: center;
  gap: 6px;
}

.name-button strong,
.member-meta small {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.member-meta small {
  display: block;
  margin-top: 3px;
  color: var(--text-muted);
  font-size: 12px;
}

.member-meta p {
  display: -webkit-box;
  overflow: hidden;
  margin: 5px 0 0;
  color: var(--text-muted);
  font-size: 13px;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.member-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 8px;
}

.member-tags span {
  min-height: 22px;
  border-radius: 6px;
  padding: 3px 7px;
  color: var(--text-muted);
  background: rgba(255, 255, 255, 0.07);
  font-size: 12px;
  line-height: 16px;
}

.empty-text {
  border: 1px dashed rgba(240, 207, 132, 0.14);
  border-radius: 12px;
  padding: 18px;
  color: var(--text-muted);
  text-align: center;
}
</style>
