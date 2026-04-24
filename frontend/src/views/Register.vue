<template>
  <section class="page">
    <form class="panel" @submit.prevent="handleSubmit">
      <h1>Register</h1>

      <label>
        Username
        <input v-model.trim="username" name="username" autocomplete="username" />
      </label>

      <label>
        Nickname
        <input v-model.trim="nickname" name="nickname" autocomplete="nickname" />
      </label>

      <label>
        Password
        <input v-model="password" name="password" type="password" autocomplete="new-password" />
      </label>

      <p v-if="error" class="error">{{ error }}</p>
      <p v-if="success" class="success">{{ success }}</p>

      <button type="submit" :disabled="submitting">
        {{ submitting ? 'Creating...' : 'Create account' }}
      </button>

      <RouterLink to="/login">Back to login</RouterLink>
    </form>
  </section>
</template>

<script setup lang="ts">
import { ref } from 'vue'

import { useAuthStore } from '../stores/auth'

const authStore = useAuthStore()

const username = ref('')
const nickname = ref('')
const password = ref('')
const submitting = ref(false)
const error = ref('')
const success = ref('')

async function handleSubmit() {
  error.value = ''
  success.value = ''
  submitting.value = true
  try {
    await authStore.register(username.value, password.value, nickname.value)
    success.value = 'Account created. You can login now.'
    password.value = ''
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Registration failed'
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.page {
  display: grid;
  min-height: calc(100vh - 56px);
  place-items: center;
  padding: 24px;
}

.panel {
  display: grid;
  gap: 12px;
  width: min(420px, 100%);
  padding: 24px;
  border: 1px solid #dde3ee;
  border-radius: 8px;
  background: #ffffff;
}

h1 {
  margin: 0 0 12px;
  font-size: 24px;
}

label {
  display: grid;
  gap: 12px;
  color: #344054;
  font-size: 14px;
  font-weight: 600;
}

input {
  height: 40px;
  padding: 0 12px;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  color: #172033;
  font: inherit;
}

button {
  height: 40px;
  border: 0;
  border-radius: 6px;
  color: #ffffff;
  background: #2563eb;
  font: inherit;
  font-weight: 700;
  cursor: pointer;
}

button:disabled {
  cursor: not-allowed;
  opacity: 0.7;
}

a,
.error,
.success {
  margin: 0;
  font-size: 14px;
}

a {
  color: #2563eb;
}

.error {
  color: #b42318;
}

.success {
  color: #027a48;
}
</style>
