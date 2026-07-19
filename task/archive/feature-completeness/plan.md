# 实施计划：功能完善增强

> 基于 `prd.md` v1，证据截止 2025-07-19

---

## 证据链摘要

### 产品不变量（从现有实现确认）

| 不变量 | 来源 |
|--------|------|
| Scheme.Providers 为有序切片，Provider.Models 为有序切片 | `models.go:5-29` |
| 内置供应商 `APIType` 在 `BuiltInProviders` 硬编码目录中定义，前端不修改 | `builtin.go:7-21` |
| `AddBuiltInProvider` 按 Key 查重，重复返回“该内置供应商已添加” | `api.go:158-161` |
| `ValidAPITypes` 是自定义供应商的可选 API 类型全集；`FetchProviderModels` 后端不按 API 类型过滤 | `builtin.go:31-40`, `fetch.go:14-78` |
| schemes.json 原子写入：temp + rename，跨卷回退直接写 | `store.go:50-62` |
| models.json 输出格式：`providers` 为 `map[string]providerJSON`（JSON 对象），models 为有序切片 | `serializer.go:8-10,37-74` |
| 前端通过 Wails 导出方法访问后端，不直接读写磁盘 | `spec/architecture/overview.md` |
| `canFetchModels` 前端计算属性控制"拉取模型"按钮可用性，限制 `openai-completions` 和 `openai-responses` | `SchemeEditor.vue` L~450-460 |
| 拖拽排序无现有实现；无批量删除 API；无连通性测试 API | `api.go` 全文件 |

### 关键调用入口

| 入口 | 文件:行 | 调用方 |
|------|---------|--------|
| `ListBuiltInProviders()` | `api.go:614-616` | 前端下拉 / `availableBuiltIns` computed |
| `AddBuiltInProvider(schemeID, providerKey, apiKey, baseURL)` | `api.go:143-188` | 前端 `handleAddBuiltIn` |
| `FetchProviderModels(schemeID, providerKey)` | `fetch.go:14-78` | 前端 `handleFetchModels` |
| `canFetchModels` (前端 computed) | `SchemeEditor.vue` | "拉取模型列表" 按钮 `:disabled` |
| `RemoveModel(schemeID, providerKey, modelID)` | `api.go:289-307` | 前端 `handleDeleteModel` |
| `SerializeToModelsJSON(scheme *Scheme)` | `serializer.go:37-74` | `ActivateScheme` → `activate.go:16-38` |
| `SaveSchemes(schemes)` | `store.go:49-62` | 所有变更 API |

---

## 实施步骤

### 步骤 1：内置供应商扩展 (AC-01, AC-02, AC-03, AC-04, AC-05)

**目标**：补齐 13 个缺失内置供应商，更新 ValidAPITypes。

**文件**：`builtin.go`

**改动**：
1. 在 `BuiltInProviders` 切片末尾追加 13 个新条目（见下表），均设 `APIType: "openai-completions"`
2. AC-03 已满足：`ValidAPITypes` 第 33 行已含 `"azure-openai-responses"`，无需改动

**新增供应商清单**：

| Key | Name | APIType |
|-----|------|---------|
| `ant-ling` | Ant Ling | openai-completions |
| `vercel-ai-gateway` | Vercel AI Gateway | openai-completions |
| `zai-coding-plan-global` | ZAI Coding Plan (Global) | openai-completions |
| `zai-coding-plan-china` | ZAI Coding Plan (China) | openai-completions |
| `opencode-zen` | OpenCode Zen | openai-completions |
| `opencode-go` | OpenCode Go | openai-completions |
| `kimi-for-coding` | Kimi For Coding | openai-completions |
| `minimax` | MiniMax | openai-completions |
| `minimax-china` | MiniMax (China) | openai-completions |
| `xiaomi-mimo` | Xiaomi MiMo | openai-completions |
| `xiaomi-mimo-china` | Xiaomi MiMo Token Plan (China) | openai-completions |
| `xiaomi-mimo-amsterdam` | Xiaomi MiMo Token Plan (Amsterdam) | openai-completions |
| `xiaomi-mimo-singapore` | Xiaomi MiMo Token Plan (Singapore) | openai-completions |

