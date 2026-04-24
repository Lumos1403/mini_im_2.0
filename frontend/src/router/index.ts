import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'

import Chat from '../views/Chat.vue'
import Login from '../views/Login.vue'
import Register from '../views/Register.vue'

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    redirect: '/login',
  },
  {
    path: '/login',
    name: 'Login',
    component: Login,
  },
  {
    path: '/register',
    name: 'Register',
    component: Register,
  },
  {
    path: '/chat',
    name: 'Chat',
    component: Chat,
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

export default router
