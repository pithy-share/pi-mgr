# PRD：SSH 配置同步（Windows → Ubuntu）

## 问题

- 用户同时在 Windows 和 Ubuntu 上使用 Pi Coding Agent，但两套配置（settings.json、models.json、prompts）相互独立，需手动分别维护
- Windows 端已通过 pi-mgr 完成 providers/models 配置、settings 调节和 prompts 定制，Ubuntu 端缺少这些配置，每次更新都要人工逐项复制
- 缺乏一键同步手段，配置不一致容易导致双端体验差异和工作重复

## 目标与非目标

### 目标

- 在 pi-mgr 中新增「SSH 同步」功能，将 Windows 本地的 Pi 配置文件一键同步到远程 Ubuntu 机器
- 同步的文件包括：
  - `~/.pi/settings.json`
  - `~/.pi/agent/models.json`
  - `~/.pi/prompts/` 目录及其下所有文件
  - `~/.pi/agent/skills/` 目录及其下所有 skill 子目录和文件
- 使用 SSH 作为传输通道，复用本地已配好的 SSH 环境（密钥、known_hosts）
- 提供 SSH 连接地址配置项（`user@host[:port]` 格式），默认值为 `zyong@192.168.1.180`，用户可修改并持久保存
- 同步完成后界面反馈成功/失败

### 非目标

- 不支持 Ubuntu → Windows 的反向同步（仅单向 Windows → Ubuntu）
- 不支持除 SSH 之外的其他传输方式（SCP、rsync 基于 SSH 的可行，但不支持 FTP、HTTP 推送等）
- 不管理或同步 auth.json、trust.json、sessions、extensions、npm、bin、agents 等 pi 目录下的其他文件
- 不做配置内容差异对比或合并——始终以 Windows 配置覆盖 Ubuntu 对应文件
- 不做同步前备份（Ubuntu 端原有文件被覆盖前不自动备份）
- 不校验 settings.json 或 prompts 内容的合法性（仅 models.json 涉及 pi-mgr 已有校验逻辑）
- 不处理 SSH 密钥首次部署或 SSH 密码输入——依赖本地已就绪的 SSH 环境

## 范围内外

### 范围之内

- pi-mgr 中新增一个「SSH 同步」页面或对话框入口（可从主界面导航访问）
- SSH 连接地址配置区域：输入框支持用户填写 `user@host[:port]` 格式，包含默认值 `zyong@192.168.1.180`；配置值持久化保存（存于 `%APPDATA%\pi-mgr\settings.json` 或其他合适位置），下次打开自动恢复
- 「测试连接」按钮：执行 `ssh user@host -p port exit` 验证 SSH 可达性，反馈成功/失败信息
- 「开始同步」按钮：一键同步三个配置文件/目录到远程机器
- 同步协议选择：使用系统 `scp` 命令（文件）或 `rsync`（目录）实现传输，依赖系统已安装 OpenSSH

### 同步映射规则

