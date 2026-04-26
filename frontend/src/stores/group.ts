import { defineStore } from 'pinia'

import {
  acceptJoinRequest,
  createGroup,
  createJoinRequest,
  dissolveGroup,
  listGroupMembers,
  listJoinRequests,
  muteGroupMember,
  rejectJoinRequest,
  searchGroups,
  setGroupAdmin,
  unmuteGroupMember,
  unsetGroupAdmin,
  updateGroupSettings,
  type GroupInfo,
  type GroupJoinRequest,
  type GroupMember,
  type GroupFriendshipStatus,
} from '../api/group'
import { createFriendRequest } from '../api/friend'
import { getApiErrorMessage } from '../api/http'
import { useChatStore } from './chat'

interface GroupState {
  searchResults: GroupInfo[]
  joinRequests: GroupJoinRequest[]
  members: GroupMember[]
  membersGroupID: string
  loadingMembers: boolean
  memberDrawerOpen: boolean
  selectedMemberID: string
  memberProfileOpen: boolean
  friendRequestingUserID: string
  loading: boolean
  operating: boolean
  errorMessage: string
  noticeMessage: string
}

export const useGroupStore = defineStore('group', {
  state: (): GroupState => ({
    searchResults: [],
    joinRequests: [],
    members: [],
    membersGroupID: '',
    loadingMembers: false,
    memberDrawerOpen: false,
    selectedMemberID: '',
    memberProfileOpen: false,
    friendRequestingUserID: '',
    loading: false,
    operating: false,
    errorMessage: '',
    noticeMessage: '',
  }),
  getters: {
    selectedMember(state): GroupMember | null {
      if (!state.selectedMemberID) {
        return null
      }
      return state.members.find((item) => item.user_id === state.selectedMemberID) || null
    },
  },
  actions: {
    async create(name: string) {
      const trimmed = name.trim()
      if (!trimmed) {
        this.errorMessage = 'Group name is required'
        return ''
      }
      return this.withOperation(async () => {
        const result = await createGroup(trimmed)
        this.noticeMessage = `Group created: ${result.group_no}`
        const chat = useChatStore()
        await chat.loadConversationList(false)
        await chat.selectConversation(result.conversation_id)
        return result.conversation_id
      })
    },

    async search(keyword: string) {
      const trimmed = keyword.trim()
      if (!trimmed) {
        this.searchResults = []
        return
      }
      this.loading = true
      this.clearMessages()
      try {
        const result = await searchGroups(trimmed)
        this.searchResults = result.list
      } catch (error) {
        this.errorMessage = getErrorMessage(error)
      } finally {
        this.loading = false
      }
    },

    async applyJoin(groupID: string, message: string) {
      await this.withOperation(async () => {
        await createJoinRequest(groupID, message.trim())
        this.noticeMessage = 'Join request sent'
      })
    },

    async loadJoinRequests(groupID: string) {
      if (!groupID) {
        this.joinRequests = []
        return
      }
      this.loading = true
      this.clearMessages()
      try {
        const result = await listJoinRequests(groupID)
        this.joinRequests = result.list
      } catch (error) {
        this.joinRequests = []
        this.errorMessage = getErrorMessage(error)
      } finally {
        this.loading = false
      }
    },

    async accept(request: GroupJoinRequest) {
      await this.withOperation(async () => {
        await acceptJoinRequest(request.request_id)
        this.noticeMessage = 'Join request accepted'
        await this.loadJoinRequests(request.group_id)
        await this.loadMembers(request.group_id)
      })
    },

    async reject(request: GroupJoinRequest) {
      await this.withOperation(async () => {
        await rejectJoinRequest(request.request_id)
        this.noticeMessage = 'Join request rejected'
        await this.loadJoinRequests(request.group_id)
      })
    },

    async loadMembers(groupID: string, options: { silent?: boolean } = {}) {
      if (!groupID) {
        this.members = []
        this.membersGroupID = ''
        return
      }
      if (this.membersGroupID !== groupID) {
        this.members = []
      }
      this.membersGroupID = groupID
      this.loading = true
      this.loadingMembers = true
      if (!options.silent) {
        this.clearMessages()
      }
      try {
        const result = await listGroupMembers(groupID)
        if (this.membersGroupID === groupID) {
          this.members = result.list
        }
      } catch (error) {
        if (this.membersGroupID === groupID) {
          this.members = []
        }
        this.errorMessage = getErrorMessage(error)
      } finally {
        this.loading = false
        this.loadingMembers = false
      }
    },

    openMemberDrawer(groupID: string) {
      if (!groupID) {
        return
      }
      this.memberDrawerOpen = true
      void this.loadMembers(groupID)
    },

    closeMemberDrawer() {
      this.memberDrawerOpen = false
      this.closeMemberProfile()
    },

    openMemberProfile(member: GroupMember) {
      this.selectedMemberID = member.user_id
      this.memberProfileOpen = true
    },

    closeMemberProfile() {
      this.memberProfileOpen = false
      this.selectedMemberID = ''
    },

    async sendFriendRequestFromMember(groupID: string, userID: string, message: string) {
      if (!groupID || !userID || this.friendRequestingUserID === userID) {
        return
      }
      this.friendRequestingUserID = userID
      this.clearMessages()
      try {
        await createFriendRequest(userID, message.trim())
        this.updateMemberFriendshipStatus(userID, 'pending_sent')
        this.noticeMessage = '好友申请已发送'
        await this.loadMembers(groupID, { silent: true })
      } catch (error) {
        this.errorMessage = getErrorMessage(error)
        await this.loadMembers(groupID, { silent: true })
      } finally {
        if (this.friendRequestingUserID === userID) {
          this.friendRequestingUserID = ''
        }
      }
    },

    updateMemberFriendshipStatus(userID: string, status: GroupFriendshipStatus) {
      this.members = this.members.map((member) =>
        member.user_id === userID
          ? {
              ...member,
              friendship_status: status,
            }
          : member,
      )
    },

    async setAdmin(groupID: string, userID: string) {
      await this.withOperation(async () => {
        await setGroupAdmin(groupID, userID)
        await this.loadMembers(groupID)
      })
    },

    async unsetAdmin(groupID: string, userID: string) {
      await this.withOperation(async () => {
        await unsetGroupAdmin(groupID, userID)
        await this.loadMembers(groupID)
      })
    },

    async mute(groupID: string, userID: string) {
      const muteUntil = new Date(Date.now() + 10 * 60 * 1000)
      await this.withOperation(async () => {
        await muteGroupMember(groupID, userID, formatBackendTime(muteUntil))
        await this.loadMembers(groupID)
      })
    },

    async unmute(groupID: string, userID: string) {
      await this.withOperation(async () => {
        await unmuteGroupMember(groupID, userID)
        await this.loadMembers(groupID)
      })
    },

    async saveInviteSetting(groupID: string, allowMemberInvite: boolean) {
      await this.withOperation(async () => {
        await updateGroupSettings(groupID, { allow_member_invite: allowMemberInvite })
        this.noticeMessage = 'Invite setting saved'
      })
    },

    async saveMaxMembers(groupID: string, maxMembers: number) {
      await this.withOperation(async () => {
        await updateGroupSettings(groupID, { max_members: maxMembers })
        this.noticeMessage = 'Max members saved'
      })
    },

    async dissolve(groupID: string) {
      await this.withOperation(async () => {
        await dissolveGroup(groupID)
        this.noticeMessage = 'Group dissolved'
        await useChatStore().loadConversationList(false)
      })
    },

    clearMessages() {
      this.errorMessage = ''
      this.noticeMessage = ''
    },

    async withOperation<T>(operation: () => Promise<T>): Promise<T | undefined> {
      this.operating = true
      this.clearMessages()
      try {
        return await operation()
      } catch (error) {
        this.errorMessage = getErrorMessage(error)
        return undefined
      } finally {
        this.operating = false
      }
    },
  },
})

function formatBackendTime(value: Date) {
  const pad = (part: number) => String(part).padStart(2, '0')
  return `${value.getFullYear()}-${pad(value.getMonth() + 1)}-${pad(value.getDate())} ${pad(value.getHours())}:${pad(value.getMinutes())}:${pad(value.getSeconds())}`
}

function getErrorMessage(error: unknown) {
  return getApiErrorMessage(error)
}
