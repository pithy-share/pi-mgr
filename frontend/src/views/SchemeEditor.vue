<template>
  <div class="two-column">
    <!-- Left sidebar: Provider list -->
    <div class="sidebar card" style="margin:12px;">
      <h3 style="margin-bottom:12px;">供应商</h3>
      
      <!-- Built-in providers -->
      <div class="section-title">内置供应商</div>
      <div v-for="prov in builtInProviders" :key="prov.key"
        :class="['provider-item', { active: selectedProviderKey === prov.key }]"
        draggable="true"
        @dragstart="onProviderDragStart($event, prov.key)"
        @dragover.prevent="onProviderDragOver($event, prov.key)"
        @dragend="onProviderDragEnd"
        @drop.prevent="onProviderDrop($event, prov.key)"
        @click="selectedProviderKey = prov.key">
        <span class="drag-handle">⋮⋮</span>
        <span class="provider-name">{{ prov.name || prov.key }}</span>
        <span class="tag tag-builtin">内置</span>
      </div>

      <!-- Custom providers -->
      <div class="section-title">自定义供应商</div>
      <div v-for="prov in customProviders" :key="prov.key"
        :class="['provider-item', { active: selectedProviderKey === prov.key }]"
        draggable="true"
        @dragstart="onProviderDragStart($event, prov.key)"
        @dragover.prevent="onProviderDragOver($event, prov.key)"
        @dragend="onProviderDragEnd"
        @drop.prevent="onProviderDrop($event, prov.key)"
        @click="selectedProviderKey = prov.key">
        <span class="drag-handle">⋮⋮</span>
        <span class="provider-name">{{ prov.name || prov.key }}</span>
        <span class="tag tag-custom">自定义</span>
      </div>

      <div style="margin-top:16px;display:flex;flex-direction:column;gap:6px;">
        <button class="btn-primary btn-small" @click="showAddBuiltIn = !showAddBuiltIn">+ 添加内置供应商</button>
        <button class="btn-secondary btn-small" @click="showAddCustom = !showAddCustom">+ 添加自定义供应商</button>
      </div>

      <!-- Add built-in dropdown -->
      <div v-if="showAddBuiltIn" class="card" style="margin-top:8px;padding:12px;">
        <select v-model="addBuiltInKey" style="margin-bottom:8px;">
          <option value="">-- 选择内置供应商 --</option>
          <option v-for="b in sortedAvailableBuiltIns" :key="b.key" :value="b.key">{{ b.name }}</option>
        </select>
        <div v-if="addBuiltInKey" style="font-size:12px;color:var(--text-secondary);margin-bottom:8px;">
          API 类型：{{ getBuiltInAPIType(addBuiltInKey) }}
        </div>
        <div class="form-group" v-if="addBuiltInKey">
          <label>API Key</label>
          <div class="password-wrapper">
            <input :type="showBuiltInKey ? 'text' : 'password'" v-model="addBuiltInAPIKey" placeholder="输入 API key" />
            <button class="toggle-password" @click="showBuiltInKey = !showBuiltInKey">
              {{ showBuiltInKey ? '隐藏' : '显示' }}
            </button>
          </div>
        </div>
        <div class="form-group" v-if="addBuiltInKey">
          <label>Base URL（可选，用于覆盖/代理）</label>
          <input v-model="addBuiltInBaseURL" placeholder="https://your-proxy.example.com/v1" />
        </div>
        <div v-if="addBuiltInError" class="field-error">{{ addBuiltInError }}</div>
        <div style="display:flex;gap:6px;" v-if="addBuiltInKey">
          <button class="btn-primary btn-small" @click="handleAddBuiltIn">添加</button>
          <button class="btn-secondary btn-small" @click="showAddBuiltIn = false; resetAddBuiltIn()">取消</button>
        </div>
      </div>

      <!-- Add custom form -->
      <div v-if="showAddCustom" class="card" style="margin-top:8px;padding:12px;">
        <div class="form-group">
          <label>供应商标识 *</label>
          <input v-model="addCustomKey" placeholder="如 my-proxy" />
        </div>
        <div class="form-group">
          <label>Base URL *</label>
          <input v-model="addCustomBaseURL" placeholder="https://my-proxy.example.com/v1" />
        </div>
        <div class="form-group">
          <label>API 类型 *</label>
          <select v-model="addCustomAPIType">
            <option value="">-- 选择 API 类型 --</option>
            <option v-for="t in apiTypes" :key="t" :value="t">{{ t }}</option>
          </select>
        </div>
        <div class="form-group">
          <label>API Key（可选）</label>
          <div class="password-wrapper">
            <input :type="showCustomKey ? 'text' : 'password'" v-model="addCustomAPIKey" placeholder="输入 API key" />
            <button class="toggle-password" @click="showCustomKey = !showCustomKey">
              {{ showCustomKey ? '隐藏' : '显示' }}
            </button>
          </div>
        </div>
        <div v-if="addCustomError" class="field-error">{{ addCustomError }}</div>
        <div style="display:flex;gap:6px;">
          <button class="btn-primary btn-small" @click="handleAddCustom">添加</button>
          <button class="btn-secondary btn-small" @click="showAddCustom = false; resetAddCustom()">取消</button>
        </div>
      </div>
    </div>

    <!-- Right main area: Provider detail + Model list -->
    <div class="main-area" style="margin:12px;">
      <div v-if="!selectedProvider" class="card empty-state" style="height:100%;">
        <h2>{{ scheme?.name || '方案' }}</h2>
        <p>从左侧选择一个供应商以查看和编辑其配置</p>
      </div>

      <div v-else>
        <!-- Provider config form -->
        <div class="card" style="margin-bottom:16px;">
          <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:12px;">
            <h3>
              {{ selectedProvider.name || selectedProvider.key }}
              <span :class="['tag', selectedProvider.builtIn ? 'tag-builtin' : 'tag-custom']">
                {{ selectedProvider.builtIn ? '内置' : '自定义' }}
              </span>
            </h3>
            <button class="btn-danger btn-small" @click="confirmRemoveProvider = true">移除供应商</button>
          </div>

          <!-- Built-in provider form -->
          <template v-if="selectedProvider.builtIn">
            <div class="form-group">
              <label>API Key</label>
              <div class="password-wrapper">
                <input :type="showProvKey ? 'text' : 'password'" v-model="provAPIKey" placeholder="输入 API key" />
                <button class="toggle-password" @click="showProvKey = !showProvKey">
                  {{ showProvKey ? '隐藏' : '显示' }}
                </button>
              </div>
            </div>
            <div class="form-group">
              <label>Base URL（可选，用于覆盖/代理）</label>
              <input v-model="provBaseURL" placeholder="https://your-proxy.example.com/v1" />
            </div>
          </template>

          <!-- Custom provider form -->
          <template v-else>
            <div class="form-group">
              <label>供应商标识</label>
              <input :value="selectedProvider.key" disabled />
            </div>
            <div class="form-group">
              <label>Base URL *</label>
              <input v-model="provBaseURL" placeholder="https://my-proxy.example.com/v1" />
            </div>
            <div class="form-group">
              <label>API 类型 *</label>
              <select v-model="provAPIType">
                <option value="">-- 选择 API 类型 --</option>
                <option v-for="t in apiTypes" :key="t" :value="t">{{ t }}</option>
              </select>
            </div>
            <div class="form-group">
              <label>API Key（可选）</label>
              <div class="password-wrapper">
                <input :type="showProvKey ? 'text' : 'password'" v-model="provAPIKey" placeholder="输入 API key" />
                <button class="toggle-password" @click="showProvKey = !showProvKey">
                  {{ showProvKey ? '隐藏' : '显示' }}
                </button>
              </div>
            </div>
          </template>

          <div v-if="provError" class="field-error" style="margin-bottom:8px;">{{ provError }}</div>
          <div style="display:flex;gap:6px;align-items:center;">
            <button class="btn-primary" @click="handleSaveProvider">保存供应商配置</button>
            <button class="btn-secondary btn-small" @click="handleTestConnectivity"
              :disabled="!canTestConnectivity || testingConnectivity"
              :title="connectivityTooltip">
              {{ testingConnectivity ? '测试中...' : '测试连接' }}
            </button>
            <span v-if="connectivityResult" :class="connectivityResultType === 'success' ? 'toast-success-inline' : 'toast-error-inline'">
              {{ connectivityResult }}
            </span>
          </div>
        </div>

        <!-- Model list -->
        <div class="card">
          <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:8px;">
            <h3>模型列表</h3>
            <div style="display:flex;align-items:center;gap:6px;">
              <button class="btn-secondary btn-small" @click="handleFetchModels"
                :disabled="!canFetchModels" :title="fetchButtonTitle">
                ⟳ 拉取模型列表
              </button>
              <span v-if="fetchingModels" class="spinner"></span>
              <button class="btn-primary btn-small" @click="showAddModel = true">+ 添加模型</button>
            </div>
          </div>

          <!-- Search and filter bar -->
          <div style="display:flex;gap:8px;margin-bottom:12px;">
            <input v-model="modelSearchQuery" placeholder="搜索模型 ID 或名称"
              style="flex:1;" />
            <select v-model="modelCapFilter" style="width:140px;">
              <option value="all">全部</option>
              <option value="reasoning">reasoning</option>
              <option value="inputImage">inputImage</option>
            </select>
          </div>

          <!-- Batch action bar -->
          <div v-if="selectedModelIDs.length > 0" style="display:flex;align-items:center;gap:8px;margin-bottom:8px;">
            <button class="btn-danger btn-small" @click="confirmBatchDelete = true">删除所选（{{ selectedModelIDs.length }}）</button>
            <span style="font-size:12px;color:var(--text-secondary);">已选 {{ selectedModelIDs.length }} 个模型</span>
          </div>
          <div style="margin-bottom:8px;" v-if="selectedProvider.models.length > 0">
            <button class="btn-secondary btn-small" @click="showBatchImport = true">批量导入 JSON</button>
          </div>

          <!-- Empty states -->
          <div v-if="selectedProvider.models.length === 0" style="color:var(--text-secondary);padding:12px 0;">
            暂无自定义模型
          </div>
          <div v-else-if="filteredModels.length === 0 && modelSearchQuery" style="color:var(--text-secondary);padding:12px 0;">
            无匹配模型
          </div>

          <!-- Model list with checkboxes and drag -->
          <div v-for="(model, idx) in filteredModels" :key="model.id"
            class="list-item" style="margin-bottom:4px;align-items:center;justify-content:flex-start;"
            draggable="true"
            @dragstart="onModelDragStart($event, model.id)"
            @dragover.prevent="onModelDragOver($event, model.id)"
            @dragend="onModelDragEnd"
            @drop.prevent="onModelDrop($event, model.id)">
            <input type="checkbox" :value="model.id" v-model="selectedModelIDs" style="flex-shrink:0;margin-right:4px;width:16px;" />
            <span class="drag-handle" style="margin-right:4px;">⋮⋮</span>
            <div style="flex:1;min-width:0;overflow:hidden;white-space:nowrap;text-overflow:ellipsis;">
              <strong>{{ model.id }}</strong>
              <span v-if="model.name && model.name !== model.id" style="color:var(--text-secondary);margin-left:4px;font-size:13px;">{{ model.name }}</span>
              <span style="font-size:11px;color:var(--text-secondary);margin-left:6px;">ctx:{{ model.contextWindow }}&nbsp;tok:{{ model.maxTokens }}</span>
            </div>
            <div class="list-item-actions" style="flex-shrink:0;">
              <button class="btn-secondary btn-small" @click="startEditModel(model)">编辑</button>
              <button class="btn-danger btn-small" @click="confirmDeleteModel = model.id">删除</button>
            </div>
          </div>

          <!-- Select all checkbox (at bottom for quick toggle) -->
          <div v-if="filteredModels.length > 0" style="margin-top:8px;padding-top:8px;border-top:1px solid var(--border-color);">
            <label style="cursor:pointer;font-size:13px;">
              <input type="checkbox" :checked="allFilteredSelected" @change="toggleSelectAllFiltered" style="margin-right:4px;width:auto;" />
              全选 ({{ filteredModels.length }} 个可见模型)
            </label>
          </div>
        </div>

        <!-- Add/Edit Model modal -->
        <div v-if="showAddModel || editingModel" class="modal-overlay" @click.self="closeModelForm">
          <div class="modal" style="max-height:80vh;overflow-y:auto;">
            <h3>{{ editingModel ? '编辑模型' : '添加模型' }}</h3>
            <!-- Preset dropdown: only in add mode -->
            <div class="form-group" v-if="!editingModel">
              <label>预设模型（可选）</label>
              <select v-model="selectedPresetLabel" @change="applyPreset">
                <option value="">自定义</option>
                <option v-for="p in MODEL_PRESETS" :key="p.label" :value="p.label">{{ p.label }}</option>
              </select>
            </div>
            <div class="form-group">
              <label>ID *</label>
              <input v-model="modelForm.id" placeholder="模型标识符" :disabled="!!editingModel" />
              <div v-if="modelFormErrors.id" class="field-error">{{ modelFormErrors.id }}</div>
            </div>
            <div class="form-group">
              <label>名称（默认同 ID）</label>
              <input v-model="modelForm.name" placeholder="显示名称" />
            </div>
            <div class="form-group">
              <label>推理模式</label>
              <div class="checkbox-group">
                <label><input type="checkbox" v-model="modelForm.reasoning" /> reasoning</label>
              </div>
            </div>
            <div class="form-group">
              <label>输入类型</label>
              <div class="checkbox-group">
                <label><input type="checkbox" v-model="modelForm.inputText" /> 文本</label>
                <label><input type="checkbox" v-model="modelForm.inputImage" /> 图片</label>
              </div>
            </div>
            <div class="form-group">
              <label>Context Window</label>
              <input type="number" v-model.number="modelForm.contextWindow" />
            </div>
            <div class="form-group">
              <label>Max Tokens</label>
              <input type="number" v-model.number="modelForm.maxTokens" />
            </div>
            <fieldset style="border:1px solid var(--border-color);border-radius:var(--radius);padding:12px;margin-bottom:16px;">
              <legend style="font-size:13px;color:var(--text-secondary);">成本 (Cost)</legend>
              <div class="cost-grid">
                <div class="form-group">
                  <label>Input</label>
                  <input type="number" step="0.0001" v-model.number="modelForm.costInput" />
                </div>
                <div class="form-group">
                  <label>Output</label>
                  <input type="number" step="0.0001" v-model.number="modelForm.costOutput" />
                </div>
                <div class="form-group">
                  <label>Cache Read</label>
                  <input type="number" step="0.0001" v-model.number="modelForm.costCacheRead" />
                </div>
                <div class="form-group">
                  <label>Cache Write</label>
                  <input type="number" step="0.0001" v-model.number="modelForm.costCacheWrite" />
                </div>
              </div>
            </fieldset>
            <div v-if="modelFormErrors.server" class="field-error" style="margin-bottom:8px;">{{ modelFormErrors.server }}</div>
            <div class="modal-actions">
              <button class="btn-secondary" @click="closeModelForm">取消</button>
              <button class="btn-primary" @click="handleSaveModel">
                {{ editingModel ? '保存' : '添加' }}
              </button>
            </div>
          </div>
        </div>

        <!-- Fetch models dialog -->
        <div v-if="showFetchDialog" class="modal-overlay" @click.self="showFetchDialog = false">
          <div class="modal" style="max-height:70vh;overflow-y:auto;">
            <h3>选择要导入的模型</h3>
            <p style="color:var(--text-secondary);margin-bottom:12px;">
              共 {{ fetchedModels.length }} 个模型，请勾选要导入的条目
            </p>
            <label class="fetch-model-row" style="cursor:pointer;margin-bottom:4px;">
              <input type="checkbox" v-model="fetchSelectAll" @change="handleFetchSelectAll" class="fetch-model-checkbox" />
              <strong>全选</strong>
            </label>
            <div v-for="m in fetchedModels" :key="m.id" class="fetch-model-row">
              <input type="checkbox" :value="m.id" v-model="fetchSelectedIds" class="fetch-model-checkbox" />
              <span class="fetch-model-id"><strong>{{ m.id }}</strong></span>
              <span v-if="m.name && m.name !== m.id" class="fetch-model-name">({{ m.name }})</span>
            </div>
            <div style="margin-top:12px;display:flex;gap:6px;justify-content:flex-end;">
              <button class="btn-secondary" @click="showFetchDialog = false">取消</button>
              <button class="btn-primary" @click="handleImportModels" :disabled="fetchSelectedIds.length === 0">导入 ({{ fetchSelectedIds.length }})</button>
            </div>
          </div>
        </div>

        <!-- Delete model confirmation (single) -->
        <div v-if="confirmDeleteModel" class="modal-overlay" @click.self="confirmDeleteModel = ''">
          <div class="modal">
            <h3>确认删除模型</h3>
            <p>确定要删除模型「{{ confirmDeleteModel }}」吗？</p>
            <div class="modal-actions">
              <button class="btn-secondary" @click="confirmDeleteModel = ''">取消</button>
              <button class="btn-danger" @click="handleDeleteModel">确认删除</button>
            </div>
          </div>
        </div>

        <!-- Batch delete confirmation -->
        <div v-if="confirmBatchDelete" class="modal-overlay" @click.self="confirmBatchDelete = false">
          <div class="modal">
            <h3>确认批量删除模型</h3>
            <p>确认删除选中的 {{ selectedModelIDs.length }} 个模型？</p>
            <div class="modal-actions">
              <button class="btn-secondary" @click="confirmBatchDelete = false">取消</button>
              <button class="btn-danger" @click="handleBatchDelete">确认删除</button>
            </div>
          </div>
        </div>

        <!-- Batch import JSON modal -->
        <div v-if="showBatchImport" class="modal-overlay" @click.self="showBatchImport = false">
          <div class="modal" style="max-height:70vh;">
            <h3>批量导入 JSON</h3>
            <p style="color:var(--text-secondary);margin-bottom:8px;">
              粘贴 JSON 模型数组：<code>[{"id": "...", "name": "...", ...}, ...]</code>
            </p>
            <textarea v-model="batchImportJSON" rows="10"
              placeholder='[{"id": "gpt-4o", "name": "GPT-4o", "contextWindow": 128000, ...}]'
              style="width:100%;font-family:monospace;font-size:13px;"></textarea>
            <div v-if="batchImportError" class="field-error" style="margin-top:8px;">{{ batchImportError }}</div>
            <div v-if="batchImportResult" :class="batchImportResultType" style="margin-top:8px;font-size:13px;">
              {{ batchImportResult }}
            </div>
            <div class="modal-actions" style="margin-top:12px;">
              <button class="btn-secondary" @click="showBatchImport = false; resetBatchImport()">取消</button>
              <button class="btn-primary" @click="handleBatchImportSubmit" :disabled="!batchImportJSON.trim()">导入</button>
            </div>
          </div>
        </div>

        <!-- Remove provider confirmation -->
        <div v-if="confirmRemoveProvider" class="modal-overlay" @click.self="confirmRemoveProvider = false">
          <div class="modal">
            <h3>确认移除供应商</h3>
            <p>确定要移除供应商「{{ selectedProvider?.name || selectedProvider?.key }}」及其所有模型吗？</p>
            <div class="modal-actions">
              <button class="btn-secondary" @click="confirmRemoveProvider = false">取消</button>
              <button class="btn-danger" @click="handleRemoveProvider">确认移除</button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch, inject } from 'vue'
