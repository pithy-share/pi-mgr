# 实施计划：模型列表自动拉取

基于 PRD `task/auto-model-list/prd.md`。

## 产品不变量（确认不变）

1. **单文件持久化**：所有状态经 `api.go` → `store.go`（schemes.json 原子写入），无其他存储
2. **Wails API 桥接**：所有变更必须通过 `App` 结构上的导出方法；前端不直接读写磁盘
3. **零 HTTP 依赖**：当前代码库无 `net/http` 导入。本功能首次引入，豁免见 PRD §网络调用豁免声明
4. **frontend/wails/api.ts** 是前端 Wails 桥接层，手动列出所有可用 API
5. **Model 唯一性**：同一 Provider 下 Model.ID 唯一（`validate.go` 保证）
6. **前端兜底**：`api.ts` 的 `api()` 在无 Wails 运行时返回空存根——本功能需在存根中添加对应方法避免未定义错误
7. **Provider.APIType**：内置供应商从 `BuiltInProviders` 继承；自定义从 `ValidAPITypes` 选择
8. **ValidAPITypes** 当前值：`["openai-completions", "anthropic-messages", "google-generative-ai"]`
9. **内置供应商 APIType**：`openai-completions`（openai, deepseek, groq, xai, openrouter, together, fireworks, cerebras, nvidia, huggingface）、`anthropic-messages`（anthropic）、`google-generative-ai`（google）、`mistral-conversations`（mistral）、`bedrock-converse-stream`（bedrock）

---

## 实施步骤

### Step 1：后端 - HTTP 客户端 + `FetchProviderModels`

**目标**：新增 `App.FetchProviderModels(schemeID, providerKey) ([]Model, error)`，实现 HTTP GET `{baseURL}/models` 并解析 OpenAI 兼容响应。

**文件**：`api.go`（新增方法），或新增 `fetch.go`（按 `serializer.go` 等多文件模式）

**改动**：

1. 在 `App` 上新增导出方法：
   - 通过 `GetScheme()` + `findProviderIndex()` 定位目标 Provider
   - 从 Provider 获取 `BaseURL`（拼接 URL：`strings.TrimRight(baseURL, "/") + "/models"`）和 `APIKey`
   - 创建 `http.Client`，`Timeout: 10 * time.Second`
   - 构造 GET 请求，设置 `Authorization: Bearer {apiKey}` 头（apiKey 为空时不设 Authorization 头，对应 AC-15）
   - 发送请求，检查 HTTP 状态码（非 200 返回包含状态码的错误）
   - 解析 JSON：`{"data": [{"id": "...", "name": "...", ...}]}`
   - 映射为 `[]Model`：仅填充 `ID` 和 `Name` 字段；`Name` 优先取响应 `name`，无则复制 `id`；其余字段保持零值（对应 defaultModel 语义：`Reasoning=false, InputText=true, InputImage=false, ContextWindow=128000, MaxTokens=16384, Cost*=0`）
   - 返回 `[]Model`，或错误

2. 错误分类（从前端区分用，统一返回 `error` 即可，前端根据 error string 或类型进行展示）：
   - 网络错误：DNS / 超时 / 连接拒绝 → `"网络不可达: %v"`
   - HTTP 非 200：→ `"API 返回状态码 %d: %s"`
   - JSON 解析错误：→ `"响应解析失败: %v"`
   - Provider 不存在等业务错误：沿用现有 `fmt.Errorf("provider not found: ...")`

**关键联动**：
- `store.go` 的 `GetScheme()` 用于查找 Scheme
- `api.go` 的 `findProviderIndex()` 用于定位 Provider
- `models.go` 的 `Model` 结构体作为返回类型
- `net/http` 为新增 import，`go mod tidy` 会自动拉取（Go 标准库，无需外部依赖）
- `encoding/json` 已有 import，用于解析 API 响应

**保持不变的控制流**：
- 不修改 `store.go` 持久化逻辑
- 不修改 `serializer.go` 序列化规则
- 不修改 `validate.go` 校验规则
- 不修改 `builtin.go` 内置供应商目录
- 不修改 Provider CRUD 控制流

**验证方式**：
- 单元测试：mock HTTP server 验证正常响应、非 200 响应、JSON 解析错误、网络超时
- 对应 AC：AC-02, AC-07, AC-08, AC-09, AC-15

---

### Step 2：后端 - 批量导入 API `ImportProviderModels`

**目标**：新增 `App.ImportProviderModels(schemeID, providerKey, models []Model) (int, error)`，一次原子写入导入多个模型（跳过重复 ID）。

**文件**：`api.go`（新增方法）

**改动**：

