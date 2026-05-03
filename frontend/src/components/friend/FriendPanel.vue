<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'

import type { FriendItem, FriendRequest, FriendUser } from '../../api/friend'
import { useAuthStore } from '../../stores/auth'
import { useFriendStore } from '../../stores/friend'

type FriendTab = 'friends' | 'requests' | 'search'

const emit = defineEmits<{
  (event: 'open-chat', friend: FriendItem): void
}>()

const auth = useAuthStore()
const friendStore = useFriendStore()

const activeTab = ref<FriendTab>('friends')
const keyword = ref('')
const requestMessage = ref('Hi, I would like to add you as a friend.')

const currentUserID = computed(() => auth.user?.user_id || '')

onMounted(() => {
  void refreshInitialData()
})

watch(currentUserID, (userID, oldUserID) => {
  if (userID !== oldUserID) {
    void refreshInitialData()
  }
})

async function refreshInitialData() {
  if (!currentUserID.value) {
    return
  }
  await Promise.all([friendStore.loadFriends(), friendStore.loadReceivedRequests()])
}

async function refreshActiveTab() {
  if (activeTab.value === 'friends') {
    await friendStore.loadFriends()
    return
  }
  if (activeTab.value === 'requests') {
    await friendStore.loadReceivedRequests()
    return
  }
  await handleSearch()
}

async function handleSearch() {
  await friendStore.search(keyword.value)
}

async function sendRequest(user: FriendUser) {
  await friendStore.sendFriendRequest(user.user_id, requestMessage.value)
}

async function acceptRequest(request: FriendRequest) {
  await friendStore.acceptRequest(request.request_id)
}

async function rejectRequest(request: FriendRequest) {
  await friendStore.rejectRequest(request.request_id)
}

async function deleteFriend(item: FriendItem) {
  if (window.confirm(`Delete friend ${friendDisplayName(item)}?`)) {
    await friendStore.removeFriend(item.friend_user_id, item.conversation_id)
  }
}

async function blockFriend(item: FriendItem) {
  if (window.confirm(`Block friend ${friendDisplayName(item)}?`)) {
    await friendStore.block(item.friend_user_id)
  }
}

async function unblockFriend(item: FriendItem) {
  await friendStore.unblock(item.friend_user_id)
}

function openChat(item: FriendItem) {
  emit('open-chat', item)
}

function isSelf(user: FriendUser) {
  return user.user_id === currentUserID.value
}

function displayName(user: FriendUser) {
  return user.nickname || user.username || user.user_id
}

function friendDisplayName(item: FriendItem) {
  return item.nickname || item.friend_user_id
}

function avatarText(user: FriendUser) {
  return displayName(user).slice(0, 1).toUpperCase() || '#'
}

function friendAvatarText(item: FriendItem) {
  return friendDisplayName(item).slice(0, 1).toUpperCase() || '#'
}

function statusText(status: FriendRequest['status']) {
  const statusMap: Record<FriendRequest['status'], string> = {
    pending: 'Pending',
    accepted: 'Accepted',
    rejected: 'Rejected',
    expired: 'Expired',
  }
  return statusMap[status] || status
}

function formatTime(value: string) {
  const time = Date.parse(value)
  return Number.isFinite(time) ? new Date(time).toLocaleString() : value
}
</script>

