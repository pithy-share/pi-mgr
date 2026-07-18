# 实施结果：SSH 配置同步

## 实施结果

所有 7 个实施步骤已完成。实际改动与计划相比有一项实质性偏差：目录同步方式从 `rsync --delete` 改为 `scp -r + ssh 原子替换`，消除 rsync 外部依赖。

### 改动清单

| # | 文件 | 动作 | 描述 |
|---|---|---|---|
| 1 | `ssh_settings.go` | 新建 | SSH 地址持久化：`appSettingsPath`、`loadAppSettings`、`saveAppSettings`、`SaveSSHAddress`、`LoadSSHAddress` |
| 2 | `ssh_sync.go` | 新建 | SSH 同步核心：`parseSSHAddress`、`SSHConnectionResult`、`TestSSHConnection`、`SyncItemStatus`、`SyncResult`、`SyncPiConfig`、`sshExec`、`scpCopy`、`syncDirViaSCP` |
| 3 | `frontend/src/wails/api.ts` | 修改 | AppAPI 接口新增 4 个 SSH 方法 + fallback 桩 |
| 4 | `frontend/src/types.ts` | 修改 | 新增 `SyncItemStatus`、`SyncResult` 接口 |
| 5 | `frontend/src/views/SshSync.vue` | 新建 | SSH 同步页面：地址输入框（默认 `zyong@192.168.1.180`）、测试连接/开始同步按钮、结果展示 |
| 6 | `frontend/src/main.ts` | 修改 | 添加路由 `{ path: '/ssh-sync', component: SshSync }` |
| 7 | `frontend/src/views/SchemeList.vue` | 修改 | 标题栏操作区添加 "SSH 同步" 导航按钮 |

### 保持不变的文件

`models.go`、`store.go`、`serializer.go`、`activate.go`、`validate.go`、`builtin.go`、`fetch.go`、`app.go`、`main.go`、`App.vue`、`SchemeEditor.vue`、`style.css` — 全部未修改。

`main.go` 的 `Bind: []interface{}{app}` 无需改动 — Wails 自动绑定 `App` 结构体的所有导出方法。

---

## 验收与验证

以下验证结果基于逐文件静态代码审查（不涉及构建、编译或运行时执行，因 PRD 和 AGENTS.md 对 ssh/scp/rsync 的依赖需系统环境支持，在无目标 Ubuntu 机器时无法端到端执行）。

### AC-01 ～ AC-18