**不变量保持**：
- `ListBuiltInProviders()` 返回更新后的完整切片，前端 `availableBuiltIns` computed 自动排除已添加的 key
- `AddBuiltInProvider` 查重逻辑不变（AC-05 已满足）
- 前端下拉按名称排序由 `allBuiltIns` 原始顺序 + 浏览器默认 select 行为决定；需确认前端在 `<option>` 渲染前对 `availableBuiltIns` 排序

**联动**：
- **前端**需确认 `availableBuiltIns` 按 `name` 排序后再渲染 `<option>`，以满足 AC-04"按名称排序展示"（当前前端 `v-for="b in availableBuiltIns"` 未显式排序）
- `IsBuiltInProvider(key)` 函数自动覆盖新增 key，无需改动

**验证**：AC-01, AC-02, AC-03, AC-04, AC-05

---

### 步骤 2：FetchProviderModels 支持 azure-openai-responses (AC-06, AC-07)

**目标**：允许 `azure-openai-responses` 类型的供应商触发模型拉取。

**文件**：
- `frontend/src/views/SchemeEditor.vue` — `canFetchModels` computed
- `frontend/src/views/SchemeEditor.vue` — `fetchButtonTitle` computed

**改动**：
1. `canFetchModels` 第二行条件从 `apiType !== 'openai-completions' && apiType !== 'openai-responses'` 改为 `... && apiType !== 'azure-openai-responses'`
2. `fetchButtonTitle` 同理更新 tooltip 文本

**不变量保持**：
- 后端 `FetchProviderModels`（`fetch.go`）无 API 类型过滤，已天然支持任何 API 类型发 GET `/models`，无需改动
- `baseUrl` 非空检查不变

**验证**：AC-06, AC-07

---

### 步骤 3：模型搜索与能力过滤 (AC-08 至 AC-14)

**目标**：方案编辑器模型列表区域增加搜索框和能力过滤下拉。

**文件**：`frontend/src/views/SchemeEditor.vue`

**改动**：
1. 新增状态：`modelSearchQuery`（ref string）、`modelCapFilter`（ref `'all' | 'reasoning' | 'inputImage'`）
2. 新增 computed：`filteredModels`——从 `selectedProvider.models` 中实时过滤：
   - 文本搜索：`m.id.toLowerCase().includes(q)` 或 `m.name.toLowerCase().includes(q)`
   - 能力过滤：`reasoning` → `m.reasoning === true`；`inputImage` → `m.inputImage === true`
   - AND 关系，不区分大小写
3. 模型列表渲染区改为遍历 `filteredModels` 而非 `selectedProvider.models`
4. UI 布局：在 `"模型列表"` 标题和 `"⟳ 拉取模型列表"` 按钮之间插入搜索框 + 过滤下拉
   - 搜索框：`placeholder="搜索模型 ID 或名称"`
   - 过滤下拉：选项 `全部` / `reasoning` / `inputImage`，默认 `全部`
5. 空状态：`filteredModels.length === 0 && selectedProvider.models.length > 0` 时显示 "无匹配模型"
6. 模型增删改操作后，`loadData()` 刷新方案数据，搜索/过滤条件保持不变，computed 自动重新计算

**不变量保持**：
- `AddModel`、`UpdateModel`、`RemoveModel`、`ImportProviderModels` API 调用不变
- 持久化逻辑不变
- 原有未过滤时的空状态 "暂无自定义模型" 保留（当 `selectedProvider.models.length === 0` 时）

**验证**：AC-08, AC-09, AC-10, AC-11, AC-12, AC-13, AC-14

---

### 步骤 4：Provider 排序 API 与拖拽 (AC-15 至 AC-18, AC-22, AC-24, AC-25)

**目标**：新增后端排序 API，前端支持拖拽重排供应商列表。

**文件**：
- `api.go` — 新增 `ReorderProviders` 方法
- `frontend/src/wails/api.ts` — `AppAPI` 接口新增 + fallback 新增
- `frontend/src/views/SchemeEditor.vue` — 供应商列表拖拽交互

**后端改动** (`api.go`)：
```go
// ReorderProviders reorders providers in a scheme by the given key order.
func (a *App) ReorderProviders(schemeID string, orderedKeys []string) error {
    // 1. LoadSchemes → find scheme
    // 2. 校验：len(orderedKeys) == len(scheme.Providers)
    //    且每个 orderedKeys[i] 在 scheme.Providers 中存在且唯一
    //    （即集合完全一致）
    // 3. 校验失败 → "供应商列表不一致"
    // 4. 按 orderedKeys 构建新的 []Provider → 赋值 → SaveSchemes
}
```

