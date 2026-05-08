<script setup>
import { ref, shallowRef } from 'vue'
import { NMessageProvider, NLayout, NLayoutSider, NMenu, NButton } from 'naive-ui'
import ProxyPanel from './views/ProxyPanel.vue'
import RoutePanel from './views/RoutePanel.vue'
import LogViewer from './views/LogViewer.vue'
import SettingsPanel from './views/SettingsPanel.vue'

const collapsed = ref(false)
const activeTab = ref('proxy')

const menuOptions = [
  { label: () => '代理', key: 'proxy', icon: () => '📡' },
  { label: () => '路由', key: 'routes', icon: () => '🔀' },
  { label: () => '日志', key: 'logs', icon: () => '📋' },
  { label: () => '设置', key: 'settings', icon: () => '⚙' },
]

const currentView = shallowRef(ProxyPanel)

function handleUpdate(key) {
  activeTab.value = key
  const views = { proxy: ProxyPanel, routes: RoutePanel, logs: LogViewer, settings: SettingsPanel }
  currentView.value = views[key] || ProxyPanel
}
</script>

<template>
  <n-message-provider>
    <n-layout has-sider position="absolute" style="height: 100vh;">
      <n-layout-sider
        bordered
        :collapsed="collapsed"
        collapse-mode="width"
        :collapsed-width="56"
        :width="180"
        :native-scrollbar="false"
        style="background: #1e1e2e;"
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
          <div style="padding: 8px; text-align: center; border-top: 1px solid #313244;">
            <n-button quaternary size="small" @click="collapsed = !collapsed" style="color: #6c7086;">
              {{ collapsed ? '»' : '« 收缩' }}
            </n-button>
          </div>
        </template>
      </n-layout-sider>
      <n-layout content-style="padding: 20px 24px; background: #181825; color: #cdd6f4;">
        <component :is="currentView" />
      </n-layout>
    </n-layout>
  </n-message-provider>
</template>

<style>
html, body { margin: 0; padding: 0; height: 100%; }
body { background: #181825; }
</style>
