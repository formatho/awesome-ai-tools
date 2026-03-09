export interface Agent {
  id: string
  name: string
  description?: string
  model?: string
  provider?: string
  status: 'idle' | 'running' | 'stopped' | 'error'
  created_at: string
  updated_at: string
  config?: Record<string, unknown>
}

export interface AgentLog {
  id: string
  agent_id: string
  level: 'debug' | 'info' | 'warn' | 'error'
  message: string
  metadata?: Record<string, unknown> | null
  created_at: string
}

export interface Config {
  id: string
  llm_config?: LLMConfig
  defaults?: Record<string, unknown>
  settings?: Record<string, unknown>
  updated_at?: string
}

export interface LLMConfig {
  provider?: string
  model?: string
  api_key?: string
  base_url?: string
}
