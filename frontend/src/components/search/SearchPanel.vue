<script setup lang="ts">
import { computed, ref } from 'vue'

import { searchFiles, searchMessages, type SearchFileItem, type SearchMessageItem } from '../../api/search'

type SearchTab = 'messages' | 'files'

const emit = defineEmits<{
  (event: 'open-conversation', conversationID: string): void
  (event: 'close'): void
}>()

const props = defineProps<{
  open: boolean
}>()

const keyword = ref('')
const activeTab = ref<SearchTab>('messages')
const results = ref<(SearchMessageItem | SearchFileItem)[]>([])
const loading = ref(false)
const errorMsg = ref('')
const page = ref(1)
const total = ref(0)
const pageSize = 20
const hasSearched = ref(false)

const totalPages = computed(() => Math.ceil(total.value / pageSize))
const hasMore = computed(() => page.value < totalPages.value)

function conversationTypeLabel(type: string) {
  return type === 'group' ? '群聊' : '私聊'
}

function formatFileSize(size?: number) {
  if (typeof size !== 'number' || !Number.isFinite(size) || size < 0) {
    return '大小未知'
  }
  const units = ['B', 'KB', 'MB', 'GB']
  let value = size
  let unitIndex = 0
  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024
    unitIndex += 1
  }
  return `${value.toFixed(unitIndex === 0 ? 0 : 1)} ${units[unitIndex]}`
}

function switchTab(tab: SearchTab) {
  if (tab === activeTab.value) {
    return
  }
  activeTab.value = tab
  results.value = []
  total.value = 0
  page.value = 1
  errorMsg.value = ''
  hasSearched.value = false
}

async function doSearch(reset = true) {
  const kw = keyword.value.trim()
  if (!kw) {
    errorMsg.value = '请输入搜索关键词'
    return
  }

  if (reset) {
    page.value = 1
    results.value = []
    total.value = 0
  }

  loading.value = true
  errorMsg.value = ''
  hasSearched.value = true

  try {
    const params = { keyword: kw, page: page.value, page_size: pageSize }
    const fetcher = activeTab.value === 'messages' ? searchMessages : searchFiles
    const result = await fetcher(params)
    results.value = reset ? result.list : [...results.value, ...result.list]
    total.value = result.total
  } catch (error) {
    errorMsg.value = error instanceof Error ? error.message : '搜索失败'
  } finally {
    loading.value = false
  }
}

function loadMore() {
  if (!hasMore.value || loading.value) {
    return
  }
  page.value += 1
  void doSearch(false)
}

function handleKeydown(event: KeyboardEvent) {
  if (event.key === 'Enter') {
    void doSearch()
  }
}

function openConversation(conversationID: string) {
  emit('open-conversation', conversationID)
}

function close() {
  emit('close')
}
</script>

<template>
  <Teleport to="body">
    <div v-if="open" class="search-drawer-layer">
      <button class="search-drawer-mask" type="button" aria-label="关闭搜索面板" @click="close"></button>
      <aside class="search-drawer" aria-label="搜索面板">
        <header class="search-drawer-header">
          <strong>搜索消息/文件</strong>
          <button type="button" @click="close">关闭</button>
        </header>

        <div class="search-bar">
          <input
            v-model="keyword"
            maxlength="200"
            placeholder="输入关键词"
            @keydown="handleKeydown"
          />
          <button type="button" :disabled="loading" @click="doSearch()">搜索</button>
        </div>

        <nav class="search-tabs">
          <button
            :class="['tab-button', { active: activeTab === 'messages' }]"
            type="button"
            @click="switchTab('messages')"
          >
            消息
          </button>
          <button
            :class="['tab-button', { active: activeTab === 'files' }]"
            type="button"
            @click="switchTab('files')"
          >
            文件
          </button>
        </nav>

        <div class="search-drawer-body">
          <p v-if="errorMsg" class="search-status error">{{ errorMsg }}</p>

          <div v-if="loading" class="search-status">搜索中...</div>

          <div
            v-else-if="!hasSearched && results.length === 0"
            class="search-status"
          >
            输入关键词开始搜索
          </div>

          <div
            v-else-if="hasSearched && !loading && results.length === 0"
            class="search-status"
          >
            未找到相关内容
          </div>

          <article
            v-for="result in results"
            v-else
            :key="activeTab === 'messages' ? (result as SearchMessageItem).message_id : (result as SearchFileItem).file_id"
            class="search-result-card"
            @click="openConversation(result.conversation_id)"
          >
            <template v-if="activeTab === 'messages'">
              <div class="result-header">
                <span class="result-sender">{{ (result as SearchMessageItem).sender_nickname || (result as SearchMessageItem).sender_id }}</span>
                <span class="result-meta">
                  {{ conversationTypeLabel(result.conversation_type) }}
                  &middot;
                  {{ result.created_at }}
                </span>
              </div>
              <p class="result-content">{{ (result as SearchMessageItem).content }}</p>
            </template>

            <template v-else>
              <div class="result-header">
                <span class="result-file-name">{{ (result as SearchFileItem).original_name }}</span>
                <span class="result-meta">
                  {{ conversationTypeLabel(result.conversation_type) }}
                  &middot;
                  {{ result.created_at }}
                </span>
              </div>
              <div class="result-file-meta">
                <span>{{ (result as SearchFileItem).uploader_nickname || (result as SearchFileItem).uploader_id }}</span>
                <span>{{ formatFileSize((result as SearchFileItem).file_size) }} / {{ (result as SearchFileItem).mime_type || '未知' }}</span>
              </div>
            </template>
          </article>

          <button
            v-if="hasMore && !loading"
            class="load-more-button"
            type="button"
            @click="loadMore"
          >
            加载更多
          </button>
        </div>
      </aside>
    </div>
  </Teleport>
