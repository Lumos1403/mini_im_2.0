<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'

import {
  clearConversationMessages,
  deleteMessage,
  getRecallEditCache,
  listConversations,
  listMessages,
  recallMessage,
  type Conversation,
  type FileMessageExtra,
  type Message,
  type MessageType,
} from '../api/conversation'
import { downloadFile, uploadFile, type FileUploadResult } from '../api/file'
import type { FriendItem } from '../api/friend'
import FriendPanel from '../components/friend/FriendPanel.vue'
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

interface RecalledData {
  message_id: string
  conversation_id: string
  recalled_by: string
  recalled_at: string
}

interface StoredRecallNotice {
  user_id: string
  conversation_id: string
  message_id: string
  recalled_at: string
  editable_until: string
}

type ChatMessage = Message & {
  error_message?: string
  is_recall_notice?: boolean
  recalled_message_id?: string
  editable_until?: string
}

const recallNoticeStoragePrefix = 'mini_im:recall_notices:'
const defaultMaxUploadSizeMB = 50
const configuredMaxUploadSizeMB = Number(import.meta.env.VITE_FILE_MAX_SIZE_MB)
const maxUploadSizeMB =
  Number.isFinite(configuredMaxUploadSizeMB) && configuredMaxUploadSizeMB > 0
    ? configuredMaxUploadSizeMB
    : defaultMaxUploadSizeMB
const maxUploadSizeBytes = maxUploadSizeMB * 1024 * 1024

const auth = useAuthStore()
const conversations = ref<Conversation[]>([])
const activeConversationID = ref('')
const messages = ref<ChatMessage[]>([])
const draft = ref('')
const fileInput = ref<HTMLInputElement | null>(null)
const loadingConversations = ref(false)
const loadingMessages = ref(false)
const uploadingFile = ref(false)
const hasMore = ref(false)
const errorMessage = ref('')
const wsConnected = ref(false)
const recallNoticeNow = ref(Date.now())

let socket: WebSocket | null = null
let recallNoticeTimer: number | undefined

const activeConversation = computed(() =>
  conversations.value.find((item) => item.conversation_id === activeConversationID.value),
)

const canSend = computed(() => wsConnected.value && activeConversationID.value && draft.value.trim().length > 0)
const canUploadFile = computed(() => wsConnected.value && Boolean(activeConversationID.value) && !uploadingFile.value)

onMounted(async () => {
  await loadConversationList()
  connectWebSocket()
  startRecallNoticeExpiryTimer()
})

