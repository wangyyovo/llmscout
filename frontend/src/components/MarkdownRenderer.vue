<script setup>
import { computed } from 'vue'
import markdownit from 'markdown-it'

let md = null
function getMd() {
  if (!md) {
    try {
      md = markdownit({ html: false, linkify: true, breaks: true })
    } catch (e) {
      console.error('markdown-it init failed:', e)
      md = { render: (s) => s || '' }
    }
  }
  return md
}

const props = defineProps({
  content: { type: String, default: '' }
})

const rendered = computed(() => {
  if (!props.content) return ''
  try {
    return getMd().render(props.content)
  } catch {
    return props.content
  }
})
</script>

<template>
  <div class="markdown-body" v-html="rendered"></div>
</template>

<style scoped>
.markdown-body {
  font-size: 13px;
  line-height: 1.7;
  color: var(--text-primary);
  word-break: break-word;
}
.markdown-body :deep(p) {
  margin: 0 0 8px;
}
.markdown-body :deep(p:last-child) {
  margin-bottom: 0;
}
.markdown-body :deep(code) {
  padding: 1px 4px;
  border-radius: 3px;
  background: var(--bg-code);
  font-size: 12px;
}
.markdown-body :deep(pre) {
  background: var(--bg-code);
  border-radius: 4px;
  padding: 10px 12px;
  overflow-x: auto;
  margin: 8px 0;
}
.markdown-body :deep(pre code) {
  background: none;
  padding: 0;
}
.markdown-body :deep(h1),
.markdown-body :deep(h2),
.markdown-body :deep(h3),
.markdown-body :deep(h4) {
  margin: 12px 0 6px;
  color: var(--text-primary);
}
.markdown-body :deep(ul),
.markdown-body :deep(ol) {
  padding-left: 20px;
  margin: 4px 0;
}
.markdown-body :deep(li) {
  margin: 2px 0;
}
.markdown-body :deep(blockquote) {
  border-left: 3px solid var(--border-color);
  padding-left: 10px;
  margin: 8px 0;
  color: var(--text-secondary);
}
.markdown-body :deep(a) {
  color: #89b4fa;
}
.markdown-body :deep(table) {
  border-collapse: collapse;
  width: 100%;
  margin: 8px 0;
  font-size: 12px;
}
.markdown-body :deep(th),
.markdown-body :deep(td) {
  border: 1px solid var(--border-color);
  padding: 6px 10px;
  text-align: left;
}
.markdown-body :deep(th) {
  background: var(--bg-message);
}
</style>
