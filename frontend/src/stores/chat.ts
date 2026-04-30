import { defineStore } from 'pinia'

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
import { useAuthStore } from './auth'

export type ChatMessage = Message & {
  error_message?: string
  is_recall_notice?: boolean
  recalled_message_id?: string
  editable_until?: string
}

export interface OutgoingChatMessage {
  conversation_id: string
  client_msg_id: string
  message_type: MessageType
  content: string
  extra_json?: Record<string, unknown>
}

interface StoredRecallNotice {
  user_id: string
  conversation_id: string
  message_id: string
  recalled_at: string
  editable_until: string
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

interface ChatState {
  conversations: Conversation[]
  activeConversationID: string
  messages: ChatMessage[]
  draft: string
  loadingConversations: boolean
  loadingMessages: boolean
  uploadingFile: boolean
  hasMore: boolean
  errorMessage: string
  recallNoticeNow: number
  recallNoticeTimer?: number
  scrollToBottomSignal: number
  seenMessageKeys: string[]
}

const recallNoticeStoragePrefix = 'mini_im:recall_notices:'
const defaultMaxUploadSizeMB = 50
const configuredMaxUploadSizeMB = Number(import.meta.env.VITE_FILE_MAX_SIZE_MB)
const maxUploadSizeMB =
  Number.isFinite(configuredMaxUploadSizeMB) && configuredMaxUploadSizeMB > 0
    ? configuredMaxUploadSizeMB
    : defaultMaxUploadSizeMB
const maxUploadSizeBytes = maxUploadSizeMB * 1024 * 1024

export const useChatStore = defineStore('chat', {
  state: (): ChatState => ({
    conversations: [],
    activeConversationID: '',
    messages: [],
    draft: '',
    loadingConversations: false,
    loadingMessages: false,
    uploadingFile: false,
    hasMore: false,
    errorMessage: '',
    recallNoticeNow: Date.now(),
    recallNoticeTimer: undefined,
    scrollToBottomSignal: 0,
    seenMessageKeys: [],
  }),
  getters: {
    activeConversation(state): Conversation | undefined {
      return state.conversations.find((item) => item.conversation_id === state.activeConversationID)
    },
  },
  actions: {
    async initialize() {
      await this.loadConversationList()
    },

    async loadConversationList(autoSelect = true) {
      this.loadingConversations = true
      this.errorMessage = ''
      try {
        const result = await listConversations()
        this.conversations = result.list
        if (
          this.activeConversationID &&
          !this.conversations.some((item) => item.conversation_id === this.activeConversationID)
        ) {
          this.resetActiveConversation()
        }
        if (autoSelect && !this.activeConversationID && this.conversations.length > 0) {
          await this.selectConversation(this.conversations[0].conversation_id)
        }
      } catch (error) {
        this.errorMessage = getErrorMessage(error)
      } finally {
        this.loadingConversations = false
      }
    },

    async selectConversation(conversationID: string) {
      if (this.activeConversationID === conversationID && this.messages.length > 0) {
        this.clearConversationUnread(conversationID)
        this.requestScrollToBottom()
        return
      }

      this.activeConversationID = conversationID
      this.clearConversationUnread(conversationID)
      this.messages = []
      this.hasMore = false
      await this.loadCurrentMessages('')
      this.requestScrollToBottom()
    },

    async openFriendChat(friend: FriendItem) {
      this.errorMessage = ''

      let conversationID = friend.conversation_id
      if (!conversationID) {
        await this.loadConversationList(false)
        conversationID = this.findFriendConversationID(friend.friend_user_id)
      } else if (!this.conversations.some((item) => item.conversation_id === conversationID)) {
        await this.loadConversationList(false)
      }

      if (!conversationID) {
        this.errorMessage = '未找到该好友会话，请刷新会话列表后重试'
        return
      }

      await this.selectConversation(conversationID)
    },

    findFriendConversationID(friendUserID: string) {
      return (
        this.conversations.find(
          (item) => item.conversation_type === 'private' && item.peer_user_id === friendUserID,
        )?.conversation_id || ''
      )
    },

    removeConversation(conversationID: string) {
      if (!conversationID) {
        return
      }
      this.conversations = this.conversations.filter((item) => item.conversation_id !== conversationID)
      if (this.activeConversationID === conversationID) {
        this.resetActiveConversation()
      }
      this.removeStoredRecallNoticesByConversation(conversationID)
    },

    removePrivateConversationByPeer(peerUserID: string) {
      if (!peerUserID) {
        return
      }
      const conversationIDs = this.conversations
        .filter((item) => item.conversation_type === 'private' && item.peer_user_id === peerUserID)
        .map((item) => item.conversation_id)
      conversationIDs.forEach((conversationID) => this.removeConversation(conversationID))
    },

    removeGroupConversationByGroupID(groupID: string) {
      if (!groupID) {
        return
      }
      const conversationIDs = this.conversations
        .filter((item) => item.conversation_type === 'group' && item.group_id === groupID)
        .map((item) => item.conversation_id)
      conversationIDs.forEach((conversationID) => this.removeConversation(conversationID))
    },

    resetActiveConversation() {
      const previousConversationID = this.activeConversationID
      this.activeConversationID = ''
      this.messages = []
      this.draft = ''
      this.hasMore = false
      if (previousConversationID) {
        this.removeStoredRecallNoticesByConversation(previousConversationID)
      }
    },

    async loadCurrentMessages(cursor: string) {
      if (!this.activeConversationID) {
        return
      }
      this.loadingMessages = true
      this.errorMessage = ''
      try {
        const result = await listMessages(this.activeConversationID, cursor)
        const incoming = result.list.map((item) => normalizeIncomingMessage(item))
        this.messages = cursor
          ? sortChatMessages(dedupeMessages([...incoming, ...this.messages]))
          : sortChatMessages(dedupeMessages(incoming))
        this.rememberMessages(incoming)
        this.mergeStoredRecallNotices(this.activeConversationID)
        this.hasMore = result.has_more
      } catch (error) {
        this.errorMessage = getErrorMessage(error)
      } finally {
        this.loadingMessages = false
      }
    },

    async loadOlderMessages() {
      const cursor = this.messages.find((message) => !message.is_recall_notice && isPersistedMessage(message))
        ?.message_id
      if (!cursor || this.loadingMessages) {
        return
      }
      await this.loadCurrentMessages(cursor)
    },

    prepareTextMessage(): OutgoingChatMessage | null {
      const content = this.draft.trim()
      if (!this.activeConversationID || !content) {
        return null
      }

      const clientMsgID = crypto.randomUUID()
      const localMessage = createLocalMessage({
        clientMsgID,
        conversationID: this.activeConversationID,
        senderID: currentUserID(),
        messageType: 'text',
        content,
        extraJSON: {},
      })

      this.messages = sortChatMessages(appendOrMergeMessage(this.messages, localMessage))
      this.rememberMessage(localMessage)
      this.upsertConversationFromMessage(localMessage, { clearUnread: true, moveToTop: true })
      this.draft = ''
      this.requestScrollToBottom()

      return {
        conversation_id: localMessage.conversation_id,
        client_msg_id: localMessage.client_msg_id,
        message_type: localMessage.message_type,
        content: localMessage.content,
        extra_json: {},
      }
    },

    async prepareUploadedFileMessage(file: File): Promise<OutgoingChatMessage | null> {
      if (!this.activeConversationID) {
        this.errorMessage = '请先选择会话'
        return null
      }
      if (file.size > maxUploadSizeBytes) {
        this.errorMessage = `文件不能超过 ${maxUploadSizeMB}MB`
        return null
      }

      this.uploadingFile = true
      this.errorMessage = ''
      try {
        const uploaded = await uploadFile(file)
        return this.prepareFileMessage(uploaded)
      } catch (error) {
        this.errorMessage = getErrorMessage(error)
        return null
      } finally {
        this.uploadingFile = false
      }
    },

    prepareFileMessage(uploaded: FileUploadResult): OutgoingChatMessage | null {
      if (!this.activeConversationID) {
        this.errorMessage = 'WebSocket 未连接，文件已上传但消息未发送'
        return null
      }

      const clientMsgID = crypto.randomUUID()
      const localExtra: Record<string, unknown> = {
        file_id: uploaded.file_id,
        file_name: uploaded.original_name,
        file_size: uploaded.file_size,
        mime_type: uploaded.mime_type,
      }
      const localMessage = createLocalMessage({
        clientMsgID,
        conversationID: this.activeConversationID,
        senderID: currentUserID(),
        messageType: 'file',
        content: uploaded.file_id,
        extraJSON: localExtra,
      })

      this.messages = sortChatMessages(appendOrMergeMessage(this.messages, localMessage))
      this.rememberMessage(localMessage)
      this.upsertConversationFromMessage(localMessage, { clearUnread: true, moveToTop: true })
      this.requestScrollToBottom()

      return {
        conversation_id: localMessage.conversation_id,
        client_msg_id: localMessage.client_msg_id,
        message_type: localMessage.message_type,
        content: localMessage.content,
      }
    },

    applyAck(data: AckData) {
      let updatedMessage: ChatMessage | undefined
      const nextMessages = this.messages.map((message) => {
        if (message.client_msg_id !== data.client_msg_id) {
          return message
        }
        updatedMessage = {
          ...message,
          message_id: data.message_id,
          send_status: data.send_status,
          created_at: data.server_time,
          error_message: '',
        }
        return updatedMessage
      })

      this.messages = sortChatMessages(dedupeMessages(nextMessages))
      if (updatedMessage) {
        this.rememberMessage(updatedMessage)
        this.upsertConversationFromMessage(updatedMessage, { clearUnread: true, moveToTop: true })
      } else {
        this.rememberRawMessageKeys(data.message_id, data.client_msg_id)
      }
    },

    applyFailed(data: FailedData) {
      let updatedMessage: ChatMessage | undefined
      this.messages = this.messages.map((message) => {
        if (message.client_msg_id !== data.client_msg_id) {
          return message
        }
        updatedMessage = {
          ...message,
          message_id: data.message_id || message.message_id,
          send_status: data.send_status,
          created_at: data.server_time || message.created_at,
          error_message: data.message,
        }
        return updatedMessage
      })
      this.messages = dedupeMessages(this.messages)
      if (updatedMessage) {
        this.rememberMessage(updatedMessage)
      } else {
        this.rememberRawMessageKeys(data.message_id || '', data.client_msg_id)
      }
    },

    async applyReceive(data: Message) {
      const message = normalizeIncomingMessage(data)
      if (!message.message_id || !message.conversation_id) {
        return
      }

      const alreadySeen = this.hasSeenMessage(message)
      const isActiveConversation = message.conversation_id === this.activeConversationID
      this.ensureConversationFromMessage(message)
      this.upsertConversationFromMessage(message, {
        clearUnread: isActiveConversation,
        unreadDelta: isActiveConversation || alreadySeen ? 0 : 1,
        moveToTop: true,
      })

      if (isActiveConversation && !this.hasMessage(message)) {
        this.messages = sortChatMessages(appendOrMergeMessage(this.messages, message))
        this.requestScrollToBottom()
      }

      this.rememberMessage(message)
      if (!this.conversationHasServerProfile(message.conversation_id)) {
        await this.refreshConversationItem(message.conversation_id)
      }
    },

    applyRecalled(data: RecalledData) {
      if (data.conversation_id !== this.activeConversationID) {
        return
      }

      this.messages = this.messages.filter((message) => message.message_id !== data.message_id)
      if (data.recalled_by === currentUserID()) {
        this.showRecallNotice(data.message_id, data.conversation_id, '', data.recalled_at)
      }
      this.requestScrollToBottom()
    },

    async deleteVisibleMessage(message: ChatMessage) {
      if (!this.canDelete(message)) {
        return
      }

      this.errorMessage = ''
      try {
        await deleteMessage(message.conversation_id, message.message_id)
        this.messages = this.messages.filter((item) => item.message_id !== message.message_id)
      } catch (error) {
        this.errorMessage = getErrorMessage(error)
      }
    },

    async clearCurrentConversation() {
      if (!this.activeConversationID) {
        return
      }

      this.errorMessage = ''
      try {
        await clearConversationMessages(this.activeConversationID)
        this.removeStoredRecallNoticesByConversation(this.activeConversationID)
        this.messages = []
        this.hasMore = false
        this.conversations = this.conversations.map((item) => {
          if (item.conversation_id !== this.activeConversationID) {
            return item
          }
          return {
            ...item,
            last_message: null,
            unread_count: 0,
          }
        })
      } catch (error) {
        this.errorMessage = getErrorMessage(error)
      }
    },

    async recallVisibleMessage(message: ChatMessage) {
      if (!this.canRecall(message)) {
        return
      }

      this.errorMessage = ''
      try {
        const result = await recallMessage(message.message_id)
        this.messages = this.messages.filter((item) => item.message_id !== message.message_id)
        this.showRecallNotice(result.message_id, message.conversation_id, result.editable_until)
      } catch (error) {
        this.errorMessage = getErrorMessage(error)
      }
    },

    async reEditMessage(message: ChatMessage) {
      const messageID = message.recalled_message_id
      if (!messageID) {
        return
      }

      this.errorMessage = ''
      try {
        const cache = await getRecallEditCache(messageID)
        this.draft = cache.content
      } catch (error) {
        this.expireRecallNotice(messageID)
        this.errorMessage = getErrorMessage(error)
      }
    },

    showRecallNotice(
      messageID: string,
      conversationID: string,
      editableUntil: string,
      recalledAt = new Date().toISOString(),
    ) {
      if (conversationID !== this.activeConversationID) {
        return
      }
      const existingNotice = this.messages.find((message) => message.recalled_message_id === messageID)
      const effectiveEditableUntil = editableUntil || existingNotice?.editable_until || ''
      const effectiveRecalledAt = recalledAt || existingNotice?.created_at || new Date().toISOString()
      const recallNotice = this.createRecallNoticeMessage(
        messageID,
        conversationID,
        effectiveRecalledAt,
        effectiveEditableUntil,
      )

      this.messages = sortChatMessages([
        ...this.messages.filter(
          (message) => message.recalled_message_id !== messageID && message.message_id !== messageID,
        ),
        recallNotice,
      ])
      this.persistRecallNotice(messageID, conversationID, effectiveRecalledAt, effectiveEditableUntil)
      this.requestScrollToBottom()
    },

    ensureConversationFromMessage(message: ChatMessage) {
      if (this.conversations.some((item) => item.conversation_id === message.conversation_id)) {
        return
      }

      const localConversation = createLocalConversationFromMessage(message, currentUserID())
      this.conversations = [localConversation, ...this.conversations]
    },

    async refreshConversationItem(conversationID: string) {
      try {
        const result = await listConversations()
        const fetched = result.list.find((item) => item.conversation_id === conversationID)
        if (!fetched) {
          return
        }

        const current = this.conversations.find((item) => item.conversation_id === conversationID)
        const merged: Conversation = {
          ...fetched,
          last_message: fetched.last_message || current?.last_message || null,
          unread_count: current?.unread_count ?? fetched.unread_count,
        }
        this.conversations = [merged, ...this.conversations.filter((item) => item.conversation_id !== conversationID)]
      } catch {
        // The local conversation item is already enough for realtime display.
      }
    },

    upsertConversationFromMessage(
      message: ChatMessage,
      options: { clearUnread?: boolean; unreadDelta?: number; moveToTop?: boolean } = {},
    ) {
      const existing = this.conversations.find((item) => item.conversation_id === message.conversation_id)
      const base = existing || createLocalConversationFromMessage(message, currentUserID())
      const unreadCount = options.clearUnread
        ? 0
        : Math.max(0, base.unread_count + (options.unreadDelta || 0))
      const updated: Conversation = {
        ...base,
        last_message: {
          content: lastMessageContent(message),
          message_type: message.message_type,
          created_at: message.created_at,
        },
        unread_count: unreadCount,
      }
      const others = this.conversations.filter((item) => item.conversation_id !== message.conversation_id)
      this.conversations = options.moveToTop ? [updated, ...others] : [...others, updated]
    },

    clearConversationUnread(conversationID: string) {
      this.conversations = this.conversations.map((item) => {
        if (item.conversation_id !== conversationID) {
          return item
        }
        return {
          ...item,
          unread_count: 0,
        }
      })
    },

    conversationHasServerProfile(conversationID: string) {
      const conversation = this.conversations.find((item) => item.conversation_id === conversationID)
      if (!conversation) {
        return false
      }
      if (conversation.conversation_type === 'group') {
        return Boolean(conversation.group_id && conversation.title)
      }
      return Boolean(conversation.peer_user_id && conversation.peer_nickname)
    },

    hasMessage(message: ChatMessage) {
      return this.messages.some((item) => isSameMessage(item, message))
    },

    hasSeenMessage(message: ChatMessage) {
      return messageKeys(message).some((key) => this.seenMessageKeys.includes(key))
    },

    rememberMessage(message: ChatMessage) {
      this.rememberRawMessageKeys(message.message_id, message.client_msg_id)
    },

    rememberMessages(messages: ChatMessage[]) {
      messages.forEach((message) => this.rememberMessage(message))
    },

    rememberRawMessageKeys(messageID: string, clientMsgID: string) {
      const nextKeys = [...this.seenMessageKeys]
      for (const key of rawMessageKeys(messageID, clientMsgID)) {
        if (!nextKeys.includes(key)) {
          nextKeys.push(key)
        }
      }
      this.seenMessageKeys = nextKeys.slice(-1000)
    },

    requestScrollToBottom() {
      this.scrollToBottomSignal += 1
    },

    formatConversationLastMessage(lastMessage: Conversation['last_message']) {
      if (!lastMessage) {
        return '暂无消息'
      }
      if (lastMessage.message_type === 'file') {
        return '文件'
      }
      return lastMessage.content || '暂无消息'
    },

    formatUnreadCount(count: number) {
      if (count > 99) {
        return '99+'
      }
      return String(count)
    },

    isMine(message: ChatMessage) {
      return message.sender_id === currentUserID()
    },

    isFailed(message: ChatMessage) {
      return message.send_status === 'failed' || message.send_status === 'failed_blocked'
    },

    canDelete(message: ChatMessage) {
      return !message.is_recall_notice && isPersistedMessage(message)
    },

    canRecall(message: ChatMessage) {
      return !message.is_recall_notice && this.isMine(message) && message.send_status === 'sent' && isPersistedMessage(message)
    },

    canReEdit(message: ChatMessage) {
      return Boolean(message.is_recall_notice && message.recalled_message_id && this.isBeforeEditableUntil(message.editable_until))
    },

    getFileExtra(message: ChatMessage): FileMessageExtra {
      const extra = message.extra_json || {}
      return {
        file_id: readString(extra.file_id),
        file_name: readString(extra.file_name),
        file_size: readNumber(extra.file_size),
        mime_type: readString(extra.mime_type),
      }
    },

    getFileDisplayName(message: ChatMessage) {
      return this.getFileExtra(message).file_name || '文件'
    },

    getFileMetaText(message: ChatMessage) {
      const extra = this.getFileExtra(message)
      return `${formatFileSize(extra.file_size)} / ${extra.mime_type || '类型未知'}`
    },

    getFileDownloadID(message: ChatMessage) {
      const extra = this.getFileExtra(message)
      return extra.file_id || message.content.trim()
    },

    async downloadVisibleFile(message: ChatMessage) {
      const fileID = this.getFileDownloadID(message)
      if (!fileID) {
        this.errorMessage = '文件信息缺失，无法下载'
        return
      }

      this.errorMessage = ''
      try {
        const result = await downloadFile(fileID)
        triggerBrowserDownload(result.blob, this.getFileExtra(message).file_name || result.fileName || '文件')
      } catch (error) {
        this.errorMessage = getErrorMessage(error)
      }
    },

    mergeStoredRecallNotices(conversationID: string) {
      const notices = this.loadStoredRecallNotices().filter((notice) => notice.conversation_id === conversationID)
      if (notices.length === 0) {
        return
      }

      const noticeMessages = notices.map((notice) =>
        this.createRecallNoticeMessage(notice.message_id, notice.conversation_id, notice.recalled_at, notice.editable_until),
      )
      const noticeIDs = new Set(notices.map((notice) => notice.message_id))
      this.messages = sortChatMessages([
        ...this.messages.filter(
          (message) =>
            !noticeIDs.has(message.message_id) &&
            (!message.recalled_message_id || !noticeIDs.has(message.recalled_message_id)),
        ),
        ...noticeMessages,
      ])
    },

    createRecallNoticeMessage(
      messageID: string,
      conversationID: string,
      recalledAt: string,
      editableUntil: string,
    ): ChatMessage {
      return {
        client_msg_id: `recall-${messageID}`,
        message_id: `recall-${messageID}`,
        conversation_id: conversationID,
        sender_id: currentUserID(),
        message_type: 'system',
        content: '你撤回了一条消息',
        extra_json: {},
        send_status: 'sent',
        created_at: recalledAt,
        is_recall_notice: true,
        recalled_message_id: messageID,
        editable_until: editableUntil,
      }
    },

    persistRecallNotice(messageID: string, conversationID: string, recalledAt: string, editableUntil: string) {
      const userID = currentUserID()
      if (!userID || !editableUntil || !this.isBeforeEditableUntil(editableUntil)) {
        return
      }

      const notices = this.loadStoredRecallNotices().filter((notice) => notice.message_id !== messageID)
      notices.push({
        user_id: userID,
        conversation_id: conversationID,
        message_id: messageID,
        recalled_at: recalledAt,
        editable_until: editableUntil,
      })
      this.saveStoredRecallNotices(notices)
    },

    loadStoredRecallNotices() {
      const key = this.recallNoticeStorageKey()
      const userID = currentUserID()
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
          this.isBeforeEditableUntil(notice.editable_until),
      )
      if (validNotices.length !== parsed.length) {
        this.saveStoredRecallNotices(validNotices)
      }
      return validNotices
    },

    saveStoredRecallNotices(notices: StoredRecallNotice[]) {
      const key = this.recallNoticeStorageKey()
      if (!key) {
        return
      }
      if (notices.length === 0) {
        localStorage.removeItem(key)
        return
      }
      localStorage.setItem(key, JSON.stringify(notices))
    },

    removeStoredRecallNotice(messageID: string) {
      this.saveStoredRecallNotices(this.loadStoredRecallNotices().filter((notice) => notice.message_id !== messageID))
    },

    removeStoredRecallNoticesByConversation(conversationID: string) {
      this.saveStoredRecallNotices(
        this.loadStoredRecallNotices().filter((notice) => notice.conversation_id !== conversationID),
      )
    },

    expireRecallNotice(messageID: string) {
      this.removeStoredRecallNotice(messageID)
      this.messages = this.messages.map((message) => {
        if (message.recalled_message_id !== messageID) {
          return message
        }
        return {
          ...message,
          editable_until: '',
        }
      })
    },

    cleanupExpiredRecallNotices() {
      const notices = this.loadStoredRecallNotices()
      const activeNoticeIDs = new Set(notices.map((notice) => notice.message_id))
      this.messages = this.messages.map((message) => {
        if (!message.recalled_message_id || activeNoticeIDs.has(message.recalled_message_id)) {
          return message
        }
        return {
          ...message,
          editable_until: '',
        }
      })
    },

    startRecallNoticeExpiryTimer() {
      if (this.recallNoticeTimer) {
        return
      }
      this.recallNoticeTimer = window.setInterval(() => {
        this.recallNoticeNow = Date.now()
        this.cleanupExpiredRecallNotices()
      }, 30000)
    },

    stopRecallNoticeExpiryTimer() {
      if (!this.recallNoticeTimer) {
        return
      }
      window.clearInterval(this.recallNoticeTimer)
      this.recallNoticeTimer = undefined
    },

    isBeforeEditableUntil(value?: string) {
      const editableUntil = parseTimeValue(value || '')
      return editableUntil > this.recallNoticeNow
    },

    recallNoticeStorageKey() {
      const userID = currentUserID()
      return userID ? `${recallNoticeStoragePrefix}${userID}` : ''
    },
  },
})