<template>
  <section class="friend-panel">
    <header class="panel-header">
      <div>
        <h2>Friends</h2>
        <small>{{ friendStore.friends.length }} contacts</small>
      </div>
      <button type="button" :disabled="friendStore.operating" @click="refreshActiveTab">Refresh</button>
    </header>

    <nav class="tabs" aria-label="Friend tools">
      <button type="button" :class="{ active: activeTab === 'friends' }" @click="activeTab = 'friends'">
        List
      </button>
      <button type="button" :class="{ active: activeTab === 'requests' }" @click="activeTab = 'requests'">
        Requests
        <span v-if="friendStore.pendingRequestCount > 0">{{ friendStore.pendingRequestCount }}</span>
      </button>
      <button type="button" :class="{ active: activeTab === 'search' }" @click="activeTab = 'search'">
        Add
      </button>
    </nav>

    <p v-if="friendStore.errorMessage" class="status error">{{ friendStore.errorMessage }}</p>
    <p v-else-if="friendStore.noticeMessage" class="status success">{{ friendStore.noticeMessage }}</p>

    <section v-if="activeTab === 'friends'" class="panel-body">
      <div v-if="friendStore.loadingFriends" class="empty-text">Loading friends...</div>
      <article v-for="item in friendStore.friends" :key="item.friend_user_id" class="user-card">
        <div class="user-main">
          <img v-if="item.avatar_url" class="avatar image" :src="item.avatar_url" alt="" />
          <span v-else class="avatar">{{ friendAvatarText(item) }}</span>
          <div class="user-meta">
            <strong>{{ friendDisplayName(item) }}</strong>
            <small>{{ item.friend_user_id }}</small>
            <p>{{ item.bio?.trim() || 'No bio' }}</p>
          </div>
        </div>
        <div class="action-row">
          <button type="button" :disabled="friendStore.operating" @click="openChat(item)">Open chat</button>
          <button
            v-if="item.is_blocked_by_me"
            type="button"
            :disabled="friendStore.operating"
            @click="unblockFriend(item)"
          >
            Unblock
          </button>
          <button v-else type="button" :disabled="friendStore.operating" @click="blockFriend(item)">
            Block
          </button>
          <button class="danger" type="button" :disabled="friendStore.operating" @click="deleteFriend(item)">
            Delete
          </button>
        </div>
      </article>
      <div v-if="!friendStore.loadingFriends && friendStore.friends.length === 0" class="empty-text">
        No friends yet.
      </div>
    </section>

    <section v-else-if="activeTab === 'requests'" class="panel-body">
      <div v-if="friendStore.loadingRequests" class="empty-text">Loading requests...</div>
      <article v-for="request in friendStore.receivedRequests" :key="request.request_id" class="user-card">
        <div class="user-main">
          <img v-if="request.user.avatar_url" class="avatar image" :src="request.user.avatar_url" alt="" />
          <span v-else class="avatar">{{ avatarText(request.user) }}</span>
          <div class="user-meta">
            <strong>{{ displayName(request.user) }}</strong>
            <small>{{ request.user.user_id }} / {{ formatTime(request.created_at) }}</small>
            <p>{{ request.message || 'No request message' }}</p>
          </div>
        </div>
        <div class="action-row">
          <span class="request-status">{{ statusText(request.status) }}</span>
          <template v-if="request.status === 'pending'">
            <button type="button" :disabled="friendStore.operating" @click="acceptRequest(request)">Accept</button>
            <button type="button" :disabled="friendStore.operating" @click="rejectRequest(request)">Reject</button>
          </template>
        </div>
      </article>
      <div v-if="!friendStore.loadingRequests && friendStore.receivedRequests.length === 0" class="empty-text">
        No friend requests.
      </div>
    </section>

    <section v-else class="panel-body">
      <form class="search-form" @submit.prevent="handleSearch">
        <input v-model="keyword" type="text" placeholder="Search by user_id or nickname" />
        <button type="submit" :disabled="friendStore.searching">
          {{ friendStore.searching ? 'Searching...' : 'Search' }}
        </button>
      </form>
      <input v-model="requestMessage" class="request-input" type="text" maxlength="100" />

      <article v-for="user in friendStore.searchResults" :key="user.user_id" class="user-card">
        <div class="user-main">
          <img v-if="user.avatar_url" class="avatar image" :src="user.avatar_url" alt="" />
          <span v-else class="avatar">{{ avatarText(user) }}</span>
          <div class="user-meta">
            <strong>{{ displayName(user) }}</strong>
            <small>{{ user.user_id }}</small>
            <p>{{ user.bio?.trim() || 'No bio' }}</p>
          </div>
        </div>
        <div class="action-row">
          <button type="button" :disabled="friendStore.operating || isSelf(user)" @click="sendRequest(user)">
            {{ isSelf(user) ? 'You' : 'Add friend' }}
          </button>
        </div>
      </article>
    </section>
  </section>
</template>

<style scoped>
.friend-panel {
  min-height: 0;
  color: var(--text);
}

.panel-header {
  display: flex;
  min-height: 58px;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  border-bottom: 1px solid var(--border);
  padding: 14px;
}

h2 {
  margin: 0;
  color: #fff7e8;
  font-size: 17px;
}

.panel-header small {
  color: var(--text-muted);
  font-size: 12px;
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

.tabs {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 6px;
  border-bottom: 1px solid rgba(240, 207, 132, 0.1);
  padding: 10px 12px;
}

.tabs button.active {
  color: #171009;
  background: linear-gradient(135deg, #f0cf84, #b68a3e);
}

.tabs span {
  margin-left: 4px;
}

.panel-body {
  display: grid;
  gap: 10px;
  padding: 12px;
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

.search-form {
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

.request-input {
  width: 100%;
}

.user-card {
  border: 1px solid rgba(240, 207, 132, 0.12);
  border-radius: 12px;
  padding: 11px;
  background: rgba(255, 255, 255, 0.055);
}

.user-main {
  display: flex;
  min-width: 0;
  gap: 10px;
}

.avatar {
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

.avatar.image {
  background: rgba(255, 255, 255, 0.08);
}

.user-meta {
  min-width: 0;
}

.user-meta strong,
.user-meta small {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.user-meta small {
  margin-top: 3px;
  color: var(--text-muted);
  font-size: 12px;
}

.user-meta p {
  display: -webkit-box;
  overflow: hidden;
  margin: 5px 0 0;
  color: var(--text-muted);
  font-size: 13px;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 2;
}

.action-row {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 10px;
}

.danger {
  color: #ffd8d8;
  border-color: rgba(248, 113, 113, 0.32);
  background: rgba(239, 68, 68, 0.12);
}

.request-status {
  display: inline-flex;
  height: 34px;
  align-items: center;
  color: var(--text-muted);
  font-size: 13px;
}

.empty-text {
  border: 1px dashed rgba(240, 207, 132, 0.14);
  border-radius: 12px;
  padding: 18px;
  color: var(--text-muted);
  text-align: center;
}
</style>
