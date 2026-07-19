export interface Config {
  providers: Provider[]
}

export interface Provider {
  key: string
  name: string
  builtIn: boolean
  enabled: boolean
  apiKey: string
  baseUrl: string
  apiType: string
  models: Model[]
}

export interface Model {
  id: string
  name: string
  reasoning: boolean
  inputText: boolean
  inputImage: boolean
  contextWindow: number
  maxTokens: number
  costInput: number
  costOutput: number
  costCacheRead: number
  costCacheWrite: number
}

export interface BuiltInProvider {
  key: string
  name: string
  apiType: string
}

export interface SyncItemStatus {
  name: string
  status: 'success' | 'skipped' | 'failed'
  message: string
}

export interface SyncResult {
  overall: 'success' | 'partial' | 'failed'
  items: SyncItemStatus[]
}

export interface Toast {
  message: string
  type: 'success' | 'error'
}

export interface PiPackage {
  source: string    // 完整来源标识，如 "npm:@foo/bar"
  path: string      // 安装路径
}

export interface PromptTemplate {
  name: string
  description: string
  argumentHint: string
  installed: boolean
}

// Default model values per AC-17
export function defaultModel(): Model {
  return {
    id: '',
    name: '',
    reasoning: true,
    inputText: true,
    inputImage: false,
    contextWindow: 256000,
    maxTokens: 64000,
    costInput: 0,
    costOutput: 0,
    costCacheRead: 0,
    costCacheWrite: 0,
  }
}
