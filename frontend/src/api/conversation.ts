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
  last_message: {
    content: string
    message_type: string
    created_at: string
  } | null
  unread_count: number
  is_pinned: boolean
  is_muted: boolean
}

export interface Message {
  client_msg_id: string
  message_id: string
  conversation_id: string
  sender_id: string
  message_type: 'text'
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
