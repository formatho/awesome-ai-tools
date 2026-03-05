import { useState } from 'react'
import { Search, Plus, Filter, CheckCircle, Circle, AlertTriangle, Calendar, User } from 'lucide-react'

interface TODO {
  id: string
  title: string
  description: string
  priority: 'low' | 'medium' | 'high'
  status: 'pending' | 'in-progress' | 'completed'
  assignee?: string
  dueDate?: string
  createdAt: string
}

const mockTODOs: TODO[] = [
  { id: '1', title: 'Update API documentation', description: 'Add new endpoints to docs', priority: 'high', status: 'in-progress', assignee: 'code-reviewer', dueDate: '2026-03-07', createdAt: '2026-03-05' },
  { id: '2', title: 'Fix authentication bug', description: 'Users unable to login with SSO', priority: 'high', status: 'pending', dueDate: '2026-03-06', createdAt: '2026-03-05' },
  { id: '3', title: 'Optimize database queries', description: 'Improve slow query performance', priority: 'medium', status: 'pending', createdAt: '2026-03-04' },
  { id: '4', title: 'Add unit tests for payment module', description: 'Increase test coverage to 80%', priority: 'medium', status: 'in-progress', assignee: 'test-runner', createdAt: '2026-03-03' },
  { id: '5', title: 'Review PR #234', description: 'Code review for new feature', priority: 'low', status: 'completed', assignee: 'code-reviewer', createdAt: '2026-03-02' },
  { id: '6', title: 'Update dependencies', description: 'Bump outdated packages', priority: 'low', status: 'pending', createdAt: '2026-03-01' },
]

export default function TODOList() {
  const [search, setSearch] = useState('')
  const [filterStatus, setFilterStatus] = useState<string>('all')
  const [filterPriority, setFilterPriority] = useState<string>('all')

  const filteredTODOs = mockTODOs.filter((todo) => {
    const matchesSearch = todo.title.toLowerCase().includes(search.toLowerCase())
    const matchesStatus = filterStatus === 'all' || todo.status === filterStatus
    const matchesPriority = filterPriority === 'all' || todo.priority === filterPriority
    return matchesSearch && matchesStatus && matchesPriority
  })

  return (
    <div className="space-y-6 animate-fade-in">
      {/* Page Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-text-primary">TODOs</h1>
          <p className="text-text-secondary mt-1">Track and manage your tasks</p>
        </div>
        <button className="btn-primary">
          <Plus className="w-4 h-4 mr-2" />
          New TODO
        </button>
      </div>

      {/* Search and Filters */}
      <div className="flex flex-wrap gap-4">
        <div className="relative flex-1 min-w-64">
          <Search className="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-text-muted" />
          <input
            type="text"
            placeholder="Search TODOs..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            className="input pl-10"
          />
        </div>
        <div className="flex items-center gap-2">
          <Filter className="w-4 h-4 text-text-muted" />
          <select
            value={filterStatus}
            onChange={(e) => setFilterStatus(e.target.value)}
            className="input w-auto"
          >
            <option value="all">All Status</option>
            <option value="pending">Pending</option>
            <option value="in-progress">In Progress</option>
            <option value="completed">Completed</option>
          </select>
          <select
            value={filterPriority}
            onChange={(e) => setFilterPriority(e.target.value)}
            className="input w-auto"
          >
            <option value="all">All Priority</option>
            <option value="high">High</option>
            <option value="medium">Medium</option>
            <option value="low">Low</option>
          </select>
        </div>
      </div>

      {/* TODO List */}
      <div className="space-y-3">
        {filteredTODOs.map((todo) => (
          <TODOCard key={todo.id} todo={todo} />
        ))}
      </div>

      {filteredTODOs.length === 0 && (
        <div className="card text-center py-12">
          <p className="text-text-muted">No TODOs found</p>
        </div>
      )}
    </div>
  )
}

function TODOCard({ todo }: { todo: TODO }) {
  const priorityColors = {
    high: 'bg-error/20 text-error',
    medium: 'bg-warning/20 text-warning',
    low: 'bg-accent/20 text-accent',
  }

  const statusIcons = {
    pending: Circle,
    'in-progress': AlertTriangle,
    completed: CheckCircle,
  }

  const StatusIcon = statusIcons[todo.status]

  return (
    <div className={`card group hover:border-border-light transition-colors ${todo.status === 'completed' ? 'opacity-60' : ''}`}>
      <div className="flex items-start gap-4">
        {/* Status checkbox */}
        <button className={`mt-1 ${todo.status === 'completed' ? 'text-success' : 'text-text-muted hover:text-accent'}`}>
          <StatusIcon className="w-5 h-5" />
        </button>

        {/* Content */}
        <div className="flex-1 min-w-0">
          <div className="flex items-start justify-between gap-4">
            <div>
              <h3 className={`font-medium ${todo.status === 'completed' ? 'line-through text-text-muted' : 'text-text-primary'}`}>
                {todo.title}
              </h3>
              <p className="text-sm text-text-muted mt-1">{todo.description}</p>
            </div>
            <span className={`badge ${priorityColors[todo.priority]}`}>
              {todo.priority}
            </span>
          </div>

          {/* Meta info */}
          <div className="flex flex-wrap items-center gap-4 mt-3 text-sm text-text-muted">
            {todo.assignee && (
              <div className="flex items-center gap-1">
                <User className="w-3.5 h-3.5" />
                <span>{todo.assignee}</span>
              </div>
            )}
            {todo.dueDate && (
              <div className="flex items-center gap-1">
                <Calendar className="w-3.5 h-3.5" />
                <span>{todo.dueDate}</span>
              </div>
            )}
            <div className="flex items-center gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
              <button className="text-text-muted hover:text-accent">Edit</button>
              <span>•</span>
              <button className="text-text-muted hover:text-error">Delete</button>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
