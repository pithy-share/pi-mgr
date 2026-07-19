## cbm 使用规则

- **代码探索走 cbm**：先定位后精读，不在 cbm 拿到位置前 `read` 代码文件
- **`read` 放行**：spec/*.md、Makefile、.gitmodules 等非代码文件直接读

```
get_architecture / search_graph           → 定向（拿 qualified_name）
search_and_read_symbols                   → 搜索+源码一步到位（探索首选）
         ↓
get_code_snippet / get_code_snippets      → 精读（已知 qualified_name 时）
resolve_symbol / read_symbol              → 已知符号名消歧+读取
         ↓
trace_path                                → 追踪调用链/数据流
         ↓
search_code                               → 精确文本（宏、配置键、字面量）
         ↓
read                                      → 回退（仅 spec、配置、或 cbm 反复失败时）
```

核心反模式：**search_graph 返回了结果，别忽略它去 read 整个文件，对 qualified_name 调 get_code_snippet。**

选工具口诀：
- 知道符号名但不确定在哪 → `resolve_symbol` 消歧，确认后 `read_symbol` 读源码
- 概念搜索想直接看代码 → `search_and_read_symbols`，省一轮往返
- 已知精确 qualified_name → `get_code_snippet` 或批量 `get_code_snippets`
- 追踪谁调了谁、数据怎么流 → `trace_path`，不用手动 grep 调用链
- 搜宏、配置键、日志字符串、字面量 → `search_code`，不是 `search_graph`
- 批量消歧读多个符号 → `read_symbols`，加 `file_path` 或 `parent_class` 消歧
