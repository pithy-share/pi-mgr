# 架构概览

**阅读时机**：涉及项目结构、模块边界、API 绑定、前端路由、模型拉取/导入、SSH 同步或跨层数据流时。  
**可核验依据**：`app.go`, `api.go`, `fetch.go`, `ssh_sync.go`, `ssh_settings.go` 中导出的 `App` 方法；`frontend/src/presets.ts`（模型预设）。

## 分层

```
┌──────────────────────────────┐
│  Vue 3 + TypeScript 前端      │  用户交互，仅通过 Wails API 访问后端
├──────────────────────────────┤
│  Wails API 桥接层 (app.go)    │  导出方法，参数校验，调用业务逻辑
├──────────────────────────────┤
│  业务逻辑层                    │
│  ├── models.go  数据模型       │
│  ├── store.go   方案持久化      │
│  ├── serializer.go 输出序列化  │
│  ├── activate.go 激活写入      │
│  ├── validate.go 校验规则      │
│  ├── builtin.go 内置供应商目录  │
│  ├── fetch.go   模型列表拉取    │
│  ├── ssh_sync.go SSH 配置同步   │
│  └── ssh_settings.go SSH 地址持久化 │
├──────────────────────────────┤
│  本地文件系统                   │
│  ├── %APPDATA%/pi-mgr/schemes.json  (工具自身数据)  │
│  └── %USERPROFILE%/.pi/agent/models.json  (pi 配置) │
└──────────────────────────────┘
```

## API 绑定（Wails 桥接）

前端通过 Wails runtime 调用 Go 端导出方法，所有状态变更必须经过这些方法。前端绝不直接读写磁盘。

### 方案 CRUD

| 方法 | 输入 | 输出 | 副作用 |
|---|---|---|---|
| `ListSchemes()` | — | `[]Scheme` | 读取 schemes.json |
| `CreateScheme(name)` | `name string` | `*Scheme, error` | 写入 schemes.json |
| `UpdateScheme(scheme)` | `Scheme` | `error` | 写入 schemes.json |
| `DeleteScheme(id)` | `id string` | `error` | 写入 schemes.json（含确认） |
| `DuplicateScheme(id)` | `id string` | `*Scheme, error` | 写入 schemes.json，名称加" - 副本" |
| `ActivateScheme(id)` | `id string` | `error` | 序列化并覆盖写入 models.json |

### 供应商管理（方案内）

| 方法 | 输入 | 输出 | 副作用 |
|---|---|---|---|
| `AddBuiltInProvider(schemeID, providerKey, apiKey, baseUrl)` | 方案 ID、内置 key、API key、可选 baseUrl | `error` | 写入 schemes.json，校验重复 |
| `AddCustomProvider(schemeID, key, baseUrl, apiType, apiKey)` | 方案 ID、自定义 key、baseUrl、API 类型、可选 apiKey | `error` | 写入 schemes.json，校验重复和必填 |
| `UpdateProvider(schemeID, provider)` | 方案 ID、Provider | `error` | 写入 schemes.json |
| `RemoveProvider(schemeID, providerKey)` | 方案 ID、provider key | `error` | 写入 schemes.json（含确认） |

### 模型管理（供应商内）

| 方法 | 输入 | 输出 | 副作用 |
|---|---|---|---|
| `AddModel(schemeID, providerKey, model)` | 方案 ID、provider key、Model | `error` | 写入 schemes.json，校验 ID 重复 |
| `UpdateModel(schemeID, providerKey, model)` | 方案 ID、provider key、Model | `error` | 写入 schemes.json |
| `RemoveModel(schemeID, providerKey, modelID)` | 方案 ID、provider key、model ID | `error` | 写入 schemes.json（含确认） |

### 排序

| 方法 | 输入 | 输出 | 副作用 |
|---|---|---|---|
| `ReorderProviders(schemeID, orderedKeys)` | 方案 ID、按新顺序排列的 provider key 切片 | `error` | 写入 schemes.json；set-equality 校验（长度、存在性、重复），失败返回 `"供应商列表不一致"` |
| `ReorderModels(schemeID, providerKey, orderedIDs)` | 方案 ID、provider key、按新顺序排列的 model ID 切片 | `error` | 写入 schemes.json；set-equality 校验，失败返回 `"模型列表不一致"` |

**排序契约**：Reorder API 仅重排切片顺序，不修改 Key/ID 或其他字段；唯一性约束由现有校验保证。Go 切片顺序决定序列化输出顺序。

### 批量操作

| 方法 | 输入 | 输出 | 副作用 |
|---|---|---|---|
| `RemoveModels(schemeID, providerKey, modelIDs)` | 方案 ID、provider key、待删除 model ID 切片 | `(int, error)` | 写入 schemes.json；返回实际删除数量；不存在的 ID 自然跳过；`removed == 0` 时不写入返回 `(0, nil)` |

### 连通性测试

| 方法 | 输入 | 输出 | 副作用 |
|---|---|---|---|
| `TestProviderConnectivity(schemeID, providerKey)` | 方案 ID、provider key | `(string, error)` | **无持久化**；HTTP GET `{baseURL}/models`，Bearer 认证（空 key 不发），10s 超时 |

**连通性测试契约**：
- 仅支持 `openai-completions`、`openai-responses`、`azure-openai-responses` 三种 API 类型，其他类型返回 `"该 API 类型暂不支持连通性测试"`
- BaseURL 为空时返回 `"请先配置 Base URL"`
- 2xx 响应 → `"连接成功，API 可达"`
- 网络错误 → `"无法连接，请检查 Base URL 和网络"`
- 非 2xx → `"API 返回错误（状态码 %d），请检查 API Key"`
- 内置供应商通过查找 `BuiltInProviders` 获取实际 APIType