onBeforeUnmount(() => {
  socket?.close()
  socket = null
  if (recallNoticeTimer) {
    window.clearInterval(recallNoticeTimer)
    recallNoticeTimer = undefined
  }
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

async function openFriendChat(friend: FriendItem) {
  errorMessage.value = ''

  let conversationID = friend.conversation_id
  if (!conversationID) {
    await loadConversationList()
    conversationID = findFriendConversationID(friend.friend_user_id)
  } else if (!conversations.value.some((item) => item.conversation_id === conversationID)) {
    await loadConversationList()
  }

  if (!conversationID) {
    errorMessage.value = '未找到该好友会话，请刷新会话列表后重试'
    return
  }

  await selectConversation(conversationID)
}

function findFriendConversationID(friendUserID: string) {
  return (
    conversations.value.find(
      (item) => item.conversation_type === 'private' && item.peer_user_id === friendUserID,
    )?.conversation_id || ''
  )
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
    mergeStoredRecallNotices(activeConversationID.value)
    hasMore.value = result.has_more
  } catch (error) {
    errorMessage.value = getErrorMessage(error)
  } finally {
    loadingMessages.value = false
  }
}

async function loadOlderMessages() {
  const cursor = messages.value.find((message) => !message.is_recall_notice && isPersistedMessage(message))?.message_id
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

  if (!socket || socket.readyState !== WebSocket.OPEN || !activeConversationID.value) {
    errorMessage.value = 'WebSocket 未连接，暂不能发送文件'
    return
  }
  if (selectedFile.size > maxUploadSizeBytes) {
    errorMessage.value = `文件不能超过 ${maxUploadSizeMB}MB`
    return
  }

  uploadingFile.value = true
  errorMessage.value = ''
  try {
    const uploaded = await uploadFile(selectedFile)
    sendUploadedFileMessage(uploaded)
  } catch (error) {
    errorMessage.value = getErrorMessage(error)
  } finally {
    uploadingFile.value = false
  }
}

function sendUploadedFileMessage(uploaded: FileUploadResult) {
  if (!socket || socket.readyState !== WebSocket.OPEN || !activeConversationID.value) {
    errorMessage.value = 'WebSocket 未连接，文件已上传但消息未发送'
    return
  }

  const clientMsgID = crypto.randomUUID()
  const seq = crypto.randomUUID()
  const localExtra: Record<string, unknown> = {
    file_id: uploaded.file_id,
    file_name: uploaded.original_name,
    file_size: uploaded.file_size,
    mime_type: uploaded.mime_type,
  }
  const localMessage: ChatMessage = {
    client_msg_id: clientMsgID,
    message_id: clientMsgID,
    conversation_id: activeConversationID.value,
    sender_id: auth.user?.user_id || '',
    message_type: 'file',
    content: uploaded.file_id,
    extra_json: localExtra,
    send_status: 'sending',
    created_at: new Date().toISOString(),
  }

  messages.value = [...messages.value, localMessage]
  updateConversationLastMessage(activeConversationID.value, uploaded.file_id, 'file')

  socket.send(
    JSON.stringify({
      seq,
      type: 'chat.message.send',
      data: {
        conversation_id: activeConversationID.value,
        client_msg_id: clientMsgID,
        message_type: 'file',
        content: uploaded.file_id,
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
    return
  }
  if (envelope.type === 'chat.message.recalled') {
    applyRecalled(envelope.data as RecalledData)
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
    updateConversationLastMessage(data.conversation_id, data.content, data.message_type)
    return
  }
  if (messages.value.some((message) => message.message_id === data.message_id)) {
    return
  }
  messages.value = [...messages.value, { ...data }]
  updateConversationLastMessage(data.conversation_id, data.content, data.message_type)
}

function applyRecalled(data: RecalledData) {
  if (data.conversation_id !== activeConversationID.value) {
    return
  }

  messages.value = messages.value.filter((message) => message.message_id !== data.message_id)
  if (data.recalled_by === auth.user?.user_id) {
    showRecallNotice(data.message_id, data.conversation_id, '', data.recalled_at)
  }
}

async function deleteVisibleMessage(message: ChatMessage) {
  if (!canDelete(message)) {
    return
  }

  errorMessage.value = ''
  try {
    await deleteMessage(message.conversation_id, message.message_id)
    messages.value = messages.value.filter((item) => item.message_id !== message.message_id)
  } catch (error) {
    errorMessage.value = getErrorMessage(error)
  }
}

async function clearCurrentConversation() {
  if (!activeConversationID.value) {
    return
  }

  errorMessage.value = ''
  try {
    await clearConversationMessages(activeConversationID.value)
    removeStoredRecallNoticesByConversation(activeConversationID.value)
    messages.value = []
    hasMore.value = false
    conversations.value = conversations.value.map((item) => {
      if (item.conversation_id !== activeConversationID.value) {
        return item
      }
      return {
        ...item,
        last_message: null,
        unread_count: 0,
      }
    })
  } catch (error) {
    errorMessage.value = getErrorMessage(error)
  }
}

async function recallVisibleMessage(message: ChatMessage) {
  if (!canRecall(message)) {
    return
  }

  errorMessage.value = ''
  try {
    const result = await recallMessage(message.message_id)
    messages.value = messages.value.filter((item) => item.message_id !== message.message_id)
    showRecallNotice(result.message_id, message.conversation_id, result.editable_until)
  } catch (error) {
    errorMessage.value = getErrorMessage(error)
  }
}

async function reEditMessage(message: ChatMessage) {
  const messageID = message.recalled_message_id
  if (!messageID) {
    return
  }

  errorMessage.value = ''
  try {
    const cache = await getRecallEditCache(messageID)
    draft.value = cache.content
  } catch (error) {
    expireRecallNotice(messageID)
    errorMessage.value = getErrorMessage(error)
  }
}

function showRecallNotice(
  messageID: string,
  conversationID: string,
  editableUntil: string,
  recalledAt = new Date().toISOString(),
) {
  if (conversationID !== activeConversationID.value) {
    return
  }
  const existingNotice = messages.value.find((message) => message.recalled_message_id === messageID)
  const effectiveEditableUntil = editableUntil || existingNotice?.editable_until || ''
  const effectiveRecalledAt = recalledAt || existingNotice?.created_at || new Date().toISOString()
  const recallNotice = createRecallNoticeMessage(messageID, conversationID, effectiveRecalledAt, effectiveEditableUntil)

  messages.value = sortChatMessages([
    ...messages.value.filter((message) => message.recalled_message_id !== messageID && message.message_id !== messageID),
    recallNotice,
  ])
  persistRecallNotice(messageID, conversationID, effectiveRecalledAt, effectiveEditableUntil)
}

function updateConversationLastMessage(conversationID: string, content: string, messageType: MessageType = 'text') {
  conversations.value = conversations.value.map((item) => {
    if (item.conversation_id !== conversationID) {
      return item
    }
    return {
      ...item,
      last_message: {
        content: messageType === 'file' ? '文件' : content,
        message_type: messageType,
        created_at: new Date().toISOString(),
      },
    }
  })
}

function formatConversationLastMessage(lastMessage: Conversation['last_message']) {
  if (!lastMessage) {
    return '暂无消息'
  }
  if (lastMessage.message_type === 'file') {
    return '文件'
  }
  return lastMessage.content || '暂无消息'
}

function getFileExtra(message: ChatMessage): FileMessageExtra {
  const extra = message.extra_json || {}
  return {
    file_id: readString(extra.file_id),
    file_name: readString(extra.file_name),
    file_size: readNumber(extra.file_size),
    mime_type: readString(extra.mime_type),
  }
}

function getFileDisplayName(message: ChatMessage) {
  return getFileExtra(message).file_name || '文件'
}

function getFileMetaText(message: ChatMessage) {
  const extra = getFileExtra(message)
  return `${formatFileSize(extra.file_size)} / ${extra.mime_type || '类型未知'}`
}

function getFileDownloadID(message: ChatMessage) {
  const extra = getFileExtra(message)
  return extra.file_id || message.content.trim()
}

async function downloadVisibleFile(message: ChatMessage) {
  const fileID = getFileDownloadID(message)
  if (!fileID) {
    errorMessage.value = '文件信息缺失，无法下载'
    return
  }

  errorMessage.value = ''
  try {
    const result = await downloadFile(fileID)
    triggerBrowserDownload(result.blob, getFileExtra(message).file_name || result.fileName || '文件')
  } catch (error) {
    errorMessage.value = getErrorMessage(error)
  }
}

function triggerBrowserDownload(blob: Blob, fileName: string) {
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = fileName
  document.body.appendChild(link)
  link.click()
  link.remove()
  window.setTimeout(() => URL.revokeObjectURL(url), 0)
}

function formatFileSize(size?: number) {
  if (typeof size !== 'number' || !Number.isFinite(size) || size < 0) {
    return '大小未知'
  }

  const units = ['B', 'KB', 'MB', 'GB']
  let value = size
  let unitIndex = 0
  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024
    unitIndex += 1
  }
  return `${value.toFixed(unitIndex === 0 ? 0 : 1)} ${units[unitIndex]}`
}

function readString(value: unknown) {
  return typeof value === 'string' ? value.trim() : ''
}

function readNumber(value: unknown) {
  if (typeof value === 'number' && Number.isFinite(value)) {
    return value
  }
  if (typeof value === 'string' && value.trim()) {
    const parsed = Number(value)
    return Number.isFinite(parsed) ? parsed : undefined
  }
  return undefined
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

function canDelete(message: ChatMessage) {
  return !message.is_recall_notice && isPersistedMessage(message)
}

function canRecall(message: ChatMessage) {
  return !message.is_recall_notice && isMine(message) && message.send_status === 'sent' && isPersistedMessage(message)
}

function canReEdit(message: ChatMessage) {
  return Boolean(message.is_recall_notice && message.recalled_message_id && isBeforeEditableUntil(message.editable_until))
}

function isPersistedMessage(message: ChatMessage) {
  return /^\d+$/.test(message.message_id)
}

function createRecallNoticeMessage(
  messageID: string,
  conversationID: string,
  recalledAt: string,
  editableUntil: string,
): ChatMessage {
  return {
    client_msg_id: `recall-${messageID}`,
    message_id: `recall-${messageID}`,
    conversation_id: conversationID,
    sender_id: auth.user?.user_id || '',
    message_type: 'system',
    content: '你撤回了一条消息',
    extra_json: {},
    send_status: 'sent',
    created_at: recalledAt,
    is_recall_notice: true,
    recalled_message_id: messageID,
    editable_until: editableUntil,
  }
}

function mergeStoredRecallNotices(conversationID: string) {
  const notices = loadStoredRecallNotices().filter((notice) => notice.conversation_id === conversationID)
  if (notices.length === 0) {
    return
  }

  const noticeMessages = notices.map((notice) =>
    createRecallNoticeMessage(notice.message_id, notice.conversation_id, notice.recalled_at, notice.editable_until),
  )
  const noticeIDs = new Set(notices.map((notice) => notice.message_id))
  messages.value = sortChatMessages([
    ...messages.value.filter(
      (message) =>
        !noticeIDs.has(message.message_id) &&
        (!message.recalled_message_id || !noticeIDs.has(message.recalled_message_id)),
    ),
    ...noticeMessages,
  ])
}

function persistRecallNotice(messageID: string, conversationID: string, recalledAt: string, editableUntil: string) {
  const userID = auth.user?.user_id
  if (!userID || !editableUntil || !isBeforeEditableUntil(editableUntil)) {
    return
  }

  const notices = loadStoredRecallNotices().filter((notice) => notice.message_id !== messageID)
  notices.push({
    user_id: userID,
    conversation_id: conversationID,
    message_id: messageID,
    recalled_at: recalledAt,
    editable_until: editableUntil,
  })
  saveStoredRecallNotices(notices)
}

function loadStoredRecallNotices() {
  const key = recallNoticeStorageKey()
  const userID = auth.user?.user_id
  if (!key || !userID) {
    return []
  }

  let parsed: StoredRecallNotice[] = []
  try {
    const raw = localStorage.getItem(key)
    parsed = raw ? (JSON.parse(raw) as StoredRecallNotice[]) : []
  } catch {
    localStorage.removeItem(key)
    return []
  }

  const validNotices = parsed.filter(
    (notice) =>
      notice.user_id === userID &&
      notice.conversation_id &&
      notice.message_id &&
      notice.recalled_at &&
      notice.editable_until &&
      isBeforeEditableUntil(notice.editable_until),
  )
  if (validNotices.length !== parsed.length) {
    saveStoredRecallNotices(validNotices)
  }
  return validNotices
}

function saveStoredRecallNotices(notices: StoredRecallNotice[]) {
  const key = recallNoticeStorageKey()
  if (!key) {
    return
  }
  if (notices.length === 0) {
    localStorage.removeItem(key)
    return
  }
  localStorage.setItem(key, JSON.stringify(notices))
}

function removeStoredRecallNotice(messageID: string) {
  saveStoredRecallNotices(loadStoredRecallNotices().filter((notice) => notice.message_id !== messageID))
}

function removeStoredRecallNoticesByConversation(conversationID: string) {
  saveStoredRecallNotices(loadStoredRecallNotices().filter((notice) => notice.conversation_id !== conversationID))
}

function expireRecallNotice(messageID: string) {
  removeStoredRecallNotice(messageID)
  messages.value = messages.value.map((message) => {
    if (message.recalled_message_id !== messageID) {
      return message
    }
    return {
      ...message,
      editable_until: '',
    }
  })
}

function cleanupExpiredRecallNotices() {
  const notices = loadStoredRecallNotices()
  const activeNoticeIDs = new Set(notices.map((notice) => notice.message_id))
  messages.value = messages.value.map((message) => {
    if (!message.recalled_message_id || activeNoticeIDs.has(message.recalled_message_id)) {
      return message
    }
    return {
      ...message,
      editable_until: '',
    }
  })
}

function startRecallNoticeExpiryTimer() {
  recallNoticeTimer = window.setInterval(() => {
    recallNoticeNow.value = Date.now()
    cleanupExpiredRecallNotices()
  }, 30000)
}

function sortChatMessages(items: ChatMessage[]) {
  return [...items].sort((left, right) => parseMessageTime(left) - parseMessageTime(right))
}

function parseMessageTime(message: ChatMessage) {
  return parseTimeValue(message.is_recall_notice ? message.created_at : message.created_at)
}

function isBeforeEditableUntil(value?: string) {
  const editableUntil = parseTimeValue(value || '')
  return editableUntil > recallNoticeNow.value
}

function parseTimeValue(value: string) {
  const trimmed = value.trim()
  if (!trimmed) {
    return 0
  }
  const normalized = trimmed.includes('T') ? trimmed : trimmed.replace(' ', 'T')
  const parsed = Date.parse(normalized)
  return Number.isNaN(parsed) ? 0 : parsed
}

function recallNoticeStorageKey() {
  const userID = auth.user?.user_id
  return userID ? `${recallNoticeStoragePrefix}${userID}` : ''
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
          <small>{{ formatConversationLastMessage(conversation.last_message) }}</small>
        </span>
      </button>
      <div v-if="!loadingConversations && conversations.length === 0" class="empty-text">暂无会话</div>
    </aside>

    <section class="chat-main">
      <header class="chat-header">
        <strong>{{ activeConversation?.title || '聊天窗口' }}</strong>
        <div class="header-actions">
          <span v-if="errorMessage" class="error-text">{{ errorMessage }}</span>
          <button
            class="clear-button"
            type="button"
            :disabled="!activeConversationID || messages.length === 0"
            @click="clearCurrentConversation"
          >
            清空
          </button>
        </div>
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
          :class="['message-row', { mine: isMine(message), notice: message.is_recall_notice }]"
        >
          <div v-if="message.is_recall_notice" class="recall-notice">
            <span>{{ message.content }}</span>
            <button v-if="canReEdit(message)" type="button" @click="reEditMessage(message)">重新编辑</button>
          </div>
          <div v-else class="bubble-wrap">
            <span
              v-if="isMine(message) && isFailed(message)"
              class="failed-mark"
              :title="message.error_message || '发送失败'"
            >
              !
            </span>
            <div class="message-bubble">
              <div v-if="message.message_type === 'file'" class="file-card">
                <span class="file-icon">文件</span>
                <span class="file-info">
                  <strong>{{ getFileDisplayName(message) }}</strong>
                  <small>{{ getFileMetaText(message) }}</small>
                </span>
                <button
                  class="file-download-button"
                  type="button"
                  :disabled="!getFileDownloadID(message)"
                  @click="downloadVisibleFile(message)"
                >
                  下载
                </button>
              </div>
              <p v-else>{{ message.content }}</p>
              <small>{{ message.send_status === 'sending' ? '发送中' : message.created_at }}</small>
              <div class="message-actions">
                <button v-if="canDelete(message)" type="button" @click="deleteVisibleMessage(message)">
                  删除
                </button>
                <button v-if="canRecall(message)" type="button" @click="recallVisibleMessage(message)">
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

    <FriendPanel @open-chat="openFriendChat" />
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
.clear-button {
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
.clear-button {
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
.clear-button:disabled {
  background: #98a2b3;
  cursor: not-allowed;
}

.clear-button {
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
    grid-template-rows: minmax(150px, 24%) minmax(320px, 1fr) minmax(260px, 38%);
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