import { useRoute } from 'vue-router'
import api from '../wails/api'
import type { Scheme, Provider, Model, BuiltInProvider } from '../types'
import { defaultModel } from '../types'
import { MODEL_PRESETS, type ModelPreset } from '../presets'

const props = defineProps<{ id: string }>()
const showToast: any = inject('showToast')
const route = useRoute()

const scheme = ref<Scheme | null>(null)
const selectedProviderKey = ref('')
const apiTypes = ref<string[]>([])
const allBuiltIns = ref<BuiltInProvider[]>([])

// Add built-in state
const showAddBuiltIn = ref(false)
const addBuiltInKey = ref('')
const addBuiltInAPIKey = ref('')
const addBuiltInBaseURL = ref('')
const addBuiltInError = ref('')
const showBuiltInKey = ref(false)

// Add custom state
const showAddCustom = ref(false)
const addCustomKey = ref('')
const addCustomBaseURL = ref('')
const addCustomAPIType = ref('')
const addCustomAPIKey = ref('')
const addCustomError = ref('')
const showCustomKey = ref(false)

// Provider form state
const provAPIKey = ref('')
const provBaseURL = ref('')
const provAPIType = ref('')
const provError = ref('')
const showProvKey = ref(false)

// Model form
const showAddModel = ref(false)
const editingModel = ref<Model | null>(null)
const confirmDeleteModel = ref('')
const confirmBatchDelete = ref(false)
const confirmRemoveProvider = ref(false)
const modelForm = reactive(defaultModel())
const modelFormErrors = reactive({ id: '', server: '' })
const selectedPresetLabel = ref('')

