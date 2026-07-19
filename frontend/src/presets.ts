// Model presets — frontend hardcoded, no backend interaction.
// Model IDs sourced from https://pi.dev/models (2026).
// Cost fields are all 0 by design — users fill those in manually.

import type { Model } from './types'

export interface ModelPreset {
  label: string
  model: Model
}

export const MODEL_PRESETS: ModelPreset[] = [
  // --- OpenAI ---
  {
    label: 'o4-mini',
    model: {
      id: 'o4-mini',
      name: 'o4-mini',
      reasoning: true,
      inputText: true,
      inputImage: true,
      contextWindow: 1000000,
      maxTokens: 100000,
      costInput: 0,
      costOutput: 0,
      costCacheRead: 0,
      costCacheWrite: 0,
    },
  },
  {
    label: 'o3',
    model: {
      id: 'o3',
      name: 'o3',
      reasoning: true,
      inputText: true,
      inputImage: true,
      contextWindow: 200000,
      maxTokens: 100000,
      costInput: 0,
      costOutput: 0,
      costCacheRead: 0,
      costCacheWrite: 0,
    },
  },
  {
    label: 'GPT-5.6 Luna',
    model: {
      id: 'gpt-5.6-luna',
      name: 'GPT-5.6 Luna',
      reasoning: false,
      inputText: true,
      inputImage: true,
      contextWindow: 1000000,
      maxTokens: 32768,
      costInput: 0,
      costOutput: 0,
      costCacheRead: 0,
      costCacheWrite: 0,
    },
  },
  {
    label: 'GPT-5.6 Sol',
    model: {
      id: 'gpt-5.6-sol',
      name: 'GPT-5.6 Sol',
      reasoning: false,
      inputText: true,
      inputImage: true,
      contextWindow: 1000000,
      maxTokens: 32768,
      costInput: 0,
      costOutput: 0,
      costCacheRead: 0,
      costCacheWrite: 0,
    },
  },
  {
    label: 'GPT-5.5',
    model: {
      id: 'gpt-5.5',
      name: 'GPT-5.5',
      reasoning: false,
      inputText: true,
      inputImage: true,
      contextWindow: 1000000,
      maxTokens: 32768,
      costInput: 0,
      costOutput: 0,
      costCacheRead: 0,
      costCacheWrite: 0,
    },
  },

  // --- Anthropic ---
  {
    label: 'Claude Sonnet 5',
    model: {
      id: 'claude-sonnet-5',
      name: 'Claude Sonnet 5',
      reasoning: true,
      inputText: true,
      inputImage: true,
      contextWindow: 200000,
      maxTokens: 8192,
      costInput: 0,
      costOutput: 0,
      costCacheRead: 0,
      costCacheWrite: 0,
    },
  },
  {
    label: 'Claude Opus 4.8',
    model: {
      id: 'claude-opus-4-8',
      name: 'Claude Opus 4.8',
      reasoning: true,
      inputText: true,
      inputImage: true,
      contextWindow: 200000,
      maxTokens: 8192,
      costInput: 0,
      costOutput: 0,
      costCacheRead: 0,
      costCacheWrite: 0,
    },
  },
  {
    label: 'Claude Haiku 4.5',
    model: {
      id: 'claude-haiku-4-5',
      name: 'Claude Haiku 4.5',
      reasoning: false,
      inputText: true,
      inputImage: true,
      contextWindow: 200000,
      maxTokens: 8192,
      costInput: 0,
      costOutput: 0,
      costCacheRead: 0,
      costCacheWrite: 0,
    },
  },

  // --- Google ---
  {
    label: 'Gemini 3.5 Flash',
    model: {
      id: 'gemini-3.5-flash',
      name: 'Gemini 3.5 Flash',
      reasoning: false,
      inputText: true,
      inputImage: true,
      contextWindow: 1048576,
      maxTokens: 65536,
      costInput: 0,
      costOutput: 0,
      costCacheRead: 0,
      costCacheWrite: 0,
    },
  },
  {
    label: 'Gemini 2.5 Pro',
    model: {
      id: 'gemini-2.5-pro',
      name: 'Gemini 2.5 Pro',
      reasoning: false,
      inputText: true,
      inputImage: true,
      contextWindow: 1048576,
      maxTokens: 65536,
      costInput: 0,
      costOutput: 0,
      costCacheRead: 0,
      costCacheWrite: 0,
    },
  },

  // --- DeepSeek ---
  {
    label: 'DeepSeek V4 Pro',
    model: {
      id: 'deepseek-v4-pro',
      name: 'DeepSeek V4 Pro',
      reasoning: false,
      inputText: true,
      inputImage: true,
      contextWindow: 128000,
      maxTokens: 8192,
      costInput: 0,
      costOutput: 0,
      costCacheRead: 0,
      costCacheWrite: 0,
    },
  },
  {
    label: 'DeepSeek V4 Flash',
    model: {
      id: 'deepseek-v4-flash',
      name: 'DeepSeek V4 Flash',
      reasoning: false,
      inputText: true,
      inputImage: false,
      contextWindow: 128000,
      maxTokens: 8192,
      costInput: 0,
      costOutput: 0,
      costCacheRead: 0,
      costCacheWrite: 0,
    },
  },
]
