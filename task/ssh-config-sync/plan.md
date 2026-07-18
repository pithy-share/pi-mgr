# 实施计划：SSH 配置同步

## 证据链

```
PRD 要求（SSH 同步页、地址持久化、scp/rsync 传输）
  → 已确认产品不变量（Windows 平台、os.UserConfigDir/HomeDir、原子写入、Wails API 自动绑定）
  → 真实调用入口 (api.go App 方法、main.ts 路由、api.ts 桥接层、Vue views/)
  → 当前行为（无 SSH 功能、无 settings.json 持久化、无 os/exec 调用）
  → 必要改动（新建 ssh_sync.go、ssh_settings.go、SshSync.vue；修改 api.ts、main.ts、SchemeList.vue、types.ts）
```

**产品不变量**（来自 AGENTS.md + spec/）：
- 仅 Windows；路径通过 `os.UserHomeDir()` / `os.UserConfigDir()` 解析
- 所有持久化使用原子写入（temp + rename，`store.go` 模式）
- Wails API 通过 `App` 结构体上的导出方法自动绑定（`main.go` Bind 列表）
- 前端通过 `window['go'].main.App` 桥接；手动 `api.ts` 维护接口 + fallback
- Vue 3 + TypeScript，hash 路由（`main.ts` 定义），组件 PascalCase 文件名
- 不修改：`schemes.json`、`models.json` 序列化、激活流程、校验规则、现有路由布局

**支持的输入形态**：
- SSH 地址 `user@host[:port]`，默认 `zyong@192.168.1.180`
- 4 个同步项：`settings.json`（scp）、`models.json`（scp）、`prompts/`（rsync）、`skills/`（rsync）
- 错误输入：空地址、格式错误（无 @、多余空格）

**禁止假设**：
- 不假设 rsync 已安装在系统 PATH 中（运行时检测）
- 不假设远程目录已存在（通过 SSH 预创建）
- 不假设本地文件/目录存在（逐项检查并跳过）
- 不处理 SSH 密钥部署或密码输入（依赖已就绪的 SSH 环境）
- 不做文件内容差异对比、备份、反向同步

---

## 实施步骤

### Step 1：Go 后端 — SSH 设置持久化

**目标**：实现 SSH 地址的持久化读写，复用 `store.go` 的原子写入模式。

**文件**：新建 `D:\src\pi-mgr\ssh_settings.go`

**内容**：
1. 定义 `appSettingsPath()` 返回 `%APPDATA%\pi-mgr\settings.json`
2. 定义 `appSettings` 结构体：`{ SSHAddress string }`
3. 实现 `saveAppSettings(settings) error`：原子写入（temp + rename，同 `SaveSchemes` 模式）
4. 实现 `loadAppSettings() (*appSettings, error)`：文件不存在时返回零值（地址为空字符串），不报错
5. 在 `App` 上添加导出方法：
   - `SaveSSHAddress(address string) error`：加载→更新地址→保存
   - `LoadSSHAddress() (string, error)`：加载设置，返回 `SSHAddress` 字段

**关键联动**：
- 无：纯持久化操作，独立于 schemes.json
- 文件不存在时 `loadAppSettings` 返回 `{SSHAddress: ""}`，与 AC-02 首次使用场景一致

**保持不变的控制流**：
- `store.go` 不修改；settings.json 使用完全独立的文件路径和写入路径
- `schemes.json` 的读写逻辑不受影响

**验证方式**：
- 写入后读取返回相同值
- 文件不存在时返回默认空地址
- `SaveSSHAddress("")` 将空地址持久化（AC-17 清空场景）

---

### Step 2：Go 后端 — SSH 同步核心逻辑

**目标**：实现 `TestSSHConnection` 和 `SyncPiConfig` 两个 Wails 方法。

**文件**：新建 `D:\src\pi-mgr\ssh_sync.go`

**内容**：

#### 2.1 解析 SSH 地址的辅助函数 `parseSSHAddress(address string) (user, host, port string, err error)`

