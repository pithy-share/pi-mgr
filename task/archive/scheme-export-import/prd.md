# 方案导入/导出 PRD

## 问题

当前 pi-mgr 的方案数据仅保存在单台电脑的 `%APPDATA%\pi-mgr\schemes.json` 中。用户若更换电脑或需要与同事共享配置，只能在另一台电脑上手动重新创建所有方案、供应商和模型，操作繁琐且容易出错。

## 目标与非目标

### 目标
- 支持将全部方案导出为 JSON 文件（内容等同于 `schemes.json`），用户可选择保存位置，复制到另一台电脑即可导入还原，无需重新手工配置。
- 导入时与原 `schemes.json` 自然合并，不破坏已有配置。
- 跨版本兼容：若 JSON 引用了当前版本不认识的内置 provider key，不拒绝导入，兼容降级为自定义供应商处理。

### 非目标
- 不实现基于网络、网盘、Git 或账号体系的同步/自动备份功能。
- 不校验导入 JSON 中 API Key 的有效性、不发起 HTTP 请求。
- 不提供导出为 Excel/CSV/YAML 等其他格式。
- 不实现导入前对 JSON 与当前硬件/系统环境的兼容性校验（如供应商在某台电脑是否可用）。

## 范围内外

### 范围内
- **导出全部方案**：方案列表页提供"导出全部"，直接复制 `%APPDATA%\pi-mgr\schemes.json` 的内容写入用户选定的 `.json` 文件。导出内容等同于当前 `schemes.json` 的完整 `[]Scheme` 数组。
- **导入 JSON**：方案列表页提供入口，用户通过系统文件选择器选择一个 `.json` 文件，解析其内容后合并写入 `schemes.json`。
- **导入冲突处理**：以 `Scheme.ID` 为匹配键，同 ID 覆盖全量（名称、供应商、模型全量替换），新 ID 追加，未覆盖的现有方案保持不变。
- **跨版本内置供应商兼容**：若导入 JSON 中的 `Provider` 标记 `builtIn: true` 但其 `key` 不在当前硬编码的 `BuiltInProviders` 列表中，自动降级为 `builtIn: false` 的自定义供应商，保留完整字段（key/name/apiKey/baseUrl/apiType/models 等）。
- **导出格式**：导出为 `.json` 文件，用户可选择保存路径。内容包含 `Scheme` 完整字段（id/name/providers），`Provider` 完整字段（key/name/builtIn/apiKey/baseUrl/apiType/models），`Model` 完整字段。不重复导出派生数据或运行时状态。
- **导入入口**：用户通过文件选择器选择 `.json` 文件进行导入。

### 范围外
- 不修改 `models.json` 序列化逻辑（`SerializeToModelsJSON`）和激活行为。
- 不修改内置供应商硬编码目录。
- 不修改现有方案 CRUD 校验规则。
- 不引入新的外部依赖或网络调用。
- 不实现导出/导入的密码加密或压缩功能。

## 验收标准

**AC-01** 方案列表页必须提供入口触发"导出全部"。操作后，程序弹出文件保存对话框，默认文件名为 `pi-mgr-schemes.json`（或含日期的 `pi-mgr-schemes-YYYYMMDD.json`），由用户选定保存路径。写入文件的内容为当前 `schemes.json` 的完整 `[]Scheme` JSON 数组，格式与原文件一致，包含所有方案及其嵌套字段。写入应为原子方式（先写临时文件再 rename）。

**AC-02** 已移除（不导出单个方案）。

**AC-04** 导入入口必须支持通过系统文件选择器选择 `.json` 文件并执行导入。程序读取选中文件后，必须能自动识别其顶层是 `Scheme` 对象还是 `[]Scheme` 数组，并统一按数组逻辑处理（兼容手动编辑或第三方生成的单方案 JSON）。

**AC-05** 导入时，程序必须以 `Scheme.ID` 为唯一匹配键：
- 若导入的 ID 已存在于 `schemes.json`，则完全覆盖该方案的 name、providers、models 等全部字段。
- 若导入的 ID 不存在，则追加为新方案。
- 未被导入覆盖的现有方案及其数据必须保持不变。

**AC-06** 导入含有 `Provider.builtIn: true` 的 JSON 时，若其 `key` 不在当前版本的 `BuiltInProviders` 硬编码列表中，程序不得拒绝该条 Provider 或整个导入请求。应自动将该 Provider 视为自定义供应商（`builtIn: false`），保留其 key、name、apiKey、baseUrl、apiType、models 等所有字段。

**AC-07** 导入写入 `schemes.json` 前，必须对新合并后的数据进行现有校验规则的检查（参见 `spec/contracts/validation.md`）：例如 Scheme 内 Provider.Key 唯一、同 Provider 内 Model.ID 唯一、必填字段非空等。若校验失败，应拒绝此次导入并返回错误，且不得产生任何部分写入或部分覆盖。

**AC-08** 导入写入 `schemes.json` 成功后，新导入/覆盖的方案必须立即在方案列表中可见，并可通过 `ActivateScheme` 正常激活并生成合法的 `models.json`。

**AC-09** 导入空数组（`[]`）时，应视为无操作，不修改现有 `schemes.json`，且返回成功。

**AC-10** 导入 JSON 为非对象/非预期格式、缺少 `id` 字段、或 JSON 语法非法时，必须返回清晰错误，且不得修改 `schemes.json`。

## 待确认项

无。
