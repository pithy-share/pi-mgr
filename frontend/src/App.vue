<template>
  <div id="app-root">
    <header class="header">
      <div style="display:flex;align-items:center;gap:16px;">
        <h1>Pi Provider &amp; Model Manager</h1>
        <nav class="nav-tabs">
          <router-link to="/" class="nav-tab" :class="{ active: $route.path === '/' }">方案管理</router-link>
          <router-link to="/ssh-sync" class="nav-tab" :class="{ active: $route.path === '/ssh-sync' }">SSH 同步</router-link>
          <router-link to="/pi-manage" class="nav-tab" :class="{ active: $route.path === '/pi-manage' }">Pi 管理</router-link>
        </nav>
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