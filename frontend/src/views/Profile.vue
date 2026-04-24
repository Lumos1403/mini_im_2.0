<template>
  <section class="profile-page">
    <form class="profile-panel" @submit.prevent="handleSubmit">
      <header class="profile-header">
        <div class="avatar-preview">
          <img v-if="form.avatar_url" :src="form.avatar_url" alt="" />
          <span v-else>{{ avatarInitial }}</span>
        </div>
        <div>
          <h1>个人资料</h1>
          <p>{{ profile?.username || authStore.user?.username || '' }}</p>
        </div>
      </header>

      <label>
        头像 URL
        <input v-model.trim="form.avatar_url" name="avatar_url" autocomplete="url" />
      </label>

      <label>
        昵称
        <input v-model.trim="form.nickname" name="nickname" autocomplete="nickname" />
      </label>

      <label>
        性别
        <select v-model="form.gender" name="gender">
          <option value="">未设置</option>
          <option value="male">男</option>
          <option value="female">女</option>
          <option value="other">其他</option>
        </select>
      </label>

      <label>
        个性签名
        <textarea v-model.trim="form.bio" name="bio" rows="4"></textarea>
      </label>

      <dl v-if="profile" class="review-state">
        <div>
          <dt>资料状态</dt>
          <dd>{{ profile.profile_status }}</dd>
        </div>
        <div>
          <dt>审核说明</dt>
          <dd>{{ profile.profile_review_reason || '无' }}</dd>
        </div>
      </dl>

      <p v-if="error" class="error">{{ error }}</p>
      <p v-if="success" class="success">{{ success }}</p>

      <button type="submit" :disabled="loading || submitting">
        {{ submitting ? '保存中...' : '保存资料' }}
      </button>
    </form>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'

import * as userApi from '../api/user'
import type { UserProfile } from '../api/user'
import { useAuthStore } from '../stores/auth'

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
    error.value = err instanceof Error ? err.message : '加载资料失败'
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
    success.value = '资料已保存'
  } catch (err) {
    error.value = err instanceof Error ? err.message : '保存资料失败'
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

<style scoped>
.profile-page {
  display: grid;
  min-height: calc(100vh - 56px);
  place-items: start center;
  padding: 32px 24px;
}

.profile-panel {
  display: grid;
  gap: 16px;
  width: min(560px, 100%);
  padding: 24px;
  border: 1px solid #dde3ee;
  border-radius: 8px;
  background: #ffffff;
}

.profile-header {
  display: flex;
  gap: 16px;
  align-items: center;
  margin-bottom: 8px;
}

.avatar-preview {
  display: grid;
  flex: 0 0 72px;
  width: 72px;
  height: 72px;
  overflow: hidden;
  place-items: center;
  border: 1px solid #cbd5e1;
  border-radius: 50%;
  color: #2563eb;
  background: #eef4ff;
  font-size: 28px;
  font-weight: 700;
}

.avatar-preview img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

h1 {
  margin: 0;
  font-size: 24px;
}

.profile-header p {
  margin: 6px 0 0;
  color: #667085;
  font-size: 14px;
}

label {
  display: grid;
  gap: 8px;
  color: #344054;
  font-size: 14px;
  font-weight: 600;
}

input,
select,
textarea {
  width: 100%;
  min-width: 0;
  border: 1px solid #cbd5e1;
  border-radius: 6px;
  color: #172033;
  background: #ffffff;
  font: inherit;
}

input,
select {
  height: 40px;
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
  padding: 12px;
  border-radius: 8px;
  background: #f8fafc;
}

.review-state div {
  min-width: 0;
}

dt {
  color: #667085;
  font-size: 12px;
}

dd {
  margin: 4px 0 0;
  color: #172033;
  overflow-wrap: anywhere;
  font-size: 14px;
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

.error,
.success {
  margin: 0;
  font-size: 14px;
}

.error {
  color: #b42318;
}

.success {
  color: #027a48;
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
