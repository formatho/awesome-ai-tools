import { useState } from 'react'
import { Save, RefreshCw, AlertCircle, Check, Code, Settings, Cpu, Zap } from 'lucide-react'

const configSections = [
  { id: 'general', label: 'General', icon: Settings },
  { id: 'api', label: 'API', icon: Code },
  { id: 'models', label: 'LLM Models', icon: Cpu },
  { id: 'agents', label: 'Agents', icon: Settings },
  { id: 'skills', label: 'Skills', icon: Zap },
  { id: 'notifications', label: 'Notifications', icon: Settings },
]

interface ModelConfig {
  provider: 'openai' | 'anthropic' | 'ollama' | 'zai'
  modelName: string
  apiKey: string
  temperature: number
  maxTokens: number
  baseUrl?: string
}

interface Config {
  general: {
    appName: string
    version: string
    debug: boolean
    logLevel: string
  }
  api: {
    baseUrl: string
    timeout: number
    retries: number
  }
  models: {
    default: ModelConfig
    openai?: ModelConfig
    anthropic?: ModelConfig
    ollama?: ModelConfig
  }
  agents: {
    maxConcurrent: number
    defaultModel: string
    defaultTemperature: number
  }
  notifications: {
    enabled: boolean
    email: string
    slackWebhook: string
  }
}

const defaultConfig: Config = {
  general: {
    appName: 'Agent Orchestrator',
    version: '0.1.0',
    debug: false,
    logLevel: 'info',
  },
  api: {
    baseUrl: 'http://localhost:18765',
    timeout: 30000,
    retries: 3,
  },
  models: {
    default: {
      provider: 'openai',
      modelName: 'gpt-4o',
      apiKey: '',
      temperature: 0.7,
      maxTokens: 4096,
    }
  },
  agents: {
    maxConcurrent: 10,
    defaultModel: 'gpt-4o',
    defaultTemperature: 0.7,
  },
  notifications: {
    enabled: true,
    email: 'admin@formatho.com',
    slackWebhook: '',
  },
}

const PROVIDERS = [
  { value: 'openai', label: 'OpenAI' },
  { value: 'anthropic', label: 'Anthropic' },
  { value: 'ollama', label: 'Ollama (Local)' },
  { value: 'zai', label: 'Z.ai (GLM)' },
]

const MODELS_BY_PROVIDER: Record<string, Array<{ value: string; label: string }>> = {
  openai: [
    { value: 'gpt-4o', label: 'GPT-4o' },
    { value: 'gpt-4-turbo', label: 'GPT-4 Turbo' },
    { value: 'gpt-4', label: 'GPT-4' },
    { value: 'gpt-3.5-turbo', label: 'GPT-3.5 Turbo' },
  ],
  anthropic: [
    { value: 'claude-3-opus', label: 'Claude 3 Opus' },
    { value: 'claude-3-sonnet', label: 'Claude 3 Sonnet' },
    { value: 'claude-3-haiku', label: 'Claude 3 Haiku' },
    { value: 'claude-2', label: 'Claude 2' },
  ],
  ollama: [
    { value: 'llama2', label: 'Llama 2' },
    { value: 'llama3', label: 'Llama 3' },
    { value: 'codellama', label: 'Code Llama' },
    { value: 'mistral', label: 'Mistral' },
    { value: 'mixtral', label: 'Mixtral' },
  ],
  zai: [
    { value: 'glm-4.7', label: 'GLM-4.7' },
    { value: 'glm-4', label: 'GLM-4' },
    { value: 'glm-3-turbo', label: 'GLM-3 Turbo' },
  ],
}

