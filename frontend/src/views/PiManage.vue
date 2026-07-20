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

    <!-- Codebase Memory usage rules (click to copy) -->
    <div class="card" style="margin-bottom:16px;padding:10px 16px;">
      <div style="display:flex;align-items:center;justify-content:space-between;">
        <span style="font-size:14px;">Codebase Memory 使用规则</span>
        <button class="btn-secondary btn-small" @click="handleCopyCodebaseMemory" :disabled="loadingCodebaseMemory">
          {{ loadingCodebaseMemory ? '加载中…' : (copiedCodebaseMemory ? '已复制' : '复制到剪贴板') }}
        </button>
      </div>
      <div v-if="codebaseMemoryError" style="margin-top:6px;font-size:12px;color:var(--danger);">{{ codebaseMemoryError }}</div>
    </div>

    <!-- Codebase Memory MCP config (click to copy) -->
    <div class="card" style="margin-bottom:16px;padding:10px 16px;">
      <div style="display:flex;align-items:center;justify-content:space-between;">
        <span style="font-size:14px;">Codebase Memory MCP 配置</span>
        <span style="font-size:11px;color:var(--text-secondary);">.mcp.json</span>
        <button class="btn-secondary btn-small" @click="handleCopyMCPConfig" :disabled="loadingMCPConfig">
          {{ loadingMCPConfig ? '加载中…' : (copiedMCPConfig ? '已复制' : '复制到剪贴板') }}
        </button>
      </div>
      <div v-if="mcpConfigError" style="margin-top:6px;font-size:12px;color:var(--danger);">{{ mcpConfigError }}</div>
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

    <!-- Built-in Prompt Templates -->
    <div class="card" style="margin-bottom:16px;">
      <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:12px;">
        <h3>内置提示词</h3>
        <div style="display:flex;gap:6px;">
          <button class="btn-secondary btn-small" @click="loadPrompts" :disabled="isLoadingPrompts">
            刷新
          </button>
          <button class="btn-primary btn-small" @click="handleInstallAllPrompts" :disabled="isInstallingPrompts || promptList.length === 0">
            {{ isInstallingPrompts ? '安装中…' : '安装全部' }}
          </button>
        </div>
      </div>

      <div v-if="promptsError" style="color:var(--danger);margin-bottom:8px;font-size:13px;">{{ promptsError }}</div>

      <div v-for="pt in promptList" :key="pt.name"
        class="list-item" style="margin-bottom:2px;align-items:flex-start;cursor:pointer;"
        @click="handleViewPrompt(pt.name)">
        <div style="flex:1;min-width:0;">
          <div style="display:flex;align-items:center;gap:6px;flex-wrap:wrap;">
            <code style="font-size:13px;font-weight:500;">/{{ pt.name }}</code>
            <span v-if="pt.installed" class="active-badge" style="font-size:11px;">已安装</span>
          </div>
          <div v-if="pt.description" style="font-size:12px;color:var(--text-secondary);margin-top:2px;">
            {{ pt.description }}
          </div>
          <div v-if="pt.argumentHint" style="font-size:11px;color:var(--text-secondary);margin-top:1px;opacity:0.8;">
            参数: {{ pt.argumentHint }}
          </div>
        </div>
        <div class="list-item-actions" style="flex-shrink:0;margin-left:8px;">
          <button v-if="!pt.installed" class="btn-primary btn-small"
            :disabled="installingPrompt === pt.name"
            @click="handleInstallPrompt(pt.name)">
            {{ installingPrompt === pt.name ? '安装中…' : '安装' }}
          </button>
          <button v-else class="btn-danger btn-small"
            :disabled="removingPrompt === pt.name"
            @click="handleRemovePrompt(pt.name)">
            {{ removingPrompt === pt.name ? '删除中…' : '删除' }}
          </button>
        </div>
      </div>

      <div v-if="promptResult" style="margin-top:8px;padding:8px;background:#f5f5f5;border-radius:var(--radius);font-size:13px;">
        {{ promptResult }}
      </div>
      <div v-else-if="promptList.length === 0 && !isLoadingPrompts && !promptsError" style="color:var(--text-secondary);padding:8px 0;font-size:13px;">
        未加载到内置提示词
      </div>
    </div>

    <!-- Prompt preview modal -->
    <div v-if="previewPrompt" class="modal-overlay" @click.self="previewPrompt = ''">
      <div class="modal" style="min-width:500px;max-width:700px;max-height:80vh;display:flex;flex-direction:column;">
        <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:12px;">
          <h3 style="margin:0;">/<span style="font-family:monospace;">{{ previewPrompt }}</span></h3>
          <button class="btn-secondary btn-small" @click="previewPrompt = ''">关闭</button>
        </div>
        <div style="flex:1;overflow-y:auto;background:#f8f9fa;border-radius:var(--radius);padding:16px;font-size:13px;font-family:monospace;white-space:pre-wrap;word-break:break-word;line-height:1.6;">
          <div v-if="isLoadingContent" class="spinner" style="margin:20px auto;"></div>
          <div v-else-if="previewError" style="color:var(--danger);">{{ previewError }}</div>
          <template v-else>{{ previewContent }}</template>
        </div>
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
import type { PiPackage, PromptTemplate } from '../types'

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
  'npm:pi-mcp-adapter',
]
const installingPkg = ref('')
const installResult = ref('')

