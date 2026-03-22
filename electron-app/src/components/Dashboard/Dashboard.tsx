import { Users, CheckSquare, Clock, Activity, TrendingUp, AlertCircle, PlayCircle } from 'lucide-react'
import { useAgents, useTODOs } from '../../hooks/useAPI'
import { startTour, isTourCompleted, shouldShowTour } from '../../lib/productTour'
import { useState, useEffect } from 'react'

export default function Dashboard() {
  const { data: agents } = useAgents()
  const { data: todos } = useTODOs({ status: 'pending' })
  const [showTourPrompt, setShowTourPrompt] = useState(false)

  // Check if we should show the tour prompt
  useEffect(() => {
    setShowTourPrompt(shouldShowTour())
  }, [])

  // Count active agents (status === 'running')
  const activeAgentsCount = agents?.filter((agent: { status: string }) => agent.status === 'running').length ?? 0

  // Count pending TODOs
  const pendingTodosCount = todos?.length ?? 0

  const stats = [
    { label: 'Active Agents', value: activeAgentsCount.toString(), icon: Users, color: 'text-accent', change: '+2' },
    { label: 'Pending TODOs', value: pendingTodosCount.toString(), icon: CheckSquare, color: 'text-warning', change: '-5' },
    { label: 'Cron Jobs', value: '8', icon: Clock, color: 'text-success', change: '0' },
    { label: 'Tasks Today', value: '156', icon: Activity, color: 'text-accent', change: '+12' },
  ]

  const recentActivity = [
    { id: 1, type: 'agent', message: 'Agent "code-reviewer" completed task', time: '2 min ago' },
    { id: 2, type: 'todo', message: 'TODO "Update docs" marked as done', time: '15 min ago' },
    { id: 3, type: 'cron', message: 'Cron "daily-report" executed successfully', time: '1 hour ago' },
    { id: 4, type: 'agent', message: 'Agent "data-scraper" started', time: '2 hours ago' },
  ]

  return (
    <div className="space-y-6 animate-fade-in">
      {/* Page Header */}
      <div id="tour-welcome">
        <h1 className="text-2xl font-bold text-text-primary">Dashboard</h1>
        <p className="text-text-secondary mt-1">Overview of your agents, tasks, and system status</p>
      </div>

      {/* Tour Prompt Banner */}
      {showTourPrompt && (
        <div className="card bg-gradient-to-r from-accent/10 to-success/10 border-accent/20 flex items-center justify-between">
          <div className="flex items-center gap-4">
            <div className="w-10 h-10 rounded-lg bg-accent/20 flex items-center justify-center">
              <PlayCircle className="w-5 h-5 text-accent" />
            </div>
            <div>
              <h3 className="font-semibold text-text-primary">New here? Take a quick tour!</h3>
              <p className="text-sm text-text-secondary">Learn how to create agents, assign tasks, and monitor progress in just 2 minutes.</p>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <button
              onClick={() => setShowTourPrompt(false)}
              className="btn-ghost text-sm text-text-secondary hover:text-text-primary"
            >
              Maybe Later
            </button>
            <button
              onClick={() => {
                setShowTourPrompt(false)
                startTour()
              }}
              className="btn-primary text-sm flex items-center gap-2"
            >
              <PlayCircle className="w-4 h-4" />
              Start Tour
            </button>
          </div>
        </div>
      )}

      {/* Stats Grid */}
      <div id="tour-dashboard-stats" className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {stats.map((stat) => (
          <div key={stat.label} className="card">
            <div className="flex items-center justify-between">
              <div className={`p-2 rounded-lg bg-surface-hover ${stat.color}`}>
                <stat.icon className="w-5 h-5" />
              </div>
              <span className={`text-sm ${stat.change.startsWith('+') ? 'text-success' : stat.change.startsWith('-') ? 'text-error' : 'text-text-muted'}`}>
                {stat.change}
              </span>
            </div>
            <div className="mt-3">
              <p className="text-2xl font-bold">{stat.value}</p>
              <p className="text-sm text-text-secondary">{stat.label}</p>
            </div>
          </div>
        ))}
      </div>

      {/* Content Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Recent Activity */}
        <div id="tour-activity-feed" className="card">
          <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
            <TrendingUp className="w-5 h-5 text-accent" />
            Recent Activity
          </h2>
          <div className="space-y-3">
            {recentActivity.map((activity) => (
              <div key={activity.id} className="flex items-start gap-3 p-2 rounded-lg hover:bg-surface-hover transition-colors">
                <div className={`w-2 h-2 rounded-full mt-2 ${
                  activity.type === 'agent' ? 'bg-accent' : 
                  activity.type === 'todo' ? 'bg-warning' : 
                  'bg-success'
                }`} />
                <div className="flex-1 min-w-0">
                  <p className="text-sm text-text-primary truncate">{activity.message}</p>
                  <p className="text-xs text-text-muted">{activity.time}</p>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* System Status */}
        <div className="card">
          <h2 className="text-lg font-semibold mb-4 flex items-center gap-2">
            <AlertCircle className="w-5 h-5 text-accent" />
            System Status
          </h2>
          <div className="space-y-4">
            <StatusItem label="API Server" status="online" />
            <StatusItem label="WebSocket" status="online" />
            <StatusItem label="Database" status="online" />
            <StatusItem label="Queue Worker" status="busy" />
          </div>
        </div>
      </div>

      {/* Quick Actions */}
      <div id="tour-quick-actions" className="card">
        <h2 className="text-lg font-semibold mb-4">Quick Actions</h2>
        <div className="flex flex-wrap gap-3">
          <button className="btn-primary">
            <Users className="w-4 h-4 mr-2 inline" />
            New Agent
          </button>
          <button className="btn-secondary">
            <CheckSquare className="w-4 h-4 mr-2 inline" />
            Add TODO
          </button>
          <button className="btn-secondary">
            <Clock className="w-4 h-4 mr-2 inline" />
            Schedule Cron
          </button>
          {!isTourCompleted() && (
            <button 
              onClick={startTour}
              className="btn-secondary flex items-center gap-2"
            >
              <PlayCircle className="w-4 h-4" />
              Start Tour
            </button>
          )}
        </div>
      </div>
    </div>
  )
}

function StatusItem({ label, status }: { label: string; status: 'online' | 'offline' | 'busy' | 'error' }) {
  const statusText = {
    online: 'Online',
    offline: 'Offline',
    busy: 'Busy',
    error: 'Error',
  }

  return (
    <div className="flex items-center justify-between">
      <span className="text-text-secondary">{label}</span>
      <div className="flex items-center gap-2">
        <span className={`status-dot ${status}`} />
        <span className="text-sm">{statusText[status]}</span>
      </div>
    </div>
  )
}