**前端改动**：
1. `api.ts`：新增 `ReorderProviders(schemeID: string, orderedKeys: string[]): Promise<void>`
2. `SchemeEditor.vue`：
   - 供应商列表使用 HTML5 Drag & Drop（或轻量拖拽库），在 `dragend` 时调用 `ReorderProviders`
   - 仅允许在同一列表内拖拽，拖拽完成后传入新的 key 顺序数组
   - 列表渲染顺序由 `scheme.providers` 数组位置决定

**不变量保持**：
- Provider Key 唯一性约束不变
- `AddBuiltInProvider` / `AddCustomProvider` 仍然 append 到末尾
- `RemoveProvider` 不变
- `models.json` 输出中 models 数组保持 `scheme.Providers[i].Models` 顺序（已由切片保证）

**联动**：无

**验证**：AC-15, AC-16, AC-17, AC-18, AC-22, AC-24, AC-25

---

### 步骤 5：Model 排序 API 与拖拽 (AC-19 至 AC-21, AC-23, AC-24, AC-25)

**目标**：新增后端排序 API，前端支持拖拽重排模型列表。

**文件**：
- `api.go` — 新增 `ReorderModels` 方法
- `frontend/src/wails/api.ts` — `AppAPI` 接口新增 + fallback 新增
- `frontend/src/views/SchemeEditor.vue` — 模型列表拖拽交互

**后端改动** (`api.go`)：
```go
// ReorderModels reorders models in a provider within a scheme.
func (a *App) ReorderModels(schemeID string, providerKey string, orderedIDs []string) error {
    // 1. LoadSchemes → find scheme → find provider
    // 2. 校验：len(orderedIDs) == len(prov.Models)
    //    且每个 orderedIDs[i] 在 prov.Models 中存在且唯一
    // 3. 校验失败 → "模型列表不一致"
    // 4. 按 orderedIDs 构建新的 []Model → 赋值 → SaveSchemes
}
```

**前端改动**：
1. `api.ts`：新增 `ReorderModels(schemeID: string, providerKey: string, orderedIDs: string[]): Promise<void>`
2. `SchemeEditor.vue`：
   - 模型列表使用 HTML5 Drag & Drop，在 `dragend` 时调用 `ReorderModels`
   - 拖拽后的新顺序传入 API；搜索/过滤激活时拖拽应在 `filteredModels` 上操作，但排序影响的是原始 `selectedProvider.models` 的全集（拖拽时临时清除过滤，完成后恢复）

**不变量保持**：
- Model ID 唯一性约束不变
- `AddModel` 仍然 append 到末尾
- `models.json` 输出中 models 数组保持新顺序（切片顺序）
- `RemoveModels` 不影响其他 model 的相对顺序

**联动**：
- 拖拽与搜索/过滤的交互：建议拖拽期间暂停过滤，或仅在未过滤状态下允许拖拽（简单且不易出错）

**验证**：AC-19, AC-20, AC-21, AC-23, AC-24, AC-25

---

### 步骤 6：批量删除 API (AC-26 至 AC-28)

**目标**：新增 `RemoveModels` API，支持批量删除模型。

**文件**：
- `api.go` — 新增 `RemoveModels` 方法
- `frontend/src/wails/api.ts` — `AppAPI` 接口新增 + fallback 新增

**后端改动** (`api.go`)：
```go
// RemoveModels removes multiple models from a provider. Best-effort: skips
// models that don't exist, returns count of successfully removed models.
// Returns (0, nil) when no models were removed (all skipped or not found).
func (a *App) RemoveModels(schemeID string, providerKey string, modelIDs []string) (int, error) {
    // 1. LoadSchemes → find scheme → find provider
    // 2. 构建 existing ID set
    // 3. 过滤 prov.Models，剔除在 modelIDs 中存在的 ID
    // 4. removed = 原长度 - 新长度
    // 5. 如果 removed > 0 → SaveSchemes → 返回 (removed, nil)
    //    如果 removed == 0 → 返回 (0, nil)（AC-28：不存在的 ID 静默跳过）
    // 注：scheme/provider 不存在或 SaveSchemes 失败为系统性错误，返回 error
}
```

