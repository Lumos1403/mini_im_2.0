<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'

import * as userApi from '../api/user'
import type { UserProfile } from '../api/user'
import AppShell from '../layouts/AppShell.vue'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const profile = ref<UserProfile | null>(null)
const loading = ref(false)
const submitting = ref(false)
const error = ref('')
const success = ref('')

const form = reactive({
  nickname: '',
  avatar_url: '',
  gender: '',
  bio: '',
})

const avatarInitial = computed(() => {
  const name = form.nickname || authStore.user?.nickname || authStore.user?.username || ''
  return name.slice(0, 1).toUpperCase()
})

onMounted(() => {
  void loadProfile()
})

async function loadProfile() {
  error.value = ''
  loading.value = true
  try {
    const result = await userApi.getMyProfile()
    applyProfile(result)
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to load profile'
  } finally {
    loading.value = false
  }
}

async function handleSubmit() {
  error.value = ''
  success.value = ''
  submitting.value = true
  try {
    await userApi.updateMyProfile({
      nickname: form.nickname,
      avatar_url: form.avatar_url,
      gender: form.gender,
      bio: form.bio,
    })
    const refreshed = await userApi.getMyProfile()
    applyProfile(refreshed)
    success.value = 'Profile saved.'
  } catch (err) {
    error.value = err instanceof Error ? err.message : 'Failed to save profile'
  } finally {
    submitting.value = false
  }
}

function applyProfile(nextProfile: UserProfile) {
  profile.value = nextProfile
  form.nickname = nextProfile.nickname
  form.avatar_url = nextProfile.avatar_url
  form.gender = nextProfile.gender
  form.bio = nextProfile.bio
  authStore.syncProfile(nextProfile)
}
</script>

<template>
  <AppShell>
    <main class="profile-page">
      <form class="profile-panel" @submit.prevent="handleSubmit">
        <header class="profile-header">
          <div class="avatar-preview">
            <img v-if="form.avatar_url" :src="form.avatar_url" alt="" />
            <span v-else>{{ avatarInitial }}</span>
          </div>
          <div>
            <h1>Profile</h1>
            <p>{{ profile?.username || authStore.user?.username || '' }}</p>
          </div>
        </header>

        <label>
          Avatar URL
          <input v-model.trim="form.avatar_url" name="avatar_url" autocomplete="url" />
        </label>

        <label>
          Nickname
          <input v-model.trim="form.nickname" name="nickname" autocomplete="nickname" />
        </label>

        <label>
          Gender
          <select v-model="form.gender" name="gender">
            <option value="">Unset</option>
            <option value="male">Male</option>
            <option value="female">Female</option>
            <option value="other">Other</option>
          </select>
        </label>

        <label>
          Bio
          <textarea v-model.trim="form.bio" name="bio" rows="4"></textarea>
        </label>

        <dl v-if="profile" class="review-state">
          <div>
            <dt>Profile status</dt>
            <dd>{{ profile.profile_status }}</dd>
          </div>
          <div>
            <dt>Review reason</dt>
            <dd>{{ profile.profile_review_reason || 'None' }}</dd>
          </div>
        </dl>

        <p v-if="error" class="status error">{{ error }}</p>
        <p v-if="success" class="status success">{{ success }}</p>

        <div class="profile-actions">
          <button type="button" @click="router.push('/chat')">Back to chat</button>
          <button class="primary" type="submit" :disabled="loading || submitting">
            {{ submitting ? 'Saving...' : 'Save profile' }}
          </button>
        </div>
      </form>
    </main>
  </AppShell>
</template>

<style scoped>
.profile-page {
  display: grid;
  height: 100%;
  min-height: 0;
  place-items: start center;
  overflow-y: auto;
  padding: 40px 24px;
  color: var(--text);
  background: transparent;
}

.profile-panel {
  display: grid;
  gap: 16px;
  width: min(560px, 100%);
  border: 1px solid rgba(240, 207, 132, 0.18);
  border-radius: 16px;
  padding: 24px;
  background: rgba(12, 14, 15, 0.82);
  box-shadow: var(--shadow-panel);
}

.profile-header {
  display: flex;
  gap: 16px;
  align-items: center;
  margin-bottom: 8px;
}

.avatar-preview {
  display: grid;
  width: 72px;
  height: 72px;
  flex: 0 0 72px;
  overflow: hidden;
  place-items: center;
  border-radius: 18px;
  color: #171009;
  background: linear-gradient(135deg, #f0cf84, #9a7535);
  font-size: 28px;
  font-weight: 900;
}

.avatar-preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

h1 {
  margin: 0;
  color: #fff7e8;
  font-size: 28px;
}

.profile-header p {
  margin: 6px 0 0;
  color: var(--text-muted);
  font-size: 14px;
}

label {
  display: grid;
  gap: 8px;
  color: var(--text-soft);
  font-size: 14px;
  font-weight: 760;
}

input,
select,
textarea {
  width: 100%;
  min-width: 0;
  border: 1px solid rgba(240, 207, 132, 0.16);
  border-radius: 10px;
  color: var(--text);
  background: rgba(0, 0, 0, 0.22);
}

input,
select {
  height: 42px;
  padding: 0 12px;
}

textarea {
  resize: vertical;
  padding: 10px 12px;
}

.review-state {
  display: grid;
  gap: 12px;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  margin: 0;
  border: 1px solid rgba(240, 207, 132, 0.12);
  border-radius: 12px;
  padding: 12px;
  background: rgba(255, 255, 255, 0.055);
}

dt {
  color: var(--text-muted);
  font-size: 12px;
}

dd {
  margin: 4px 0 0;
  color: var(--text-soft);
  overflow-wrap: anywhere;
  font-size: 14px;
}

.profile-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

button {
  height: 40px;
  border: 1px solid rgba(240, 207, 132, 0.16);
  border-radius: 10px;
  padding: 0 14px;
  color: var(--text-soft);
  background: rgba(255, 255, 255, 0.07);
  cursor: pointer;
  font-weight: 780;
}

button.primary {
  color: #171009;
  background: linear-gradient(135deg, #f0cf84, #b68a3e);
}

.status {
  margin: 0;
  border-radius: 10px;
  padding: 10px 12px;
  font-size: 14px;
}

.status.error {
  color: #ffd4d4;
  background: rgba(239, 68, 68, 0.14);
}

.status.success {
  color: #cbffe3;
  background: rgba(94, 226, 160, 0.13);
}

@media (max-width: 560px) {
  .profile-page {
    padding: 16px;
  }

  .profile-panel {
    padding: 18px;
  }

  .review-state {
    grid-template-columns: 1fr;
  }
}
</style>
