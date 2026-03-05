import { useState } from 'react'
import { Link } from 'react-router-dom'
import { Search, Plus, Filter, MoreVertical, Play, Pause, Trash2 } from 'lucide-react'

interface Agent {
  id: string
  name: string
  type: string
  status: 'running' | 'idle' | 'error'
  lastActive: string
  tasksCompleted: number
}

const mockAgents: Agent[] = [
  { id: '1', name: 'code-reviewer', type: 'Code Analysis', status: 'running', lastActive: '2 min ago', tasksCompleted: 156 },
  { id: '2', name: 'data-scraper', type: 'Web Scraping', status: 'idle', lastActive: '1 hour ago', tasksCompleted: 89 },
  { id: '3', name: 'doc-writer', type: 'Documentation', status: 'running', lastActive: '5 min ago', tasksCompleted: 234 },
  { id: '4', name: 'test-runner', type: 'Testing', status: 'error', lastActive: '3 hours ago', tasksCompleted: 67 },
  { id: '5', name: 'api-monitor', type: 'Monitoring', status: 'running', lastActive: '1 min ago', tasksCompleted: 1012 },
]

export default function AgentList() {
  const [search, setSearch] = useState('')
  const [filter, setFilter] = useState<string>('all')

  const filteredAgents = mockAgents.filter((agent) => {
    const matchesSearch = agent.name.toLowerCase().includes(search.toLowerCase())
    const matchesFilter = filter === 'all' || agent.status === filter
    return matchesSearch && matchesFilter
  })

  return (
    <div className="space-y-6 animate-fade-in">
      {/* Page Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-text-primary">Agents</h1>
          <p className="text-text-secondary mt-1">Manage and monitor your AI agents</p>
        </div>
        <button className="btn-primary">
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

      {/* Agent List */}
      <div className="space-y-3">
        {filteredAgents.map((agent) => (
          <AgentCard key={agent.id} agent={agent} />
        ))}
      </div>

      {filteredAgents.length === 0 && (
        <div className="card text-center py-12">
          <p className="text-text-muted">No agents found</p>
        </div>
      )}
    </div>
  )
}

function AgentCard({ agent }: { agent: Agent }) {
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
            <p className="text-sm text-text-muted">{agent.type}</p>
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
            <button className="p-2 hover:bg-surface-hover rounded-lg text-text-muted hover:text-text-primary">
              {agent.status === 'running' ? <Pause className="w-4 h-4" /> : <Play className="w-4 h-4" />}
            </button>
            <button className="p-2 hover:bg-surface-hover rounded-lg text-text-muted hover:text-text-primary">
              <MoreVertical className="w-4 h-4" />
            </button>
          </div>
        </div>
      </div>
    </div>
  )
}
