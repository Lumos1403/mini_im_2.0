import { http, unwrap, type ApiResponse } from './http'

export interface SearchMessageItem {
  message_id: string
  conversation_id: string
  conversation_type: string
  sender_id: string
  sender_nickname: string
  sender_avatar_url: string
  message_type: string
  content: string
  created_at: string
}

export interface SearchFileItem {
  file_id: string
  original_name: string
  file_size: number
  mime_type: string
  uploader_id: string
  uploader_nickname: string
  message_id: string
  conversation_id: string
  conversation_type: string
  created_at: string
}

export interface SearchPageResult<T> {
  list: T[]
  total: number
  page: number
  page_size: number
}

export interface SearchParams {
  keyword: string
  page?: number
  page_size?: number
}

export async function searchMessages(
  params: SearchParams,
): Promise<SearchPageResult<SearchMessageItem>> {
  const { data } = await http.get<ApiResponse<SearchPageResult<SearchMessageItem>>>(
    '/api/search/messages',
    {
      params: {
        keyword: params.keyword,
        page: params.page ?? 1,
        page_size: params.page_size ?? 20,
      },
    },
  )
  return unwrap(data)
}

export async function searchFiles(
  params: SearchParams,
): Promise<SearchPageResult<SearchFileItem>> {
  const { data } = await http.get<ApiResponse<SearchPageResult<SearchFileItem>>>(
    '/api/search/files',
    {
      params: {
        keyword: params.keyword,
        page: params.page ?? 1,
        page_size: params.page_size ?? 20,
      },
    },
  )
  return unwrap(data)
}
