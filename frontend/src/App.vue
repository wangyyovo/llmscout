<script setup>
import { ref, shallowRef, h, computed, onMounted, onUnmounted } from 'vue'
import { NConfigProvider, NMessageProvider, NLayout, NLayoutSider, NMenu, NButton, NIcon } from 'naive-ui'
import { useTheme } from './composables/useTheme.js'
import { ServerOutline, GitBranchOutline, DocumentTextOutline, SettingsOutline, ChevronBackOutline, ChevronForwardOutline } from '@vicons/ionicons5'
import ProxyPanel from './views/ProxyPanel.vue'
import RoutePanel from './views/RoutePanel.vue'
import LogViewer from './views/LogViewer.vue'
import SettingsPanel from './views/SettingsPanel.vue'
import logoSvg from './assets/logo.svg'

const { naiveTheme, themeClass } = useTheme()

const collapsed = ref(false)
const autoCollapsed = ref(false)
const activeTab = ref('proxy')

const sidebarCollapsed = computed(() => autoCollapsed.value || collapsed.value)

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
  padding: '28px 32px',
  background: 'var(--bg-main)',
  color: 'var(--text-primary)'
}))

// Auto-collapse sidebar on narrow windows
let mql = null
function onResize(e) {
  autoCollapsed.value = e.matches
}
onMounted(() => {
  mql = window.matchMedia('(max-width: 900px)')
  autoCollapsed.value = mql.matches
  mql.addEventListener('change', onResize)
})
onUnmounted(() => {
  if (mql) mql.removeEventListener('change', onResize)
})

// Shared NaiveUI theme overrides for consistent control sizing
const themeOverrides = {
  common: {
    fontSize: '13px',
    fontSizeSmall: '12px',
    fontSizeTiny: '11px',
    borderRadius: '6px',
    borderRadiusSmall: '4px',
  },
  Input: { height: '34px', fontSize: '13px' },
  Select: { peers: { InternalSelection: { height: '34px', fontSize: '13px' } } },
  Button: {
    heightSmall: '30px',
    heightMedium: '34px',
    heightLarge: '40px',
    fontSizeSmall: '12px',
    fontSizeMedium: '13px',
    fontSizeLarge: '15px',
    borderRadius: '6px',
  },
  Tag: { borderRadius: '4px' },
  Card: { borderRadius: '10px', padding: '20px' },
  Switch: { railHeight: '18px', railWidth: '32px', buttonHeight: '14px', buttonWidth: '14px' },
}
</script>

