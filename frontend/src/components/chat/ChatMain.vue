<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { computed, nextTick, ref, watch } from 'vue'

import type { Conversation } from '../../api/conversation'
import GroupRoleBadge from '../group/GroupRoleBadge.vue'
import { useChatStore, type ChatMessage } from '../../stores/chat'
import StreamingMessageContent from './StreamingMessageContent.vue'

const props = defineProps<{
  activeConversation?: Conversation
  activeConversationId: string
  messages: ChatMessage[]
  loadingMessages: boolean
  hasMore: boolean
  errorMessage: string
  canSend: boolean
  canUploadFile: boolean
  uploadingFile: boolean
  scrollSignal: number
}>()

const emit = defineEmits<{
  (event: 'send'): void
  (event: 'select-file', file: File): void
  (event: 'load-older'): void
  (event: 'clear-current'): void
  (event: 'open-members'): void
}>()

const chat = useChatStore()
const { draft } = storeToRefs(chat)
const fileInput = ref<HTMLInputElement | null>(null)
const messageArea = ref<HTMLElement | null>(null)

const isGroupConversation = computed(() => props.activeConversation?.conversation_type === 'group')
const activeGroupNo = computed(() => props.activeConversation?.group_no || '')

watch(
  () => props.scrollSignal,
  () => {
    void scrollToBottom()
  },
)

watch(
  () => props.activeConversationId,
  () => {
    void scrollToBottom('auto')
  },
)

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

function triggerFileSelect() {
  if (props.canUploadFile) {
    fileInput.value?.click()
  }
}

function handleFileSelected(event: Event) {
  const input = event.target as HTMLInputElement
  const selectedFile = input.files?.[0]
  input.value = ''
  if (selectedFile) {
    emit('select-file', selectedFile)
  }
}

function groupSenderName(message: ChatMessage) {
  if (message.sender_nickname) {
    return message.sender_nickname
  }
  if (chat.isMine(message)) {
    return 'You'
  }
  return message.sender_id
}

function handleStreamingTyped() {
  void scrollToBottom('auto')
}
</script>

<template>
  <section class="chat-main-panel">
    <header class="chat-header">
      <div class="chat-title">
        <strong>{{ activeConversation?.title || 'Select a conversation' }}</strong>
        <small v-if="activeConversation">
          {{ isGroupConversation ? `Group ${activeGroupNo || activeConversation.group_id}` : 'Private conversation' }}
        </small>
        <small v-else>No chat data is rendered until a conversation is selected.</small>
      </div>
      <div class="chat-actions">
        <span v-if="errorMessage" class="inline-error">{{ errorMessage }}</span>
        <button
          v-if="isGroupConversation"
          class="secondary-button"
          type="button"
          :disabled="!activeConversation?.group_id"
          @click="emit('open-members')"
        >
          Members
        </button>
        <button
          class="danger-button"
          type="button"
          :disabled="!activeConversationId || messages.length === 0"
          @click="emit('clear-current')"
        >
          Clear
        </button>
      </div>
    </header>

    <div ref="messageArea" class="message-area">
      <button
        v-if="hasMore"
        class="load-more"
        type="button"
        :disabled="loadingMessages"
        @click="emit('load-older')"
      >
        {{ loadingMessages ? 'Loading...' : 'Load earlier messages' }}
      </button>

      <div v-if="loadingMessages && messages.length === 0" class="state-block">Loading messages...</div>
      <div v-else-if="!activeConversationId" class="state-block">
        Choose a conversation from the left sidebar.
      </div>
      <div v-else-if="messages.length === 0" class="state-block">
        No visible messages in this conversation.
      </div>

      <article
        v-for="message in messages"
        :key="message.client_msg_id || message.message_id"
        :class="['message-row', { mine: chat.isMine(message), notice: message.is_recall_notice }]"
      >
        <div v-if="message.is_recall_notice" class="recall-notice">
          <span>{{ message.content }}</span>
          <button v-if="chat.canReEdit(message)" type="button" @click="chat.reEditMessage(message)">
            Re-edit
          </button>
        </div>

        <div v-else class="bubble-wrap">
          <span
            v-if="chat.isMine(message) && chat.isFailed(message)"
            class="failed-mark"
            :title="message.error_message || 'Send failed'"
          >
            !
          </span>
          <div class="message-bubble">
            <span v-if="isGroupConversation" class="sender-line">
              <span class="sender-name">{{ groupSenderName(message) }}</span>
              <GroupRoleBadge :role="message.sender_group_role || 'member'" />
            </span>

            <div v-if="message.message_type === 'file'" class="file-card">
              <span class="file-icon">FILE</span>
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
                Download
              </button>
            </div>
            <StreamingMessageContent
              v-else
              :content="message.content"
              :streaming="message.stream_status === 'streaming'"
              :error="message.stream_status === 'error'"
              :error-message="message.error_message"
              @typed="handleStreamingTyped"
            />

            <small>
              {{
                message.stream_status === 'streaming'
                  ? 'Generating...'
                  : message.send_status === 'sending'
                    ? 'Sending...'
                    : message.created_at
              }}
            </small>
            <div class="message-actions">
              <button v-if="chat.canDelete(message)" type="button" @click="chat.deleteVisibleMessage(message)">
                Delete
              </button>
              <button v-if="chat.canRecall(message)" type="button" @click="chat.recallVisibleMessage(message)">
                Recall
              </button>
            </div>
          </div>
        </div>
      </article>
    </div>

    <footer class="composer">
      <input ref="fileInput" class="hidden-file-input" type="file" @change="handleFileSelected" />
      <button class="secondary-button" type="button" :disabled="!canUploadFile" @click="triggerFileSelect">
        {{ uploadingFile ? 'Uploading...' : 'File' }}
      </button>
      <textarea
        v-model="draft"
        maxlength="2000"
        placeholder="Type a message"
        :disabled="!activeConversationId"
        @keydown.enter.exact.prevent="emit('send')"
      ></textarea>
      <button class="send-button" type="button" :disabled="!canSend" @click="emit('send')">Send</button>
    </footer>
  </section>
