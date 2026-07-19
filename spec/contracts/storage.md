# 存储格式与序列化契约

**阅读时机**：改动 schemes.json 持久化、settings.json 应用设置、models.json 输出格式、数据模型字段、序列化逻辑或内置供应商目录时。  
**可核验依据**：
- 数据模型：`models.go`（Scheme, Provider, Model, BuiltInProvider）
- 存储层：`store.go`（LoadSchemes, SaveSchemes, GetScheme, 活跃方案追踪）
- 应用设置：`ssh_settings.go`（loadAppSettings, saveAppSettings, appSettings）
- 内置目录：`builtin.go`（BuiltInProviders, ValidAPITypes）
- 序列化器：`serializer.go`（SerializeToModelsJSON）
- 激活写入：`activate.go`（ActivateScheme）

## 数据模型（Go 结构体）

```go
type Scheme struct {
    ID        string     `json:"id"`        // UUID
    Name      string     `json:"name"`      // 用户可见名称
    Providers []Provider `json:"providers"` // 有序
}

type Provider struct {
    Key     string  `json:"key"`     // provider ID（内置的如 "openai"，自定义的为用户输入）
    Name    string  `json:"name"`    // 显示名（自定义时有值，内置为空）
    BuiltIn bool    `json:"builtIn"` // true=内置，false=自定义
    APIKey  string  `json:"apiKey"`
    BaseURL string  `json:"baseUrl"`
    APIType string  `json:"apiType"` // 仅自定义：openai-completions / anthropic-messages / google-generative-ai
    Models  []Model `json:"models"`
}

type Model struct {
    ID            string  `json:"id"`
    Name          string  `json:"name"`
    Reasoning     bool    `json:"reasoning"`
    InputText     bool    `json:"inputText"`
    InputImage    bool    `json:"inputImage"`
    ContextWindow int     `json:"contextWindow"`
    MaxTokens     int     `json:"maxTokens"`
    CostInput     float64 `json:"costInput"`
    CostOutput    float64 `json:"costOutput"`
    CostCacheRead float64 `json:"costCacheRead"`
    CostCacheWrite float64 `json:"costCacheWrite"`
}

type BuiltInProvider struct {
    Key     string `json:"key"`
    Name    string `json:"name"`
    APIType string `json:"apiType"`
}
```

**唯一性约束**（存储层保证）：
- Scheme.ID 全局唯一
- 同一 Scheme 内 Provider.Key 唯一
- 同一 Provider 内 Model.ID 唯一

## schemes.json 持久化

- **路径**：`%APPDATA%\pi-mgr\schemes.json`
- **格式**：`[]Scheme` 的 JSON 数组
- **原子写入**：先写临时文件再 rename，防止写入中断导致数据损坏；跨卷 rename 失败时回退直接写并清理临时文件
- **初始化**：文件不存在或为空时返回空切片
- **并发**：单进程单写者，不做锁

### 核心函数（`store.go`）

| 函数 | 签名 | 说明 |
|---|---|---|
| `storePath()` | `string` | 返回 `%APPDATA%/pi-mgr/schemes.json` 路径 |
| `LoadSchemes()` | `([]Scheme, error)` | 读取全部方案，文件不存在返回空切片 |
| `SaveSchemes(schemes)` | `error` | 原子写入（temp + rename） |
| `GetScheme(id)` | `(*Scheme, error)` | 按 ID 查找单个方案，未找到返回 error |
| `newUUID()` | `string` | `crypto/rand` 实现 UUID v4 |

### 活跃方案追踪（`active.json`）

激活操作后记录当前激活的方案 ID，存储于 `%APPDATA%/pi-mgr/active.json`。

| 函数 | 说明 |
|---|---|
| `GetActiveSchemeID() (string, error)` | 读取活跃方案 ID，文件不存在返回空字符串 |
| `SaveActiveSchemeID(id string) error` | 保存活跃方案 ID（原子写入） |
| `ClearActiveSchemeID()` | 删除 active.json（方案被删除时调用） |

## 内置供应商目录（硬编码）

