<script setup>
import { computed } from 'vue'
import { NCollapse, NCollapseItem, NTag, NSpace } from 'naive-ui'

const props = defineProps({
  data: { type: String, default: '' },
  mode: { type: String, default: 'auto' } // auto | request | response
})

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
  // Request format: { messages: [...] }
  if (parsed.value.messages) return parsed.value.messages
  // Response format: { choices: [{ message: {...} }] }
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

const roleColors = {
  system: 'info',
  user: 'primary',
  assistant: 'success',
  tool: 'warning',
  function: 'warning'
}

function truncate(text, max = 500) {
  if (!text) return ''
  if (text.length <= max) return text
  return text.slice(0, max) + '...'
}
</script>

<template>
  <div>
    <!-- Model name -->
    <div v-if="modelInfo" style="margin-bottom: 12px;">
      <n-tag size="small" type="primary" style="margin-right: 6px;">{{ modelInfo }}</n-tag>
      <n-tag v-if="parsed.stream" size="small" type="warning">stream</n-tag>
    </div>

    <!-- Non-LLM content fallback -->
    <div v-if="!hasLLMContent && parsed">
      <pre style="background: #11111b; border-radius: 4px; padding: 12px 16px; font-size: 12px; line-height: 1.6; overflow-x: auto; color: #cdd6f4; white-space: pre-wrap;">{{ JSON.stringify(parsed, null, 2) }}</pre>
    </div>

    <!-- Messages chat view -->
    <div v-if="messages" style="display: flex; flex-direction: column; gap: 10px;">
      <div
        v-for="(msg, i) in messages"
        :key="i"
        style="background: #1e1e2e; border-radius: 8px; padding: 12px 16px; border-left: 3px solid transparent;"
        :style="{
          borderLeftColor: msg.role === 'user' ? '#89b4fa' : msg.role === 'assistant' ? '#a6e3a1' : msg.role === 'system' ? '#cba6f7' : '#fab387'
        }"
      >
        <!-- Role header -->
        <div style="display: flex; align-items: center; gap: 8px; margin-bottom: 8px;">
          <n-tag :type="roleColors[msg.role] || 'default'" size="small">
            {{ msg.role || 'unknown' }}
          </n-tag>
          <span v-if="msg.name" style="color: var(--text-muted); font-size: 12px;">{{ msg.name }}</span>
        </div>

        <!-- Content -->
        <div v-if="msg.content" style="color: var(--text-primary); font-size: 13px; line-height: 1.6; white-space: pre-wrap; word-break: break-word;">
          {{ msg.content }}
        </div>
        <div v-else style="color: var(--text-muted); font-size: 12px; font-style: italic;">（空）</div>

        <!-- Reasoning content (collapsible) -->
        <div v-if="msg.reasoning_content" style="margin-top: 8px;">
          <details>
            <summary style="color: #fab387; font-size: 12px; cursor: pointer;">🧠 reasoning_content ({{ msg.reasoning_content.length }} 字符)</summary>
            <div style="margin-top: 6px; padding: 8px 12px; background: #11111b; border-radius: 4px; color: #fab387; font-size: 12px; line-height: 1.5; white-space: pre-wrap; word-break: break-word;">
              {{ msg.reasoning_content }}
            </div>
          </details>
        </div>

        <!-- Tool calls -->
        <div v-if="msg.tool_calls && msg.tool_calls.length" style="margin-top: 8px;">
          <div style="color: #fab387; font-size: 12px; margin-bottom: 4px;">🔧 tool_calls:</div>
          <div v-for="(tc, j) in msg.tool_calls" :key="j" style="margin-top: 4px;">
            <div v-if="tc.function" style="padding: 8px; background: #11111b; border-radius: 4px;">
              <code style="color: #89dceb; font-size: 12px;">{{ tc.function.name }}</code>
              <pre style="color: #cdd6f4; font-size: 11px; margin-top: 4px; white-space: pre-wrap;">{{ tc.function.arguments }}</pre>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Tools section -->
    <div v-if="tools" style="margin-top: 16px;">
      <n-collapse>
        <n-collapse-item title="🔧 工具定义 ({{ tools.length }})" name="tools">
          <div v-for="(tool, i) in tools" :key="i" style="margin-bottom: 8px; padding: 8px; background: #11111b; border-radius: 4px;">
            <div v-if="tool.function">
              <code style="color: #89dceb; font-size: 12px;">{{ tool.function.name }}</code>
              <div style="color: var(--text-muted); font-size: 11px; margin-top: 2px;">{{ tool.function.description || '' }}</div>
              <pre v-if="tool.function.parameters" style="color: #cdd6f4; font-size: 11px; margin-top: 4px; white-space: pre-wrap;">{{ JSON.stringify(tool.function.parameters, null, 2) }}</pre>
            </div>
          </div>
          <div v-if="toolChoice" style="color: var(--text-muted); font-size: 11px; margin-top: 4px;">
            tool_choice: <code style="color: #a6e3a1;">{{ JSON.stringify(toolChoice) }}</code>
          </div>
        </n-collapse-item>
      </n-collapse>
    </div>

    <!-- Empty state -->
    <div v-if="!parsed && data" style="color: var(--text-muted); text-align: center; padding: 20px;">
      无法解析 JSON
    </div>

    <!-- Raw JSON toggle -->
    <div v-if="hasLLMContent" style="margin-top: 16px;">
      <details>
        <summary style="color: var(--text-muted); font-size: 12px; cursor: pointer;">查看原始 JSON</summary>
        <pre style="background: #11111b; border-radius: 4px; padding: 12px 16px; font-size: 11px; line-height: 1.5; overflow-x: auto; color: #6c7086; margin-top: 8px; white-space: pre-wrap;">{{ JSON.stringify(parsed, null, 2) }}</pre>
      </details>
    </div>
  </div>
</template>
