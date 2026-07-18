# 实施结果：模型列表自动拉取

## 实施结果

**状态**：已完成。所有 17 条 PRD 验收标准均已实现，静态验证通过。

**验证方式说明**：基于 AGENTS.md 约束，仅进行静态核验（源码审阅），未运行 `go build`、`wails dev` 或单元测试。以下结论均基于源码和调用链证据。

---

## 改动清单

### 新增文件

| 文件 | 说明 |
|---|---|
| `fetch.go` (D:/src/pi-mgr/fetch.go) | 新增 `App.FetchProviderModels` 方法。HTTP GET `{baseURL}/models`，解析 OpenAI 兼容 List Models 响应，返回 `[]Model`（仅填充 id/name，其余零值）。使用 Go 标准库 `net/http`，超时 10s。 |

### 修改文件

| 文件 | 改动 |
|---|---|
| `api.go` | 新增 `App.ImportProviderModels(schemeID, providerKey, models []Model) (int, error)` — 批量导入模型，按 ID 去重跳过，单次原子写入，返回实际导入数量。插在 Model 管理区段 `RemoveModel` 之后、Export/Import 区段之前。 |
| `frontend/src/wails/api.ts` | `AppAPI` 接口新增 `FetchProviderModels` 和 `ImportProviderModels` 签名；`api()` fallback 中添加对应空存根。 |
| `frontend/src/views/SchemeEditor.vue` | 模板：模型列表标题行添加「⟳ 拉取模型列表」按钮（条件禁用 + title）、spinner、「选择要导入的模型」模态对话框（复选框列表、全选、确认导入）。脚本：新增 `fetchingModels`/`fetchedModels`/`showFetchDialog`/`fetchError`/`fetchSelectAll`/`fetchSelectedIds` 响应式状态；`canFetchModels`/`fetchButtonTitle` 计算属性；`getEffectiveAPIType` 辅助函数（内置供应商从目录查 APIType）；`handleFetchModels`/`handleFetchSelectAll`/`handleImportModels` 方法。 |
| `frontend/src/style.css` | 新增 `.spinner` 旋转动画样式。 |

### 未修改文件（确认不变）

`store.go`、`serializer.go`、`validate.go`、`builtin.go`、`activate.go`、`models.go`、`app.go`、`main.go`、`frontend/src/types.ts`、`SchemeList.vue` — 均未改动。

---

## 验收与验证

### 正常场景

| AC | 源码证据 | 调用链 |
|---|---|---|
| **AC-01** | `SchemeEditor.vue` 模板中 `:disabled="!canFetchModels"` 按钮 + `canFetchModels` 计算属性：`apiType === 'openai-completions'` 且 `prov.baseUrl` 非空时返回 `true`。内置供应商通过 `getEffectiveAPIType` → `getBuiltInAPIType(prov.key)` 从 `allBuiltIns` 查 APIType。 | 选择内置 `openai` 或自定义 `openai-completions` → 按钮可见可点 |
| **AC-02** | `fetch.go:16-75`：`FetchProviderModels` 构建 `GET {TrimRight(baseURL,"/")}/models`，`apiKey != ""` 时设 `Authorization: Bearer {apiKey}`。 | 点击按钮 → `handleFetchModels` → `api().FetchProviderModels(schemeID, providerKey)` → Go HTTP GET |
| **AC-03** | 模板 `v-if="showFetchDialog"` 模态框：遍历 `fetchedModels` 显示 `m.id` 和 `(m.name)`（条件），checkbox `v-model="fetchSelectedIds"`，全选 checkbox。 | 请求成功 → `fetchedModels = result; showFetchDialog = true` |
| **AC-04** | `api.go:365-406`：`ImportProviderModels` 用 `map[string]bool` 去重，仅追加不重复 ID。`fetch.go:66-76`：`Name` 优先取响应 `name`，空则复制 `id`。其余字段零值。`SchemeEditor.vue`：导入后调 `loadData()` 刷新。 | 勾选 → `handleImportModels` → `ImportProviderModels` → `SaveSchemes` → `loadData()` |
| **AC-05** | `loadData()` 刷新后模型出现在列表，编辑/删除按钮均存在，`UpdateModel` 路径未改动。 | 导入后 → 模型显示 → 点编辑 → 修改 → `UpdateModel` → 保存 |
| **AC-06** | `getEffectiveAPIType`：`!prov.builtIn` 时返回 `prov.apiType`，即自定义的 `openai-completions`。`canFetchModels` 逻辑同 AC-01。后端调用同一 `FetchProviderModels`。 | 同 AC-01~AC-05，供应商为自定义 `openai-completions` 类型 |

