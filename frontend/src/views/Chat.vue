<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'

import {
  listConversations,
  listMessages,
  type Conversation,
  type Message,
} from '../api/conversation'
import { useAuthStore } from '../stores/auth'

interface Envelope<T = unknown> {
  seq: string
  type: string
  data: T
  timestamp: number
}

interface AckData {
  client_msg_id: string
  message_id: string
  conversation_id: string
  send_status: string
  server_time: string
}

interface FailedData {
  client_msg_id: string
  message_id?: string
  conversation_id: string
  send_status: string
  code: string
  message: string
  server_time?: string
}

type ChatMessage = Message & {
  error_message?: string
}

const auth = useAuthStore()
const conversations = ref<Conversation[]>([])
const activeConversationID = ref('')
const messages = ref<ChatMessage[]>([])
const draft = ref('')
const loadingConversations = ref(false)
const loadingMessages = ref(false)
const hasMore = ref(false)
const errorMessage = ref('')
const wsConnected = ref(false)

let socket: WebSocket | null = null

const activeConversation = computed(() =>
  conversations.value.find((item) => item.conversation_id === activeConversationID.value),
)

const canSend = computed(() => wsConnected.value && activeConversationID.value && draft.value.trim().length > 0)

onMounted(async () => {
  await loadConversationList()
  connectWebSocket()
})

onBeforeUnmount(() => {
  socket?.close()
  socket = null
})

async function loadConversationList() {
  loadingConversations.value = true
  errorMessage.value = ''
  try {
    const result = await listConversations()
    conversations.value = result.list
    if (!activeConversationID.value && conversations.value.length > 0) {
      await selectConversation(conversations.value[0].conversation_id)
    }
  } catch (error) {
    errorMessage.value = getErrorMessage(error)
  } finally {
    loadingConversations.value = false
  }
}

async function selectConversation(conversationID: string) {
  if (activeConversationID.value === conversationID && messages.value.length > 0) {
    return
  }
  activeConversationID.value = conversationID
  messages.value = []
  hasMore.value = false
  await loadCurrentMessages('')
}

async function loadCurrentMessages(cursor: string) {
  if (!activeConversationID.value) {
    return
  }
  loadingMessages.value = true
  errorMessage.value = ''
  try {
    const result = await listMessages(activeConversationID.value, cursor)
    const incoming = result.list.map((item) => ({ ...item }))
    messages.value = cursor ? [...incoming, ...messages.value] : incoming
    hasMore.value = result.has_more
  } catch (error) {
    errorMessage.value = getErrorMessage(error)
  } finally {
    loadingMessages.value = false
  }
}

async function loadOlderMessages() {
  const cursor = messages.value[0]?.message_id
  if (!cursor || loadingMessages.value) {
    return
  }
  await loadCurrentMessages(cursor)
}

function connectWebSocket() {
  const token = localStorage.getItem('access_token')
  if (!token || socket) {
    return
  }

  socket = new WebSocket(buildWebSocketURL(token))
  socket.addEventListener('open', () => {
    wsConnected.value = true
  })
  socket.addEventListener('close', () => {
    wsConnected.value = false
    socket = null
  })
  socket.addEventListener('message', (event) => {
    handleEnvelope(event.data)
  })
}

function sendMessage() {
  const content = draft.value.trim()
  if (!socket || socket.readyState !== WebSocket.OPEN || !activeConversationID.value || !content) {
    return
  }

  const clientMsgID = crypto.randomUUID()
  const seq = crypto.randomUUID()
  const localMessage: ChatMessage = {
    client_msg_id: clientMsgID,
    message_id: clientMsgID,
    conversation_id: activeConversationID.value,
    sender_id: auth.user?.user_id || '',
    message_type: 'text',
    content,
    extra_json: {},
    send_status: 'sending',
    created_at: new Date().toISOString(),
  }
  messages.value = [...messages.value, localMessage]
  updateConversationLastMessage(activeConversationID.value, content)
  draft.value = ''

  socket.send(
    JSON.stringify({
      seq,
      type: 'chat.message.send',
      data: {
        conversation_id: activeConversationID.value,
        client_msg_id: clientMsgID,
        message_type: 'text',
        content,
        extra_json: {},
      },
      timestamp: Date.now(),
    }),
  )
}

function handleEnvelope(raw: string) {
  let envelope: Envelope
  try {
    envelope = JSON.parse(raw) as Envelope
  } catch {
    return
  }

  if (envelope.type === 'chat.message.ack') {
    applyAck(envelope.data as AckData)
    return
  }
  if (envelope.type === 'chat.message.failed') {
    applyFailed(envelope.data as FailedData)
    return
  }
  if (envelope.type === 'chat.message.receive') {
    applyReceive(envelope.data as Message)
  }
}

function applyAck(data: AckData) {
  messages.value = messages.value.map((message) => {
    if (message.client_msg_id !== data.client_msg_id) {
      return message
    }
    return {
      ...message,
      message_id: data.message_id,
      send_status: data.send_status,
      created_at: data.server_time,
      error_message: '',
    }
  })
}

