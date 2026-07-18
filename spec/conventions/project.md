# 项目约定

**阅读时机**：涉及项目初始化、构建配置、命名规范、平台条件时。  
**可核验依据**：`task/pi-provider-model-manager/plan.md` Decision 1–3, Step 1.1。

## 技术栈

| 层 | 技术 | 版本约束 |
|---|---|---|
| 桌面框架 | Wails | v2 |
| 后端语言 | Go | Wails v2 支持版本 |
| 前端框架 | Vue 3 + TypeScript | Wails 默认模板 |
| 包管理 | npm / pnpm（前端） + Go modules（后端） | — |
| 目标平台 | Windows | 仅 Windows，不交叉编译 |

## 项目结构

```
pi-mgr/                       # Wails 项目根（即当前仓库根）
├── main.go                   # Wails 入口
├── app.go                    # App struct, startup
├── models.go                 # 数据模型（Scheme, Provider, Model, BuiltInProvider）
├── store.go                  # schemes.json / active.json 持久化 + UUID 生成
├── serializer.go             # models.json 序列化
├── activate.go               # 激活写入 models.json
├── validate.go               # 校验规则
├── builtin.go                # 内置供应商目录 + 有效 API 类型
├── api.go                    # Wails API 方法（CRUD + 导入导出 + 目录查询 + 辅助函数）
├── fetch.go                  # HTTP 模型列表拉取（FetchProviderModels）
├── ssh_sync.go               # SSH 连接测试 + 配置同步（SyncPiConfig）
├── ssh_settings.go           # SSH 地址持久化（settings.json）
├── wails.json                # Wails 项目配置
├── frontend/                 # Vue 3 前端
│   ├── src/
│   │   ├── views/            # 页面组件（SchemeList, SchemeEditor）
│   │   ├── components/       # 可复用组件
│   │   └── wails/            # Wails 生成的运行时绑定 + dev mode fallback
│   └── package.json
└── build/                    # Wails 构建产物
```

## 命名规范

- Go 导出名：PascalCase（`ListSchemes`, `SerializeToModelsJSON`）
- Go 文件名：snake_case（`models.go`, `store.go`）
- Go JSON tag：camelCase（`json:"apiKey"`）
- Vue 组件：PascalCase 文件，kebab-case 在模板中
- 路由路径：kebab-case（`/scheme/:id`）
- 方案 ID：UUID v4

## 构建

```bash
wails dev       # 开发模式，热重载
wails build     # 生产构建，输出 Windows exe
```

## 文件路径（Windows）

| 用途 | 路径 |
|---|---|
| 工具持久化数据（方案） | `%APPDATA%\pi-mgr\schemes.json` |
| 工具持久化数据（应用设置） | `%APPDATA%\pi-mgr\settings.json` |
| 活跃方案追踪 | `%APPDATA%\pi-mgr\active.json` |
| pi models.json | `%USERPROFILE%\.pi\agent\models.json` |
| pi agent 目录 | `%USERPROFILE%\.pi\agent\` |

**路径解析**：Go 端使用 `os.UserConfigDir()` + `os.UserHomeDir()` 获取对应目录，不硬编码盘符。

## 设计决策

### 决策 1：存储格式

- **选型**：单个 JSON 文件（`schemes.json`）+ 原子写入（temp + rename）
- **理由**：数据量小（数十个方案、数百个 model），JSON 可调试，无 CGo 依赖
- **后果**：极大数据量时有性能风险，但 v1 场景不会触发

### 决策 2：前端框架

- **选型**：Vue 3 + TypeScript（Wails 默认模板）
- **理由**：Wails v2 最佳支持，响应式表单开发效率高

### 决策 3：不导入现有 models.json

- **决定**：v1 不从现有 pi models.json 导入配置
- **理由**：PRD 明确范围外
