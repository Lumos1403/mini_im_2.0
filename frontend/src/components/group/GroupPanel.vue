<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { computed, ref, watch } from 'vue'

import type { GroupJoinRequest, GroupMember } from '../../api/group'
import { useAuthStore } from '../../stores/auth'
import { useChatStore } from '../../stores/chat'
import { useGroupStore } from '../../stores/group'
import GroupRoleBadge from './GroupRoleBadge.vue'

const emit = defineEmits<{
  (event: 'open-conversation', conversationID: string): void
}>()

const auth = useAuthStore()
const chat = useChatStore()
const groupStore = useGroupStore()

const { searchResults, joinRequests, members, loading, operating, errorMessage, noticeMessage } =
  storeToRefs(groupStore)

const createName = ref('')
const searchKeyword = ref('')
const joinMessage = ref('I want to join this group.')
const allowMemberInvite = ref(true)
const maxMembers = ref(50)

const activeConversation = computed(() => chat.activeConversation)
const activeGroupID = computed(() => activeConversation.value?.group_id || '')
const currentUserID = computed(() => auth.user?.user_id || '')
const currentMember = computed(() => members.value.find((item) => item.user_id === currentUserID.value))
const currentRole = computed(() => currentMember.value?.role || 'member')
const canManageRequests = computed(() => currentRole.value === 'owner' || currentRole.value === 'admin')
const canManageAdmins = computed(() => currentRole.value === 'owner')
const canManageMembers = computed(() => currentRole.value === 'owner' || currentRole.value === 'admin')
const canEditMaxMembers = computed(() => currentRole.value === 'owner')

watch(
  activeGroupID,
  (groupID) => {
    if (!groupID) {
      groupStore.members = []
      groupStore.joinRequests = []
      return
    }
    void groupStore.loadMembers(groupID)
    void groupStore.loadJoinRequests(groupID)
  },
  { immediate: true },
)

async function createGroup() {
  const conversationID = await groupStore.create(createName.value)
  if (conversationID) {
    createName.value = ''
    emit('open-conversation', conversationID)
  }
}

async function search() {
  await groupStore.search(searchKeyword.value)
}

async function applyJoin(groupID: string) {
  await groupStore.applyJoin(groupID, joinMessage.value)
}

function openConversation(conversationID: string) {
  if (conversationID) {
    emit('open-conversation', conversationID)
  }
}

function refreshActiveGroup() {
  if (!activeGroupID.value) {
    return
  }
  void groupStore.loadMembers(activeGroupID.value)
  void groupStore.loadJoinRequests(activeGroupID.value)
}

function accept(request: GroupJoinRequest) {
  void groupStore.accept(request)
}

function reject(request: GroupJoinRequest) {
  void groupStore.reject(request)
}

function setAdmin(member: GroupMember) {
  if (activeGroupID.value) {
    void groupStore.setAdmin(activeGroupID.value, member.user_id)
  }
}

function unsetAdmin(member: GroupMember) {
  if (activeGroupID.value) {
    void groupStore.unsetAdmin(activeGroupID.value, member.user_id)
  }
}

function mute(member: GroupMember) {
  if (activeGroupID.value) {
    void groupStore.mute(activeGroupID.value, member.user_id)
  }
}

function unmute(member: GroupMember) {
  if (activeGroupID.value) {
    void groupStore.unmute(activeGroupID.value, member.user_id)
  }
}

function saveInviteSetting() {
  if (activeGroupID.value) {
    void groupStore.saveInviteSetting(activeGroupID.value, allowMemberInvite.value)
  }
}

function saveMaxMembers() {
  if (activeGroupID.value) {
    void groupStore.saveMaxMembers(activeGroupID.value, maxMembers.value)
  }
}

function dissolveGroup() {
  if (!activeGroupID.value || !window.confirm('Dissolve this group?')) {
    return
  }
  void groupStore.dissolve(activeGroupID.value)
}

