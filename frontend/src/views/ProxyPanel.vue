<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { NInput, NButton, NCard, NStatistic, NSpace, NTag } from 'naive-ui'
import { StartProxy, StopProxy, ProxyStatus, ListRoutes, QueryLogs, GetSetting } from '../../wailsjs/go/main/App'

const port = ref(8899)
const running = ref(false)
const uptime = ref('')
const routeCount = ref(0)
const todayCount = ref(0)

let statusTimer = null

async function refreshStatus() {
  const s = await ProxyStatus()
  running.value = s.running
  uptime.value = s.uptime || ''
  port.value = s.port
  const routes = await ListRoutes()
  routeCount.value = routes.length
  const result = await QueryLogs({ keyword: '', routeName: '', statusCode: 0, protocol: '', startTime: '', endTime: '', page: 1, pageSize: 1 })
  todayCount.value = Number(result.total) || 0
}

async function toggleProxy() {
  if (running.value) {
    await StopProxy()
  } else {
    await StartProxy()
  }
  refreshStatus()
}

onMounted(async () => {
  const savedPort = await GetSetting('port', '8899')
  port.value = parseInt(savedPort) || 8899
  refreshStatus()
  statusTimer = setInterval(refreshStatus, 3000)
})

onUnmounted(() => {
  if (statusTimer) clearInterval(statusTimer)
})
</script>

<template>
  <div>
    <h2 style="color: var(--text-primary); margin-bottom: 20px;">📡 代理控制</h2>
    <n-card style="background: var(--bg-card); border: none; max-width: 500px;">
      <n-space vertical size="large">
        <div style="display: flex; align-items: center; gap: 12px;">
          <span style="color: var(--text-secondary);">端口</span>
          <n-input
            v-model:value="port"
            :disabled="running"
            style="width: 120px;"
            type="number"
          />
          <span style="color: var(--text-muted); font-size: 13px;">localhost:{{ port }}</span>
        </div>
        <div style="display: flex; align-items: center; gap: 12px;">
          <span style="color: var(--text-secondary);">状态</span>
          <n-tag :type="running ? 'success' : 'default'" size="small">
            {{ running ? '● 运行中' : '已停止' }}
          </n-tag>
          <n-button
            :type="running ? 'error' : 'primary'"
            @click="toggleProxy"
          >
            {{ running ? '停止代理' : '启动代理' }}
          </n-button>
        </div>
      </n-space>
    </n-card>

    <div style="display: flex; gap: 20px; margin-top: 20px;">
      <n-card style="background: var(--bg-card); border: none; flex: 1;" size="small">
        <n-statistic label="已配置路由" :value="routeCount" />
      </n-card>
      <n-card style="background: var(--bg-card); border: none; flex: 1;" size="small">
        <n-statistic label="今日请求" :value="todayCount" />
      </n-card>
      <n-card style="background: var(--bg-card); border: none; flex: 1;" size="small">
        <n-statistic label="运行时长" :value="uptime || '-'" />
      </n-card>
    </div>
  </div>
</template>
