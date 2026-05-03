<script setup lang="ts">
import { computed } from 'vue'

import type { Conversation } from '../../api/conversation'
import { useAuthStore } from '../../stores/auth'
import { useFriendStore } from '../../stores/friend'
import { useGroupStore } from '../../stores/group'
import GroupRoleBadge from '../group/GroupRoleBadge.vue'

const props = defineProps<{
  conversation?: Conversation
}>()

const emit = defineEmits<{
  (event: 'open-members'): void
}>()

const auth = useAuthStore()
const friendStore = useFriendStore()
const groupStore = useGroupStore()

const isGroup = computed(() => props.conversation?.conversation_type === 'group')
const peerUserID = computed(() => props.conversation?.peer_user_id || '')
const friend = computed(() =>
  friendStore.friends.find((item) => item.friend_user_id === peerUserID.value),
)
const currentMember = computed(() =>
  groupStore.members.find((item) => item.user_id === auth.user?.user_id),
)
const currentRole = computed(() => currentMember.value?.role || 'member')
const canLeaveGroup = computed(() => isGroup.value && currentRole.value !== 'owner')
const canDissolveGroup = computed(() => isGroup.value && currentRole.value === 'owner')

function avatarText() {
  const text = props.conversation?.title || props.conversation?.peer_nickname || props.conversation?.group_no || '#'
  return text.slice(0, 1).toUpperCase()
}

function removeFriend() {
  if (!props.conversation || !peerUserID.value) {
    return
  }
  if (window.confirm('Delete this friend and remove the private conversation?')) {
    void friendStore.removeFriend(peerUserID.value, props.conversation.conversation_id)
  }
}

function blockFriend() {
  if (peerUserID.value && window.confirm('Block this friend? Incoming messages will be rejected.')) {
    void friendStore.block(peerUserID.value)
  }
}

function unblockFriend() {
  if (peerUserID.value) {
    void friendStore.unblock(peerUserID.value)
  }
}

function leaveGroup() {
  const groupID = props.conversation?.group_id
  if (groupID && window.confirm('Leave this group?')) {
    void groupStore.leave(groupID)
  }
}

function dissolveGroup() {
  const groupID = props.conversation?.group_id
  if (groupID && window.confirm('Dissolve this group for all members?')) {
    void groupStore.dissolve(groupID)
  }
}
</script>

<template>
  <aside class="detail-panel">
    <div v-if="!conversation" class="empty-detail">
      <span class="detail-avatar">#</span>
      <strong>No conversation selected</strong>
      <p>Conversation details appear only after a logged-in user selects a thread.</p>
    </div>

    <template v-else>
      <header class="detail-header">
        <span class="detail-avatar">{{ avatarText() }}</span>
        <div>
          <h2>{{ conversation.title || 'Untitled conversation' }}</h2>
          <small>{{ isGroup ? 'Group conversation' : 'Private conversation' }}</small>
        </div>
      </header>

      <section class="detail-section">
        <h3>Identity</h3>
        <dl>
          <div>
            <dt>Conversation ID</dt>
            <dd>{{ conversation.conversation_id }}</dd>
          </div>
          <div v-if="!isGroup">
            <dt>Peer user ID</dt>
            <dd>{{ peerUserID || 'Unavailable' }}</dd>
          </div>
          <div v-if="isGroup">
            <dt>Group ID</dt>
            <dd>{{ conversation.group_id }}</dd>
          </div>
          <div v-if="isGroup">
            <dt>Group No</dt>
            <dd>{{ conversation.group_no || 'Unavailable' }}</dd>
          </div>
        </dl>
      </section>

      <section v-if="isGroup" class="detail-section">
        <div class="section-title-row">
          <h3>Members</h3>
          <GroupRoleBadge :role="currentRole" />
        </div>
        <p>{{ groupStore.members.length }} loaded members. Your role is {{ currentRole }}.</p>
        <button type="button" @click="emit('open-members')">Open member list</button>
      </section>

      <section v-else class="detail-section">
        <h3>Friend Controls</h3>
        <p v-if="friend?.is_blocked_by_me">You are blocking this friend.</p>
        <p v-else>Private controls use the peer user_id from the conversation contract.</p>
        <div class="action-grid">
          <button v-if="friend?.is_blocked_by_me" type="button" :disabled="friendStore.operating" @click="unblockFriend">
            Unblock
          </button>
          <button v-else type="button" :disabled="friendStore.operating || !peerUserID" @click="blockFriend">
            Block
          </button>
          <button class="danger-button" type="button" :disabled="friendStore.operating || !peerUserID" @click="removeFriend">
            Delete friend
          </button>
        </div>
      </section>

      <section v-if="isGroup" class="detail-section">
        <h3>Group Controls</h3>
        <div class="action-grid">
          <button class="danger-button" type="button" :disabled="groupStore.operating || !canLeaveGroup" @click="leaveGroup">
            Leave group
          </button>
          <button class="danger-button" type="button" :disabled="groupStore.operating || !canDissolveGroup" @click="dissolveGroup">
            Dissolve group
          </button>
        </div>
      </section>

      <p v-if="friendStore.errorMessage || groupStore.errorMessage" class="detail-status error">
        {{ friendStore.errorMessage || groupStore.errorMessage }}
      </p>
      <p v-else-if="friendStore.noticeMessage || groupStore.noticeMessage" class="detail-status success">
        {{ friendStore.noticeMessage || groupStore.noticeMessage }}
      </p>
    </template>
  </aside>
