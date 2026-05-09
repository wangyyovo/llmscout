<script setup>
import { ref, onMounted } from 'vue'
import { NButton, NCard, NSpace, NTag, NModal, NInput, NSelect, useMessage } from 'naive-ui'
import { ListRoutes, AddRoute, UpdateRoute, DeleteRoute } from '../../wailsjs/go/main/App'

const message = useMessage()
const rules = ref([])
const showModal = ref(false)
const editingId = ref(null)

const form = ref({ name: '', type: 'prefix', path: '', targetUrl: '' })
const typeOptions = [
  { label: 'prefix — 路径前缀剥离', value: 'prefix' },
  { label: 'exact — 精确路径映射', value: 'exact' },
]

async function load() {
  rules.value = await ListRoutes()
}

function openAdd() {
  editingId.value = null
  form.value = { name: '', type: 'prefix', path: '', targetUrl: '' }
  showModal.value = true
}

function openEdit(rule) {
  editingId.value = rule.id
  form.value = { name: rule.name, type: rule.type, path: rule.path, targetUrl: rule.targetUrl }
  showModal.value = true
}

async function save() {
  if (!form.value.name || !form.value.path || !form.value.targetUrl) {
    message.error('请填完所有字段')
    return
  }
  if (editingId.value) {
    await UpdateRoute(editingId.value, form.value.name, form.value.type, form.value.path, form.value.targetUrl)
    message.success('路由已更新')
  } else {
    await AddRoute(form.value.name, form.value.type, form.value.path, form.value.targetUrl)
    message.success('路由已添加')
  }
  showModal.value = false
  load()
}

async function remove(id) {
  await DeleteRoute(id)
  message.success('路由已删除')
  load()
}

onMounted(load)
</script>

<template>
  <div>
    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px;">
      <h2 style="color: var(--text-primary);">🔀 路由规则</h2>
      <n-button type="primary" @click="openAdd">+ 添加路由</n-button>
    </div>

    <n-card v-for="rule in rules" :key="rule.id"
      style="background: var(--bg-card); border: none; margin-bottom: 8px; border-left: 3px solid var(--accent); transition: all 0.15s ease;">
      <div style="display: flex; align-items: center; gap: 12px;">
        <n-tag :type="rule.type === 'prefix' ? 'info' : 'warning'" size="small">
          {{ rule.type }}
        </n-tag>
        <code style="color: var(--text-primary);">{{ rule.path }}</code>
        <span style="color: var(--text-muted);">→</span>
        <code style="color: #a6e3a1;">{{ rule.targetUrl }}</code>
        <span style="color: var(--text-muted); font-size: 12px; margin-left: 4px;">({{ rule.name }})</span>
        <span style="margin-left: auto;">
          <n-button quaternary size="small" @click="openEdit(rule)" style="color: var(--text-muted);">编辑</n-button>
          <n-button quaternary size="small" @click="remove(rule.id)" style="color: #f38ba8;">删除</n-button>
        </span>
      </div>
    </n-card>

    <div v-if="rules.length === 0" style="color: var(--text-muted); text-align: center; padding: 40px;">
      暂无路由规则，点击上方按钮添加
    </div>

    <n-modal v-model:show="showModal" preset="card" title="路由规则" style="max-width: 500px;">
      <n-space vertical size="large">
        <n-input v-model:value="form.name" placeholder="名称（如 openai）" />
        <n-select v-model:value="form.type" :options="typeOptions" />
        <n-input v-model:value="form.path" placeholder="代理路径（如 /openai）" />
        <n-input v-model:value="form.targetUrl" placeholder="目标域名（如 api.openai.com）" />
        <n-button type="primary" @click="save">保存</n-button>
      </n-space>
    </n-modal>
  </div>
</template>
