<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { computed, ref, watch } from 'vue'

import type { GroupJoinRequest, GroupMember } from '../../api/group'
import { useAuthStore } from '../../stores/auth'
import { useChatStore } from '../../stores/chat'
import { useGroupStore } from '../../stores/group'

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
const joinMessage = ref('I want to join this group')
const allowMemberInvite = ref(true)
const maxMembers = ref(50)

const activeConversation = computed(() => chat.activeConversation)
const activeGroupID = computed(() => activeConversation.value?.group_id || '')
const currentUserID = computed(() => auth.user?.user_id || '')
const currentMember = computed(() => members.value.find((item) => item.user_id === currentUserID.value))
const currentRole = computed(() => currentMember.value?.role || '')
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
    <header class="group-header">
      <h2>Groups</h2>
      <button type="button" :disabled="loading || operating" @click="refreshActiveGroup">Refresh</button>
    </header>

    <p v-if="errorMessage" class="status-text error">{{ errorMessage }}</p>
    <p v-else-if="noticeMessage" class="status-text success">{{ noticeMessage }}</p>

    <section class="group-section">
      <h3>Create</h3>
      <form class="inline-form" @submit.prevent="createGroup">
        <input v-model="createName" maxlength="100" placeholder="Group name" />
        <button type="submit" :disabled="operating">Create</button>
      </form>
    </section>

    <section class="group-section">
      <h3>Search</h3>
      <form class="inline-form" @submit.prevent="search">
        <input v-model="searchKeyword" placeholder="Group no" />
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
          <button v-else type="button" :disabled="operating" @click="applyJoin(group.group_id)">Join</button>
        </div>
      </article>
    </section>

    <section class="group-section">
      <h3>Active Group</h3>
      <div v-if="!activeGroupID" class="empty-text">Select a group conversation</div>
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
        <button
          class="danger-button"
          type="button"
          :disabled="operating || currentRole !== 'owner'"
          @click="dissolveGroup"
        >
          Dissolve
        </button>

        <h3>Join Requests</h3>
        <article v-for="request in joinRequests" :key="request.request_id" class="mini-card">
          <strong>{{ request.user.nickname || request.user.user_id }}</strong>
          <small>{{ request.status }} / {{ request.message || 'No message' }}</small>
          <div v-if="request.status === 'pending' && canManageRequests" class="action-row">
            <button type="button" :disabled="operating" @click="accept(request)">Accept</button>
            <button type="button" :disabled="operating" @click="reject(request)">Reject</button>
          </div>
        </article>
        <div v-if="joinRequests.length === 0" class="empty-text">No requests</div>

        <h3>Members</h3>
        <article v-for="member in members" :key="member.user_id" class="mini-card">
          <strong>{{ displayName(member) }}</strong>
          <small>{{ member.user_id }} / {{ member.role }} / {{ member.mute_until || 'not muted' }}</small>
          <div class="action-row">
            <button v-if="canPromote(member)" type="button" :disabled="operating" @click="setAdmin(member)">
              Set admin
            </button>
            <button v-if="canDemote(member)" type="button" :disabled="operating" @click="unsetAdmin(member)">
              Unset admin
            </button>
            <button v-if="canMute(member)" type="button" :disabled="operating" @click="mute(member)">Mute 10m</button>
            <button v-if="canMute(member)" type="button" :disabled="operating" @click="unmute(member)">Unmute</button>
          </div>
        </article>
      </template>
    </section>
  </section>
</template>

<style scoped>
.group-panel {
  display: flex;
  min-height: 0;
  flex-direction: column;
  overflow-y: auto;
  border-top: 1px solid #dde3ee;
  background: #ffffff;
}

.group-header {
  display: flex;
  flex: 0 0 48px;
  align-items: center;
  justify-content: space-between;
  padding: 0 14px;
  border-bottom: 1px solid #eef2f6;
}

.group-header h2,
.group-section h3 {
  margin: 0;
  font-size: 16px;
}

.group-section {
  padding: 12px;
  border-bottom: 1px solid #eef2f6;
}

.inline-form,
.settings-row {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 86px;
  gap: 8px;
  margin-top: 8px;
}

input {
  min-width: 0;
  height: 34px;
  border: 1px solid #cfd6e4;
  border-radius: 7px;
  padding: 0 9px;
  font: inherit;
}

.wide-input {
  width: 100%;
  margin-top: 8px;
}

button {
  min-height: 34px;
  border: 0;
  border-radius: 7px;
  background: #eef4ff;
  color: #175cd3;
  font: inherit;
  font-size: 13px;
  font-weight: 700;
  cursor: pointer;
}

button:disabled {
  background: #eaecf0;
  color: #98a2b3;
  cursor: not-allowed;
}

.danger-button {
  width: 100%;
  margin-top: 8px;
  background: #fff1f3;
  color: #c01048;
}

.mini-card {
  padding: 10px;
  margin-top: 8px;
  border: 1px solid #e4e7ec;
  border-radius: 8px;
}

.mini-card strong,
.mini-card small,
.meta-line {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.mini-card small,
.meta-line,
.empty-text {
  color: #667085;
  font-size: 13px;
}

.action-row {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 8px;
}

.action-row button {
  min-width: 74px;
  padding: 0 8px;
}

.status-text {
  margin: 10px 12px 0;
  padding: 8px 10px;
  border-radius: 7px;
  font-size: 13px;
}

.status-text.error {
  background: #fff1f3;
  color: #c01048;
}

.status-text.success {
  background: #ecfdf3;
  color: #027a48;
}

.empty-text {
  padding: 12px;
  text-align: center;
}
</style>
