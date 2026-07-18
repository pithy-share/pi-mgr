# PRD: Pi Provider & Model Manager

## 问题

Pi 用户在 Windows 上需要频繁切换模型和供应商（provider），目前只能手动编辑 `models.json` 配置文件。配置涉及内置供应商（如 OpenAI、Anthropic、DeepSeek 等只需 API key）、第三方中转站（需 baseUrl，可能兼容或不兼容内置 API 类型）、以及每个供应商下多个模型的增删改查。缺少一个桌面 GUI 工具来管理这些配置，导致切换操作繁琐且易出错。

## 目标与非目标

### 目标

- 提供一个 Windows 桌面应用（Go + Wails v2），可视化管理 pi 的 provider 和 model 配置
- 支持内置供应商的快速配置（只需填写 API key）
- 支持第三方中转站的新增（需填写 baseUrl + 选择 API 类型 + 手动添加 model）
- 支持对已有内置供应商 override baseUrl（如代理/中转场景）
- 工具自身保存多套配置方案，用户选择一套后一键覆盖写入 pi 的 `models.json`
- 对每个 provider 下的 model 支持完整的增删改查

### 非目标

- 不实现 model 的自动发现/拉取（v1 仅手动添加）
- 不管理 OAuth / subscription 类型的登录流程
- 不管理 `auth.json`（API key 直接写入 `models.json` 的 provider 配置中）
- 不实现配置的云端同步、导入导出、版本回滚
- 不管理 pi 的其他配置文件（如 settings.json、extensions 等）
- 不支持 macOS / Linux（仅 Windows）

## 范围内外

### 范围内

