# PRD：功能完善增强

## 问题

pi-mgr v1 作为 Pi Coding Agent 的 provider/model 方案管理工具，在功能完整度上存在七个核心缺口：

1. **内置供应商目录缺失** — 当前硬编码 14 个内置供应商，但 pi 官方支持的常规 API key 供应商中约 13 个缺失（Ant Ling、Vercel AI Gateway、ZAI Coding Plan、OpenCode Zen/Go、Kimi For Coding、MiniMax、Xiaomi MiMo 系列等）。用户只能以自定义供应商方式手动输入，缺乏内置供应商带来的 API 类型自动关联。Azure OpenAI、Cloudflare、Google Vertex 等需额外配置字段（resource name、account ID、project/location）的云供应商本期不纳入，下期扩展 Provider 数据模型后再支持。

2. **模型拉取缺少 azure-openai-responses** — `FetchProviderModels` 当前仅允许 `openai-completions` 和 `openai-responses`，同为 OpenAI 兼容的 `azure-openai-responses` 尚未支持。Anthropic 和 Google 的模型列表端点返回格式不同，需独立调研，本期不纳入。

3. **模型列表无搜索和过滤** — 单个供应商（如 OpenRouter）可能有上百个模型，当前仅展示顺序列表，缺少按 ID/名称搜索和按能力（reasoning、inputImage）过滤的能力。

4. **无法拖拽排序** — Provider 在当前方案内的顺序、Model 在 Provider 内的顺序均不可调整。`models.json` 中的顺序直接影响 pi 的模型选择 UI（Ctrl+L），用户无法控制模型展示优先级。

5. **缺少批量操作** — 无法多选模型后批量删除；FetchProviderModels 拉取之外，无其他批量导入途径（如粘贴 JSON 数组）。

6. **缺少已知模型预设** — 手动添加模型时，`contextWindow` 默认 256K、`maxTokens` 默认 64K，缺乏常见模型（GPT-4o、Claude Sonnet、Gemini 等）的预设值，用户每次需手动查找填写。

7. **缺少供应商连通性测试** — 用户配置 API key 和 Base URL 后，无法在 pi-mgr 内快速验证是否能连通该供应商的 API 端点。目前只能激活方案后到 pi 终端中实际调用模型才能知道配置是否正确，反馈周期太长。

## 目标与非目标

### 目标

- 补齐 pi 官方常规 API key 供应商（13 个）
- `FetchProviderModels` 新增 `azure-openai-responses` 支持
- 为方案编辑器的模型列表增加搜索和过滤（文本搜索 + 能力过滤）
- 支持 Provider 列表拖拽排序（新增 `ReorderProviders` API 一次性持久化）
- 支持 Model 列表拖拽排序（新增 `ReorderModels` API 一次性持久化）
- 新增 `RemoveModels` API 支持模型批量删除
- 增加模型批量导入 JSON 功能
- 为手动添加模型提供已知模型预设列表（11 个，前端硬编码）
- 供应商连通性测试：对 OpenAI 兼容端点发送轻量请求，验证 API 可达性和 key 是否有效

### 非目标

- 不管理 `auth.json`、`settings.json`、`keybindings.json`、`prompts/`、`skills/`、`extensions/`
- 不做自动化的 API key 有效性校验（连通性测试为手动触发，非自动校验）
- 不改变 `models.json` 激活写入格式或序列化逻辑
- 不修改方案导入/导出格式
- 不做方案级对比 / diff
- 不做拖拽之外的排序 UI（如排序下拉框、按字段排序）
- 云供应商（Azure OpenAI、Cloudflare AI Gateway / Workers AI、Google Vertex）本期不纳入内置列表

## 范围内外

### 范围内

- `builtin.go` 内置供应商常量扩展（13 个新增）
- `ValidAPITypes` 补充 `azure-openai-responses`
- 方案编辑器模型列表区域（搜索输入框 + 能力过滤下拉：reasoning、inputImage）
- Provider 列表拖拽排序 + 新增 `ReorderProviders` Wails API
- Model 列表拖拽排序 + 新增 `ReorderModels` Wails API
- Provider 和 Model 顺序在 `schemes.json` 中按数组位置持久化
- 新增 `RemoveModels` Wails API（批量删除）
- 模型多选复选框 + 批量删除（含确认）
- 模型批量导入 JSON（粘贴 JSON 数组）
- `Model` 预设目录（前端硬编码 11 个知名模型）
- 供应商连通性测试（手动触发，仅对 OpenAI 兼容 API 类型发送 GET 请求验证连通性）

### 范围外

- 订阅制供应商（OpenAI Codex、Claude Pro/Max、GitHub Copilot、xAI Grok 订阅、Radius）
- 云供应商的额外配置字段（Azure OpenAI resource name、Cloudflare account ID、Google Vertex project/location）
- 自动检测自定义供应商的 API 类型
- 桌面应用的深色主题
- 自动更新机制
- 撤销 / 重做

