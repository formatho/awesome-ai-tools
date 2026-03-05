import { useState } from 'react'
import { Link } from 'react-router-dom'
import { Search, Plus, Filter, MoreVertical, Play, Pause, Trash2, X, MessageSquare } from 'lucide-react'
import { useAgents, useAgentMutations } from '../../hooks/useAPI'

interface Agent {
  id: string
  name: string
  type: string
  model?: string
  status: 'running' | 'idle' | 'error'
  lastActive: string
  tasksCompleted: number
}

const AVAILABLE_MODELS = [
  { value: 'gpt-4o', label: 'GPT-4o' },
  { value: 'gpt-4-turbo', label: 'GPT-4 Turbo' },
  { value: 'claude-3-opus', label: 'Claude 3 Opus' },
  { value: 'claude-3-sonnet', label: 'Claude 3 Sonnet' },
  { value: 'ollama/llama2', label: 'Llama 2 (Ollama)' },
]

const AGENT_TYPES = [
  'Code Analysis',
  'Web Scraping',
  'Documentation',
  'Testing',
  'Monitoring',
  'Data Processing',
  'Automation',
  'Other',
]

interface CreateAgentModalProps {
  isOpen: boolean
  onClose: () => void
  onSubmit: (data: CreateAgentRequest) => Promise<void>
}

interface CreateAgentRequest {
  name: string
  provider: string
  type: string
  model: string
  work_dir: string
}

function CreateAgentModal({ isOpen, onClose, onSubmit }: CreateAgentModalProps) {
  const [formData, setFormData] = useState<CreateAgentRequest>({
    name: '',
    provider: 'openai',
    type: AGENT_TYPES[0],
    model: 'gpt-4o',
    work_dir: '~/sandbox',
  })
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!formData.name.trim()) {
      setError('Agent name is required')
      return
    }
    
    setIsSubmitting(true)
    setError(null)
    
    try {
      await onSubmit(formData)
      setFormData({ name: '', provider: 'openai', type: AGENT_TYPES[0], model: 'gpt-4o', work_dir: '~/sandbox' })
      onClose()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to create agent')
    } finally {
      setIsSubmitting(false)
    }
  }

  if (!isOpen) return null

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm animate-fade-in">
      <div className="bg-surface border border-border rounded-lg shadow-xl w-full max-w-md mx-4">
        <div className="flex items-center justify-between p-4 border-b border-border">
          <h2 className="text-xl font-semibold text-text-primary">Create New Agent</h2>
          <button
            onClick={onClose}
            className="p-2 hover:bg-surface-hover rounded-lg text-text-muted hover:text-text-primary transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>
        
        <form onSubmit={handleSubmit} className="p-6 space-y-4">
          {error && (
            <div className="p-3 bg-error/10 border border-error/20 rounded-lg text-error text-sm">
              {error}
            </div>
          )}
          
          <div>
            <label className="block text-sm font-medium text-text-secondary mb-2">
              Agent Name
            </label>
            <input
              type="text"
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              placeholder="e.g., code-reviewer"
              className="input w-full"
              autoFocus
            />
          </div>
          
          <div>
            <label className="block text-sm font-medium text-text-secondary mb-2">
              Provider
            </label>
            <select
              value={formData.provider}
              onChange={(e) => setFormData({ ...formData, provider: e.target.value })}
              className="input w-full"
            >
              <option value="openai">OpenAI</option>
              <option value="anthropic">Anthropic</option>
              <option value="ollama">Ollama (Local)</option>
            </select>
          </div>
          
          <div>
            <label className="block text-sm font-medium text-text-secondary mb-2">
              Model
            </label>
            <select
              value={formData.model}
              onChange={(e) => setFormData({ ...formData, model: e.target.value })}
              className="input w-full"
            >
              {AVAILABLE_MODELS.map((model) => (
                <option key={model.value} value={model.value}>{model.label}</option>
              ))}
            </select>
          </div>
          
          <div>
            <label className="block text-sm font-medium text-text-secondary mb-2">
              Working Directory
            </label>
            <input
              type="text"
              value={formData.work_dir}
              onChange={(e) => setFormData({ ...formData, work_dir: e.target.value })}
              placeholder="e.g., ~/sandbox or /Users/you/projects"
              className="input w-full font-mono text-sm"
            />
            <p className="mt-1 text-xs text-text-muted">Directory where the agent will execute tasks</p>
          </div>
          
          <div className="flex gap-3 pt-4">
            <button
              type="button"
              onClick={onClose}
              className="btn-secondary flex-1"
              disabled={isSubmitting}
            >
              Cancel
            </button>
            <button
              type="submit"
              className="btn-primary flex-1"
              disabled={isSubmitting}
            >
              {isSubmitting ? 'Creating...' : 'Create Agent'}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}

