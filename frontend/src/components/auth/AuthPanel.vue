<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { useAuthStore } from '../../stores/auth'

type AuthMode = 'login' | 'register'

const props = defineProps<{
  initialMode: AuthMode
}>()

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()

const mode = ref<AuthMode>(props.initialMode)
const username = ref('')
const nickname = ref('')
const password = ref('')
const submitting = ref(false)
const error = ref('')
const success = ref('')

const title = computed(() => (mode.value === 'login' ? 'Welcome back' : 'Create account'))
const subtitle = computed(() =>
  mode.value === 'login'
    ? 'Sign in to open your private messaging workspace.'
    : 'Register with username, password, and nickname.',
)
const submitLabel = computed(() => {
  if (submitting.value) {
    return mode.value === 'login' ? 'Signing in...' : 'Creating...'
  }
  return mode.value === 'login' ? 'Sign in' : 'Create account'
})

watch(
  () => props.initialMode,
  (next) => {
    mode.value = next
    clearFeedback()
  },
)

async function handleSubmit() {
  clearFeedback()
  submitting.value = true
  try {
    if (mode.value === 'login') {
      await auth.login(username.value, password.value)
      await router.push((route.query.redirect as string) || '/chat')
      return
    }

    await auth.register(username.value, password.value, nickname.value)
    success.value = 'Account created. Sign in with the new credentials.'
    password.value = ''
    switchMode('login')
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Authentication failed'
  } finally {
    submitting.value = false
  }
}

function switchMode(nextMode: AuthMode) {
  mode.value = nextMode
  clearFeedback()
  void router.replace(nextMode === 'login' ? '/login' : '/register')
}

function clearFeedback() {
  error.value = ''
  success.value = ''
}

function handleButtonMove(event: PointerEvent) {
  const target = event.currentTarget as HTMLElement
  const rect = target.getBoundingClientRect()
  window.dispatchEvent(
    new CustomEvent('auth-magnet-focus', {
      detail: {
        x: rect.left + rect.width / 2,
        y: rect.top + rect.height / 2,
        radius: Math.max(rect.width, rect.height) * 3.1,
        strength: 1.55,
      },
    }),
  )
}

function clearButtonMagnet() {
  window.dispatchEvent(new CustomEvent('auth-magnet-clear'))
}
</script>

<template>
  <section class="auth-panel" aria-label="Authentication panel">
    <div class="mode-switch" role="tablist" aria-label="Authentication mode">
      <button
        type="button"
        :class="{ active: mode === 'login' }"
        role="tab"
        :aria-selected="mode === 'login'"
        @click="switchMode('login')"
      >
        Login
      </button>
      <button
        type="button"
        :class="{ active: mode === 'register' }"
        role="tab"
        :aria-selected="mode === 'register'"
        @click="switchMode('register')"
      >
        Register
      </button>
    </div>

    <header>
      <h2>{{ title }}</h2>
      <p>{{ subtitle }}</p>
    </header>

    <form @submit.prevent="handleSubmit">
      <label>
        Username
        <input v-model.trim="username" name="username" autocomplete="username" required />
      </label>

      <label v-if="mode === 'register'">
        Nickname
        <input v-model.trim="nickname" name="nickname" autocomplete="nickname" required />
      </label>

      <label>
        Password
        <input
          v-model="password"
          name="password"
          type="password"
          :autocomplete="mode === 'login' ? 'current-password' : 'new-password'"
          required
        />
      </label>

      <p v-if="error" class="status error">{{ error }}</p>
      <p v-if="success" class="status success">{{ success }}</p>

      <button
        class="auth-submit"
        type="submit"
        :disabled="submitting"
        @pointermove="handleButtonMove"
        @pointerenter="handleButtonMove"
        @pointerleave="clearButtonMagnet"
        @blur="clearButtonMagnet"
      >
        <span>{{ submitLabel }}</span>
      </button>
    </form>
  </section>
</template>

