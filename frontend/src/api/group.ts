import { http, unwrap, type ApiResponse } from './http'

export interface PageResult<T> {
  list: T[]
  total: number
  page: number
  page_size: number
}

export interface GroupInfo {
  group_id: string
  group_no: string
  conversation_id: string
  owner_id: string
  name: string
  avatar_url: string
  max_members: number
  allow_member_invite: boolean
  status: string
  is_member?: boolean
}

export interface CreateGroupResult {
  group_id: string
  group_no: string
  conversation_id: string
}

export interface GroupJoinRequest {
  request_id: string
  group_id: string
  user_id: string
  user: {
    user_id: string
    username: string
    nickname: string
    avatar_url: string
  }
  message: string
  status: string
  handled_by: string
  created_at: string
  updated_at: string
}

export interface GroupMember {
  user_id: string
  nickname: string
  avatar_url: string
  bio: string
  role: 'owner' | 'admin' | 'member'
  mute_until: string | null
  joined_at: string
  status: string
  friendship_status: string
}

export async function createGroup(name: string, avatarURL = ''): Promise<CreateGroupResult> {
  const { data } = await http.post<ApiResponse<CreateGroupResult>>('/api/groups', {
    name,
    avatar_url: avatarURL,
  })
  return unwrap(data)
}

export async function searchGroups(keyword: string): Promise<PageResult<GroupInfo>> {
  const { data } = await http.get<ApiResponse<PageResult<GroupInfo>>>('/api/groups/search', {
    params: { keyword },
  })
  return unwrap(data)
}

export async function createJoinRequest(groupID: string, message: string): Promise<void> {
  const { data } = await http.post<ApiResponse<Record<string, never>>>(
    `/api/groups/${groupID}/join-requests`,
    { message },
  )
  unwrap(data)
}

export async function listJoinRequests(groupID: string): Promise<PageResult<GroupJoinRequest>> {
  const { data } = await http.get<ApiResponse<PageResult<GroupJoinRequest>>>(
    `/api/groups/${groupID}/join-requests`,
    { params: { page: 1, page_size: 50 } },
  )
  return unwrap(data)
}

export async function acceptJoinRequest(requestID: string): Promise<void> {
  const { data } = await http.post<ApiResponse<Record<string, never>>>(
    `/api/groups/join-requests/${requestID}/accept`,
  )
  unwrap(data)
}

export async function rejectJoinRequest(requestID: string): Promise<void> {
  const { data } = await http.post<ApiResponse<Record<string, never>>>(
    `/api/groups/join-requests/${requestID}/reject`,
  )
  unwrap(data)
}

export async function listGroupMembers(groupID: string): Promise<PageResult<GroupMember>> {
  const { data } = await http.get<ApiResponse<PageResult<GroupMember>>>(
    `/api/groups/${groupID}/members`,
    { params: { page: 1, page_size: 50 } },
  )
  return unwrap(data)
}

export async function setGroupAdmin(groupID: string, userID: string): Promise<void> {
  const { data } = await http.post<ApiResponse<Record<string, never>>>(
    `/api/groups/${groupID}/admins/${userID}`,
  )
  unwrap(data)
}

export async function unsetGroupAdmin(groupID: string, userID: string): Promise<void> {
  const { data } = await http.delete<ApiResponse<Record<string, never>>>(
    `/api/groups/${groupID}/admins/${userID}`,
  )
  unwrap(data)
}

export async function muteGroupMember(groupID: string, userID: string, muteUntil: string): Promise<void> {
  const { data } = await http.post<ApiResponse<Record<string, never>>>(
    `/api/groups/${groupID}/members/${userID}/mute`,
    { mute_until: muteUntil },
  )
  unwrap(data)
}

export async function unmuteGroupMember(groupID: string, userID: string): Promise<void> {
  const { data } = await http.delete<ApiResponse<Record<string, never>>>(
    `/api/groups/${groupID}/members/${userID}/mute`,
  )
  unwrap(data)
}

export async function updateGroupSettings(
  groupID: string,
  payload: { allow_member_invite?: boolean; max_members?: number },
): Promise<void> {
  const { data } = await http.put<ApiResponse<Record<string, never>>>(
    `/api/groups/${groupID}/settings`,
    payload,
  )
  unwrap(data)
}

export async function dissolveGroup(groupID: string): Promise<void> {
  const { data } = await http.delete<ApiResponse<Record<string, never>>>(`/api/groups/${groupID}`)
  unwrap(data)
}
