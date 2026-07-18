# AGENTS.md —— pi-mgr 编码知识库导航

pi-mgr 是 Pi Coding Agent 的 Windows 桌面配置管理工具（Go + Wails v2 + Vue 3），管理 provider 和 model 配置方案，一键激活写入 pi 的 `models.json`。

## 全局硬规则

- **平台**：仅 Windows；`~/.pi/` 解析为 `%USERPROFILE%\.pi\`，应用数据存 `%APPDATA%\pi-mgr\`
- **网络调用**：`FetchProviderModels` 发起 HTTP GET 请求拉取模型列表（仅此一处）；不校验 API key 有效性；SSH 同步使用本地 ssh/scp 命令
- **不管理 pi 其他配置**：仅 `models.json`；不碰 `auth.json`、`settings.json`、extensions
- **不导入现有 models.json**：v1 从零创建方案，不读取已有 pi 配置
- **多实例**：允许，最后激活者生效；不做冲突检测
- **代码探索走 cbm**：先定位后精读，不在 cbm 拿到位置前 `read` 代码文件
- **`read` 放行**：spec/*.md、wails.json、package.json 等非代码文件直接读

## cbm 三阶

```
get_architecture / search_graph    → 定向（拿 qualified_name）
         ↓
get_code_snippet / get_code_snippets → 精读（别搜完又去 read 整个文件）
         ↓
read                                → 回退（仅 spec、配置、或 cbm 反复失败时）
```

核心反模式就一个：**search_graph 返回了结果，别忽略它去 read 整个文件，对 qualified_name 调 get_code_snippet。**

其他：`read_symbol` 加 `file_path` 消歧，批量用 `get_code_snippets`，文字匹配用 `search_code`。

## 改动信号 → 知识域路由

| 改动信号 | 知识域 / 叶子 spec |
|---|---|
| 涉及 `schemes.json` 读写、方案 CRUD、持久化 | `spec/contracts/storage.md` |
| 涉及 `models.json` 序列化、激活写入、字段映射 | `spec/contracts/storage.md` (§序列化) |
| 涉及校验规则、错误消息、边界条件 | `spec/contracts/validation.md` |
| 涉及 Wails API 绑定、Go↔前端接口 | `spec/architecture/overview.md` (§API 绑定) |
| 涉及前端页面、路由、组件结构 | `spec/architecture/overview.md` (§前端) |
| 涉及内置供应商列表、API 类型枚举 | `spec/contracts/storage.md` (§内置供应商目录) |
| 涉及 HTTP 模型拉取（FetchProviderModels）、批量导入（ImportProviderModels） | `spec/architecture/overview.md` (§模型拉取与导入) |
| 涉及 SSH 连接测试、远程配置同步、SSH 地址持久化 | `spec/architecture/overview.md` (§SSH 同步) |
| 涉及 settings.json（SSH 地址等应用级设置） | `spec/contracts/storage.md` (§应用设置) |
| 涉及项目构建、命名规范、平台条件 | `spec/conventions/project.md` |

## 快速链接

- 架构概览、模块边界、API 绑定、SSH 同步：`spec/architecture/overview.md`
- 存储格式、序列化契约、内置供应商目录、应用设置：`spec/contracts/storage.md`
- 验证规则与错误矩阵：`spec/contracts/validation.md`
- 项目约定与构建：`spec/conventions/project.md`
