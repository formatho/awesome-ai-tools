import { useState, useMemo } from 'react'
import {
  Search,
  Filter,
  SortAsc,
  SortDesc,
  Plus,
  CheckCircle,
  Circle,
  Clock,
  AlertTriangle,
  Calendar,
  User,
  Trash2,
  Edit,
  ArrowRight,
  CheckSquare,
  LayoutList,
  LayoutGrid,
  RefreshCw,
} from 'lucide-react'
import { useTODOs, useTODOMutations } from '../../hooks/useAPI'

interface TODO {
  id: string
  title: string
  description: string
  priority: number
  status: 'pending' | 'in-progress' | 'completed'
  assignee?: string
  dueDate?: string
  createdAt: string
}

type ViewMode = 'list' | 'kanban' | 'compact'
type SortField = 'priority' | 'createdAt' | 'dueDate' | 'assignee' | 'status'
type SortDirection = 'asc' | 'desc'

export default function TODOQueue() {
  const [viewMode, setViewMode] = useState<ViewMode>('list')
  const [searchQuery, setSearchQuery] = useState('')
  const [statusFilter, setStatusFilter] = useState<string>('all')
  const [priorityFilter, setPriorityFilter] = useState<string>('all')
  const [sortField, setSortField] = useState<SortField>('priority')
  const [sortDirection, setSortDirection] = useState<SortDirection>('desc')
  const [selectedItems, setSelectedItems] = useState<Set<string>>(new Set())
  const [showFilters, setShowFilters] = useState(false)

  // Fetch TODOs with filters
  const { data: todos, isLoading, error, refetch } = useTODOs({
    status: statusFilter !== 'all' ? statusFilter : undefined,
    priority: priorityFilter !== 'all' ? priorityFilter : undefined,
  })

  const { update, delete: deleteTodo } = useTODOMutations()

  // Filter and sort TODOs
  const filteredAndSortedTodos = useMemo(() => {
    if (!todos) return []

    let result = [...todos]

    // Search filter
    if (searchQuery) {
      const query = searchQuery.toLowerCase()
      result = result.filter(
        (todo: TODO) =>
          todo.title.toLowerCase().includes(query) ||
          todo.description.toLowerCase().includes(query) ||
          todo.assignee?.toLowerCase().includes(query)
      )
    }

    // Sorting
    result.sort((a: TODO, b: TODO) => {
      let comparison = 0

      switch (sortField) {
        case 'priority':
          comparison = a.priority - b.priority
          break
        case 'createdAt':
          comparison = new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime()
          break
        case 'dueDate':
          if (!a.dueDate) return 1
          if (!b.dueDate) return -1
          comparison = new Date(a.dueDate).getTime() - new Date(b.dueDate).getTime()
          break
        case 'assignee':
          comparison = (a.assignee || '').localeCompare(b.assignee || '')
          break
        case 'status':
          const statusOrder = { pending: 0, 'in-progress': 1, completed: 2 }
          comparison = statusOrder[a.status] - statusOrder[b.status]
          break
      }

      return sortDirection === 'asc' ? comparison : -comparison
    })

    return result
  }, [todos, searchQuery, sortField, sortDirection])

  // Kanban columns
  const kanbanColumns = {
    pending: filteredAndSortedTodos.filter((t: TODO) => t.status === 'pending'),
    'in-progress': filteredAndSortedTodos.filter((t: TODO) => t.status === 'in-progress'),
    completed: filteredAndSortedTodos.filter((t: TODO) => t.status === 'completed'),
  }

  const handleSelectItem = (id: string) => {
    const newSelected = new Set(selectedItems)
    if (newSelected.has(id)) {
      newSelected.delete(id)
    } else {
      newSelected.add(id)
    }
    setSelectedItems(newSelected)
  }

  const handleSelectAll = () => {
    if (selectedItems.size === filteredAndSortedTodos.length) {
      setSelectedItems(new Set())
    } else {
      setSelectedItems(new Set(filteredAndSortedTodos.map((t: TODO) => t.id)))
    }
  }

  const handleBulkStatusUpdate = async (status: TODO['status']) => {
    const updates = Array.from(selectedItems).map((id) => update.mutateAsync({ id, data: { status } }))
    await Promise.all(updates)
    setSelectedItems(new Set())
  }

  const handleBulkDelete = async () => {
    if (!confirm(`Delete ${selectedItems.size} items?`)) return
    const deletions = Array.from(selectedItems).map((id) => deleteTodo.mutateAsync(id))
    await Promise.all(deletions)
    setSelectedItems(new Set())
  }

  const handleStatusChange = async (id: string, newStatus: TODO['status']) => {
    await update.mutateAsync({ id, data: { status: newStatus } })
  }

  const getPriorityColor = (priority: number) => {
    if (priority >= 8) return 'text-error bg-error/10 border-error/20'
    if (priority >= 5) return 'text-warning bg-warning/10 border-warning/20'
    return 'text-success bg-success/10 border-success/20'
  }

  const getStatusIcon = (status: TODO['status']) => {
    switch (status) {
      case 'completed':
        return <CheckCircle className="w-4 h-4 text-success" />
      case 'in-progress':
        return <Clock className="w-4 h-4 text-warning" />
      default:
        return <Circle className="w-4 h-4 text-text-muted" />
    }
  }

  const formatDate = (dateString: string) => {
    const date = new Date(dateString)
    const now = new Date()
    const diffMs = now.getTime() - date.getTime()
    const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24))

    if (diffDays === 0) return 'Today'
    if (diffDays === 1) return 'Yesterday'
    if (diffDays < 7) return `${diffDays} days ago`
    return date.toLocaleDateString()
  }

  // Loading state
  if (isLoading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-accent"></div>
      </div>
    )
  }

  // Error state
  if (error) {
    return (
      <div className="card bg-error/10 border-error/20">
        <div className="flex items-center gap-3">
          <AlertTriangle className="w-5 h-5 text-error" />
          <div>
            <h3 className="font-semibold text-error">Failed to load TODOs</h3>
            <p className="text-sm text-text-muted mt-1">
              {error instanceof Error ? error.message : 'Unknown error'}
            </p>
          </div>
          <button onClick={() => refetch()} className="btn-secondary ml-auto">
            <RefreshCw className="w-4 h-4 mr-2" />
            Retry
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6 animate-fade-in">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-text-primary">TODO Queue</h1>
          <p className="text-text-secondary mt-1">
            {filteredAndSortedTodos.length} items • {selectedItems.size} selected
          </p>
        </div>

        <div className="flex items-center gap-3">
          {/* View Mode Toggle */}
          <div className="flex items-center bg-surface-hover rounded-lg p-1">
            <button
              onClick={() => setViewMode('list')}
              className={`p-2 rounded ${viewMode === 'list' ? 'bg-surface text-accent' : 'text-text-muted'}`}
              title="List View"
            >
              <LayoutList className="w-4 h-4" />
            </button>
            <button
              onClick={() => setViewMode('kanban')}
              className={`p-2 rounded ${viewMode === 'kanban' ? 'bg-surface text-accent' : 'text-text-muted'}`}
              title="Kanban Board"
            >
              <LayoutGrid className="w-4 h-4" />
            </button>
          </div>

          {/* Filters Toggle */}
          <button
            onClick={() => setShowFilters(!showFilters)}
            className={`btn-secondary ${showFilters ? 'bg-accent/10 text-accent' : ''}`}
          >
            <Filter className="w-4 h-4 mr-2" />
            Filters
          </button>

          {/* Refresh */}
          <button onClick={() => refetch()} className="btn-secondary">
            <RefreshCw className="w-4 h-4" />
          </button>

          {/* Create New */}
          <button className="btn-primary">
            <Plus className="w-4 h-4 mr-2" />
            New TODO
          </button>
        </div>
      </div>

      {/* Filters Bar */}
      {showFilters && (
        <div className="card space-y-4 animate-slide-down">
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
            {/* Search */}
            <div className="relative">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-text-muted" />
              <input
                type="text"
                placeholder="Search TODOs..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="input w-full pl-10"
              />
            </div>

            {/* Status Filter */}
            <select
              value={statusFilter}
              onChange={(e) => setStatusFilter(e.target.value)}
              className="input"
            >
              <option value="all">All Statuses</option>
              <option value="pending">Pending</option>
              <option value="in-progress">In Progress</option>
              <option value="completed">Completed</option>
            </select>

            {/* Priority Filter */}
            <select
              value={priorityFilter}
              onChange={(e) => setPriorityFilter(e.target.value)}
              className="input"
            >
              <option value="all">All Priorities</option>
              <option value="high">High (8-10)</option>
              <option value="medium">Medium (5-7)</option>
              <option value="low">Low (1-4)</option>
            </select>

            {/* Sort */}
            <div className="flex gap-2">
              <select
                value={sortField}
                onChange={(e) => setSortField(e.target.value as SortField)}
                className="input flex-1"
              >
                <option value="priority">Priority</option>
                <option value="createdAt">Created Date</option>
                <option value="dueDate">Due Date</option>
                <option value="assignee">Assignee</option>
                <option value="status">Status</option>
              </select>
              <button
                onClick={() => setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc')}
                className="btn-secondary px-3"
                title={sortDirection === 'asc' ? 'Ascending' : 'Descending'}
              >
                {sortDirection === 'asc' ? <SortAsc className="w-4 h-4" /> : <SortDesc className="w-4 h-4" />}
              </button>
            </div>
          </div>

          {/* Bulk Actions */}
          {selectedItems.size > 0 && (
            <div className="flex items-center gap-2 pt-4 border-t border-border">
              <span className="text-sm text-text-muted">{selectedItems.size} selected</span>
              <button onClick={handleSelectAll} className="btn-ghost text-sm">
                {selectedItems.size === filteredAndSortedTodos.length ? 'Deselect All' : 'Select All'}
              </button>
              <div className="flex-1" />
              <button
                onClick={() => handleBulkStatusUpdate('in-progress')}
                className="btn-secondary text-sm"
              >
                <Clock className="w-3 h-3 mr-1" />
                Start
              </button>
              <button
                onClick={() => handleBulkStatusUpdate('completed')}
                className="btn-secondary text-sm"
              >
                <CheckCircle className="w-3 h-3 mr-1" />
                Complete
              </button>
              <button onClick={handleBulkDelete} className="btn-danger text-sm">
                <Trash2 className="w-3 h-3 mr-1" />
                Delete
              </button>
            </div>
          )}
        </div>
      )}

      {/* Main Content */}
      {viewMode === 'kanban' ? (
        // Kanban Board View
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {Object.entries(kanbanColumns).map(([status, items]) => (
            <div key={status} className="space-y-4">
              {/* Column Header */}
              <div className="flex items-center justify-between">
                <h3 className="font-semibold text-text-primary capitalize">{status.replace('-', ' ')}</h3>
                <span className="text-sm text-text-muted bg-surface-hover px-2 py-1 rounded">
                  {items.length}
                </span>
              </div>

              {/* Column Cards */}
              <div className="space-y-3">
                {items.map((todo: TODO) => (
                  <div key={todo.id} className="card hover:border-accent/50 transition-all cursor-pointer">
                    <div className="space-y-3">
                      {/* Priority Badge */}
                      <div className="flex items-center justify-between">
                        <span className={`text-xs px-2 py-1 rounded border ${getPriorityColor(todo.priority)}`}>
                          P{todo.priority}
                        </span>
                        {getStatusIcon(todo.status)}
                      </div>

                      {/* Title */}
                      <h4 className="font-medium text-text-primary">{todo.title}</h4>

                      {/* Description */}
                      {todo.description && (
                        <p className="text-sm text-text-secondary line-clamp-2">{todo.description}</p>
                      )}

                      {/* Meta */}
                      <div className="flex items-center gap-3 text-xs text-text-muted">
                        {todo.assignee && (
                          <div className="flex items-center gap-1">
                            <User className="w-3 h-3" />
                            {todo.assignee}
                          </div>
                        )}
                        {todo.dueDate && (
                          <div className="flex items-center gap-1">
                            <Calendar className="w-3 h-3" />
                            {formatDate(todo.dueDate)}
                          </div>
                        )}
                      </div>

                      {/* Quick Actions */}
                      <div className="flex items-center gap-2 pt-2 border-t border-border">
                        {status === 'pending' && (
                          <button
                            onClick={() => handleStatusChange(todo.id, 'in-progress')}
                            className="btn-ghost text-xs flex-1"
                          >
                            <ArrowRight className="w-3 h-3 mr-1" />
                            Start
                          </button>
                        )}
                        {status === 'in-progress' && (
                          <button
                            onClick={() => handleStatusChange(todo.id, 'completed')}
                            className="btn-ghost text-xs flex-1"
                          >
                            <CheckCircle className="w-3 h-3 mr-1" />
                            Complete
                          </button>
                        )}
                        <button className="btn-ghost text-xs">
                          <Edit className="w-3 h-3" />
                        </button>
                        <button
                          onClick={() => deleteTodo.mutate(todo.id)}
                          className="btn-ghost text-xs text-error"
                        >
                          <Trash2 className="w-3 h-3" />
                        </button>
                      </div>
                    </div>
                  </div>
                ))}

                {/* Empty State */}
                {items.length === 0 && (
                  <div className="card bg-surface-hover border-dashed text-center py-8">
                    <p className="text-sm text-text-muted">No items</p>
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>
      ) : (
        // List View
        <div className="card overflow-hidden">
          <table className="w-full">
            <thead className="bg-surface-hover border-b border-border">
              <tr>
                <th className="p-4 text-left">
                  <button onClick={handleSelectAll} className="hover:text-accent">
                    {selectedItems.size === filteredAndSortedTodos.length ? (
                      <CheckSquare className="w-4 h-4" />
                    ) : (
                      <Circle className="w-4 h-4" />
                    )}
                  </button>
                </th>
                <th className="p-4 text-left text-sm font-medium text-text-secondary">Priority</th>
                <th className="p-4 text-left text-sm font-medium text-text-secondary">Title</th>
                <th className="p-4 text-left text-sm font-medium text-text-secondary">Status</th>
                <th className="p-4 text-left text-sm font-medium text-text-secondary">Assignee</th>
                <th className="p-4 text-left text-sm font-medium text-text-secondary">Due Date</th>
                <th className="p-4 text-left text-sm font-medium text-text-secondary">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {filteredAndSortedTodos.map((todo: TODO) => (
                <tr
                  key={todo.id}
                  className="hover:bg-surface-hover transition-colors"
                >
                  <td className="p-4">
                    <button
                      onClick={() => handleSelectItem(todo.id)}
                      className="hover:text-accent"
                    >
                      {selectedItems.has(todo.id) ? (
                        <CheckSquare className="w-4 h-4 text-accent" />
                      ) : (
                        <Circle className="w-4 h-4" />
                      )}
                    </button>
                  </td>
                  <td className="p-4">
                    <span className={`text-xs px-2 py-1 rounded border ${getPriorityColor(todo.priority)}`}>
                      P{todo.priority}
                    </span>
                  </td>
                  <td className="p-4">
                    <div>
                      <div className="font-medium text-text-primary">{todo.title}</div>
                      {todo.description && (
                        <div className="text-sm text-text-secondary mt-1 line-clamp-1">
                          {todo.description}
                        </div>
                      )}
                    </div>
                  </td>
                  <td className="p-4">
                    <div className="flex items-center gap-2">
                      {getStatusIcon(todo.status)}
                      <span className="text-sm capitalize">{todo.status.replace('-', ' ')}</span>
                    </div>
                  </td>
                  <td className="p-4">
                    {todo.assignee ? (
                      <div className="flex items-center gap-2">
                        <div className="w-6 h-6 rounded-full bg-accent/20 flex items-center justify-center text-xs">
                          {todo.assignee.charAt(0).toUpperCase()}
                        </div>
                        <span className="text-sm">{todo.assignee}</span>
                      </div>
                    ) : (
                      <span className="text-sm text-text-muted">Unassigned</span>
                    )}
                  </td>
                  <td className="p-4">
                    {todo.dueDate ? (
                      <div className="flex items-center gap-1 text-sm">
                        <Calendar className="w-3 h-3 text-text-muted" />
                        {formatDate(todo.dueDate)}
                      </div>
                    ) : (
                      <span className="text-sm text-text-muted">—</span>
                    )}
                  </td>
                  <td className="p-4">
                    <div className="flex items-center gap-1">
                      <button className="btn-ghost p-1">
                        <Edit className="w-4 h-4" />
                      </button>
                      <button
                        onClick={() => deleteTodo.mutate(todo.id)}
                        className="btn-ghost p-1 text-error"
                      >
                        <Trash2 className="w-4 h-4" />
                      </button>
                    </div>
                  </td>
                </tr>
              ))}

              {/* Empty State */}
              {filteredAndSortedTodos.length === 0 && (
                <tr>
                  <td colSpan={7} className="p-12 text-center">
                    <div className="flex flex-col items-center">
                      <CheckCircle className="w-12 h-12 text-text-muted mb-3" />
                      <p className="text-text-secondary font-medium">No TODOs found</p>
                      <p className="text-sm text-text-muted mt-1">
                        {searchQuery || statusFilter !== 'all' || priorityFilter !== 'all'
                          ? 'Try adjusting your filters'
                          : 'Create your first TODO to get started'}
                      </p>
                    </div>
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}
