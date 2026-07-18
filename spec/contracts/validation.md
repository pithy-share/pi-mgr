# 校验规则与错误矩阵

**阅读时机**：改动前后端校验逻辑、错误消息、表单提交处理时。  
**可核验依据**：`validate.go`（ValidateScheme, ValidateProvider, ValidateModel），`api.go` 中各变更方法调用前的校验逻辑。

## 校验职责

前端做即时校验供用户体验，后端在所有变更写入前重复校验（安全网）。所有校验函数返回 `[]string` 错误消息列表，空列表表示通过。

## 校验函数

```go
func ValidateScheme(scheme *Scheme) []string
func ValidateProvider(prov *Provider, allProviders []Provider) []string
func ValidateModel(m *Model, existingModels []Model) []string
```

## 错误矩阵

### Scheme 级

| 条件 | 错误 | AC |
|---|---|---|
| Name 为空或仅空白 | "方案名称不能为空" | AC-24 |

### Provider 级

| 条件 | 错误 | 适用范围 | AC |
|---|---|---|---|
| Key 为空 | "供应商标识不能为空" | 所有 | — |
| !BuiltIn 且 BaseURL 为空 | "自定义供应商的 baseUrl 不能为空" | 仅自定义 | AC-25 |
| !BuiltIn 且 APIType 为空 | "自定义供应商的 API 类型不能为空" | 仅自定义 | AC-26 |
| Key 与同 Scheme 内其他 Provider.Key 重复 | "供应商标识已存在" | 所有 | AC-29 / AC-33 |
| BuiltIn 且同 Scheme 已有同 Key 的内置 | "该内置供应商已添加" | 仅内置 | AC-33 |

**注意**：内置供应商通过 `BuiltIn` 标记防止重复添加（AC-33）；自定义供应商通过 Key 唯一性防止重复（AC-29）。两者共用同一校验：同一 Scheme 内所有 Provider.Key 必须唯一。

### Model 级

| 条件 | 错误 | AC |
|---|---|---|
| ID 为空或仅空白 | "模型 ID 不能为空" | AC-27 |
| ID 与同 Provider 内已有 Model.ID 重复 | "模型 ID 在该供应商下已存在" | AC-28 |

## 调用时机

1. **前端**：表单 onBlur / onSubmit 时调用对应校验，显示内联错误
2. **后端 API**：每个变更方法在执行写操作前调用校验，校验失败返回 error 而不写入

## 不在校验范围的事项（明确排除）

- API key 有效性（不做网络请求）
- baseUrl 可访问性
- 模型 ID 与 pi 内置模型 ID 冲突（pi 运行时按 upsert 处理）
- 跨方案数据一致性（方案互相独立）