1. 导出方法签名：`(a *App) ImportProviderModels(schemeID, providerKey string, models []Model) (int, error)`
   - 返回 `int` 为实际导入数量（0 表示全部跳过或无新增），前端据此展示提示
   - 定位 Scheme + Provider
   - 遍历传入 models，对每个 model：
     - 跳过 `Model.ID` 在 `Provider.Models` 中已存在的条目（AC-10）
     - 其余追加到 `Provider.Models`
   - 调用 `SaveSchemes()` 单次原子写入
   - 返回实际添加数量
   - 若全部跳过（添加 0 个），仍然正常返回，不上报错误（AC-11）

2. **不**复用 `ValidateModel` 校验已有 ID 重复——因为 `ValidateModel` 的设计是对新单条目的校验（返回错误），而批量导入的语义是"跳过重复"而非"报错"。重复检测仅通过 ID 集合比对实现。

**关键联动**：
- `store.go` 的 `LoadSchemes()` / `SaveSchemes()`
- 复用 `findSchemeIndex()` / `findProviderIndex()`

**保持不变的控制流**：
- 不修改 `AddModel` 的单条添加逻辑（两个路径并存）
- 不修改 `ValidateModel` 校验规则
- `schemes.json` 仍然单次原子写入

**验证方式**：
- 单元测试：含重复 ID 的列表、全部重复、全部新模型
- 对应 AC：AC-04, AC-10, AC-11, AC-17

---

### Step 3：前端 - API 桥接层更新

**目标**：在 `wails/api.ts` 中添加 `FetchProviderModels` 和 `ImportProviderModels` 的类型定义和存根实现。

**文件**：`frontend/src/wails/api.ts`

**改动**：

1. 在 `AppAPI` 接口中添加：
```typescript
FetchProviderModels(schemeID: string, providerKey: string): Promise<Model[]>
ImportProviderModels(schemeID: string, providerKey: string, models: Model[]): Promise<number>
```

2. 在 `api()` 的 fallback 对象中添加对应存根：
```typescript
FetchProviderModels: () => Promise.resolve([] as Model[]),
ImportProviderModels: () => Promise.resolve(0),
```

3. 确认 import 中已有 `Model` 类型引用（已有 `import type { Scheme, Provider, Model, BuiltInProvider } from '../types'`，`Model` 已在其中）

**关键联动**：`types.ts` 的 `Model` 接口——返回的 `Model` 实例中仅 `id` 和 `name` 有值，其余字段为零值，与 `defaultModel()` 的默认值一致。

**保持不变的控制流**：
- 不修改 `AppAPI` 中的现有方法签名
- 不修改 `types.ts`

**验证方式**：
- 检查 TypeScript 编译无类型错误
- 对应 AC：AC-02（前端调用入口）

---

### Step 4：前端 - SchemeEditor.vue 按钮与交互

**目标**：在方案编辑器的模型管理区域添加"拉取模型列表"按钮、条件可见性、加载状态和选择对话框。

**文件**：`frontend/src/views/SchemeEditor.vue`

#### 4a：按钮条件显示与触发

**改动**：

1. 在模型列表区域标题旁（"模型列表" h3 同一行，`+ 添加模型` 按钮旁边或之前），添加"拉取模型列表"按钮：
```vue
<button class="btn-secondary btn-small" @click="handleFetchModels"
  :disabled="!canFetchModels" :title="fetchButtonTitle">
  ⟳ 拉取模型列表
</button>
```

2. 计算属性 `canFetchModels` 和 `fetchButtonTitle`：
   - `selectedProvider` 不存在 → `disabled=true`, title="请先选择供应商"
   - `selectedProvider.apiType` 不是 `openai-completions` 也不是 `openai-responses` → `disabled=true`, title="该 API 类型不支持自动拉取"（AC-12, AC-13）
   - `selectedProvider.baseUrl` 为空 → `disabled=true`, title="请先配置 baseUrl"（AC-14）
   - 默认 → `disabled=false`, title=""

3. 添加状态变量：
   - `fetchingModels: ref(false)` — 控制 spinner
   - `fetchedModels: ref<Model[]>([])` — 缓存拉取结果
   - `showFetchDialog: ref(false)` — 控制选择对话框显示
   - `fetchError: ref('')` — 错误消息

4. `handleFetchModels` 方法：
   - 设置 `fetchingModels = true`, `fetchError = ''`
   - 调用 `api().FetchProviderModels(props.id, selectedProviderKey.value)`
   - 成功 → `fetchedModels = result; showFetchDialog = true`
   - 失败 → `fetchError = error message; showToast(error, 'error')`（AC-07, AC-08, AC-09）
   - finally → `fetchingModels = false`

5. 按钮旁添加加载指示器：
```vue
<span v-if="fetchingModels" class="spinner" style="margin-left:8px;"></span>
```
（spinner 样式参考现有 UI 风格，具体通过 CSS class 实现；按 Q4，不加取消按钮）

