<script setup>
import { ref, onMounted, onUnmounted, onErrorCaptured } from 'vue'
import { NInput, NSelect, NButton, NTag, NTable, NPagination, NSwitch, NModal, NTabs, NTabPane } from 'naive-ui'
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
    <h2 style="color: var(--text-primary); margin-bottom: 16px;">📋 请求日志</h2>

    <div style="display: flex; gap: 10px; margin-bottom: 14px; align-items: center; flex-wrap: wrap;">
      <n-input v-model:value="keyword" placeholder="🔍 搜索关键词..." clearable style="width: 180px;" @keyup.enter="search" />
      <n-select v-model:value="routeName" :options="routeOptions" style="width: 130px;" @update:value="search" />
      <n-select v-model:value="statusCode" :options="statusOptions" style="width: 110px;" @update:value="search" />
      <n-select v-model:value="protocol" :options="protocolOptions" style="width: 110px;" @update:value="search" />
      <n-button type="primary" size="small" @click="search">搜索</n-button>
      <span style="margin-left: auto; display: flex; align-items: center; gap: 6px; color: var(--text-secondary); font-size: 13px;">
        <span>🔄</span>
        <n-switch v-model:value="autoRefresh" @update:value="toggleAuto" />
        <n-select v-model:value="refreshInterval" :options="[{label:'3 秒',value:3},{label:'5 秒',value:5},{label:'10 秒',value:10}]" style="width: 80px;" />
      </span>
    </div>

    <div style="background: var(--bg-card); border-radius: 8px; overflow: hidden;">
      <div v-if="selectedIds.length > 0" style="display: flex; align-items: center; gap: 10px; margin-bottom: 8px; padding: 6px 12px; background: var(--bg-card); border-radius: 6px;">
        <span style="color: var(--text-secondary); font-size: 12px;">已选 {{ selectedIds.length }} 条</span>
        <n-button size="tiny" type="error" @click="deleteSelected">删除选中</n-button>
      </div>
      <n-table :single-line="false" style="background: transparent;">
        <thead>
          <tr style="background: var(--border-color);">
            <th style="width: 36px;">
              <input type="checkbox" :checked="selectedIds.length === logs.length && logs.length > 0" @click.stop="selectAll()" style="cursor: pointer;" />
            </th>
            <th style="color: var(--text-secondary); width: 55px;">协议</th>
            <th style="color: var(--text-secondary); width: 50px;">方法</th>
            <th style="color: var(--text-secondary); width: 60px;">状态</th>
            <th style="color: var(--text-secondary); width: 75px;">服务商</th>
            <th style="color: var(--text-secondary); width: 60px;">耗时</th>
            <th style="color: var(--text-secondary); max-width: 350px;">请求 / 目标地址</th>
            <th style="color: var(--text-secondary); width: 140px;">时间</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="log in logs" :key="log.id" style="cursor: pointer;">
            <td @click.stop>
              <input type="checkbox" :checked="isSelected(log.id)" @click.stop="toggleSelect(log.id)" style="cursor: pointer;" />
            </td>
            <td @click="openDetail(log.id)"><n-tag :type="log.protocol === 'SSE' ? 'warning' : 'info'" size="tiny">{{ log.protocol }}</n-tag></td>
            <td @click="openDetail(log.id)" style="color: #89b4fa; font-size: 12px;">{{ log.method }}</td>
            <td @click="openDetail(log.id)"><n-tag :type="statusTagType(log.statusCode)" size="tiny">{{ log.statusCode }}</n-tag></td>
            <td @click="openDetail(log.id)" style="color: var(--text-primary); font-size: 12px;">{{ log.routeName }}</td>
            <td @click="openDetail(log.id)" style="color: var(--text-primary); font-size: 12px;">{{ formatLatency(log.latencyMs) }}</td>
            <td @click="openDetail(log.id)" style="max-width: 350px; overflow: hidden;">
              <div v-if="log.targetUrl" style="font-size: 11px; line-height: 1.3;">
                <div style="color: var(--text-secondary); white-space: nowrap; overflow: hidden; text-overflow: ellipsis;" :title="log.requestUrl">→ {{ log.requestUrl || log.path }}</div>
                <div style="color: #a6e3a1; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;" :title="log.targetUrl">↳ {{ log.targetUrl }}</div>
              </div>
              <code v-else style="color: var(--text-muted); font-size: 11px;">{{ log.path }}</code>
            </td>
            <td @click="openDetail(log.id)" style="color: var(--text-secondary); font-size: 12px; white-space: nowrap;">{{ formatTime(log.createdAt) }}</td>
          </tr>
          <tr v-if="logs.length === 0">
            <td colspan="8" style="text-align: center; color: var(--text-muted); padding: 40px;">暂无日志记录</td>
          </tr>
        </tbody>
      </n-table>
    </div>

    <div style="display: flex; justify-content: space-between; align-items: center; margin-top: 12px;">
      <div style="color: var(--text-secondary); font-size: 12px;">
        每页 <n-select v-model:value="pageSize" :options="[{label:'20',value:20},{label:'50',value:50},{label:'100',value:100}]" style="width: 70px; display: inline-block;" @update:value="load" /> 条
      </div>
      <n-pagination
        v-model:page="page"
        :page-count="Math.ceil(total / pageSize) || 1"
        @update:page="load"
        simple
      />
    </div>

    <n-modal v-model:show="showDetail" preset="card" title="请求详情" style="max-width: 860px;" :segmented="{ content: true }">
      <template v-if="detailLog">
        <n-tabs type="line">
          <n-tab-pane name="req" tab="📤 请求">
            <llm-message-viewer :data="detailLog.reqBody" mode="request" :showRaw="showRaw" @update:showRaw="showRaw = $event" />
          </n-tab-pane>
          <n-tab-pane name="resp" tab="📥 响应">
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
      </template>
    </n-modal>
  </div>
</template>
