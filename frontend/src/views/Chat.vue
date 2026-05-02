<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'

import type { FriendItem } from '../api/friend'
import type { GroupMember } from '../api/group'
import FriendPanel from '../components/friend/FriendPanel.vue'
import GroupMemberDrawer from '../components/group/GroupMemberDrawer.vue'
import GroupMemberProfileModal from '../components/group/GroupMemberProfileModal.vue'
import GroupPanel from '../components/group/GroupPanel.vue'
import GroupRoleBadge from '../components/group/GroupRoleBadge.vue'
import SearchPanel from '../components/search/SearchPanel.vue'
import { useChatStore, type ChatMessage } from '../stores/chat'
import { useGroupStore } from '../stores/group'
import { useWsStore } from '../stores/ws'

const chat = useChatStore()
const ws = useWsStore()
const groupStore = useGroupStore()

const {
  conversations,
  activeConversationID,
  messages,
  draft,
  loadingConversations,
  loadingMessages,
  uploadingFile,
  hasMore,
  errorMessage,
  scrollToBottomSignal,
} = storeToRefs(chat)
const { connected: wsConnected } = storeToRefs(ws)
const {
  members: groupMembers,
  loadingMembers,
  memberDrawerOpen,
  memberProfileOpen,
  selectedMember,
  friendRequestingUserID,
} = storeToRefs(groupStore)

const fileInput = ref<HTMLInputElement | null>(null)
const messageArea = ref<HTMLElement | null>(null)
const showSearchPanel = ref(false)

const activeConversation = computed(() => chat.activeConversation)
const activeGroupID = computed(() =>
  activeConversation.value?.conversation_type === 'group' ? activeConversation.value.group_id : '',
)
const isGroupConversation = computed(() => Boolean(activeGroupID.value))
const canSend = computed(() => wsConnected.value && Boolean(activeConversationID.value) && draft.value.trim().length > 0)
const canUploadFile = computed(
  () =>
    wsConnected.value &&
    Boolean(activeConversationID.value) &&
    activeConversation.value?.conversation_type !== 'group' &&
    !uploadingFile.value,
)

onMounted(async () => {
  await chat.initialize()
  ws.connect()
  chat.startRecallNoticeExpiryTimer()
  await scrollToBottom('auto')
})

onBeforeUnmount(() => {
  ws.disconnect()
  chat.stopRecallNoticeExpiryTimer()
})

watch(scrollToBottomSignal, () => {
  void scrollToBottom()
})

watch(activeGroupID, (groupID, oldGroupID) => {
  if (groupID === oldGroupID || !memberDrawerOpen.value) {
    return
  }
  if (!groupID) {
    groupStore.closeMemberDrawer()
    return
  }
  groupStore.openMemberDrawer(groupID)
})

function sendMessage() {
  if (!canSend.value) {
    return
  }
  const payload = chat.prepareTextMessage()
  if (payload) {
    ws.sendChatMessage(payload)
  }
}

function triggerFileSelect() {
  if (!canUploadFile.value) {
    return
  }
  fileInput.value?.click()
}

async function handleFileSelected(event: Event) {
  const input = event.target as HTMLInputElement
  const selectedFile = input.files?.[0]
  input.value = ''
  if (!selectedFile) {
    return
  }
  if (!wsConnected.value) {
    errorMessage.value = 'WebSocket 未连接，暂不能发送文件'
    return
  }

  const payload = await chat.prepareUploadedFileMessage(selectedFile)
  if (payload) {
    ws.sendChatMessage(payload)
  }
}

async function scrollToBottom(behavior: ScrollBehavior = 'smooth') {
  await nextTick()
  const element = messageArea.value
  if (!element) {
    return
  }
  element.scrollTo({
    top: element.scrollHeight,
    behavior,
  })
}

function openFriendChat(friend: FriendItem) {
  void chat.openFriendChat(friend)
}

function openConversation(conversationID: string) {
  void chat.selectConversation(conversationID)
}

async function openSearchConversation(conversationID: string) {
  if (!chat.conversations.some((item) => item.conversation_id === conversationID)) {
    await chat.loadConversationList(false)
  }
  await chat.selectConversation(conversationID)
  showSearchPanel.value = false
}

function openGroupMembers() {
  if (activeGroupID.value) {
    groupStore.openMemberDrawer(activeGroupID.value)
  }
}

function refreshGroupMembers() {
  if (activeGroupID.value) {
    void groupStore.loadMembers(activeGroupID.value)
  }
}

function openGroupMemberProfile(member: GroupMember) {
  groupStore.openMemberProfile(member)
}

function addFriendFromGroup(member: GroupMember, message: string) {
  if (activeGroupID.value) {
    void groupStore.sendFriendRequestFromMember(activeGroupID.value, member.user_id, message)
  }
}