// Fetch models state
const fetchingModels = ref(false)
const fetchedModels = ref<Model[]>([])
const showFetchDialog = ref(false)
const fetchError = ref('')
const fetchSelectAll = ref(false)
const fetchSelectedIds = ref<string[]>([])

// Search and filter state
const modelSearchQuery = ref('')
const modelCapFilter = ref<'all' | 'reasoning' | 'inputImage'>('all')

// Multi-select state
const selectedModelIDs = ref<string[]>([])

// Batch import state
const showBatchImport = ref(false)
const batchImportJSON = ref('')
const batchImportError = ref('')
const batchImportResult = ref('')
const batchImportResultType = ref<'toast-success-inline' | 'toast-error-inline'>('toast-success-inline')

// Connectivity test state
const testingConnectivity = ref(false)
const connectivityResult = ref('')
const connectivityResultType = ref<'success' | 'error'>('success')

// Drag state
const providerDragKey = ref('')
const providerDragOverKey = ref('')
const modelDragID = ref('')
const modelDragOverID = ref('')

// Computed
const builtInProviders = computed(() => scheme.value?.providers.filter(p => p.builtIn) || [])
const customProviders = computed(() => scheme.value?.providers.filter(p => !p.builtIn) || [])
const selectedProvider = computed(() => {
  if (!scheme.value || !selectedProviderKey.value) return null
  return scheme.value.providers.find(p => p.key === selectedProviderKey.value) || null
})

