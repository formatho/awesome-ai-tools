import { useParams, Link } from 'react-router-dom'
import { ArrowLeft, Play, Pause, Trash2, Settings, Activity, Clock, CheckCircle, XCircle, MessageSquare, Loader2 } from 'lucide-react'
import { useAgent, useAgentMutations } from '../../hooks/useAPI'

export default function AgentDetail() {
  const { id } = useParams()
  const { data: agent, isLoading, error } = useAgent(id || '')
  const mutations = useAgentMutations()

  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="w-8 h-8 animate-spin text-accent" />
      </div>
    )
  }

  if (error || !agent) {
    return (
      <div className="card text-center py-12">
        <p className="text-error">Failed to load agent. Please check if the backend is running.</p>
        <Link to="/agents" className="btn-secondary mt-4 inline-block">
          Back to Agents
        </Link>
      </div>
    )
  }

  const handleToggle = async () => {
    if (agent.status === 'running') {
      await mutations.stop.mutateAsync(agent.id)
    } else {
      await mutations.start.mutateAsync(agent.id)
    }
  }

  const handleDelete = async () => {
    if (confirm(`Are you sure you want to delete agent "${agent.name}"?`)) {
      await mutations.delete.mutateAsync(agent.id)
    }
  }

  // Calculate uptime from started_at if available
  const getUptime = () => {
    if (!agent.started_at) return 'Not started'
    const started = new Date(agent.started_at)
    const now = new Date()
    const diffMs = now.getTime() - started.getTime()
    const hours = Math.floor(diffMs / (1000 * 60 * 60))
    const mins = Math.floor((diffMs % (1000 * 60 * 60)) / (1000 * 60))
    return `${hours}h ${mins}m`
  }

  return (
    <div className="space-y-6 animate-fade-in">
      {/* Back button */}
      <Link to="/agents" className="inline-flex items-center gap-2 text-text-secondary hover:text-text-primary transition-colors">
        <ArrowLeft className="w-4 h-4" />
        Back to Agents
      </Link>

      {/* Agent Header */}
      <div className="card">
        <div className="flex items-start justify-between">
          <div className="flex items-center gap-4">
            <div className={`w-16 h-16 rounded-xl flex items-center justify-center ${agent.status === 'running' ? 'bg-success/20' : agent.status === 'error' ? 'bg-error/20' : 'bg-accent/20'}`}>
              <span className={`text-2xl font-bold ${agent.status === 'running' ? 'text-success' : agent.status === 'error' ? 'text-error' : 'text-accent'}`}>{agent.name.charAt(0).toUpperCase()}</span>
            </div>
            <div>
              <h1 className="text-2xl font-bold">{agent.name}</h1>
              <p className="text-text-secondary">{agent.provider || 'Agent'} {agent.model && `• ${agent.model}`}</p>
              <div className="flex items-center gap-2 mt-2">
                <span className={`status-dot ${agent.status === 'running' ? 'online' : agent.status === 'error' ? 'error' : 'offline'}`} />
                <span className="text-sm capitalize">{agent.status}</span>
              </div>
            </div>
          </div>

          <div className="flex items-center gap-2">
            <Link 
              to={`/agents/${agent.id}/chat`}
              className="btn-primary"
            >
              <MessageSquare className="w-4 h-4 mr-2" />
              Chat
            </Link>
            <button className="btn-secondary">
              <Settings className="w-4 h-4 mr-2" />
              Configure
            </button>
            {agent.status === 'running' ? (
              <button onClick={handleToggle} className="btn-secondary" disabled={mutations.stop.isPending}>
                {mutations.stop.isPending ? <Loader2 className="w-4 h-4 mr-2 animate-spin" /> : <Pause className="w-4 h-4 mr-2" />}
                Stop
              </button>
            ) : (
              <button onClick={handleToggle} className="btn-primary" disabled={mutations.start.isPending}>
                {mutations.start.isPending ? <Loader2 className="w-4 h-4 mr-2 animate-spin" /> : <Play className="w-4 h-4 mr-2" />}
                Start
              </button>
            )}
            <button onClick={handleDelete} className="btn-ghost text-error hover:bg-error/10" disabled={mutations.delete.isPending}>
              <Trash2 className="w-4 h-4" />
            </button>
          </div>
        </div>
        {agent.error && (
          <div className="mt-4 p-3 bg-error/10 border border-error/20 rounded-lg text-error text-sm">
            {agent.error}
          </div>
        )}
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard icon={Activity} label="Status" value={agent.status} color={agent.status === 'running' ? 'text-success' : agent.status === 'error' ? 'text-error' : 'text-text-muted'} />
        <StatCard icon={Clock} label="Uptime" value={getUptime()} color="text-accent" />
        <StatCard icon={CheckCircle} label="Provider" value={agent.provider || 'N/A'} color="text-success" />
        <StatCard icon={XCircle} label="Model" value={agent.model || 'N/A'} color="text-text-secondary" />
      </div>

      {/* Content Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Configuration */}
        <div className="card">
          <h2 className="text-lg font-semibold mb-4">Configuration</h2>
          <div className="space-y-3">
            <ConfigItem label="ID" value={agent.id} />
            <ConfigItem label="Provider" value={agent.provider || 'N/A'} />
            <ConfigItem label="Model" value={agent.model || 'N/A'} />
            <ConfigItem label="Created" value={new Date(agent.created_at).toLocaleString()} />
            <ConfigItem label="Updated" value={new Date(agent.updated_at).toLocaleString()} />
            {agent.started_at && <ConfigItem label="Started" value={new Date(agent.started_at).toLocaleString()} />}
            {agent.stopped_at && <ConfigItem label="Stopped" value={new Date(agent.stopped_at).toLocaleString()} />}
          </div>
        </div>

        {/* Status Info */}
        <div className="card">
          <h2 className="text-lg font-semibold mb-4">Status Information</h2>
          <div className="space-y-3">
            <div className="flex items-center justify-between py-2 border-b border-border">
              <span className="text-text-secondary">Current Status</span>
              <span className={`flex items-center gap-2`}>
                <span className={`status-dot ${agent.status === 'running' ? 'online' : agent.status === 'error' ? 'error' : 'offline'}`} />
                <span className="capitalize">{agent.status}</span>
              </span>
            </div>
            {agent.error && (
              <div className="p-3 bg-error/10 border border-error/20 rounded-lg text-error text-sm">
                <strong>Error:</strong> {agent.error}
              </div>
            )}
            {!agent.error && agent.status === 'idle' && (
              <p className="text-text-muted text-sm">Agent is idle and ready to start.</p>
            )}
            {agent.status === 'running' && (
              <p className="text-success text-sm">Agent is currently running and processing tasks.</p>
            )}
          </div>
        </div>
      </div>
    </div>
  )
}

function StatCard({ icon: Icon, label, value, color }: { icon: React.ElementType; label: string; value: string; color: string }) {
  return (
    <div className="card">
      <div className="flex items-center gap-3">
        <div className={`p-2 rounded-lg bg-surface-hover ${color}`}>
          <Icon className="w-5 h-5" />
        </div>
        <div>
          <p className="text-xl font-bold">{value}</p>
          <p className="text-sm text-text-muted">{label}</p>
        </div>
      </div>
    </div>
  )
}

function ConfigItem({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-center justify-between py-2 border-b border-border last:border-0">
      <span className="text-text-secondary">{label}</span>
      <span className="font-mono text-sm">{value}</span>
    </div>
  )
}
