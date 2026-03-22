import { driver } from 'driver.js'
import 'driver.js/dist/driver.css'
import type { DriveStep } from 'driver.js'

const TOUR_COMPLETED_KEY = 'agent-orchestrator-tour-completed'
const TOUR_SKIPPED_KEY = 'agent-orchestrator-tour-skipped'

// Tour steps configuration
export const tourSteps: DriveStep[] = [
  {
    element: '#tour-welcome',
    popover: {
      title: '👋 Welcome to Agent Orchestrator!',
      description: 'Your command center for managing AI agents, tasks, and automation. Let us show you around!',
      side: 'bottom',
      align: 'start',
    },
  },
  {
    element: '#tour-sidebar',
    popover: {
      title: '📍 Navigation Sidebar',
      description: 'Navigate between Dashboard, Agents, TODOs, Cron Jobs, Skills, and Settings. Everything you need is just one click away.',
      side: 'right',
      align: 'start',
    },
  },
  {
    element: '#tour-dashboard-stats',
    popover: {
      title: '📊 Dashboard Overview',
      description: 'See real-time stats for active agents, pending tasks, cron jobs, and daily activity at a glance.',
      side: 'bottom',
      align: 'start',
    },
  },
  {
    element: '#tour-agents-nav',
    popover: {
      title: '🤖 Create Your First Agent',
      description: 'Agents are AI assistants that execute tasks autonomously. Click here to create and manage your agents.',
      side: 'right',
      align: 'start',
    },
  },
  {
    element: '#tour-todos-nav',
    popover: {
      title: '✅ Assign Tasks',
      description: 'Create and assign tasks to your agents. Define priorities, descriptions, and track progress.',
      side: 'right',
      align: 'start',
    },
  },
  {
    element: '#tour-activity-feed',
    popover: {
      title: '📈 Monitor Progress',
      description: 'Watch your agents work in real-time. View activity logs, task status, and system health.',
      side: 'left',
      align: 'start',
    },
  },
  {
    element: '#tour-config-nav',
    popover: {
      title: '⚙️ Explore Settings',
      description: 'Configure your API keys, providers, and system preferences to customize your experience.',
      side: 'right',
      align: 'start',
    },
  },
  {
    element: '#tour-quick-actions',
    popover: {
      title: '🚀 Ready to Go!',
      description: 'Use these quick actions to create agents, add TODOs, or schedule cron jobs. You\'re all set to start!',
      side: 'top',
      align: 'center',
    },
  },
]

// Create driver instance
const driverObj = driver({
  showProgress: true,
  showButtons: ['next', 'previous', 'close'],
  nextBtnText: 'Next →',
  prevBtnText: '← Back',
  doneBtnText: 'Finish Tour',
  progressText: '{{current}} of {{total}}',
  popoverClass: 'tour-popover',
  steps: tourSteps,
  onDestroyStarted: () => {
    // Mark tour as completed when finished
    if (!driverObj.hasNextStep()) {
      localStorage.setItem(TOUR_COMPLETED_KEY, 'true')
    }
    driverObj.destroy()
  },
  onCloseClick: () => {
    // Mark as skipped if closed early
    localStorage.setItem(TOUR_SKIPPED_KEY, Date.now().toString())
    driverObj.destroy()
  },
})

// Check if tour should be shown
export function shouldShowTour(): boolean {
  const completed = localStorage.getItem(TOUR_COMPLETED_KEY)
  const skipped = localStorage.getItem(TOUR_SKIPPED_KEY)
  
  // If completed, don't show
  if (completed === 'true') return false
  
  // If skipped less than 7 days ago, don't show
  if (skipped) {
    const skippedTime = parseInt(skipped, 10)
    const daysSinceSkipped = (Date.now() - skippedTime) / (1000 * 60 * 60 * 24)
    if (daysSinceSkipped < 7) return false
  }
  
  return true
}

// Check if tour is completed
export function isTourCompleted(): boolean {
  return localStorage.getItem(TOUR_COMPLETED_KEY) === 'true'
}

// Start the tour
export function startTour(): void {
  // Reset skip status when manually starting
  localStorage.removeItem(TOUR_SKIPPED_KEY)
  driverObj.drive()
}

// Reset tour progress
export function resetTourProgress(): void {
  localStorage.removeItem(TOUR_COMPLETED_KEY)
  localStorage.removeItem(TOUR_SKIPPED_KEY)
}

// Export driver for advanced usage
export { driverObj }
