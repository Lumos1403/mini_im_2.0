<script setup lang="ts">
import type { Conversation } from '../../api/conversation'
import { useChatStore } from '../../stores/chat'

defineProps<{
  conversations: Conversation[]
  activeConversationId: string
  loading: boolean
}>()

const emit = defineEmits<{
  (event: 'select', conversationID: string): void
}>()

const chat = useChatStore()

function avatarText(conversation: Conversation) {
  return (conversation.title || conversation.peer_nickname || conversation.group_no || '#').slice(0, 1).toUpperCase()
}
</script>

<template>
  <aside class="conversation-sidebar">
    <header>
      <div>
        <h2>Conversations</h2>
        <small>{{ conversations.length }} active threads</small>
      </div>
    </header>

    <div v-if="loading" class="state-block">Loading conversations...</div>

    <button
      v-for="conversation in conversations"
      :key="conversation.conversation_id"
      :class="['conversation-row', { active: conversation.conversation_id === activeConversationId }]"
      type="button"
      @click="emit('select', conversation.conversation_id)"
    >
      <span class="avatar">{{ avatarText(conversation) }}</span>
      <span class="meta">
        <span class="title-line">
          <strong>{{ conversation.title || 'Untitled conversation' }}</strong>
          <span v-if="conversation.unread_count > 0" class="unread">
            {{ chat.formatUnreadCount(conversation.unread_count) }}
          </span>
        </span>
        <small>{{ chat.formatConversationLastMessage(conversation.last_message) }}</small>
      </span>
    </button>

    <div v-if="!loading && conversations.length === 0" class="state-block">
      No conversations yet.
    </div>
  </aside>
</template>

<style scoped>
.conversation-sidebar {
  display: flex;
  min-width: 0;
  min-height: 0;
  flex-direction: column;
  gap: 8px;
  overflow-y: auto;
  border-right: 1px solid var(--border);
  padding: 18px 12px;
  background: rgba(12, 14, 15, 0.68);
  backdrop-filter: blur(18px);
}

header {
  display: flex;
  min-height: 48px;
  align-items: center;
  padding: 0 8px 8px;
}

h2 {
  margin: 0;
  color: #fff7e8;
  font-size: 17px;
}

header small {
  display: block;
  margin-top: 4px;
  color: var(--text-muted);
  font-size: 12px;
}

.conversation-row {
  display: flex;
  width: 100%;
  min-height: 64px;
  align-items: center;
  gap: 10px;
  border: 1px solid transparent;
  border-radius: 10px;
  padding: 9px;
  color: var(--text-soft);
  background: transparent;
  cursor: pointer;
  text-align: left;
}

.conversation-row:hover,
.conversation-row.active {
  border-color: rgba(240, 207, 132, 0.2);
  background: rgba(255, 255, 255, 0.07);
}

.conversation-row.active {
  box-shadow: 0 0 0 1px rgba(240, 207, 132, 0.08) inset;
}

.avatar {
  display: grid;
  width: 42px;
  height: 42px;
  flex: 0 0 42px;
  place-items: center;
  border: 1px solid rgba(240, 207, 132, 0.16);
  border-radius: 12px;
  color: #19130b;
  background: linear-gradient(135deg, #e5c26e, #876c34);
  font-weight: 850;
}

.meta {
  min-width: 0;
  flex: 1 1 auto;
}

.title-line {
  display: flex;
  min-width: 0;
  align-items: center;
  gap: 8px;
}

.title-line strong,
.meta small {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.title-line strong {
  min-width: 0;
  flex: 1 1 auto;
  color: var(--text);
  font-size: 14px;
}

.meta small {
  margin-top: 4px;
  color: var(--text-muted);
  font-size: 12px;
}

.unread {
  display: grid;
  min-width: 20px;
  height: 20px;
  flex: 0 0 auto;
  place-items: center;
  border-radius: 999px;
  padding: 0 6px;
  color: #210f0f;
  background: #f87171;
  font-size: 11px;
  font-weight: 820;
}

.state-block {
  margin: 8px;
  border: 1px dashed rgba(240, 207, 132, 0.18);
  border-radius: 10px;
  padding: 24px 12px;
  color: var(--text-muted);
  background: rgba(255, 255, 255, 0.035);
  text-align: center;
  font-size: 13px;
}
</style>
