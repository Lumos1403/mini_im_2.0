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
const requestMessage = ref('你好，我想添加你为好友')

const currentUserID = computed(() => auth.user?.user_id || '')

onMounted(() => {
  void refreshInitialData()
})

watch(currentUserID, (userID, oldUserID) => {
  if (userID === oldUserID) {
    return
  }
  void refreshInitialData()
})

async function refreshInitialData() {
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
  if (!window.confirm(`确定删除好友 ${friendDisplayName(item)} 吗？`)) {
    return
  }
  await friendStore.removeFriend(item.friend_user_id, item.conversation_id)
}

async function blockFriend(item: FriendItem) {
  if (!window.confirm(`确定拉黑好友 ${friendDisplayName(item)} 吗？`)) {
    return
  }
  await friendStore.block(item.friend_user_id)
}

async function unblockFriend(item: FriendItem) {
  await friendStore.unblock(item.friend_user_id)
}

function openChat(item: FriendItem) {
  emit('open-chat', item)
}

function isBlocked(item: FriendItem) {
  return item.is_blocked_by_me
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

function bioText(user: FriendUser) {
  return user.bio?.trim() || '暂无个性签名'
}

function friendBioText(item: FriendItem) {
  return item.bio?.trim() || '暂无个性签名'
}

function statusText(status: FriendRequest['status']) {
  const statusMap: Record<FriendRequest['status'], string> = {
    pending: '待处理',
    accepted: '已同意',
    rejected: '已拒绝',
    expired: '已过期',
  }
  return statusMap[status] || status
}

function formatTime(value: string) {
  const time = Date.parse(value)
  if (!Number.isFinite(time)) {
    return value
  }
  return new Date(time).toLocaleString()
}
</script>

<template>
  <aside class="friend-panel">
    <header class="friend-header">
      <div class="friend-title">
        <h2>好友</h2>
        <span v-if="friendStore.pendingRequestCount > 0" class="request-badge">
          {{ friendStore.pendingRequestCount }}
        </span>
      </div>
      <button class="ghost-button" type="button" :disabled="friendStore.operating" @click="refreshActiveTab">
        刷新
      </button>
    </header>

    <nav class="friend-tabs" aria-label="好友面板">
      <button
        type="button"
        :class="{ active: activeTab === 'friends' }"
        @click="activeTab = 'friends'"
      >
        好友列表
      </button>
      <button
        type="button"
        :class="{ active: activeTab === 'requests' }"
        @click="activeTab = 'requests'"
      >
        好友申请
      </button>
      <button
        type="button"
        :class="{ active: activeTab === 'search' }"
        @click="activeTab = 'search'"
      >
        搜索用户
      </button>
    </nav>

    <p v-if="friendStore.errorMessage" class="status-text error">{{ friendStore.errorMessage }}</p>
    <p v-else-if="friendStore.noticeMessage" class="status-text success">{{ friendStore.noticeMessage }}</p>

    <section v-if="activeTab === 'friends'" class="friend-body">
      <div v-if="friendStore.loadingFriends" class="empty-text">加载中</div>
      <article v-for="item in friendStore.friends" :key="item.friend_user_id" class="user-card">
        <div class="user-main">
          <img v-if="item.avatar_url" class="user-avatar image" :src="item.avatar_url" alt="" />
          <span v-else class="user-avatar">{{ friendAvatarText(item) }}</span>
          <div class="user-meta">
            <strong>{{ friendDisplayName(item) }}</strong>
            <small>{{ item.friend_user_id }}</small>
            <p>{{ friendBioText(item) }}</p>
          </div>
        </div>
        <div class="action-row">
          <button type="button" :disabled="friendStore.operating" @click="openChat(item)">打开聊天</button>
          <button type="button" :disabled="friendStore.operating" @click="deleteFriend(item)">删除</button>
          <button
            v-if="isBlocked(item)"
            type="button"
            :disabled="friendStore.operating"
            @click="unblockFriend(item)"
          >
            解除拉黑
          </button>
          <button
            v-else
            class="danger-button"
            type="button"
            :disabled="friendStore.operating"
            @click="blockFriend(item)"
          >
            拉黑
          </button>
        </div>
      </article>
      <div v-if="!friendStore.loadingFriends && friendStore.friends.length === 0" class="empty-text">
        暂无好友
      </div>
    </section>

    <section v-else-if="activeTab === 'requests'" class="friend-body">
      <div v-if="friendStore.loadingRequests" class="empty-text">加载中</div>
      <article v-for="request in friendStore.receivedRequests" :key="request.request_id" class="user-card">
        <div class="user-main">
          <img v-if="request.user.avatar_url" class="user-avatar image" :src="request.user.avatar_url" alt="" />
          <span v-else class="user-avatar">{{ avatarText(request.user) }}</span>
          <div class="user-meta">
            <strong>{{ displayName(request.user) }}</strong>
            <small>{{ request.user.user_id }} · {{ formatTime(request.created_at) }}</small>
            <p>{{ request.message || '对方请求添加你为好友' }}</p>
          </div>
        </div>
        <div class="action-row">
          <span class="request-status">{{ statusText(request.status) }}</span>
          <template v-if="request.status === 'pending'">
            <button type="button" :disabled="friendStore.operating" @click="acceptRequest(request)">同意</button>
            <button type="button" :disabled="friendStore.operating" @click="rejectRequest(request)">拒绝</button>
          </template>
        </div>
      </article>
      <div
        v-if="!friendStore.loadingRequests && friendStore.receivedRequests.length === 0"
        class="empty-text"
      >
        暂无好友申请
      </div>
    </section>

    <section v-else class="friend-body">
      <form class="search-form" @submit.prevent="handleSearch">
        <input v-model="keyword" type="text" placeholder="输入 user_id 或昵称" />
        <button type="submit" :disabled="friendStore.searching">
          {{ friendStore.searching ? '搜索中' : '搜索' }}
        </button>
      </form>
      <input v-model="requestMessage" class="request-input" type="text" maxlength="100" />

      <article v-for="user in friendStore.searchResults" :key="user.user_id" class="user-card">
        <div class="user-main">
          <img v-if="user.avatar_url" class="user-avatar image" :src="user.avatar_url" alt="" />
          <span v-else class="user-avatar">{{ avatarText(user) }}</span>
          <div class="user-meta">
            <strong>{{ displayName(user) }}</strong>
            <small>{{ user.user_id }}</small>
            <p>{{ bioText(user) }}</p>
          </div>
        </div>
        <div class="action-row">
          <button type="button" :disabled="friendStore.operating || isSelf(user)" @click="sendRequest(user)">
            {{ isSelf(user) ? '本人' : '添加好友' }}
          </button>
        </div>
      </article>
    </section>
  </aside>
</template>

<style scoped>
.friend-panel {
  display: flex;
  min-width: 0;
  min-height: 0;
  flex-direction: column;
  border-left: 1px solid #dde3ee;
  background: #ffffff;
}

.friend-header {
  display: flex;
  flex: 0 0 56px;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
  border-bottom: 1px solid #dde3ee;
}

.friend-title {
  display: flex;
  align-items: center;
  gap: 8px;
}

.friend-title h2 {
  margin: 0;
  font-size: 18px;
}

.request-badge {
  min-width: 20px;
  height: 20px;
  padding: 0 6px;
  border-radius: 10px;
  background: #d92d20;
  color: #ffffff;
  font-size: 12px;
  font-weight: 700;
  line-height: 20px;
  text-align: center;
}

.friend-tabs {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 4px;
  padding: 10px 12px;
  border-bottom: 1px solid #eef2f6;
}

.friend-tabs button,
.ghost-button,
.action-row button,
.search-form button {
  height: 34px;
  border: 0;
  border-radius: 7px;
  background: #eef4ff;
  color: #175cd3;
  font: inherit;
  font-size: 13px;
  font-weight: 700;
  cursor: pointer;
}

.friend-tabs button.active {
  background: #1570ef;
  color: #ffffff;
}

.ghost-button {
  min-width: 56px;
}

.friend-body {
  min-height: 0;
  overflow-y: auto;
  padding: 12px;
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

.search-form {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 68px;
  gap: 8px;
  margin-bottom: 8px;
}

.search-form input,
.request-input {
  min-width: 0;
  height: 36px;
  border: 1px solid #cfd6e4;
  border-radius: 7px;
  padding: 0 10px;
  font: inherit;
}

.request-input {
  width: 100%;
  margin-bottom: 10px;
}

.user-card {
  padding: 12px;
  margin-bottom: 10px;
  border: 1px solid #e4e7ec;
  border-radius: 8px;
  background: #ffffff;
}

.user-main {
  display: flex;
  min-width: 0;
  gap: 10px;
}

.user-avatar {
  display: grid;
  width: 40px;
  height: 40px;
  flex: 0 0 40px;
  place-items: center;
  border-radius: 50%;
  background: #344054;
  color: #ffffff;
  font-weight: 700;
  object-fit: cover;
}

.user-avatar.image {
  background: #f2f4f7;
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
  margin-top: 2px;
  color: #667085;
}

.user-meta p {
  display: -webkit-box;
  overflow: hidden;
  margin: 5px 0 0;
  color: #475467;
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

.action-row button {
  min-width: 68px;
  padding: 0 10px;
}

.action-row button.danger-button {
  background: #fff1f3;
  color: #c01048;
}

.action-row button:disabled,
.ghost-button:disabled,
.search-form button:disabled {
  background: #eaecf0;
  color: #98a2b3;
  cursor: not-allowed;
}

.request-status {
  display: inline-flex;
  height: 34px;
  align-items: center;
  color: #667085;
  font-size: 13px;
}

.empty-text {
  padding: 16px;
  color: #667085;
  text-align: center;
}
</style>
