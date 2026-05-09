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
  <div v-if="headers.length > 0" class="headers-viewer">
    <div
      v-for="h in headers" :key="h.key"
      class="header-row"
    >
      <span class="header-key">{{ h.key }}</span>
      <span class="header-value">{{ isMaskable(h.key) ? maskIf(h.value) : h.value }}</span>
    </div>
  </div>
  <div v-else class="empty-state">（无内容）</div>
</template>

<style scoped>
.headers-viewer {
  font-size: 12px;
  font-family: 'SF Mono', 'Fira Code', monospace;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  overflow: hidden;
}
.header-row {
  display: flex;
  gap: 12px;
  padding: 6px 14px;
  border-bottom: 1px solid var(--border-color);
  align-items: baseline;
}
.header-row:last-child { border-bottom: none; }
.header-key {
  color: var(--accent);
  white-space: nowrap;
  min-width: 160px;
  flex-shrink: 0;
  font-weight: 500;
}
.header-value {
  color: var(--text-primary);
  word-break: break-all;
}
.empty-state {
  color: var(--text-muted);
  font-size: 12px;
  text-align: center;
  padding: 16px;
}
</style>
