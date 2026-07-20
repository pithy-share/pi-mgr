// Wails runtime bridge
// In a wails build, this is auto-generated. For dev, we use a manual bridge.
// @ts-ignore
const wails = window['go']?.main?.App

export interface AppAPI {
  // Config
  GetConfig(): Promise<Config>

  // Provider
  AddBuiltInProvider(providerKey: string, apiKey: string, baseUrl: string): Promise<void>
  AddCustomProvider(key: string, baseUrl: string, apiType: string, apiKey: string): Promise<void>
  UpdateProvider(provider: Provider): Promise<void>
  SetProviderEnabled(providerKey: string, enabled: boolean): Promise<void>
  RemoveProvider(providerKey: string): Promise<void>

  // Model
  AddModel(providerKey: string, model: Model): Promise<void>
  UpdateModel(providerKey: string, model: Model): Promise<void>
  RemoveModel(providerKey: string, modelID: string): Promise<void>
  RemoveModels(providerKey: string, modelIDs: string[]): Promise<number>
  ReorderModels(providerKey: string, orderedIDs: string[]): Promise<void>
  FetchProviderModels(providerKey: string): Promise<Model[]>
  ImportProviderModels(providerKey: string, models: Model[]): Promise<number>

  // Provider reorder
  ReorderProviders(orderedKeys: string[]): Promise<void>

  // Connectivity
  TestProviderConnectivity(providerKey: string): Promise<string>

  // Export / Import
  ExportConfig(): Promise<void>
  ImportConfig(): Promise<void>

  // SSH Sync
  TestSSHConnection(address: string): Promise<{ success: boolean; message: string }>
  SaveSSHAddress(address: string): Promise<void>
  LoadSSHAddress(): Promise<string>
  SyncPiConfig(address: string): Promise<SyncResult>

  // Catalog
  ListBuiltInProviders(): Promise<BuiltInProvider[]>
  ListAPITypes(): Promise<string[]>

  // Pi Management
  GetPiVersion(): Promise<string>
  GetPiPackages(): Promise<string>
  UpdatePiSelf(): Promise<string>
  UpdateAllPiPackages(): Promise<string>
  UpdatePiPackage(source: string): Promise<string>
  RemovePiPackage(source: string): Promise<string>

  // Pi Management (Remote via SSH)
  GetRemotePiVersion(sshAddress: string): Promise<string>
  GetRemotePiPackages(sshAddress: string): Promise<string>
  UpdateRemotePiSelf(sshAddress: string): Promise<string>
  UpdateRemoteAllPiPackages(sshAddress: string): Promise<string>
  UpdateRemotePiPackage(sshAddress: string, source: string): Promise<string>
  RemoveRemotePiPackage(sshAddress: string, source: string): Promise<string>

  // Pi Install
  InstallPiPackage(source: string): Promise<string>
  InstallRemotePiPackage(sshAddress: string, source: string): Promise<string>

  // Codebase Memory
  GetCodebaseMemoryRules(): Promise<string>
  GetCodebaseMemoryMCPConfig(): Promise<string>

  // Built-in Prompt Templates
  ListBuiltInPrompts(): Promise<PromptTemplate[]>
  InstallBuiltInPrompts(names: string[]): Promise<number>
  RemoveInstalledPrompt(name: string): Promise<void>
  GetBuiltInPromptContent(name: string): Promise<string>
}

import type { Config, Provider, Model, BuiltInProvider, SyncResult, PromptTemplate } from '../types'

