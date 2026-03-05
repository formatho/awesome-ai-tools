import { useState } from 'react'
import { Search, Plus, Play, Pause, Clock, MoreVertical, Trash2, Edit } from 'lucide-react'

interface CronJob {
  id: string
  name: string
  schedule: string
  scheduleHuman: string
  command: string
  status: 'active' | 'paused' | 'error'
  lastRun?: string
  nextRun: string
  successCount: number
  failCount: number
}

const mockCronJobs: CronJob[] = [
  { id: '1', name: 'daily-report', schedule: '0 9 * * *', scheduleHuman: 'Every day at 9:00 AM', command: 'node scripts/report.js', status: 'active', lastRun: '2026-03-05 09:00', nextRun: '2026-03-06 09:00', successCount: 45, failCount: 0 },
  { id: '2', name: 'hourly-sync', schedule: '0 * * * *', scheduleHuman: 'Every hour', command: 'node scripts/sync.js', status: 'active', lastRun: '2026-03-05 17:00', nextRun: '2026-03-05 18:00', successCount: 120, failCount: 2 },
  { id: '3', name: 'weekly-cleanup', schedule: '0 0 * * 0', scheduleHuman: 'Every Sunday at midnight', command: 'node scripts/cleanup.js', status: 'paused', lastRun: '2026-03-02 00:00', nextRun: 'Paused', successCount: 12, failCount: 0 },
  { id: '4', name: 'backup-db', schedule: '0 2 * * *', scheduleHuman: 'Every day at 2:00 AM', command: 'pg_dump > backup.sql', status: 'active', lastRun: '2026-03-05 02:00', nextRun: '2026-03-06 02:00', successCount: 89, failCount: 1 },
  { id: '5', name: 'health-check', schedule: '*/5 * * * *', scheduleHuman: 'Every 5 minutes', command: 'curl /health', status: 'error', lastRun: '2026-03-05 17:05', nextRun: 'Error', successCount: 2000, failCount: 15 },
]

export default function CronList() {
  const [search, setSearch] = useState('')

  const filteredJobs = mockCronJobs.filter((job) =>
    job.name.toLowerCase().includes(search.toLowerCase())
  )

  return (
    <div className="space-y-6 animate-fade-in">
      {/* Page Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-text-primary">Cron Jobs</h1>
          <p className="text-text-secondary mt-1">Schedule and automate recurring tasks</p>
        </div>
        <button className="btn-primary">
          <Plus className="w-4 h-4 mr-2" />
          New Cron Job
        </button>
      </div>

      {/* Search */}
      <div className="relative max-w-md">
        <Search className="w-4 h-4 absolute left-3 top-1/2 -translate-y-1/2 text-text-muted" />
        <input
          type="text"
          placeholder="Search cron jobs..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="input pl-10"
        />
      </div>

      {/* Cron Jobs Table */}
      <div className="card overflow-hidden p-0">
        <table className="w-full">
          <thead>
            <tr className="border-b border-border">
              <th className="text-left p-4 text-sm font-medium text-text-secondary">Name</th>
              <th className="text-left p-4 text-sm font-medium text-text-secondary hidden md:table-cell">Schedule</th>
              <th className="text-left p-4 text-sm font-medium text-text-secondary hidden lg:table-cell">Command</th>
              <th className="text-left p-4 text-sm font-medium text-text-secondary">Status</th>
              <th className="text-left p-4 text-sm font-medium text-text-secondary hidden sm:table-cell">Next Run</th>
              <th className="text-right p-4 text-sm font-medium text-text-secondary">Actions</th>
            </tr>
          </thead>
          <tbody>
            {filteredJobs.map((job) => (
              <tr key={job.id} className="border-b border-border last:border-0 hover:bg-surface-hover transition-colors">
                <td className="p-4">
                  <div>
                    <p className="font-medium text-text-primary">{job.name}</p>
                    <p className="text-xs text-text-muted md:hidden">{job.scheduleHuman}</p>
                  </div>
                </td>
                <td className="p-4 hidden md:table-cell">
                  <div className="flex items-center gap-2">
                    <Clock className="w-4 h-4 text-text-muted" />
                    <div>
                      <code className="text-sm text-accent">{job.schedule}</code>
                      <p className="text-xs text-text-muted">{job.scheduleHuman}</p>
                    </div>
                  </div>
                </td>
                <td className="p-4 hidden lg:table-cell">
                  <code className="text-sm text-text-secondary">{job.command}</code>
                </td>
                <td className="p-4">
                  <div className="flex items-center gap-2">
                    <span className={`status-dot ${
                      job.status === 'active' ? 'online' :
                      job.status === 'error' ? 'error' :
                      'offline'
                    }`} />
                    <span className="text-sm capitalize">{job.status}</span>
                  </div>
                </td>
                <td className="p-4 hidden sm:table-cell">
                  <span className="text-sm text-text-secondary">{job.nextRun}</span>
                </td>
                <td className="p-4">
                  <div className="flex items-center justify-end gap-1">
                    <button className="p-2 hover:bg-surface rounded-lg text-text-muted hover:text-text-primary" title={job.status === 'active' ? 'Pause' : 'Resume'}>
                      {job.status === 'active' ? <Pause className="w-4 h-4" /> : <Play className="w-4 h-4" />}
                    </button>
                    <button className="p-2 hover:bg-surface rounded-lg text-text-muted hover:text-text-primary" title="Edit">
                      <Edit className="w-4 h-4" />
                    </button>
                    <button className="p-2 hover:bg-surface rounded-lg text-text-muted hover:text-error" title="Delete">
                      <Trash2 className="w-4 h-4" />
                    </button>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>

        {filteredJobs.length === 0 && (
          <div className="text-center py-12">
            <p className="text-text-muted">No cron jobs found</p>
          </div>
        )}
      </div>

      {/* Stats Summary */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div className="card">
          <p className="text-sm text-text-secondary">Total Jobs</p>
          <p className="text-2xl font-bold mt-1">{mockCronJobs.length}</p>
        </div>
        <div className="card">
          <p className="text-sm text-text-secondary">Active Jobs</p>
          <p className="text-2xl font-bold mt-1 text-success">{mockCronJobs.filter(j => j.status === 'active').length}</p>
        </div>
        <div className="card">
          <p className="text-sm text-text-secondary">Total Executions</p>
          <p className="text-2xl font-bold mt-1">
            {mockCronJobs.reduce((acc, j) => acc + j.successCount + j.failCount, 0).toLocaleString()}
          </p>
        </div>
      </div>
    </div>
  )
}