function groupSenderName(message: ChatMessage) {
  if (message.sender_nickname) {
    return message.sender_nickname
  }
  if (chat.isMine(message)) {
    return '我'
  }
  return message.sender_id
}
</script>

<template>
  <section class="chat-shell">
    <aside class="conversation-list">
      <div class="list-header">
        <h2>会话</h2>
        <span :class="['connection-dot', wsConnected ? 'online' : 'offline']"></span>
      </div>
      <div v-if="loadingConversations" class="empty-text">加载中</div>
      <button
        v-for="conversation in conversations"
        :key="conversation.conversation_id"
        :class="['conversation-item', { active: conversation.conversation_id === activeConversationID }]"
        type="button"
        @click="chat.selectConversation(conversation.conversation_id)"
      >
        <span class="avatar">{{ conversation.title.slice(0, 1) || '#' }}</span>
        <span class="conversation-meta">
          <span class="conversation-title-row">
            <strong>{{ conversation.title || '未命名会话' }}</strong>
            <span v-if="conversation.unread_count > 0" class="unread-badge">
              {{ chat.formatUnreadCount(conversation.unread_count) }}
            </span>
          </span>
          <small>{{ chat.formatConversationLastMessage(conversation.last_message) }}</small>
        </span>
      </button>
      <div v-if="!loadingConversations && conversations.length === 0" class="empty-text">暂无会话</div>
    </aside>

    <section class="chat-main">
      <header class="chat-header">
        <div class="chat-title">
          <strong>{{ activeConversation?.title || '聊天窗口' }}</strong>
          <small v-if="isGroupConversation && activeConversation?.group_no">
            群号 {{ activeConversation.group_no }}
          </small>
        </div>
        <div class="header-actions">
          <span v-if="errorMessage" class="error-text">{{ errorMessage }}</span>
          <button
            class="search-button"
            type="button"
            @click="showSearchPanel = true"
          >
            搜索
          </button>
          <button
            v-if="isGroupConversation"
            class="member-button"
            type="button"
            :disabled="!activeGroupID || loadingMembers"
            @click="openGroupMembers"
          >
            {{ loadingMembers ? '加载中' : '成员' }}
          </button>
          <button
            class="clear-button"
            type="button"
            :disabled="!activeConversationID || messages.length === 0"
            @click="chat.clearCurrentConversation"
          >
            清空
          </button>
        </div>
      </header>

      <div ref="messageArea" class="message-area">
        <button
          v-if="hasMore"
          class="load-more"
          type="button"
          :disabled="loadingMessages"
          @click="chat.loadOlderMessages"
        >
          加载更早消息
        </button>
        <div v-if="loadingMessages && messages.length === 0" class="empty-text">加载中</div>
        <div v-else-if="messages.length === 0" class="empty-text">暂无消息</div>
        <article
          v-for="message in messages"
          :key="message.message_id"
          :class="['message-row', { mine: chat.isMine(message), notice: message.is_recall_notice }]"
        >
          <div v-if="message.is_recall_notice" class="recall-notice">
            <span>{{ message.content }}</span>
            <button v-if="chat.canReEdit(message)" type="button" @click="chat.reEditMessage(message)">重新编辑</button>
          </div>
          <div v-else class="bubble-wrap">
            <span
              v-if="chat.isMine(message) && chat.isFailed(message)"
              class="failed-mark"
              :title="message.error_message || '发送失败'"
            >
              !
            </span>
            <div class="message-bubble">
              <span
                v-if="activeConversation?.conversation_type === 'group'"
                class="sender-line"
              >
                <span class="sender-name">{{ groupSenderName(message) }}</span>
                <GroupRoleBadge :role="message.sender_group_role || 'member'" />
              </span>
              <div v-if="message.message_type === 'file'" class="file-card">
                <span class="file-icon">文件</span>
                <span class="file-info">
                  <strong>{{ chat.getFileDisplayName(message) }}</strong>
                  <small>{{ chat.getFileMetaText(message) }}</small>
                </span>
                <button
                  class="file-download-button"
                  type="button"
                  :disabled="!chat.getFileDownloadID(message)"
                  @click="chat.downloadVisibleFile(message)"
                >
                  下载
                </button>
              </div>
              <p v-else>{{ message.content }}</p>
              <small>{{ message.send_status === 'sending' ? '发送中' : message.created_at }}</small>
              <div class="message-actions">
                <button v-if="chat.canDelete(message)" type="button" @click="chat.deleteVisibleMessage(message)">
                  删除
                </button>
                <button v-if="chat.canRecall(message)" type="button" @click="chat.recallVisibleMessage(message)">
                  撤回
                </button>
              </div>
            </div>
          </div>
        </article>
      </div>

      <footer class="composer">
        <input ref="fileInput" class="hidden-file-input" type="file" @change="handleFileSelected" />
        <button
          class="file-upload-button"
          type="button"
          :disabled="!canUploadFile"
          @click="triggerFileSelect"
        >
          {{ uploadingFile ? '上传中' : '文件' }}
        </button>
        <textarea
          v-model="draft"
          maxlength="2000"
          placeholder="输入消息"
          @keydown.enter.exact.prevent="sendMessage"
        ></textarea>
        <button type="button" :disabled="!canSend" @click="sendMessage">发送</button>
      </footer>
    </section>

    <aside class="side-panel">
      <FriendPanel @open-chat="openFriendChat" />
      <GroupPanel @open-conversation="openConversation" />
    </aside>

    <GroupMemberDrawer
      :open="memberDrawerOpen"
      :title="activeConversation?.title"
      :members="groupMembers"
      :loading="loadingMembers"
      :selected-user-id="selectedMember?.user_id || ''"
      @close="groupStore.closeMemberDrawer"
      @refresh="refreshGroupMembers"
      @select="openGroupMemberProfile"
    />
    <GroupMemberProfileModal
      :open="memberProfileOpen"
      :member="selectedMember"
      :requesting="Boolean(selectedMember && friendRequestingUserID === selectedMember.user_id)"
      @close="groupStore.closeMemberProfile"
      @add-friend="addFriendFromGroup"
    />
    <SearchPanel
      :open="showSearchPanel"
      @close="showSearchPanel = false"
      @open-conversation="openSearchConversation"
    />
  </section>