- 格式：`user@host`（默认 port 22）或 `user@host:port`
- 输入已验证（前端保证），但后端仍做安全检查
- 返回解析后的 user、host、port

#### 2.2 `App.TestSSHConnection(address string) (success bool, message string)`

- 调用 `parseSSHAddress` 解析地址
- 执行：`exec.Command("ssh", "-p", port, "-o", "ConnectTimeout=10", "-o", "BatchMode=yes", user+"@"+host, "exit")`
- 设置 15 秒超时（`context.WithTimeout`）
- `success=true` + 成功消息（如 "SSH 连接成功"），或 `success=false` + stderr 截取的消息
- 不暴露原始 SSH 错误细节（密钥路径等），仅返回用户可读的中文消息

#### 2.3 `App.SyncPiConfig(address string) SyncResult`

定义 `SyncResult` 结构体（Go 端 + TypeScript 端互相映射）：

```go
type SyncItemStatus struct {
    Name    string `json:"name"`    // "settings.json" / "models.json" / "prompts" / "skills"
    Status  string `json:"status"`  // "success" / "skipped" / "failed"
    Message string `json:"message"` // 详细说明
}

type SyncResult struct {
    Overall string          `json:"overall"` // "success" / "partial" / "failed"
    Items   []SyncItemStatus `json:"items"`
}
```

实现流程：

1. **前置检测**：
   - 解析地址；失败则返回整体失败
   - 检测 `rsync` 可用性：`exec.LookPath("rsync")`；失败则 AC-07/AC-08 不可达，返回整体失败（"rsync 未找到，请通过 Git for Windows 或 MSYS2 安装"）
   - 连接预检：执行 `ssh -p port -o ConnectTimeout=10 user@host exit`；失败则 AC-09/AC-10 语义，返回整体失败

2. **远程目录预创建**：
   - 执行 `ssh user@host "mkdir -p ~/.pi/agent ~/.pi/prompts ~/.pi/agent/skills"`（AC-05、AC-06）
   - 失败则整体失败（远程完全不可写）

3. **逐项同步**（独立 `exec.Command`，每项有独立错误捕获）：

   | 序号 | 项 | 命令 | AC 对应 |
   |---|---|---|---|
   | 1 | `settings.json` | `scp -P port %USERPROFILE%\.pi\settings.json user@host:~/.pi/settings.json` | AC-13 |
   | 2 | `models.json` | `scp -P port %USERPROFILE%\.pi\agent\models.json user@host:~/.pi/agent/models.json` | AC-14 |
   | 3 | `prompts/` | `rsync -r --delete -e "ssh -p port" "%USERPROFILE%\.pi\prompts/" user@host:~/.pi/prompts/` | AC-07, AC-15 |
   | 4 | `skills/` | `rsync -r --delete -e "ssh -p port" "%USERPROFILE%\.pi\agent\skills\" user@host:~/.pi/agent/skills/" | AC-08, AC-16 |

4. **本地存在性检查**（每项同步前）：
   - 文件项：`os.Stat(sourcePath)` → `os.IsNotExist` → 标记 `skipped`（"本地文件不存在，已跳过"）
   - 目录项：`os.Stat(sourcePath)` → `os.IsNotExist` → 标记 `skipped`（"本地目录不存在，已跳过"）；空目录通过 rsync 自然处理（传输空目录）
   - 不检查目录内容（rsync 会处理空目录情况）

5. **错误聚合**：
   - 每项独立执行，stderr 截取为 `message`
   - `Overall` 判定：全部 success→"success"；部分成功→"partial"；全部 failed→"failed"（AC-12）
   - 单个文件失败不阻断其他项（AC-11）

6. **成功后摘要**：同步摘要中如 Q6 约定，添加提示"请在 Ubuntu 上重启 Pi 以使配置生效"

**关键联动**：
- `parseSSHAddress` 供前端状态显示和后端连接使用
- `SyncResult` 类型需在 `frontend/src/types.ts` 镜像定义

**保持不变的控制流**：
- 不修改 `api.go` 中任何现有方法
- 不修改 `store.go`、`serializer.go`、`activate.go` 等
- Wails 自动绑定新增的 `App.TestSSHConnection` 和 `App.SyncPiConfig`

**验证方式**：
- 单元测试覆盖 `parseSSHAddress`（Go 可测）
- 手动验证：地址解析、连接检测、rsync 缺失检测

---

### Step 3：前端 — API 桥接扩展

**目标**：在前端 TypeScript 桥接层中添加 SSH 相关方法。

**文件**：`D:\src\pi-mgr\frontend\src\wails\api.ts`

**改动**：

1. `AppAPI` 接口新增方法：
   ```typescript
   TestSSHConnection(address: string): Promise<{ success: boolean; message: string }>
   SaveSSHAddress(address: string): Promise<void>
   LoadSSHAddress(): Promise<string>
   SyncPiConfig(address: string): Promise<SyncResult>
   ```

2. `api()` 函数的 fallback 分支添加对应桩实现：
   - `TestSSHConnection` → `Promise.resolve({ success: false, message: 'dev mode' })`
   - `SaveSSHAddress` / `LoadSSHAddress` / `SyncPiConfig` → 无操作桩

**保持不变的控制流**：
- 不修改已有方法的签名或实现
- fallback 仅在浏览器开发模式生效（无 backend 时）

---

### Step 4：前端 — 类型定义扩展

**目标**：添加 SSH 同步相关的 TypeScript 接口。

**文件**：`D:\src\pi-mgr\frontend\src\types.ts`

**改动**：新增：

```typescript
export interface SyncItemStatus {
  name: string
  status: 'success' | 'skipped' | 'failed'
  message: string
}