```go
var BuiltInProviders = []BuiltInProvider{
    {Key: "openai",                    Name: "OpenAI",                            APIType: "openai-completions"},
    {Key: "anthropic",                 Name: "Anthropic",                         APIType: "anthropic-messages"},
    {Key: "deepseek",                  Name: "DeepSeek",                          APIType: "openai-completions"},
    {Key: "google",                    Name: "Google Gemini",                     APIType: "google-generative-ai"},
    {Key: "mistral",                   Name: "Mistral",                           APIType: "mistral-conversations"},
    {Key: "groq",                      Name: "Groq",                              APIType: "openai-completions"},
    {Key: "xai",                       Name: "xAI",                               APIType: "openai-completions"},
    {Key: "openrouter",                Name: "OpenRouter",                        APIType: "openai-completions"},
    {Key: "together",                  Name: "Together AI",                       APIType: "openai-completions"},
    {Key: "fireworks",                 Name: "Fireworks",                         APIType: "openai-completions"},
    {Key: "cerebras",                  Name: "Cerebras",                          APIType: "openai-completions"},
    {Key: "bedrock",                   Name: "Amazon Bedrock",                    APIType: "bedrock-converse-stream"},
    {Key: "nvidia",                    Name: "NVIDIA NIM",                        APIType: "openai-completions"},
    {Key: "huggingface",              Name: "Hugging Face",                      APIType: "openai-completions"},
    {Key: "ant-ling",                  Name: "Ant Ling",                          APIType: "openai-completions"},
    {Key: "vercel-ai-gateway",         Name: "Vercel AI Gateway",                 APIType: "openai-completions"},
    {Key: "zai-coding-plan-global",    Name: "ZAI Coding Plan (Global)",          APIType: "openai-completions"},
    {Key: "zai-coding-plan-china",     Name: "ZAI Coding Plan (China)",           APIType: "openai-completions"},
    {Key: "opencode-zen",              Name: "OpenCode Zen",                      APIType: "openai-completions"},
    {Key: "opencode-go",               Name: "OpenCode Go",                       APIType: "openai-completions"},
    {Key: "kimi-for-coding",           Name: "Kimi For Coding",                   APIType: "openai-completions"},
    {Key: "minimax",                   Name: "MiniMax",                           APIType: "openai-completions"},
    {Key: "minimax-china",             Name: "MiniMax (China)",                   APIType: "openai-completions"},
    {Key: "xiaomi-mimo",               Name: "Xiaomi MiMo",                       APIType: "openai-completions"},
    {Key: "xiaomi-mimo-china",         Name: "Xiaomi MiMo Token Plan (China)",    APIType: "openai-completions"},
    {Key: "xiaomi-mimo-amsterdam",     Name: "Xiaomi MiMo Token Plan (Amsterdam)",APIType: "openai-completions"},
    {Key: "xiaomi-mimo-singapore",     Name: "Xiaomi MiMo Token Plan (Singapore)",APIType: "openai-completions"},
}
```

**自定义供应商可用 API 类型**（见 `builtin.go` `ValidAPITypes`）：`openai-completions`、`anthropic-messages`、`openai-responses`、`azure-openai-responses`、`openai-codex-responses`、`mistral-conversations`、`google-generative-ai`、`google-vertex`、`bedrock-converse-stream`。

**重验条件**：pi 官方 providers 文档更新增删内置供应商时需同步更新。

## models.json 序列化规则

`SerializeToModelsJSON(scheme *Scheme) ([]byte, error)` 生成 pi 可消费的 `models.json` 格式。

### 输出结构

```json
{
  "providers": {
    "<provider-key>": { ... }
  }
}
```

### Provider 级规则

| 条件 | 输出行为 |
|---|---|
| 内置供应商，无 APIKey、无 BaseURL、无 Models | **跳过**，不在输出中出现 |
| 内置供应商，有 APIKey 或 BaseURL 或 Models | 生成条目，仅包含非空字段及 models（如有） |
| 内置供应商 | **禁止**输出 `api` 字段 |
| 自定义供应商 | 必须输出 `baseUrl`、`api`，可选 `apiKey`（非空时） |
| 自定义供应商 | 必须输出 `models` 数组 |

### Model 级规则

| 字段 | 输出规则 |
|---|---|
| `id` | 始终输出 |
| `name` | 仅当与 `id` 不同时输出 |
| `reasoning` | 仅当为 `true` 时输出（省略 `false`） |
| `input` | 从 InputText/InputImage 构建数组；若仅为 `["text"]`（默认）则省略 |
| `contextWindow` | 仅当 ≠ 0 时输出（含默认值 256000） |
| `maxTokens` | 仅当 ≠ 0 时输出（含默认值 64000） |
| `cost` | 仅当任一 cost 子字段 ≠ 0 时输出；cost 内省略零值字段 |

### 内置供应商 models 合并

内置供应商的自定义 model 与 pi 内置 model 按 `id` upsert：同名 ID 覆盖内置定义，新 ID 追加到内置列表。输出到 models.json 时，仅输出用户自定义的 models（pi 运行时合并内置 models）。

## 激活写入

`ActivateScheme(scheme *Scheme) error`（`activate.go`）：
1. 调用 `SerializeToModelsJSON(scheme)` 获取 JSON
2. 解析 `%USERPROFILE%\.pi\agent\` 目录，不存在则创建
3. 覆盖写入 `models.json`（非原子，目标路径可能跨文件系统）
4. 写入失败返回错误
5. 上层 `App.ActivateScheme(id)` 额外调用 `SaveActiveSchemeID(id)` 写入 `active.json`

**注意**：不改动 `schemes.json`，激活是纯读+输出操作。

## 应用设置（settings.json）

`ssh_settings.go` 管理应用级设置，存储于 `%APPDATA%/pi-mgr/settings.json`，独立于 `schemes.json`。

### 数据模型

```go
type appSettings struct {
    SSHAddress string `json:"sshAddress"` // user@host[:port] 格式
}
```

### 持久化

- **路径**：`%APPDATA%\pi-mgr\settings.json`
- **格式**：`{"sshAddress": "..."}` 单对象 JSON
- **原子写入**：同 schemes.json（temp + rename，跨卷回退直接写）
- **初始化**：文件不存在或解析失败时返回空 `appSettings{}`
- **方法**：`loadAppSettings()` / `saveAppSettings(settings)`

### Wails API

| 方法 | 说明 |
|---|---|
| `SaveSSHAddress(address string) error` | 持久化 SSH 地址 |
| `LoadSSHAddress() (string, error)` | 读取已保存的 SSH 地址 |