</template>

<style scoped>
.chat-shell {
  display: grid;
  grid-template-columns: 280px minmax(0, 1fr) 340px;
  height: calc(100vh - 56px);
  height: calc(100dvh - 56px);
  min-height: 0;
  overflow: hidden;
  background: #f5f7fb;
}

.side-panel {
  display: grid;
  min-width: 0;
  min-height: 0;
  grid-template-rows: minmax(0, 1fr) minmax(0, 1fr);
  border-left: 1px solid #dde3ee;
  background: #ffffff;
}

.conversation-list {
  min-height: 0;
  overflow-y: auto;
  padding: 18px;
  border-right: 1px solid #dde3ee;
  background: #ffffff;
}

.list-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 14px;
}

.list-header h2 {
  margin: 0;
  font-size: 18px;
}

.connection-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: #d92d20;
}

.connection-dot.online {
  background: #12b76a;
}

.conversation-item {
  display: flex;
  width: 100%;
  gap: 10px;
  align-items: center;
  padding: 10px;
  margin-bottom: 6px;
  border: 1px solid transparent;
  border-radius: 8px;
  background: transparent;
  color: #101828;
  text-align: left;
  cursor: pointer;
}

.conversation-item.active,
.conversation-item:hover {
  border-color: #c7d7fe;
  background: #eef4ff;
}

.avatar {
  display: grid;
  width: 36px;
  height: 36px;
  flex: 0 0 36px;
  place-items: center;
  border-radius: 50%;
  background: #344054;
  color: #ffffff;
  font-weight: 700;
}

.conversation-meta {
  min-width: 0;
  flex: 1 1 auto;
}

.conversation-title-row {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 8px;
}

