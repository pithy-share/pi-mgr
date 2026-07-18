<template>
  <div style="max-width:800px;margin:0 auto;padding:20px;">
    <div style="display:flex;align-items:center;justify-content:space-between;margin-bottom:20px;">
      <h2>配置方案</h2>
      <div style="display:flex;gap:8px;">
        <button class="btn-secondary" @click="handleExport">导出全部</button>
        <button class="btn-secondary" @click="handleImport">导入</button>
        <button class="btn-primary" @click="showCreate = true">+ 新建方案</button>
      </div>
    </div>

    <!-- Create inline form -->
    <div v-if="showCreate" class="card" style="margin-bottom:16px;">
      <div class="inline-form">
        <input
          v-model="newName"
          placeholder="输入方案名称"
          @keyup.enter="handleCreate"
          ref="createInput"
        />
        <button class="btn-primary" @click="handleCreate" :disabled="!newName.trim()">创建</button>
        <button class="btn-secondary" @click="showCreate = false; newName = ''">取消</button>
      </div>
      <div v-if="createError" class="field-error">{{ createError }}</div>
    </div>

    <!-- Empty state (AC-01) -->
    <div v-if="schemes.length === 0" class="card empty-state">
      <h2>暂无配置方案</h2>
      <p>点击「新建方案」创建您的第一个配置方案</p>
    </div>

    <!-- Scheme list -->
    <div v-for="scheme in schemes" :key="scheme.id" class="card list-item" style="margin-bottom:8px;">
      <div style="flex:1;">
        <span v-if="editingId !== scheme.id" style="font-weight:500;">{{ scheme.name }}</span>
        <input
          v-else
          v-model="editName"
          @keyup.enter="handleSaveEdit(scheme)"
          @keyup.escape="editingId = ''"
          style="max-width:300px;"
          ref="editInput"
        />
        <span style="color:var(--text-secondary);font-size:12px;margin-left:8px;">
          {{ scheme.providers.length }} 个供应商
        </span>
      </div>
      <div class="list-item-actions">
        <template v-if="editingId === scheme.id">
          <button class="btn-primary btn-small" @click="handleSaveEdit(scheme)" :disabled="!editName.trim()">保存</button>
          <button class="btn-secondary btn-small" @click="editingId = ''">取消</button>
        </template>
        <template v-else>
          <button class="btn-secondary btn-small" @click="startEdit(scheme)">编辑</button>
          <button class="btn-secondary btn-small" @click="handleDuplicate(scheme.id)">复制</button>
          <button class="btn-success btn-small" @click="handleActivate(scheme.id)">激活</button>
          <button class="btn-danger btn-small" @click="confirmDelete = scheme.id">删除</button>
          <button class="btn-primary btn-small" @click="$router.push(`/scheme/${scheme.id}`)">配置</button>
        </template>
      </div>
    </div>

    <!-- Delete confirmation modal (AC-05, AC-30) -->
    <div v-if="confirmDelete" class="modal-overlay" @click.self="confirmDelete = ''">
      <div class="modal">
        <h3>确认删除</h3>
        <p>确定要删除该方案吗？此操作不可撤销。</p>
        <div class="modal-actions">
          <button class="btn-secondary" @click="confirmDelete = ''">取消</button>
          <button class="btn-danger" @click="handleDelete">确认删除</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref, nextTick, inject } from 'vue'
import api from '../wails/api'
import type { Scheme } from '../types'

const showToast: any = inject('showToast')

const schemes = ref<Scheme[]>([])
const showCreate = ref(false)
const newName = ref('')
const createError = ref('')
const editingId = ref('')
const editName = ref('')
const confirmDelete = ref('')

const createInput = ref<HTMLInputElement | null>(null)
const editInput = ref<HTMLInputElement | null>(null)

onMounted(async () => {
  await loadSchemes()
})

async function loadSchemes() {
  try {
    const a = api()
    schemes.value = await a.ListSchemes()
  } catch (e: any) {
    showToast?.(e?.message || e, 'error')
  }
}

async function handleCreate() {
  const name = newName.value.trim()
  if (!name) {
    createError.value = '方案名称不能为空'
    return
  }
  try {
    const a = api()
    await a.CreateScheme(name)
    showCreate.value = false
    newName.value = ''
    createError.value = ''
    await loadSchemes()
    showToast?.('方案创建成功', 'success')
  } catch (e: any) {
    createError.value = e?.message || e
  }
}

function startEdit(scheme: Scheme) {
  editingId.value = scheme.id
  editName.value = scheme.name
  nextTick(() => {
    editInput.value?.focus()
  })
}

async function handleSaveEdit(scheme: Scheme) {
  if (!editName.value.trim()) return
  try {
    const a = api()
    await a.UpdateScheme({ ...scheme, name: editName.value.trim() })
    editingId.value = ''
    await loadSchemes()
    showToast?.('方案名称已更新', 'success')
  } catch (e: any) {
    showToast?.(e?.message || e, 'error')
  }
}

async function handleDuplicate(id: string) {
  try {
    const a = api()
    await a.DuplicateScheme(id)
    await loadSchemes()
    showToast?.('方案已复制', 'success')
  } catch (e: any) {
    showToast?.(e?.message || e, 'error')
  }
}

async function handleExport() {
  try {
    const a = api()
    await a.ExportSchemes()
    showToast?.('方案已导出', 'success')
  } catch (e: any) {
    showToast?.(e?.message || e, 'error')
  }
}

async function handleImport() {
  try {
    const a = api()
    await a.ImportSchemes()
    await loadSchemes()  // AC-08: refresh list immediately after import
    showToast?.('方案已导入', 'success')
  } catch (e: any) {
    showToast?.(e?.message || e, 'error')
  }
}

async function handleActivate(id: string) {
  try {
    const a = api()
    await a.ActivateScheme(id)
    showToast?.('方案已激活，models.json 已更新', 'success')
  } catch (e: any) {
    showToast?.(e?.message || e, 'error')
  }
}

async function handleDelete() {
  const id = confirmDelete.value
  if (!id) return
  try {
    const a = api()
    await a.DeleteScheme(id)
    confirmDelete.value = ''
    await loadSchemes()
    showToast?.('方案已删除', 'success')
  } catch (e: any) {
    showToast?.(e?.message || e, 'error')
  }
}
</script>