**关键联动**：
- 新增计算属性 `canFetchModels` 和 `fetchButtonTitle`
- `FetchProviderModels` API 调用

#### 4b：模型选择对话框

**改动**：在 `<template>` 中新增模态对话框（在 "添加/编辑模型 模态框" 与 "删除模型确认框" 之间或附近）：

```vue
<!-- Fetch models dialog -->
<div v-if="showFetchDialog" class="modal-overlay" @click.self="showFetchDialog = false">
  <div class="modal" style="max-height:70vh;overflow-y:auto;">
    <h3>选择要导入的模型</h3>
    <p style="color:var(--text-secondary);margin-bottom:12px;">
      共 {{ fetchedModels.length }} 个模型，请勾选要导入的条目
    </p>
    <!-- Select all -->
    <label style="display:flex;align-items:center;gap:6px;margin-bottom:8px;cursor:pointer;">
      <input type="checkbox" v-model="fetchSelectAll" @change="handleFetchSelectAll" />
      <strong>全选</strong>
    </label>
    <div v-for="m in fetchedModels" :key="m.id" style="display:flex;align-items:center;gap:8px;padding:4px 0;">
      <input type="checkbox" :value="m.id" v-model="fetchSelectedIds" />
      <span><strong>{{ m.id }}</strong></span>
      <span v-if="m.name && m.name !== m.id" style="color:var(--text-secondary);">({{ m.name }})</span>
    </div>
    <div style="margin-top:12px;display:flex;gap:6px;justify-content:flex-end;">
      <button class="btn-secondary" @click="showFetchDialog = false">取消</button>
      <button class="btn-primary" @click="handleImportModels" :disabled="fetchSelectedIds.length === 0">导入 ({{ fetchSelectedIds.length }})</button>
    </div>
  </div>
</div>
```

**新增状态**：
- `fetchSelectAll: ref(false)` — 全选状态
- `fetchSelectedIds: ref<string[]>([])` — 已勾选的模型 ID 列表
- `handleFetchSelectAll()` — 全选/取消全选切换

**关键联动**：
- 对话框展示条件：`showFetchDialog`
- 选择状态管理：`fetchSelectAll`, `fetchSelectedIds`

#### 4c：批量导入与提示

**改动**：新增 `handleImportModels` 方法：

```typescript
async function handleImportModels() {
  if (fetchSelectedIds.value.length === 0) {
    showToast?.('请至少选择一个模型', 'error')
    return // AC-16
  }
  try {
    const selectedModels = fetchedModels.value.filter(m => fetchSelectedIds.value.includes(m.id))
    const a = api()
    const count = await a.ImportProviderModels(props.id, selectedProviderKey.value, selectedModels)
    showFetchDialog.value = false
    fetchedModels.value = []
    fetchSelectedIds.value = []
    fetchSelectAll.value = false
    await loadData() // 刷新界面，显示新增的模型（AC-05）
    if (count > 0) {
      const skipped = selectedModels.length - count
      const msg = skipped > 0 ? `已导入 ${count} 个模型，${skipped} 个因重复跳过` : `已导入 ${count} 个模型`
      showToast?.(msg, 'success')
    } else {
      showToast?.('无新增模型（所有选中模型均已存在）', 'success') // AC-11
    }
  } catch (e: any) {
    showToast?.(e?.message || e, 'error')
  }
}
```

**关键联动**：
- `ImportProviderModels` API
- `loadData()` 刷新整个方案数据——保证导入后界面显示最新模型列表（AC-05）
- `selectedProvider` watch 中已包含 `provAPIKey`, `provBaseURL`, `provAPIType` 的同步——当用户在导入前已修改但未保存这些字段时，显示的仍是已保存版本（AC-17：未保存的编辑内容不受影响）

**保持不变的控制流**：
- 不修改现有模型添加/编辑/删除功能
- 不修改供应商配置保存/修改功能
- 不修改方案列表页
- `loadData()` 的调用模式不变（重新从后端加载全量数据）

**验证方式**：
- 手动测试：内置 openai-completions 供应商点按钮、自定义供应商、无 baseUrl、非 openai API 类型
- 对应 AC：AC-01, AC-03, AC-04, AC-05, AC-06, AC-07, AC-08, AC-09, AC-10, AC-11, AC-12, AC-13, AC-14, AC-15, AC-16, AC-17

---

## 验收与验证

每个 AC 的验证方式如下。引用 PRD 原始 AC 编号。

