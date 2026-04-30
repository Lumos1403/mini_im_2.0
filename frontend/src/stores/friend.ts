import { defineStore } from 'pinia'

import * as friendApi from '../api/friend'
import { getApiErrorMessage } from '../api/http'
import { searchUsers, type UserSearchResult } from '../api/user'
import type { FriendItem, FriendRequest } from '../api/friend'
import { useChatStore } from './chat'

interface FriendState {
  friends: FriendItem[]
  receivedRequests: FriendRequest[]
  searchResults: UserSearchResult[]
  loadingFriends: boolean
  loadingRequests: boolean
  searching: boolean
  operating: boolean
  errorMessage: string
  noticeMessage: string
}

export const useFriendStore = defineStore('friend', {
  state: (): FriendState => ({
    friends: [],
    receivedRequests: [],
    searchResults: [],
    loadingFriends: false,
    loadingRequests: false,
    searching: false,
    operating: false,
    errorMessage: '',
    noticeMessage: '',
  }),
  getters: {
    pendingRequestCount(state) {
      return state.receivedRequests.filter((item) => item.status === 'pending').length
    },
  },
  actions: {
    async loadFriends() {
      this.loadingFriends = true
      this.clearMessages()
      try {
        const result = await friendApi.listFriends()
        this.friends = result.list
      } catch (error) {
        this.errorMessage = getApiErrorMessage(error)
      } finally {
        this.loadingFriends = false
      }
    },
    async loadReceivedRequests() {
      this.loadingRequests = true
      this.clearMessages()
      try {
        const result = await friendApi.listFriendRequests('received')
        this.receivedRequests = result.list
      } catch (error) {
        this.errorMessage = getApiErrorMessage(error)
      } finally {
        this.loadingRequests = false
      }
    },
    async search(keyword: string) {
      const trimmed = keyword.trim()
      if (!trimmed) {
        this.searchResults = []
        this.errorMessage = '请输入 user_id 或昵称'
        this.noticeMessage = ''
        return
      }

      this.searching = true
      this.clearMessages()
      try {
        const result = await searchUsers(trimmed)
        this.searchResults = result.list
        if (result.list.length === 0) {
          this.noticeMessage = '没有找到匹配用户'
        }
      } catch (error) {
        this.searchResults = []
        this.errorMessage = getApiErrorMessage(error)
      } finally {
        this.searching = false
      }
    },
    async sendFriendRequest(toUserID: string, message: string) {
      this.operating = true
      this.clearMessages()
      try {
        await friendApi.createFriendRequest(toUserID, message.trim())
        this.noticeMessage = '好友申请已发送'
      } catch (error) {
        this.errorMessage = getApiErrorMessage(error)
      } finally {
        this.operating = false
      }
    },
    async acceptRequest(requestID: string) {
      this.operating = true
      this.clearMessages()
      try {
        await friendApi.acceptFriendRequest(requestID)
        await Promise.all([this.loadReceivedRequests(), this.loadFriends()])
        this.noticeMessage = '已同意好友申请'
      } catch (error) {
        this.errorMessage = getApiErrorMessage(error)
      } finally {
        this.operating = false
      }
    },
    async rejectRequest(requestID: string) {
      this.operating = true
      this.clearMessages()
      try {
        await friendApi.rejectFriendRequest(requestID)
        await this.loadReceivedRequests()
        this.noticeMessage = '已拒绝好友申请'
      } catch (error) {
        this.errorMessage = getApiErrorMessage(error)
      } finally {
        this.operating = false
      }
    },
    async removeFriend(userID: string, conversationID = '') {
      this.operating = true
      this.clearMessages()
      try {
        await friendApi.deleteFriend(userID)
        const chat = useChatStore()
        chat.removeConversation(conversationID)
        chat.removePrivateConversationByPeer(userID)
        await Promise.all([this.loadFriends(), chat.loadConversationList(false)])
        this.noticeMessage = '好友已删除'
      } catch (error) {
        this.errorMessage = getApiErrorMessage(error)
      } finally {
        this.operating = false
      }
    },
    async block(userID: string) {
      this.operating = true
      this.clearMessages()
      try {
        await friendApi.blockFriend(userID)
        await this.loadFriends()
        this.noticeMessage = '已拉黑好友'
      } catch (error) {
        this.errorMessage = getApiErrorMessage(error)
      } finally {
        this.operating = false
      }
    },
    async unblock(userID: string) {
      this.operating = true
      this.clearMessages()
      try {
        await friendApi.unblockFriend(userID)
        await this.loadFriends()
        this.noticeMessage = '已解除拉黑'
      } catch (error) {
        this.errorMessage = getApiErrorMessage(error)
      } finally {
        this.operating = false
      }
    },
    clearMessages() {
      this.errorMessage = ''
      this.noticeMessage = ''
    },
  },
})
