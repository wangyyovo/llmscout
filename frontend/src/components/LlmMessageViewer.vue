<script setup>
import { computed, ref, onErrorCaptured } from 'vue'
import { NCollapse, NCollapseItem, NTag, NButton } from 'naive-ui'
import MarkdownRenderer from './MarkdownRenderer.vue'

const props = defineProps({
  data: { type: String, default: '' },
  mode: { type: String, default: 'auto' },
  showRaw: { type: Boolean, default: false }
})

const emit = defineEmits(['update:showRaw'])
const renderFailed = ref(false)
onErrorCaptured((err) => {
  console.error('LlmMessageViewer render error:', err)
  renderFailed.value = true
  return false
})

const parsed = computed(() => {
  if (!props.data) return null
  // Try direct JSON parse
  try { return JSON.parse(props.data) } catch { /* continue */ }
  // Handle SSE raw text: extract all data lines and try to parse the last non-[DONE] event
  if (props.data.includes('data:')) {
    const lines = props.data.split('\n').filter(l => l.startsWith('data:'))
    const jsons = lines.map(l => l.slice(5).trim()).filter(s => s && s !== '[DONE]')
    // Try parsing the first valid JSON event
    for (const j of jsons) {
      try { return JSON.parse(j) } catch { /* skip */ }
    }
  }
  return null
})

// Extract streaming content from SSE body (OpenAI + Anthropic)
const sseContent = computed(() => {
  if (!props.data || !props.data.includes('data:')) return null
  const lines = props.data.split('\n').filter(l => l.startsWith('data:'))
  let content = '', reasoning = '', modelName = '', stopReason = ''
  for (const l of lines) {
    const j = l.slice(5).trim()
    if (!j || j === '[DONE]') continue
    try {
      const evt = JSON.parse(j)
      if (evt.model && !modelName) modelName = evt.model
      // OpenAI
      if (evt.choices) for (const c of evt.choices) {
        if (c.delta) {
          if (c.delta.content) content += c.delta.content
          if (c.delta.reasoning_content) reasoning += c.delta.reasoning_content
        }
        if (c.finish_reason) stopReason = c.finish_reason
      }
      // Anthropic
      if (evt.type === 'message_start' && evt.message && evt.message.model) modelName = evt.message.model
      if (evt.type === 'content_block_delta' && evt.delta) {
        if (evt.delta.type === 'text_delta' || evt.delta.type === 'text') content += evt.delta.text || ''
        if (evt.delta.type === 'thinking_delta') reasoning += evt.delta.thinking || ''
      }
      if (evt.type === 'message_delta' && evt.delta && evt.delta.stop_reason) stopReason = evt.delta.stop_reason
    } catch { /* skip */ }
  }
  if (content || reasoning || stopReason) return { content: content.trim(), reasoning: reasoning.trim(), model: modelName, stopReason }
  return null
})

const modelInfo = computed(() => {
  if (!parsed.value) return null
  if (parsed.value.model) return parsed.value.model
  return null
})