**不变量保持**：
- `RemoveModel`（单条）保持不变
- `SaveSchemes` 原子写入不变
- 已存在的 Model ID 唯一性约束不变

**验证**：AC-26, AC-27, AC-28

---

### 步骤 7：批量操作前端交互 (AC-29 至 AC-34)

**目标**：模型列表增加多选复选框、批量删除和批量导入 JSON。

**文件**：`frontend/src/views/SchemeEditor.vue`

**改动**：

**7a. 多选 + 批量删除 (AC-29 至 AC-31)**：
1. 新增状态：`selectedModelIDs`（ref `Set<string>` / `string[]`）
2. 模型列表每行左侧增加 checkbox，绑定到 `selectedModelIDs`
3. 列表头增加全选 checkbox：搜索/过滤激活时仅选中 `filteredModels` 的 ID
4. 选中 ≥1 个模型时显示"删除所选（N）"按钮（否则隐藏）
5. 点击后弹出确认对话框 "确认删除选中的 N 个模型？"
6. 确认后调用 `RemoveModels` API，成功后 `loadData()` 刷新并清空选中

**7b. 批量导入 JSON (AC-32 至 AC-34)**：
1. 模型列表上方增加"批量导入 JSON"按钮
2. 点击后弹出模态框，内含 `<textarea>`（供粘贴 JSON 数组）
3. 前端解析 JSON：`JSON.parse(text)`，验证为数组且每项含 `id` 字段
4. 格式无效 → "JSON 格式错误"提示
5. 解析成功后调用 `ImportProviderModels`，展示导入结果（成功 N，跳过 M）
6. 导入完成后 `loadData()` 刷新

**不变量保持**：
- `ImportProviderModels` API 不变（按 ID 跳过已存在）
- `RemoveModels` API 为步骤 6 新增

**验证**：AC-29, AC-30, AC-31, AC-32, AC-33, AC-34

---

### 步骤 8：模型预设 (AC-35 至 AC-39)

**目标**：手动添加模型时提供 11 个已知模型预设。

**文件**：`frontend/src/views/SchemeEditor.vue`（或新建 `frontend/src/presets.ts`）

**改动**：
1. 新建常量 `MODEL_PRESETS`（前端硬编码，11 条）：
```typescript
const MODEL_PRESETS: { label: string; model: Model }[] = [
  { label: 'GPT-4o',              model: { id: 'gpt-4o',          name: 'GPT-4o',          reasoning: false, inputText: true,  inputImage: true,  contextWindow: 128000, maxTokens: 16384, ... } },
  { label: 'GPT-4o-mini',         model: { id: 'gpt-4o-mini',     name: 'GPT-4o-mini',     reasoning: false, inputText: true,  inputImage: true,  contextWindow: 128000, maxTokens: 16384, ... } },
  { label: 'GPT-4.1',             model: { ... } },
  { label: 'GPT-4.1-mini',        model: { ... } },
  { label: 'Claude Sonnet 4.5',   model: { ... } },
  { label: 'Claude Opus 4.5',     model: { ... } },
  { label: 'Claude Haiku 4.5',    model: { ... } },
  { label: 'Gemini 2.5 Pro',      model: { ... } },
  { label: 'Gemini 2.5 Flash',    model: { ... } },
  { label: 'DeepSeek V3',         model: { ... } },
  { label: 'DeepSeek R1',         model: { ... } },
]
```
  每个预设需填入准确的 `id`、`name`、`reasoning`、`inputText`、`inputImage`、`contextWindow`、`maxTokens`；Cost 全为 0。

2. 在模型表单（添加/编辑模态框）顶部增加"预设模型"下拉：
   - 选项："自定义"（默认） + 11 个预设名称
   - 选择预设后自动填充 `modelForm` 各字段（覆盖当前值）
   - 用户可继续修改任意字段（AC-38）
3. 预设下拉仅在"添加模型"时显示（`!editingModel`），编辑模式下隐藏

**不变量保持**：
- 纯前端数据，无后端交互
- Cost 字段由用户自行填写
- 不影响已有模型编辑