function currentUserID() {
  return useAuthStore().user?.user_id || ''
}

function createLocalMessage(input: {
  clientMsgID: string
  conversationID: string
  senderID: string
  messageType: MessageType
  content: string
  extraJSON: Record<string, unknown>
}): ChatMessage {
  return {
    client_msg_id: input.clientMsgID,
    message_id: input.clientMsgID,
    conversation_id: input.conversationID,
    sender_id: input.senderID,
    message_type: input.messageType,
    content: input.content,
    extra_json: input.extraJSON,
    send_status: 'sending',
    created_at: new Date().toISOString(),
  }
}

function normalizeIncomingMessage(message: Message): ChatMessage {
  return {
    ...message,
    extra_json: message.extra_json || {},
    send_status: message.send_status || 'sent',
  }
}

function createLocalConversationFromMessage(message: ChatMessage, userID: string): Conversation {
  const peerUserID = message.sender_id && message.sender_id !== userID ? message.sender_id : ''
  return {
    conversation_id: message.conversation_id,
    conversation_type: 'private',
    title: peerUserID ? `用户 ${peerUserID}` : `会话 ${message.conversation_id}`,
    avatar_url: '',
    peer_user_id: peerUserID,
    peer_nickname: '',
    peer_avatar_url: '',
    group_id: '',
    group_no: '',
    group_status: '',
    last_message: null,
    unread_count: 0,
    is_pinned: false,
    is_muted: false,
  }
}