const availableBuiltIns = computed(() => {
  const existing = new Set((scheme.value?.providers || []).map(p => p.key))
  return allBuiltIns.value.filter(b => !existing.has(b.key))
})

// AC-04: sorted by name
const sortedAvailableBuiltIns = computed(() => {
  return [...availableBuiltIns.value].sort((a, b) => a.name.localeCompare(b.name))
})

function getBuiltInAPIType(key: string): string {
  return allBuiltIns.value.find(b => b.key === key)?.apiType || ''
}

function getEffectiveAPIType(prov: Provider | null): string {
  if (!prov) return ''
  if (prov.builtIn) {
    return getBuiltInAPIType(prov.key)
  }
  return prov.apiType
}

// Model search and filter
const filteredModels = computed(() => {
  if (!selectedProvider.value) return []
  let models = selectedProvider.value.models || []
  const q = modelSearchQuery.value.toLowerCase().trim()
  const cap = modelCapFilter.value

  if (q) {
    models = models.filter(m =>
      m.id.toLowerCase().includes(q) || m.name.toLowerCase().includes(q)
    )
  }
  if (cap === 'reasoning') {
    models = models.filter(m => m.reasoning === true)
  } else if (cap === 'inputImage') {
    models = models.filter(m => m.inputImage === true)
  }
  return models
})