**验证**：AC-35, AC-36, AC-37, AC-38, AC-39

---

### 步骤 9：供应商连通性测试 (AC-40 至 AC-45)

**目标**：新增后端连通性测试 API + 前端测试按钮。

**文件**：
- `api.go` — 新增 `TestProviderConnectivity` 方法
- `frontend/src/wails/api.ts` — `AppAPI` 接口新增 + fallback 新增
- `frontend/src/views/SchemeEditor.vue` — 测试按钮及状态

**后端改动** (`api.go` 或新建 `connectivity.go`)：
```go
// TestProviderConnectivity tests connectivity to a provider's API endpoint.
// Returns a Chinese status message. Does not persist any data.
func (a *App) TestProviderConnectivity(schemeID string, providerKey string) (string, error) {
    // 1. GetScheme → find provider
    // 2. 校验 API type: 仅 openai-completions / openai-responses / azure-openai-responses
    //    否则 → "", fmt.Errorf("该 API 类型暂不支持连通性测试")
    // 3. 校验 baseUrl 非空 → "", fmt.Errorf("请先配置 Base URL")
    // 4. HTTP GET {baseURL}/models, Authorization: Bearer {apiKey}（apiKey 为空不发送）
    //    Timeout: 10s
    // 5. 2xx → "连接成功，API 可达", nil
    // 6. 超时/DNS/连接拒绝 → "无法连接，请检查 Base URL 和网络", nil
    // 7. 非 2xx → fmt.Sprintf("API 返回错误（状态码 %d），请检查 API Key", code), nil
    // 8. 其他错误 → 原始错误摘要, nil
}
```
  注：返回 `(string, error)` 格式：成功或可预期的失败均返回 `(message, nil)`；仅前置校验失败返回 `("", error)`。

**前端改动**：
1. `api.ts`：新增 `TestProviderConnectivity(schemeID: string, providerKey: string): Promise<string>`
2. `SchemeEditor.vue`：
   - 供应商详情区域（保存按钮旁）增加"测试连接"按钮
   - 按钮可用条件（computed `canTestConnectivity`）：
     - `apiType` 为 `openai-completions` / `openai-responses` / `azure-openai-responses`
     - `baseUrl` 非空
   - 不可用时显示对应 tooltip（AC-40"该 API 类型暂不支持连通性测试" / AC-41"请先配置 Base URL"）
   - 新增状态：`testingConnectivity`（ref boolean）
   - 测试期间按钮显示加载状态并禁用
   - 结果通过 `showToast` 显示（成功绿色 / 失败红色）

**不变量保持**：
- 不持久化测试结果
- 不影响 `FetchProviderModels` 逻辑
- 与 AGENTS.md"不校验 API key 有效性"不冲突（手动触发 vs 自动校验）

**验证**：AC-40, AC-41, AC-42, AC-43, AC-44, AC-45

---

## 验收与验证