</template>

<style scoped>
.detail-panel {
  min-width: 0;
  min-height: 0;
  overflow-y: auto;
  border-left: 1px solid var(--border);
  padding: 18px;
  background: rgba(12, 14, 15, 0.7);
  backdrop-filter: blur(18px);
}

.empty-detail,
.detail-header {
  border: 1px solid rgba(240, 207, 132, 0.13);
  border-radius: 14px;
  padding: 16px;
  background: rgba(255, 255, 255, 0.055);
}

.empty-detail {
  display: grid;
  gap: 10px;
  justify-items: start;
  color: var(--text-muted);
}

.empty-detail strong {
  color: var(--text);
}

.empty-detail p {
  margin: 0;
  line-height: 1.5;
}

.detail-header {
  display: flex;
  align-items: center;
  gap: 12px;
}

.detail-avatar {
  display: grid;
  width: 48px;
  height: 48px;
  flex: 0 0 48px;
  place-items: center;
  border-radius: 14px;
  color: #171009;
  background: linear-gradient(135deg, #f0cf84, #9a7535);
  font-size: 18px;
  font-weight: 900;
}

h2,
h3,
p {
  margin: 0;
}

h2 {
  color: #fff7e8;
  font-size: 18px;
}

.detail-header small {
  display: block;
  margin-top: 4px;
  color: var(--text-muted);
  font-size: 12px;
}

.detail-section {
  margin-top: 14px;
  border: 1px solid rgba(240, 207, 132, 0.12);
  border-radius: 14px;
  padding: 14px;
  background: rgba(255, 255, 255, 0.045);
}

.section-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

h3 {
  color: var(--text);
  font-size: 14px;
}

.detail-section p {
  margin-top: 10px;
  color: var(--text-muted);
  font-size: 13px;
  line-height: 1.5;
}

dl {
  display: grid;
  gap: 10px;
  margin: 12px 0 0;
}

dt {
  color: var(--text-muted);
  font-size: 11px;
  text-transform: uppercase;
}

dd {
  margin: 4px 0 0;
  color: var(--text-soft);
  overflow-wrap: anywhere;
  font-size: 13px;
}

.action-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px;
  margin-top: 12px;
}

button {
  height: 36px;
  border: 1px solid rgba(240, 207, 132, 0.16);
  border-radius: 9px;
  color: var(--text-soft);
  background: rgba(255, 255, 255, 0.07);
  cursor: pointer;
  font-size: 13px;
  font-weight: 760;
}

button:hover {
  border-color: rgba(240, 207, 132, 0.3);
  color: var(--text);
  background: rgba(255, 255, 255, 0.1);
}

.danger-button {
  color: #ffd8d8;
  border-color: rgba(248, 113, 113, 0.32);
  background: rgba(239, 68, 68, 0.12);
}

.detail-status {
  margin-top: 14px;
  border-radius: 10px;
  padding: 10px 12px;
  font-size: 13px;
}

.detail-status.error {
  color: #ffd4d4;
  background: rgba(239, 68, 68, 0.14);
}

.detail-status.success {
  color: #cbffe3;
  background: rgba(94, 226, 160, 0.13);
}
</style>
