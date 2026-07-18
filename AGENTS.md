# AGENTS.md —— pi-mgr 编码知识库导航

pi-mgr 是 Pi Coding Agent 的 Windows 桌面配置管理工具（Go + Wails v2 + Vue 3），管理 provider 和 model 配置方案，一键激活写入 pi 的 `models.json`。

## 全局硬规则

- **平台**：仅 Windows；`~/.pi/` 解析为 `%USERPROFILE%\.pi\`，应用数据存 `%APPDATA%\pi-mgr\`
- **无网络调用**：不校验 API key 或 endpoint，不发起 HTTP 请求
- **不管理 pi 其他配置**：仅 `models.json`；不碰 `auth.json`、`settings.json`、extensions
- **不导入现有 models.json**：v1 从零创建方案，不读取已有 pi 配置
- **多实例**：允许，最后激活者生效；不做冲突检测

## 改动信号 → 知识域路由

| 改动信号 | 知识域 / 叶子 spec |
|---|---|
| 涉及 `schemes.json` 读写、方案 CRUD、持久化 | `spec/contracts/storage.md` |
| 涉及 `models.json` 序列化、激活写入、字段映射 | `spec/contracts/storage.md` (§序列化) |
| 涉及校验规则、错误消息、边界条件 | `spec/contracts/validation.md` |
| 涉及 Wails API 绑定、Go↔前端接口 | `spec/architecture/overview.md` (§API 绑定) |
| 涉及前端页面、路由、组件结构 | `spec/architecture/overview.md` (§前端) |
| 涉及内置供应商列表、API 类型枚举 | `spec/contracts/storage.md` (§内置供应商目录) |
| 涉及项目构建、命名规范、平台条件 | `spec/conventions/project.md` |

## 快速链接

- 架构概览与模块边界：`spec/architecture/overview.md`
- 存储格式与序列化契约：`spec/contracts/storage.md`
- 验证规则与错误矩阵：`spec/contracts/validation.md`
- 项目约定与构建：`spec/conventions/project.md`