### 目录查询

| 方法 | 输入 | 输出 |
|---|---|---|
| `ListBuiltInProviders()` | — | `[]BuiltInProvider` |
| `ListAPITypes()` | — | `[]string`（9 种有效 API 类型） |
| `GetActiveSchemeID()` | — | `string`（空字符串表示无激活方案） |

### 模型拉取与导入

| 方法 | 输入 | 输出 | 副作用 |
|---|---|---|---|
| `FetchProviderModels(schemeID, providerKey)` | 方案 ID、provider key | `([]Model, error)` | HTTP GET `{baseURL}/models`，仅返回 id 和 name |
| `ImportProviderModels(schemeID, providerKey, models)` | 方案 ID、provider key、Model 切片 | `(int, error)` | 写入 schemes.json，按 ID 跳过已存在项 |

### 方案导入/导出

| 方法 | 输入 | 输出 | 副作用 |
|---|---|---|---|
| `ExportSchemes()` | — | `error` | 保存文件对话框 → 原子写入用户选定路径 |
| `ImportSchemes()` | — | `error` | 打开文件对话框 → 合并写入 schemes.json，全量校验 |

### SSH 同步

| 方法 | 输入 | 输出 | 副作用 |
|---|---|---|---|
| `SaveSSHAddress(address)` | `address string` | `error` | 写入 `%APPDATA%/pi-mgr/settings.json` |
| `LoadSSHAddress()` | — | `(string, error)` | 读取 `settings.json` |
| `TestSSHConnection(address)` | `address string` | `SSHConnectionResult` | 执行 `ssh -p -o BatchMode=yes user@host exit`，15s 超时 |
| `SyncPiConfig(address)` | `address string` | `SyncResult` | scp + ssh 同步 models.json、settings.json、prompts/、skills/ 到远端 |

## 前端路由

| 路由 | 页面 | 职责 |
|---|---|---|
| `/` | 方案列表 | 展示方案，提供新建/编辑/复制/删除/激活操作 |
| `/scheme/:id` | 方案编辑器 | 左侧供应商列表，右侧供应商详情+模型列表 |

## 状态与数据流

```
用户操作 → 前端组件 → Wails API 调用 → 业务逻辑层
    → 校验 (validate.go)
    → 更新内存 Scheme 结构
    → 持久化 (store.go → schemes.json)
    → 返回结果给前端
```

激活流程：
```
用户点击激活 → ActivateScheme(id)
    → LoadSchemes → 找到目标 Scheme
    → SerializeToModelsJSON(scheme)
    → 解析 ~/.pi/agent/ 目录（不存在则创建）
    → 覆盖写入 models.json
    → SaveActiveSchemeID(id) → 写入 active.json
    → 返回成功/错误
```

## 模型拉取（FetchProviderModels）

`fetch.go` 通过 HTTP GET 从 `{provider.BaseURL}/models` 拉取 OpenAI 兼容的模型列表。

- **请求**：GET `{baseURL}/models`，`Authorization: Bearer {apiKey}`（apiKey 为空时不发送 Authorization header）
- **超时**：10 秒
- **响应解析**：期望 `{"data": [{"id": "...", "name": "..."}]}`，仅填充 ID 和 Name
- **默认值**：`Reasoning=false, InputText=true, InputImage=false, ContextWindow=256000, MaxTokens=64000, Cost*=0`
- **错误**：网络不可达、非 200 状态码、JSON 解析失败、缺少 data 字段 → 均返回中文错误

## 模型预设

`frontend/src/presets.ts` 导出 `MODEL_PRESETS` 常量数组，包含 11 个常用模型的默认参数。预设仅在**添加模型**时可选（编辑模式不显示），选择后自动填充 id/name/reasoning/inputText/inputImage/contextWindow/maxTokens 字段，Cost 字段保持 0。预设为纯前端硬编码，无后端 API 调用。

**重验条件**：模型参数随供应商更新变化时需手动同步。

## SSH 同步

`ssh_sync.go` + `ssh_settings.go` 实现通过 SSH/SCP 将本地 pi 配置同步到远程机器。

### 地址解析

`parseSSHAddress(address)` 解析 `user@host[:port]` 格式，端口默认 22。

### 连接测试（TestSSHConnection）

执行 `ssh -p {port} -o ConnectTimeout=10 -o BatchMode=yes {user}@{host} exit`（15s 超时），按错误模式分类返回中文消息：超时、拒绝、主机密钥验证失败、认证失败、无法到达、DNS 解析失败。

### 配置同步（SyncPiConfig）

1. 解析地址 → 预检查 SSH 连通性 → 创建远程 `~/.pi/agent/` 目录
2. 逐项同步（每项独立，失败不影响其他项）：
   - `settings.json` / `models.json` / `keybindings.json`：`scp` 直接传输
   - `prompts/` / `skills/`：`scp -r` 到临时目录 → `ssh rm -rf + mv` 原子替换（模拟 rsync --delete）
3. 整体状态：全部成功 → `"success"`，部分失败 → `"partial"`，全部失败 → `"failed"`

### SSH 地址持久化

`%APPDATA%/pi-mgr/settings.json` 存储 `{"sshAddress": "user@host[:port]"}`，原子写入（temp + rename）。

### Windows 适配

所有 `exec.CommandContext` 通过 `syscall.SysProcAttr{HideWindow: true}` 隐藏终端窗口。