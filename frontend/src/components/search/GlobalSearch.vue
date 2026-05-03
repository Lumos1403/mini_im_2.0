<script setup lang="ts">
import { computed, nextTick, ref, watch } from 'vue'

import { searchFiles, searchMessages, type SearchFileItem, type SearchMessageItem } from '../../api/search'

type SearchTab = 'messages' | 'files'

const props = defineProps<{
  open: boolean
}>()

const emit = defineEmits<{
  (event: 'open-conversation', conversationID: string): void
  (event: 'close'): void
}>()

const inputRef = ref<HTMLInputElement | null>(null)
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

watch(
  () => props.open,
  async (open) => {
    if (open) {
      await nextTick()
      inputRef.value?.focus()
      return
    }
    resetSearch()
  },
)

function conversationTypeLabel(type: string) {
  return type === 'group' ? 'Group' : 'Private'
}

function formatFileSize(size?: number) {
  if (typeof size !== 'number' || !Number.isFinite(size) || size < 0) {
    return 'Unknown size'
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
    errorMsg.value = 'Enter a keyword to search.'
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
    errorMsg.value = error instanceof Error ? error.message : 'Search failed'
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

function openConversation(conversationID: string) {
  emit('open-conversation', conversationID)
}

function resetSearch() {
  keyword.value = ''
  activeTab.value = 'messages'
  results.value = []
  loading.value = false
  errorMsg.value = ''
  page.value = 1
  total.value = 0
  hasSearched.value = false
}

function close() {
  emit('close')
}
</script>

<template>
  <Teleport to="body">
    <div v-if="open" class="search-layer">
      <button class="search-mask" type="button" aria-label="Close search" @click="close"></button>
      <section class="search-panel" aria-label="Global search">
        <header>
          <div>
            <strong>Global search</strong>
            <small>Results come from the backend search APIs only.</small>
          </div>
          <button type="button" @click="close">Close</button>
        </header>

        <form class="search-form" @submit.prevent="doSearch()">
          <input
            ref="inputRef"
            v-model="keyword"
            maxlength="200"
            placeholder="Search messages or files"
          />
          <button type="submit" :disabled="loading">
            {{ loading ? 'Searching...' : 'Search' }}
          </button>
        </form>

        <nav class="tabs" aria-label="Search type">
          <button
            type="button"
            :class="{ active: activeTab === 'messages' }"
            @click="switchTab('messages')"
          >
            Messages
          </button>
          <button
            type="button"
            :class="{ active: activeTab === 'files' }"
            @click="switchTab('files')"
          >
            Files
          </button>
        </nav>

        <div class="body">
          <p v-if="errorMsg" class="status error">{{ errorMsg }}</p>
          <div v-if="loading" class="status">Searching...</div>
          <div v-else-if="!hasSearched && results.length === 0" class="status">
            Type a keyword to search visible conversations.
          </div>
          <div v-else-if="hasSearched && results.length === 0" class="status">
            No backend results for this keyword.
          </div>

          <article
            v-for="result in results"
            :key="activeTab === 'messages' ? (result as SearchMessageItem).message_id : (result as SearchFileItem).file_id"
            class="result-card"
            @click="openConversation(result.conversation_id)"
          >
            <template v-if="activeTab === 'messages'">
              <div class="result-header">
                <span>{{ (result as SearchMessageItem).sender_nickname || (result as SearchMessageItem).sender_id }}</span>
                <small>{{ conversationTypeLabel(result.conversation_type) }} / {{ result.created_at }}</small>
              </div>
              <p>{{ (result as SearchMessageItem).content }}</p>
            </template>

            <template v-else>
              <div class="result-header">
                <span>{{ (result as SearchFileItem).original_name }}</span>
                <small>{{ conversationTypeLabel(result.conversation_type) }} / {{ result.created_at }}</small>
              </div>
              <p>
                {{ (result as SearchFileItem).uploader_nickname || (result as SearchFileItem).uploader_id }}
                /
                {{ formatFileSize((result as SearchFileItem).file_size) }}
                /
                {{ (result as SearchFileItem).mime_type || 'Unknown type' }}
              </p>
            </template>
          </article>

          <button v-if="hasMore && !loading" class="load-more" type="button" @click="loadMore">
            Load more
          </button>
        </div>
      </section>
    </div>
  </Teleport>
</template>

<style scoped>
.search-layer {
  position: fixed;
  inset: 0;
  z-index: 70;
  display: grid;
  place-items: start center;
  padding: min(7vh, 56px) 18px 18px;
}

.search-mask {
  position: absolute;
  inset: 0;
  border: 0;
  background: rgba(0, 0, 0, 0.56);
}

.search-panel {
  position: relative;
  z-index: 1;
  display: flex;
  width: min(780px, 96vw);
  max-height: min(760px, 88vh);
  min-height: 420px;
  flex-direction: column;
  overflow: hidden;
  border: 1px solid rgba(240, 207, 132, 0.18);
  border-radius: 18px;
  color: var(--text);
  background: rgba(12, 14, 15, 0.94);
  box-shadow: 0 36px 90px rgba(0, 0, 0, 0.55);
  backdrop-filter: blur(24px);
}

header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 14px;
  border-bottom: 1px solid var(--border);
  padding: 16px 18px;
}

header strong,
header small {
  display: block;
}

header strong {
  color: #fff7e8;
  font-size: 18px;
}

header small {
  margin-top: 4px;
  color: var(--text-muted);
  font-size: 12px;
}

button {
  height: 36px;
  border: 1px solid rgba(240, 207, 132, 0.16);
  border-radius: 9px;
  color: var(--text-soft);
  background: rgba(255, 255, 255, 0.07);
  cursor: pointer;
  font-weight: 760;
}

button:hover {
  border-color: rgba(240, 207, 132, 0.3);
  color: var(--text);
  background: rgba(255, 255, 255, 0.1);
}

.search-form {
  display: grid;
  grid-template-columns: minmax(0, 1fr) 110px;
  gap: 10px;
  border-bottom: 1px solid rgba(240, 207, 132, 0.1);
  padding: 14px 18px;
}

input {
  height: 42px;
  min-width: 0;
  border: 1px solid rgba(240, 207, 132, 0.16);
  border-radius: 10px;
  padding: 0 13px;
  color: var(--text);
  background: rgba(0, 0, 0, 0.22);
}

.search-form button {
  height: 42px;
  color: #171009;
  background: linear-gradient(135deg, #f0cf84, #b68a3e);
}

.tabs {
  display: flex;
  gap: 6px;
  border-bottom: 1px solid var(--border);
  padding: 0 18px;
}

.tabs button {
  border: 0;
  border-bottom: 2px solid transparent;
  border-radius: 0;
  padding: 0 14px;
  color: var(--text-muted);
  background: transparent;
}

.tabs button.active {
  color: var(--accent-strong);
  border-bottom-color: var(--accent-strong);
}

.body {
  min-height: 0;
  flex: 1 1 auto;
  overflow-y: auto;
  padding: 16px 18px 18px;
}

.status {
  border: 1px dashed rgba(240, 207, 132, 0.16);
  border-radius: 14px;
  padding: 28px 16px;
  color: var(--text-muted);
  background: rgba(255, 255, 255, 0.04);
  text-align: center;
}

.status.error {
  color: #ffd4d4;
  background: rgba(239, 68, 68, 0.14);
}

.result-card {
  border: 1px solid rgba(240, 207, 132, 0.12);
  border-radius: 12px;
  padding: 13px;
  margin-bottom: 10px;
  background: rgba(255, 255, 255, 0.055);
  cursor: pointer;
}

.result-card:hover {
  border-color: rgba(240, 207, 132, 0.26);
  background: rgba(255, 255, 255, 0.085);
}

.result-header {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 12px;
}

.result-header span {
  min-width: 0;
  overflow: hidden;
  color: var(--text);
  text-overflow: ellipsis;
  white-space: nowrap;
  font-weight: 820;
}

.result-header small {
  flex: 0 0 auto;
  color: var(--text-muted);
  font-size: 12px;
}

.result-card p {
  margin: 8px 0 0;
  overflow: hidden;
  color: var(--text-muted);
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 13px;
}

.load-more {
  width: 100%;
  margin-top: 8px;
}

@media (max-width: 620px) {
  .search-form {
    grid-template-columns: 1fr;
  }

  .result-header {
    display: block;
  }

  .result-header small {
    display: block;
    margin-top: 4px;
  }
}
</style>
