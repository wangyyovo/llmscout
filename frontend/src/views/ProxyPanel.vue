<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { NInput, NButton, NCard, NStatistic, NTag, NIcon } from 'naive-ui'
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
          <div class="control-header-icon">
            <n-icon size="20" color="var(--accent)"><ServerOutline /></n-icon>
          </div>
          <div>
            <div class="control-title">代理服务</div>
            <div class="control-subtitle">管理本地代理服务器</div>
          </div>
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
              <n-icon :size="20"><component :is="running ? StopOutline : PlayOutline" /></n-icon>
            </template>
            {{ running ? '停止代理' : '启动代理' }}
          </n-button>
        </div>
      </n-card>

      <div class="stats-row">
        <n-card class="stat-card stat-routes" :bordered="false" size="small">
          <div class="stat-content">
            <div class="stat-icon routes-icon">
              <n-icon size="18" color="var(--accent)"><GitBranchOutline /></n-icon>
            </div>
            <div class="stat-body">
              <n-statistic label="已配置路由" :value="routeCount" />
            </div>
          </div>
        </n-card>

        <n-card class="stat-card stat-today" :bordered="false" size="small">
          <div class="stat-content">
            <div class="stat-icon today-icon">
              <n-icon size="18" color="var(--accent-success)"><DocumentTextOutline /></n-icon>
            </div>
            <div class="stat-body">
              <n-statistic label="今日请求" :value="todayCount" />
            </div>
          </div>
        </n-card>

        <n-card class="stat-card stat-uptime" :bordered="false" size="small">
          <div class="stat-content">
            <div class="stat-icon uptime-icon">
              <n-icon size="18" color="var(--accent-warning)"><TimeOutline /></n-icon>
            </div>
            <div class="stat-body">
              <n-statistic label="运行时长" :value="uptime || '-'" />
            </div>
          </div>
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
}

/* Control card */
.control-card {
  background: var(--bg-card) !important;
  border-radius: var(--radius) !important;
  box-shadow: var(--shadow-sm);
  padding: 0 !important;
}
:deep(.control-card .n-card__content) { padding: 24px; }

.control-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding-bottom: 20px;
  margin-bottom: 4px;
  border-bottom: 1px solid var(--border-color);
}
.control-header-icon {
  width: 38px; height: 38px;
  border-radius: var(--radius-sm);
  background: rgba(124,140,248,0.1);
  display: flex; align-items: center; justify-content: center;
  flex-shrink: 0;
}
.control-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
}
.control-subtitle {
  font-size: 12px;
  color: var(--text-muted);
  margin-top: 1px;
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
  font-size: 13px;
  font-weight: 500;
  flex-shrink: 0;
  min-width: 56px;
}
.port-input { width: 110px; }
.port-addr {
  color: var(--text-muted);
  font-size: 12px;
  font-family: 'SF Mono', 'Fira Code', monospace;
}
.control-btn {
  align-self: flex-start;
  min-width: 150px;
  font-weight: 600;
  border-radius: var(--radius-sm);
}

/* Status dot */
.status-dot {
  display: inline-block;
  width: 6px; height: 6px;
  border-radius: 50%;
}
.dot-live {
  background: var(--accent-success);
  box-shadow: 0 0 6px rgba(92,230,184,0.5);
}
.dot-idle { background: var(--text-muted); }

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
  padding: 0 !important;
  overflow: hidden;
  border: none !important;
}
.stat-card::before {
  content: '';
  position: absolute;
  left: 0; top: 14px; bottom: 14px;
  width: 3px;
  border-radius: 0 2px 2px 0;
}
.stat-routes::before { background: var(--accent); }
.stat-today::before { background: var(--accent-success); }
.stat-uptime::before { background: var(--accent-warning); }

.stat-content {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 16px 18px;
}
.stat-icon {
  width: 40px; height: 40px;
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.routes-icon { background: rgba(124,140,248,0.1); }
.today-icon { background: rgba(92,230,184,0.1); }
.uptime-icon { background: rgba(245,169,127,0.1); }

.stat-body {
  min-width: 0;
}
:deep(.stat-card .n-statistic) { white-space: normal; }
:deep(.stat-card .n-statistic__label) {
  font-size: 11px;
  color: var(--text-muted);
  font-weight: 500;
  letter-spacing: 0.2px;
}
:deep(.stat-card .n-statistic__value) {
  font-size: 24px;
  font-weight: 700;
  color: var(--text-primary);
  letter-spacing: -0.5px;
  word-break: break-all;
  min-width: 0;
}

/* Responsive */
@media (max-width: 900px) {
  .stats-row { gap: 10px; }
  :deep(.stat-card .n-statistic__value) { font-size: 20px; }
}
@media (max-width: 600px) {
  .stats-row { grid-template-columns: 1fr; gap: 8px; }
  .port-row { flex-wrap: wrap; gap: 8px; }
  .port-addr { width: 100%; margin-left: 68px; }
  :deep(.stat-card .n-statistic__value) { font-size: 18px; }
}
</style>