</template>

<style scoped>
.chat-main-panel {
  display: flex;
  min-width: 0;
  min-height: 0;
  flex-direction: column;
  background: rgba(18, 21, 22, 0.58);
}

.chat-header,
.composer {
  display: flex;
  flex: 0 0 auto;
  align-items: center;
  gap: 12px;
  border-bottom: 1px solid var(--border);
  padding: 12px 18px;
  background: rgba(15, 17, 18, 0.7);
  backdrop-filter: blur(18px);
}

.chat-header {
  min-height: 64px;
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

.chat-title strong {
  color: #fff7e8;
  font-size: 17px;
}

.chat-title small {
  margin-top: 4px;
  color: var(--text-muted);
  font-size: 12px;
}

.chat-actions {
  display: flex;
  min-width: 0;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
}

.message-area {
  flex: 1 1 auto;
  min-height: 0;
  overflow-y: auto;
  padding: 22px min(4vw, 40px);
}

.message-row {
  display: flex;
  margin-bottom: 14px;
}

.message-row.mine {
  justify-content: flex-end;
}

.message-row.notice {
  justify-content: center;
}

.bubble-wrap {
  display: flex;
  max-width: min(680px, 78%);
  align-items: center;
  gap: 9px;
}

.message-bubble {
  min-width: 0;
  border: 1px solid rgba(240, 207, 132, 0.1);
  border-radius: 14px;
  padding: 11px 13px;
  color: var(--text);
  background: rgba(255, 255, 255, 0.075);
  box-shadow: 0 10px 24px rgba(0, 0, 0, 0.22);
}

.message-row.mine .message-bubble {
  border-color: rgba(240, 207, 132, 0.22);
  color: #171009;
  background: linear-gradient(135deg, #e9c97d, #c49a4f);
}

.message-bubble p {
  margin: 0;
  white-space: pre-wrap;
  overflow-wrap: anywhere;
  line-height: 1.55;
}

.text-message {
  overflow-wrap: anywhere;
  line-height: 1.55;
}

.text-message-copy {
  white-space: pre-wrap;
}

.markdown-image-link {
  display: block;
  margin: 10px 0 2px;
}

.markdown-image {
  display: block;
  width: min(480px, 100%);
  max-height: 360px;
  border: 1px solid rgba(240, 207, 132, 0.14);
  border-radius: 10px;
  background: #fff;
  object-fit: contain;
}

.message-row.mine .markdown-image {
  border-color: rgba(23, 16, 9, 0.18);
}

.sender-line {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 6px;
  margin-bottom: 6px;
  color: rgba(243, 239, 229, 0.72);
  font-size: 12px;
  font-weight: 800;
}

.sender-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.message-row.mine .sender-line {
  color: rgba(23, 16, 9, 0.72);
}

.message-bubble small {
  display: block;
  margin-top: 7px;
  color: var(--text-muted);
  font-size: 11px;
}

.message-row.mine .message-bubble small {
  color: rgba(23, 16, 9, 0.64);
}

.file-card {
  display: flex;
  width: min(380px, 72vw);
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
  border-radius: 10px;
  color: var(--accent-strong);
  background: rgba(0, 0, 0, 0.28);
  font-size: 11px;
  font-weight: 900;
}

.message-row.mine .file-icon {
  color: #171009;
  background: rgba(255, 255, 255, 0.28);
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
  text-overflow: ellipsis;
  white-space: nowrap;
}

.file-download-button,
.message-actions button,
.recall-notice button,
.secondary-button,
.send-button,
.load-more,
.danger-button {
  height: 36px;
  border: 1px solid rgba(240, 207, 132, 0.16);
  border-radius: 9px;
  padding: 0 11px;
  color: var(--text-soft);
  background: rgba(255, 255, 255, 0.07);
  cursor: pointer;
  font-size: 13px;
  font-weight: 760;
}

.file-download-button:hover,
.message-actions button:hover,
.recall-notice button:hover,
.secondary-button:hover,
.load-more:hover {
  border-color: rgba(240, 207, 132, 0.3);
  color: var(--text);
  background: rgba(255, 255, 255, 0.1);
}

.send-button {
  min-width: 76px;
  color: #171009;
  background: linear-gradient(135deg, #f0cf84, #b68a3e);
}

.danger-button {
  color: #ffd8d8;
  border-color: rgba(248, 113, 113, 0.32);
  background: rgba(239, 68, 68, 0.12);
}

.danger-button:hover {
  color: #fff;
  border-color: rgba(248, 113, 113, 0.52);
  background: rgba(239, 68, 68, 0.2);
}

.failed-mark {
  display: grid;
  width: 22px;
  height: 22px;
  place-items: center;
  border-radius: 50%;
  color: #2b0c0c;
  background: var(--danger);
  box-shadow: 0 0 20px rgba(248, 113, 113, 0.4);
  font-size: 13px;
  font-weight: 900;
}

.message-actions {
  display: flex;
  justify-content: flex-end;
  gap: 7px;
  margin-top: 8px;
}

.message-actions button {
  height: 28px;
  padding: 0 8px;
  font-size: 12px;
}

.message-row.mine .message-actions button,
.message-row.mine .file-download-button {
  color: #171009;
  border-color: rgba(23, 16, 9, 0.18);
  background: rgba(255, 255, 255, 0.24);
}

.recall-notice {
  display: inline-flex;
  max-width: 90%;
  align-items: center;
  gap: 10px;
  border: 1px solid rgba(240, 207, 132, 0.12);
  border-radius: 999px;
  padding: 8px 12px;
  color: var(--text-muted);
  background: rgba(255, 255, 255, 0.06);
  font-size: 13px;
}

.load-more {
  display: block;
  margin: 0 auto 18px;
}

.state-block {
  max-width: 420px;
  margin: 36px auto;
  border: 1px dashed rgba(240, 207, 132, 0.16);
  border-radius: 14px;
  padding: 28px 18px;
  color: var(--text-muted);
  background: rgba(255, 255, 255, 0.04);
  text-align: center;
}

.inline-error {
  max-width: min(360px, 28vw);
  overflow: hidden;
  color: #ffd4d4;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 13px;
}

.composer {
  min-height: 86px;
  border-top: 1px solid var(--border);
  border-bottom: 0;
}

.hidden-file-input {
  display: none;
}

textarea {
  width: 100%;
  height: 54px;
  min-width: 0;
  resize: none;
  border: 1px solid rgba(240, 207, 132, 0.16);
  border-radius: 12px;
  padding: 11px 13px;
  color: var(--text);
  background: rgba(0, 0, 0, 0.22);
}

textarea:hover,
textarea:focus {
  border-color: rgba(240, 207, 132, 0.34);
}

@media (max-width: 760px) {
  .chat-header,
  .composer {
    align-items: stretch;
    flex-wrap: wrap;
  }

  .chat-actions {
    justify-content: flex-start;
  }

  .bubble-wrap {
    max-width: 94%;
  }
}
</style>
