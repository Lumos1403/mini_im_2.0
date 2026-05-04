<script setup lang="ts">
import { computed, ref, watch } from 'vue'

const props = defineProps<{
  url: string
  alt?: string
}>()

const loadStatus = ref<'loading' | 'loaded' | 'error'>('loading')

const isAllowedURL = computed(() => {
  try {
    const url = new URL(props.url)
    return url.protocol === 'https:' && url.hostname === 'mermaid.ink' && /^\/(img|svg)\//.test(url.pathname)
  } catch {
    return false
  }
})

watch(
  () => props.url,
  () => {
    loadStatus.value = 'loading'
  },
)

function handleLoad() {
  loadStatus.value = 'loaded'
}

function handleError() {
  loadStatus.value = 'error'
}
</script>

<template>
  <div class="mermaid-preview">
    <div v-if="!isAllowedURL" class="mermaid-state error">Invalid diagram link</div>

    <template v-else>
      <div v-if="loadStatus === 'loading'" class="mermaid-skeleton">
        <span class="skeleton-line wide"></span>
        <span class="skeleton-line"></span>
        <span class="skeleton-label">图表生成中</span>
      </div>

      <a
        v-if="loadStatus === 'error'"
        class="mermaid-state error"
        :href="url"
        target="_blank"
        rel="noopener noreferrer"
      >
        图表加载失败
      </a>

      <a
        v-show="loadStatus === 'loaded'"
        class="mermaid-link"
        :href="url"
        target="_blank"
        rel="noopener noreferrer"
      >
        <img class="mermaid-image" :src="url" :alt="alt || 'Mermaid diagram'" @load="handleLoad" @error="handleError" />
      </a>

      <img
        v-if="loadStatus === 'loading'"
        class="hidden-loader"
        :src="url"
        :alt="alt || 'Mermaid diagram'"
        @load="handleLoad"
        @error="handleError"
      />
    </template>
  </div>
</template>

<style scoped>
.mermaid-preview {
  display: block;
  margin: 10px 0 2px;
}

.mermaid-skeleton,
.mermaid-state {
  width: min(480px, 100%);
  min-height: 140px;
  border: 1px solid rgba(240, 207, 132, 0.14);
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.055);
}

.mermaid-skeleton {
  display: grid;
  align-content: center;
  gap: 10px;
  padding: 18px;
  overflow: hidden;
}

.skeleton-line {
  display: block;
  height: 12px;
  width: 62%;
  border-radius: 999px;
  background: linear-gradient(90deg, rgba(240, 207, 132, 0.1), rgba(240, 207, 132, 0.24), rgba(240, 207, 132, 0.1));
  background-size: 220% 100%;
  animation: shimmer 1.4s ease-in-out infinite;
}

.skeleton-line.wide {
  width: 84%;
}

.skeleton-label {
  margin-top: 4px;
  color: var(--text-muted);
  font-size: 12px;
  font-weight: 760;
}

.mermaid-state {
  display: grid;
  min-height: 72px;
  place-items: center;
  color: var(--text-muted);
  text-decoration: none;
  font-size: 13px;
  font-weight: 760;
}

.mermaid-state.error {
  color: #ffd4d4;
  background: rgba(239, 68, 68, 0.12);
}

.mermaid-link {
  display: block;
}

.mermaid-image {
  display: block;
  width: min(480px, 100%);
  max-height: 360px;
  border: 1px solid rgba(240, 207, 132, 0.14);
  border-radius: 10px;
  background: #fff;
  object-fit: contain;
}

.hidden-loader {
  display: none;
}

@keyframes shimmer {
  0% {
    background-position: 120% 0;
  }

  100% {
    background-position: -120% 0;
  }
}
</style>
