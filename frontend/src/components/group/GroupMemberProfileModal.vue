<script setup lang="ts">
import { computed, ref, watch } from 'vue'

import type { GroupFriendshipStatus, GroupMember, GroupRole } from '../../api/group'
import GroupRoleBadge from './GroupRoleBadge.vue'

const props = defineProps<{
  open: boolean
  member: GroupMember | null
  requesting?: boolean
}>()

const emit = defineEmits<{
  (event: 'close'): void
  (event: 'add-friend', member: GroupMember, message: string): void
}>()

const requestMessage = ref('我是群里的成员')

const displayName = computed(() => props.member?.nickname || props.member?.user_id || '')
const canAddFriend = computed(() => props.member?.friendship_status === 'not_friend')

watch(
  () => props.member?.user_id,
  () => {
    requestMessage.value = '我是群里的成员'
  },
)

function avatarText(member: GroupMember) {
  return (member.nickname || member.user_id).slice(0, 1).toUpperCase() || '#'
}

function roleText(role: GroupRole) {
  const labels: Record<GroupRole, string> = {
    owner: '群主',
    admin: '管理员',
    member: '成员',
  }
  return labels[role] || '成员'
}

function friendshipText(status: GroupFriendshipStatus) {
  const labels: Record<GroupFriendshipStatus, string> = {
    self: '这是你自己',
    friend: '已是好友',
    not_friend: '可以添加好友',
    pending_sent: '申请中',
    pending_received: '对方已申请添加你',
  }
  return labels[status] || status
}

function addFriend() {
  if (!props.member || !canAddFriend.value || props.requesting) {
    return
  }
  emit('add-friend', props.member, requestMessage.value)
}
</script>

<template>
  <Teleport to="body">
    <div v-if="open && member" class="modal-layer">
      <button class="modal-mask" type="button" aria-label="关闭成员资料" @click="emit('close')"></button>
      <section class="profile-modal" aria-label="群成员资料">
        <header class="profile-header">
          <strong>成员资料</strong>
          <button type="button" @click="emit('close')">关闭</button>
        </header>

        <div class="profile-body">
          <img v-if="member.avatar_url" class="profile-avatar image" :src="member.avatar_url" alt="" />
          <span v-else class="profile-avatar">{{ avatarText(member) }}</span>

          <div class="profile-name-row">
            <h3>{{ displayName }}</h3>
            <GroupRoleBadge :role="member.role" />
          </div>
          <small class="profile-id">{{ member.user_id }}</small>
          <p class="profile-bio">{{ member.bio?.trim() || '暂无个性签名' }}</p>

          <dl class="profile-fields">
            <div>
              <dt>群身份</dt>
              <dd>{{ roleText(member.role) }}</dd>
            </div>
            <div>
              <dt>好友状态</dt>
              <dd>{{ friendshipText(member.friendship_status) }}</dd>
            </div>
          </dl>

          <div class="profile-actions">
            <span v-if="member.friendship_status === 'self'" class="state-text">这是你自己</span>
            <button v-else-if="member.friendship_status === 'friend'" type="button" disabled>已是好友</button>
            <template v-else-if="member.friendship_status === 'not_friend'">
              <input v-model="requestMessage" maxlength="100" type="text" />
              <button type="button" :disabled="requesting" @click="addFriend">
                {{ requesting ? '发送中' : '添加好友' }}
              </button>
            </template>
            <button v-else-if="member.friendship_status === 'pending_sent'" type="button" disabled>
              申请中
            </button>
            <span v-else-if="member.friendship_status === 'pending_received'" class="state-text">
              对方已申请添加你
            </span>
          </div>
        </div>
      </section>
    </div>
  </Teleport>
</template>

<style scoped>
.modal-layer {
  position: fixed;
  inset: 0;
  z-index: 50;
  display: grid;
  place-items: center;
  padding: 20px;
}

.modal-mask {
  position: absolute;
  inset: 0;
  border: 0;
  background: rgba(16, 24, 40, 0.36);
  cursor: default;
}

.profile-modal {
  position: relative;
  z-index: 1;
  width: min(380px, 92vw);
  overflow: hidden;
  border-radius: 8px;
  background: #ffffff;
  box-shadow: 0 18px 44px rgba(16, 24, 40, 0.2);
}

.profile-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 14px 16px;
  border-bottom: 1px solid #e4e7ec;
}

.profile-header button,
.profile-actions button {
  height: 34px;
  border: 0;
  border-radius: 7px;
  padding: 0 12px;
  background: #eef4ff;
  color: #175cd3;
  font: inherit;
  font-size: 13px;
  font-weight: 700;
  cursor: pointer;
}

.profile-header button {
  min-width: 56px;
}

.profile-actions button:disabled {
  background: #eaecf0;
  color: #98a2b3;
  cursor: not-allowed;
}

.profile-body {
  padding: 18px;
  text-align: center;
}

.profile-avatar {
  display: grid;
  width: 72px;
  height: 72px;
  margin: 0 auto 12px;
  place-items: center;
  border-radius: 50%;
  background: #344054;
  color: #ffffff;
  font-size: 24px;
  font-weight: 700;
  object-fit: cover;
}

.profile-avatar.image {
  background: #f2f4f7;
}

.profile-name-row {
  display: flex;
  min-width: 0;
  align-items: center;
  justify-content: center;
  gap: 7px;
}

.profile-name-row h3 {
  overflow: hidden;
  margin: 0;
  font-size: 18px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.profile-id {
  display: block;
  margin-top: 4px;
  color: #667085;
  font-size: 12px;
}

.profile-bio {
  margin: 12px 0 0;
  color: #475467;
  font-size: 13px;
  line-height: 1.5;
  word-break: break-word;
}

.profile-fields {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px;
  margin: 16px 0 0;
  text-align: left;
}

.profile-fields div {
  border-radius: 8px;
  padding: 9px 10px;
  background: #f8fafc;
}

.profile-fields dt,
.profile-fields dd {
  margin: 0;
}

.profile-fields dt {
  color: #667085;
  font-size: 12px;
}

.profile-fields dd {
  margin-top: 4px;
  color: #101828;
  font-size: 13px;
  font-weight: 700;
}

.profile-actions {
  display: grid;
  gap: 8px;
  margin-top: 16px;
}

.profile-actions input {
  width: 100%;
  height: 36px;
  border: 1px solid #cfd6e4;
  border-radius: 7px;
  padding: 0 10px;
  font: inherit;
}

.state-text {
  display: block;
  border-radius: 8px;
  padding: 10px;
  background: #f2f4f7;
  color: #475467;
  font-size: 13px;
}
</style>
