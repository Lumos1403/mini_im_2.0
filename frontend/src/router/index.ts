import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'

import Chat from '../views/Chat.vue'
import Login from '../views/Login.vue'
import Profile from '../views/Profile.vue'
import Register from '../views/Register.vue'
import { useAuthStore } from '../stores/auth'
import { pinia } from '../stores'
import { clearClientSession, resetPostAuthState } from '../stores/session'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: '/login',
  },
  {
    path: '/login',
    name: 'Login',
    component: Login,
    meta: { guestOnly: true },
  },
  {
    path: '/register',
    name: 'Register',
    component: Register,
    meta: { guestOnly: true },
  },
  {
    path: '/chat',
    name: 'Chat',
    component: Chat,
    meta: { requiresAuth: true },
  },
  {
    path: '/profile',
    name: 'Profile',
    component: Profile,
    meta: { requiresAuth: true },
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach(async (to) => {
  const auth = useAuthStore(pinia)

  if (to.meta.requiresAuth) {
    const ok = await auth.ensureSession()
    if (!ok) {
      clearClientSession()
      return {
        name: 'Login',
        query: to.fullPath === '/chat' ? undefined : { redirect: to.fullPath },
      }
    }
    return true
  }

  if (to.meta.guestOnly && auth.hasToken) {
    const ok = auth.bootstrapped && auth.user ? true : await auth.ensureSession()
    if (ok) {
      resetPostAuthState()
      return { name: 'Chat' }
    }
    clearClientSession()
  }

  return true
})

export default router
