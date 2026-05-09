<script setup>
import { computed } from 'vue'

const props = defineProps({
  data: { type: String, default: '' },
  maxLen: { type: Number, default: 80 }
})

const headers = computed(() => {
  if (!props.data) return []
  try {
    const obj = JSON.parse(props.data)
    return Object.entries(obj).map(([key, value]) => ({ key, value }))
  } catch {
    return []
  }
})

function maskIf(token) {
  const lower = token.toLowerCase()
  if (lower.startsWith('bearer ') || lower.startsWith('sk-')) {
    return lower.slice(0, 12) + '...'
  }
  if (token.length > props.maxLen) {
    return token.slice(0, props.maxLen) + '...'
  }
  return token
}

function isMaskable(key) {
  return ['authorization', 'api-key', 'x-api-key'].includes(key.toLowerCase())
}
</script>

<template>
  <div v-if="headers.length > 0" style="font-size: 12px; font-family: monospace;">
    <div
      v-for="h in headers" :key="h.key"
      style="display: flex; gap: 12px; padding: 3px 0; border-bottom: 1px solid var(--border-color); align-items: baseline;"
    >
      <span style="color: #89b4fa; white-space: nowrap; min-width: 160px; flex-shrink: 0;">{{ h.key }}</span>
      <span style="color: var(--text-primary); word-break: break-all;">{{ isMaskable(h.key) ? maskIf(h.value) : h.value }}</span>
    </div>
  </div>
  <div v-else style="color: var(--text-muted); font-size: 12px; text-align: center; padding: 10px;">（无内容）</div>
</template>