export interface SyncResult {
  overall: 'success' | 'partial' | 'failed'
  items: SyncItemStatus[]
}
```

**保持不变**：已有接口（`Scheme`, `Provider`, `Model`, `BuiltInProvider`, `Toast`, `defaultModel`）完全不动。

---

### Step 5：前端 — SSH 同步页面组件

**目标**：实现完整的 SSH 同步配置页面。

**文件**：新建 `D:\src\pi-mgr\frontend\src\views\SshSync.vue`

**页面结构**：

```
┌─────────────────────────────────┐
│  ← 返回方案列表                  │
│  SSH 配置同步                    │
│                                 │
│  SSH 地址                        │
│  [zyong@192.168.1.180      ]    │
│  (格式提示/错误提示)              │
│                                 │
│  [测试连接]  [开始同步]          │
│                                 │
│  ── 状态区域 ──                 │
│  (连接成功/失败提示)              │
│  (同步结果摘要)                  │
└─────────────────────────────────┘
```

**组件逻辑**：

1. **数据**：
   - `address: string` — 绑定输入框，初始化时通过 `LoadSSHAddress()` 加载
   - `connectionResult: { success: boolean, message: string } | null`
   - `syncResult: SyncResult | null`
   - `isTesting: boolean` / `isSyncing: boolean`

2. **计算属性**（AC-17、AC-18）：
   - `addressError: string` — 空→"请填写 SSH 地址"；格式错（正则 `/^[^\s@]+@[^\s@]+(:[0-9]+)?$/`）→"SSH 地址格式应为 user@host[:port]"；无错→空
   - `canOperate: boolean` — `!addressError && address.trim() !== ''`
   - 按钮 `disabled` 绑定 `!canOperate || isTesting || isSyncing`

3. **方法**：
   - `handleTestConnection()` — 调用 `TestSSHConnection`，设置 `connectionResult`
   - `handleSync()` — 调用 `SyncPiConfig`，设置 `syncResult`
   - 地址修改时：自动保存 `SaveSSHAddress`（防抖或 `watch` + 即时保存），并清除连接/同步结果

4. **UI 状态**：
   - 测试连接成功：绿色卡片 "SSH 连接成功"
   - 测试连接失败：红色卡片 + 具体原因
   - 同步中：按钮禁用 + 加载中状态
   - 同步完成：成功摘要卡片（每个项的状态 + 最终提示）

**关键联动**：
- 页面加载时调用 `LoadSSHAddress()` 恢复上次保存的地址（AC-02）
- 地址变化时自动调用 `SaveSSHAddress()` 持久化

**保持不变**：
- 不修改 `App.vue` 布局
- 不引入额外外部依赖

---

### Step 6：前端 — 路由注册

**目标**：将 SSH 同步页面加入路由。

**文件**：`D:\src\pi-mgr\frontend\src\main.ts`

**改动**：
1. 导入 `SshSync` 组件
2. 添加路由 `{ path: '/ssh-sync', component: SshSync }`

**验证**：导航到 `#/ssh-sync` 即可访问页面。

