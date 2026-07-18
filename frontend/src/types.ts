export interface Scheme {
  id: string
  name: string
  providers: Provider[]
}

export interface Provider {
  key: string
  name: string
  builtIn: boolean
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

export interface Toast {
  message: string
  type: 'success' | 'error'
}

// Default model values per AC-17
export function defaultModel(): Model {
  return {
    id: '',
    name: '',
    reasoning: false,
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
