<script setup>
import { computed } from 'vue'
import { NCollapse, NCollapseItem, NTag, NButton } from 'naive-ui'
import MarkdownRenderer from './MarkdownRenderer.vue'

const props = defineProps({
  data: { type: String, default: '' },
  mode: { type: String, default: 'auto' },
  showRaw: { type: Boolean, default: false }
})

const emit = defineEmits(['update:showRaw'])

const parsed = computed(() => {
  if (!props.data) return null
  try {
    return JSON.parse(props.data)
  } catch {
    return null
  }
})

const modelInfo = computed(() => {
  if (!parsed.value) return null
  if (parsed.value.model) return parsed.value.model
  return null
})

const messages = computed(() => {
  if (!parsed.value) return null
  if (parsed.value.messages) return parsed.value.messages
  if (parsed.value.choices) {
    return parsed.value.choices
      .filter(c => c.message)
      .map(c => c.message)
  }
  return null
})

const tools = computed(() => {
  if (!parsed.value) return null
  if (parsed.value.tools) return parsed.value.tools
  return null
})

const toolChoice = computed(() => {
  if (!parsed.value) return null
  return parsed.value.tool_choice || null
})

const hasLLMContent = computed(() => {
  return messages.value || tools.value
})

// Detect if content is HTML and format it with proper indentation
function formatHtml(text) {
  if (!text) return null
  // Only detect as HTML if there are structural block-level tags
  if (!/<\/?(html|div|table|tr|td|th|tbody|thead|ul|ol|li|p|h[1-6]|span|section|article|header|footer|main|nav|form|input|select|option|button|a|img|pre|code|blockquote|dl|dt|dd)[\s>]/i.test(text)) return null
  let indent = 0
  const lines = []
  const tagRegex = /(<\/?[\w-]+(?:\s[^>]*)?>)|([^<]+)/g
  let match
  while ((match = tagRegex.exec(text)) !== null) {
    if (match[2]) {
      const trimmed = match[2].trim()
      if (trimmed) lines.push('  '.repeat(indent) + trimmed)
    } else if (match[1]) {
      const isClosing = match[1].startsWith('</')
      const isSelfClosing = match[1].endsWith('/>')
      if (isClosing) indent = Math.max(0, indent - 1)
      lines.push('  '.repeat(indent) + match[1])
      if (!isClosing && !isSelfClosing) indent++
    }
  }
  return lines.filter(l => l.trim()).join('\n')
}

const roleColors = {
  system: 'info',
  user: 'primary',
  assistant: 'success',
  tool: 'warning',
  function: 'warning'
}
</script>