</template>

<style scoped>
.search-drawer-layer {
  position: fixed;
  inset: 0;
  z-index: 50;
  display: flex;
  justify-content: flex-end;
}

.search-drawer-mask {
  position: absolute;
  inset: 0;
  border: 0;
  background: rgba(16, 24, 40, 0.28);
  cursor: default;
}

.search-drawer {
  position: relative;
  z-index: 1;
  display: flex;
  width: min(440px, 92vw);
  height: 100%;
  min-height: 0;
  flex-direction: column;
  background: #ffffff;
  box-shadow: -12px 0 30px rgba(16, 24, 40, 0.16);
}

.search-drawer-header {
  display: flex;
  flex: 0 0 auto;
  align-items: center;
  justify-content: space-between;
  padding: 16px;
  border-bottom: 1px solid #e4e7ec;
}

.search-drawer-header strong {
  font-size: 17px;
}

.search-drawer-header button {
  height: 34px;
  border: 0;
  border-radius: 7px;
  padding: 0 10px;
  background: #eef4ff;
  color: #175cd3;
  font: inherit;
  font-size: 13px;
  font-weight: 700;
  cursor: pointer;
}

.search-bar {
  display: flex;
  flex: 0 0 auto;
  gap: 8px;
  padding: 12px 16px;
  border-bottom: 1px solid #eef2f6;
}

.search-bar input {
  min-width: 0;
  flex: 1 1 auto;
  height: 38px;
  border: 1px solid #cfd6e4;
  border-radius: 8px;
  padding: 0 10px;
  font: inherit;
}

.search-bar button {
  flex: 0 0 auto;
  min-width: 64px;
  height: 38px;
  border: 0;
  border-radius: 8px;
  background: #1570ef;
  color: #ffffff;
  font: inherit;
  font-weight: 700;
  cursor: pointer;
}

.search-bar button:disabled {
  background: #98a2b3;
  cursor: not-allowed;
}

.search-tabs {
  display: flex;
  flex: 0 0 auto;
  gap: 0;
  border-bottom: 1px solid #e4e7ec;
  padding: 0 16px;
}

.tab-button {
  padding: 10px 18px;
  border: 0;
  border-bottom: 2px solid transparent;
  background: transparent;
  color: #667085;
  font: inherit;
  font-size: 14px;
  font-weight: 600;
  cursor: pointer;
}

.tab-button.active {
  border-bottom-color: #1570ef;
  color: #1570ef;
}

.search-drawer-body {
  min-height: 0;
  flex: 1 1 auto;
  overflow-y: auto;
  padding: 14px 16px;
}

.search-status {
  padding: 32px 0;
  color: #667085;
  text-align: center;
  font-size: 14px;
}

.search-status.error {
  color: #d92d20;
}

.search-result-card {
  padding: 12px;
  margin-bottom: 8px;
  border: 1px solid #e4e7ec;
  border-radius: 8px;
  cursor: pointer;
}

.search-result-card:hover {
  border-color: #c7d7fe;
  background: #eef4ff;
}

.result-header {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 6px;
}

.result-sender,
.result-file-name {
  min-width: 0;
  overflow: hidden;
  font-weight: 700;
  font-size: 14px;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.result-meta {
  flex: 0 0 auto;
  color: #667085;
  font-size: 12px;
  white-space: nowrap;
}

.result-content {
  margin: 0;
  overflow: hidden;
  color: #475467;
  font-size: 13px;
  line-height: 1.5;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.result-file-meta {
  display: flex;
  justify-content: space-between;
  color: #667085;
  font-size: 12px;
  margin-top: 4px;
}

.load-more-button {
  display: block;
  width: 100%;
  height: 38px;
  margin-top: 12px;
  border: 0;
  border-radius: 8px;
  background: #eef4ff;
  color: #175cd3;
  font: inherit;
  font-size: 14px;
  font-weight: 700;
  cursor: pointer;
}

.load-more-button:disabled {
  background: #eaecf0;
  color: #98a2b3;
  cursor: not-allowed;
}
</style>