const allFilteredSelected = computed(() => {
  if (filteredModels.value.length === 0) return false
  return filteredModels.value.every(m => selectedModelIDs.value.includes(m.id))
})

function toggleSelectAllFiltered() {
  if (allFilteredSelected.value) {
    const filteredIDs = new Set(filteredModels.value.map(m => m.id))
    selectedModelIDs.value = selectedModelIDs.value.filter(id => !filteredIDs.has(id))
  } else {
    const filteredIDs = filteredModels.value.map(m => m.id)
    const existing = new Set(selectedModelIDs.value)
    for (const id of filteredIDs) {
      if (!existing.has(id)) {
        selectedModelIDs.value.push(id)
      }
    }
  }
}

// Fetch models computed
const canFetchModels = computed(() => {
  const prov = selectedProvider.value
  if (!prov) return false
  const apiType = getEffectiveAPIType(prov)
  if (apiType !== 'openai-completions' && apiType !== 'openai-responses' && apiType !== 'azure-openai-responses') return false
  if (!prov.baseUrl) return false
  return true
})

const fetchButtonTitle = computed(() => {
  const prov = selectedProvider.value
  if (!prov) return '请先选择供应商'
  const apiType = getEffectiveAPIType(prov)
  if (apiType !== 'openai-completions' && apiType !== 'openai-responses' && apiType !== 'azure-openai-responses') {
    return '该 API 类型不支持自动拉取'
  }
  if (!prov.baseUrl) return '请先配置 baseUrl'
  return ''
})