| PRD AC | 步骤 | 验证方式 |
|--------|------|----------|
| AC-01 | 步骤 1 | 对比 `builtin.go` 中 `BuiltInProviders` 列表与 PRD 清单；`ListBuiltInProviders()` 调用返回包含全部 27 个条目 |
| AC-02 | 步骤 1 | 检查新增 13 个条目的 `APIType` 字段均为 `"openai-completions"` |
| AC-03 | 步骤 1 | 检查 `builtin.go` 中 `ValidAPITypes` 已包含 `"azure-openai-responses"`（已满足） |
| AC-04 | 步骤 1 | 启动应用 → 进入方案编辑器 → 添加内置供应商下拉：13 个新供应商可见，按名称排序 |
| AC-05 | 步骤 1 | 已添加过的内置供应商不在下拉中显示；API 直接调用 `AddBuiltInProvider` 返回"该内置供应商已添加"（已有逻辑，无需新增） |
| AC-06 | 步骤 2 | `azure-openai-responses` 供应商配置 baseUrl + apiKey → "拉取模型列表"按钮可用 → 点击发起请求 |
| AC-07 | 步骤 2 | 检查 `canFetchModels` computed 包含 `azure-openai-responses` |
| AC-08 | 步骤 3 | 模型列表顶部出现搜索输入框，placeholder="搜索模型 ID 或名称" |
| AC-09 | 步骤 3 | 输入关键词 → 列表实时过滤（仅匹配的模型可见） |
| AC-10 | 步骤 3 | 清空搜索 → 列表恢复完整 |
| AC-11 | 步骤 3 | 输入无匹配关键词 → 显示"无匹配模型" |
| AC-12 | 步骤 3 | 搜索状态下添加/编辑/删除模型 → 操作完成后列表按当前搜索条件重新过滤 |
| AC-13 | 步骤 3 | 搜索框右侧有能力过滤下拉，选项：全部 / reasoning / inputImage |
| AC-14 | 步骤 3 | 同时设置文本搜索 + 能力过滤 → 仅显示同时满足两者的模型 |
| AC-15 | 步骤 4 | `ReorderProviders(schemeID, orderedKeys)` API 存在；签名正确 |
| AC-16 | 步骤 4 | 传入增/删/改过的 `orderedKeys` → 返回"供应商列表不一致" |
| AC-17 | 步骤 4 | 合法 `orderedKeys` → 重排后 `schemes.json` 中 Provider 数组顺序与传入一致；一次性写入（无中间态） |
| AC-18 | 步骤 4 | 前端拖拽排序 → 调用 `ReorderProviders` → 刷新后顺序保持 |
| AC-19 | 步骤 5 | `ReorderModels(schemeID, providerKey, orderedIDs)` API 存在；签名正确 |
| AC-20 | 步骤 5 | 传入增/删/改过的 `orderedIDs` → 返回"模型列表不一致" |
| AC-21 | 步骤 5 | 合法 `orderedIDs` → 重排后 `schemes.json` 中 Model 数组顺序与传入一致；一次性写入 |
| AC-22 | 步骤 4 | 左侧供应商列表支持拖拽重排；拖拽后 `ReorderProviders` 被调用 |
| AC-23 | 步骤 5 | 右侧模型列表支持拖拽重排；拖拽后 `ReorderModels` 被调用 |
| AC-24 | 步骤 4+5 | 排序后 Key 唯一、ID 唯一的现有校验不变（不新增校验错误） |
| AC-25 | 步骤 4+5 | 激活方案后 `models.json` 中 models 数组顺序与 `schemes.json` 一致（Go 切片保证）；provider 对象内属性保持 struct 字段顺序 |
| AC-26 | 步骤 6 | `RemoveModels(schemeID, providerKey, modelIDs)` API 存在；签名 `(int, error)` |
| AC-27 | 步骤 6 | 传入 3 个 ID（1 存在、1 不存在、1 空字符串）→ 返回 `(1, nil)` |
| AC-28 | 步骤 6 | 传入全不存在的 ID → 返回 `(0, nil)`（静默跳过） |
| AC-29 | 步骤 7 | 模型列表每行左侧有复选框；全选复选框仅影响当前可见（搜索/过滤后）的模型 |
| AC-30 | 步骤 7 | 选中模型后显示"删除所选（N）"；未选中时隐藏 |
| AC-31 | 步骤 7 | 点击批量删除 → 确认对话框 → 确认后调用 `RemoveModels` → 刷新 |
| AC-32 | 步骤 7 | "批量导入 JSON"按钮 → 文本输入框 → 粘贴 JSON 数组 |
| AC-33 | 步骤 7 | 粘贴合法 JSON → 调用 `ImportProviderModels` → 显示导入 N，跳过 M |
| AC-34 | 步骤 7 | 粘贴非法 JSON → 前端提示"JSON 格式错误"，不发起 API 调用 |
| AC-35 | 步骤 8 | 添加模型模态框中出现"预设模型"下拉，默认"自定义" |
| AC-36 | 步骤 8 | 选择预设 → 自动填充 ID/Name/reasoning/inputText/inputImage/contextWindow/maxTokens；Cost 全部为 0 |
| AC-37 | 步骤 8 | 下拉包含全部 11 个预设模型 |
| AC-38 | 步骤 8 | 选择预设后修改 ID → 表单接受修改（不锁定） |
| AC-39 | 步骤 8 | 预设数据仅在 `presets.ts` 中定义，无后端 API 调用 |
| AC-40 | 步骤 9 | "测试连接"按钮：`openai-completions`/`openai-responses`/`azure-openai-responses` 可用；其他禁用并显示正确 tooltip |
| AC-41 | 步骤 9 | baseUrl 为空时按钮禁用，tooltip="请先配置 Base URL" |
| AC-42 | 步骤 9 | 点击 → GET `{baseURL}/models`，`Authorization: Bearer {apiKey}`（空 key 不发），超时 10s |
| AC-43 | 步骤 9 | 返回 2xx → 绿色 toast "连接成功，API 可达" |
| AC-44 | 步骤 9 | 超时/DNS/连接拒绝 → "无法连接，请检查 Base URL 和网络"；401/403 → "API 返回错误（状态码 XXX），请检查 API Key" |
| AC-45 | 步骤 9 | 测试期间按钮显示加载状态并禁用；结果返回后恢复 |

