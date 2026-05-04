<script setup lang="ts">
import { computed, onBeforeUnmount, ref, watch } from 'vue'

import MermaidImagePreview from './MermaidImagePreview.vue'

const props = defineProps<{
  content: string
  streaming?: boolean
  error?: boolean
  errorMessage?: string
}>()

const emit = defineEmits<{
  (event: 'typed'): void
}>()

type MessagePart =
  | { type: 'text'; value: string }
  | { type: 'image'; alt: string; url: string }
  | { type: 'mermaid_image'; alt: string; url: string }
  | { type: 'mermaid_pending' }

const markdownImagePattern = /!\[([^\]\n]*)\]\((https?:\/\/[^\s)]+)\)/g
const bareMermaidURLPattern = /https:\/\/mermaid\.ink\/(?:img|svg)\/[^\s)]+/gi
const pseudoStreamIntervalMS = 18

const visibleContent = ref(props.streaming ? '' : props.content || '')
const targetContent = ref(props.content || '')
const hasStreamed = ref(Boolean(props.streaming))
let pseudoStreamTimer: number | undefined

const visualStreaming = computed(() =>
  Boolean(!props.error && hasStreamed.value && (props.streaming || visibleContent.value.length < targetContent.value.length)),
)
const parts = computed(() => parseMessageParts(visibleContent.value || '', visualStreaming.value))
const hasContent = computed(() => visibleContent.value.trim().length > 0)

watch(
  () => [props.content, props.streaming, props.error] as const,
  () => {
    if (props.streaming) {
      hasStreamed.value = true
    }
    targetContent.value = props.content || ''

    if (props.error) {
      stopPseudoStream()
      visibleContent.value = targetContent.value
      return
    }

    if (!hasStreamed.value) {
      stopPseudoStream()
      visibleContent.value = targetContent.value
      return
    }

    if (!targetContent.value.startsWith(visibleContent.value)) {
      if (props.streaming) {
        const mergedTarget = buildForwardOnlyTarget(visibleContent.value, targetContent.value)
        visibleContent.value = mergedTarget.slice(0, visibleContent.value.length)
        targetContent.value = mergedTarget
      } else {
        visibleContent.value = targetContent.value.slice(0, commonPrefixLength(visibleContent.value, targetContent.value))
      }
    }

    if (visibleContent.value.length < targetContent.value.length) {
      startPseudoStream()
    } else {
      stopPseudoStream()
    }
  },
  { immediate: true },
)

onBeforeUnmount(() => {
  stopPseudoStream()
})

function startPseudoStream() {
  if (pseudoStreamTimer) {
    return
  }
  pseudoStreamTimer = window.setInterval(() => {
    if (props.error) {
      stopPseudoStream()
      return
    }
    if (visibleContent.value.length >= targetContent.value.length) {
      stopPseudoStream()
      return
    }

    const remaining = targetContent.value.length - visibleContent.value.length
    visibleContent.value = targetContent.value.slice(0, visibleContent.value.length + pseudoStreamBatchSize(remaining))
    emit('typed')
  }, pseudoStreamIntervalMS)
}

function stopPseudoStream() {
  if (!pseudoStreamTimer) {
    return
  }
  window.clearInterval(pseudoStreamTimer)
  pseudoStreamTimer = undefined
}

function pseudoStreamBatchSize(remaining: number) {
  if (remaining > 240) {
    return 8
  }
  if (remaining > 120) {
    return 5
  }
  if (remaining > 60) {
    return 3
  }
  return 1
}

function buildForwardOnlyTarget(current: string, incoming: string) {
  const currentTrimmed = current.trim()
  const incomingTrimmed = incoming.trim()
  if (!currentTrimmed) {
    return incoming
  }
  if (!incomingTrimmed || current.includes(incomingTrimmed)) {
    return current
  }
  return `${current.trimEnd()}\n\n${incoming.trimStart()}`
}

function commonPrefixLength(left: string, right: string) {
  const maxLength = Math.min(left.length, right.length)
  let index = 0
  while (index < maxLength && left[index] === right[index]) {
    index += 1
  }
  return index
}