<template>
  <div>
    <!-- Model info row + raw toggle -->
    <div v-if="hasLLMContent" style="display: flex; align-items: center; gap: 8px; margin-bottom: 12px;">
      <n-tag v-if="modelInfo" size="small" type="primary">{{ modelInfo }}</n-tag>
      <n-tag v-if="parsed && parsed.stream" size="small" type="warning">stream</n-tag>
      <span style="flex: 1;"></span>
      <n-button quaternary size="tiny" @click="emit('update:showRaw', !props.showRaw)" style="color: var(--text-muted);">
        {{ props.showRaw ? '📖 解析视图' : '📄 原始报文' }}
      </n-button>
    </div>

    <!-- Raw JSON mode -->
    <template v-if="props.showRaw">
      <pre style="background: var(--bg-code); border-radius: 4px; padding: 12px 16px; font-size: 12px; line-height: 1.6; overflow-x: auto; color: var(--text-primary); white-space: pre-wrap;">{{ JSON.stringify(parsed, null, 2) }}</pre>
    </template>

    <!-- Parsed mode -->
    <template v-if="!props.showRaw">
      <div v-if="!hasLLMContent && parsed">
        <pre style="background: var(--bg-code); border-radius: 4px; padding: 12px 16px; font-size: 12px; line-height: 1.6; overflow-x: auto; color: var(--text-primary); white-space: pre-wrap;">{{ JSON.stringify(parsed, null, 2) }}</pre>
      </div>

      <div v-if="messages" style="display: flex; flex-direction: column; gap: 10px;">
        <div
          v-for="(msg, i) in messages" :key="i"
          style="background: var(--bg-message); border-radius: 8px; padding: 12px 16px; border-left: 3px solid transparent;"
          :style="{ borderLeftColor: msg.role === 'user' ? '#89b4fa' : msg.role === 'assistant' ? '#a6e3a1' : msg.role === 'system' ? '#cba6f7' : '#fab387' }"
        >
          <div style="display: flex; align-items: center; gap: 8px; margin-bottom: 8px;">
            <n-tag :type="roleColors[msg.role] || 'default'" size="small">{{ msg.role || 'unknown' }}</n-tag>
            <span v-if="msg.name" style="color: var(--text-muted); font-size: 12px;">{{ msg.name }}</span>
          </div>

          <div v-if="msg.reasoning_content" style="margin-bottom: 8px; padding: 8px 12px; background: var(--bg-code); border-radius: 4px; color: #fab387; font-size: 12px; line-height: 1.5; white-space: pre-wrap; word-break: break-word; border-left: 2px solid #fab387;">
            <div style="color: #fab387; font-size: 11px; margin-bottom: 4px; opacity: 0.7;">🧠 思考过程</div>
            {{ msg.reasoning_content }}
          </div>

          <template v-if="msg.content">
            <div v-if="formatHtml(msg.content)" style="margin-bottom: 4px;">
              <div style="color: var(--text-muted); font-size: 11px; margin-bottom: 4px;">📄 HTML 输出</div>
              <pre style="background: var(--bg-code); border-radius: 4px; padding: 10px 12px; font-size: 12px; line-height: 1.5; overflow-x: auto; color: var(--text-primary); white-space: pre; tab-size: 2;"><code>{{ formatHtml(msg.content) }}</code></pre>
            </div>
            <markdown-renderer v-else :content="msg.content" />
          </template>
          <div v-else style="color: var(--text-muted); font-size: 12px; font-style: italic;">（空）</div>

          <div v-if="msg.tool_calls && msg.tool_calls.length" style="margin-top: 8px;">
            <div style="color: #fab387; font-size: 12px; margin-bottom: 4px;">🔧 tool_calls:</div>
            <div v-for="(tc, j) in msg.tool_calls" :key="j" style="margin-top: 4px;">
              <div v-if="tc.function" style="padding: 8px; background: var(--bg-code); border-radius: 4px;">
                <code style="color: #89dceb; font-size: 12px;">{{ tc.function.name }}</code>
                <pre style="color: var(--text-primary); font-size: 11px; margin-top: 4px; white-space: pre-wrap;">{{ tc.function.arguments }}</pre>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div v-if="tools" style="margin-top: 16px;">
        <n-collapse>
          <n-collapse-item title="🔧 工具定义 ({{ tools.length }})" name="tools">
            <div v-for="(tool, i) in tools" :key="i" style="margin-bottom: 8px; padding: 8px; background: var(--bg-code); border-radius: 4px;">
              <div v-if="tool.function">
                <code style="color: #89dceb; font-size: 12px;">{{ tool.function.name }}</code>
                <div style="color: var(--text-muted); font-size: 11px; margin-top: 2px;">{{ tool.function.description || '' }}</div>
                <pre v-if="tool.function.parameters" style="color: var(--text-primary); font-size: 11px; margin-top: 4px; white-space: pre-wrap;">{{ JSON.stringify(tool.function.parameters, null, 2) }}</pre>
              </div>
            </div>
            <div v-if="toolChoice" style="color: var(--text-muted); font-size: 11px; margin-top: 4px;">tool_choice: <code style="color: #a6e3a1;">{{ JSON.stringify(toolChoice) }}</code></div>
          </n-collapse-item>
        </n-collapse>
      </div>

      <div v-if="!parsed && data" style="color: var(--text-muted); text-align: center; padding: 20px;">无法解析 JSON</div>
    </template>
  </div>
</template>
