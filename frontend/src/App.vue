<template>
  <div id="app-root">
    <header class="header">
      <div style="display:flex;align-items:center;gap:16px;">
        <h1>Pi Provider & Model Manager</h1>
      </div>
    </header>
    <router-view />
    <div v-if="toast" :class="['toast', 'toast-' + toast.type]">{{ toast.message }}</div>
  </div>
</template>

<script setup lang="ts">
import { provide, ref } from 'vue'
import type { Toast } from './types'

const toast = ref<Toast | null>(null)
let toastTimer: ReturnType<typeof setTimeout> | null = null

function showToast(message: string, type: 'success' | 'error') {
  toast.value = { message, type }
  if (toastTimer) clearTimeout(toastTimer)
  toastTimer = setTimeout(() => { toast.value = null }, 3000)
}

provide('showToast', showToast)
</script>