function parseMessageParts(content: string, streaming: boolean): MessagePart[] {
  const parts: MessagePart[] = []
  markdownImagePattern.lastIndex = 0

  let lastIndex = 0
  let match: RegExpExecArray | null
  while ((match = markdownImagePattern.exec(content)) !== null) {
    const [source, alt, url] = match
    appendTextSegment(parts, content.slice(lastIndex, match.index), streaming)
    if (isMermaidImageURL(url)) {
      parts.push({ type: 'mermaid_image', alt: alt.trim() || 'Mermaid diagram', url })
    } else if (isRenderableImageURL(url)) {
      parts.push({ type: 'image', alt: alt.trim() || 'image', url })
    } else {
      parts.push({ type: 'text', value: source })
    }
    lastIndex = match.index + source.length
  }

  appendTextSegment(parts, content.slice(lastIndex), streaming)
  return parts.length > 0 ? parts : [{ type: 'text', value: content }]
}

function appendTextSegment(parts: MessagePart[], segment: string, streaming: boolean) {
  if (!segment) {
    return
  }

  if (streaming) {
    const fenceIndex = segment.toLowerCase().lastIndexOf('```mermaid')
    if (fenceIndex >= 0) {
      const afterFence = segment.slice(fenceIndex + '```mermaid'.length)
      if (!afterFence.includes('```')) {
        appendTextWithBareMermaidLinks(parts, segment.slice(0, fenceIndex))
        parts.push({ type: 'mermaid_pending' })
        return
      }
    }
  }

  const pendingIndex = streaming ? findPendingMermaidIndex(segment) : -1
  appendTextWithBareMermaidLinks(parts, pendingIndex >= 0 ? segment.slice(0, pendingIndex) : segment)
  if (pendingIndex >= 0) {
    parts.push({ type: 'mermaid_pending' })
  }
}

function appendTextWithBareMermaidLinks(parts: MessagePart[], segment: string) {
  bareMermaidURLPattern.lastIndex = 0
  let lastIndex = 0
  let match: RegExpExecArray | null
  while ((match = bareMermaidURLPattern.exec(segment)) !== null) {
    if (match.index > lastIndex) {
      parts.push({ type: 'text', value: segment.slice(lastIndex, match.index) })
    }
    parts.push({ type: 'mermaid_image', alt: 'Mermaid diagram', url: match[0] })
    lastIndex = match.index + match[0].length
  }
  if (lastIndex < segment.length) {
    parts.push({ type: 'text', value: segment.slice(lastIndex) })
  }
}

function findPendingMermaidIndex(segment: string) {
  const lower = segment.toLowerCase()
  const imageStart = lower.lastIndexOf('![')
  if (imageStart >= 0) {
    const tail = lower.slice(imageStart)
    if (tail.includes('(https://mermaid.ink') && !tail.includes(')')) {
      return imageStart
    }
  }

  const linkStart = lower.lastIndexOf('https://mermaid.ink')
  if (linkStart >= 0) {
    const tail = lower.slice(linkStart)
    const fullLink = tail.match(/^https:\/\/mermaid\.ink\/(?:img|svg)\/[^\s)]+/)
    if (!fullLink || /\/(?:img|svg)\/?$/.test(tail)) {
      return linkStart
    }
  }
  return -1
}

function isMermaidImageURL(value: string) {
  try {
    const url = new URL(value)
    return url.protocol === 'https:' && url.hostname === 'mermaid.ink' && /^\/(img|svg)\//.test(url.pathname)
  } catch {
    return false
  }
}

function isRenderableImageURL(value: string) {
  try {
    const url = new URL(value)
    return url.protocol === 'http:' || url.protocol === 'https:'
  } catch {
    return false
  }
}
</script>

