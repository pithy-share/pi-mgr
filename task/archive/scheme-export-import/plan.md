# 方案导入/导出 — 实施计划

## 实施步骤

### Step 1: Go — 新增 `IsBuiltInProvider` 辅助函数

- **文件**：`builtin.go`
- **目标**：提供可复用的内置供应商 key 查询，供导入逻辑判断是否需降级（AC-06）。
- **内容**：
  ```go
  // IsBuiltInProvider checks if a key is in the built-in provider catalog
  func IsBuiltInProvider(key string) bool {
      for _, bp := range BuiltInProviders {
          if bp.Key == key {
              return true
          }
      }
      return false
  }
  ```
- **保持不变**：`BuiltInProviders` 硬编码目录本身不做任何修改（属范围内"不修改内置供应商硬编码目录"）。
- **验证方式**：静态确认函数正确遍历 `BuiltInProviders` 并返回匹配结果。

---

### Step 2: Go — 新增 `ExportSchemes` API 方法

- **文件**：`api.go`
- **目标**：实现 AC-01 导出全部方案。
- **签名**：`func (a *App) ExportSchemes() error`
- **实现逻辑**：
  1. 调用 `runtime.SaveFileDialog(a.ctx, &runtime.SaveDialogOptions{
         DefaultFilename: "pi-mgr-schemes.json",
         Title: "导出配置方案",
         Filters: []runtime.FileFilter{{DisplayName: "JSON 文件 (*.json)", Pattern: "*.json"}},
     })` 获取用户选择的保存路径。
     - 若用户取消（返回空字符串），直接返回 `nil`，不报错。
  2. 调用 `LoadSchemes()` 获取全部方案。
  3. 使用 `json.MarshalIndent(schemes, "", "  ")` 序列化为 JSON。
  4. 原子写入用户选定路径：
     - 先写入临时文件（`path + ".tmp"`），
     - 再 `os.Rename` 重命名。
     - rename 失败时回退到直接写入（同 `SaveSchemes` 的 fallback 模式）。
- **关键联动**：`SaveSchemes` 的原子写入模式被复用，但作用于用户选定的路径而非 `schemes.json`。
- **错误边界**：写入失败返回 error，由前端显示 toast。
- **对应验收**：AC-01。

---

### Step 3: Go — 新增 `ImportSchemes` API 方法

- **文件**：`api.go`
- **目标**：实现 AC-04、AC-05、AC-06、AC-07、AC-09、AC-10。
- **签名**：`func (a *App) ImportSchemes() error`
- **实现逻辑**：

  **3a. 文件对话框**：
  ```go
  path, err := runtime.OpenFileDialog(a.ctx, &runtime.OpenDialogOptions{
      Title: "导入配置方案",
      Filters: []runtime.FileFilter{{DisplayName: "JSON 文件 (*.json)", Pattern: "*.json"}},
  })
  ```
  - 用户取消 → 返回 `nil`。

  **3b. 读取与初步解析**：
  - 读取文件全部内容。
  - 用 `json.RawMessage` 先解析顶层，判断是 `[]Scheme` 还是 `Scheme`：
    - 若为对象（`{` 开头），包装为 `[]Scheme{scheme}`。
    - 若为数组，解析为 `[]Scheme`。
    - 若为空数组 `[]`，视为 AC-09 无操作，返回 `nil`（成功）。
  - JSON 语法非法 → 返回清晰错误（AC-10），不写入。

  **3c. 单条 Scheme 完整性检查**（AC-10 的一部分）：
  - 遍历每个导入的 scheme，若 `scheme.ID == ""`（空字符串或缺失），返回 `"导入的方案缺少 id 字段"` 错误。
  - 此检查在现有 `ValidateScheme` 之外单独做，因为 `ValidateScheme` 仅校验 `Name` 而不校验 `ID`（不修改现有校验规则）。

  **3d. 内置供应商降级**（AC-06）：
  - 遍历每个导入 scheme 的 providers：若 `Provider.BuiltIn == true` 且 `!IsBuiltInProvider(Provider.Key)`，则设置 `Provider.BuiltIn = false`。
  - 保留该 Provider 的所有字段（key、name、apiKey、baseUrl、apiType、models）。
  - ⚠️ 降级后的 Provider 可能因缺少 `baseUrl` 或 `apiType` 而在后续校验中失败。这是正确行为（数据不完整应被拒绝），不是 bug。

  **3e. 合并**（AC-05）：
  - 调用 `LoadSchemes()` 获取当前方案。
  - 遍历导入的 scheme 数组，以 `scheme.ID` 为键：
    - 若 ID 已存在 → 完全覆盖该位置的 name、providers 等全部字段。
    - 若 ID 不存在 → `append` 到末尾。
  - 未被覆盖的现有方案不做任何改动。

  **3f. 全量校验**（AC-07）：
  - 对合并后的 `[]Scheme` 逐个调用 `ValidateScheme`，若失败 → 返回错误，不写入。
  - 对每个 scheme 内的每个 provider 调用 `ValidateProvider`，通过时传入同 scheme 内全部 providers（含自身）以检测 key 唯一性。
  - 对每个 provider 内的每个 model 调用 `ValidateModel`，传入同 provider 内全部 models 以检测 ID 唯一性。
  - 若任意校验失败 → 返回首条错误，**绝不调用 `SaveSchemes`**，保证无部分写入。

  **3g. 持久化**：
  - 调用 `SaveSchemes(merged)` 原子写入。
  - 写入失败返回 error。

