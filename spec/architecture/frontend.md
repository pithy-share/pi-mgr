# 前端架构

**阅读时机**：涉及前端页面、路由、组件结构、状态管理、表单交互、拖拽排序、Toast 通知、API 调用模式、模型搜索/过滤或 CSS 样式系统时。  
**可核验依据**：`frontend/src/main.ts`（路由注册）、`frontend/src/App.vue`（Shell + Toast + provide）、`frontend/src/wails/api.ts`（Wails 桥接工厂）、`frontend/src/types.ts`（前端类型 + defaultModel）、`frontend/src/presets.ts`（MODEL_PRESETS）、`frontend/src/views/SchemeList.vue`、`frontend/src/views/SchemeEditor.vue`、`frontend/src/views/SshSync.vue`、`frontend/src/style.css`（全局样式系统）。

## 技术栈与约定

- Vue 3 Composition API（`<script setup lang="ts">`）
- Vue Router（hash mode）
- 无 Pinia / Vuex，状态通过 Vue 3 `provide`/`inject` + `ref`/`reactive` 管理
- TypeScript 类型定义在 `frontend/src/types.ts`
- 全局 CSS 在 `frontend/src/style.css`（CSS 变量体系，无第三方 UI 库）
- 前端组件平铺在 `views/` 目录，无 `components/` 子目录

## 路由表

| 路由 | 组件 | 职责 |
|---|---|---|
| `/` | `SchemeList` | 方案列表，提供新建/编辑/复制/删除/激活/导出/导入操作 |
| `/scheme/:id` | `SchemeEditor` | 方案编辑器，左侧供应商列表 + 右侧供应商详情与模型管理 |
| `/ssh-sync` | `SshSync` | SSH 连接测试与远程配置同步 |

路由注册在 `main.ts`，使用 `createWebHashHistory`。

## 组件树

```
App.vue (Shell)
├── Header (标题 + 返回按钮 + 激活方案徽标)
├── <router-view>
│   ├── SchemeList.vue       路由 /
│   │   ├── 方案卡片列表
│   │   ├── 新建方案内联表单
│   │   ├── 内联编辑（方案重命名）
│   │   └── 删除确认 Modal
│   ├── SchemeEditor.vue      路由 /scheme/:id
│   │   ├── 左侧 Sidebar
│   │   │   ├── 内置供应商列表（可拖拽排序）
│   │   │   ├── 自定义供应商列表（可拖拽排序）
│   │   │   ├── 添加内置供应商下拉表单（含密码显隐切换）
│   │   │   └── 添加自定义供应商表单（含密码显隐切换）
│   │   ├── 右侧 Main Area
│   │   │   ├── 供应商配置表单（内置/自定义表单分支）
│   │   │   │   ├── 保存 / 连通性测试 / 移除供应商
│   │   │   ├── 模型列表（可拖拽排序 + 搜索/过滤 + 多选）
│   │   │   │   ├── 拉取模型列表（Fetch → 选择导入 Dialog）
│   │   │   │   ├── 批量删除确认 Modal
│   │   │   │   └── 批量导入 JSON Modal
│   │   │   └── 添加/编辑模型 Modal（预设下拉仅在添加模式显示）
│   │   └── 移除供应商确认 Modal
│   └── SshSync.vue            路由 /ssh-sync
│       ├── SSH 地址输入（自动加载/保存）
│       ├── 连接测试结果展示
│       └── 同步结果展示（逐项状态 + 提示）
└── Toast（固定定位，3s 自动消失）
```

## 状态管理

### provide/inject 机制

App.vue 作为根组件通过 `provide` 向下注入三个共享状态，子孙组件通过 `inject` 消费：

| Key | 类型 | 提供者 | 消费者 | 说明 |
|---|---|---|---|---|
| `showToast` | `(msg: string, type: 'success' \| 'error') => void` | App.vue | SchemeList, SchemeEditor, SshSync | 统一 Toast 通知 |
| `activeSchemeId` | `Ref<string>` | App.vue | SchemeList | 当前激活方案 ID，用于高亮 |
| `refreshActiveScheme` | `() => Promise<void>` | App.vue | SchemeList | 激活后刷新方案名 |

### 组件局部状态

每个 View 组件维护自己的 `ref`/`reactive` 状态，不跨组件共享。SchemeEditor 通过 `watch(selectedProvider, ...)` 将当前选中供应商同步到表单局部变量（`provAPIKey`、`provBaseURL`、`provAPIType`），保存时才写回。

**关键模式**：表单修改不直接修改 `scheme.value.providers`，而是通过局部 ref 编辑 → 点击保存 → 调 Wails API → `loadData()` 重新拉取全量数据。这确保了后端校验通过后才更新前端视图。

## 数据流

```
用户操作 → 表单局部状态 (ref/reactive)
    → Wails API 调用 (api().XxxMethod(...))
    → Go 后端校验 + 持久化
    → 返回结果
    → loadData() 重新拉取数据（保持前端与后端一致）
    → showToast() 通知用户
```

**错误处理**：每个 API 调用包裹在 `try/catch` 中，catch 块通过 `showToast(e?.message || e, 'error')` 展示错误。表单级错误（如校验失败）展示在表单内联 `field-error` 区域而非 Toast。

**重要**：前端不直接读写磁盘，所有状态变更通过 Wails API。

## API 调用模式

`frontend/src/wails/api.ts` 导出 `api()` 工厂函数：

```typescript
// 模式示意：非可复制代码
function api(): AppAPI {
  if (window['go']?.main?.App) return window['go']?.main?.App  // Wails 运行时桥接
  return { /* dev mode fallback: 所有方法返回空值/空数组 */ }
}
```