## 验收标准

### 内置供应商扩展

| 编号 | 验收标准 |
|------|----------|
| AC-01 | `ListBuiltInProviders()` 返回的内置供应商列表包含以下新增条目（名称与 pi 官方 `providers.md` 一致）：Ant Ling、Vercel AI Gateway、ZAI Coding Plan (Global)、ZAI Coding Plan (China)、OpenCode Zen、OpenCode Go、Kimi For Coding、MiniMax、MiniMax (China)、Xiaomi MiMo、Xiaomi MiMo Token Plan (China)、Xiaomi MiMo Token Plan (Amsterdam)、Xiaomi MiMo Token Plan (Singapore) |
| AC-02 | 新增内置供应商的 `APIType` 全部映射为 `openai-completions` |
| AC-03 | `ValidAPITypes` 补充 `azure-openai-responses` |
| AC-04 | 新增内置供应商在方案编辑器的“添加内置供应商”下拉列表中可选，按名称排序展示 |
| AC-05 | 同 Scheme 下重复添加同一内置供应商（相同 Key）时返回“该内置供应商已添加”错误 |

### 模型拉取扩展

| 编号 | 验收标准 |
|------|----------|
| AC-06 | `FetchProviderModels` 支持的 API 类型新增 `azure-openai-responses`：当供应商的 `apiType` 为该值时，可发起模型拉取请求 |
| AC-07 | 前端 `canFetchModels` 计算逻辑同步更新，`azure-openai-responses` 供应商的“拉取模型”按钮可用 |

### 模型搜索与过滤

| 编号 | 验收标准 |
|------|----------|
| AC-08 | 方案编辑器右侧模型列表顶部显示搜索输入框（placeholder：“搜索模型 ID 或名称”） |
| AC-09 | 输入搜索关键词后，模型列表实时过滤：仅显示 ID 或 Name 包含关键词（不区分大小写）的模型 |
| AC-10 | 清空搜索关键词后，模型列表恢复完整展示 |
| AC-11 | 无匹配结果时显示“无匹配模型”空状态提示 |
| AC-12 | 搜索后仍可正常执行模型的添加、编辑、删除操作，操作完成后列表按当前搜索条件重新过滤 |
| AC-13 | 搜索输入框右侧提供能力过滤下拉（可选）：选项为“reasoning”和“inputImage”，默认为“全部”（不筛选） |
| AC-14 | 选中某个能力过滤后，模型列表仅显示对应字段为 `true` 的模型；文本搜索与能力过滤为 AND 关系（同时满足） |

### Provider 排序 API

| 编号 | 验收标准 |
|------|----------|
| AC-15 | 新增 Wails API `ReorderProviders(schemeID string, orderedKeys []string) error`，传入方案 ID 和 Provider key 顺序数组 |
| AC-16 | 后端校验 `orderedKeys` 与方案当前 Provider key 集合完全一致（不可增删），校验失败返回“供应商列表不一致”错误 |
| AC-17 | 校验通过后按 `orderedKeys` 重排 Provider 数组并一次性原子写入 `schemes.json` |
| AC-18 | 前端拖拽排序完成后调用 `ReorderProviders` 持久化 |

### Model 排序 API

| 编号 | 验收标准 |
|------|----------|
| AC-19 | 新增 Wails API `ReorderModels(schemeID string, providerKey string, orderedIDs []string) error`，传入方案 ID、供应商 key 和 Model ID 顺序数组 |
| AC-20 | 后端校验 `orderedIDs` 与当前 Provider 的 Model ID 集合完全一致（不可增删），校验失败返回“模型列表不一致”错误 |
| AC-21 | 校验通过后按 `orderedIDs` 重排 Model 数组并一次性原子写入 `schemes.json` |

### 拖拽排序（前端交互）

| 编号 | 验收标准 |
|------|----------|
| AC-22 | 方案编辑器左侧供应商列表支持拖拽重排：用户拖拽某供应商到新位置后，触发 `ReorderProviders` 持久化 |
| AC-23 | 方案编辑器右侧模型列表支持拖拽重排：用户拖拽某模型到新位置后，触发 `ReorderModels` 持久化 |
| AC-24 | 拖拽排序后的校验保持原有规则（Key 唯一、ID 唯一），不因排序引入新的校验错误 |
| AC-25 | 激活方案后写入 `models.json` 时，providers 对象内的属性和 models 数组保持用户在 pi-mgr 中设定的顺序 |

### 批量删除 API

| 编号 | 验收标准 |
|------|----------|
| AC-26 | 新增 Wails API `RemoveModels(schemeID string, providerKey string, modelIDs []string) (int, error)`，传入方案 ID、供应商 key 和待删除 Model ID 数组；返回成功删除数 |
| AC-27 | 后端逐条执行删除：某条因不存在或其他原因失败时跳过，继续处理剩余条目；全部处理完后返回成功计数；仅当全部失败时返回 error |
| AC-28 | 方案中不存在的 Model ID 不被视为错误，在成功计数中不计入（静默跳过） |