1. **配置方案管理**：工具自身存储多套配置方案（方案 = 一组 provider + model 配置），支持方案的创建、编辑、删除、复制
2. **激活方案**：用户选择一个方案，点击激活，工具将其覆盖写入 pi 的 `models.json`（`~/.pi/agent/models.json`）
3. **内置供应商配置**：硬编码 pi 支持的内置供应商列表（来源于 [pi providers 文档](https://pi.dev/docs/latest/providers)），用户选择后只需填写 API key（可选填写 override baseUrl）
4. **自定义供应商配置**：用户新建供应商，填写名称、baseUrl、API 类型（openai-completions / anthropic-messages / google-generative-ai）、API key
5. **Model 管理**：在任意供应商下，手动添加/编辑/删除 model，字段包括 id（必填）、name、reasoning、input、contextWindow、maxTokens、cost
6. **兼容模式**：自定义供应商可选择 API 类型（openai-completions / anthropic-messages / google-generative-ai），适配中转站兼容情况
7. **生成 models.json**：激活时将当前方案序列化为符合 pi models.json 规范的 JSON，覆盖写入

### 范围外

- 不管理 extension 形式的自定义 provider（TypeScript 代码注册）
- 不管理 `compat` 字段的高级配置（如 `supportsDeveloperRole`、`thinkingLevelMap` 等，v1 不涉及）
- 不管理 `headers`、`authHeader`、`oauth` 字段
- 不管理 `modelOverrides`（per-model overrides for built-in models）
- 不管理 `cost.tiers`（分层定价）
- 不检测/校验 API key 的有效性（不发起网络请求验证）
- 不检测 pi 是否已安装或 `models.json` 是否存在
- 不支持从现有 `models.json` 导入配置为方案（v1 从零创建）
- 不做多实例冲突检测（允许多实例运行，最后激活者生效）

## 验收标准

### 方案管理

- **AC-01**：启动应用后，显示已有配置方案列表（若为空则显示空状态提示）
- **AC-02**：用户可以创建新方案，输入方案名称，创建后方案列表中出现该条目
- **AC-03**：用户可以编辑方案名称
- **AC-04**：用户可以复制方案（生成副本，名称为原名称 + " - 副本"）
- **AC-05**：用户可以删除方案，删除前弹出确认对话框；删除后方案列表不再显示该条目
- **AC-06**：方案列表中的"激活"按钮/操作可将当前方案的配置写入 `~/.pi/agent/models.json`，覆盖已有内容
- **AC-07**：激活成功后给出明确的可感知提示（如 Toast / 状态栏消息）

### 内置供应商

- **AC-08**：在方案编辑界面，展示内置供应商选择列表。硬编码列表包含每个内置供应商的默认 `api` 类型映射，至少包括：openai（openai-completions）、anthropic（anthropic-messages）、deepseek（openai-completions）、google（google-generative-ai）、mistral（mistral-conversations）、groq（openai-completions）、xai（openai-completions）、openrouter（openai-completions）、together（openai-completions）、fireworks（openai-completions）、cerebras（openai-completions）、bedrock（bedrock-converse-stream）、nvidia（openai-completions）、huggingface（openai-completions）等
- **AC-09**：用户选择一个内置供应商后，显示该供应商的配置表单：API key 输入框（明文/可见切换）、可选 baseUrl 覆盖输入框
- **AC-10**：API key 输入框支持密文模式（默认）和明文切换
- **AC-11**：保存后，该内置供应商出现在当前方案的供应商列表中
- **AC-12**：内置供应商在 models.json 中生成的配置仅包含 `baseUrl`（若用户填写了）和 `apiKey`，不包含 `models` 数组和 `api` 字段（保留内置模型和 API 类型）。若用户仅填写了 API key 而未填写 baseUrl 且未添加自定义 model，则只生成 `apiKey` 字段
- **AC-12-1**：内置供应商表单中提供可选 baseUrl 覆盖输入框，用于将内置供应商指向第三方中转站（如将 openai 的 baseUrl 指向代理地址），与 pi 官方文档的 Overriding Built-in Providers 行为一致

### 自定义供应商

- **AC-13**：用户可以新建自定义供应商，填写：供应商名称（必填，作为 provider key）、baseUrl（必填）、API 类型（下拉选择，必填，选项为 openai-completions / anthropic-messages / google-generative-ai）、API key（选填）
- **AC-14**：自定义供应商创建后出现在当前方案的供应商列表中，与内置供应商区分显示（如标记"自定义"）
- **AC-15**：自定义供应商在 models.json 中生成的配置包含 `baseUrl`、`api`、`apiKey`（若有）和 `models` 数组

### Model 管理

- **AC-16**：在任意供应商条目下，用户可以查看该供应商已有的 model 列表
- **AC-17**：用户可以添加 model，必填字段为 `id`（model 标识符），选填字段为 `name`（默认同 id）、`reasoning`（默认 false）、`input`（默认 ["text"]，可选 ["text", "image"]）、`contextWindow`（默认 128000）、`maxTokens`（默认 16384）、`cost`（包含 input/output/cacheRead/cacheWrite 四个子字段，默认均为 0）
- **AC-18**：用户可以编辑已有 model 的任意字段
- **AC-19**：用户可以删除已有 model，删除前弹出确认
- **AC-20**：对于内置供应商，若用户未添加任何自定义 model 且未填写 baseUrl，则不在 models.json 中生成该供应商条目（因为没有需要覆盖的字段）
- **AC-21**：对于内置供应商，若用户添加了自定义 model 或填写了 baseUrl，则 models.json 中包含该供应商条目，自定义 model 与内置 model 合并（upsert by id）

### 数据持久化

- **AC-22**：所有方案及其配置数据持久化到工具自身的本地存储中（如 SQLite 或 JSON 文件），应用重启后数据不丢失
- **AC-23**：工具自身的配置数据与 pi 的 `models.json` 完全解耦，修改方案不影响 pi 当前配置，除非用户主动激活

### 边界与错误

- **AC-24**：方案名称为空时不允许保存，给出校验提示
- **AC-25**：自定义供应商的 baseUrl 为空时不允许保存，给出校验提示
- **AC-26**：自定义供应商的 API 类型未选择时不允许保存
- **AC-27**：model 的 id 为空时不允许保存，给出校验提示
- **AC-28**：同一方案下同一供应商的 model id 不可重复，重复时给出校验提示
- **AC-29**：同一方案下自定义供应商的名称不可与已有供应商（内置或自定义）重复，重复时给出校验提示
- **AC-30**：删除方案或 model 的确认对话框中，取消操作不执行删除
- **AC-31**：激活方案时，若 `~/.pi/agent/` 目录不存在，自动创建目录后写入
- **AC-32**：激活方案时，若写入失败（如权限不足），给出明确的错误提示

- **AC-33**：同一方案下同一内置供应商不可重复添加（一个方案中每个内置供应商最多出现一次）

## 待确认项

（已全部澄清，无待确认项）