// Connectivity test computed
const canTestConnectivity = computed(() => {
  const prov = selectedProvider.value
  if (!prov) return false
  const apiType = getEffectiveAPIType(prov)
  if (apiType !== 'openai-completions' && apiType !== 'openai-responses' && apiType !== 'azure-openai-responses') return false
  if (!provBaseURL.value.trim()) return false
  return true
})

const connectivityTooltip = computed(() => {
  const prov = selectedProvider.value
  if (!prov) return ''
  const apiType = getEffectiveAPIType(prov)
  if (apiType !== 'openai-completions' && apiType !== 'openai-responses' && apiType !== 'azure-openai-responses') {
    return '该 API 类型暂不支持连通性测试'
  }
  if (!provBaseURL.value.trim()) return '请先配置 Base URL'
  return ''
})

// Preset application
function applyPreset() {
  if (!selectedPresetLabel.value) return
  const preset = MODEL_PRESETS.find(p => p.label === selectedPresetLabel.value)
  if (!preset) return
  modelForm.id = preset.model.id
  modelForm.name = preset.model.name
  modelForm.reasoning = preset.model.reasoning
  modelForm.inputText = preset.model.inputText
  modelForm.inputImage = preset.model.inputImage
  modelForm.contextWindow = preset.model.contextWindow
  modelForm.maxTokens = preset.model.maxTokens
  // Cost fields remain 0 (AC-36)
}

onMounted(async () => {
  await loadData()
})

watch(() => route.params.id, async (newId) => {
  if (newId) await loadData()
})

// Clear search/filter/selection when provider changes
watch(selectedProviderKey, () => {
  modelSearchQuery.value = ''
  modelCapFilter.value = 'all'
  selectedModelIDs.value = []
  connectivityResult.value = ''
})

async function loadData() {
  try {
    const a = api()
    const schemes = await a.ListSchemes()
    scheme.value = schemes.find(s => s.id === props.id) || null
    apiTypes.value = await a.ListAPITypes()
    allBuiltIns.value = await a.ListBuiltInProviders()
    if (scheme.value && scheme.value.providers.length > 0 && !selectedProviderKey.value) {
      selectedProviderKey.value = scheme.value.providers[0].key
    }
  } catch (e: any) {
    showToast?.(e?.message || e, 'error')
  }
}

// Sync selected provider to form
watch(selectedProvider, (prov) => {
  if (prov) {
    provAPIKey.value = prov.apiKey
    provBaseURL.value = prov.baseUrl
    provAPIType.value = prov.apiType
    provError.value = ''
  }
})

// ---- Provider drag ----
function onProviderDragStart(e: DragEvent, key: string) {
  providerDragKey.value = key
}

function onProviderDragOver(e: DragEvent, key: string) {
  providerDragOverKey.value = key
}

function onProviderDragEnd() {
  providerDragKey.value = ''
  providerDragOverKey.value = ''
}

async function onProviderDrop(e: DragEvent, targetKey: string) {
  const srcKey = providerDragKey.value
  if (!srcKey || srcKey === targetKey || !scheme.value) return

  const providers = [...scheme.value.providers]
  const srcIdx = providers.findIndex(p => p.key === srcKey)
  const tgtIdx = providers.findIndex(p => p.key === targetKey)
  if (srcIdx < 0 || tgtIdx < 0) return

  // Reorder
  const [moved] = providers.splice(srcIdx, 1)
  providers.splice(tgtIdx, 0, moved)

  const orderedKeys = providers.map(p => p.key)
  try {
    const a = api()
    await a.ReorderProviders(props.id, orderedKeys)
    await loadData()
    showToast?.('供应商顺序已更新', 'success')
  } catch (e: any) {
    showToast?.(e?.message || e, 'error')
  }
}

// ---- Model drag ----
function onModelDragStart(e: DragEvent, id: string) {
  modelDragID.value = id
}

function onModelDragOver(e: DragEvent, id: string) {
  modelDragOverID.value = id
}

function onModelDragEnd() {
  modelDragID.value = ''
  modelDragOverID.value = ''
}

async function onModelDrop(e: DragEvent, targetID: string) {
  const srcID = modelDragID.value
  if (!srcID || srcID === targetID || !selectedProvider.value) return

  const models = [...selectedProvider.value.models]
  const srcIdx = models.findIndex(m => m.id === srcID)
  const tgtIdx = models.findIndex(m => m.id === targetID)
  if (srcIdx < 0 || tgtIdx < 0) return

  // Reorder
  const [moved] = models.splice(srcIdx, 1)
  models.splice(tgtIdx, 0, moved)

  const orderedIDs = models.map(m => m.id)
  try {
    const a = api()
    await a.ReorderModels(props.id, selectedProviderKey.value, orderedIDs)
    await loadData()
    showToast?.('模型顺序已更新', 'success')
  } catch (e: any) {
    showToast?.(e?.message || e, 'error')
  }
}

