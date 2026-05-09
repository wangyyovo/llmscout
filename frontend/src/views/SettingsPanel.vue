<script setup>
import { ref, onMounted } from 'vue'
import { NButton, NCard, NRadio, NRadioGroup, NIcon, useMessage } from 'naive-ui'
import { TrashOutline, MoonOutline, SunnyOutline, DesktopOutline, FolderOpenOutline } from '@vicons/ionicons5'
import { ClearLogs, GetDbPath } from '../../wailsjs/go/main/App'
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
        <!-- Theme section -->
        <div class="card-section">
          <div class="section-header">
            <div class="section-icon" :style="{ background: mode === 'dark' ? 'rgba(245,169,127,0.1)' : 'rgba(240,140,66,0.1)' }">
              <n-icon size="18" :color="mode === 'dark' ? 'var(--accent-warning)' : '#f08c42'">
                <MoonOutline v-if="mode === 'dark'" /><SunnyOutline v-else-if="mode === 'light'" /><DesktopOutline v-else />
              </n-icon>
            </div>
            <div>
              <div class="section-title">外观主题</div>
              <div class="section-desc">切换浅色/深色模式，或跟随系统设置</div>
            </div>
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

        <!-- DB section -->
        <div class="card-section">
          <div class="section-header">
            <div class="section-icon" style="background:rgba(124,140,248,0.1);">
              <n-icon size="18" color="var(--accent)"><FolderOpenOutline /></n-icon>
            </div>
            <div>
              <div class="section-title">数据存储</div>
              <div class="section-desc">SQLite 数据库文件存放位置</div>
            </div>
          </div>
          <div class="db-path">
            <code class="db-path-text">{{ dbPath }}</code>
          </div>
        </div>

        <div class="section-divider" />

        <!-- Danger zone -->
        <div class="card-section danger-section">
          <div class="section-header">
            <div class="section-icon" style="background:rgba(243,139,168,0.1);">
              <n-icon size="18" color="var(--accent-error)"><TrashOutline /></n-icon>
            </div>
            <div>
              <div class="section-title">危险操作</div>
              <div class="section-desc">清空后将不可恢复，请谨慎操作</div>
            </div>
          </div>
          <n-button type="error" @click="handleClear" secondary>
            <template #icon><n-icon size="16"><TrashOutline /></n-icon></template>
            清空所有日志
          </n-button>
        </div>
      </n-card>
    </div>
  </div>
</template>

<style scoped>
.settings-grid { max-width: 560px; }

.settings-card {
  background: var(--bg-card) !important;
  border-radius: var(--radius) !important;
  box-shadow: var(--shadow-sm);
  overflow: hidden;
  padding: 0 !important;
}
:deep(.settings-card .n-card__content) { padding: 0; }

.card-section { padding: 22px 24px; }
.danger-section {
  background: rgba(243,139,168,0.03);
}

.section-divider {
  height: 1px;
  background: var(--border-color);
  margin: 0 24px;
}

.section-header {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 16px;
}
.section-icon {
  width: 36px; height: 36px;
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.section-title {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  line-height: 1.3;
}
.section-desc {
  font-size: 12px;
  color: var(--text-muted);
  margin-top: 2px;
}

/* Theme radios */
.theme-radios {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-left: 48px;
}
.radio-item { padding: 3px 0; }
.radio-label {
  display: flex;
  align-items: center;
  gap: 7px;
  font-size: 13px;
  color: var(--text-primary);
}

/* DB path */
.db-path {
  background: var(--bg-code);
  border-radius: var(--radius-sm);
  padding: 12px 14px;
  border: 1px solid var(--border-color);
  margin-left: 48px;
}
.db-path-text {
  font-size: 12px;
  color: var(--text-secondary);
  font-family: 'SF Mono', 'Fira Code', monospace;
  word-break: break-all;
  line-height: 1.5;
}

/* Danger zone */
.danger-section .section-desc { color: var(--accent-error); opacity: 0.7; }
.danger-section .n-button { margin-left: 48px; }
</style>
