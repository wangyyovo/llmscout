<script setup>
import { ref, onMounted } from 'vue'
import { NButton, NCard, NRadio, NRadioGroup, NIcon, NDivider, useMessage } from 'naive-ui'
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
    <h2 class="page-title" style="margin-bottom: 24px;">设置</h2>

    <div class="settings-grid">
      <!-- 外观主题 -->
      <n-card :bordered="false" class="block-card">
        <div class="block-header">
          <div class="block-icon" :style="{
            background: mode === 'dark' ? 'rgba(245,169,127,0.12)' : 'rgba(240,140,66,0.1)',
          }">
            <n-icon size="20" :color="mode === 'dark' ? 'var(--accent-warning)' : '#f08c42'">
              <MoonOutline v-if="mode === 'dark'" />
              <SunnyOutline v-else-if="mode === 'light'" />
              <DesktopOutline v-else />
            </n-icon>
          </div>
          <div>
            <div class="block-title">外观主题</div>
            <div class="block-desc">切换浅色 / 深色模式，或跟随系统自动切换</div>
          </div>
        </div>

        <n-radio-group :value="mode" @update:value="setMode" class="block-body">
          <n-radio value="light">
            <span class="radio-label"><n-icon size="16"><SunnyOutline /></n-icon> 浅色</span>
          </n-radio>
          <n-radio value="dark">
            <span class="radio-label"><n-icon size="16"><MoonOutline /></n-icon> 深色</span>
          </n-radio>
          <n-radio value="system">
            <span class="radio-label"><n-icon size="16"><DesktopOutline /></n-icon> 跟随系统</span>
          </n-radio>
        </n-radio-group>
      </n-card>

      <!-- 数据存储 -->
      <n-card :bordered="false" class="block-card">
        <div class="block-header">
          <div class="block-icon" style="background: rgba(124,140,248,0.1);">
            <n-icon size="20" color="var(--accent)"><FolderOpenOutline /></n-icon>
          </div>
          <div>
            <div class="block-title">数据存储</div>
            <div class="block-desc">SQLite 数据库文件存放位置</div>
          </div>
        </div>

        <div class="block-body">
          <div class="db-path">
            <code class="db-path-text">{{ dbPath }}</code>
          </div>
        </div>
      </n-card>

      <!-- 危险操作 -->
      <n-card :bordered="false" class="block-card danger-card">
        <div class="block-header">
          <div class="block-icon" style="background: rgba(243,139,168,0.1);">
            <n-icon size="20" color="var(--accent-error)"><TrashOutline /></n-icon>
          </div>
          <div>
            <div class="block-title" style="color: var(--accent-error);">危险操作</div>
            <div class="block-desc">清空所有日志数据，操作不可恢复</div>
          </div>
        </div>

        <div class="block-body">
          <n-button type="error" @click="handleClear" secondary size="medium">
            <template #icon><n-icon size="17"><TrashOutline /></n-icon></template>
            清空所有日志
          </n-button>
        </div>
      </n-card>
    </div>
  </div>
</template>

<style scoped>
.settings-grid {
  max-width: 600px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

/* Card block */
.block-card {
  background: var(--bg-card) !important;
  border-radius: var(--radius) !important;
  box-shadow: var(--shadow-sm);
  padding: 0 !important;
  transition: box-shadow var(--transition);
}
.block-card:hover {
  box-shadow: var(--shadow);
}
:deep(.block-card .n-card__content) { padding: 22px 24px; }

.block-header {
  display: flex;
  align-items: flex-start;
  gap: 14px;
  margin-bottom: 18px;
}
.block-icon {
  width: 40px; height: 40px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}
.block-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
  line-height: 1.3;
}
.block-desc {
  font-size: 12px;
  color: var(--text-muted);
  margin-top: 3px;
  line-height: 1.4;
}
.block-body {
  margin-left: 54px;
}

/* Theme radios */
:deep(.block-body .n-radio-group) {
  display: flex;
  flex-direction: column;
  gap: 12px;
}
.radio-label {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
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
  line-height: 1.6;
}

/* Danger card */
.danger-card {
  border: 1px solid rgba(243,139,168,0.2) !important;
}
</style>