const messages = computed(() => {
  if (!parsed.value) return null
  // Request format: { messages: [...] }
  if (parsed.value.messages) {
    const msgs = []
    // Anthropic: system as separate field -> virtual system message
    if (parsed.value.system) {
      msgs.push({ role: 'system', content: typeof parsed.value.system === 'string' ? parsed.value.system : parsed.value.system.map(b => b.text || '').join('\n') })
    }
    for (const m of parsed.value.messages) {
      const msg = { ...m }
      // Anthropic: content is array of blocks -> normalize to string for markdown
      if (Array.isArray(m.content)) {
        const texts = []
        const toolCalls = []
        let reasoning = []
        for (const b of m.content) {
          if (b.type === 'text') texts.push(b.text)
          else if (b.type === 'thinking') reasoning.push(b.thinking || b.text || '')
          else if (b.type === 'tool_use') {
            toolCalls.push({ id: b.id, function: { name: b.name, arguments: JSON.stringify(b.input) } })
          }
          else if (b.type === 'tool_result') {
            if (!msg.tool_results) msg.tool_results = []
            msg.tool_results.push({ id: b.tool_use_id, content: b.content, isError: b.is_error || false })
          }
          else texts.push(JSON.stringify(b))
        }
        msg.content = texts.filter(t => t != null && ('' + t).trim()).map(t => ('' + t).trim()).join('\n')
        if (toolCalls.length) msg.tool_calls = (msg.tool_calls || []).concat(toolCalls)
        if (reasoning.length && !msg.reasoning_content) msg.reasoning_content = reasoning.filter(t => ('' + t).trim()).map(t => ('' + t).trim()).join('\n')
      }
      msgs.push(msg)
    }
    return msgs
  }
  // OpenAI response format: { choices: [{ message: {...} }] }
  if (parsed.value.choices) {
    const msgs = []
    for (const c of parsed.value.choices) {
      if (c.message) msgs.push(c.message)
      else if (c.delta && c.delta.content) msgs.push({ role: c.delta.role || 'assistant', content: c.delta.content })
      else if (c.delta && c.delta.reasoning_content) msgs.push({ role: 'assistant', reasoning_content: c.delta.reasoning_content })
    }
    if (msgs.length > 0) return msgs
  }
  // Anthropic response format: { role: "assistant", content: [...] }
  if (parsed.value.content && Array.isArray(parsed.value.content) && parsed.value.role) {
    const texts = []
    const toolCalls = []
    let reasoning = []
    for (const b of parsed.value.content) {
      if (b.type === 'text') texts.push(b.text)
      else if (b.type === 'thinking') reasoning.push(b.thinking || b.text || '')
      else if (b.type === 'tool_use') {
        toolCalls.push({ id: b.id, function: { name: b.name, arguments: JSON.stringify(b.input) } })
      }
      else texts.push(JSON.stringify(b))
    }
    const msg = { role: parsed.value.role, content: texts.filter(t => t != null && ('' + t).trim()).map(t => ('' + t).trim()).join('\n') }
    if (toolCalls.length) msg.tool_calls = toolCalls
    if (reasoning.length) msg.reasoning_content = reasoning.filter(t => ('' + t).trim()).map(t => ('' + t).trim()).join('\n')
    if (parsed.value.stop_reason) msg.stop_reason = parsed.value.stop_reason
    return [msg]
  }
  return null
})

const tools = computed(() => {
  if (!parsed.value) return null
  if (!parsed.value.tools) return null
  // Normalize both OpenAI and Anthropic tool formats
  return parsed.value.tools.map(t => {
    // OpenAI format: { type: "function", function: { name, description, parameters } }
    if (t.function) {
      return {
        name: t.function.name,
        description: t.function.description,
        inputSchema: t.function.parameters,
        type: t.type || 'function'
      }
    }
    // Anthropic format: { name, description, input_schema }
    if (t.name) {
      return {
        name: t.name,
        description: t.description || '',
        inputSchema: t.input_schema || null,
        type: 'function'
      }
    }
    return t
  })
})

const toolChoice = computed(() => {
  if (!parsed.value) return null
  return parsed.value.tool_choice || null
})

const hasLLMContent = computed(() => {
  return messages.value || tools.value
})

// Detect if content is HTML and format it with proper indentation via DOMParser
function formatHtml(text) {
  if (!text || typeof text !== 'string') return null
  let trimmed = text.trim()
  if (!/^<\w+[\s>]/i.test(trimmed)) return null
  if (!/<\/?(html|div|table|tr|td|th|tbody|thead|ul|ol|li|p|h[1-6]|span|section|article|header|footer|main|nav|form|input|select|option|button|a|img|pre|code|blockquote|dl|dt|dd)[\s>]/i.test(text)) return null
  try {
    const parser = new DOMParser()
    const doc = parser.parseFromString(text, 'text/html')
    const body = doc.body
    let result = ''
    function walk(node, depth) {
      let indent = '  '.repeat(depth)
      for (let child = node.firstChild; child; child = child.nextSibling) {
        if (child.nodeType === 3) {
          let t = (child.textContent || '').trim()
          if (t) result += indent + t + '\n'
        } else if (child.nodeType === 1) {
          let tag = child.tagName.toLowerCase()
          let selfClosing = ['br','hr','img','input','meta','link','area','base','col','embed','source','track','wbr'].includes(tag)
          let attrs = ''
          for (let a of child.attributes) {
            attrs += ' ' + a.name + '="' + a.value + '"'
          }
          if (selfClosing) {
            result += indent + '<' + tag + attrs + ' />\n'
          } else {
            result += indent + '<' + tag + attrs + '>\n'
            walk(child, depth + 1)
            result += indent + '</' + tag + '>\n'
          }
        }
      }
    }
    walk(body, 0)
    return result.trim()
  } catch {
    return null
  }
}

