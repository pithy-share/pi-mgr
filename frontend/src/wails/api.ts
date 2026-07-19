// Wails runtime bridge
// In a wails build, this is auto-generated. For dev, we use a manual bridge.
// @ts-ignore
const wails = window['go']?.main?.App

export interface AppAPI {
  // Scheme CRUD
  ListSchemes(): Promise<Scheme[]>
  CreateScheme(name: string): Promise<Scheme>
  UpdateScheme(scheme: Scheme): Promise<void>
  DeleteScheme(id: string): Promise<void>
  DuplicateScheme(id: string): Promise<Scheme>
  ActivateScheme(id: string): Promise<void>
  GetActiveSchemeID(): Promise<string>

  // Provider
  AddBuiltInProvider(schemeID: string, providerKey: string, apiKey: string, baseUrl: string): Promise<void>
  AddCustomProvider(schemeID: string, key: string, baseUrl: string, apiType: string, apiKey: string): Promise<void>
  UpdateProvider(schemeID: string, provider: Provider): Promise<void>
  RemoveProvider(schemeID: string, providerKey: string): Promise<void>

  // Model
  AddModel(schemeID: string, providerKey: string, model: Model): Promise<void>
  UpdateModel(schemeID: string, providerKey: string, model: Model): Promise<void>
  RemoveModel(schemeID: string, providerKey: string, modelID: string): Promise<void>
  RemoveModels(schemeID: string, providerKey: string, modelIDs: string[]): Promise<number>
  ReorderModels(schemeID: string, providerKey: string, orderedIDs: string[]): Promise<void>
  FetchProviderModels(schemeID: string, providerKey: string): Promise<Model[]>
  ImportProviderModels(schemeID: string, providerKey: string, models: Model[]): Promise<number>

  // Provider reorder
  ReorderProviders(schemeID: string, orderedKeys: string[]): Promise<void>

  // Connectivity
  TestProviderConnectivity(schemeID: string, providerKey: string): Promise<string>

  // Export / Import
  ExportSchemes(): Promise<void>
  ImportSchemes(): Promise<void>

  // SSH Sync
  TestSSHConnection(address: string): Promise<{ success: boolean; message: string }>
  SaveSSHAddress(address: string): Promise<void>
  LoadSSHAddress(): Promise<string>
  SyncPiConfig(address: string): Promise<SyncResult>

  // Catalog
  ListBuiltInProviders(): Promise<BuiltInProvider[]>
  ListAPITypes(): Promise<string[]>
}

import type { Scheme, Provider, Model, BuiltInProvider, SyncResult } from '../types'

function api(): AppAPI {
  if (wails) return wails as unknown as AppAPI
  // Fallback for browser dev
  return {
    ListSchemes: () => Promise.resolve([]),
    CreateScheme: (name: string) => Promise.resolve({ id: '1', name, providers: [] } as Scheme),
    UpdateScheme: () => Promise.resolve(),
    DeleteScheme: () => Promise.resolve(),
    DuplicateScheme: () => Promise.resolve({ id: '2', name: 'dummy', providers: [] } as Scheme),
    ActivateScheme: () => Promise.resolve(),
    GetActiveSchemeID: () => Promise.resolve(''),
    AddBuiltInProvider: () => Promise.resolve(),
    AddCustomProvider: () => Promise.resolve(),
    UpdateProvider: () => Promise.resolve(),
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
    ExportSchemes: () => Promise.resolve(),
    ImportSchemes: () => Promise.resolve(),
    TestSSHConnection: () => Promise.resolve({ success: false, message: 'dev mode' }),
    SaveSSHAddress: () => Promise.resolve(),
    LoadSSHAddress: () => Promise.resolve(''),
    SyncPiConfig: () => Promise.resolve({ overall: 'failed' as const, items: [] }),
    ListBuiltInProviders: () => Promise.resolve([]),
    ListAPITypes: () => Promise.resolve([]),
  }
}

export default api