- **关键联动**：`LoadSchemes` / `SaveSchemes` 是唯一持久化入口，不做额外并发控制（产品不变量：单写者不锁）。
- **保持不变**：
  - 不修改 `ValidateScheme`、`ValidateProvider`、`ValidateModel` 函数体。
  - 不修改 `SerializeToModelsJSON` 或 `ActivateScheme`。
  - 不发起任何 HTTP 请求。
- **错误边界**：
  - 文件读取失败 → 返回 `"读取文件失败: ..."`。
  - JSON 解析失败 → 返回 `"JSON 格式错误: ..."`。
  - 缺少 id → 返回 `"导入的方案缺少 id 字段"`。
  - 校验失败 → 返回校验错误消息。
  - 写入失败 → 返回 `"保存方案失败: ..."`。
- **对应验收**：AC-04、AC-05、AC-06、AC-07、AC-09、AC-10。

---

### Step 4: 前端 — 更新 `api.ts` 接口定义

- **文件**：`frontend/src/wails/api.ts`
- **目标**：导出新 API 方法的类型签名和 fallback。
- **改动**：
  1. 在 `AppAPI` interface 中添加：
     ```typescript
     ExportSchemes(): Promise<void>
     ImportSchemes(): Promise<void>
     ```
  2. 在 fallback 对象中添加对应 stub：
     ```typescript
     ExportSchemes: () => Promise.resolve(),
     ImportSchemes: () => Promise.resolve(),
     ```
- **对应验收**：AC-01、AC-04（前端可用调用入口）。

---

### Step 5: 前端 — 方案列表页添加导出/导入按钮

- **文件**：`frontend/src/views/SchemeList.vue`
- **目标**：提供用户可触发的导出/导入入口（AC-01、AC-04）。
- **改动**：

  **5a. Header 区域按钮**：在现有"新建方案"按钮旁添加两个按钮：
  - `<button class="btn-secondary" @click="handleExport">导出全部</button>`
  - `<button class="btn-secondary" @click="handleImport">导入</button>`

  **5b. 导出 handler**：
  ```typescript
  async function handleExport() {
      try {
          const a = api()
          await a.ExportSchemes()
          showToast?.('方案已导出', 'success')
      } catch (e: any) {
          showToast?.(e?.message || e, 'error')
      }
  }
  ```

  **5c. 导入 handler**：
  ```typescript
  async function handleImport() {
      try {
          const a = api()
          await a.ImportSchemes()
          await loadSchemes()  // AC-08: 导入后立即刷新列表
          showToast?.('方案已导入', 'success')
      } catch (e: any) {
          showToast?.(e?.message || e, 'error')
      }
  }
  ```

- **关键联动**：导入成功后调用 `loadSchemes()` 刷新列表，确保 AC-08（新导入方案立即可见）得以满足。
- **保持不变**：不修改现有按钮布局和交互模式（编辑、复制、激活、删除、配置按钮行为不变）。
- **对应验收**：AC-01、AC-04、AC-08。

---

## 验收与验证

### 验证矩阵

