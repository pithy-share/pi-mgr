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
}

import type { Config, Provider, Model, BuiltInProvider, SyncResult } from '../types'

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
  }
}

export default api
