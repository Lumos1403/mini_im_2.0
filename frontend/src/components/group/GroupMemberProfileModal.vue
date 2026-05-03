<script setup lang="ts">
import { computed, ref, watch } from 'vue'

import type { GroupFriendshipStatus, GroupMember, GroupRole } from '../../api/group'
import GroupRoleBadge from './GroupRoleBadge.vue'

const props = defineProps<{
  open: boolean
  member: GroupMember | null
  requesting?: boolean
}>()

const emit = defineEmits<{
  (event: 'close'): void
  (event: 'add-friend', member: GroupMember, message: string): void
}>()

const requestMessage = ref('I am a member of the same group.')

const displayName = computed(() => props.member?.nickname || props.member?.user_id || '')
const canAddFriend = computed(() => props.member?.friendship_status === 'not_friend')

watch(
  () => props.member?.user_id,
  () => {
    requestMessage.value = 'I am a member of the same group.'
  },
)

function avatarText(member: GroupMember) {
  return (member.nickname || member.user_id).slice(0, 1).toUpperCase() || '#'
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
    self: 'This is you',
    friend: 'Already friends',
    not_friend: 'Can add friend',
    pending_sent: 'Request sent',
    pending_received: 'They sent you a request',
  }
  return labels[status] || status
}

function addFriend() {
  if (!props.member || !canAddFriend.value || props.requesting) {
    return
  }
  emit('add-friend', props.member, requestMessage.value)
}
</script>

<template>
  <Teleport to="body">
    <div v-if="open && member" class="modal-layer">
      <button class="modal-mask" type="button" aria-label="Close member profile" @click="emit('close')"></button>
      <section class="profile-modal" aria-label="Group member profile">
        <header class="profile-header">
          <strong>Member profile</strong>
          <button type="button" @click="emit('close')">Close</button>
        </header>

        <div class="profile-body">
          <img v-if="member.avatar_url" class="profile-avatar image" :src="member.avatar_url" alt="" />
          <span v-else class="profile-avatar">{{ avatarText(member) }}</span>

          <div class="profile-name-row">
            <h3>{{ displayName }}</h3>
            <GroupRoleBadge :role="member.role" />
          </div>
          <small class="profile-id">{{ member.user_id }}</small>
          <p class="profile-bio">{{ member.bio?.trim() || 'No bio' }}</p>

          <dl class="profile-fields">
            <div>
              <dt>Role</dt>
              <dd>{{ roleText(member.role) }}</dd>
            </div>
            <div>
              <dt>Friendship</dt>
              <dd>{{ friendshipText(member.friendship_status) }}</dd>
            </div>
          </dl>

          <div class="profile-actions">
            <span v-if="member.friendship_status === 'self'" class="state-text">This is you</span>
            <button v-else-if="member.friendship_status === 'friend'" type="button" disabled>Already friends</button>
            <template v-else-if="member.friendship_status === 'not_friend'">
              <input v-model="requestMessage" maxlength="100" type="text" />
              <button type="button" :disabled="requesting" @click="addFriend">
                {{ requesting ? 'Sending...' : 'Add friend' }}
              </button>
            </template>
            <button v-else-if="member.friendship_status === 'pending_sent'" type="button" disabled>
              Request sent
            </button>
            <span v-else-if="member.friendship_status === 'pending_received'" class="state-text">
              They sent you a request
            </span>
          </div>
        </div>
      </section>
    </div>
  </Teleport>
</template>

<style scoped>
.modal-layer {
  position: fixed;
  inset: 0;
  z-index: 72;
  display: grid;
  place-items: center;
  padding: 20px;
}

.modal-mask {
  position: absolute;
  inset: 0;
  border: 0;
  background: rgba(0, 0, 0, 0.5);
}

.profile-modal {
  position: relative;
  z-index: 1;
  width: min(390px, 92vw);
  overflow: hidden;
  border: 1px solid rgba(240, 207, 132, 0.18);
  border-radius: 16px;
  color: var(--text);
  background: rgba(12, 14, 15, 0.96);
  box-shadow: 0 28px 72px rgba(0, 0, 0, 0.5);
}

.profile-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  border-bottom: 1px solid var(--border);
  padding: 14px 16px;
}

.profile-header button,
.profile-actions button {
  height: 34px;
  border: 1px solid rgba(240, 207, 132, 0.16);
  border-radius: 9px;
  padding: 0 12px;
  color: var(--text-soft);
  background: rgba(255, 255, 255, 0.07);
  cursor: pointer;
  font-size: 13px;
  font-weight: 760;
}

.profile-actions button:disabled {
  cursor: not-allowed;
  opacity: 0.58;
}

.profile-body {
  padding: 18px;
  text-align: center;
}

.profile-avatar {
  display: grid;
  width: 76px;
  height: 76px;
  margin: 0 auto 12px;
  place-items: center;
  border-radius: 18px;
  color: #171009;
  background: linear-gradient(135deg, #f0cf84, #9a7535);
  font-size: 24px;
  font-weight: 900;
  object-fit: cover;
}

.profile-avatar.image {
  background: rgba(255, 255, 255, 0.08);
}

.profile-name-row {
  display: flex;
  min-width: 0;
  align-items: center;
  justify-content: center;
  gap: 7px;
}

.profile-name-row h3 {
  overflow: hidden;
  margin: 0;
  color: #fff7e8;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 18px;
}

.profile-id {
  display: block;
  margin-top: 5px;
  color: var(--text-muted);
  font-size: 12px;
}

.profile-bio {
  margin: 12px 0 0;
  color: var(--text-muted);
  font-size: 13px;
  line-height: 1.5;
  word-break: break-word;
}

.profile-fields {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px;
  margin: 16px 0 0;
  text-align: left;
}

.profile-fields div {
  border-radius: 10px;
  padding: 9px 10px;
  background: rgba(255, 255, 255, 0.06);
}

.profile-fields dt,
.profile-fields dd {
  margin: 0;
}

.profile-fields dt {
  color: var(--text-muted);
  font-size: 12px;
}

.profile-fields dd {
  margin-top: 4px;
  color: var(--text-soft);
  font-size: 13px;
  font-weight: 800;
}

.profile-actions {
  display: grid;
  gap: 8px;
  margin-top: 16px;
}

.profile-actions input {
  width: 100%;
  height: 36px;
  border: 1px solid rgba(240, 207, 132, 0.16);
  border-radius: 9px;
  padding: 0 10px;
  color: var(--text);
  background: rgba(0, 0, 0, 0.22);
}

.state-text {
  display: block;
  border-radius: 10px;
  padding: 10px;
  color: var(--text-muted);
  background: rgba(255, 255, 255, 0.06);
  font-size: 13px;
}
</style>