| AC | 验证方式 | 步骤 |
|---|---|---|
| AC-01 | 前端视觉确认 | 在 SchemeEditor 中选中 `openai-completions` 内置/自定义供应商，"⟳ 拉取模型列表" 按钮可见且可用 |
| AC-02 | 浏览器 DevTools Network + 日志 | 点击按钮后，Go 端发起 `GET {baseUrl}/models` 并携带 `Authorization: Bearer {apiKey}` 头 |
| AC-03 | 前端视觉确认 | 请求返回后弹窗显示模型列表（id + name），每项有 checkbox，支持全选 |
| AC-04 | 数据库（schemes.json）验证 | 勾选多个模型→确认导入→schemes.json 中该 Provider.Models 新增条目，仅 id/name 有值，其余为默认值 |
| AC-05 | 前端视觉 + 表单操作 | 导入后模型出现在方案编辑器的模型列表；点击编辑可修改 contextWindow 等；保存后刷新页面数据保留 |
| AC-06 | 全流程验证 | 对自定义的 openai-completions 类型供应商重复 AC-01~AC-05，行为一致 |
| AC-07 | 模拟网络不可达 | 断开网络/关闭代理，点击按钮显示错误提示（toast），"添加模型" 按钮仍可用 |
| AC-08 | 模拟非 200 | 配置指向返回 401/404/500 的端点，点击按钮显示 HTTP 状态码和错误，手动添加仍可用 |
| AC-09 | 模拟畸形 JSON | 配置指向返回非 JSON 或 JSON 不合预期的端点，显示解析错误，手动添加仍可用 |
| AC-10 | 数据库验证 | 导入包含已有 ID 的列表，schemes.json 中该 ID 仍为原有记录，其余新 ID 正常添加；toast 显示跳过数量 |
| AC-11 | 前端验证 | 导入列表中全是已有 ID，toast 提示"无新增模型" |
| AC-12 | 前端视觉 | 选中 `anthropic-messages` 供应商，按钮置灰，hover 显示"该 API 类型不支持自动拉取" |
| AC-13 | 前端视觉 | 选中 `google-generative-ai` 供应商，同 AC-12 |
| AC-14 | 前端视觉 | 新建无 baseUrl 的自定义供应商，按钮置灰，hover 显示"请先配置 baseUrl" |
| AC-15 | 请求日志 + 功能 | 无 apiKey 的供应商按钮可用；请求日志确认无 Authorization 头 |
| AC-16 | 前端验证 | 打开选择弹窗，不勾选任何模型直接点确认，toast 显示"请至少选择一个模型" |
| AC-17 | 前端验证 | 在导入前修改某个模型的编辑表单字段但不保存；执行导入后，未保存的编辑内容仍在表单中，模型列表新增条目 |

---

## 决策与风险

### 决策

| # | 决策 | 理由 |
|---|---|---|
| D1 | **后端新增 `fetch.go` 存放 HTTP 客户端代码** | 保持 `api.go` 聚焦于 API 桥接，遵循 `serializer.go`、`activate.go` 等职责分离模式 |
| D2 | **新增 `ImportProviderModels` 批量 API 而非前端循环调 `AddModel`** | 避免 N 次原子写入（N 次磁盘 I/O），且单次写入可保证 AC-10/AC-11 的全部跳过语义与错误消息原子性 |
| D3 | **按钮条件按 `openai-completions` 和 `openai-responses` 两个字符串匹配** | 覆盖 PRD 要求的全部支持类型；`openai-responses` 虽未在 `ValidAPITypes` 中存在，但 PRD 明确纳入范围 |
| D4 | **批量导入返回 `int`（导入数量），前端据此推导跳过数量** | 前端可计算 `selectedModels.length - count = skipped`，满足 AC-10 的提示需求 |
| D5 | **apiKey 为空时不设 `Authorization` 头** | 对应 AC-15：某些本地服务不需要 key |

### 风险

| # | 风险 | 影响 | 缓解 |
|---|---|---|---|
| R1 | 某些中转站 / Ollama 的 `/v1/models` 端点路径不同 | 拉取失败走兜底手动添加（AC-07~AC-09） | 按 OpenAI 标准实现，非标准路径由用户手动配置 |
| R2 | `openai-responses` 当前不在 `ValidAPITypes`，用户无法选择 | 按钮通过字符串匹配仍会正确隐藏 | 不影响现有功能；后续 `ValidAPITypes` 扩增时自动生效 |
| R3 | 首次引入 `net/http`，需确保在 Wails 内不触发 TLS 证书问题 | Windows 环境使用系统证书池 | Go 默认使用系统根证书，Wails 环境亦然 |

### 不发生的场景（明确不入计划）

- 不做模型列表缓存（每次用户主动拉取）
- 不改 `serializer.go`、`validate.go`、`builtin.go`、`activate.go`
- 不改供应商 CRUD（Add/Update/Remove Provider）
- 不改方案 CRUD
- 不做并发请求/冲突检测（单进程单写者）
- 不为 `anthropic-messages` / `google-generative-ai` 提供 fetch（已明确非目标）
- 不为 `mistral-conversations` / `bedrock-converse-stream` 提供 fetch（非 OpenAI 兼容 List Models 端点）

---

## 未解决 Blocker

无。