function leaveGroup() {
  if (!activeGroupID.value || !window.confirm('Leave this group?')) {
    return
  }
  void groupStore.leave(activeGroupID.value)
}

function displayName(member: GroupMember) {
  return member.nickname || member.user_id
}

function canPromote(member: GroupMember) {
  return canManageAdmins.value && member.role === 'member' && member.user_id !== currentUserID.value
}

function canDemote(member: GroupMember) {
  return canManageAdmins.value && member.role === 'admin' && member.user_id !== currentUserID.value
}

function canMute(member: GroupMember) {
  if (!canManageMembers.value || member.user_id === currentUserID.value || member.role === 'owner') {
    return false
  }
  if (currentRole.value === 'admin' && member.role !== 'member') {
    return false
  }
  return true
}
</script>

<template>
  <section class="group-panel">
    <header class="panel-header">
      <div>
        <h2>Groups</h2>
        <small>Global group tools and current group controls</small>
      </div>
      <button type="button" :disabled="loading || operating || !activeGroupID" @click="refreshActiveGroup">
        Refresh
      </button>
    </header>

    <p v-if="errorMessage" class="status error">{{ errorMessage }}</p>
    <p v-else-if="noticeMessage" class="status success">{{ noticeMessage }}</p>

    <section class="group-section">
      <h3>Create group</h3>
      <form class="inline-form" @submit.prevent="createGroup">
        <input v-model="createName" maxlength="100" placeholder="Group name" />
        <button type="submit" :disabled="operating">Create</button>
      </form>
    </section>

    <section class="group-section">
      <h3>Find group</h3>
      <form class="inline-form" @submit.prevent="search">
        <input v-model="searchKeyword" placeholder="Group no or keyword" />
        <button type="submit" :disabled="loading">Search</button>
      </form>
      <input v-model="joinMessage" class="wide-input" maxlength="255" />
      <article v-for="group in searchResults" :key="group.group_id" class="mini-card">
        <strong>{{ group.name }}</strong>
        <small>{{ group.group_no }} / {{ group.status }}</small>
        <div class="action-row">
          <button
            v-if="group.is_member"
            type="button"
            :disabled="!group.conversation_id"
            @click="openConversation(group.conversation_id)"
          >
            Open
          </button>
          <button v-else type="button" :disabled="operating" @click="applyJoin(group.group_id)">
            Request join
          </button>
        </div>
      </article>
    </section>

    <section class="group-section">
      <div class="section-title-row">
        <h3>Current group</h3>
        <GroupRoleBadge :role="currentRole" />
      </div>
      <div v-if="!activeGroupID" class="empty-text">Select a group conversation first.</div>
      <template v-else>
        <p class="meta-line">{{ activeConversation?.title }} / {{ activeConversation?.group_no }}</p>

        <div class="settings-row">
          <label>
            <input v-model="allowMemberInvite" type="checkbox" />
            Allow member invite
          </label>
          <button type="button" :disabled="operating || !canManageRequests" @click="saveInviteSetting">
            Save
          </button>
        </div>
        <div class="settings-row">
          <input v-model.number="maxMembers" min="1" type="number" />
          <button type="button" :disabled="operating || !canEditMaxMembers" @click="saveMaxMembers">
            Save max
          </button>
        </div>
        <div class="action-row">
          <button class="danger" type="button" :disabled="operating || currentRole !== 'owner'" @click="dissolveGroup">
            Dissolve
          </button>
          <button class="danger" type="button" :disabled="operating || currentRole === 'owner'" @click="leaveGroup">
            Leave
          </button>
        </div>

        <h3>Join requests</h3>
        <article v-for="request in joinRequests" :key="request.request_id" class="mini-card">
          <strong>{{ request.user.nickname || request.user.user_id }}</strong>
          <small>{{ request.status }} / {{ request.message || 'No message' }}</small>
          <div v-if="request.status === 'pending' && canManageRequests" class="action-row">
            <button type="button" :disabled="operating" @click="accept(request)">Accept</button>
            <button type="button" :disabled="operating" @click="reject(request)">Reject</button>
          </div>
        </article>
        <div v-if="joinRequests.length === 0" class="empty-text">No join requests.</div>

        <h3>Members</h3>
        <article v-for="member in members" :key="member.user_id" class="mini-card">
          <strong>{{ displayName(member) }}</strong>
          <small>{{ member.user_id }} / {{ member.role }} / {{ member.mute_until || 'not muted' }}</small>
          <div class="action-row">
            <button v-if="canPromote(member)" type="button" :disabled="operating" @click="setAdmin(member)">
              Set admin
            </button>
            <button v-if="canDemote(member)" type="button" :disabled="operating" @click="unsetAdmin(member)">
              Remove admin
            </button>
            <button v-if="canMute(member)" type="button" :disabled="operating" @click="mute(member)">
              Mute
            </button>
            <button v-if="canMute(member)" type="button" :disabled="operating" @click="unmute(member)">
              Unmute
            </button>
          </div>
        </article>
        <div v-if="members.length === 0" class="empty-text">No members loaded.</div>
      </template>
    </section>
  </section>
