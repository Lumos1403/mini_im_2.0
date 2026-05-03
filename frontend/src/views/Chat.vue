<script setup lang="ts">
import { storeToRefs } from 'pinia'
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'

import type { FriendItem } from '../api/friend'
import type { GroupMember } from '../api/group'
import AppTopBar from '../components/chat/AppTopBar.vue'
import ChatMain from '../components/chat/ChatMain.vue'
import ConversationDetailPanel from '../components/chat/ConversationDetailPanel.vue'
import ConversationSidebar from '../components/chat/ConversationSidebar.vue'
import GlobalDrawer from '../components/common/GlobalDrawer.vue'
import FriendPanel from '../components/friend/FriendPanel.vue'
import GroupMemberDrawer from '../components/group/GroupMemberDrawer.vue'
import GroupMemberProfileModal from '../components/group/GroupMemberProfileModal.vue'
import GroupPanel from '../components/group/GroupPanel.vue'
import GlobalSearch from '../components/search/GlobalSearch.vue'
import AppShell from '../layouts/AppShell.vue'
import { useAuthStore } from '../stores/auth'
import { useChatStore } from '../stores/chat'
import { useFriendStore } from '../stores/friend'
import { useGroupStore } from '../stores/group'
import { logoutAndReset } from '../stores/session'
import { useWsStore } from '../stores/ws'

const router = useRouter()
const auth = useAuthStore()
const chat = useChatStore()
const ws = useWsStore()
const friendStore = useFriendStore()
const groupStore = useGroupStore()

const {
  conversations,
  activeConversationID,
  messages,
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

const showSearch = ref(false)
const showFriends = ref(false)
const showGroups = ref(false)

const activeConversation = computed(() => chat.activeConversation)
const activeGroupID = computed(() =>
  activeConversation.value?.conversation_type === 'group' ? activeConversation.value.group_id : '',
)
const canSend = computed(() => wsConnected.value && Boolean(activeConversationID.value) && chat.draft.trim().length > 0)
const canUploadFile = computed(
  () =>
    wsConnected.value &&
    Boolean(activeConversationID.value) &&
    activeConversation.value?.conversation_type !== 'group' &&
    !uploadingFile.value,
)
const isReady = computed(() => Boolean(auth.accessToken && auth.user))

onMounted(async () => {
  if (!isReady.value) {
    await router.replace('/login')
    return
  }

  await Promise.all([
    chat.initialize(),
    friendStore.loadFriends(),
    friendStore.loadReceivedRequests(),
  ])
  ws.connect()
  chat.startRecallNoticeExpiryTimer()
})

onBeforeUnmount(() => {
  ws.disconnect()
  chat.stopRecallNoticeExpiryTimer()
  showSearch.value = false
  showFriends.value = false
  showGroups.value = false
})

watch(activeGroupID, (groupID, oldGroupID) => {
  if (groupID === oldGroupID) {
    return
  }
  if (!groupID) {
    groupStore.closeMemberDrawer()
    groupStore.members = []
    groupStore.joinRequests = []
    return
  }
  void groupStore.loadMembers(groupID)
  void groupStore.loadJoinRequests(groupID)
  if (memberDrawerOpen.value) {
    groupStore.openMemberDrawer(groupID)
  }
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

async function handleFileSelected(selectedFile: File) {
  if (!wsConnected.value) {
    chat.errorMessage = 'WebSocket is not connected. File messages cannot be sent.'
    return
  }

  const payload = await chat.prepareUploadedFileMessage(selectedFile)
  if (payload) {
    ws.sendChatMessage(payload)
  }
}

function selectConversation(conversationID: string) {
  void chat.selectConversation(conversationID)
}

function openFriendChat(friend: FriendItem) {
  showFriends.value = false
  void chat.openFriendChat(friend)
}

async function openConversation(conversationID: string) {
  if (!conversationID) {
    return
  }
  if (!chat.conversations.some((item) => item.conversation_id === conversationID)) {
    await chat.loadConversationList(false)
  }
  await chat.selectConversation(conversationID)
  showSearch.value = false
  showGroups.value = false
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

function goProfile() {
  void router.push('/profile')
}

function handleLogout() {
  void logoutAndReset(router)
}
</script>

<template>
  <AppShell>
    <section v-if="isReady" class="chat-layout">
      <AppTopBar
        :connected="wsConnected"
        :pending-requests="friendStore.pendingRequestCount"
        @open-search="showSearch = true"
        @open-friends="showFriends = true"
        @open-groups="showGroups = true"
        @open-profile="goProfile"
        @logout="handleLogout"
      />

      <div class="chat-grid">
        <ConversationSidebar
          :conversations="conversations"
          :active-conversation-id="activeConversationID"
          :loading="loadingConversations"
          @select="selectConversation"
        />

        <ChatMain
          :active-conversation="activeConversation"
          :active-conversation-id="activeConversationID"
          :messages="messages"
          :loading-messages="loadingMessages"
          :has-more="hasMore"
          :error-message="errorMessage"
          :can-send="canSend"
          :can-upload-file="canUploadFile"
          :uploading-file="uploadingFile"
          :scroll-signal="scrollToBottomSignal"
          @send="sendMessage"
          @select-file="handleFileSelected"
          @load-older="chat.loadOlderMessages"
          @clear-current="chat.clearCurrentConversation"
          @open-members="openGroupMembers"
        />

        <ConversationDetailPanel
          :conversation="activeConversation"
          @open-members="openGroupMembers"
        />
      </div>

      <GlobalDrawer
        :open="showFriends"
        title="Friends"
        description="Contacts, requests, and user search"
        @close="showFriends = false"
      >
        <FriendPanel @open-chat="openFriendChat" />
      </GlobalDrawer>

      <GlobalDrawer
        :open="showGroups"
        title="Groups"
        description="Create, find, join, and manage groups"
        @close="showGroups = false"
      >
        <GroupPanel @open-conversation="openConversation" />
      </GlobalDrawer>

      <GlobalSearch
        :open="showSearch"
        @close="showSearch = false"
        @open-conversation="openConversation"
      />

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
    </section>
  </AppShell>
</template>

<style scoped>
.chat-layout {
  display: grid;
  height: 100%;
  min-height: 0;
  grid-template-rows: auto minmax(0, 1fr);
}

.chat-grid {
  display: grid;
  min-height: 0;
  grid-template-columns: minmax(240px, 292px) minmax(0, 1fr) minmax(280px, 340px);
  overflow: hidden;
}

@media (max-width: 1160px) {
  .chat-grid {
    grid-template-columns: minmax(220px, 260px) minmax(0, 1fr);
  }

  .chat-grid :deep(.detail-panel) {
    display: none;
  }
}

@media (max-width: 760px) {
  .chat-grid {
    grid-template-columns: 1fr;
    grid-template-rows: minmax(150px, 26%) minmax(0, 1fr);
  }

  .chat-grid :deep(.conversation-sidebar) {
    border-right: 0;
    border-bottom: 1px solid var(--border);
  }
}
</style>
