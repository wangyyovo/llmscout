import { ref, computed, watch, provide, inject } from 'vue'
import { darkTheme } from 'naive-ui'
import { GetSetting, SetSetting } from '../../wailsjs/go/main/App'

const THEME_KEY = 'theme'

// Shared state
const mode = ref('system')
const systemDark = ref(false)
let initialized = false

// Setup system preference listener once
if (typeof window !== 'undefined') {
  const mq = window.matchMedia('(prefers-color-scheme: dark)')
  systemDark.value = mq.matches
  mq.addEventListener('change', (e) => { systemDark.value = e.matches })
}

export function useTheme() {
  if (!initialized) {
    initialized = true
    GetSetting(THEME_KEY, 'system').then(val => {
      mode.value = val
    })
  }

  const isDark = computed(() => {
    if (mode.value === 'system') return systemDark.value
    return mode.value === 'dark'
  })

  const naiveTheme = computed(() => isDark.value ? darkTheme : null)

  const themeClass = computed(() => {
    if (mode.value === 'system') return systemDark.value ? 'theme-dark' : 'theme-light'
    return mode.value === 'dark' ? 'theme-dark' : 'theme-light'
  })

  function setMode(newMode) {
    mode.value = newMode
    SetSetting(THEME_KEY, newMode)
  }

  // Apply theme class to body
  watch(themeClass, (cls) => {
    document.body.className = cls
  }, { immediate: true })

  provide(THEME_KEY, { mode, setMode, isDark, naiveTheme })

  return { mode, setMode, isDark, naiveTheme, themeClass }
}

export function useThemeInject() {
  return inject(THEME_KEY, {
    mode: ref('dark'),
    setMode: () => {},
    isDark: ref(true),
    naiveTheme: ref(darkTheme),
  })
}
