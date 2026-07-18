<template>
  <div id="app-root">
    <header class="header">
      <div style="display:flex;align-items:center;gap:16px;">
        <button v-if="showBack" class="nav-back" @click="$router.push('/')">← 返回方案列表</button>
        <h1>Pi Provider & Model Manager</h1>
        <span v-if="activeSchemeName" class="active-badge" :title="'当前激活方案: ' + activeSchemeName">
          🟢 {{ activeSchemeName }}
        </span>
      </div>
    </header>
    <router-view />
    <div v-if="toast" :class="['toast', 'toast-' + toast.type]">{{ toast.message }}</div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, provide, ref } from 'vue'
import { useRoute } from 'vue-router'
import type { Toast, Scheme } from './types'
import api from './wails/api'

const route = useRoute()
const showBack = computed(() => route.path.startsWith('/scheme/'))

const toast = ref<Toast | null>(null)
let toastTimer: ReturnType<typeof setTimeout> | null = null

function showToast(message: string, type: 'success' | 'error') {
  toast.value = { message, type }
  if (toastTimer) clearTimeout(toastTimer)
  toastTimer = setTimeout(() => { toast.value = null }, 3000)
}

provide('showToast', showToast)

// Active scheme tracking
const activeSchemeId = ref('')
const activeSchemeName = ref('')

async function refreshActiveScheme() {
  try {
    const a = api()
    const id = await a.GetActiveSchemeID()
    activeSchemeId.value = id || ''
    if (id) {
      const schemes = await a.ListSchemes()
      const scheme = schemes.find(s => s.id === id)
      activeSchemeName.value = scheme?.name || ''
    } else {
      activeSchemeName.value = ''
    }
  } catch {
    // dev fallback
  }
}

provide('activeSchemeId', activeSchemeId)
provide('refreshActiveScheme', refreshActiveScheme)

onMounted(async () => {
  await refreshActiveScheme()
})
</script>
