<script setup>
import { computed } from 'vue'
import { NButton, NIcon } from 'naive-ui'
import { CopyOutline } from '@vicons/ionicons5'

const props = defineProps({ data: { type: String, default: '' } })

const formatted = computed(() => {
  if (!props.data) return ''
  try {
    return JSON.stringify(JSON.parse(props.data), null, 2)
  } catch {
    return props.data
  }
})

function copy() {
  navigator.clipboard.writeText(formatted.value)
}
</script>

<template>
  <div class="json-viewer">
    <div class="copy-btn">
      <n-button quaternary size="tiny" @click="copy">
        <template #icon><n-icon size="15"><CopyOutline /></n-icon></template>
      </n-button>
    </div>
    <pre class="json-pre"><code>{{ formatted }}</code></pre>
  </div>
</template>

<style scoped>
.json-viewer {
  position: relative;
}
.copy-btn {
  position: absolute;
  top: 8px;
  right: 8px;
  z-index: 1;
}
.json-pre {
  background: var(--bg-code);
  border-radius: var(--radius-sm);
  padding: 14px 18px;
  font-size: 12px;
  line-height: 1.7;
  overflow-x: auto;
  color: var(--text-primary);
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
  border: 1px solid var(--border-color);
}
.json-pre code {
  font-family: 'SF Mono', 'Fira Code', 'Cascadia Code', monospace;
}
</style>
