import type { Router } from 'vue-router'

import { setAuthFailureHandler, setTokenUpdateHandler } from '../api/http'
import { useAuthStore } from './auth'
import { useChatStore } from './chat'
import { useFriendStore } from './friend'
import { useGroupStore } from './group'
import { pinia } from './index'
import { useWsStore } from './ws'

let handlingAuthFailure = false

export function resetPostAuthState() {
  useWsStore(pinia).resetAll()
  useChatStore(pinia).resetAll()
  useFriendStore(pinia).resetAll()
  useGroupStore(pinia).resetAll()
}

export function clearClientSession() {
  resetPostAuthState()
  useAuthStore(pinia).clearAuth()
}

export async function logoutAndReset(router: Router) {
  const auth = useAuthStore(pinia)
  try {
    if (auth.accessToken) {
      await auth.logout()
    }
  } finally {
    clearClientSession()
    if (router.currentRoute.value.name !== 'Login') {
      await router.replace({ name: 'Login' })
    }
  }
}

export function installSessionHandlers(router: Router) {
  setTokenUpdateHandler((accessToken, refreshToken) => {
    useAuthStore(pinia).setTokens(accessToken, refreshToken)
  })

  setAuthFailureHandler(async () => {
    if (handlingAuthFailure) {
      return
    }
    handlingAuthFailure = true
    try {
      clearClientSession()
      if (router.currentRoute.value.meta.requiresAuth) {
        await router.replace({ name: 'Login' })
      }
    } finally {
      handlingAuthFailure = false
    }
  })
}
