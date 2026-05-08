<script setup>
import { ref } from 'vue'
import { NMessageProvider, NLayout, NLayoutSider, NMenu, NButton } from 'naive-ui'

const collapsed = ref(false)
const menuOptions = [
  {
    label: '对话',
    key: 'chat',
  },
  {
    label: '设置',
    key: 'settings',
  },
]
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
        />
        <template #collapse-extra>
          <div style="padding: 8px; text-align: center; border-top: 1px solid #313244;">
            <n-button quaternary size="small" @click="collapsed = !collapsed">
              {{ collapsed ? '»' : '« 收缩' }}
            </n-button>
          </div>
        </template>
      </n-layout-sider>
      <n-layout content-style="padding: 0; background: #181825;">
        <router-view />
      </n-layout>
    </n-layout>
  </n-message-provider>
</template>

<style>
html, body { margin: 0; padding: 0; height: 100%; }
body { background: #181825; color: #cdd6f4; }
</style>
