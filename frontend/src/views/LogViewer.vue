<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { NInput, NSelect, NButton, NTag, NTable, NPagination, NSwitch, NModal, NTabs, NTabPane, NIcon } from 'naive-ui'
import { SearchOutline, RefreshOutline, TrashOutline } from '@vicons/ionicons5'
import { QueryLogs, GetLog, GetLogRouteNames, DeleteLogs } from '../../wailsjs/go/main/App'
import JsonViewer from '../components/JsonViewer.vue'
import LlmMessageViewer from '../components/LlmMessageViewer.vue'
import HeadersViewer from '../components/HeadersViewer.vue'

const logs = ref([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const autoRefresh = ref(false)
const refreshInterval = ref(3)
const showRaw = ref(false)
const detailLog = ref(null)
const showDetail = ref(false)

const keyword = ref('')
const routeName = ref('')
const statusCode = ref(null)
const protocol = ref('')
const routeOptions = ref([])
const selectedIds = ref([])

function toggleSelect(id) {
  const idx = selectedIds.value.indexOf(id)
  if (idx >= 0) {
    selectedIds.value.splice(idx, 1)
  } else {
    selectedIds.value.push(id)
  }
}

function selectAll() {
  if (selectedIds.value.length === logs.value.length) {
    selectedIds.value = []
  } else {
    selectedIds.value = logs.value.map(l => l.id)
  }
}

function isSelected(id) {
  return selectedIds.value.includes(id)
}

async function deleteSelected() {
  if (selectedIds.value.length === 0) return
  const ids = selectedIds.value.map(Number)
  await DeleteLogs(ids)
  selectedIds.value = []
  load()
}

let timer = null

const protocolOptions = [
  { label: '全部协议', value: '' },
  { label: 'REST', value: 'REST' },
  { label: 'SSE', value: 'SSE' },
]

const statusOptions = [
  { label: '全部状态', value: null },
  { label: '2xx', value: 2 },
  { label: '3xx', value: 3 },
  { label: '4xx', value: 4 },
  { label: '5xx', value: 5 },
]

async function load() {
  const filter = {
    keyword: keyword.value,
    routeName: routeName.value,
    statusCode: statusCode.value ? statusCode.value * 100 : 0,
    protocol: protocol.value,
    startTime: '',
    endTime: '',
    page: page.value,
    pageSize: pageSize.value,
  }
  try {
    const result = await QueryLogs(filter)
    logs.value = result.list || []
    total.value = result.total || 0
  } catch (e) {
    // silent
  }
}

async function loadRouteNames() {
  try {
    const names = await GetLogRouteNames()
    routeOptions.value = [{ label: '全部服务商', value: '' }, ...names.map(n => ({ label: n, value: n }))]
  } catch {
    routeOptions.value = [{ label: '全部服务商', value: '' }]
  }
}

function search() {
  page.value = 1
  load()
}

function toggleAuto(v) {
  autoRefresh.value = v
  if (v) {
    timer = setInterval(load, refreshInterval.value * 1000)
  } else if (timer) {
    clearInterval(timer)
    timer = null
  }
}

async function openDetail(id) {
  try {
    detailLog.value = await GetLog(id)
    showDetail.value = true
  } catch { /* ignore */ }
}

function formatTime(t) {
  if (!t) return ''
  const d = new Date(t)
  const pad = n => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth()+1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

function formatLatency(ms) {
  if (!ms && ms !== 0) return '-'
  if (ms < 1000) return ms + ' ms'
  if (ms < 10000) return (ms / 1000).toFixed(1) + ' s'
  return Math.round(ms / 1000) + ' s'
}

function statusTagType(code) {
  if (code >= 200 && code < 300) return 'success'
  if (code >= 300 && code < 400) return 'info'
  if (code >= 400 && code < 500) return 'warning'
  if (code >= 500) return 'error'
  return 'default'
}

onMounted(() => {
  load()
  loadRouteNames()
})
onUnmounted(() => { if (timer) clearInterval(timer) })
</script>

<template>
  <div>
    <div class="page-header">
      <h2 class="page-title">请求日志</h2>
      <div class="header-actions">
        <n-button quaternary size="small" @click="load">
          <template #icon><n-icon size="17"><RefreshOutline /></n-icon></template>
        </n-button>
      </div>
    </div>

    <div class="filter-bar">
      <div class="filter-group">
        <n-input v-model:value="keyword" placeholder="搜索关键词..." clearable class="filter-input" @keyup.enter="search">
          <template #prefix><n-icon size="14" color="var(--text-muted)"><SearchOutline /></n-icon></template>
        </n-input>
        <n-button type="primary" size="small" @click="search">
          <template #icon><n-icon size="14"><SearchOutline /></n-icon></template>
          搜索
        </n-button>
      </div>
      <div class="filter-group">
        <n-select v-model:value="routeName" :options="routeOptions" class="filter-select" @update:value="search" />
        <n-select v-model:value="statusCode" :options="statusOptions" class="filter-select-sm" @update:value="search" />
        <n-select v-model:value="protocol" :options="protocolOptions" class="filter-select-sm" @update:value="search" />
      </div>
      <div class="filter-group auto-group">
        <span class="auto-label">自动刷新</span>
        <n-switch v-model:value="autoRefresh" @update:value="toggleAuto" size="small" />
        <n-select v-model:value="refreshInterval" :options="[{label:'3s',value:3},{label:'5s',value:5},{label:'10s',value:10}]" class="interval-select" size="small" />
      </div>
    </div>

    <div class="table-container">
      <div v-if="selectedIds.length > 0" class="batch-bar">
        <span class="batch-info">已选 {{ selectedIds.length }} 条</span>
        <n-button size="tiny" type="error" @click="deleteSelected">
          <template #icon><n-icon size="12"><TrashOutline /></n-icon></template>
          删除选中
        </n-button>
      </div>

      <n-table :single-line="false" class="log-table">
        <thead>
          <tr class="table-head-row">
            <th class="col-cb">
              <input type="checkbox" :checked="selectedIds.length === logs.length && logs.length > 0" @click.stop="selectAll()" class="table-checkbox" />
            </th>
            <th class="col-id">#</th>
            <th class="col-proto">协议</th>
            <th class="col-method">方法</th>
            <th class="col-status">状态</th>
            <th class="col-route">服务商</th>
            <th class="col-latency">耗时</th>
            <th class="col-url">请求地址</th>
            <th class="col-target">转发到</th>
            <th class="col-time">时间</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="log in logs" :key="log.id" class="table-row" @click="openDetail(log.id)">
            <td class="col-cb" @click.stop>
              <input type="checkbox" :checked="isSelected(log.id)" @click.stop="toggleSelect(log.id)" class="table-checkbox" />
            </td>
            <td class="col-id cell-muted cell-mono">{{ log.id }}</td>
            <td class="col-proto"><n-tag :type="log.protocol === 'SSE' ? 'warning' : 'info'" size="tiny" round>{{ log.protocol }}</n-tag></td>
            <td class="col-method cell-accent cell-mono">{{ log.method }}</td>
            <td class="col-status"><n-tag :type="statusTagType(log.statusCode)" size="tiny" round>{{ log.statusCode }}</n-tag></td>
            <td class="col-route cell-primary">{{ log.routeName }}</td>
            <td class="col-latency cell-primary cell-mono">{{ formatLatency(log.latencyMs) }}</td>
            <td class="col-url cell-ellipsis" :title="log.requestUrl">
              <code class="url-code">{{ log.requestUrl || log.path }}</code>
            </td>
            <td class="col-target cell-ellipsis" :title="log.targetUrl">
              <code class="target-code">{{ log.targetUrl || '-' }}</code>
            </td>
            <td class="col-time cell-secondary cell-mono">{{ formatTime(log.createdAt) }}</td>
          </tr>
          <tr v-if="logs.length === 0">
            <td colspan="10" class="empty-row">暂无日志记录</td>
          </tr>
        </tbody>
      </n-table>
    </div>

    <div class="table-footer">
      <div class="page-size">
        <span class="page-size-label">每页</span>
        <n-select
          v-model:value="pageSize"
          :options="[{label:'20',value:20},{label:'50',value:50},{label:'100',value:100}]"
          class="page-size-select"
          size="small"
          @update:value="load"
        />
        <span class="page-size-label">条</span>
        <span class="page-size-label" style="margin-left: 8px;">共 {{ total }} 条</span>
      </div>
      <n-pagination
        v-model:page="page"
        :page-count="Math.ceil(total / pageSize) || 1"
        @update:page="load"
        simple
      />
    </div>

    <n-modal
      v-model:show="showDetail"
      preset="card"
      title="请求详情"
      class="detail-modal"
      :bordered="false"
    >
      <template v-if="detailLog">
        <div class="detail-body">
          <n-tabs type="line" animated>
            <n-tab-pane name="req" tab="请求">
              <llm-message-viewer :data="detailLog.reqBody" mode="request" :showRaw="showRaw" @update:showRaw="showRaw = $event" />
            </n-tab-pane>
            <n-tab-pane name="resp" tab="响应">
              <llm-message-viewer :data="detailLog.respBody" mode="response" :showRaw="showRaw" @update:showRaw="showRaw = $event" />
            </n-tab-pane>
            <n-tab-pane name="reqHeaders" tab="请求头">
              <headers-viewer v-if="!showRaw" :data="detailLog.reqHeaders" />
              <json-viewer v-else :data="detailLog.reqHeaders" />
            </n-tab-pane>
            <n-tab-pane name="respHeaders" tab="响应头">
              <headers-viewer v-if="!showRaw" :data="detailLog.respHeaders" />
              <json-viewer v-else :data="detailLog.respHeaders" />
            </n-tab-pane>
          </n-tabs>
        </div>
      </template>
    </n-modal>
  </div>
</template>

<style scoped>
/* Header */
.header-actions {
  display: flex;
  align-items: center;
  gap: 6px;
}

/* Filter bar */
.filter-bar {
  display: flex;
  gap: 8px;
  margin-bottom: 14px;
  align-items: center;
  flex-wrap: wrap;
}
.filter-group {
  display: flex;
  gap: 6px;
  align-items: center;
}
.filter-input { width: 190px; }
.filter-select { width: 140px; }
.filter-select-sm { width: 100px; }
.auto-group {
  margin-left: auto;
  gap: 6px;
  padding: 0 2px;
}
.auto-label {
  color: var(--text-muted);
  font-size: 12px;
  white-space: nowrap;
  user-select: none;
}
.interval-select { width: 64px; }

/* Table container */
.table-container {
  background: var(--bg-card);
  border-radius: var(--radius);
  overflow: auto hidden;
  box-shadow: var(--shadow-sm);
  border: 1px solid var(--border-color);
  position: relative;
}

/* Batch bar */
.batch-bar {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 7px 14px;
  background: rgba(124,140,248,0.08);
  border-bottom: 1px solid var(--border-color);
}
.batch-info {
  color: var(--accent);
  font-size: 12px;
  font-weight: 600;
}

/* Table */
.log-table { background: transparent; }

.table-head-row { background: var(--bg-hover) !important; }
.table-head-row th {
  padding: 8px 14px;
  font-size: 11px;
  font-weight: 600;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  border-bottom: 2px solid var(--border-color);
  white-space: nowrap;
}

.table-row {
  cursor: pointer;
  transition: background 0.15s ease;
}
.table-row:nth-child(even) td {
  background: rgba(124,140,248,0.015);
}
.table-row td {
  padding: 9px 14px;
  font-size: 13px;
  border-bottom: 1px solid var(--border-color);
}
.table-row:last-child td { border-bottom: none; }
.table-row:hover td { background: var(--bg-hover) !important; }

/* Cell widths */
.col-cb { width: 38px; text-align: center; }
.col-id { width: 48px; }
.col-proto { width: 56px; }
.col-method { width: 50px; }
.col-status { width: 52px; }
.col-route { width: 100px; }
.col-latency { width: 64px; }
.col-url { min-width: 140px; }
.col-target { min-width: 140px; }
.col-time { width: 148px; white-space: nowrap; }

/* Frozen columns */
.col-cb,
.col-time {
  position: sticky;
  z-index: 2;
}
.col-cb { left: 0; }
.col-time { right: 0; }

.table-head-row .col-cb,
.table-head-row .col-time {
  background: var(--bg-hover);
  z-index: 3;
}
.table-row .col-cb,
.table-row .col-time {
  background: var(--bg-card);
}
.table-row:hover .col-cb,
.table-row:hover .col-time {
  background: var(--bg-hover);
}

/* Shadow on frozen columns for depth */
.table-row .col-cb {
  box-shadow: 2px 0 4px rgba(0,0,0,0.08);
  clip-path: inset(0 -6px 0 0);
}
.table-row .col-time {
  box-shadow: -2px 0 4px rgba(0,0,0,0.08);
  clip-path: inset(0 0 0 -6px);
}

/* Ensure td/th support pseudo-elements */
.table-head-row th,
.table-row td {
  position: relative;
}

/* Cell text styles */
.cell-muted { color: var(--text-muted); }
.cell-mono { font-family: 'SF Mono', 'Fira Code', monospace; font-size: 12px; }
.cell-accent { color: var(--accent); font-weight: 600; font-size: 12px; }
.cell-primary { color: var(--text-primary); font-size: 13px; }
.cell-secondary { color: var(--text-secondary); font-size: 12px; }

.cell-ellipsis {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.url-code {
  color: var(--text-secondary);
  font-size: 12px;
  font-family: 'SF Mono', 'Fira Code', monospace;
}
.target-code {
  color: var(--accent-success);
  font-size: 12px;
  font-family: 'SF Mono', 'Fira Code', monospace;
}

.table-checkbox {
  cursor: pointer;
  accent-color: var(--accent);
}

.empty-row {
  text-align: center;
  color: var(--text-muted);
  padding: 40px 14px !important;
  font-size: 13px;
}

/* Table footer */
.table-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 14px;
}
.page-size {
  display: flex;
  align-items: center;
  gap: 5px;
}
.page-size-label {
  color: var(--text-muted);
  font-size: 12px;
}
.page-size-select { width: 72px; }

/* Detail modal */
.detail-modal { max-width: 1000px; }
.detail-body {
  max-height: calc(90vh - 80px);
  overflow-y: auto;
}

/* Responsive */
@media (max-width: 900px) {
  .detail-modal { max-width: 95vw; }
  .filter-bar { gap: 6px; }
  .filter-input { width: 150px; }
  .filter-select { width: 120px; }
  .filter-select-sm { width: 90px; }
  .auto-group { margin-left: 0; }
  .table-container { overflow-x: auto; }
  .col-proto, .col-latency { display: none; }
  .col-id { width: 40px; }
  .col-url, .col-target { min-width: 120px; }
}
@media (max-width: 600px) {
  .filter-bar {
    flex-direction: column;
    align-items: stretch;
  }
  .filter-group {
    flex-wrap: wrap;
  }
  .filter-input { width: 100%; }
  .filter-select { flex: 1; min-width: 100px; }
  .filter-select-sm { flex: 1; min-width: 80px; }
  .auto-group {
    flex-direction: row;
    justify-content: space-between;
  }
  .col-route, .col-target { display: none; }
  .col-url { min-width: 100px; }
  .col-time { width: 130px; }
  .table-footer {
    flex-direction: column;
    gap: 10px;
    align-items: stretch;
  }
  .page-size { justify-content: center; }
  :deep(.table-footer .n-pagination) { justify-content: center; }
}
</style>

<style>
/* Non-scoped: modal is teleported outside component, scoped styles can't reach it.
   Use a specific class prefix to avoid polluting global scope. */
.detail-modal .n-card > .n-card-header {
  position: sticky;
  top: 0;
  z-index: 10;
  background: var(--bg-card);
  border-bottom: 1px solid var(--border-color);
  padding-bottom: 14px;
}
</style>
