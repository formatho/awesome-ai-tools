import { useParams, Link } from 'react-router-dom'
import { ArrowLeft, Play, Pause, Trash2, Settings, Activity, Clock, CheckCircle, XCircle, MessageSquare } from 'lucide-react'

const mockAgentDetail = {
  id: '1',
  name: 'code-reviewer',
  type: 'Code Analysis',
  status: 'running',
  createdAt: '2026-01-15',
  lastActive: '2 min ago',
  tasksCompleted: 156,
  tasksFailed: 3,
  uptime: '12d 4h 32m',
  config: {
    model: 'gpt-4',
    temperature: 0.7,
    maxTokens: 2000,
  },
  recentTasks: [
    { id: 1, type: 'Review', status: 'completed', duration: '45s', time: '2 min ago' },
    { id: 2, type: 'Review', status: 'completed', duration: '32s', time: '8 min ago' },
    { id: 3, type: 'Review', status: 'failed', duration: '12s', time: '15 min ago' },
    { id: 4, type: 'Review', status: 'completed', duration: '58s', time: '22 min ago' },
    { id: 5, type: 'Review', status: 'completed', duration: '41s', time: '35 min ago' },
  ],
}

export default function AgentDetail() {
  const { id } = useParams()

  // In a real app, you would fetch agent data based on id
  // For now, using mock data but id is available for API calls
  console.debug('Viewing agent:', id)
  const agent = mockAgentDetail

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
            <div className="w-16 h-16 rounded-xl bg-accent/20 flex items-center justify-center">
              <span className="text-2xl font-bold text-accent">{agent.name.charAt(0).toUpperCase()}</span>
            </div>
            <div>
              <h1 className="text-2xl font-bold">{agent.name}</h1>
              <p className="text-text-secondary">{agent.type}</p>
              <div className="flex items-center gap-2 mt-2">
                <span className={`status-dot ${agent.status === 'running' ? 'online' : 'offline'}`} />
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
              <button className="btn-secondary">
                <Pause className="w-4 h-4 mr-2" />
                Stop
              </button>
            ) : (
              <button className="btn-primary">
                <Play className="w-4 h-4 mr-2" />
                Start
              </button>
            )}
            <button className="btn-ghost text-error hover:bg-error/10">
              <Trash2 className="w-4 h-4" />
            </button>
          </div>
        </div>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        <StatCard icon={Activity} label="Tasks Completed" value={agent.tasksCompleted.toString()} color="text-success" />
        <StatCard icon={XCircle} label="Tasks Failed" value={agent.tasksFailed.toString()} color="text-error" />
        <StatCard icon={Clock} label="Uptime" value={agent.uptime} color="text-accent" />
        <StatCard icon={CheckCircle} label="Success Rate" value="98.1%" color="text-success" />
      </div>

      {/* Content Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Configuration */}
        <div className="card">
          <h2 className="text-lg font-semibold mb-4">Configuration</h2>
          <div className="space-y-3">
            <ConfigItem label="Model" value={agent.config.model} />
            <ConfigItem label="Temperature" value={agent.config.temperature.toString()} />
            <ConfigItem label="Max Tokens" value={agent.config.maxTokens.toString()} />
            <ConfigItem label="Created" value={agent.createdAt} />
          </div>
        </div>

        {/* Recent Tasks */}
        <div className="card">
          <h2 className="text-lg font-semibold mb-4">Recent Tasks</h2>
          <div className="space-y-2">
            {agent.recentTasks.map((task) => (
              <div key={task.id} className="flex items-center justify-between p-2 rounded-lg hover:bg-surface-hover">
                <div className="flex items-center gap-3">
                  <span className={`status-dot ${task.status === 'completed' ? 'online' : 'error'}`} />
                  <span className="text-sm">{task.type}</span>
                </div>
                <div className="flex items-center gap-4 text-sm text-text-muted">
                  <span>{task.duration}</span>
                  <span>{task.time}</span>
                </div>
              </div>
            ))}
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