export default function AgentList() {
  const [search, setSearch] = useState('')
  const [filter, setFilter] = useState<string>('all')
  const [showCreateModal, setShowCreateModal] = useState(false)
  const [toast, setToast] = useState<{ type: 'success' | 'error'; message: string } | null>(null)

  const { data: agents, isLoading, error } = useAgents()
  const mutations = useAgentMutations()

  const showToast = (type: 'success' | 'error', message: string) => {
    setToast({ type, message })
    setTimeout(() => setToast(null), 3000)
  }

  const handleCreateAgent = async (data: CreateAgentRequest) => {
    await mutations.create.mutateAsync(data)
    showToast('success', `Agent "${data.name}" created successfully!`)
  }

  const handleDeleteAgent = async (id: string, name: string) => {
    if (confirm(`Are you sure you want to delete agent "${name}"?`)) {
      try {
        await mutations.delete.mutateAsync(id)
        showToast('success', `Agent "${name}" deleted successfully!`)
      } catch (err) {
        showToast('error', 'Failed to delete agent')
      }
    }
  }

  const handleToggleAgent = async (agent: Agent) => {
    try {
      if (agent.status === 'running') {
        await mutations.stop.mutateAsync(agent.id)
        showToast('success', `Agent "${agent.name}" stopped`)
      } else {
        await mutations.start.mutateAsync(agent.id)
        showToast('success', `Agent "${agent.name}" started`)
      }
    } catch (err) {
      showToast('error', 'Failed to toggle agent')
    }
  }

  const filteredAgents = (agents || []).filter((agent: Agent) => {
    const matchesSearch = agent.name.toLowerCase().includes(search.toLowerCase())
    const matchesFilter = filter === 'all' || agent.status === filter
    return matchesSearch && matchesFilter
  })

  if (error) {
    return (
      <div className="card text-center py-12">
        <p className="text-error">Failed to load agents. Please check if the backend is running.</p>
      </div>
    )
  }

  return (
    <div className="space-y-6 animate-fade-in">
      {/* Toast Notification */}
      {toast && (
        <div className={`fixed top-4 right-4 z-50 p-4 rounded-lg shadow-lg animate-fade-in ${
          toast.type === 'success' ? 'bg-success/20 border border-success/30 text-success' : 'bg-error/20 border border-error/30 text-error'
        }`}>
          {toast.message}
        </div>
      )}

      {/* Page Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-text-primary">Agents</h1>
          <p className="text-text-secondary mt-1">Manage and monitor your AI agents</p>
        </div>
        <button 
          onClick={() => setShowCreateModal(true)}
          className="btn-primary"
        >
          <Plus className="w-4 h-4 mr-2" />
          New Agent
        </button>
      </div>

      {/* Search and Filter */}
      <div className="flex gap-4">
        <div className="relative flex-1">
          <Search className="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-text-muted" />
          <input
            type="text"
            placeholder="Search agents..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="input pl-10"
          />
        </div>
        <div className="flex items-center gap-2">
          <Filter className="w-4 h-4 text-text-muted" />
          <select
            value={filter}
            onChange={(e) => setFilter(e.target.value)}
            className="input w-auto"
          >
            <option value="all">All Status</option>
            <option value="running">Running</option>
            <option value="idle">Idle</option>
            <option value="error">Error</option>
          </select>
        </div>
      </div>

      {/* Loading State */}
      {isLoading && (
        <div className="card text-center py-12">
          <p className="text-text-muted">Loading agents...</p>
        </div>
      )}

      {/* Agent List */}
      {!isLoading && (
        <div className="space-y-3">
          {filteredAgents.map((agent: Agent) => (
            <AgentCard 
              key={agent.id} 
              agent={agent} 
              onToggle={() => handleToggleAgent(agent)}
              onDelete={() => handleDeleteAgent(agent.id, agent.name)}
            />
          ))}
        </div>
      )}

      {!isLoading && filteredAgents.length === 0 && (
        <div className="card text-center py-12">
          <p className="text-text-muted">No agents found</p>
          <button 
            onClick={() => setShowCreateModal(true)}
            className="btn-primary mt-4"
          >
            Create your first agent
          </button>
        </div>
      )}

      {/* Create Agent Modal */}
      <CreateAgentModal
        isOpen={showCreateModal}
        onClose={() => setShowCreateModal(false)}
        onSubmit={handleCreateAgent}
      />
    </div>
  )
}

function AgentCard({ agent, onToggle, onDelete }: { agent: Agent; onToggle: () => void; onDelete: () => void }) {
  return (
    <div className="card hover:border-border-light transition-colors group">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <div className={`w-10 h-10 rounded-lg flex items-center justify-center ${
            agent.status === 'running' ? 'bg-accent/20 text-accent' :
            agent.status === 'error' ? 'bg-error/20 text-error' :
            'bg-surface-hover text-text-muted'
          }`}>
            <span className="text-lg font-bold">{agent.name.charAt(0).toUpperCase()}</span>
          </div>
          <div>
            <Link to={`/agents/${agent.id}`} className="font-medium text-text-primary hover:text-accent">
              {agent.name}
            </Link>
            <p className="text-sm text-text-muted">{agent.type} {agent.model && `• ${agent.model}`}</p>
          </div>
        </div>

        <div className="flex items-center gap-6">
          <div className="text-right hidden sm:block">
            <p className="text-sm text-text-secondary">{agent.tasksCompleted} tasks</p>
            <p className="text-xs text-text-muted">Last active: {agent.lastActive}</p>
          </div>

          <div className="flex items-center gap-2">
            <span className={`status-dot ${agent.status === 'running' ? 'online' : agent.status === 'error' ? 'error' : 'offline'}`} />
            <span className="text-sm capitalize">{agent.status}</span>
          </div>

          <div className="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
            <Link 
              to={`/agents/${agent.id}/chat`}
              className="p-2 hover:bg-surface-hover rounded-lg text-text-muted hover:text-accent"
              title="Chat with agent"
            >
              <MessageSquare className="w-4 h-4" />
            </Link>
            <button 
              onClick={onToggle}
              className="p-2 hover:bg-surface-hover rounded-lg text-text-muted hover:text-text-primary"
            >
              {agent.status === 'running' ? <Pause className="w-4 h-4" /> : <Play className="w-4 h-4" />}
            </button>
            <button className="p-2 hover:bg-surface-hover rounded-lg text-text-muted hover:text-text-primary">
              <MoreVertical className="w-4 h-4" />
            </button>
            <button 
              onClick={onDelete}
              className="p-2 hover:bg-surface-hover rounded-lg text-text-muted hover:text-error"
            >
              <Trash2 className="w-4 h-4" />
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
