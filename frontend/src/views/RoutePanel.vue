<script setup>
import { ref, onMounted } from 'vue'
import { NButton, NCard, NSpace, NTag, NModal, NInput, NSelect, NIcon, useMessage } from 'naive-ui'
import { AddOutline, CreateOutline, TrashOutline, GitBranchOutline, ArrowForwardOutline } from '@vicons/ionicons5'
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
    <div class="page-header">
      <h2 class="page-title">路由规则</h2>
      <n-button type="primary" @click="openAdd" :border-radius="6">
        <template #icon><n-icon :size="16"><AddOutline /></n-icon></template>
        添加路由
      </n-button>
    </div>

    <div class="route-list">
      <n-card
        v-for="rule in rules"
        :key="rule.id"
        class="route-card"
        :bordered="false"
      >
        <div class="route-card-body">
          <div class="route-left">
            <n-tag :type="rule.type === 'prefix' ? 'info' : 'warning'" size="small" round>
              {{ rule.type }}
            </n-tag>
            <code class="route-path">{{ rule.path }}</code>
            <n-icon size="16" color="var(--text-muted)" class="route-arrow"><ArrowForwardOutline /></n-icon>
            <code class="route-target">{{ rule.targetUrl }}</code>
            <span class="route-name">{{ rule.name }}</span>
          </div>
          <div class="route-actions">
            <n-button quaternary size="small" @click="openEdit(rule)" class="action-btn edit-btn">
              <template #icon><n-icon size="15"><CreateOutline /></n-icon></template>
            </n-button>
            <n-button quaternary size="small" @click="remove(rule.id)" class="action-btn del-btn">
              <template #icon><n-icon size="15" color="var(--accent-error)"><TrashOutline /></n-icon></template>
            </n-button>
          </div>
        </div>
      </n-card>
    </div>

    <div v-if="rules.length === 0" class="empty-state">
      <div class="empty-icon">
        <n-icon size="40" color="var(--text-muted)"><GitBranchOutline /></n-icon>
      </div>
      <div class="empty-title">暂无路由规则</div>
      <div class="empty-desc">点击上方「添加路由」按钮创建第一条规则</div>
    </div>

    <n-modal v-model:show="showModal" preset="card" title="路由规则" class="route-modal" :bordered="false">
      <n-space vertical size="large">
        <div class="form-group">
          <label class="form-label">名称</label>
          <n-input v-model:value="form.name" placeholder="如 openai" />
        </div>
        <div class="form-group">
          <label class="form-label">匹配类型</label>
          <n-select v-model:value="form.type" :options="typeOptions" />
        </div>
        <div class="form-group">
          <label class="form-label">代理路径</label>
          <n-input v-model:value="form.path" placeholder="如 /openai" />
        </div>
        <div class="form-group">
          <label class="form-label">目标域名</label>
          <n-input v-model:value="form.targetUrl" placeholder="如 api.openai.com" />
        </div>
        <n-button type="primary" @click="save" block :border-radius="6">保存</n-button>
      </n-space>
    </n-modal>
  </div>
</template>

<style scoped>
.route-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.route-card {
  background: var(--bg-card) !important;
  border-radius: var(--radius) !important;
  border-left: 3px solid var(--accent) !important;
  transition: all var(--transition);
  box-shadow: var(--shadow-sm);
}
.route-card:hover {
  box-shadow: var(--shadow);
  transform: translateX(2px);
}

.route-card-body {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}
.route-left {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
  min-width: 0;
}
.route-path {
  color: var(--text-primary);
  font-size: 13px;
  font-weight: 500;
}
.route-arrow {
  flex-shrink: 0;
}
.route-target {
  color: var(--accent-success);
  font-size: 13px;
}
.route-name {
  color: var(--text-muted);
  font-size: 11px;
  background: var(--bg-hover);
  padding: 2px 8px;
  border-radius: 10px;
}
.route-actions {
  display: flex;
  gap: 2px;
  flex-shrink: 0;
}
.action-btn {
  opacity: 0.5;
  transition: opacity var(--transition);
}
.action-btn:hover { opacity: 1; }

/* Empty state */
.empty-state {
  text-align: center;
  padding: 48px 20px;
}
.empty-icon {
  margin-bottom: 12px;
  opacity: 0.4;
}
.empty-title {
  color: var(--text-secondary);
  font-size: 15px;
  font-weight: 500;
  margin-bottom: 4px;
}
.empty-desc {
  color: var(--text-muted);
  font-size: 13px;
}

/* Modal form */
.route-modal {
  max-width: 480px;
}
.form-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.form-label {
  font-size: 13px;
  color: var(--text-secondary);
  font-weight: 500;
}
</style>
