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

  // Provider
  AddBuiltInProvider(schemeID: string, providerKey: string, apiKey: string, baseUrl: string): Promise<void>
  AddCustomProvider(schemeID: string, key: string, baseUrl: string, apiType: string, apiKey: string): Promise<void>
  UpdateProvider(schemeID: string, provider: Provider): Promise<void>
  RemoveProvider(schemeID: string, providerKey: string): Promise<void>

  // Model
  AddModel(schemeID: string, providerKey: string, model: Model): Promise<void>
  UpdateModel(schemeID: string, providerKey: string, model: Model): Promise<void>
  RemoveModel(schemeID: string, providerKey: string, modelID: string): Promise<void>

  // Export / Import
  ExportSchemes(): Promise<void>
  ImportSchemes(): Promise<void>

  // Catalog
  ListBuiltInProviders(): Promise<BuiltInProvider[]>
  ListAPITypes(): Promise<string[]>
}

import type { Scheme, Provider, Model, BuiltInProvider } from '../types'

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
    AddBuiltInProvider: () => Promise.resolve(),
    AddCustomProvider: () => Promise.resolve(),
    UpdateProvider: () => Promise.resolve(),
    RemoveProvider: () => Promise.resolve(),
    AddModel: () => Promise.resolve(),
    UpdateModel: () => Promise.resolve(),
    RemoveModel: () => Promise.resolve(),
    ExportSchemes: () => Promise.resolve(),
    ImportSchemes: () => Promise.resolve(),
    ListBuiltInProviders: () => Promise.resolve([]),
    ListAPITypes: () => Promise.resolve([]),
  }
}

export default api
