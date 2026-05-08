<script setup>
import { ref, onMounted, onUnmounted, onErrorCaptured } from 'vue'
import { NInput, NSelect, NButton, NTag, NTable, NPagination, NSwitch, NModal, NTabs, NTabPane } from 'naive-ui'
import { QueryLogs, GetLog, GetLogRouteNames } from '../../wailsjs/go/main/App'
import JsonViewer from '../components/JsonViewer.vue'
import LlmMessageViewer from '../components/LlmMessageViewer.vue'

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
  return `${pad(d.getMonth()+1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
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
      <n-table :single-line="false" style="background: transparent;">
        <thead>
          <tr style="background: var(--border-color);">
            <th style="color: var(--text-secondary); width: 55px;">协议</th>
            <th style="color: var(--text-secondary); width: 55px;">方法</th>
            <th style="color: var(--text-secondary); width: 65px;">状态</th>
            <th style="color: var(--text-secondary); width: 90px;">服务商</th>
            <th style="color: var(--text-secondary);">路径</th>
            <th style="color: var(--text-secondary); width: 65px;">耗时</th>
            <th style="color: var(--text-secondary); width: 140px;">时间</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="log in logs" :key="log.id" @click="openDetail(log.id)" style="cursor: pointer;">
            <td><n-tag :type="log.protocol === 'SSE' ? 'warning' : 'info'" size="tiny">{{ log.protocol }}</n-tag></td>
            <td style="color: #89b4fa;">{{ log.method }}</td>
            <td><n-tag :type="statusTagType(log.statusCode)" size="tiny">{{ log.statusCode }}</n-tag></td>
            <td style="color: var(--text-primary);">{{ log.routeName }}</td>
            <td><code style="color: #a6e3a1; font-size: 12px;">{{ log.path }}</code></td>
            <td style="color: var(--text-primary);">{{ log.latencyMs }}ms</td>
            <td style="color: var(--text-secondary); font-size: 12px;">{{ formatTime(log.createdAt) }}</td>
          </tr>
          <tr v-if="logs.length === 0">
            <td colspan="7" style="text-align: center; color: var(--text-muted); padding: 40px;">暂无日志记录</td>
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
            <json-viewer :data="detailLog.reqHeaders" />
          </n-tab-pane>
          <n-tab-pane name="respHeaders" tab="响应头">
            <json-viewer :data="detailLog.respHeaders" />
          </n-tab-pane>
        </n-tabs>
      </template>
    </n-modal>
  </div>
</template>
