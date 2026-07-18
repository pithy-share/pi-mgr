# 架构概览

**阅读时机**：涉及项目结构、模块边界、API 绑定、前端路由或跨层数据流时。  
**可核验依据**：`task/pi-provider-model-manager/plan.md` Phase 1–3。

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
│  └── builtin.go 内置供应商目录  │
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

### 目录查询

| 方法 | 输入 | 输出 |
|---|---|---|
| `ListBuiltInProviders()` | — | `[]BuiltInProvider` |
| `ListAPITypes()` | — | `[]string`（`openai-completions`, `anthropic-messages`, `google-generative-ai`） |

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
    → 返回成功/错误
```