<template>
  <div class="streaming-content">
    <div v-if="visualStreaming && !hasContent && !error" class="agent-thinking">
      <span>Agent is replying</span>
      <span class="loading-dots" aria-hidden="true">
        <i></i>
        <i></i>
        <i></i>
      </span>
    </div>

    <template v-else>
      <template v-for="(part, index) in parts" :key="index">
        <span v-if="part.type === 'text'" class="text-message-copy">{{ part.value }}</span>
        <a
          v-else-if="part.type === 'image'"
          class="markdown-image-link"
          :href="part.url"
          target="_blank"
          rel="noopener noreferrer"
        >
          <img class="markdown-image" :src="part.url" :alt="part.alt" loading="lazy" />
        </a>
        <MermaidImagePreview
          v-else-if="part.type === 'mermaid_image'"
          :url="part.url"
          :alt="part.alt"
        />
        <div v-else class="mermaid-pending">
          <span class="pending-grid"></span>
          <span>Generating diagram</span>
        </div>
      </template>
      <span v-if="visualStreaming && !error" class="stream-cursor" aria-hidden="true"></span>
    </template>

    <div v-if="error" class="stream-error">{{ errorMessage || 'Agent generation failed' }}</div>
  </div>
</template>

<style scoped>
.streaming-content {
  overflow-wrap: anywhere;
  line-height: 1.55;
}

.text-message-copy {
  white-space: pre-wrap;
}

.agent-thinking {
  display: inline-flex;
  align-items: center;
  gap: 9px;
  color: var(--text-muted);
  font-size: 13px;
  font-weight: 760;
}

.loading-dots {
  display: inline-flex;
  gap: 4px;
}

.loading-dots i {
  width: 5px;
  height: 5px;
  border-radius: 999px;
  background: var(--accent-strong);
  opacity: 0.36;
  animation: dotPulse 1.1s ease-in-out infinite;
}

.loading-dots i:nth-child(2) {
  animation-delay: 0.15s;
}

.loading-dots i:nth-child(3) {
  animation-delay: 0.3s;
}

.stream-cursor {
  display: inline-block;
  width: 7px;
  height: 1.1em;
  margin-left: 2px;
  border-radius: 999px;
  background: var(--accent-strong);
  vertical-align: -0.18em;
  animation: cursorBlink 0.9s steps(2, start) infinite;
}

.markdown-image-link {
  display: block;
  margin: 10px 0 2px;
}

.markdown-image {
  display: block;
  width: min(480px, 100%);
  max-height: 360px;
  border: 1px solid rgba(240, 207, 132, 0.14);
  border-radius: 10px;
  background: #fff;
  object-fit: contain;
}

.mermaid-pending {
  display: grid;
  width: min(480px, 100%);
  min-height: 116px;
  align-content: center;
  gap: 11px;
  margin: 10px 0 2px;
  border: 1px solid rgba(240, 207, 132, 0.14);
  border-radius: 10px;
  padding: 16px;
  color: var(--text-muted);
  background: rgba(255, 255, 255, 0.055);
  font-size: 12px;
  font-weight: 760;
}

.pending-grid {
  width: 70%;
  height: 42px;
  border-radius: 8px;
  background:
    linear-gradient(90deg, rgba(240, 207, 132, 0.08), rgba(240, 207, 132, 0.22), rgba(240, 207, 132, 0.08)),
    repeating-linear-gradient(90deg, transparent 0 22px, rgba(240, 207, 132, 0.08) 22px 23px),
    repeating-linear-gradient(0deg, transparent 0 16px, rgba(240, 207, 132, 0.08) 16px 17px);
  background-size: 220% 100%, auto, auto;
  animation: shimmer 1.4s ease-in-out infinite;
}

.stream-error {
  margin-top: 7px;
  color: #ffd4d4;
  font-size: 12px;
  font-weight: 760;
}

@keyframes dotPulse {
  0%,
  80%,
  100% {
    opacity: 0.28;
    transform: translateY(0);
  }

  40% {
    opacity: 0.9;
    transform: translateY(-2px);
  }
}

@keyframes cursorBlink {
  0%,
  45% {
    opacity: 1;
  }

  46%,
  100% {
    opacity: 0;
  }
}

@keyframes shimmer {
  0% {
    background-position: 120% 0, 0 0, 0 0;
  }

  100% {
    background-position: -120% 0, 0 0, 0 0;
  }
}
</style>
