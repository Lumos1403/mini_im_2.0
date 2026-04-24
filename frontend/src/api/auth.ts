import { http, unwrap, type ApiResponse } from './http'

export interface UserInfo {
  user_id: string
  username: string
  nickname: string
  avatar_url: string
}

export interface RegisterPayload {
  username: string
  password: string
  nickname: string
}

export interface RegisterResult {
  user_id: string
  username: string
  nickname: string
}

export interface LoginPayload {
  username: string
  password: string
}

export interface AuthResult {
  access_token: string
  refresh_token: string
  expires_in: number
  user: UserInfo
}

export interface RefreshResult {
  access_token: string
  refresh_token: string
  expires_in: number
}

export async function register(payload: RegisterPayload): Promise<RegisterResult> {
  const { data } = await http.post<ApiResponse<RegisterResult>>('/api/auth/register', payload)
  return unwrap(data)
}

export async function login(payload: LoginPayload): Promise<AuthResult> {
  const { data } = await http.post<ApiResponse<AuthResult>>('/api/auth/login', payload)
  return unwrap(data)
}

export async function refresh(refreshToken: string): Promise<RefreshResult> {
  const { data } = await http.post<ApiResponse<RefreshResult>>('/api/auth/refresh', {
    refresh_token: refreshToken,
  })
  return unwrap(data)
}

export async function logout(): Promise<void> {
  const { data } = await http.post<ApiResponse<Record<string, never>>>('/api/auth/logout')
  unwrap(data)
}

export async function getMe(): Promise<UserInfo> {
  const { data } = await http.get<ApiResponse<UserInfo>>('/api/users/me')
  return unwrap(data)
}
