import { http, unwrap, type ApiResponse } from './http'
import type { PageResult, UserSearchResult } from './user'

export type FriendRequestDirection = 'received' | 'sent'
export type FriendRequestStatus = 'pending' | 'accepted' | 'rejected' | 'expired'

export type FriendUser = UserSearchResult

export interface FriendRequest {
  request_id: string
  from_user_id: string
  to_user_id: string
  user: FriendUser
  message: string
  status: FriendRequestStatus
  created_at: string
  updated_at: string
}

export interface FriendRequestResult {
  request_id: string
  status: FriendRequestStatus
}

export interface FriendItem {
  friend_user_id: string
  nickname: string
  avatar_url: string
  bio: string
  conversation_id: string
  is_blocked_by_me: boolean
  created_at: string
  updated_at: string
}

export async function createFriendRequest(toUserID: string, message: string): Promise<FriendRequestResult> {
  const { data } = await http.post<ApiResponse<FriendRequestResult>>('/api/friends/requests', {
    to_user_id: toUserID,
    message,
  })
  return unwrap(data)
}

export async function listFriendRequests(
  direction: FriendRequestDirection = 'received',
  page = 1,
  pageSize = 20,
): Promise<PageResult<FriendRequest>> {
  const { data } = await http.get<ApiResponse<PageResult<FriendRequest>>>('/api/friends/requests', {
    params: {
      direction,
      page,
      page_size: pageSize,
    },
  })
  return unwrap(data)
}

export async function acceptFriendRequest(requestID: string): Promise<void> {
  const { data } = await http.post<ApiResponse<Record<string, never>>>(
    `/api/friends/requests/${requestID}/accept`,
  )
  unwrap(data)
}

export async function rejectFriendRequest(requestID: string): Promise<void> {
  const { data } = await http.post<ApiResponse<Record<string, never>>>(
    `/api/friends/requests/${requestID}/reject`,
  )
  unwrap(data)
}

export async function listFriends(page = 1, pageSize = 50): Promise<PageResult<FriendItem>> {
  const { data } = await http.get<ApiResponse<PageResult<FriendItem>>>('/api/friends', {
    params: {
      page,
      page_size: pageSize,
    },
  })
  return unwrap(data)
}

export async function deleteFriend(userID: string): Promise<void> {
  const { data } = await http.delete<ApiResponse<Record<string, never>>>(`/api/friends/${userID}`)
  unwrap(data)
}

export async function blockFriend(userID: string): Promise<void> {
  const { data } = await http.post<ApiResponse<Record<string, never>>>(`/api/friends/${userID}/block`)
  unwrap(data)
}

export async function unblockFriend(userID: string): Promise<void> {
  const { data } = await http.delete<ApiResponse<Record<string, never>>>(`/api/friends/${userID}/block`)
  unwrap(data)
}
