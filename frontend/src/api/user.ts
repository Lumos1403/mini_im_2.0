import { http, unwrap, type ApiResponse } from './http'

export interface UserProfile {
  user_id: string
  username: string
  nickname: string
  avatar_url: string
  gender: string
  bio: string
  profile_status: string
  profile_review_reason: string
}

export interface PageResult<T> {
  list: T[]
  total: number
  page: number
  page_size: number
}

export interface UserSearchResult {
  user_id: string
  username?: string
  nickname: string
  avatar_url: string
  bio?: string
}

export interface UpdateProfilePayload {
  nickname: string
  avatar_url: string
  gender: string
  bio: string
}

export async function getMyProfile(): Promise<UserProfile> {
  const { data } = await http.get<ApiResponse<UserProfile>>('/api/users/me/profile')
  return unwrap(data)
}

export async function updateMyProfile(payload: UpdateProfilePayload): Promise<UserProfile> {
  const { data } = await http.put<ApiResponse<UserProfile>>('/api/users/me/profile', payload)
  return unwrap(data)
}

export async function searchUsers(keyword: string, page = 1, pageSize = 20): Promise<PageResult<UserSearchResult>> {
  const { data } = await http.get<ApiResponse<PageResult<UserSearchResult>>>('/api/users/search', {
    params: {
      keyword,
      page,
      page_size: pageSize,
    },
  })
  return unwrap(data)
}