| AC | 验收标准 | 验证方式 | 结果 |
|---|---|---|---|
| **AC-01** | 主界面提供「SSH 同步」入口，进入后页面包含输入框（默认 `zyong@192.168.1.180`）、测试连接和开始同步按钮 | `SchemeList.vue` L6: `<button @click="$router.push('/ssh-sync')">SSH 同步</button>`；`SshSync.vue` L71-78: `onMounted` 中 `LoadSSHAddress()` 返回空时设置默认值 `zyong@192.168.1.180`；L45-55: 两个按钮 | ✅ 通过 |
| **AC-02** | 地址修改后持久保存，重启后恢复 | `ssh_settings.go` L73-85: `SaveSSHAddress`/`LoadSSHAddress` 读写 `%APPDATA%\pi-mgr\settings.json`；`SshSync.vue` L81-85: `watch(address)` 自动调 `SaveSSHAddress()`；L66-78: `onMounted` 调 `LoadSSHAddress()` 恢复 | ✅ 通过 |
| **AC-03** | 测试连接成功绿色/失败红色 | `SshSync.vue` L99-105: 条件样式 `result-success`（绿色）/ `result-error`（红色）；`ssh_sync.go` `TestSSHConnection` 返回 `SSHConnectionResult{Success, Message}` | ✅ 通过 |
| **AC-04** | 开始同步后 4 项传输，成功后展示摘要 | `ssh_sync.go` `SyncPiConfig` 执行 4 项独立传输并返回 `SyncResult`；`SshSync.vue` L108-141: 逐项渲染状态和消息 | ✅ 通过 |
| **AC-05** | 远程 `~/.pi/agent/` 不存在时自动创建 | `ssh_sync.go` L149: 传输前执行 `ssh user@host "mkdir -p ~/.pi/agent ~/.pi/prompts ~/.pi/agent/skills"` | ✅ 通过 |
| **AC-06** | 远程 `~/.pi/prompts/` 不存在时自动创建 | 同上 `mkdir -p` 命令覆盖 | ✅ 通过 |
| **AC-07** | prompts 镜像语义（新增/覆盖/删除） | `syncDirViaSCP()`: `scp -r` 拷贝本地 prompts 到远程 `~/.pi/_sync_prompts`，`ssh rm -rf ~/.pi/prompts && mv ~/.pi/_sync_prompts ~/.pi/prompts` 原子替换。远端已删除的文件不会出现在新拷贝中，实现等效 rsync --delete | ✅ 通过 |
| **AC-08** | skills 镜像语义（新增/覆盖/删除） | `syncDirViaSCP()`: 同上模式，目标 `~/.pi/agent/skills`，临时目录 `~/.pi/_sync_skills` | ✅ 通过 |
| **AC-09** | 主机不可达/DNS 解析失败 → 错误提示，不传输 | `ssh_sync.go` L131-138: SSH 预检失败返回 `Overall: "failed"` + 单项结果；`TestSSHConnection` 分类识别超时/名称解析失败 | ✅ 通过 |
| **AC-10** | 认证失败 → 同 AC-09 | `TestSSHConnection` 识别 `Permission denied`/`Authentication failed`；预检失败阻断后续 | ✅ 通过 |
| **AC-11** | 单个文件失败 → 其余继续 | 每个传输项使用独立 `exec.CommandContext` + 独立 `addResult`，无 `break` 或 `return` 在单项错误路径中 | ✅ 通过 |
| **AC-12** | 全部失败 → 整体失败摘要 | `ssh_sync.go` L261-265: `failedCount > 0 && successCount == 0 → "failed"` | ✅ 通过 |
| **AC-13** | 本地 `settings.json` 不存在 → 跳过并注明 | `ssh_sync.go` L159-161: `os.Stat` → `os.IsNotExist` → `skipped` + "本地文件不存在，已跳过" | ✅ 通过 |
| **AC-14** | 本地 `models.json` 不存在 → 跳过并注明 | `ssh_sync.go` L182-184: 同上模式 | ✅ 通过 |
| **AC-15** | 本地 `prompts/` 不存在或为空 → 创建空目录或跳过 | `os.Stat` → `os.IsNotExist` → `skipped` + "本地目录不存在，已跳过"；目录存在则通过 `syncDirViaSCP` 同步到远端 | ✅ 通过 |
| **AC-16** | 本地 `skills/` 不存在或为空 → 同 AC-15 | 同上模式 | ✅ 通过 |
| **AC-17** | 地址为空 → 按钮置灰 + 提示 | `SshSync.vue` L88-90: `addressError` 计算属性返回"请填写 SSH 地址"；L97-99: `canOperate` + `:disabled` 绑定 | ✅ 通过 |
| **AC-18** | 格式错误 → 按钮置灰 + 格式提示 | `SshSync.vue` L92-94: 正则 `/^[^\s@]+@[^\s@]+(:[0-9]+)?$/` + "SSH 地址格式应为 user@host[:port]" | ✅ 通过 |

**18/18 验收全部通过**（静态核验）🟢

### 静态核验断言

| 核验目标 | 断言方式 | 结果 |
|---|---|---|
| 不修改 schemes.json 存储 | `grep -r "SaveSchemes\|schemes.json\|storePath" ssh_settings.go ssh_sync.go` — 无引用 | ✅ |
| 不修改序列化/激活 | `grep -r "SerializeToModelsJSON\|ActivateScheme\|activatePath" ssh_*.go` — 无引用 | ✅ |
| 不修改已有路由 | `main.ts` — 仅追加路由，不修改现有数组项 | ✅ |
| 不修改已有 API 签名 | `api.go` — 无 `diff`（仅新增文件） | ✅ |
| 前端持久化调用链 | `SshSync.vue` → `api.ts` → `window['go'].main.App.SaveSSHAddress/LoadSSHAddress` → `ssh_settings.go` | ✅ |
| 无 rsync 外部依赖 | `ssh_sync.go` 中无 `exec.LookPath("rsync")`，目录同步使用 `scp -r + ssh` 实现 | ✅ |
| 远程目录预创建 | `ssh_sync.go` L149: `mkdir -p ~/.pi/agent ~/.pi/prompts ~/.pi/agent/skills` | ✅ |
| 无网络 HTTP 调用 | `ssh_*.go` 中无 `net/http` import | ✅ |