<style scoped>
.auth-panel {
  position: relative;
  z-index: 2;
  width: min(460px, 100%);
  justify-self: end;
  padding: 28px;
  border: 1px solid rgba(240, 207, 132, 0.18);
  border-radius: 18px;
  background:
    linear-gradient(145deg, rgba(255, 255, 255, 0.12), rgba(255, 255, 255, 0.035)),
    rgba(12, 14, 15, 0.62);
  box-shadow:
    0 0 0 1px rgba(255, 255, 255, 0.04) inset,
    0 28px 80px rgba(0, 0, 0, 0.44);
  backdrop-filter: blur(22px);
}

.auth-panel::before {
  position: absolute;
  inset: 1px;
  border-radius: 17px;
  content: "";
  pointer-events: none;
  background: linear-gradient(120deg, rgba(240, 207, 132, 0.16), transparent 38%, rgba(255, 255, 255, 0.07));
  mask: linear-gradient(#000, transparent 42%);
}

.mode-switch {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 6px;
  padding: 5px;
  border: 1px solid rgba(240, 207, 132, 0.12);
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.055);
}

.mode-switch button {
  height: 38px;
  border: 0;
  border-radius: 8px;
  color: rgba(243, 239, 229, 0.72);
  background: transparent;
  cursor: pointer;
  font-weight: 720;
}

.mode-switch button.active {
  color: #16120a;
  background: linear-gradient(135deg, #f0cf84, #c99f50);
  box-shadow: 0 10px 24px rgba(215, 180, 106, 0.2);
}

header {
  margin: 28px 0 22px;
}

h2 {
  margin: 0;
  color: #fff8eb;
  font-size: 36px;
  line-height: 1;
}

p {
  margin: 10px 0 0;
  color: rgba(243, 239, 229, 0.68);
  line-height: 1.55;
}

form {
  display: grid;
  gap: 15px;
}

label {
  display: grid;
  gap: 8px;
  color: rgba(243, 239, 229, 0.84);
  font-size: 13px;
  font-weight: 760;
}

input {
  height: 46px;
  border: 1px solid rgba(240, 207, 132, 0.16);
  border-radius: 10px;
  padding: 0 14px;
  color: var(--text);
  background: rgba(0, 0, 0, 0.24);
  box-shadow: 0 1px 0 rgba(255, 255, 255, 0.05) inset;
}

input:hover {
  border-color: rgba(240, 207, 132, 0.32);
}

input:focus {
  border-color: rgba(240, 207, 132, 0.62);
}

.status {
  margin: 0;
  border-radius: 9px;
  padding: 10px 12px;
  font-size: 13px;
}

.status.error {
  color: #ffd4d4;
  background: rgba(239, 68, 68, 0.14);
}

.status.success {
  color: #cbffe3;
  background: rgba(94, 226, 160, 0.13);
}

.auth-submit {
  position: relative;
  height: 48px;
  overflow: hidden;
  border: 1px solid rgba(240, 207, 132, 0.34);
  border-radius: 10px;
  color: #14100a;
  background:
    radial-gradient(circle at 50% 0, rgba(255, 255, 255, 0.36), transparent 45%),
    linear-gradient(135deg, #f3d48d, #c6953c);
  box-shadow:
    0 14px 34px rgba(215, 180, 106, 0.24),
    0 0 0 rgba(240, 207, 132, 0);
  cursor: pointer;
  font-weight: 820;
}

.auth-submit::before {
  position: absolute;
  inset: -80% -20%;
  content: "";
  opacity: 0;
  background: radial-gradient(circle, rgba(255, 255, 255, 0.75), transparent 42%);
  transform: translateX(-30%);
  transition:
    opacity 200ms ease,
    transform 260ms ease;
}

.auth-submit:hover:not(:disabled) {
  border-color: rgba(255, 234, 174, 0.72);
  box-shadow:
    0 18px 38px rgba(215, 180, 106, 0.3),
    0 0 36px rgba(240, 207, 132, 0.28);
}

.auth-submit:hover:not(:disabled)::before {
  opacity: 0.72;
  transform: translateX(22%);
}

.auth-submit span {
  position: relative;
  z-index: 1;
}

@media (max-width: 860px) {
  .auth-panel {
    justify-self: stretch;
  }
}
</style>