// ---- Built-in add ----
function resetAddBuiltIn() {
  addBuiltInKey.value = ''
  addBuiltInAPIKey.value = ''
  addBuiltInBaseURL.value = ''
  addBuiltInError.value = ''
  showBuiltInKey.value = false
}

async function handleAddBuiltIn() {
  if (!addBuiltInKey.value) return
  try {
    const a = api()
    await a.AddBuiltInProvider(props.id, addBuiltInKey.value, addBuiltInAPIKey.value, addBuiltInBaseURL.value)
    showAddBuiltIn.value = false
    resetAddBuiltIn()
    await loadData()
    selectedProviderKey.value = addBuiltInKey.value
    showToast?.('内置供应商已添加', 'success')
  } catch (e: any) {
    addBuiltInError.value = e?.message || e
  }
}

// ---- Custom add ----
function resetAddCustom() {
  addCustomKey.value = ''
  addCustomBaseURL.value = ''
  addCustomAPIType.value = ''
  addCustomAPIKey.value = ''
  addCustomError.value = ''
  showCustomKey.value = false
}

async function handleAddCustom() {
  try {
    const a = api()
    await a.AddCustomProvider(props.id, addCustomKey.value, addCustomBaseURL.value, addCustomAPIType.value, addCustomAPIKey.value)
    showAddCustom.value = false
    resetAddCustom()
    await loadData()
    selectedProviderKey.value = addCustomKey.value
    showToast?.('自定义供应商已添加', 'success')
  } catch (e: any) {
    addCustomError.value = e?.message || e
  }
}

// ---- Provider save ----
async function handleSaveProvider() {
  if (!selectedProvider.value) return
  try {
    const a = api()
    const updated: Provider = {
      ...selectedProvider.value,
      apiKey: provAPIKey.value,
      baseUrl: provBaseURL.value,
      apiType: provAPIType.value,
    }
    await a.UpdateProvider(props.id, updated)
    provError.value = ''
    await loadData()
    showToast?.('供应商配置已保存', 'success')
  } catch (e: any) {
    provError.value = e?.message || e
  }
}

async function handleRemoveProvider() {
  if (!selectedProvider.value) return
  try {
    const a = api()
    await a.RemoveProvider(props.id, selectedProvider.value.key)
    confirmRemoveProvider.value = false
    selectedProviderKey.value = ''
    await loadData()
    showToast?.('供应商已移除', 'success')
  } catch (e: any) {
    showToast?.(e?.message || e, 'error')
  }
}

// ---- Connectivity test ----
async function handleTestConnectivity() {
  if (!selectedProvider.value || testingConnectivity.value) return
  testingConnectivity.value = true
  connectivityResult.value = ''
  try {
    const a = api()
    const msg = await a.TestProviderConnectivity(props.id, selectedProvider.value.key)
    connectivityResult.value = msg
    connectivityResultType.value = msg.includes('成功') ? 'success' : 'error'
  } catch (e: any) {
    connectivityResult.value = e?.message || e
    connectivityResultType.value = 'error'
  } finally {
    testingConnectivity.value = false
  }
}

// ---- Model ----
function startEditModel(m: Model) {
  editingModel.value = m
  Object.assign(modelForm, { ...m })
  selectedPresetLabel.value = ''
  modelFormErrors.id = ''
  modelFormErrors.server = ''
}

function closeModelForm() {
  showAddModel.value = false
  editingModel.value = null
  selectedPresetLabel.value = ''
  Object.assign(modelForm, defaultModel())
  modelFormErrors.id = ''
  modelFormErrors.server = ''
}

async function handleSaveModel() {
  if (!modelForm.id.trim()) {
    modelFormErrors.id = '模型 ID 不能为空'
    return
  }
  if (!editingModel.value && selectedProvider.value) {
    const exists = selectedProvider.value.models.some(m => m.id === modelForm.id.trim())
    if (exists) {
      modelFormErrors.id = '模型 ID 在该供应商下已存在'
      return
    }
  }

  try {
    const a = api()
    const m: Model = {
      id: modelForm.id.trim(),
      name: modelForm.name.trim() || modelForm.id.trim(),
      reasoning: modelForm.reasoning,
      inputText: modelForm.inputText,
      inputImage: modelForm.inputImage,
      contextWindow: modelForm.contextWindow,
      maxTokens: modelForm.maxTokens,
      costInput: modelForm.costInput,
      costOutput: modelForm.costOutput,
      costCacheRead: modelForm.costCacheRead,
      costCacheWrite: modelForm.costCacheWrite,
    }
    if (editingModel.value) {
      await a.UpdateModel(props.id, selectedProviderKey.value, m)
    } else {
      await a.AddModel(props.id, selectedProviderKey.value, m)
    }
    closeModelForm()
    await loadData()
    showToast?.(editingModel.value ? '模型已更新' : '模型已添加', 'success')
  } catch (e: any) {
    modelFormErrors.server = e?.message || e
  }
}

async function handleDeleteModel() {
  const id = confirmDeleteModel.value
  if (!id || !selectedProvider.value) return
  try {
    const a = api()
    await a.RemoveModel(props.id, selectedProvider.value.key, id)
    confirmDeleteModel.value = ''
    await loadData()
    showToast?.('模型已删除', 'success')
  } catch (e: any) {
    showToast?.(e?.message || e, 'error')
  }
}