---

### Step 7：前端 — 导航入口

**目标**：在主界面上提供进入 SSH 同步页面的入口。

**文件**：`D:\src\pi-mgr\frontend\src\views\SchemeList.vue`

**改动**：
在标题栏的操作按钮组（"导出全部"、"导入"、"新建方案"）旁添加一个按钮：
```html
<button class="btn-secondary" @click="$router.push('/ssh-sync')">SSH 同步</button>
```

或者放在现有按钮之后。

**保持不变**：不修改页面布局、列表逻辑、操作处理函数。

---

## 验收与验证

### AC-01 ~ AC-18 覆盖

| AC | 验证方式 | 实施步骤覆盖 |
|---|---|---|
| **AC-01** | 页面包含输入框（默认值）、测试连接按钮、开始同步按钮 | Step 5（SshSync.vue 结构 + 初始加载地址） |
| **AC-02** | 修改地址 → 重启应用 → 地址恢复 | Step 1（SaveSSHAddress/LoadSSHAddress）+ Step 5（onMounted 加载 + watch 自动保存） |
| **AC-03** | 点击测试连接 → 成功绿/失败红 | Step 2.2（TestSSHConnection）+ Step 5（handleTestConnection + 条件样式） |
| **AC-04** | 点击开始同步 → 4 项传输 → 成功摘要 | Step 2.3（SyncPiConfig 执行）+ Step 5（handleSync + SyncResult 渲染） |
| **AC-05** | 远程 `~/.pi/agent/` 不存在 → 自动创建后写入 models.json | Step 2.3（远程目录预创建：`mkdir -p ~/.pi/agent`） |
| **AC-06** | 远程 `~/.pi/prompts/` 不存在 → 自动创建后写入 | Step 2.3（远程目录预创建：`mkdir -p ~/.pi/prompts`） |
| **AC-07** | prompts 已有文件 → rsync --delete 镜像 | Step 2.3（rsync -r --delete） |
| **AC-08** | skills 已有文件 → rsync --delete 镜像 | Step 2.3（rsync -r --delete） |
| **AC-09** | 主机不可达/DNS 解析失败 → 错误提示，不传输 | Step 2.3（连接预检）+ Step 2.2（TestSSHConnection 返回 false） |
| **AC-10** | 认证失败 → 同 AC-09 | Step 2.2（stderr 截取）+ Step 2.3（预检失败阻断） |
| **AC-11** | 单个文件失败 → 其余继续 | Step 2.3（逐项独立 exec + 错误收集） |
| **AC-12** | 全部失败 → 整体失败摘要 | Step 2.3（overall 判定逻辑） |
| **AC-13** | 本地 `settings.json` 不存在 → 跳过并注明 | Step 2.3（os.Stat + skipped 标记） |
| **AC-14** | 本地 `models.json` 不存在 → 跳过并注明 | Step 2.3（os.Stat + skipped 标记） |
| **AC-15** | 本地 `prompts/` 不存在或为空 → 创建空目录或注明 | Step 2.3（os.Stat + 空目录 rsync 自然处理） |
| **AC-16** | 本地 `skills/` 不存在或为空 → 同 AC-15 | Step 2.3（os.Stat + 空目录 rsync 自然处理） |
| **AC-17** | 地址为空 → 按钮置灰 + 提示 | Step 5（addressError 计算属性 + :disabled绑定） |
| **AC-18** | 格式错误 → 按钮置灰 + 格式提示 | Step 5（addressError 正则校验） |

