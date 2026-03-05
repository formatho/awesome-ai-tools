import { useState } from 'react'
import { Save, RefreshCw, AlertCircle, Check, Code, Settings } from 'lucide-react'

const configSections = [
  { id: 'general', label: 'General', icon: Settings },
  { id: 'api', label: 'API', icon: Code },
  { id: 'agents', label: 'Agents', icon: Settings },
  { id: 'notifications', label: 'Notifications', icon: Settings },
]

const mockConfig = {
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
  agents: {
    maxConcurrent: 10,
    defaultModel: 'gpt-4',
    defaultTemperature: 0.7,
  },
  notifications: {
    enabled: true,
    email: 'admin@formatho.com',
    slackWebhook: '',
  },
}

export default function ConfigEditor() {
  const [activeSection, setActiveSection] = useState('general')
  const [config, setConfig] = useState(mockConfig)
  const [hasChanges, setHasChanges] = useState(false)
  const [isSaving, setIsSaving] = useState(false)
  const [saveStatus, setSaveStatus] = useState<'idle' | 'success' | 'error'>('idle')

  const handleSave = async () => {
    setIsSaving(true)
    // Simulate API call
    await new Promise(resolve => setTimeout(resolve, 1000))
    setIsSaving(false)
    setHasChanges(false)
    setSaveStatus('success')
    setTimeout(() => setSaveStatus('idle'), 2000)
  }

  const handleReset = () => {
    setConfig(mockConfig)
    setHasChanges(false)
  }

  const updateConfig = (section: string, key: string, value: string | number | boolean) => {
    setConfig(prev => ({
      ...prev,
      [section]: {
        ...prev[section as keyof typeof prev],
        [key]: value,
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

          {activeSection === 'agents' && (
            <ConfigSection title="Agent Settings">
              <ConfigField label="Max Concurrent Agents" type="number" value={config.agents.maxConcurrent} onChange={(v) => updateConfig('agents', 'maxConcurrent', v)} />
              <ConfigField label="Default Model" type="text" value={config.agents.defaultModel} onChange={(v) => updateConfig('agents', 'defaultModel', v)} />
              <ConfigField label="Default Temperature" type="number" value={config.agents.defaultTemperature} onChange={(v) => updateConfig('agents', 'defaultTemperature', v)} step="0.1" />
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
        <pre className="bg-background p-4 rounded-lg overflow-auto text-sm font-mono text-text-secondary">
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
