<template>
  <div style="max-width:700px;margin:0 auto;padding:20px;">
    <h2 style="margin-bottom:20px;">Pi 管理</h2>

    <!-- Mode toggle -->
    <div class="card" style="margin-bottom:16px;padding:10px 16px;">
      <div style="display:flex;align-items:center;gap:12px;">
        <label style="margin:0;display:flex;align-items:center;gap:4px;cursor:pointer;">
          <input type="checkbox" v-model="useRemote" :disabled="!sshAddress" style="width:auto;" />
          <span>远程模式</span>
        </label>
        <span v-if="sshAddress" style="font-size:13px;color:var(--text-secondary);">{{ sshAddress }}</span>
        <span v-else style="font-size:13px;color:var(--text-secondary);">未配置 SSH 地址（请在 SSH 同步页设置）</span>
      </div>
    </div>

    <!-- Pi version section -->
    <div class="card" style="margin-bottom:16px;">
      <div style="display:flex;align-items:center;justify-content:space-between;">
        <div>
          <strong>Pi 版本</strong>
          <span v-if="piVersion" style="margin-left:8px;">{{ piVersion }}</span>
          <span v-if="piVersionError" style="margin-left:8px;color:var(--danger);">{{ piVersionError }}</span>
          <span v-if="isLoadingVersion" class="spinner" style="margin-left:8px;"></span>
        </div>
        <div style="display:flex;gap:6px;">
          <button class="btn-secondary btn-small" @click="loadVersion" :disabled="isLoadingVersion">
            刷新
          </button>
          <button class="btn-primary btn-small" @click="handleUpdatePi" :disabled="isUpdating">
            {{ isUpdating ? '升级中…' : '升级 Pi' }}
          </button>
        </div>
      </div>
      <div v-if="updateResult" style="margin-top:8px;padding:8px;background:#f5f5f5;border-radius:var(--radius);font-size:13px;white-space:pre-wrap;word-break:break-all;max-height:200px;overflow-y:auto;">
        {{ updateResult }}
      </div>
    </div>

    <!-- Recommended plugins -->
    <div class="card" style="margin-bottom:16px;">
      <h3 style="margin-bottom:12px;">推荐插件</h3>
      <div v-for="src in recommendedPkgs" :key="src"
        class="list-item" style="margin-bottom:2px;align-items:center;">
        <div style="flex:1;min-width:0;overflow:hidden;white-space:nowrap;text-overflow:ellipsis;">
          <span style="font-weight:500;">{{ src }}</span>
        </div>
        <div class="list-item-actions" style="flex-shrink:0;">
          <span v-if="isInstalled(src)" style="font-size:12px;color:var(--success);margin-right:8px;">已安装</span>
          <button v-else class="btn-primary btn-small"
            :disabled="installingPkg === src"
            @click="handleInstall(src)">
            {{ installingPkg === src ? '安装中…' : '安装' }}
          </button>
        </div>
      </div>
      <div v-if="installResult" style="margin-top:8px;padding:8px;background:#f5f5f5;border-radius:var(--radius);font-size:13px;white-space:pre-wrap;word-break:break-all;">
        {{ installResult }}
      </div>
    </div>

    <!-- CBM usage rules (click to copy) -->
    <div class="card" style="margin-bottom:16px;padding:10px 16px;">
      <div style="display:flex;align-items:center;justify-content:space-between;">
        <span style="font-size:14px;">CBM 使用规则</span>
        <button class="btn-secondary btn-small" @click="handleCopyCbm" :disabled="loadingCbm">
          {{ loadingCbm ? '加载中…' : (copiedCbm ? '已复制' : '复制到剪贴板') }}
        </button>
      </div>
      <div v-if="cbmError" style="margin-top:6px;font-size:12px;color:var(--danger);">{{ cbmError }}</div>
    </div>

    <!-- Packages section -->
    <div class="card">
      <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:12px;">
        <h3>已安装插件</h3>
        <div style="display:flex;gap:6px;">
          <button class="btn-secondary btn-small" @click="loadPackages" :disabled="isLoadingPackages">
            刷新
          </button>
          <button class="btn-primary btn-small" @click="handleUpdateAll" :disabled="isUpdatingAll">
            {{ isUpdatingAll ? '升级中…' : '升级全部' }}
          </button>
        </div>
      </div>

      <!-- Package list -->
      <div v-if="packagesError" style="color:var(--danger);margin-bottom:8px;font-size:13px;">
        {{ packagesError }}
      </div>
      <div v-else-if="packages.length === 0 && !isLoadingPackages" style="color:var(--text-secondary);padding:12px 0;font-size:13px;">
        暂无已安装插件
      </div>
      <div v-for="(pkg, idx) in packages" :key="pkg.source"
        class="list-item" style="margin-bottom:2px;align-items:center;">
        <div style="flex:1;min-width:0;overflow:hidden;white-space:nowrap;text-overflow:ellipsis;">
          <span style="font-weight:500;">{{ pkg.source }}</span>
          <span style="font-size:11px;color:var(--text-secondary);margin-left:6px;" :title="pkg.path">{{ pkg.path }}</span>
        </div>
        <div class="list-item-actions" style="flex-shrink:0;">
          <button class="btn-secondary btn-small"
            :disabled="updatingPkg === pkg.source"
            @click="handleUpdateOne(pkg.source)">
            {{ updatingPkg === pkg.source ? '升级中…' : '升级' }}
          </button>
          <button class="btn-danger btn-small"
            :disabled="updatingPkg === pkg.source"
            @click="confirmRemove = pkg.source">
            删除
          </button>
        </div>
      </div>

      <!-- Operation result -->
      <div v-if="operationResult" style="margin-top:8px;padding:8px;background:#f5f5f5;border-radius:var(--radius);font-size:13px;white-space:pre-wrap;word-break:break-all;max-height:150px;overflow-y:auto;">
        {{ operationResult }}
      </div>
    </div>

    <!-- Delete confirmation modal -->
    <div v-if="confirmRemove" class="modal-overlay" @click.self="confirmRemove = ''">
      <div class="modal">
        <h3>确认删除</h3>
        <p style="font-size:14px;">确定删除插件 <strong>{{ confirmRemove }}</strong>？</p>
        <div class="modal-actions">
          <button class="btn-secondary" @click="confirmRemove = ''">取消</button>
          <button class="btn-danger" :disabled="updatingPkg === confirmRemove" @click="handleRemove(confirmRemove)">
            {{ updatingPkg === confirmRemove ? '删除中…' : '确认删除' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import api from '../wails/api'
import type { PiPackage } from '../types'

// Pi version
const piVersion = ref('')
const piVersionError = ref('')
const isLoadingVersion = ref(false)
const updateResult = ref('')
const isUpdating = ref(false)

// Packages
const packages = ref<PiPackage[]>([])
const packagesError = ref('')
const isLoadingPackages = ref(false)
const operationResult = ref('')
const updatingPkg = ref('')
const isUpdatingAll = ref(false)
const confirmRemove = ref('')

// Recommended plugins
const recommendedPkgs = [
  'npm:@vanillagreen/pi-codex-minimal-tools',
  'npm:@gotgenes/pi-subagents',
  'npm:pi-cbm',
]
const installingPkg = ref('')
const installResult = ref('')

// CBM rules
const loadingCbm = ref(false)
const copiedCbm = ref(false)
const cbmError = ref('')
let cbmTimer: ReturnType<typeof setTimeout> | null = null

async function handleCopyCbm() {
  loadingCbm.value = true
  cbmError.value = ''
  try {
    const a = api()
    const rules = await a.GetCbmRules()
    await navigator.clipboard.writeText(rules)
    copiedCbm.value = true
    if (cbmTimer) clearTimeout(cbmTimer)
    cbmTimer = setTimeout(() => { copiedCbm.value = false }, 3000)
  } catch (e: any) {
    cbmError.value = '复制失败'
  } finally {
    loadingCbm.value = false
  }
}

function isInstalled(src: string): boolean {
  return packages.value.some(p => p.source === src)
}

// Remote mode
const sshAddress = ref('')
const useRemote = ref(false)

// Load saved SSH address on mount
async function loadSSHAddress() {
  try {
    const a = api()
    const addr = await a.LoadSSHAddress()
    if (addr) {
      sshAddress.value = addr
    }
  } catch {
    // ignore
  }
}

// --- Pi version ---
async function loadVersion() {
  isLoadingVersion.value = true
  piVersionError.value = ''
  piVersion.value = ''
  try {
    const a = api()
    const ver = useRemote.value && sshAddress.value
      ? await a.GetRemotePiVersion(sshAddress.value)
      : await a.GetPiVersion()
    piVersion.value = ver
  } catch (e: any) {
    piVersionError.value = typeof e === 'string' ? e : (e?.message || '未检测到 Pi，请确认已安装 pi')
  } finally {
    isLoadingVersion.value = false
  }
}

async function handleUpdatePi() {
  isUpdating.value = true
  updateResult.value = ''
  try {
    const a = api()
    const result = useRemote.value && sshAddress.value
      ? await a.UpdateRemotePiSelf(sshAddress.value)
      : await a.UpdatePiSelf()
    updateResult.value = result || 'Pi 已是最新版本'
  } catch (e: any) {
    updateResult.value = typeof e === 'string' ? e : (e?.message || '升级失败')
  } finally {
    isUpdating.value = false
  }
}

// --- Packages ---
function parsePiList(output: string): PiPackage[] {
  const result: PiPackage[] = []
  let currentPkg: Partial<PiPackage> | null = null
  const lines = output.split('\n')
  for (const raw of lines) {
    const text = raw.trimEnd()
    if (!text.trim()) continue

    // Normalize leading tabs to 2 spaces for consistent indent measurement
    const normalized = text.replace(/^\t+/, m => '  '.repeat(m.length))
    const trimmed = normalized.trim()
    const indent = normalized.length - normalized.trimStart().length

    // Determine indentation level (1 tab or 2 spaces = level 1)
    const level = Math.round(indent / 2)

    if (level === 0) {
      // Section header — push any in-flight package
      if (currentPkg && currentPkg.source) {
        result.push(currentPkg as PiPackage)
        currentPkg = null
      }
      continue
    }

    if (level === 1) {
      // Package name line
      if (currentPkg && currentPkg.source) {
        result.push(currentPkg as PiPackage)
      }
      currentPkg = { source: trimmed, path: '' }
    } else if (currentPkg) {
      // Level 2+ = metadata (path, etc.)
      if (!currentPkg.path) {
        currentPkg.path = trimmed
      }
    }
  }
  if (currentPkg && currentPkg.source) {
    result.push(currentPkg as PiPackage)
  }
  return result
}

async function loadPackages() {
  isLoadingPackages.value = true
  packagesError.value = ''
  packages.value = []
  try {
    const a = api()
    const raw = useRemote.value && sshAddress.value
      ? await a.GetRemotePiPackages(sshAddress.value)
      : await a.GetPiPackages()
    if (raw) {
      packages.value = parsePiList(raw)
    }
  } catch (e: any) {
    packagesError.value = typeof e === 'string' ? e : (e?.message || '获取插件列表失败')
  } finally {
    isLoadingPackages.value = false
  }
}

async function handleUpdateAll() {
  isUpdatingAll.value = true
  operationResult.value = ''
  try {
    const a = api()
    const result = useRemote.value && sshAddress.value
      ? await a.UpdateRemoteAllPiPackages(sshAddress.value)
      : await a.UpdateAllPiPackages()
    operationResult.value = result || '所有插件已更新'
    await loadPackages()
  } catch (e: any) {
    operationResult.value = typeof e === 'string' ? e : (e?.message || '升级全部插件失败')
  } finally {
    isUpdatingAll.value = false
  }
}

async function handleUpdateOne(source: string) {
  updatingPkg.value = source
  operationResult.value = ''
  try {
    const a = api()
    const result = useRemote.value && sshAddress.value
      ? await a.UpdateRemotePiPackage(sshAddress.value, source)
      : await a.UpdatePiPackage(source)
    operationResult.value = result || `${source} 已更新`
    await loadPackages()
  } catch (e: any) {
    operationResult.value = typeof e === 'string' ? e : (e?.message || `升级 ${source} 失败`)
  } finally {
    updatingPkg.value = ''
  }
}

async function handleRemove(source: string) {
  updatingPkg.value = source
  confirmRemove.value = ''
  operationResult.value = ''
  try {
    const a = api()
    const result = useRemote.value && sshAddress.value
      ? await a.RemoveRemotePiPackage(sshAddress.value, source)
      : await a.RemovePiPackage(source)
    operationResult.value = result || `${source} 已删除`
    await loadPackages()
  } catch (e: any) {
    operationResult.value = typeof e === 'string' ? e : (e?.message || `删除 ${source} 失败`)
  } finally {
    updatingPkg.value = ''
  }
}

async function handleInstall(source: string) {
  installingPkg.value = source
  installResult.value = ''
  try {
    const a = api()
    const result = useRemote.value && sshAddress.value
      ? await a.InstallRemotePiPackage(sshAddress.value, source)
      : await a.InstallPiPackage(source)
    installResult.value = result || `${source} 安装成功`
    await loadPackages()
  } catch (e: any) {
    installResult.value = typeof e === 'string' ? e : (e?.message || `安装 ${source} 失败`)
  } finally {
    installingPkg.value = ''
  }
}

onMounted(() => {
  loadSSHAddress()
  loadVersion()
  loadPackages()
})
</script>
