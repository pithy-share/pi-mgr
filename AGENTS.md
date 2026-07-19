# AGENTS.md —— pi-mgr 编码知识库导航

pi-mgr 是 Pi Coding Agent 的 Windows 桌面配置管理工具（Go + Wails v2 + Vue 3），管理 provider 和 model 配置方案，一键激活写入 pi 的 `models.json`。

## 全局硬规则

- **平台**：仅 Windows；`~/.pi/` 解析为 `%USERPROFILE%\.pi\`，应用数据存 `%APPDATA%\pi-mgr\`
- **网络调用**：`FetchProviderModels` 和 `TestProviderConnectivity` 发起 HTTP GET 请求（仅此两处）；不校验 API key 有效性；SSH 同步使用本地 ssh/scp 命令
- **不管理 pi 其他配置**：仅 `models.json`；不碰 `auth.json`、`settings.json`、extensions
- **不导入现有 models.json**：v1 从零创建方案，不读取已有 pi 配置
- **多实例**：允许，最后激活者生效；不做冲突检测

## cbm 使用规则

- **代码探索走 cbm**：先定位后精读，不在 cbm 拿到位置前 `read` 代码文件
- **`read` 放行**：spec/*.md、Makefile、.gitmodules 等非代码文件直接读

```
get_architecture / search_graph           → 定向（拿 qualified_name）
search_and_read_symbols                   → 搜索+源码一步到位（探索首选）
         ↓
get_code_snippet / get_code_snippets      → 精读（已知 qualified_name 时）
resolve_symbol / read_symbol              → 已知符号名消歧+读取
         ↓
trace_path                                → 追踪调用链/数据流
         ↓
search_code                               → 精确文本（宏、配置键、字面量）
         ↓
read                                      → 回退（仅 spec、配置、或 cbm 反复失败时）
```

核心反模式：**search_graph 返回了结果，别忽略它去 read 整个文件，对 qualified_name 调 get_code_snippet。**

选工具口诀：
- 知道符号名但不确定在哪 → `resolve_symbol` 消歧，确认后 `read_symbol` 读源码
- 概念搜索想直接看代码 → `search_and_read_symbols`，省一轮往返
- 已知精确 qualified_name → `get_code_snippet` 或批量 `get_code_snippets`
- 追踪谁调了谁、数据怎么流 → `trace_path`，不用手动 grep 调用链
- 搜宏、配置键、日志字符串、字面量 → `search_code`，不是 `search_graph`
- 批量消歧读多个符号 → `read_symbols`，加 `file_path` 或 `parent_class` 消歧

## 改动信号 → 知识域路由

| 改动信号 | 知识域 / 叶子 spec |
|---|---|
| 涉及 `schemes.json` 读写、方案 CRUD、持久化 | `spec/contracts/storage.md` |
| 涉及 `models.json` 序列化、激活写入、字段映射 | `spec/contracts/storage.md` (§序列化) |
| 涉及校验规则、错误消息、边界条件 | `spec/contracts/validation.md` |
| 涉及 Wails API 绑定、Go↔前端接口 | `spec/architecture/overview.md` (§API 绑定) |
| 涉及前端页面、路由、组件结构、状态管理、Toast、CSS | `spec/architecture/frontend.md` |
| 涉及前端表单交互、拖拽排序、模型搜索/过滤、API 调用模式 | `spec/architecture/frontend.md` |
| 涉及模型预设（MODEL_PRESETS）选择交互 | `spec/architecture/frontend.md` (§模型预设) |
| 涉及内置供应商列表、API 类型枚举 | `spec/contracts/storage.md` (§内置供应商目录) |
| 涉及 HTTP 模型拉取（FetchProviderModels）、批量导入（ImportProviderModels） | `spec/architecture/overview.md` (§模型拉取与导入) |
| 涉及 Provider/Model 排序（ReorderProviders, ReorderModels） | `spec/architecture/overview.md` (§排序) |
| 涉及批量删除模型（RemoveModels） | `spec/architecture/overview.md` (§批量操作) |
| 涉及供应商连通性测试（TestProviderConnectivity） | `spec/architecture/overview.md` (§连通性测试) |
| 涉及模型预设数据定义（MODEL_PRESETS） | `spec/architecture/overview.md` (§模型预设) |
| 涉及 SSH 连接测试、远程配置同步、SSH 地址持久化 | `spec/architecture/overview.md` (§SSH 同步) |
| 涉及 settings.json（SSH 地址等应用级设置） | `spec/contracts/storage.md` (§应用设置) |
| 涉及项目构建、命名规范、平台条件 | `spec/conventions/project.md` |

## 快速链接

- 架构概览、模块边界、API 绑定、SSH 同步、排序、批量操作、连通性测试、模型预设：`spec/architecture/overview.md`
- 前端组件树、状态管理、数据流、交互模式、样式系统：`spec/architecture/frontend.md`
- 存储格式、序列化契约、内置供应商目录、应用设置：`spec/contracts/storage.md`
- 验证规则与错误矩阵：`spec/contracts/validation.md`
- 项目约定与构建：`spec/conventions/project.md`
