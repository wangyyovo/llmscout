<script setup>
import { ref, onMounted } from 'vue'
import { NButton, NCard, NSpace, NRadio, NRadioGroup, NIcon, useMessage } from 'naive-ui'
import { TrashOutline, MoonOutline, SunnyOutline, DesktopOutline, FolderOpenOutline } from '@vicons/ionicons5'
import { GetSetting, SetSetting, ClearLogs, GetDbPath } from '../../wailsjs/go/main/App'
import { useThemeInject } from '../composables/useTheme.js'

const message = useMessage()
const { mode, setMode } = useThemeInject()
const dbPath = ref('')

onMounted(async () => {
  dbPath.value = await GetDbPath()
})

async function handleClear() {
  await ClearLogs()
  message.success('日志已清空')
}
</script>

<template>
  <div>
    <h2 class="page-title" style="margin-bottom: 20px;">设置</h2>

    <div class="settings-grid">
      <n-card class="settings-card" :bordered="false">
        <div class="card-section">
          <div class="section-header">
            <n-icon size="18" :color="mode === 'dark' ? '#f5a97f' : '#f08c42'"><MoonOutline v-if="mode === 'dark'" /><SunnyOutline v-else-if="mode === 'light'" /><DesktopOutline v-else /></n-icon>
            <span class="section-title">外观主题</span>
          </div>
          <n-radio-group :value="mode" @update:value="setMode" class="theme-radios">
            <n-radio value="light" class="radio-item">
              <span class="radio-label">
                <n-icon size="15"><SunnyOutline /></n-icon> 浅色
              </span>
            </n-radio>
            <n-radio value="dark" class="radio-item">
              <span class="radio-label">
                <n-icon size="15"><MoonOutline /></n-icon> 深色
              </span>
            </n-radio>
            <n-radio value="system" class="radio-item">
              <span class="radio-label">
                <n-icon size="15"><DesktopOutline /></n-icon> 跟随系统
              </span>
            </n-radio>
          </n-radio-group>
        </div>

        <div class="section-divider" />

        <div class="card-section">
          <div class="section-header">
            <n-icon size="18" color="var(--accent)"><FolderOpenOutline /></n-icon>
            <span class="section-title">数据存储</span>
          </div>
          <div class="db-path">
            <code class="db-path-text">{{ dbPath }}</code>
          </div>
        </div>

        <div class="section-divider" />

        <div class="card-section">
          <div class="section-header">
            <n-icon size="18" color="var(--accent-error)"><TrashOutline /></n-icon>
            <span class="section-title">危险操作</span>
          </div>
          <p class="danger-desc">清空后将不可恢复</p>
          <n-button type="error" @click="handleClear" :border-radius="6" secondary>
            <template #icon><n-icon size="16"><TrashOutline /></n-icon></template>
            清空所有日志
          </n-button>
        </div>
      </n-card>
    </div>
  </div>
</template>

<style scoped>
.settings-grid {
  max-width: 560px;
}

.settings-card {
  background: var(--bg-card) !important;
  border-radius: var(--radius) !important;
  box-shadow: var(--shadow-sm);
  overflow: hidden;
}

.card-section {
  padding: 20px 24px;
}

.section-divider {
  height: 1px;
  background: var(--border-color);
  margin: 0 24px;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 14px;
}
.section-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
}

/* Theme radios */
.theme-radios {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.radio-item {
  padding: 4px 0;
}
.radio-label {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
  color: var(--text-primary);
}

/* DB path */
.db-path {
  background: var(--bg-code);
  border-radius: var(--radius-sm);
  padding: 12px 16px;
  border: 1px solid var(--border-color);
}
.db-path-text {
  font-size: 12px;
  color: var(--text-secondary);
  font-family: 'SF Mono', 'Fira Code', monospace;
  word-break: break-all;
}

/* Danger zone */
.danger-desc {
  color: var(--text-muted);
  font-size: 12px;
  margin: 0 0 10px 0;
}
</style>
