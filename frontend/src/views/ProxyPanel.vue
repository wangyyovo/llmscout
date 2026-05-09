<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { NInput, NButton, NCard, NStatistic, NSpace, NTag, NIcon } from 'naive-ui'
import { PlayOutline, StopOutline, ServerOutline, GitBranchOutline, DocumentTextOutline, TimeOutline } from '@vicons/ionicons5'
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
    <div class="page-header">
      <h2 class="page-title">代理控制</h2>
      <n-tag :type="running ? 'success' : 'default'" size="medium" round>
        <template #icon>
          <span class="status-dot" :class="running ? 'dot-live' : 'dot-idle'" />
        </template>
        {{ running ? '运行中' : '已停止' }}
      </n-tag>
    </div>

    <div class="proxy-grid">
      <n-card class="control-card" :bordered="false">
        <div class="control-header">
          <n-icon size="20" color="var(--accent)"><ServerOutline /></n-icon>
          <span class="control-title">代理服务</span>
        </div>

        <div class="control-body">
          <div class="port-row">
            <div class="port-label">监听端口</div>
            <n-input
              v-model:value="port"
              :disabled="running"
              class="port-input"
              type="number"
              placeholder="8899"
            />
            <code class="port-addr">localhost:{{ port }}</code>
          </div>

          <n-button
            :type="running ? 'error' : 'primary'"
            size="large"
            @click="toggleProxy"
            class="control-btn"
            :secondary="running"
          >
            <template #icon>
              <n-icon :size="18"><component :is="running ? StopOutline : PlayOutline" /></n-icon>
            </template>
            {{ running ? '停止代理' : '启动代理' }}
          </n-button>
        </div>
      </n-card>

      <div class="stats-row">
        <n-card class="stat-card stat-routes" :bordered="false" size="small">
          <div class="stat-icon routes-icon">
            <n-icon size="18" color="var(--accent)"><GitBranchOutline /></n-icon>
          </div>
          <n-statistic label="已配置路由" :value="routeCount" />
        </n-card>

        <n-card class="stat-card stat-today" :bordered="false" size="small">
          <div class="stat-icon today-icon">
            <n-icon size="18" color="var(--accent-success)"><DocumentTextOutline /></n-icon>
          </div>
          <n-statistic label="今日请求" :value="todayCount" />
        </n-card>

        <n-card class="stat-card stat-uptime" :bordered="false" size="small">
          <div class="stat-icon uptime-icon">
            <n-icon size="18" color="var(--accent-warning)"><TimeOutline /></n-icon>
          </div>
          <n-statistic label="运行时长" :value="uptime || '-'" />
        </n-card>
      </div>
    </div>
  </div>
</template>

<style scoped>
.proxy-grid {
  display: flex;
  flex-direction: column;
  gap: 20px;
  max-width: 680px;
}

/* Control card */
.control-card {
  background: var(--bg-card) !important;
  border-radius: var(--radius) !important;
  box-shadow: var(--shadow-sm);
}
.control-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding-bottom: 16px;
  margin-bottom: 4px;
  border-bottom: 1px solid var(--border-color);
}
.control-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
}
.control-body {
  display: flex;
  flex-direction: column;
  gap: 20px;
  padding-top: 4px;
}
.port-row {
  display: flex;
  align-items: center;
  gap: 12px;
}
.port-label {
  color: var(--text-secondary);
  font-size: 14px;
  flex-shrink: 0;
}
.port-input {
  width: 100px;
}
.port-addr {
  color: var(--text-muted);
  font-size: 13px;
}
.control-btn {
  align-self: flex-start;
  min-width: 140px;
  border-radius: var(--radius-sm);
  font-weight: 600;
}

/* Status dot */
.status-dot {
  display: inline-block;
  width: 7px;
  height: 7px;
  border-radius: 50%;
}
.dot-live {
  background: #5ce6b8;
  box-shadow: 0 0 6px rgba(92,230,184,0.5);
}
.dot-idle {
  background: #6b6f85;
}

/* Stats row */
.stats-row {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 14px;
}
.stat-card {
  background: var(--bg-card) !important;
  border-radius: var(--radius) !important;
  box-shadow: var(--shadow-sm);
  padding: 4px 0;
  position: relative;
  overflow: hidden;
}
.stat-card::before {
  content: '';
  position: absolute;
  left: 0;
  top: 12px;
  bottom: 12px;
  width: 3px;
  border-radius: 0 2px 2px 0;
}
.stat-routes::before { background: var(--accent); }
.stat-today::before { background: var(--accent-success); }
.stat-uptime::before { background: var(--accent-warning); }

.stat-icon {
  width: 36px;
  height: 36px;
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 8px;
}
.routes-icon { background: rgba(124,140,248,0.1); }
.today-icon { background: rgba(92,230,184,0.1); }
.uptime-icon { background: rgba(245,169,127,0.1); }

:deep(.stat-card .n-statistic__label) {
  font-size: 12px;
  color: var(--text-muted);
}
:deep(.stat-card .n-statistic__value) {
  font-size: 22px;
  font-weight: 700;
  color: var(--text-primary);
}
</style>