### 静态核验方式

| 核验目标 | 方式 |
|---|---|
| 不修改 schemes.json 存储 | grep schemes.json / store.go / SaveSchemes 调用 — 无新增引用 |
| 不修改序列化/激活 | grep SerializeToModelsJSON / ActivateScheme — 无新增引用 |
| 不修改已有路由 | main.ts — 仅追加路由，不修改现有数组项 |
| 不修改已有 API 签名 | api.go — 仅新增方法，不修改已有方法签名 |
| 前端持久化调用 | SshSync.vue 中 SaveSSHAddress/LoadSSHAddress 调用链 → api.ts → Go 端 |
| rsync 未安装阻断 | ssh_sync.go exec.LookPath("rsync") 前置检测 |
| 远程目录预创建 | ssh_sync.go mkdir -p 命令 |

---

## 决策与风险

### 决策

| # | 事项 | 决策 | 理由 |
|---|---|---|---|
| D1 | 设置存储位置 | `%APPDATA%\pi-mgr\settings.json` | 与 schemes.json 同目录，模式一致 |
| D2 | 同步失败模型 | 4 项独立 exec，逐项收集结果 | AC-11 明确要求单项失败不影响其余 |
| D3 | 远程目录创建时机 | 在同步开始前一次性 `mkdir -p` | 避免每项创建增加 SSH 往返；单次创建覆盖所有目标 |
| D4 | 地址持久化时机 | 前端每次输入变化时自动保存（watch + 即时调用） | 无需额外"保存"按钮；PRD AC-02："修改后即被持久保存" |
| D5 | 连接预检方式 | `ssh -o BatchMode=yes -o ConnectTimeout=10 user@host exit` | BatchMode=yes 防止交互式密码输入等待；10s 超时 |

### 风险

| # | 风险 | 影响 | 缓解 |
|---|---|---|---|
| R1 | rsync 在 Windows 上未安装 | AC-07/AC-08 无法执行 | `exec.LookPath("rsync")` 运行时检测，返回明确中文错误 |
| R2 | 中文路径或空格 | scp/rsync 命令参数传递错误 | `exec.Command` 原生处理，非字符串拼接 |
| R3 | SSH 端口非 22 但用户输入格式错误 | 连接失败 | 前端正则校验 + 后端 `parseSSHAddress` 双重保障 |
| R4 | 远程磁盘空间不足 | rsync/scp 写入失败 | 由 stderr 捕获，归入单项失败（AC-11） |
| R5 | 用户 SSH 环境未就绪（无密钥、known_hosts 未确认） | 首次连接失败 | BatchMode=yes 快速失败；PRD 明确"依赖本地已就绪的 SSH 环境" |

### 不纳入计划的风险

- **并发连接**：PRD 说允许多实例但最后激活者生效。SSH 同步是多实例安全操作（只读本地文件 + 写入远程），不做冲突检测。
- **请求级事务**：不用；逐项独立，失败不回滚已成功项。PRD 不要求原子性。
- **全局 dirty 管理**：不用；SSH 同步不修改 schemes.json，不与现有功能共享状态。

---

## 未解决 Blocker

**无。** `Plan` 子 Agent 审查确认所有 18 个 AC 均可从已验证的真实入口点实现，无需：

- 修改核心 Go 数据模型（`models.go`）
- 修改 `schemes.json` 存储或序列化逻辑
- 修改现有路由或布局
- 引入新的 Go 外部依赖
- 进行未经 PRD 豁免的网络 HTTP 调用（SSH 已豁免）

已识别的 `rsync` 在 Windows 上的部署依赖在 **R1** 中记录并通过运行时 `exec.LookPath` 检测 + 错误提示缓解，不阻止 AC-07/AC-08 在用户安装 rsync 后正常执行。
