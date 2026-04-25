import { http, unwrap, type ApiResponse } from './http'

export interface PageResult<T> {
  list: T[]
  total: number
  page: number
  page_size: number
}

export interface Conversation {
  conversation_id: string
  conversation_type: string
  title: string
  avatar_url: string
  peer_user_id: string
  peer_nickname: string
  peer_avatar_url: string
  group_id: string
  group_no: string
  group_status: string
  last_message: {
    content: string
    message_type: string
    created_at: string
  } | null
  unread_count: number
  is_pinned: boolean
  is_muted: boolean
}

export type MessageType = 'text' | 'emoji' | 'file' | 'system'

export interface FileMessageExtra {
  file_id?: string
  file_name?: string
  file_size?: number
  mime_type?: string
}

export interface Message {
  client_msg_id: string
  message_id: string
  conversation_id: string
  sender_id: string
  sender_nickname?: string
  sender_avatar_url?: string
  sender_group_role?: 'owner' | 'admin' | 'member'
  message_type: MessageType
  content: string
  extra_json: Record<string, unknown>
  send_status: string
  created_at: string
}

export interface MessagePage {
  list: Message[]
  next_cursor: string
  has_more: boolean
  limit: number
}

export interface RecallMessageResult {
  message_id: string
  editable_until: string
}

export interface RecallEditCache {
  message_id: string
  content: string
}

export async function listConversations(page = 1, pageSize = 50): Promise<PageResult<Conversation>> {
  const { data } = await http.get<ApiResponse<PageResult<Conversation>>>('/api/conversations', {
    params: {
      page,
      page_size: pageSize,
    },
  })
  return unwrap(data)
}

export async function listMessages(conversationID: string, cursor = '', limit = 30): Promise<MessagePage> {
  const { data } = await http.get<ApiResponse<MessagePage>>(`/api/conversations/${conversationID}/messages`, {
    params: {
      cursor: cursor || undefined,
      limit,
    },
  })
  return unwrap(data)
}

export async function deleteMessage(conversationID: string, messageID: string): Promise<void> {
  const { data } = await http.delete<ApiResponse<Record<string, never>>>(
    `/api/conversations/${conversationID}/messages/${messageID}`,
  )
  unwrap(data)
}

export async function clearConversationMessages(conversationID: string): Promise<void> {
  const { data } = await http.delete<ApiResponse<Record<string, never>>>(
    `/api/conversations/${conversationID}/messages`,
  )
  unwrap(data)
}

export async function recallMessage(messageID: string): Promise<RecallMessageResult> {
  const { data } = await http.post<ApiResponse<RecallMessageResult>>(`/api/messages/${messageID}/recall`)
  return unwrap(data)
}

export async function getRecallEditCache(messageID: string): Promise<RecallEditCache> {
  const { data } = await http.get<ApiResponse<RecallEditCache>>(
    `/api/messages/${messageID}/recall-edit-cache`,
  )
  return unwrap(data)
}