// Handle content that may be string, array (Anthropic format), or other
function toTextContent(content) {
  if (!content) return ''
  if (typeof content === 'string') return content.trim()
  if (Array.isArray(content)) {
    return content.filter(c => c.type === 'text').map(c => (c.text || '').trim()).filter(t => t).join('\n')
  }
  return JSON.stringify(content)
}

function summaryText(msg) {
  let t = msg.content || msg.reasoning_content || ''
  if (typeof t !== 'string') return '(非文本内容)'
  return t.length > 60 ? t.slice(0, 60) + '...' : t
}

function isFailed(msg, tc) {
  if (!msg.tool_results) return false
  return msg.tool_results.some(tr => tr.id === tc.id && tr.isError)
}

function tryParseJson(str) {
  if (!str) return null
  try { return JSON.parse(str) } catch { return null }
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
  <div v-if="renderFailed">
    <pre style="background: var(--bg-code); border-radius: 4px; padding: 12px; font-size: 12px; color: var(--text-primary); white-space: pre-wrap; overflow-x: auto;">{{ data || '（无数据）' }}</pre>
  </div>
  <div v-else>
    <!-- Model info row + raw toggle -->
    <div v-if="hasLLMContent || sseContent || parsed" style="display: flex; align-items: center; gap: 8px; margin-bottom: 12px;">
      <n-tag v-if="modelInfo || (sseContent && sseContent.model)" size="small" type="primary">{{ modelInfo || sseContent.model }}</n-tag>
      <n-tag v-if="parsed && parsed.stream" size="small" type="warning">stream</n-tag>
      <span style="flex: 1;"></span>
      <n-button quaternary size="tiny" @click="emit('update:showRaw', !props.showRaw)" style="color: var(--text-muted);">
        {{ props.showRaw ? (sseContent ? '📖 提取内容' : '📖 解析视图') : (sseContent ? '📄 原始报文' : '📄 原始报文') }}
      </n-button>
    </div>

    <!-- Raw mode: show raw JSON or raw SSE text -->
    <template v-if="props.showRaw">
      <template v-if="sseContent">
        <pre style="background: var(--bg-code); border-radius: 4px; padding: 12px 16px; font-size: 12px; line-height: 1.6; overflow-x: auto; color: var(--text-primary); white-space: pre-wrap;">{{ data }}</pre>
      </template>
      <template v-else>
        <pre style="background: var(--bg-code); border-radius: 4px; padding: 12px 16px; font-size: 12px; line-height: 1.6; overflow-x: auto; color: var(--text-primary); white-space: pre-wrap;">{{ JSON.stringify(parsed, null, 2) }}</pre>
      </template>
    </template>

    <!-- Parsed mode or SSE extracted view -->
    <template v-if="!props.showRaw">
      <!-- SSE extracted content -->
      <template v-if="sseContent">
        <div v-if="sseContent.reasoning" style="margin-bottom: 8px; padding: 8px 12px; background: var(--bg-code); border-radius: 4px; color: #fab387; font-size: 12px; line-height: 1.5; white-space: pre-wrap; word-break: break-word; border-left: 2px solid #fab387;">
          <div style="color: #fab387; font-size: 11px; margin-bottom: 4px; opacity: 0.7;">🧠 思考过程</div>
          {{ sseContent.reasoning }}
        </div>
        <div style="color: var(--text-primary); font-size: 13px; line-height: 1.6; white-space: pre-wrap; word-break: break-word;">
          <markdown-renderer v-if="sseContent.content" :content="sseContent.content" />
          <span v-else style="color: var(--text-muted); font-size: 12px; font-style: italic;">（空）</span>
        </div>
      </template>

      <!-- Non-SSE: no LLM content fallback -->
      <div v-else-if="!hasLLMContent && parsed">
        <pre style="background: var(--bg-code); border-radius: 4px; padding: 12px 16px; font-size: 12px; line-height: 1.6; overflow-x: auto; color: var(--text-primary); white-space: pre-wrap;">{{ JSON.stringify(parsed, null, 2) }}</pre>
      </div>

      <div v-if="messages" style="display: flex; flex-direction: column; gap: 10px;">
        <details
          v-for="(msg, i) in messages" :key="i"
          style="background: var(--bg-message); border-radius: 8px; border-left: 3px solid transparent;"
          :style="{ borderLeftColor: msg.role === 'user' ? '#89b4fa' : msg.role === 'assistant' ? '#a6e3a1' : msg.role === 'system' ? '#cba6f7' : '#fab387' }"
        >
          <summary style="padding: 10px 16px; cursor: pointer; display: flex; align-items: center; gap: 8px; user-select: none;">
            <n-tag :type="roleColors[msg.role] || 'default'" size="small">{{ msg.role || 'unknown' }}</n-tag>
            <span v-if="msg.tool_call_id" style="color: var(--text-muted); font-size: 10px; font-family: monospace;">{{ msg.tool_call_id }}</span>
            <span v-if="msg.name" style="color: var(--text-muted); font-size: 12px;">{{ msg.name }}</span>
            <span style="color: var(--text-muted); font-size: 11px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;">{{ summaryText(msg) }}</span>
          </summary>
          <div style="padding: 0 16px 12px 16px;">

          <div v-if="msg.reasoning_content" style="margin-bottom: 8px; padding: 8px 12px; background: var(--bg-code); border-radius: 4px; color: #fab387; font-size: 12px; line-height: 1.5; white-space: pre-wrap; word-break: break-word; border-left: 2px solid #fab387;">
            <div style="color: #fab387; font-size: 11px; margin-bottom: 4px; opacity: 0.7;">🧠 思考过程</div>
            {{ msg.reasoning_content }}
          </div>

          <template v-if="msg.content">
            <div v-if="formatHtml(msg.content)" style="margin-bottom: 4px;">
              <div style="color: var(--text-muted); font-size: 11px; margin-bottom: 4px;">📄 HTML 输出</div>
              <pre style="background: var(--bg-code); border-radius: 4px; padding: 10px 12px; font-size: 12px; line-height: 1.5; overflow-x: auto; color: var(--text-primary); white-space: pre; tab-size: 2;"><code>{{ formatHtml(msg.content) }}</code></pre>
            </div>
            <markdown-renderer v-else :content="toTextContent(msg.content)" />
          </template>
          <div v-else style="color: var(--text-muted); font-size: 12px; font-style: italic;">（空）</div>

          <div v-if="msg.tool_calls && msg.tool_calls.length" style="margin-top: 8px;">
            <div style="color: #89dceb; font-size: 12px; margin-bottom: 4px;">🔧 工具调用:</div>
            <div v-for="(tc, j) in msg.tool_calls" :key="j" style="margin-top: 4px;">
              <div v-if="tc.function" style="padding: 8px; background: var(--bg-code); border-radius: 4px;"
                :style="{ borderLeft: '2px solid ' + (isFailed(msg, tc) ? '#f38ba8' : '#89dceb') }">
                <div style="display: flex; align-items: center; gap: 8px;">
                  <code style="color: #89dceb; font-size: 13px; font-weight: bold;">{{ tc.function.name }}</code>
                  <span v-if="tc.id" style="color: var(--text-muted); font-size: 10px;">{{ tc.id }}</span>
                  <n-tag v-if="isFailed(msg, tc)" size="tiny" type="error">✗ 失败</n-tag>
                </div>
                <div v-if="tc.function.arguments" style="margin-top: 6px;">
                  <template v-if="tryParseJson(tc.function.arguments)">
                    <div v-for="(val, key) in tryParseJson(tc.function.arguments)" :key="key" style="display: flex; align-items: baseline; gap: 6px; padding: 3px 6px; font-size: 12px;">
                      <code style="color: #89dceb; white-space: nowrap;">{{ key }}</code>
                      <span style="color: var(--text-muted);">=</span>
                      <span style="color: #a6e3a1; word-break: break-all;">{{ typeof val === 'string' ? val : JSON.stringify(val) }}</span>
                    </div>
                  </template>
                  <pre v-else style="color: var(--text-primary); font-size: 11px; white-space: pre-wrap;">{{ tc.function.arguments }}</pre>
                </div>
              </div>
            </div>
          </div>

          <div v-if="msg.tool_results && msg.tool_results.length" style="margin-top: 8px;">
            <div style="color: #a6e3a1; font-size: 12px; margin-bottom: 4px;">📋 工具结果:</div>
            <div v-for="(tr, j) in msg.tool_results" :key="j" style="margin-top: 4px; padding: 8px; background: var(--bg-code); border-radius: 4px; border-left: 2px solid #a6e3a1;">
              <div style="display: flex; align-items: center; gap: 6px; margin-bottom: 4px;">
                <span style="color: var(--text-muted); font-size: 11px;">id: {{ tr.id }}</span>
                <n-tag v-if="tr.isError" size="tiny" type="error">✗ 失败</n-tag>
                <n-tag v-else size="tiny" type="success">✓ 成功</n-tag>
              </div>
              <template v-if="tryParseJson(tr.content)">
                <div v-for="(val, key) in tryParseJson(tr.content)" :key="key" style="display: flex; align-items: baseline; gap: 6px; padding: 3px 6px; font-size: 12px;">
                  <code style="color: #89dceb; white-space: nowrap;">{{ key }}</code>
                  <span style="color: var(--text-muted);">=</span>
                  <span style="color: #a6e3a1; word-break: break-all;">{{ typeof val === 'string' ? val : JSON.stringify(val) }}</span>
                </div>
              </template>
              <div v-else style="color: var(--text-primary); font-size: 12px; line-height: 1.5; white-space: pre-wrap; word-break: break-word;">{{ typeof tr.content === 'string' ? tr.content : JSON.stringify(tr.content) }}</div>
          </div>
          </div>
          </div>
        </details>
      </div>

      <div v-if="tools && tools.length > 0" style="margin-top: 16px;">
        <n-collapse>
          <n-collapse-item :title="'🔧 工具定义 (' + tools.length + ')'" name="tools">
            <div v-for="(tool, i) in tools" :key="i" style="margin-bottom: 12px;">
              <div style="background: var(--bg-message); border-radius: 8px; padding: 12px; border-left: 3px solid #89dceb;">
                <div style="display: flex; align-items: center; gap: 8px; margin-bottom: 6px;">
                  <n-tag size="small" type="info">{{ tool.type || 'function' }}</n-tag>
                  <code style="color: #89dceb; font-size: 14px; font-weight: bold;">{{ tool.name }}</code>
                </div>
                <div v-if="tool.description" style="color: var(--text-secondary); font-size: 12px; margin-bottom: 8px; line-height: 1.4;">{{ tool.description }}</div>
                <div v-if="tool.inputSchema" style="margin-top: 6px;">
                  <template v-if="tool.inputSchema.properties">
                    <div v-for="(param, pName) in tool.inputSchema.properties" :key="pName" style="display: flex; align-items: baseline; gap: 6px; padding: 4px 8px; margin: 2px 0; background: var(--bg-code); border-radius: 4px; font-size: 12px;">
                      <code style="color: #89dceb; font-weight: bold; white-space: nowrap;">{{ pName }}</code>
                      <span v-if="param.type" style="color: var(--text-muted); font-size: 11px;">{{ param.type }}</span>
                      <span v-if="tool.inputSchema.required && tool.inputSchema.required.includes(pName)" style="color: #f38ba8; font-size: 11px;">*必填</span>
                      <span v-if="param.description" style="color: var(--text-secondary); margin-left: 4px;">— {{ param.description }}</span>
                    </div>
                  </template>
                  <pre v-else style="color: var(--text-primary); font-size: 11px; margin-top: 4px; white-space: pre-wrap;">{{ JSON.stringify(tool.inputSchema, null, 2) }}</pre>
                </div>
              </div>
            </div>
            <div v-if="toolChoice" style="color: var(--text-muted); font-size: 11px; margin-top: 8px;">tool_choice: <code style="color: #a6e3a1;">{{ JSON.stringify(toolChoice) }}</code></div>
          </n-collapse-item>
        </n-collapse>
      </div>

      <div v-if="!parsed && data">
        <pre style="background: var(--bg-code); border-radius: 4px; padding: 12px 16px; font-size: 12px; line-height: 1.6; overflow-x: auto; color: var(--text-primary); white-space: pre-wrap;">{{ data }}</pre>
      </div>
    </template>
  </div>
</template>
