<template>
  <div style="max-width:600px;margin:0 auto;padding:20px;">
    <h2 style="margin-bottom:20px;">SSH 配置同步</h2>

    <!-- SSH Address input -->
    <div class="card" style="margin-bottom:16px;">
      <div class="form-group">
        <label>SSH 地址</label>
        <input
          v-model="address"
          placeholder="user@host[:port]"
          :disabled="isTesting || isSyncing"
          @input="clearResults"
        />
        <div v-if="addressError" class="field-error">{{ addressError }}</div>
      </div>
    </div>

    <!-- Action buttons -->
    <div style="display:flex;gap:8px;margin-bottom:20px;">
      <button
        class="btn-secondary"
        :disabled="!canOperate || isTesting || isSyncing"
        @click="handleTestConnection"
      >
        {{ isTesting ? '连接测试中...' : '测试连接' }}
      </button>
      <button
        class="btn-primary"
        :disabled="!canOperate || isTesting || isSyncing"
        @click="handleSync"
      >
        {{ isSyncing ? '同步中...' : '开始同步' }}
      </button>
    </div>

    <!-- Connection test result -->
    <div v-if="connectionResult" :class="['card', connectionResult.success ? 'result-success' : 'result-error']" style="margin-bottom:16px;">
      <div style="display:flex;align-items:center;gap:8px;">
        <span :style="{ fontSize: '18px' }">{{ connectionResult.success ? '✓' : '✗' }}</span>
        <span>{{ connectionResult.message }}</span>
      </div>
    </div>

    <!-- Sync result -->
    <div v-if="syncResult" class="card" style="margin-bottom:16px;">
      <h3 style="margin-bottom:12px;">
        同步结果
        <span v-if="syncResult.overall === 'success'" style="color:var(--success);">(全部成功)</span>
        <span v-else-if="syncResult.overall === 'partial'" style="color:var(--accent);">(部分成功)</span>
        <span v-else style="color:var(--danger);">(全部失败)</span>
      </h3>

      <div v-for="(item, idx) in syncResult.items" :key="idx"
        :class="['sync-item', 'sync-item-' + item.status]">
        <span class="sync-item-status">
          <span v-if="item.status === 'success'">✓</span>
          <span v-else-if="item.status === 'skipped'">−</span>
          <span v-else>✗</span>
        </span>
        <span class="sync-item-name">{{ item.name || '系统' }}</span>
        <span class="sync-item-msg">{{ item.message }}</span>
      </div>

      <!-- Post-sync tip (Q6) -->
      <div v-if="syncResult.overall !== 'failed'" style="margin-top:12px;padding:10px;background:#f0f8ff;border-radius:var(--radius);font-size:13px;color:#2a6496;">
        💡 请在 Ubuntu 上重启 Pi 以使配置生效
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import api from '../wails/api'
import type { SyncResult } from '../types'

const address = ref('')
const connectionResult = ref<{ success: boolean; message: string } | null>(null)
const syncResult = ref<SyncResult | null>(null)
const isTesting = ref(false)
const isSyncing = ref(false)

// Load saved address on mount (AC-02)
onMounted(async () => {
  try {
    const a = api()
    const saved = await a.LoadSSHAddress()
    if (saved) {
      address.value = saved
    } else {
      // Set default value when no saved address (AC-01)
      address.value = 'zyong@192.168.1.180'
    }
  } catch {
    // Fallback to default
    address.value = 'zyong@192.168.1.180'
  }
})

// Auto-save address on change (AC-02)
watch(address, (newVal) => {
  if (newVal !== undefined) {
    const a = api()
    a.SaveSSHAddress(newVal).catch(() => { /* silent */ })
  }
})

// Compute address error (AC-17, AC-18)
const addressError = computed(() => {
  const val = address.value
  if (!val || !val.trim()) {
    return '请填写 SSH 地址'
  }
  // Format: user@host or user@host:port
  // No spaces, exactly one @, host non-empty, port digits only if present
  if (!/^[^\s@]+@[^\s@]+(:[0-9]+)?$/.test(val.trim())) {
    return 'SSH 地址格式应为 user@host[:port]'
  }
  return ''
})

// Can operate: no error and non-empty
const canOperate = computed(() => {
  return !addressError.value && address.value.trim() !== ''
})

// Clear results when address changes
function clearResults() {
  connectionResult.value = null
  syncResult.value = null
}

// Test SSH connection (AC-03)
async function handleTestConnection() {
  isTesting.value = true
  connectionResult.value = null
  try {
    const a = api()
    const result = await a.TestSSHConnection(address.value.trim())
    connectionResult.value = result
  } catch (e: any) {
    connectionResult.value = { success: false, message: e?.message || '连接测试失败' }
  } finally {
    isTesting.value = false
  }
}

// Start sync (AC-04)
async function handleSync() {
  isSyncing.value = true
  syncResult.value = null
  try {
    const a = api()
    const result = await a.SyncPiConfig(address.value.trim())
    syncResult.value = result
  } catch (e: any) {
    syncResult.value = {
      overall: 'failed',
      items: [{ name: '', status: 'failed', message: e?.message || '同步失败' }],
    }
  } finally {
    isSyncing.value = false
  }
}
</script>

<style scoped>
.result-success {
  background: var(--success-bg, #eafaf1);
  border-color: var(--success);
}
.result-error {
  background: #fdedec;
  border-color: var(--danger);
}
.sync-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 0;
  border-bottom: 1px solid var(--border-color);
  font-size: 14px;
}
.sync-item:last-child {
  border-bottom: none;
}
.sync-item-status {
  width: 20px;
  text-align: center;
  flex-shrink: 0;
}
.sync-item-success .sync-item-status { color: var(--success); }
.sync-item-skipped .sync-item-status { color: var(--text-secondary); }
.sync-item-failed .sync-item-status { color: var(--danger); }
.sync-item-name {
  font-weight: 500;
  min-width: 100px;
  flex-shrink: 0;
}
.sync-item-msg {
  color: var(--text-secondary);
  font-size: 13px;
  overflow: hidden;
  text-overflow: ellipsis;
}
</style>
