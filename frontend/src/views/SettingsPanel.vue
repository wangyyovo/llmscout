<script setup>
import { ref, onMounted } from 'vue'
import { NInput, NButton, NCard, NSpace, useMessage } from 'naive-ui'
import { GetSetting, SetSetting, ClearLogs } from '../../wailsjs/go/main/App'

const message = useMessage()
const port = ref('8899')
const dbPath = ref('')

onMounted(async () => {
  port.value = await GetSetting('port', '8899')
  dbPath.value = await GetSetting('dbPath', '~/.llmscout/data.db')
})

async function savePort() {
  await SetSetting('port', port.value)
  message.success('端口设置已保存，下次启动代理时生效')
}

async function handleClear() {
  await ClearLogs()
  message.success('日志已清空')
}
</script>

<template>
  <div>
    <h2 style="color: #cdd6f4; margin-bottom: 20px;">⚙ 设置</h2>
    <n-card style="background: #1e1e2e; border: none; max-width: 500px;">
      <n-space vertical size="large">
        <div>
          <div style="color: #a6adc8; margin-bottom: 6px;">代理端口</div>
          <div style="display: flex; gap: 10px;">
            <n-input v-model:value="port" type="number" style="width: 120px;" />
            <n-button type="primary" @click="savePort">保存</n-button>
          </div>
        </div>
        <div>
          <div style="color: #a6adc8; margin-bottom: 6px;">数据库路径</div>
          <div style="color: #6c7086; font-size: 13px;">{{ dbPath }}</div>
        </div>
        <n-button type="error" @click="handleClear">清空所有日志</n-button>
      </n-space>
    </n-card>
  </div>
</template>
