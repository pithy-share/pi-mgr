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
├── api.go                    # Wails API 方法（CRUD + 排序 + 批量操作 + 连通性测试 + 导入导出 + 目录查询）
├── fetch.go                  # HTTP 模型列表拉取（FetchProviderModels）
├── ssh_sync.go               # SSH 连接测试 + 配置同步（SyncPiConfig）
├── ssh_settings.go           # SSH 地址持久化（settings.json）
├── pi_manage.go              # Pi CLI 管理（版本、插件、提示词）+ //go:embed 内置提示词
├── wails.json                # Wails 项目配置
├── pi/                       # pi 相关内置资源
│   ├── codebase-memory.md    # Codebase Memory 使用规则（//go:embed）
│   └── agent/prompts/        # 内置提示词模板（//go:embed, 4 个 .md）
├── frontend/                 # Vue 3 前端
│   ├── src/
│   │   ├── views/            # 页面组件（ConfigPage, PiManage, SshSync）
│   │   ├── types.ts          # TypeScript 类型定义（Config, Provider, Model, PromptTemplate 等）
│   │   ├── presets.ts        # 模型预设常量（MODEL_PRESETS）
│   │   └── wails/api.ts      # Wails API TypeScript 绑定 + dev mode fallback
│   └── package.json
└── build/                    # Wails 构建输出
    └── bin/
        └── pi-mgr.exe        # 生产构建产物（wails build 输出到此）
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
wails dev                     # 开发模式，热重载
wails build                   # 生产构建，输出到 build/bin/pi-mgr.exe
```

**注意**：生产构建产物位于 `build/bin/pi-mgr.exe`（约 12MB），**不是**项目根目录的 `pi-mgr.exe`（约 6.4MB，调试版本）。发布 Release 时必须使用 `build/bin/pi-mgr.exe`。

## 发布

```bash
# 1. 打标签
# 版本号示例：v0.1.0（feature）、v0.1.1（fix）
git tag v0.1.0
git push origin v0.1.0

# 2. 创建 Release 并上传 build/bin/pi-mgr.exe
git tag v0.1.0                                         # 如未打标签
git push origin v0.1.0
gh release create v0.1.0 \
  --title "v0.1.0 - 版本名称" \
  --notes "变更日志..." \
  build/bin/pi-mgr.exe

# 3. 更新已发布 Release（替换二进制）
gh release upload v0.1.0 build/bin/pi-mgr.exe --clobber
```

**常见错误**:
- ❌ 发布根目录的 `pi-mgr.exe`（调试版，Wails dev 产物）
- ✅ 必须发布 `build/bin/pi-mgr.exe`（生产版，`wails build` 产物）

## 文件路径（Windows）

| 用途 | 路径 |
|---|---|
| 工具持久化数据（方案） | `%APPDATA%\pi-mgr\schemes.json` |
| 工具持久化数据（应用设置） | `%APPDATA%\pi-mgr\settings.json` |
| 活跃方案追踪 | `%APPDATA%\pi-mgr\active.json` |
| pi models.json | `%USERPROFILE%\.pi\agent\models.json` |
| pi agent 目录 | `%USERPROFILE%\.pi\agent\` |
| pi 提示词目录 | `%USERPROFILE%\.pi\agent\prompts\` |
| pi-mgr 内置提示词 | `pi/agent/prompts/`（//go:embed 嵌入二进制） |

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
