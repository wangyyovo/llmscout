<script setup>
import { computed } from 'vue'
import { NButton } from 'naive-ui'

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
  <div style="position: relative;">
    <div style="position: absolute; top: 8px; right: 8px;">
      <n-button quaternary size="tiny" @click="copy" style="color: #6c7086;">📋</n-button>
    </div>
    <pre style="background: #11111b; border-radius: 4px; padding: 12px 16px; font-size: 12px; line-height: 1.6; overflow-x: auto; color: #cdd6f4;"><code>{{ formatted }}</code></pre>
  </div>
</template>