export default function ConfigEditor() {
  const [activeSection, setActiveSection] = useState('models')
  const [config, setConfig] = useState<Config>(defaultConfig)
  const [hasChanges, setHasChanges] = useState(false)
  const [isSaving, setIsSaving] = useState(false)
  const [saveStatus, setSaveStatus] = useState<'idle' | 'success' | 'error'>('idle')
  const [testStatus, setTestStatus] = useState<'idle' | 'testing' | 'success' | 'error'>('idle')
  const [showApiKey, setShowApiKey] = useState(false)

  // Uncomment when backend is ready
  // const { data: serverConfig, isLoading } = useConfig()
  // const mutation = useConfigMutation()

  const handleSave = async () => {
    setIsSaving(true)
    setTestStatus('idle')
    
    try {
      // When backend is ready:
      // await mutation.mutateAsync(config)
      
      // For now, simulate save
      await new Promise(resolve => setTimeout(resolve, 1000))
      
      setIsSaving(false)
      setHasChanges(false)
      setSaveStatus('success')
      setTimeout(() => setSaveStatus('idle'), 2000)
    } catch (err) {
      setIsSaving(false)
      setSaveStatus('error')
      setTimeout(() => setSaveStatus('idle'), 3000)
    }
  }

  const handleReset = () => {
    setConfig(defaultConfig)
    setHasChanges(false)
  }

  const handleTestConnection = async () => {
    setTestStatus('testing')
    
    try {
      const modelConfig = config.models.default
      
      // Call the actual API to test LLM connection
      const response = await fetch(`${config.api.baseUrl}/api/config/test-llm`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          provider: modelConfig.provider,
          api_key: modelConfig.apiKey,
          model: modelConfig.modelName,
          base_url: modelConfig.baseUrl,
        }),
      })
      
      const result = await response.json()
      
      if (result.success) {
        setTestStatus('success')
        setTimeout(() => setTestStatus('idle'), 3000)
      } else {
        console.error('LLM test failed:', result.message)
        setTestStatus('error')
        setTimeout(() => setTestStatus('idle'), 5000)
      }
    } catch (err) {
      console.error('Failed to test LLM connection:', err)
      setTestStatus('error')
      setTimeout(() => setTestStatus('idle'), 5000)
    }
  }

  const updateConfig = (section: keyof Config, key: string, value: string | number | boolean) => {
    setConfig(prev => ({
      ...prev,
      [section]: {
        ...prev[section],
        [key]: value,
      },
    }))
    setHasChanges(true)
  }

  const updateModelConfig = (key: keyof ModelConfig, value: string | number) => {
    setConfig(prev => ({
      ...prev,
      models: {
        ...prev.models,
        default: {
          ...prev.models.default,
          [key]: value,
        },
      },
    }))
    setHasChanges(true)
  }

  return (
    <div className="space-y-6 animate-fade-in">
      {/* Page Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-text-primary">Configuration</h1>
          <p className="text-text-secondary mt-1">Manage application settings</p>
        </div>
        <div className="flex items-center gap-2">
          {hasChanges && (
            <button onClick={handleReset} className="btn-secondary">
              <RefreshCw className="w-4 h-4 mr-2" />
              Reset
            </button>
          )}
          <button
            onClick={handleSave}
            disabled={!hasChanges || isSaving}
            className={`btn-primary ${(!hasChanges || isSaving) ? 'opacity-50 cursor-not-allowed' : ''}`}
          >
            {isSaving ? (
              <>
                <RefreshCw className="w-4 h-4 mr-2 animate-spin" />
                Saving...
              </>
            ) : saveStatus === 'success' ? (
              <>
                <Check className="w-4 h-4 mr-2" />
                Saved!
              </>
            ) : (
              <>
                <Save className="w-4 h-4 mr-2" />
                Save Changes
              </>
            )}
          </button>
        </div>
      </div>

      {/* Config Editor */}
      <div className="grid grid-cols-1 lg:grid-cols-4 gap-6">
        {/* Sidebar */}
        <div className="lg:col-span-1">
          <nav className="space-y-1">
            {configSections.map((section) => (
              <button
                key={section.id}
                onClick={() => setActiveSection(section.id)}
                className={`sidebar-item w-full ${activeSection === section.id ? 'active' : ''}`}
              >
                <section.icon className="w-5 h-5" />
                <span>{section.label}</span>
              </button>
            ))}
          </nav>
        </div>

        {/* Content */}
        <div className="lg:col-span-3">
          {saveStatus === 'error' && (
            <div className="mb-4 p-4 bg-error/10 border border-error/20 rounded-lg flex items-center gap-2 text-error">
              <AlertCircle className="w-5 h-5" />
              Failed to save configuration. Please try again.
            </div>
          )}

          {activeSection === 'general' && (
            <ConfigSection title="General Settings">
              <ConfigField label="App Name" type="text" value={config.general.appName} onChange={(v) => updateConfig('general', 'appName', v)} />
              <ConfigField label="Version" type="text" value={config.general.version} disabled />
              <ConfigField label="Debug Mode" type="checkbox" value={config.general.debug} onChange={(v) => updateConfig('general', 'debug', v)} />
              <ConfigField label="Log Level" type="select" value={config.general.logLevel} options={['debug', 'info', 'warn', 'error']} onChange={(v) => updateConfig('general', 'logLevel', v)} />
            </ConfigSection>
          )}

          {activeSection === 'api' && (
            <ConfigSection title="API Settings">
              <ConfigField label="Base URL" type="text" value={config.api.baseUrl} onChange={(v) => updateConfig('api', 'baseUrl', v)} />
              <ConfigField label="Timeout (ms)" type="number" value={config.api.timeout} onChange={(v) => updateConfig('api', 'timeout', v)} />
              <ConfigField label="Retries" type="number" value={config.api.retries} onChange={(v) => updateConfig('api', 'retries', v)} />
            </ConfigSection>
          )}

          {activeSection === 'models' && (
            <ConfigSection title="LLM Model Configuration">
              <div className="space-y-6">
                {/* Provider Selection */}
                <div>
                  <label className="block text-sm font-medium text-text-secondary mb-2">
                    Provider
                  </label>
                  <select
                    value={config.models.default.provider}
                    onChange={(e) => {
                      const provider = e.target.value as 'openai' | 'anthropic' | 'ollama' | 'zai'
                      const defaultModel = MODELS_BY_PROVIDER[provider][0].value
                      updateModelConfig('provider', provider)
                      updateModelConfig('modelName', defaultModel)
                    }}
                    className="input w-full"
                  >
                    {PROVIDERS.map((p) => (
                      <option key={p.value} value={p.value}>{p.label}</option>
                    ))}
                  </select>
                </div>

                {/* Model Selection */}
                <div>
                  <label className="block text-sm font-medium text-text-secondary mb-2">
                    Model
                  </label>
                  <select
                    value={config.models.default.modelName}
                    onChange={(e) => updateModelConfig('modelName', e.target.value)}
                    className="input w-full"
                  >
                    {MODELS_BY_PROVIDER[config.models.default.provider]?.map((m) => (
                      <option key={m.value} value={m.value}>{m.label}</option>
                    ))}
                  </select>
                </div>

                {/* API Key */}
                {config.models.default.provider !== 'ollama' && (
                  <div>
                    <label className="block text-sm font-medium text-text-secondary mb-2">
                      API Key
                    </label>
                    <div className="relative">
                      <input
                        type={showApiKey ? 'text' : 'password'}
                        value={config.models.default.apiKey}
                        onChange={(e) => updateModelConfig('apiKey', e.target.value)}
                        placeholder={config.models.default.provider === 'openai' ? 'sk-...' : 'sk-ant-...'}
                        className="input w-full pr-20"
                      />
                      <button
                        type="button"
                        onClick={() => setShowApiKey(!showApiKey)}
                        className="absolute right-2 top-1/2 -translate-y-1/2 text-xs text-text-muted hover:text-text-primary"
                      >
                        {showApiKey ? 'Hide' : 'Show'}
                      </button>
                    </div>
                    <p className="text-xs text-text-muted mt-1">
                      Your API key is stored securely and never shared
                    </p>
                  </div>
                )}

                {/* Base URL for Ollama or ZAI */}
                {(config.models.default.provider === 'ollama' || config.models.default.provider === 'zai') && (
                  <div>
                    <label className="block text-sm font-medium text-text-secondary mb-2">
                      {config.models.default.provider === 'ollama' ? 'Ollama Base URL' : 'Z.ai Base URL'}
                    </label>
                    <input
                      type="text"
                      value={config.models.default.baseUrl || (config.models.default.provider === 'ollama' ? 'http://localhost:11434' : 'https://open.bigmodel.cn/api/paas/v4')}
                      onChange={(e) => updateModelConfig('baseUrl', e.target.value)}
                      placeholder={config.models.default.provider === 'ollama' ? 'http://localhost:11434' : 'https://open.bigmodel.cn/api/paas/v4'}
                      className="input w-full"
                    />
                  </div>
                )}

                {/* Temperature Slider */}
                <div>
                  <label className="block text-sm font-medium text-text-secondary mb-2">
                    Temperature: {config.models.default.temperature.toFixed(1)}
                  </label>
                  <input
                    type="range"
                    min="0"
                    max="2"
                    step="0.1"
                    value={config.models.default.temperature}
                    onChange={(e) => updateModelConfig('temperature', parseFloat(e.target.value))}
                    className="w-full h-2 bg-surface-hover rounded-lg appearance-none cursor-pointer"
                  />
                  <div className="flex justify-between text-xs text-text-muted mt-1">
                    <span>Focused (0)</span>
                    <span>Balanced (1)</span>
                    <span>Creative (2)</span>
                  </div>
                </div>

                {/* Max Tokens */}
                <div>
                  <label className="block text-sm font-medium text-text-secondary mb-2">
                    Max Tokens
                  </label>
                  <input
                    type="number"
                    min="100"
                    max="128000"
                    value={config.models.default.maxTokens}
                    onChange={(e) => updateModelConfig('maxTokens', parseInt(e.target.value))}
                    className="input w-full"
                  />
                  <p className="text-xs text-text-muted mt-1">
                    Maximum number of tokens in the response
                  </p>
                </div>

                {/* Test Connection Button */}
                <div className="pt-4 border-t border-border">
                  <button
                    onClick={handleTestConnection}
                    disabled={testStatus === 'testing'}
                    className={`btn-secondary ${testStatus === 'testing' ? 'opacity-50 cursor-not-allowed' : ''}`}
                  >
                    {testStatus === 'testing' ? (
                      <>
                        <RefreshCw className="w-4 h-4 mr-2 animate-spin" />
                        Testing Connection...
                      </>
                    ) : testStatus === 'success' ? (
                      <>
                        <Check className="w-4 h-4 mr-2 text-success" />
                        Connection Successful!
                      </>
                    ) : testStatus === 'error' ? (
                      <>
                        <AlertCircle className="w-4 h-4 mr-2 text-error" />
                        Connection Failed
                      </>
                    ) : (
                      <>
                        <Zap className="w-4 h-4 mr-2" />
                        Test Connection
                      </>
                    )}
                  </button>
                  
                  {testStatus === 'success' && (
                    <p className="text-sm text-success mt-2">
                      ✓ Successfully connected to {PROVIDERS.find(p => p.value === config.models.default.provider)?.label}
                    </p>
                  )}
                  
                  {testStatus === 'error' && (
                    <p className="text-sm text-error mt-2">
                      ✗ Failed to connect. Please check your API key and try again.
                    </p>
                  )}
                </div>
              </div>
            </ConfigSection>
          )}

          {activeSection === 'agents' && (
            <ConfigSection title="Agent Settings">
              <ConfigField label="Max Concurrent Agents" type="number" value={config.agents.maxConcurrent} onChange={(v) => updateConfig('agents', 'maxConcurrent', v)} />
              <ConfigField label="Default Model" type="text" value={config.agents.defaultModel} onChange={(v) => updateConfig('agents', 'defaultModel', v)} />
              <ConfigField label="Default Temperature" type="number" value={config.agents.defaultTemperature} onChange={(v) => updateConfig('agents', 'defaultTemperature', v)} step="0.1" />
            </ConfigSection>
          )}

          {activeSection === 'skills' && (
            <ConfigSection title="Skills Management">
              <div className="space-y-4">
                <p className="text-text-secondary">
                  Create and manage agent skills using markdown. Skills define capabilities and permissions for agents.
                </p>
                <a
                  href="#/skills"
                  className="btn-primary inline-flex items-center"
                >
                  <Zap className="w-4 h-4 mr-2" />
                  Open Skills Editor
                </a>
                
                <div className="mt-6 p-4 bg-surface-hover rounded-lg border border-border">
                  <h4 className="font-semibold mb-2">Quick Tips</h4>
                  <ul className="text-sm text-text-secondary space-y-2">
                    <li>• Skills are defined in markdown format</li>
                    <li>• Each skill has a name, description, and permissions</li>
                    <li>• Permissions control what actions the skill can perform</li>
                    <li>• Skills can be assigned to agents to give them capabilities</li>
                  </ul>
                </div>
              </div>
            </ConfigSection>
          )}

          {activeSection === 'notifications' && (
            <ConfigSection title="Notification Settings">
              <ConfigField label="Enable Notifications" type="checkbox" value={config.notifications.enabled} onChange={(v) => updateConfig('notifications', 'enabled', v)} />
              <ConfigField label="Email" type="text" value={config.notifications.email} onChange={(v) => updateConfig('notifications', 'email', v)} />
              <ConfigField label="Slack Webhook URL" type="text" value={config.notifications.slackWebhook} onChange={(v) => updateConfig('notifications', 'slackWebhook', v)} placeholder="https://hooks.slack.com/..." />
            </ConfigSection>
          )}
        </div>
      </div>

      {/* Raw JSON Preview */}
      <div className="card">
        <h3 className="text-lg font-semibold mb-4">Raw Configuration</h3>
        <pre className="bg-background p-4 rounded-lg overflow-auto text-sm font-mono text-text-secondary max-h-96">
          {JSON.stringify(config, null, 2)}
        </pre>
      </div>
    </div>
  )
}

