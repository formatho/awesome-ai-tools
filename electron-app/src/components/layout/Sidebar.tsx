import { NavLink } from 'react-router-dom'
import { 
  LayoutDashboard, 
  Users, 
  CheckSquare, 
  Clock, 
  Settings,
  Activity
} from 'lucide-react'

const navItems = [
  { path: '/', icon: LayoutDashboard, label: 'Dashboard' },
  { path: '/agents', icon: Users, label: 'Agents' },
  { path: '/todos', icon: CheckSquare, label: 'TODOs' },
  { path: '/cron', icon: Clock, label: 'Cron Jobs' },
  { path: '/config', icon: Settings, label: 'Config' },
]

export default function Sidebar() {
  return (
    <aside className="w-64 bg-surface border-r border-border flex flex-col h-full">
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
            className={({ isActive }) =>
              `sidebar-item ${isActive ? 'active' : ''}`
            }
          >
            <item.icon className="w-5 h-5" />
            <span>{item.label}</span>
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