| Windows 源路径 | Ubuntu 目标路径 | 传输方式 |
|---|---|---|
| `%USERPROFILE%\.pi\settings.json` | `~/.pi/settings.json` | scp 覆盖 |
| `%USERPROFILE%\.pi\agent\models.json` | `~/.pi/agent/models.json` | scp 覆盖 |
| `%USERPROFILE%\.pi\prompts\` | `~/.pi/prompts/` | rsync -r 同步（含新增/更新/删除） |
| `%USERPROFILE%\.pi\agent\skills\` | `~/.pi/agent/skills/` | rsync -r 同步（含新增/更新/删除） |

### 错误处理

- SSH 连接失败（主机不可达、认证失败、超时等）：界面展示明确错误消息，不继续传输
- 单个文件传输失败：界面展示失败文件路径和错误原因，其余文件继续传输
- 目标路径不存在（如 `~/.pi/agent/` 在 Ubuntu 上初次使用可能不存在）：自动创建目录后传输

### 范围之外

- 不修改 pi-mgr 现有的方案 CRUD、激活、模型拉取等核心功能
- 不修改现有的 schemes.json 存储格式、序列化规则或验证规则
- 不修改前端路由或现有页面的布局
- 不做跨平台的配置文件内容转换（如 Windows CRLF 到 Linux LF——由 SSH/scp 自动处理或保持原样）

### 网络调用豁免说明

AGENTS.md 中「无网络调用」的硬规则在此功能范围内豁免——SSH 同步是用户主动触发的推送操作，不是后端服务的校验或验证。

## 验收标准

### 正常场景

- **AC-01**：pi-mgr 主界面提供「SSH 同步」入口，点击进入 SSH 同步页面，页面包含 SSH 地址输入框（默认填充 `zyong@192.168.1.180`）、「测试连接」按钮和「开始同步」按钮
- **AC-02**：用户修改 SSH 地址输入框内容后，地址值被持久保存；关闭再打开 pi-mgr 后，输入框自动恢复为用户上次保存的值
- **AC-03**：点击「测试连接」，SSH 连接成功时界面显示绿色成功提示；失败时显示红色错误信息（含具体原因：超时/认证失败/主机不可达等）
- **AC-04**：点击「开始同步」，四个配置文件/目录按映射规则传输到 Ubuntu；同步成功后界面展示成功摘要（列出已同步的文件及大小/状态）
- **AC-05**：Ubuntu 上 `~/.pi/agent/` 目录不存在时，同步自动创建该目录并写入 models.json
- **AC-06**：Ubuntu 上 `~/.pi/prompts/` 目录不存在时，同步自动创建该目录并写入所有 prompt 文件
- **AC-07**：Ubuntu 上 `~/.pi/prompts/` 已存在部分 prompts 文件时，同步后 Windows 端所有 prompt 文件出现在 Ubuntu 上（新增的添加、已有同名文件被覆盖、Windows 已删除的同步后在 Ubuntu 上也被删除——即 rsync 镜像语义）
- **AC-08**：Ubuntu 上 `~/.pi/agent/skills/` 已存在部分 skills 时，同步后 Windows 端所有 skill 目录和文件出现在 Ubuntu 上（新增的添加、已有同名文件/目录被覆盖、Windows 已删除的同步后 Ubuntu 上也删除——即 rsync 镜像语义）

### 错误场景

- **AC-09**：SSH 连接失败（主机不可达、DNS 解析失败）时，界面展示错误信息，「开始同步」按钮点击后不发起文件传输，立即返回连接失败提示
- **AC-10**：SSH 连接失败（密钥认证失败、密码错误）时，同 AC-09 的错误提示逻辑
- **AC-11**：单个文件/目录传输失败（如 `settings.json` 源文件被占用无法读取）时，界面标记该项同步失败及其原因，其余项继续传输不受影响
- **AC-12**：所有项目均传输失败时，界面展示整体失败摘要，不显示部分成功

### 边界场景

- **AC-13**：本地 `~/.pi/settings.json` 不存在时，同步操作跳过该文件，同步摘要中注明"settings.json：本地文件不存在，已跳过"
- **AC-14**：本地 `~/.pi/agent/models.json` 不存在时（用户未通过 pi-mgr 激活过任何方案），同样跳过并注明
- **AC-15**：本地 `~/.pi/prompts/` 目录不存在或为空时，远程创建空目录（或跳过并注明）
- **AC-16**：本地 `~/.pi/agent/skills/` 目录不存在或为空时，远程创建空目录（或跳过并注明）
- **AC-17**：SSH 地址输入为空时，「测试连接」和「开始同步」按钮置灰不可点击，界面提示"请填写 SSH 地址"
- **AC-18**：SSH 地址格式错误（如不含 `@`、多余空格）时，「测试连接」和「开始同步」按钮置灰，界面提示"SSH 地址格式应为 user@host[:port]"

## 待确认项

| # | 事项 | 默认推荐 | 确认状态 |
|---|---|---|---|
| Q1 | 同步时是否需要在 Ubuntu 端自动备份被覆盖的文件？ | v1 不做备份，直接覆盖；后续可增加 | ✅ 已确认 |
| Q2 | SSH 端口：默认使用 22，是否需要额外配置端口输入？ | 地址格式支持 `user@host:port` 扩展，默认不显示端口栏 | ✅ 已确认 |
| Q3 | 同步范围：`~/.pi/agent/agents/`（自定义 agent 定义）是否也一并同步？ | 暂不同步，只同步用户提到的三个配置项；后续可扩展 | ✅ 已确认 |
| Q4 | 传输方式：使用 Go 的 `os/exec` 调用系统 scp/rsync 命令，还是用 Go SSH 库（如 `golang.org/x/crypto/ssh` + SFTP）纯 Go 实现？ | 推荐用 `os/exec` 调用系统命令，减少依赖，且用户 SSH 环境已就绪 | ✅ 已确认 |
| Q5 | UI 位置：在现有方案列表页加一个「SSH 同步」按钮跳转，还是单独的设置页面入口？ | 推荐主界面加导航按钮，跳转到独立同步页 | ✅ 已确认 |
| Q6 | 同步完成后是否需要提示重启 Ubuntu 端的 pi？ | 推荐同步成功摘要中添加提示"请在 Ubuntu 上重启 Pi 以使配置生效" | ✅ 已确认 |