- **生产模式**：Wails 自动注入 `window.go.main.App`，`api()` 直接返回桥接对象
- **浏览器开发模式**：返回 mock 对象，所有方法返回空结果（不报错）
- **调用方**：每个 View 组件在 `setup` 中通过 `const a = api()` 获取引用后调用

**AppAPI 接口**（完整方法签名见 `api.ts`）：方案 CRUD（6 个）、供应商管理（4 个）、模型管理（8 个，含排序/拉取/导入）、连通性测试（1 个）、导入导出（2 个）、SSH 同步（4 个）、目录查询（2 个）。

## 关键交互模式

### 供应商选择与表单同步

SchemeEditor 的 `selectedProviderKey` 驱动右侧视图。切换供应商时：
1. `watch(selectedProviderKey)` 清除搜索文本、过滤条件、多选状态、连通性测试结果
2. `watch(selectedProvider)` 将新选中供应商的 `apiKey`/`baseUrl`/`apiType` 同步到局部表单变量

### 内置 vs 自定义供应商表单分支

- **内置供应商**：仅显示 API Key 和 Base URL 字段（API 类型由 `allBuiltIns` 查询，不可编辑）
- **自定义供应商**：显示所有字段（Key 禁用，Base URL、API 类型、API Key 可编辑）

`getEffectiveAPIType(prov)` 用于连通性测试判断：内置供应商查 `BuiltInProviders` 列表获取 API 类型，自定义供应商直接取 `prov.apiType`。

### 模型预设（MODEL_PRESETS）

- 仅在**添加模式**显示（`!editingModel`）
- 选择预设后调用 `applyPreset()` 覆盖 `modelForm` 的 id/name/reasoning/inputText/inputImage/contextWindow/maxTokens
- Cost 字段保持 0，不随预设填充
- 预设为前端硬编码，无后端调用

### 拖拽排序

- **供应商排序**：`dragstart`/`dragover`/`drop` 事件 → 前端 splice 重排 → `api().ReorderProviders(schemeID, orderedKeys)` → `loadData()` 刷新
- **模型排序**：同上模式，调用 `api().ReorderModels(schemeID, providerKey, orderedIDs)`
- 拖拽状态通过局部 `ref`（`providerDragKey`/`modelDragID` 等）管理

### 模型搜索与过滤

全部为客户端计算（`computed`），不发起后端请求：
- `modelSearchQuery`：按 ID 或 Name 子串匹配（大小写不敏感）
- `modelCapFilter`：按 `reasoning` 或 `inputImage` 能力过滤
- `allFilteredSelected` + `toggleSelectAllFiltered`：对过滤后的可见模型全选/取消

### 模型拉取流程

1. 点击「拉取模型列表」→ `api().FetchProviderModels(schemeID, providerKey)` → 返回 `Model[]`
2. 弹出选择 Dialog（`showFetchDialog`），显示拉取的模型列表 + 全选复选框
3. 用户勾选 → 点击导入 → `api().ImportProviderModels(schemeID, providerKey, selectedModels)` → 返回实际导入数
4. 关闭 Dialog → `loadData()` 刷新

### 批量导入 JSON

用户粘贴 JSON 模型数组 → 前端解析并校验每项必须含 `id` 字段 → 调用 `api().ImportProviderModels` → 显示导入/跳过统计

### 密码字段显隐切换

所有 API Key 输入框附带「显示/隐藏」按钮，通过局部 `showXxxKey` ref 切换 `input type="text"` / `"password"`。

## 错误处理矩阵

| 错误场景 | 展示方式 | 来源 |
|---|---|---|
| API 调用异常 | Toast（右上角，红色，3s） | `catch (e) → showToast(e?.message, 'error')` |
| 表单字段校验失败 | 表单内联 `field-error`（红色小字） | 前端 computed / 后端返回 error |
| 添加供应商表单错误 | `addBuiltInError` / `addCustomError` ref | 后端返回 error |
| 供应商保存错误 | `provError` ref | 后端返回 error |
| 模型表单错误 | `modelFormErrors.id` / `modelFormErrors.server` | 前端即时校验 + 后端返回 |
| 批量导入 JSON 解析错误 | `batchImportError` ref | 前端 JSON.parse 校验 |
| 连通性测试结果 | 按钮旁内联文字（绿色/红色） | 后端返回消息字符串 |

**Toast 实现**：App.vue 维护 `toast` ref，`showToast` 函数设置值并启动 3s 定时器自动清除。新 Toast 覆盖旧 Toast（重置定时器）。

## CSS 样式系统

全局样式在 `style.css`，使用 CSS 自定义属性（变量）体系：

| 变量 | 用途 |
|---|---|
| `--bg-primary` / `--bg-card` | 页面/卡片背景 |
| `--text-primary` / `--text-secondary` | 主/次文字颜色 |
| `--border-color` | 边框和分隔线 |
| `--accent` / `--accent-hover` | 主色调（蓝色按钮） |
| `--danger` / `--danger-hover` | 危险操作（红色） |
| `--success` / `--success-bg` | 成功状态（绿色） |
| `--tag-builtin` / `--tag-custom` | 内置/自定义标签底色 |
| `--radius` | 统一圆角 |

**组件样式**：Vue SFC 使用 `<style scoped>` 进行组件级隔离。全局工具类（`.card`、`.modal-overlay`、`.toast`、`.btn-*`、`.tag-*` 等）定义在 `style.css`，组件内直接使用。

## 重验条件

- 前端路由变更或新增页面时
- Vue 依赖升级（Vue 3 / Vue Router 大版本）
- Wails 桥接模式变更（`window.go.main.App` 接口）
- MODEL_PRESETS 增删时