</template>

<style scoped>
.group-panel {
  min-height: 0;
  color: var(--text);
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  border-bottom: 1px solid var(--border);
  padding: 14px;
}

h2,
h3,
p {
  margin: 0;
}

h2 {
  color: #fff7e8;
  font-size: 17px;
}

.panel-header small,
.meta-line {
  color: var(--text-muted);
  font-size: 12px;
}

.status {
  margin: 10px 12px 0;
  border-radius: 10px;
  padding: 9px 10px;
  font-size: 13px;
}

.status.error {
  color: #ffd4d4;
  background: rgba(239, 68, 68, 0.14);
}

.status.success {
  color: #cbffe3;
  background: rgba(94, 226, 160, 0.13);
}

.group-section {
  display: grid;
  gap: 10px;
  border-bottom: 1px solid rgba(240, 207, 132, 0.1);
  padding: 14px;
}

.section-title-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

h3 {
  color: var(--text);
  font-size: 14px;
}

.inline-form {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 92px;
  gap: 8px;
}

input {
  height: 36px;
  min-width: 0;
  border: 1px solid rgba(240, 207, 132, 0.16);
  border-radius: 9px;
  padding: 0 10px;
  color: var(--text);
  background: rgba(0, 0, 0, 0.22);
}

.wide-input {
  width: 100%;
}

button {
  height: 34px;
  border: 1px solid rgba(240, 207, 132, 0.16);
  border-radius: 9px;
  padding: 0 10px;
  color: var(--text-soft);
  background: rgba(255, 255, 255, 0.07);
  cursor: pointer;
  font-weight: 760;
}

button:hover {
  border-color: rgba(240, 207, 132, 0.3);
  color: var(--text);
}

.mini-card {
  border: 1px solid rgba(240, 207, 132, 0.12);
  border-radius: 12px;
  padding: 11px;
  background: rgba(255, 255, 255, 0.055);
}

.mini-card strong,
.mini-card small {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.mini-card small {
  margin-top: 4px;
  color: var(--text-muted);
  font-size: 12px;
}

.settings-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.settings-row label {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  color: var(--text-muted);
  font-size: 13px;
}

.settings-row input[type="checkbox"] {
  width: 16px;
  height: 16px;
}

.settings-row input[type="number"] {
  width: 110px;
}

.action-row {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 8px;
}

.danger {
  color: #ffd8d8;
  border-color: rgba(248, 113, 113, 0.32);
  background: rgba(239, 68, 68, 0.12);
}

.empty-text {
  border: 1px dashed rgba(240, 207, 132, 0.14);
  border-radius: 12px;
  padding: 18px;
  color: var(--text-muted);
  text-align: center;
}
</style>