function api(): AppAPI {
  if (wails) return wails as unknown as AppAPI
  // Fallback for browser dev
  return {
    GetConfig: () => Promise.resolve({ providers: [] } as Config),
    AddBuiltInProvider: () => Promise.resolve(),
    AddCustomProvider: () => Promise.resolve(),
    UpdateProvider: () => Promise.resolve(),
    SetProviderEnabled: () => Promise.resolve(),
    RemoveProvider: () => Promise.resolve(),
    AddModel: () => Promise.resolve(),
    UpdateModel: () => Promise.resolve(),
    RemoveModel: () => Promise.resolve(),
    RemoveModels: () => Promise.resolve(0),
    ReorderModels: () => Promise.resolve(),
    FetchProviderModels: () => Promise.resolve([] as Model[]),
    ImportProviderModels: () => Promise.resolve(0),
    ReorderProviders: () => Promise.resolve(),
    TestProviderConnectivity: () => Promise.resolve(''),
    ExportConfig: () => Promise.resolve(),
    ImportConfig: () => Promise.resolve(),
    TestSSHConnection: () => Promise.resolve({ success: false, message: 'dev mode' }),
    SaveSSHAddress: () => Promise.resolve(),
    LoadSSHAddress: () => Promise.resolve(''),
    SyncPiConfig: () => Promise.resolve({ overall: 'failed' as const, items: [] }),
    ListBuiltInProviders: () => Promise.resolve([]),
    ListAPITypes: () => Promise.resolve([]),
    GetPiVersion: () => Promise.resolve('0.80.10 (dev mode)'),
    GetPiPackages: () => Promise.resolve(''),
    UpdatePiSelf: () => Promise.resolve('dev mode: upgrade skipped'),
    UpdateAllPiPackages: () => Promise.resolve('dev mode: upgrade skipped'),
    UpdatePiPackage: () => Promise.resolve('dev mode: upgrade skipped'),
    RemovePiPackage: () => Promise.resolve('dev mode: remove skipped'),
    GetRemotePiVersion: () => Promise.resolve('0.80.10 (remote dev)'),
    GetRemotePiPackages: () => Promise.resolve(''),
    UpdateRemotePiSelf: () => Promise.resolve('dev mode: remote upgrade skipped'),
    UpdateRemoteAllPiPackages: () => Promise.resolve('dev mode: remote upgrade skipped'),
    UpdateRemotePiPackage: () => Promise.resolve('dev mode: remote upgrade skipped'),
    RemoveRemotePiPackage: () => Promise.resolve('dev mode: remote remove skipped'),
    InstallPiPackage: () => Promise.resolve('dev mode: install skipped'),
    InstallRemotePiPackage: () => Promise.resolve('dev mode: remote install skipped'),
    GetCodebaseMemoryRules: () => Promise.resolve('## Codebase Memory 使用规则\n\n开发模式示例内容'),
    GetCodebaseMemoryMCPConfig: () => Promise.resolve('{\n  "mcpServers": {\n    "codebase-memory": {\n      "command": "C:\\\\Users\\\\Administrator\\\\\\.local\\\\bin\\\\codebase-memory-mcp.exe",\n      "args": [],\n      "lifecycle": "keep-alive",\n      "requestTimeoutMs": 300000,\n      "directTools": [\n        "index_repository",\n        "index_status"\n      ],\n      "excludeTools": [\n        "delete_project"\n      ]\n    }\n  }\n}'),
    // Built-in Prompt Templates
    ListBuiltInPrompts: () => Promise.resolve([
      { name: 'implement', description: '按任务计划实现、验证并将结果保存到同一任务目录', argumentHint: '<task/<task-name>/plan.md>', installed: false },
      { name: 'to_plan', description: '基于已落盘 PRD 生成计划并保存到同一任务目录', argumentHint: '<task/<task-name>/prd.md>', installed: false },
      { name: 'to_prd', description: '生成 PRD 并保存到独立任务目录', argumentHint: '[简短摘要或需求文件]', installed: false },
      { name: 'to_spec', description: '基于项目代码初始化或增量更新供 AI 编程使用的 AGENTS.md/spec 知识库', argumentHint: '[task/<task-name> 或范围]', installed: false },
    ]),
    InstallBuiltInPrompts: () => Promise.resolve(4),
    RemoveInstalledPrompt: () => Promise.resolve(),
    GetBuiltInPromptContent: () => Promise.resolve('# 开发模式\n\n示例内容'),
  }
}

export default api
