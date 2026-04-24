import { defineStore } from 'pinia'

import * as authApi from '../api/auth'
import type { UserInfo } from '../api/auth'
import type { UserProfile } from '../api/user'

interface AuthState {
  accessToken: string
  refreshToken: string
  user: UserInfo | null
}

export const useAuthStore = defineStore('auth', {
  state: (): AuthState => ({
    accessToken: localStorage.getItem('access_token') || '',
    refreshToken: localStorage.getItem('refresh_token') || '',
    user: loadStoredUser(),
  }),
  actions: {
    async login(username: string, password: string) {
      const result = await authApi.login({ username, password })
      this.setAuth(result.access_token, result.refresh_token, result.user)
    },
    async register(username: string, password: string, nickname: string) {
      return authApi.register({ username, password, nickname })
    },
    async refresh() {
      if (!this.refreshToken) {
        throw new Error('missing refresh token')
      }
      const result = await authApi.refresh(this.refreshToken)
      this.accessToken = result.access_token
      this.refreshToken = result.refresh_token
      localStorage.setItem('access_token', result.access_token)
      localStorage.setItem('refresh_token', result.refresh_token)
    },
    async loadMe() {
      this.user = await authApi.getMe()
      localStorage.setItem('current_user', JSON.stringify(this.user))
    },
    async logout() {
      try {
        if (this.accessToken) {
          await authApi.logout()
        }
      } finally {
        this.clearAuth()
      }
    },
    setAuth(accessToken: string, refreshToken: string, user: UserInfo) {
      this.accessToken = accessToken
      this.refreshToken = refreshToken
      this.user = user
      localStorage.setItem('access_token', accessToken)
      localStorage.setItem('refresh_token', refreshToken)
      localStorage.setItem('current_user', JSON.stringify(user))
    },
    syncProfile(profile: UserProfile) {
      const user = {
        user_id: profile.user_id,
        username: profile.username,
        nickname: profile.nickname,
        avatar_url: profile.avatar_url,
      }
      this.user = user
      localStorage.setItem('current_user', JSON.stringify(user))
    },
    clearAuth() {
      this.accessToken = ''
      this.refreshToken = ''
      this.user = null
      localStorage.removeItem('access_token')
      localStorage.removeItem('refresh_token')
      localStorage.removeItem('current_user')
    },
  },
})

function loadStoredUser(): UserInfo | null {
  const raw = localStorage.getItem('current_user')
  if (!raw) {
    return null
  }

  try {
    return JSON.parse(raw) as UserInfo
  } catch {
    localStorage.removeItem('current_user')
    return null
  }
}