// Prompt templates
const promptList = ref<PromptTemplate[]>([])
const isLoadingPrompts = ref(false)
const promptsError = ref('')
const installingPrompt = ref('')
const isInstallingPrompts = ref(false)
const removingPrompt = ref('')
const promptResult = ref('')

// Prompt preview
const previewPrompt = ref('')
const previewContent = ref('')
const isLoadingContent = ref(false)
const previewError = ref('')

async function handleViewPrompt(name: string) {
  previewPrompt.value = name
  previewContent.value = ''
  previewError.value = ''
  isLoadingContent.value = true
  try {
    const a = api()
    previewContent.value = await a.GetBuiltInPromptContent(name)
  } catch (e: any) {
    previewError.value = typeof e === 'string' ? e : (e?.message || '加载失败')
  } finally {
    isLoadingContent.value = false
  }
}

async function loadPrompts() {
  isLoadingPrompts.value = true
  promptsError.value = ''
  promptList.value = []
  try {
    const a = api()
    promptList.value = await a.ListBuiltInPrompts()
  } catch (e: any) {
    promptsError.value = typeof e === 'string' ? e : (e?.message || '获取内置提示词失败')
  } finally {
    isLoadingPrompts.value = false
  }
}

async function handleInstallPrompt(name: string) {
  installingPrompt.value = name
  promptResult.value = ''
  try {
    const a = api()
    const count = await a.InstallBuiltInPrompts([name])
    promptResult.value = `已安装 /${name}`
    await loadPrompts()
  } catch (e: any) {
    promptResult.value = typeof e === 'string' ? e : (e?.message || `安装 /${name} 失败`)
  } finally {
    installingPrompt.value = ''
  }
}

async function handleInstallAllPrompts() {
  isInstallingPrompts.value = true
  promptResult.value = ''
  try {
    const a = api()
    const count = await a.InstallBuiltInPrompts([])
    promptResult.value = `已安装 ${count} 个提示词`
    await loadPrompts()
  } catch (e: any) {
    promptResult.value = typeof e === 'string' ? e : (e?.message || '安装提示词失败')
  } finally {
    isInstallingPrompts.value = false
  }
}

async function handleRemovePrompt(name: string) {
  removingPrompt.value = name
  promptResult.value = ''
  try {
    const a = api()
    await a.RemoveInstalledPrompt(name)
    promptResult.value = `已删除 /${name}`
    await loadPrompts()
  } catch (e: any) {
    promptResult.value = typeof e === 'string' ? e : (e?.message || `删除 /${name} 失败`)
  } finally {
    removingPrompt.value = ''
  }
}

// Codebase Memory rules
const loadingCodebaseMemory = ref(false)
const copiedCodebaseMemory = ref(false)
const codebaseMemoryError = ref('')
let codebaseMemoryTimer: ReturnType<typeof setTimeout> | null = null

async function handleCopyCodebaseMemory() {
  loadingCodebaseMemory.value = true
  codebaseMemoryError.value = ''
  try {
    const a = api()
    const rules = await a.GetCodebaseMemoryRules()
    await navigator.clipboard.writeText(rules)
    copiedCodebaseMemory.value = true
    if (codebaseMemoryTimer) clearTimeout(codebaseMemoryTimer)
    codebaseMemoryTimer = setTimeout(() => { copiedCodebaseMemory.value = false }, 3000)
  } catch (e: any) {
    codebaseMemoryError.value = '复制失败'
  } finally {
    loadingCodebaseMemory.value = false
  }
}

// Codebase Memory MCP config
const loadingMCPConfig = ref(false)
const copiedMCPConfig = ref(false)
const mcpConfigError = ref('')
let mcpConfigTimer: ReturnType<typeof setTimeout> | null = null

async function handleCopyMCPConfig() {
  loadingMCPConfig.value = true
  mcpConfigError.value = ''
  try {
    const a = api()
    const config = await a.GetCodebaseMemoryMCPConfig()
    await navigator.clipboard.writeText(config)
    copiedMCPConfig.value = true
    if (mcpConfigTimer) clearTimeout(mcpConfigTimer)
    mcpConfigTimer = setTimeout(() => { copiedMCPConfig.value = false }, 3000)
  } catch (e: any) {
    mcpConfigError.value = '复制失败'
  } finally {
    loadingMCPConfig.value = false
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
  loadPrompts()
})
</script>