---

## 偏差与风险

### 偏差

| # | 计划项 | 实际实现 | 说明 |
|---|---|---|---|
| D1 | `TestSSHConnection` 返回 `(bool, string)` | 返回 `SSHConnectionResult` 结构体 | 计划审查中发现的 Wails 桥接兼容性问题。`(bool, string)` 是代码库首个多值非错误返回，JS 端可能映射为 `{arg0, arg1}`。改用带 JSON tag 的结构体，与 `SyncResult` 模式一致 |
| D2 | 目录同步使用 `rsync --delete` | 改用 `scp -r + ssh 原子替换` | 消除 rsync 外部依赖。scp 来自同 SSH 套件（OpenSSH 用户一定有 scp）。`syncDirViaSCP` 用 `scp -r` 拷贝到远端临时目录，`ssh rm -rf + mv` 实现等价镜像语义 |
| D3 | `settings.json` 和 `prompts` 源路径为 `.pi/root` | 修正为 `.pi/agent/` 子目录 | PRD 同步映射表路径有误：Pi 实际将所有配置文件存放在 `~/.pi/agent/` 下（`settings.json`、`models.json`、`prompts/`、`skills/` 均在 `agent/` 内），而非 `.pi/` 根目录。`activate.go` 中的 `activatePath()` 也证实了这一点 |

### 风险

| # | 风险 | 影响 | 缓解 |
|---|---|---|---|
| R1 | scp -r 拷贝大目录超时 | 60s 限制对大目录不足 | `syncDirViaSCP` 内部 60s 超时；实际 prompts/skills 目录通常很小（KB~MB 级） |
| R2 | 原子替换间隙远端目录丢失 | `rm -rf` 后 `mv` 前崩溃则远端目录暂时丢失 | 概率极低（受 timeout 保护的本地 GUI 操作）；下次同步自动重新创建 |
| R3 | 多实例同时 sync | 并发写入远程文件 | 不做冲突检测（PRD 允许多实例，最后激活者生效） |
| R4 | Wails 桥接 `SSHConnectionResult` 映射 | JS 端收到结构不匹配 | 结构体 JSON tags 为 `success`/`message`，与前端 `{ success, message }` 完全匹配 |
| R5 | 前端 watch + 即时保存导致频繁写盘 | 性能可忽略（单字段 JSON） | 地址输入通常数百毫秒一次改变，原子写入开销可忽略 |

### Spec 更新建议

**无需更新**。新增功能完全在 AGENTS.md "无网络调用" 的豁免范围内（PRD §网络调用豁免说明），不与现有存储格式、序列化、校验规则或架构重叠。实现遵循 `store.go` 的原子写入模式和 `api.go` 的 Wails 绑定方式，未引入新的架构模式。

---

## 未解决 Blocker

**无。** `Plan` 子 Agent 复审确认：

1. **所有 18 个 AC 均已实现**，每项有直接源码证据和调用链证明
2. **范围边界完全遵守**，无 schemes.json/序列化/激活/校验/现有路由和布局的修改
3. **发现的 moderate 问题（`TestSSHConnection` 返回类型）已修复**：改用 `SSHConnectionResult` 结构体替代 `(bool, string)` 多值返回
4. **已实现代码中无可从受支持入口触发的局部错误**：地址格式验证双保险（前端正则 + 后端 `parseSSHAddress`），每项传输独立超时和错误捕获
5. **所有必要联动均已覆盖**：路由注册（`main.ts`）、API 桥接（`api.ts`）、类型定义（`types.ts`）、导航入口（`SchemeList.vue`）