### 拉取失败 / 兜底

| AC | 源码证据 |
|---|---|
| **AC-07** | `fetch.go:46`：`client.Do(req)` 失败 → `fmt.Errorf("网络不可达: %w", err)`。前端 `catch` → `showToast?.(fetchError.value, 'error')`。手动「+ 添加模型」按钮始终可用。 |
| **AC-08** | `fetch.go:48-56`：`resp.StatusCode != 200` → 读 256B body → `fmt.Errorf("API 返回状态码 %d: %s", statusCode, body)`。前端同 AC-07 toast 显示。 |
| **AC-09** | `fetch.go:61-63`：JSON 解码失败 → `fmt.Errorf("响应解析失败: %w", err)`。`fetch.go:65-67`：`Data == nil` → `fmt.Errorf("响应解析失败: 响应中缺少 data 字段")`。前端同 AC-07。 |

### 重复与冲突

| AC | 源码证据 |
|---|---|
| **AC-10** | `api.go:388-395`：`if existing[m.ID] { continue }`。前端计算 `skipped = selectedModels.length - count`，toast 显示 "已导入 N 个模型，M 个因重复跳过"。 |
| **AC-11** | `api.go:400-401`：全部跳过时返回 `(0, nil)`。前端 `count === 0` → toast "无新增模型（所有选中模型均已存在）"。 |

### API 类型限制

| AC | 源码证据 |
|---|---|
| **AC-12** | `canFetchModels`：`apiType !== 'openai-completions' && apiType !== 'openai-responses'` → `false`。`fetchButtonTitle`：返回 "该 API 类型不支持自动拉取"。按钮 `:disabled="true"` + `:title="该 API 类型不支持自动拉取"`。 |
| **AC-13** | 同 AC-12，`google-generative-ai` 也匹配该条件。 |

### 边界与错误

| AC | 源码证据 |
|---|---|
| **AC-14** | `canFetchModels`：`if (!prov.baseUrl) return false`。`fetchButtonTitle`：返回 "请先配置 baseUrl"。按钮禁用。 |
| **AC-15** | `canFetchModels` 不检查 apiKey。`fetch.go:33-35`：`if prov.APIKey != "" { req.Header.Set(...) }`，空 key 时不设 Authorization 头。 |
| **AC-16** | `handleImportModels`：`if (fetchSelectedIds.value.length === 0) { showToast?.('请至少选择一个模型', 'error'); return }`。 |
| **AC-17** | `ImportProviderModels` 只追加模型到 `prov.Models`，不修改其他模型。模型编辑表单（`showAddModel`/`editingModel`）独立于导入流程，不受影响。`loadData()` 刷新供应商表单字段（与保存供应商后的行为一致），不影响模型编辑模态框内的未保存内容。 |

---

## 偏差与风险

### 计划偏差

| # | 偏差 | 原因 | 影响 |
|---|---|---|---|
| B1 | 无。所有改动严格遵循 `plan.md` 的 Step 1~4 和决策 D1~D5。 | — | — |

### 已知风险

| 风险 | 说明 | 缓解 |
|---|---|---|
| `openai-responses` 类型当前在 `ValidAPITypes` 和 `BuiltInProviders` 均不存在 | 按钮的条件判断已预留对 `openai-responses` 的匹配（Plan D3 要求），但该字符串在系统中不可选 | 未来 `ValidAPITypes` 扩增时自动生效，当前为无害的死代码 |
| 导入后 `loadData()` 会覆盖供应商表单字段 | 与保存供应商后 `loadData()` 的现有行为一致 | 供应商表单未保存内容会丢失，但导入本身并非数据丢失来源 |

### Spec 更新建议：无需

本功能未引入需要更新 `AGENTS.md` 或 `spec/` 目录的新知识。PRD 已明确声明网络调用豁免，改动完全在存储契约和架构定义的范围内扩展。

---

## 未解决 Blocker

无。Plan 子 Agent 审阅确认所有 AC 已实现，无可达 bug、无缺失联动、无范围外改动。

---

## 终态

- **PRD**：`D:/src/pi-mgr/task/auto-model-list/prd.md`
- **Plan**：`D:/src/pi-mgr/task/auto-model-list/plan.md`
- **Result**：`D:/src/pi-mgr/task/auto-model-list/result.md`
- **评审轮数**：1 轮（Plan 子 Agent 审阅）
- **剩余 Blocker**：0