.conversation-meta strong,
.conversation-meta small {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.conversation-meta strong {
  min-width: 0;
  flex: 1 1 auto;
}

.conversation-meta small {
  margin-top: 3px;
  color: #667085;
}

.unread-badge {
  display: grid;
  min-width: 18px;
  height: 18px;
  flex: 0 0 auto;
  place-items: center;
  border-radius: 999px;
  padding: 0 5px;
  background: #d92d20;
  color: #ffffff;
  font-size: 11px;
  font-weight: 700;
  line-height: 1;
}

.chat-main {
  display: flex;
  min-height: 0;
  min-width: 0;
  flex-direction: column;
}

.chat-header,
.composer {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 0 20px;
  border-bottom: 1px solid #dde3ee;
  background: #ffffff;
}

.chat-header {
  flex: 0 0 56px;
  justify-content: space-between;
}

.chat-title {
  min-width: 0;
}

.chat-title strong,
.chat-title small {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.chat-title small {
  margin-top: 2px;
  color: #667085;
  font-size: 12px;
}

.header-actions {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 10px;
}

.message-area {
  flex: 1 1 auto;
  min-height: 0;
  overflow-y: auto;
  padding: 20px;
}

.message-row {
  display: flex;
  margin-bottom: 12px;
}

.message-row.mine {
  justify-content: flex-end;
}

.message-row.notice {
  justify-content: center;
}

.bubble-wrap {
  display: flex;
  max-width: min(620px, 80%);
  align-items: center;
  gap: 8px;
}

.message-row.mine .bubble-wrap {
  flex-direction: row;
}

.message-bubble {
  padding: 10px 12px;
  border-radius: 8px;
  background: #ffffff;
  box-shadow: 0 1px 2px rgba(16, 24, 40, 0.08);
}

.message-row.mine .message-bubble {
  background: #1570ef;
  color: #ffffff;
}

.message-bubble p {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-word;
}

.sender-line {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 5px;
  color: #475467;
  font-size: 12px;
  font-weight: 700;
}

.sender-name {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.message-row.mine .sender-line {
  color: rgba(255, 255, 255, 0.82);
}

.message-bubble small {
  display: block;
  margin-top: 6px;
  color: #667085;
  font-size: 11px;
}

.message-row.mine .message-bubble small {
  color: rgba(255, 255, 255, 0.76);
}

.file-card {
  display: flex;
  width: min(360px, 72vw);
  max-width: 100%;
  align-items: center;
  gap: 10px;
}

.file-icon {
  display: grid;
  width: 42px;
  height: 42px;
  flex: 0 0 42px;
  place-items: center;
  border-radius: 8px;
  background: #eef4ff;
  color: #175cd3;
  font-size: 12px;
  font-weight: 700;
}

.message-row.mine .file-icon {
  background: rgba(255, 255, 255, 0.18);
  color: #ffffff;
}

.file-info {
  display: flex;
  min-width: 0;
  flex: 1 1 auto;
  flex-direction: column;
  gap: 3px;
}

.file-info strong {
  overflow: hidden;
  color: inherit;
  font-size: 14px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-info small {
  margin: 0;
  color: #667085;
}

.message-row.mine .file-info small {
  color: rgba(255, 255, 255, 0.76);
}

.file-download-button {
  flex: 0 0 auto;
  height: 32px;
  border: 0;
  border-radius: 6px;
  background: #eef4ff;
  color: #175cd3;
  font: inherit;
  font-size: 13px;
  font-weight: 700;
  cursor: pointer;
}

.message-row.mine .file-download-button {
  background: rgba(255, 255, 255, 0.18);
  color: #ffffff;
}

.file-download-button:disabled {
  cursor: not-allowed;
  opacity: 0.56;
}

.failed-mark {
  display: grid;
  width: 18px;
  height: 18px;
  place-items: center;
  border-radius: 50%;
  background: #d92d20;
  color: #ffffff;
  font-size: 12px;
  font-weight: 700;
}

.message-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 8px;
}

.message-actions button,
.recall-notice button,
.clear-button,
.member-button {
  border: 0;
  border-radius: 6px;
  background: transparent;
  color: #175cd3;
  font: inherit;
  font-size: 12px;
  cursor: pointer;
}

.message-row.mine .message-actions button {
  color: rgba(255, 255, 255, 0.86);
}

.recall-notice {
  display: inline-flex;
  max-width: 90%;
  align-items: center;
  gap: 10px;
  padding: 7px 10px;
  border-radius: 8px;
  background: #eef2f6;
  color: #475467;
  font-size: 13px;
}

.composer {
  flex: 0 0 96px;
  border-top: 1px solid #dde3ee;
  border-bottom: 0;
}

.hidden-file-input {
  display: none;
}

.file-upload-button {
  flex: 0 0 auto;
}

.composer textarea {
  width: 100%;
  height: 56px;
  resize: none;
  border: 1px solid #cfd6e4;
  border-radius: 8px;
  padding: 10px 12px;
  font: inherit;
}

.composer button,
.load-more,
.clear-button,
.member-button,
.search-button {
  min-width: 78px;
  height: 38px;
  border: 0;
  border-radius: 8px;
  background: #1570ef;
  color: #ffffff;
  font-weight: 700;
  cursor: pointer;
}

.composer button:disabled,
.load-more:disabled,
.clear-button:disabled,
.member-button:disabled,
.search-button:disabled {
  background: #98a2b3;
  cursor: not-allowed;
}

.clear-button,
.member-button,
.search-button {
  min-width: 64px;
  color: #ffffff;
  font-size: 13px;
  font-weight: 700;
}

.load-more {
  display: block;
  margin: 0 auto 16px;
}

.empty-text {
  padding: 16px;
  color: #667085;
  text-align: center;
}

.error-text {
  color: #d92d20;
  font-size: 13px;
}

@media (max-width: 1080px) {
  .chat-shell {
    grid-template-columns: 240px minmax(0, 1fr) 320px;
  }
}

@media (max-width: 720px) {
  .chat-shell {
    grid-template-columns: 1fr;
    grid-template-rows: minmax(150px, 22%) minmax(320px, 1fr) minmax(320px, 42%);
  }

  .conversation-list {
    border-right: 0;
    border-bottom: 1px solid #dde3ee;
  }

  .side-panel {
    border-left: 0;
    border-top: 1px solid #dde3ee;
  }

  .bubble-wrap {
    max-width: 92%;
  }
}
</style>