function appendOrMergeMessage(messages: ChatMessage[], incoming: ChatMessage) {
  const index = messages.findIndex((message) => isSameMessage(message, incoming))
  if (index === -1) {
    return [...messages, incoming]
  }

  const next = [...messages]
  next[index] = {
    ...next[index],
    ...incoming,
  }
  return next
}

function dedupeMessages(messages: ChatMessage[]) {
  return messages.reduce<ChatMessage[]>((items, message) => appendOrMergeMessage(items, message), [])
}

function isSameMessage(left: ChatMessage, right: ChatMessage) {
  if (left.message_id && right.message_id && left.message_id === right.message_id) {
    return true
  }
  return Boolean(left.client_msg_id && right.client_msg_id && left.client_msg_id === right.client_msg_id)
}

function rawMessageKeys(messageID: string, clientMsgID: string) {
  const keys: string[] = []
  if (messageID) {
    keys.push(`message:${messageID}`)
  }
  if (clientMsgID) {
    keys.push(`client:${clientMsgID}`)
  }
  return keys
}

function messageKeys(message: ChatMessage) {
  return rawMessageKeys(message.message_id, message.client_msg_id)
}

function sortChatMessages(items: ChatMessage[]) {
  return [...items].sort((left, right) => parseMessageTime(left) - parseMessageTime(right))
}

function parseMessageTime(message: ChatMessage) {
  return parseTimeValue(message.created_at)
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

function isPersistedMessage(message: ChatMessage) {
  return /^\d+$/.test(message.message_id)
}

function lastMessageContent(message: ChatMessage) {
  return message.message_type === 'file' ? '文件' : message.content
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

function getErrorMessage(error: unknown) {
  return error instanceof Error ? error.message : '请求失败'
}