### 批量操作（前端交互）

| 编号 | 验收标准 |
|------|----------|
| AC-29 | 模型列表每行左侧显示复选框；列表头部显示全选复选框（搜索/过滤激活时，全选仅选中当前可见的匹配模型） |
| AC-30 | 选中一个或多个模型后，列表上方出现“删除所选（N）”按钮；未选中任何模型时按钮隐藏或禁用 |
| AC-31 | 点击批量删除后弹出确认对话框（“确认删除选中的 N 个模型？”），确认后调用 `RemoveModels` API 执行删除，刷新列表 |
| AC-32 | 模型列表上方提供“批量导入 JSON”按钮，点击后弹出文本框，用户粘贴 JSON 数组 `[{"id": "...", "name": "...", ...}, ...]` |
| AC-33 | 批量导入 JSON 解析成功后调用 `ImportProviderModels` 逐条导入，按 ID 跳过已存在的模型；返回导入成功数和跳过数 |
| AC-34 | JSON 格式无效时在前端展示“JSON 格式错误”提示，不发起任何导入操作 |

### 模型预设

| 编号 | 验收标准 |
|------|----------|
| AC-35 | 手动添加模型时，提供一个“预设模型”下拉列表（可选，默认为“自定义”） |
| AC-36 | 选择某个预设模型后，自动填充该模型的 `ID`、`Name`、`reasoning`、`inputText`、`inputImage`、`contextWindow`、`maxTokens` 字段；Cost 字段保持 0（用户自行填写） |
| AC-37 | 下拉列表包含至少以下模型的预设：GPT-4o、GPT-4o-mini、GPT-4.1、GPT-4.1-mini、Claude Sonnet 4.5、Claude Opus 4.5、Claude Haiku 4.5、Gemini 2.5 Pro、Gemini 2.5 Flash、DeepSeek V3、DeepSeek R1 |
| AC-38 | 选择预设后，用户可修改任意字段（预设仅为快捷填充，不锁定字段） |
| AC-39 | 预设数据为前端硬编码常量，不与后端 API 交互 |

### 供应商连通性测试

| 编号 | 验收标准 |
|------|----------|
| AC-40 | 方案编辑器供应商详情区域（API Key / Base URL 配置区旁）提供“测试连接”按钮；当供应商 `apiType` 为以下类型时按钮可用：`openai-completions`、`openai-responses`、`azure-openai-responses`；其他类型禁用并显示 tooltip“该 API 类型暂不支持连通性测试” |
| AC-41 | `baseUrl` 为空时按钮禁用，tooltip 显示“请先配置 Base URL” |
| AC-42 | 点击“测试连接”后发起 HTTP GET `{baseURL}/models`，携带 `Authorization: Bearer {apiKey}` 头（apiKey 为空时不发送 Authorization 头），超时 10 秒 |
| AC-43 | 返回 2xx 状态码时显示绿色成功提示“连接成功，API 可达” |
| AC-44 | 请求失败时按错误类型返回中文提示：无法连接（超时、DNS 失败、连接拒绝等）→“无法连接，请检查 Base URL 和网络”；非 2xx（含 401/403）→“API 返回错误（状态码 XXX），请检查 API Key”；其他 → 展示原始错误摘要 |
| AC-45 | 测试期间按钮显示加载状态并禁用，防止重复点击；结果返回后恢复 |

## 决议记录

以下为已确认的产品决议（不纳入验收）：

1. **云供应商**（Azure OpenAI、Cloudflare AI Gateway / Workers AI、Google Vertex）：本期不加入内置列表，下期扩展 Provider 数据模型支持额外配置字段后再纳入。
2. **FetchProviderModels 扩展**：本期仅增加 `azure-openai-responses`；Anthropic Messages / Google Gemini 端点格式不同，推迟到下期调研。
3. **Provider 排序**：新增专用 API `ReorderProviders(schemeID, orderedKeys)`，一次性原子写入。
4. **模型预设**：以 11 个模型为基础，前端硬编码，后续按社区反馈补充。
5. **连通性测试**：仅支持 OpenAI 兼容端点（`openai-completions` / `openai-responses` / `azure-openai-responses`），手动触发，非自动校验。与 AGENTS.md 中“不校验 API key 有效性”不冲突——该校验为自动挡，连通性测试为手动挡。
6. **Model 排序**：新增专用 API `ReorderModels(schemeID, providerKey, orderedIDs)`，与 Provider 排序一致的一次性原子写入方案。
7. **批量删除**：新增专用 API `RemoveModels(schemeID, providerKey, modelIDs)`，返回成功删除数，尽力而为语义。