function applyFailed(data: FailedData) {
  messages.value = messages.value.map((message) => {
    if (message.client_msg_id !== data.client_msg_id) {
      return message
    }
    return {
      ...message,
      message_id: data.message_id || message.message_id,
      send_status: data.send_status,
      created_at: data.server_time || message.created_at,
      error_message: data.message,
    }
  })
}

function applyReceive(data: Message) {
  if (data.conversation_id !== activeConversationID.value) {
    updateConversationLastMessage(data.conversation_id, data.content)
    return
  }
  if (messages.value.some((message) => message.message_id === data.message_id)) {
    return
  }
  messages.value = [...messages.value, { ...data }]
  updateConversationLastMessage(data.conversation_id, data.content)
}

function updateConversationLastMessage(conversationID: string, content: string) {
  conversations.value = conversations.value.map((item) => {
    if (item.conversation_id !== conversationID) {
      return item
    }
    return {
      ...item,
      last_message: {
        content,
        message_type: 'text',
        created_at: new Date().toISOString(),
      },
    }
  })
}

function buildWebSocketURL(token: string) {
  const explicitURL = import.meta.env.VITE_WS_URL
  if (explicitURL) {
    const separator = explicitURL.includes('?') ? '&' : '?'
    return `${explicitURL}${separator}token=${encodeURIComponent(token)}`
  }

  const protocol = window.location.protocol === 'https:' ? 'wss' : 'ws'
  const host = import.meta.env.DEV ? 'localhost:8081' : window.location.host
  return `${protocol}://${host}/ws?token=${encodeURIComponent(token)}`
}

function isMine(message: ChatMessage) {
  return message.sender_id === auth.user?.user_id
}

function isFailed(message: ChatMessage) {
  return message.send_status === 'failed' || message.send_status === 'failed_blocked'
}

function getErrorMessage(error: unknown) {
  return error instanceof Error ? error.message : '请求失败'
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
        @click="selectConversation(conversation.conversation_id)"
      >
        <span class="avatar">{{ conversation.title.slice(0, 1) || '#' }}</span>
        <span class="conversation-meta">
          <strong>{{ conversation.title || '未命名会话' }}</strong>
          <small>{{ conversation.last_message?.content || '暂无消息' }}</small>
        </span>
      </button>
      <div v-if="!loadingConversations && conversations.length === 0" class="empty-text">暂无会话</div>
    </aside>

    <section class="chat-main">
      <header class="chat-header">
        <strong>{{ activeConversation?.title || '聊天窗口' }}</strong>
        <span v-if="errorMessage" class="error-text">{{ errorMessage }}</span>
      </header>

      <div class="message-area">
        <button
          v-if="hasMore"
          class="load-more"
          type="button"
          :disabled="loadingMessages"
          @click="loadOlderMessages"
        >
          加载更早消息
        </button>
        <div v-if="loadingMessages && messages.length === 0" class="empty-text">加载中</div>
        <div v-else-if="messages.length === 0" class="empty-text">暂无消息</div>
        <article
          v-for="message in messages"
          :key="message.message_id"
          :class="['message-row', { mine: isMine(message) }]"
        >
          <div class="bubble-wrap">
            <span
              v-if="isMine(message) && isFailed(message)"
              class="failed-mark"
              :title="message.error_message || '发送失败'"
            >
              !
            </span>
            <div class="message-bubble">
              <p>{{ message.content }}</p>
              <small>{{ message.send_status === 'sending' ? '发送中' : message.created_at }}</small>
            </div>
          </div>
        </article>
      </div>

      <footer class="composer">
        <textarea
          v-model="draft"
          maxlength="2000"
          placeholder="输入消息"
          @keydown.enter.exact.prevent="sendMessage"
        ></textarea>
        <button type="button" :disabled="!canSend" @click="sendMessage">发送</button>
      </footer>
    </section>
  </section>
</template>

<style scoped>
.chat-shell {
  display: grid;
  grid-template-columns: 280px minmax(0, 1fr);
  min-height: calc(100vh - 56px);
  background: #f5f7fb;
}

.conversation-list {
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
}

.conversation-meta strong,
.conversation-meta small {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.conversation-meta small {
  margin-top: 3px;
  color: #667085;
}

.chat-main {
  display: grid;
  grid-template-rows: 56px minmax(0, 1fr) 96px;
  min-width: 0;
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
  justify-content: space-between;
}

.message-area {
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

.message-bubble small {
  display: block;
  margin-top: 6px;
  color: #667085;
  font-size: 11px;
}

.message-row.mine .message-bubble small {
  color: rgba(255, 255, 255, 0.76);
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

.composer {
  border-top: 1px solid #dde3ee;
  border-bottom: 0;
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
.load-more {
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
.load-more:disabled {
  background: #98a2b3;
  cursor: not-allowed;
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

@media (max-width: 720px) {
  .chat-shell {
    grid-template-columns: 1fr;
  }

  .conversation-list {
    border-right: 0;
    border-bottom: 1px solid #dde3ee;
  }

  .bubble-wrap {
    max-width: 92%;
  }
}
</style>
