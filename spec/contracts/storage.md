# 存储格式与序列化契约

**阅读时机**：改动 schemes.json 持久化、models.json 输出格式、数据模型字段、序列化逻辑或内置供应商目录时。  
**可核验依据**：
- 数据模型：`task/pi-provider-model-manager/plan.md` Step 1.2（Go 结构体）
- 存储层：Step 1.3（schemes.json 路径与原子写入）
- 内置目录：Step 1.4（BuiltInProviders）
- 序列化器：Step 1.5（SerializeToModelsJSON 逻辑）
- 激活写入：Step 1.6（ActivateScheme）

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
- **原子写入**：先写临时文件再 rename，防止写入中断导致数据损坏
- **初始化**：文件不存在时返回空切片
- **并发**：单进程单写者，不做锁

## 内置供应商目录（硬编码）

```go
var BuiltInProviders = []BuiltInProvider{
    {Key: "openai",       Name: "OpenAI",           APIType: "openai-completions"},
    {Key: "anthropic",    Name: "Anthropic",         APIType: "anthropic-messages"},
    {Key: "deepseek",     Name: "DeepSeek",          APIType: "openai-completions"},
    {Key: "google",       Name: "Google Gemini",     APIType: "google-generative-ai"},
    {Key: "mistral",      Name: "Mistral",           APIType: "mistral-conversations"},
    {Key: "groq",         Name: "Groq",              APIType: "openai-completions"},
    {Key: "xai",          Name: "xAI",               APIType: "openai-completions"},
    {Key: "openrouter",   Name: "OpenRouter",        APIType: "openai-completions"},
    {Key: "together",     Name: "Together AI",       APIType: "openai-completions"},
    {Key: "fireworks",    Name: "Fireworks",         APIType: "openai-completions"},
    {Key: "cerebras",     Name: "Cerebras",          APIType: "openai-completions"},
    {Key: "bedrock",      Name: "Amazon Bedrock",    APIType: "bedrock-converse-stream"},
    {Key: "nvidia",       Name: "NVIDIA NIM",        APIType: "openai-completions"},
    {Key: "huggingface",  Name: "Hugging Face",      APIType: "openai-completions"},
}
```

**自定义供应商可用 API 类型**：`openai-completions`、`anthropic-messages`、`google-generative-ai`（前端的下拉列表，与内置的类型枚举是子集关系）。

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
| `contextWindow` | 仅当 ≠ 128000 时输出 |
| `maxTokens` | 仅当 ≠ 16384 时输出 |
| `cost` | 仅当任一 cost 子字段 ≠ 0 时输出；cost 内省略零值字段 |

### 内置供应商 models 合并

内置供应商的自定义 model 与 pi 内置 model 按 `id` upsert：同名 ID 覆盖内置定义，新 ID 追加到内置列表。输出到 models.json 时，仅输出用户自定义的 models（pi 运行时合并内置 models）。

## 激活写入

`ActivateScheme(scheme *Scheme) error`：
1. 调用 `SerializeToModelsJSON(scheme)` 获取 JSON
2. 解析 `%USERPROFILE%\.pi\agent\` 目录，不存在则创建
3. 覆盖写入 `models.json`
4. 写入失败返回错误

**注意**：不改动 `schemes.json`，激活是纯读+输出操作。