---

## 决策与风险

### 决策点

| # | 决策 | 选项 | 推荐 |
|---|------|------|------|
| D1 | 供应商 key 命名风格 | (A) kebab-case (`ant-ling`)；(B) 与 pi 官方 providers.md 严格一致 | **B**：查看 pi 官方文档确认准确 key 名；当前假设使用合理 slug |
| D2 | 拖拽实现方案 | (A) HTML5 Drag & Drop 原生 API；(B) 引入 vuedraggable 等库 | **A**：避免新增依赖；本项目规模小，原生 DnD 足够；方案/模型列表项数有限 |
| D3 | 搜索/过滤期间允许拖拽 | (A) 禁止，需清空搜索后拖拽；(B) 允许，但排序影响全集 | **A**：实现简单，语义清晰，避免用户困惑 |
| D4 | 连通性测试返回格式 | (A) 总是返回 `(message, error)`，error 仅前置校验；(B) 统一返回 `(message, error)` | **A**：与前端 toast 用法一致（error 显示为红色，message 按类型着色） |
| D5 | 模型预设数据来源 | (A) 查阅各模型官方文档填入准确值；(B) 使用常见社区值 | **A**：需在实施时逐模型核实 window/tokens 数值 |

### 风险

| # | 风险 | 影响 | 缓解 |
|---|------|------|------|
| R1 | 13 个新供应商的 key 名可能与 pi 官方不一致 | AC-01 失败 | 实施前查阅 pi 官方 `providers.md` 确认准确 key 名 |
| R2 | 前端改动量大（7 个步骤涉及 SchemeEditor.vue），可能引入回归 | 已有功能的验收失效 | 分步骤 CR，每步手动验收核心流程（添加/删除/激活） |
| R3 | HTML5 DnD 在 Windows WebView2 中的行为可能异常 | AC-22/AC-23 | Wails v2 使用 Edge WebView2，HTML5 DnD 支持良好；实施后测试 |
| R4 | 模型预设的 contextWindow/maxTokens 值不准确 | AC-36 填写值不正确 | 实施时查阅各模型官方文档确认，并将来源记录在注释中 |

---

## 未解决 Blocker

经核查源码、调用链和产品不变量，**当前无 blocker**。

### 已排查的潜在问题

| 潜在问题 | 排查结论 |
|----------|----------|
| AC-25 provider 顺序在 models.json 中是否可保持 | `modelsJSONOutput.Providers` 为 `map[string]providerJSON`，Go `json.Marshal` 按键排序。但 AC-25 原文"providers 对象内的属性"应解读为每个 provider 对象内部的字段（apiKey/baseUrl/models 等，struct 字段顺序固定），非 provider 条目在 map 中的顺序。`models` 数组为切片，顺序天然保持。JSON 对象的 key 顺序无实际语义意义（pi 按 key 查找，非按位置）。**不构成 blocker**。 |
| AC-03 ValidAPITypes 是否已含 azure-openai-responses | `builtin.go:33` 已包含，**AC-03 已满足**。 |
| 后端 FetchProviderModels 是否需要新增 API 类型白名单 | 当前 `fetch.go` 无 API 类型限制（仅前端控制 `canFetchModels`），天然支持任意 API 类型发 GET `/models`。PRD 要求仅为前端可用性控制（AC-06/AC-07），**后端无需改动**。 |

---

*计划生成时间：2025-07-19 | 基于 PRD v1 | 证据覆盖 builtin.go / api.go / fetch.go / serializer.go / store.go / SchemeEditor.vue / types.ts / wails/api.ts*