| AC | 验证方式 | 验证步骤 |
|---|---|---|
| **AC-01** | 手动测试 + 代码审查 | 点击"导出全部"→ 对话框弹出，默认名 `pi-mgr-schemes.json` → 选定路径保存 → 文件内容为完整的 `[]Scheme` JSON（与 `schemes.json` 一致）→ 确认原子写入（临时文件不存在残留） |
| **AC-04** | 手动测试 | 点击"导入"→ 文件选择器弹出，过滤 `.json` → 选择文件 → 成功提示 |
| **AC-04 (对象格式兼容)** | 手动测试 | 准备包含单 `Scheme` 对象的 JSON 文件（非数组）→ 导入 → 方案出现在列表中 |
| **AC-05 (覆盖)** | 手动测试 | 导出现有方案 → 修改其中某方案的 name/providers → 导入修改后的文件 → 确认该方案被覆盖，其他方案不变 |
| **AC-05 (新增)** | 手动测试 | 准备含全新 ID 的 JSON → 导入 → 确认新方案追加，现有方案不变 |
| **AC-06** | 手动测试 | 准备 JSON 含 `builtIn: true` 但 key 不在 `BuiltInProviders` 中（例如 `"x-nsfw"`）→ 导入 → 确认该 provider 的 `builtIn` 被置为 `false`，其他字段保留 → 若缺少 baseUrl 则校验失败报错（合理） |
| **AC-07** | 手动测试 | 准备含重复 Provider.Key 的导入 JSON → 导入 → 被拒绝，报"供应商标识已存在"→ schemes.json 未改变 |
| **AC-07 (事务性)** | 代码审查 | 确认 `SaveSchemes` 在验证通过后才被调用；失败时绝不写入 |
| **AC-08** | 手动测试 | 导入含新方案的 JSON → 列表立即显示新方案 → 点击"激活"→ models.json 生成正确 |
| **AC-09** | 手动测试 | 导入空文件内容 `[]` → 成功提示 → 方案列表无变化 → schemes.json 未改变 |
| **AC-10 (语法错误)** | 手动测试 | 导入非 JSON 文件 → 报"JSON 格式错误"→ schemes.json 未改变 |
| **AC-10 (缺少 id)** | 手动测试 | 导入方案对象的 JSON（无 id 字段或 `"id": ""`）→ 报"导入的方案缺少 id 字段"→ schemes.json 未改变 |
| **AC-10 (非对象)** | 手动测试 | 导入 JSON 为字符串或数字 → 报格式错误 → schemes.json 未改变 |

### 回归验证

- 现有方案 CRUD（新建、编辑、复制、删除、激活）不受影响。
- 方案列表空状态、删除确认弹窗、Toast 提示等功能正常。
- `models.json` 序列化输出不受影响。

---

## 决策与风险

### 设计决策

| 决策 | 选项 | 选择理由 |
|---|---|---|
| 文件对话框在 Go 端还是前端 | Go 端使用 `runtime.SaveFileDialog` / `runtime.OpenFileDialog` | 符合"前端绝不直接读写磁盘"架构原则；Wails v2 的 runtime API 已在 go.mod 中，无需新依赖 |
| 导出内容格式 | 完整 `[]Scheme` JSON 数组 | 与 `schemes.json` 内部格式一致，可直接复制到另一台机器导入；非自定义格式降低学习成本 |
| 导入时单对象 vs 数组兼容 | 先用 `json.RawMessage` 判断顶层类型 | 兼容手动编辑或第三方生成的单方案 JSON，符合 PRD AC-04 |
| AC-10 缺少 id 的检查位置 | 在 `ImportSchemes` 内单独检查，不修改 `ValidateScheme` | PRD 范围外规定不修改现有校验规则；`ValidateScheme` 本就不校验 ID（由 `newUUID` 生成保证非空） |

### 风险

| 风险 | 影响 | 缓解措施 |
|---|---|---|
| `runtime.SaveFileDialog` 写入目标路径可能在跨卷（如 C:→D:）时 rename 失败 | 导出失败 | 复用 `SaveSchemes` 的 fallback：rename 失败时直接覆写目标文件 |
| 导入的 JSON 中 `Provider.BuiltIn` 降级后缺少 `baseUrl` 校验失败 | 导入被拒绝 | 这是正确行为（数据不完整应拒绝），向用户展示清晰错误信息；如需导入不完整数据，用户可先补齐字段 |
| 导入文件非常大（数千个方案） | UI 卡顿 | Go 端同步执行，Wails 调用会阻塞 UI 线程。当前不引入 goroutine（PRD 不要求异步，且单文件 JSON 大小通常 < 1MB） |

---

## 未解决 Blocker

**无。**

Plan agent 审查结论：每个 AC 均有真实、可达的入口点；无 scope 边界突破；无新外部依赖；`runtime` API 确认可用。审查发现的 3 项改进（缺少 id 检查、AC-06 降级后校验失败是正确行为、导入后需显式 `loadSchemes`）已纳入 Step 3c / Step 3d / Step 5c。