function ConfigSection({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div className="card">
      <h3 className="text-lg font-semibold mb-6">{title}</h3>
      <div className="space-y-6">
        {children}
      </div>
    </div>
  )
}

interface ConfigFieldProps {
  label: string
  type: 'text' | 'number' | 'checkbox' | 'select'
  value: string | number | boolean
  onChange?: (value: string | number | boolean) => void
  options?: string[]
  disabled?: boolean
  placeholder?: string
  step?: string
}

function ConfigField({ label, type, value, onChange, options, disabled, placeholder, step }: ConfigFieldProps) {
  if (type === 'checkbox') {
    return (
      <div className="flex items-center justify-between">
        <label className="label mb-0">{label}</label>
        <button
          onClick={() => onChange?.(!value)}
          className={`w-12 h-6 rounded-full transition-colors ${value ? 'bg-accent' : 'bg-border'}`}
        >
          <div className={`w-5 h-5 rounded-full bg-white shadow transition-transform ${value ? 'translate-x-6' : 'translate-x-0.5'}`} />
        </button>
      </div>
    )
  }

  if (type === 'select') {
    return (
      <div>
        <label className="label">{label}</label>
        <select
          value={value as string}
          onChange={(e) => onChange?.(e.target.value)}
          disabled={disabled}
          className="input"
        >
          {options?.map((opt) => (
            <option key={opt} value={opt}>{opt}</option>
          ))}
        </select>
      </div>
    )
  }

  return (
    <div>
      <label className="label">{label}</label>
      <input
        type={type}
        value={value as string | number}
        onChange={(e) => onChange?.(type === 'number' ? parseFloat(e.target.value) : e.target.value)}
        disabled={disabled}
        placeholder={placeholder}
        step={step}
        className={`input ${disabled ? 'opacity-50 cursor-not-allowed' : ''}`}
      />
    </div>
  )
}
