import { NavLink } from 'react-router-dom'
import { 
  LayoutDashboard, 
  Users, 
  CheckSquare, 
  Clock, 
  Settings,
  Activity,
  Zap,
  Sparkles
} from 'lucide-react'

const navItems = [
  { path: '/', icon: LayoutDashboard, label: 'Dashboard', tourId: 'tour-dashboard-nav' },
  { path: '/tutorial', icon: Sparkles, label: 'Tutorial', highlight: true, tourId: 'tour-tutorial-nav' },
  { path: '/agents', icon: Users, label: 'Agents', tourId: 'tour-agents-nav' },
  { path: '/todos', icon: CheckSquare, label: 'TODOs', tourId: 'tour-todos-nav' },
  { path: '/cron', icon: Clock, label: 'Cron Jobs', tourId: 'tour-cron-nav' },
  { path: '/skills', icon: Zap, label: 'Skills', tourId: 'tour-skills-nav' },
  { path: '/config', icon: Settings, label: 'Config', tourId: 'tour-config-nav' },
]

export default function Sidebar() {
  return (
    <aside id="tour-sidebar" className="w-64 bg-surface border-r border-border flex flex-col h-full">
      {/* Logo */}
      <div className="p-4 border-b border-border">
        <div className="flex items-center gap-2">
          <div className="w-8 h-8 rounded-lg bg-accent flex items-center justify-center">
            <Activity className="w-5 h-5 text-white" />
          </div>
          <span className="font-semibold text-lg text-text-primary">Agent Orchestrator</span>
        </div>
      </div>

      {/* Navigation */}
      <nav className="flex-1 p-3 space-y-1">
        {navItems.map((item) => (
          <NavLink
            key={item.path}
            to={item.path}
            id={item.tourId}
            className={({ isActive }) =>
              `sidebar-item ${isActive ? 'active' : ''} ${item.highlight ? 'text-accent' : ''}`
            }
          >
            <item.icon className={`w-5 h-5 ${item.highlight ? 'text-accent' : ''}`} />
            <span>{item.label}</span>
            {item.highlight && (
              <span className="ml-auto text-xs bg-accent/20 text-accent px-1.5 py-0.5 rounded">NEW</span>
            )}
          </NavLink>
        ))}
      </nav>

      {/* Footer */}
      <div className="p-4 border-t border-border">
        <div className="text-xs text-text-muted">
          <p>Version 0.1.0</p>
          <p className="mt-1">Formatho © 2026</p>
        </div>
      </div>
    </aside>
  )
}
