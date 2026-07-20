# AGENTS.md —— pi-mgr 编码知识库导航

pi-mgr 是 Pi Coding Agent 的 Windows 桌面配置管理工具（Go + Wails v2 + Vue 3），管理 provider 和 model 配置方案，一键激活写入 pi 的 `models.json`。

> **导读**：通读「全局硬规则」掌握项目边界 → 查阅「改动信号→知识域路由」表按工作内容定位 spec → 结合「Codebase Memory 使用规则」高效分析代码。

## 全局硬规则

- **平台**：仅 Windows；`~/.pi/` 解析为 `%USERPROFILE%\.pi\`，应用数据存 `%APPDATA%\pi-mgr\`
- **网络调用**：`FetchProviderModels` 和 `TestProviderConnectivity` 发起 HTTP GET 请求（仅此两处）；不校验 API key 有效性；SSH 同步使用本地 ssh/scp 命令
- **不管理 pi 其他配置**：仅 `models.json`；不碰 `auth.json`、`settings.json`、extensions
- **不导入现有 models.json**：v1 从零创建方案，不读取已有 pi 配置
- **多实例**：允许，最后激活者生效；不做冲突检测

## Codebase Memory 使用规则

- 分析架构、流程、调用链、数据流、影响面或跨文件改动时，先用 `codebase-memory` 收敛范围，再读取当前工作区源码；已知精确位置的单文件小改可直接读文件。
- 开始前用 `list_projects`/`index_status` 确认索引为 `ready`，并比较索引 branch/HEAD 与当前 Git 状态；索引过期或存在未提交修改时，图谱只用于发现候选关系。
- 按目标选择工具：`get_architecture` 看局部架构与入口，`search_graph` 找符号和实现，`trace_path` 查调用者/下游/参数传播，`search_code` 查宏、事件、注册表和文本，复杂聚合才用 `query_graph`。
- 查询应限定目录、标签和深度；检查 `has_more`、`total`、`limit` 等截断信号，必要时缩小条件或翻页，不能把首屏结果当全集。翻页模式：`offset += limit` 循环直到 `has_more == false`。
- 修改前用 inbound `trace_path` 检查调用者和影响面；修改后用当前源码搜索旧引用、漏注册和相关调用点，不要求为未提交改动重建索引。例外：新增大量文件或重构包结构后，建议手动触发一次 `index_repository(mode: "fast")` 刷新索引。
- 图谱是导航证据，不是最终事实。条件编译、宏、函数指针、回调、异步队列、timer/ISR、资源所有权、构建配置、忽略目录、未提交代码、接口多态运行时绑定、依赖注入容器、泛型实例化，必须通过当前源码、测试与项目文档核验。
- 工具不可用、项目未索引或查询失败时，退回 `rg`/直接读文件并说明限制；除非用户明确要求，不自动触发全仓索引。


## 改动信号 → 知识域路由

| 改动信号 | 知识域 / 叶子 spec |
|---|---|
| 涉及 `schemes.json` 读写、方案 CRUD、持久化 | `spec/contracts/storage.md` |
| 涉及方案复制（DuplicateScheme）、导入/导出（ExportSchemes, ImportSchemes） | `spec/architecture/overview.md` (§方案导入/导出) |
| 涉及 `models.json` 序列化、激活写入、字段映射 | `spec/contracts/storage.md` (§models.json 序列化规则) |
| 涉及校验规则、错误消息、边界条件 | `spec/contracts/validation.md` |
| 涉及 Wails API 绑定、Go↔前端接口 | `spec/architecture/overview.md` (§API 绑定) |
| 涉及前端页面、路由、组件结构、状态管理、Toast、CSS | `spec/architecture/frontend.md` |
| 涉及前端表单交互、拖拽排序、模型搜索/过滤、API 调用模式 | `spec/architecture/frontend.md` |
| 涉及前端模型预设（MODEL_PRESETS）选择交互 | `spec/architecture/frontend.md` (§模型预设) |
| 涉及内置供应商列表、API 类型枚举 | `spec/contracts/storage.md` (§内置供应商目录) |
| 涉及供应商 CRUD（AddCustomProvider, UpdateProvider, RemoveProvider） | `spec/architecture/overview.md` (§供应商管理) |
| 涉及 HTTP 模型拉取（FetchProviderModels）、批量导入（ImportProviderModels） | `spec/architecture/overview.md` (§模型拉取与导入) |
| 涉及 Provider/Model 排序（ReorderProviders, ReorderModels） | `spec/architecture/overview.md` (§排序) |
| 涉及批量删除模型（RemoveModels） | `spec/architecture/overview.md` (§批量操作) |
| 涉及供应商连通性测试（TestProviderConnectivity） | `spec/architecture/overview.md` (§连通性测试) |
| 涉及后端模型预设数据定义（MODEL_PRESETS） | `spec/architecture/overview.md` (§模型预设) |
| 涉及 SSH 连接测试、远程配置同步、SSH 地址持久化 | `spec/architecture/overview.md` (§SSH 同步) |
| 涉及 settings.json（SSH 地址等应用级设置） | `spec/contracts/storage.md` (§应用设置) |
| 涉及应用启动/初始化（startup） | `spec/architecture/overview.md` (§分层) |
| 涉及项目构建、命名规范、平台条件、发布 Release | `spec/conventions/project.md` |
| 涉及 Pi 管理（pi_manage.go）、内置提示词/安装/删除/预览、Pi 版本/插件管理 | `spec/conventions/project.md`（§发布 + §项目结构） |

## 快速链接

- 架构概览、模块边界、API 绑定、SSH 同步、排序、批量操作、连通性测试、模型预设：`spec/architecture/overview.md`
- 前端组件树、状态管理、数据流、交互模式、样式系统：`spec/architecture/frontend.md`
- 存储格式、序列化契约、内置供应商目录、应用设置：`spec/contracts/storage.md`
- 验证规则与错误矩阵：`spec/contracts/validation.md`
- 项目约定、构建、发布 Release：`spec/conventions/project.md`
