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

// Detect if content is HTML and format it with proper indentation via DOMParser
function formatHtml(text) {
  if (!text) return null
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
            <div style="color: #89dceb; font-size: 12px; margin-bottom: 4px;">🔧 工具调用:</div>
            <div v-for="(tc, j) in msg.tool_calls" :key="j" style="margin-top: 4px;">
              <div v-if="tc.function" style="padding: 8px; background: var(--bg-code); border-radius: 4px; border-left: 2px solid #89dceb;">
                <code style="color: #89dceb; font-size: 13px; font-weight: bold;">{{ tc.function.name }}</code>
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
        </div>
      </div>

      <div v-if="tools && tools.length > 0" style="margin-top: 16px;">
        <n-collapse>
          <n-collapse-item :title="'🔧 工具定义 (' + tools.length + ')'" name="tools">
            <div v-for="(tool, i) in tools" :key="i" style="margin-bottom: 12px;">
              <div v-if="tool.function" style="background: var(--bg-message); border-radius: 8px; padding: 12px; border-left: 3px solid #89dceb;">
                <div style="display: flex; align-items: center; gap: 8px; margin-bottom: 6px;">
                  <n-tag size="small" type="info">{{ tool.type || 'function' }}</n-tag>
                  <code style="color: #89dceb; font-size: 14px; font-weight: bold;">{{ tool.function.name }}</code>
                </div>
                <div v-if="tool.function.description" style="color: var(--text-secondary); font-size: 12px; margin-bottom: 8px; line-height: 1.4;">{{ tool.function.description }}</div>
                <div v-if="tool.function.parameters" style="margin-top: 6px;">
                  <div v-if="tool.function.parameters.properties">
                    <div v-for="(param, pName) in tool.function.parameters.properties" :key="pName" style="display: flex; align-items: baseline; gap: 6px; padding: 4px 8px; margin: 2px 0; background: var(--bg-code); border-radius: 4px; font-size: 12px;">
                      <code style="color: #89dceb; font-weight: bold; white-space: nowrap;">{{ pName }}</code>
                      <span v-if="param.type" style="color: var(--text-muted); font-size: 11px;">{{ param.type }}</span>
                      <span v-if="tool.function.parameters.required && tool.function.parameters.required.includes(pName)" style="color: #f38ba8; font-size: 11px;">*必填</span>
                      <span v-if="param.description" style="color: var(--text-secondary); margin-left: 4px;">— {{ param.description }}</span>
                    </div>
                  </div>
                  <pre v-else style="color: var(--text-primary); font-size: 11px; margin-top: 4px; white-space: pre-wrap;">{{ JSON.stringify(tool.function.parameters, null, 2) }}</pre>
                </div>
              </div>
            </div>
            <div v-if="toolChoice" style="color: var(--text-muted); font-size: 11px; margin-top: 8px;">tool_choice: <code style="color: #a6e3a1;">{{ JSON.stringify(toolChoice) }}</code></div>
          </n-collapse-item>
        </n-collapse>
      </div>

      <div v-if="!parsed && data" style="color: var(--text-muted); text-align: center; padding: 20px;">无法解析 JSON</div>
    </template>
  </div>
</template>