// ---- Batch delete ----
async function handleBatchDelete() {
  const ids = selectedModelIDs.value
  if (ids.length === 0 || !selectedProvider.value) return
  try {
    const a = api()
    const removed = await a.RemoveModels(props.id, selectedProvider.value.key, ids)
    confirmBatchDelete.value = false
    selectedModelIDs.value = []
    await loadData()
    showToast?.(`已删除 ${removed} 个模型`, 'success')
  } catch (e: any) {
    showToast?.(e?.message || e, 'error')
  }
}

// ---- Batch import JSON ----
function resetBatchImport() {
  batchImportJSON.value = ''
  batchImportError.value = ''
  batchImportResult.value = ''
}

async function handleBatchImportSubmit() {
  if (!batchImportJSON.value.trim() || !selectedProvider.value) return
  batchImportError.value = ''
  batchImportResult.value = ''

  let parsed: any
  try {
    parsed = JSON.parse(batchImportJSON.value)
  } catch {
    batchImportError.value = 'JSON 格式错误'
    return
  }

  if (!Array.isArray(parsed)) {
    batchImportError.value = 'JSON 格式错误：应为数组'
    return
  }

  // Validate at least id field per item
  const models: Model[] = []
  for (const item of parsed) {
    if (!item.id || typeof item.id !== 'string') {
      batchImportError.value = 'JSON 格式错误：每项必须包含 id 字段'
      return
    }
    const m: Model = {
      id: item.id,
      name: item.name || item.id,
      reasoning: !!item.reasoning,
      inputText: item.inputText !== undefined ? !!item.inputText : true,
      inputImage: !!item.inputImage,
      contextWindow: typeof item.contextWindow === 'number' ? item.contextWindow : 256000,
      maxTokens: typeof item.maxTokens === 'number' ? item.maxTokens : 64000,
      costInput: typeof item.costInput === 'number' ? item.costInput : 0,
      costOutput: typeof item.costOutput === 'number' ? item.costOutput : 0,
      costCacheRead: typeof item.costCacheRead === 'number' ? item.costCacheRead : 0,
      costCacheWrite: typeof item.costCacheWrite === 'number' ? item.costCacheWrite : 0,
    }
    models.push(m)
  }

  try {
    const a = api()
    const added = await a.ImportProviderModels(props.id, selectedProviderKey.value, models)
    const skipped = models.length - added
    if (added > 0) {
      batchImportResult.value = `成功导入 ${added} 个模型，跳过 ${skipped} 个已存在模型`
      batchImportResultType.value = 'toast-success-inline'
    } else {
      batchImportResult.value = '无新增模型（所有模型均已存在）'
      batchImportResultType.value = 'toast-error-inline'
    }
    await loadData()
  } catch (e: any) {
    batchImportError.value = e?.message || e
  }
}

// ---- Fetch models ----
async function handleFetchModels() {
  const prov = selectedProvider.value
  if (!prov) return

  fetchingModels.value = true
  fetchError.value = ''
  try {
    const a = api()
    const result = await a.FetchProviderModels(props.id, prov.key)
    fetchedModels.value = result
    fetchSelectedIds.value = []
    fetchSelectAll.value = false
    showFetchDialog.value = true
  } catch (e: any) {
    fetchError.value = e?.message || e
    showToast?.(fetchError.value, 'error')
  } finally {
    fetchingModels.value = false
  }
}

function handleFetchSelectAll() {
  if (fetchSelectAll.value) {
    fetchSelectedIds.value = fetchedModels.value.map(m => m.id)
  } else {
    fetchSelectedIds.value = []
  }
}

async function handleImportModels() {
  if (fetchSelectedIds.value.length === 0) {
    showToast?.('请至少选择一个模型', 'error')
    return
  }
  try {
    const selectedModels = fetchedModels.value.filter(m => fetchSelectedIds.value.includes(m.id))
    const a = api()
    const count = await a.ImportProviderModels(props.id, selectedProviderKey.value, selectedModels)
    showFetchDialog.value = false
    fetchedModels.value = []
    fetchSelectedIds.value = []
    fetchSelectAll.value = false
    await loadData()
    if (count > 0) {
      const skipped = selectedModels.length - count
      const msg = skipped > 0 ? `已导入 ${count} 个模型，${skipped} 个因重复跳过` : `已导入 ${count} 个模型`
      showToast?.(msg, 'success')
    } else {
      showToast?.('无新增模型（所有选中模型均已存在）', 'success')
    }
  } catch (e: any) {
    showToast?.(e?.message || e, 'error')
  }
}
</script>

<style>
.drag-handle {
  cursor: grab;
  color: var(--text-secondary);
  margin-right: 6px;
  user-select: none;
  font-size: 14px;
}
.drag-handle:active {
  cursor: grabbing;
}
.provider-item[draggable="true"]:active {
  cursor: grabbing;
}
.toast-success-inline {
  color: #22c55e;
  font-size: 13px;
}
.toast-error-inline {
  color: #ef4444;
  font-size: 13px;
}
</style>