<template>
  <n-config-provider :theme="naiveTheme" :theme-overrides="themeOverrides" :class="themeClass">
    <n-message-provider>
      <n-layout has-sider position="absolute" style="height: 100vh;">
        <n-layout-sider
          bordered
          :collapsed="sidebarCollapsed"
          collapse-mode="width"
          :collapsed-width="56"
          :width="200"
          :native-scrollbar="false"
          :style="siderStyle"
        >
          <div class="sider-brand">
            <img :src="logoSvg" class="sider-logo" alt="LLM Scout" />
            <span v-show="!sidebarCollapsed" class="sider-title">LLM Scout</span>
          </div>
          <n-menu
            :collapsed="sidebarCollapsed"
            :collapsed-width="56"
            :collapsed-icon-size="20"
            :options="menuOptions"
            :value="activeTab"
            @update:value="handleUpdate"
          />
          <template #collapse-extra>
            <div class="sider-footer">
              <n-button quaternary size="small" @click="collapsed = !collapsed" class="collapse-btn">
                <template #icon>
                  <n-icon size="16"><ChevronBackOutline v-if="!sidebarCollapsed" /><ChevronForwardOutline v-else /></n-icon>
                </template>
                <span v-show="!sidebarCollapsed" class="collapse-label">收起菜单</span>
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
html, body { margin: 0; padding: 0; height: 100%; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', 'PingFang SC', 'Microsoft YaHei', Roboto, system-ui, sans-serif; }

/* Dark theme */
body, .theme-dark {
  --bg-main: #0f0f1a;
  --bg-card: #18182a;
  --bg-sider: #141425;
  --bg-code: #0a0a14;
  --bg-message: #18182a;
  --bg-hover: rgba(124,140,248,0.06);
  --border-color: #262640;
  --text-primary: #e2e4f0;
  --text-secondary: #a0a4b8;
  --text-muted: #6b6f85;
  --accent: #7c8cf8;
  --accent-hover: #909ef9;
  --accent-success: #5ce6b8;
  --accent-warning: #f5a97f;
  --accent-error: #f38ba8;
  --shadow-sm: 0 1px 2px rgba(0,0,0,0.2);
  --shadow: 0 2px 8px rgba(0,0,0,0.25);
  --shadow-lg: 0 4px 16px rgba(0,0,0,0.3);
  --radius: 10px;
  --radius-sm: 6px;
  --radius-xs: 4px;
  --transition: 0.2s ease;
}

/* Light theme */
.theme-light {
  --bg-main: #f0f2f5;
  --bg-card: #ffffff;
  --bg-sider: #f8f9fb;
  --bg-code: #f5f5f5;
  --bg-message: #fafafa;
  --bg-hover: rgba(124,140,248,0.05);
  --border-color: #e5e6eb;
  --text-primary: #1d1f2a;
  --text-secondary: #606470;
  --text-muted: #9498a3;
  --accent: #5b6cf0;
  --accent-hover: #4a5de0;
  --accent-success: #22b88b;
  --accent-warning: #f08c42;
  --accent-error: #e0556f;
  --shadow-sm: 0 1px 2px rgba(0,0,0,0.04);
  --shadow: 0 2px 8px rgba(0,0,0,0.06);
  --shadow-lg: 0 4px 16px rgba(0,0,0,0.08);
  --radius: 10px;
  --radius-sm: 6px;
  --radius-xs: 4px;
  --transition: 0.2s ease;
}

body { background: var(--bg-main); color: var(--text-primary); }

::-webkit-scrollbar { width: 6px; height: 6px; }
::-webkit-scrollbar-track { background: transparent; }
::-webkit-scrollbar-thumb { background: var(--border-color); border-radius: 3px; }
::-webkit-scrollbar-thumb:hover { background: var(--text-muted); }

/* Sidebar brand */
.sider-brand {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 20px 18px 16px;
  border-bottom: 1px solid var(--border-color);
  margin-bottom: 8px;
}
.sider-logo {
  width: 28px;
  height: 28px;
  flex-shrink: 0;
  border-radius: 6px;
}
.sider-title {
  font-size: 16px;
  font-weight: 700;
  color: var(--text-primary);
  letter-spacing: -0.3px;
  white-space: nowrap;
}

/* Sidebar footer */
.sider-footer {
  padding: 8px;
  text-align: center;
  border-top: 1px solid var(--border-color);
}
.collapse-btn {
  color: var(--text-muted);
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  transition: color var(--transition);
  border-radius: var(--radius-sm);
}
.collapse-btn:hover { color: var(--text-secondary); }
.collapse-label { font-size: 12px; }

/* Sidebar width transition */
.n-layout-sider {
  transition: width 0.25s ease !important;
}

/* Page section headers */
.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
}
.page-title {
  font-size: 20px;
  font-weight: 700;
  color: var(--text-primary);
  letter-spacing: -0.3px;
  margin: 0;
}

/* Card refinements */
.n-card {
  border-radius: var(--radius) !important;
  transition: box-shadow var(--transition), transform var(--transition);
}
.n-card:hover {
  box-shadow: var(--shadow);
}
.card-accent {
  position: relative;
  overflow: hidden;
}
.card-accent::before {
  content: '';
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 3px;
  background: var(--accent);
  border-radius: 0 2px 2px 0;
}

/* Fade transitions */
.fade-enter-active, .fade-leave-active { transition: opacity 0.2s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; }

/* Responsive */
@media (max-width: 900px) {
  .n-layout-content { padding: 20px 20px !important; }
  .page-title { font-size: 18px; }
}
@media (max-width: 600px) {
  .n-layout-content { padding: 16px 14px !important; }
  .page-title { font-size: 17px; }
  .page-header { margin-bottom: 14px; }
}
</style>
