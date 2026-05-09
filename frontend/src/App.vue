<script setup>
import { ref, shallowRef, h, computed } from 'vue'
import { NConfigProvider, NMessageProvider, NLayout, NLayoutSider, NMenu, NButton, NIcon } from 'naive-ui'
import { useTheme } from './composables/useTheme.js'
import { ServerOutline, GitBranchOutline, DocumentTextOutline, SettingsOutline } from '@vicons/ionicons5'
import ProxyPanel from './views/ProxyPanel.vue'
import RoutePanel from './views/RoutePanel.vue'
import LogViewer from './views/LogViewer.vue'
import SettingsPanel from './views/SettingsPanel.vue'

const { naiveTheme, themeClass } = useTheme()

const collapsed = ref(false)
const activeTab = ref('proxy')

const renderIcon = (icon) => () => h(NIcon, { size: 18 }, () => h(icon))

const menuOptions = [
  { label: () => '代理', key: 'proxy', icon: renderIcon(ServerOutline) },
  { label: () => '路由', key: 'routes', icon: renderIcon(GitBranchOutline) },
  { label: () => '日志', key: 'logs', icon: renderIcon(DocumentTextOutline) },
  { label: () => '设置', key: 'settings', icon: renderIcon(SettingsOutline) },
]

const currentView = shallowRef(ProxyPanel)

function handleUpdate(key) {
  activeTab.value = key
  const views = { proxy: ProxyPanel, routes: RoutePanel, logs: LogViewer, settings: SettingsPanel }
  currentView.value = views[key] || ProxyPanel
}

const siderStyle = computed(() => ({
  background: 'var(--bg-sider)',
  borderRight: '1px solid var(--border-color)'
}))

const layoutStyle = computed(() => ({
  padding: '20px 24px',
  background: 'var(--bg-main)',
  color: 'var(--text-primary)'
}))
</script>

<template>
  <n-config-provider :theme="naiveTheme" :class="themeClass">
    <n-message-provider>
      <n-layout has-sider position="absolute" style="height: 100vh;">
        <n-layout-sider
          bordered
          :collapsed="collapsed"
          collapse-mode="width"
          :collapsed-width="56"
          :width="180"
          :native-scrollbar="false"
          :style="siderStyle"
        >
          <n-menu
            :collapsed="collapsed"
            :collapsed-width="56"
            :collapsed-icon-size="20"
            :options="menuOptions"
            :value="activeTab"
            @update:value="handleUpdate"
          />
          <template #collapse-extra>
            <div style="padding: 8px; text-align: center; border-top: 1px solid var(--border-color);">
              <n-button quaternary size="small" @click="collapsed = !collapsed" style="color: var(--text-muted);">
                {{ collapsed ? '»' : '« 收缩' }}
              </n-button>
            </div>
          </template>
        </n-layout-sider>
        <n-layout :content-style="layoutStyle">
          <component :is="currentView" />
        </n-layout>
      </n-layout>
    </n-message-provider>
  </n-config-provider>
</template>

<style>
*, *::before, *::after { box-sizing: border-box; }
html, body { margin: 0; padding: 0; height: 100%; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, system-ui, sans-serif; }

/* Dark theme (default) */
body, .theme-dark {
  --bg-main: #181825;
  --bg-card: #1e1e2e;
  --bg-sider: #1e1e2e;
  --bg-code: #11111b;
  --bg-message: #1e1e2e;
  --bg-hover: rgba(137,180,250,0.04);
  --border-color: #313244;
  --text-primary: #cdd6f4;
  --text-secondary: #a6adc8;
  --text-muted: #6c7086;
  --accent: #89b4fa;
  --shadow: 0 1px 3px rgba(0,0,0,0.3);
  --radius: 8px;
  --radius-sm: 4px;
}

/* Light theme */
.theme-light {
  --bg-main: #f5f5f5;
  --bg-card: #ffffff;
  --bg-sider: #fafafa;
  --bg-code: #f0f0f0;
  --bg-message: #f8f8f8;
  --bg-hover: rgba(0,0,0,0.03);
  --border-color: #e0e0e0;
  --text-primary: #333333;
  --text-secondary: #666666;
  --text-muted: #999999;
  --accent: #2563eb;
  --shadow: 0 1px 3px rgba(0,0,0,0.08);
  --radius: 8px;
  --radius-sm: 4px;
}

body { background: var(--bg-main); color: var(--text-primary); }

::-webkit-scrollbar { width: 6px; height: 6px; }
::-webkit-scrollbar-track { background: transparent; }
::-webkit-scrollbar-thumb { background: var(--border-color); border-radius: 3px; }
::-webkit-scrollbar-thumb:hover { background: var(--text-muted); }

.n-card { transition: box-shadow 0.15s ease; }
.n-card:hover { box-shadow: var(--shadow); }

.fade-enter-active, .fade-leave-active { transition: opacity 0.2s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
</style>
