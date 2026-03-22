import { useState, useEffect, useCallback } from 'react'
import { 
  Bot, 
  CheckCircle2, 
  Circle, 
  ArrowRight, 
  ArrowLeft, 
  Play, 
  Pause,
  RefreshCw,
  Eye,
  Terminal,
  Sparkles,
  ChevronDown,
  ChevronUp,
  Rocket,
  Target,
  Monitor,
  FileCheck,
  Trophy,
  ExternalLink,
  HelpCircle
} from 'lucide-react'

interface TutorialStep {
  id: number
  title: string
  description: string
  icon: typeof Bot
  content: React.ReactNode
  tips: string[]
}

const PROGRESS_KEY = 'agent-orchestrator-tutorial-progress'

export default function Tutorial() {
  const [currentStep, setCurrentStep] = useState(0)
  const [completedSteps, setCompletedSteps] = useState<Set<number>>(new Set())
  const [isPlaying, setIsPlaying] = useState(false)
  const [showTips, setShowTips] = useState(true)

  // Load progress from localStorage
  useEffect(() => {
    const saved = localStorage.getItem(PROGRESS_KEY)
    if (saved) {
      try {
        const data = JSON.parse(saved)
        setCompletedSteps(new Set(data.completedSteps || []))
        setCurrentStep(data.currentStep || 0)
      } catch (e) {
        console.error('Failed to load tutorial progress:', e)
      }
    }
  }, [])

  // Save progress to localStorage
  useEffect(() => {
    localStorage.setItem(PROGRESS_KEY, JSON.stringify({
      completedSteps: Array.from(completedSteps),
      currentStep
    }))
  }, [completedSteps, currentStep])

  const completeStep = useCallback((stepId: number) => {
    setCompletedSteps(prev => new Set([...prev, stepId]))
  }, [])

  const goNext = useCallback(() => {
    completeStep(currentStep)
    setCurrentStep(prev => Math.min(prev + 1, steps.length - 1))
  }, [currentStep, completeStep])

  const goPrev = useCallback(() => {
    setCurrentStep(prev => Math.max(prev - 1, 0))
  }, [])

  const goToStep = useCallback((step: number) => {
    setCurrentStep(step)
  }, [])

  const resetProgress = useCallback(() => {
    if (confirm('Are you sure you want to reset your tutorial progress?')) {
      setCompletedSteps(new Set())
      setCurrentStep(0)
      localStorage.removeItem(PROGRESS_KEY)
    }
  }, [])

  const progressPercentage = (completedSteps.size / steps.length) * 100
  const StepIcon = steps[currentStep]?.icon || Bot

  // Auto-play functionality
  useEffect(() => {
    if (!isPlaying) return
    
    const timer = setTimeout(() => {
      if (currentStep < steps.length - 1) {
        goNext()
      } else {
        setIsPlaying(false)
      }
    }, 8000)

    return () => clearTimeout(timer)
  }, [isPlaying, currentStep, goNext])

  return (
    <div className="space-y-6 animate-fade-in">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold text-text-primary flex items-center gap-3">
            <Sparkles className="w-7 h-7 text-accent" />
            Agent Orchestrator Tutorial
          </h1>
          <p className="text-text-secondary mt-1">Learn how to create and manage AI agents in 4 simple steps</p>
        </div>
        <div className="flex items-center gap-3">
          <button 
            onClick={resetProgress}
            className="btn-ghost text-sm flex items-center gap-2"
            title="Reset progress"
          >
            <RefreshCw className="w-4 h-4" />
            Reset
          </button>
        </div>
      </div>

      {/* Overall Progress */}
      <div className="card">
        <div className="flex items-center justify-between mb-3">
          <span className="text-sm text-text-secondary">Overall Progress</span>
          <span className="text-sm font-medium text-accent">{completedSteps.size} of {steps.length} completed</span>
        </div>
        <div className="h-2 bg-surface-hover rounded-full overflow-hidden">
          <div 
            className="h-full bg-gradient-to-r from-accent to-success transition-all duration-500 ease-out"
            style={{ width: `${progressPercentage}%` }}
          />
        </div>
        
        {/* Step indicators */}
        <div className="flex items-center justify-between mt-4">
          {steps.map((step, index) => (
            <button
              key={step.id}
              onClick={() => goToStep(index)}
              className={`flex flex-col items-center gap-1 transition-all duration-300 ${
                index === currentStep 
                  ? 'text-accent scale-110' 
                  : completedSteps.has(index) 
                    ? 'text-success' 
                    : 'text-text-muted hover:text-text-secondary'
              }`}
            >
              <div className={`w-8 h-8 rounded-full flex items-center justify-center transition-all duration-300 ${
                index === currentStep 
                  ? 'bg-accent/20 ring-2 ring-accent' 
                  : completedSteps.has(index) 
                    ? 'bg-success/20' 
                    : 'bg-surface-hover'
              }`}>
                {completedSteps.has(index) ? (
                  <CheckCircle2 className="w-5 h-5" />
                ) : (
                  <step.icon className="w-4 h-4" />
                )}
              </div>
              <span className="text-xs font-medium hidden sm:block">{step.title.split(' ').slice(-1)}</span>
            </button>
          ))}
        </div>
      </div>

      {/* Main Content */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Step Content */}
        <div className="lg:col-span-2 space-y-6">
          {/* Current Step Card */}
          <div className="card relative overflow-hidden">
            {/* Animated background gradient */}
            <div className="absolute inset-0 bg-gradient-to-br from-accent/5 via-transparent to-success/5 pointer-events-none" />
            
            <div className="relative">
              {/* Step Header */}
              <div className="flex items-start justify-between mb-6">
                <div className="flex items-center gap-4">
                  <div className="w-14 h-14 rounded-xl bg-accent/20 flex items-center justify-center animate-pulse-subtle">
                    <StepIcon className="w-7 h-7 text-accent" />
                  </div>
                  <div>
                    <div className="badge badge-accent mb-1">Step {currentStep + 1} of {steps.length}</div>
                    <h2 className="text-xl font-bold text-text-primary">{steps[currentStep].title}</h2>
                  </div>
                </div>
                
                {/* Auto-play controls */}
                <div className="flex items-center gap-2">
                  <button
                    onClick={() => setIsPlaying(!isPlaying)}
                    className={`p-2 rounded-lg transition-colors ${
                      isPlaying ? 'bg-accent text-white' : 'bg-surface-hover text-text-secondary hover:text-text-primary'
                    }`}
                    title={isPlaying ? 'Pause' : 'Auto-play'}
                  >
                    {isPlaying ? <Pause className="w-4 h-4" /> : <Play className="w-4 h-4" />}
                  </button>
                </div>
              </div>

              {/* Step Description */}
              <p className="text-text-secondary mb-6">{steps[currentStep].description}</p>

              {/* Interactive Content */}
              <div className="min-h-[300px]">
                {steps[currentStep].content}
              </div>

              {/* Navigation */}
              <div className="flex items-center justify-between mt-8 pt-6 border-t border-border">
                <button
                  onClick={goPrev}
                  disabled={currentStep === 0}
                  className="btn-secondary flex items-center gap-2 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  <ArrowLeft className="w-4 h-4" />
                  Previous
                </button>
                
                <div className="flex items-center gap-2">
                  {currentStep === steps.length - 1 && completedSteps.size === steps.length - 1 && (
                    <div className="flex items-center gap-2 text-success mr-4">
                      <Trophy className="w-5 h-5" />
                      <span className="text-sm font-medium">Tutorial Complete!</span>
                    </div>
                  )}
                </div>

                {currentStep < steps.length - 1 ? (
                  <button
                    onClick={goNext}
                    className="btn-primary flex items-center gap-2"
                  >
                    Next
                    <ArrowRight className="w-4 h-4" />
                  </button>
                ) : (
                  <button
                    onClick={() => {
                      completeStep(currentStep)
                      // Navigate to agents page
                      window.location.href = '/agents'
                    }}
                    className="btn-primary flex items-center gap-2 bg-success hover:bg-success/90"
                  >
                    <Rocket className="w-4 h-4" />
                    Get Started
                  </button>
                )}
              </div>
            </div>
          </div>
        </div>

        {/* Sidebar */}
        <div className="space-y-6">
          {/* Tips Panel */}
          <div className="card">
            <button
              onClick={() => setShowTips(!showTips)}
              className="flex items-center justify-between w-full text-left"
            >
              <h3 className="font-semibold text-text-primary flex items-center gap-2">
                <HelpCircle className="w-4 h-4 text-accent" />
                Tips & Hints
              </h3>
              {showTips ? <ChevronUp className="w-4 h-4" /> : <ChevronDown className="w-4 h-4" />}
            </button>
            
            {showTips && (
              <div className="mt-4 space-y-2">
                {steps[currentStep].tips.map((tip, index) => (
                  <div 
                    key={index}
                    className="flex items-start gap-2 text-sm text-text-secondary p-2 rounded-lg bg-surface-hover/50"
                  >
                    <Sparkles className="w-4 h-4 text-accent shrink-0 mt-0.5" />
                    <span>{tip}</span>
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* Quick Links */}
          <div className="card">
            <h3 className="font-semibold text-text-primary mb-4">Quick Links</h3>
            <div className="space-y-2">
              <a href="/agents" className="flex items-center gap-3 p-2 rounded-lg hover:bg-surface-hover transition-colors text-text-secondary hover:text-text-primary">
                <Bot className="w-4 h-4" />
                <span className="text-sm">Agents Dashboard</span>
                <ExternalLink className="w-3 h-3 ml-auto" />
              </a>
              <a href="/todos" className="flex items-center gap-3 p-2 rounded-lg hover:bg-surface-hover transition-colors text-text-secondary hover:text-text-primary">
                <Target className="w-4 h-4" />
                <span className="text-sm">Task List</span>
                <ExternalLink className="w-3 h-3 ml-auto" />
              </a>
              <a href="/config" className="flex items-center gap-3 p-2 rounded-lg hover:bg-surface-hover transition-colors text-text-secondary hover:text-text-primary">
                <Terminal className="w-4 h-4" />
                <span className="text-sm">Configuration</span>
                <ExternalLink className="w-3 h-3 ml-auto" />
              </a>
            </div>
          </div>

          {/* Progress Summary */}
          <div className="card bg-gradient-to-br from-accent/10 to-success/10 border-accent/20">
            <h3 className="font-semibold text-text-primary mb-3">Your Progress</h3>
            <div className="space-y-3">
              {steps.map((step, index) => (
                <div 
                  key={step.id}
                  className={`flex items-center gap-3 text-sm ${
                    completedSteps.has(index) ? 'text-success' : 'text-text-muted'
                  }`}
                >
                  {completedSteps.has(index) ? (
                    <CheckCircle2 className="w-4 h-4" />
                  ) : (
                    <Circle className="w-4 h-4" />
                  )}
                  <span className={index === currentStep ? 'text-accent font-medium' : ''}>
                    {step.title}
                  </span>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

// Step Components
function CreateAgentStep() {
  const [formData, setFormData] = useState({
    name: 'my-first-agent',
    provider: 'openai',
    model: 'gpt-4o',
    work_dir: '~/sandbox'
  })
  const [isCreating, setIsCreating] = useState(false)
  const [created, setCreated] = useState(false)

  const handleCreate = () => {
    setIsCreating(true)
    setTimeout(() => {
      setIsCreating(false)
      setCreated(true)
    }, 1500)
  }

  if (created) {
    return (
      <div className="flex flex-col items-center justify-center py-8 animate-fade-in">
        <div className="w-20 h-20 rounded-full bg-success/20 flex items-center justify-center mb-4">
          <CheckCircle2 className="w-10 h-10 text-success" />
        </div>
        <h3 className="text-lg font-semibold text-text-primary mb-2">Agent Created!</h3>
        <p className="text-text-secondary text-center mb-4">
          Your agent <span className="text-accent font-mono">{formData.name}</span> is ready to go!
        </p>
        <button 
          onClick={() => setCreated(false)}
          className="btn-secondary text-sm"
        >
          Try Again
        </button>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      <div className="bg-surface-hover/50 rounded-lg p-4 mb-4">
        <p className="text-sm text-text-secondary flex items-center gap-2">
          <Bot className="w-4 h-4 text-accent" />
          An agent is an AI assistant that can execute tasks autonomously
        </p>
      </div>

      <div>
        <label className="label">Agent Name</label>
        <input
          type="text"
          value={formData.name}
          onChange={(e) => setFormData({ ...formData, name: e.target.value })}
          className="input"
          placeholder="e.g., code-reviewer"
        />
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div>
          <label className="label">Provider</label>
          <select
            value={formData.provider}
            onChange={(e) => setFormData({ ...formData, provider: e.target.value })}
            className="input"
          >
            <option value="openai">OpenAI</option>
            <option value="anthropic">Anthropic</option>
            <option value="ollama">Ollama (Local)</option>
          </select>
        </div>
        <div>
          <label className="label">Model</label>
          <select
            value={formData.model}
            onChange={(e) => setFormData({ ...formData, model: e.target.value })}
            className="input"
          >
            <option value="gpt-4o">GPT-4o</option>
            <option value="gpt-4-turbo">GPT-4 Turbo</option>
            <option value="claude-3-opus">Claude 3 Opus</option>
          </select>
        </div>
      </div>

      <div>
        <label className="label">Working Directory</label>
        <input
          type="text"
          value={formData.work_dir}
          onChange={(e) => setFormData({ ...formData, work_dir: e.target.value })}
          className="input font-mono text-sm"
        />
      </div>

      <button
        onClick={handleCreate}
        disabled={isCreating}
        className="btn-primary w-full flex items-center justify-center gap-2"
      >
        {isCreating ? (
          <>
            <RefreshCw className="w-4 h-4 animate-spin" />
            Creating...
          </>
        ) : (
          <>
            <Bot className="w-4 h-4" />
            Create Agent
          </>
        )}
      </button>
    </div>
  )
}

function AssignTaskStep() {
  const [taskData, setTaskData] = useState({
    title: 'Analyze codebase structure',
    description: 'Review the main components and provide a summary',
    priority: 7
  })
  const [isAssigning, setIsAssigning] = useState(false)
  const [assigned, setAssigned] = useState(false)

  const handleAssign = () => {
    setIsAssigning(true)
    setTimeout(() => {
      setIsAssigning(false)
      setAssigned(true)
    }, 1500)
  }

  if (assigned) {
    return (
      <div className="flex flex-col items-center justify-center py-8 animate-fade-in">
        <div className="w-20 h-20 rounded-full bg-success/20 flex items-center justify-center mb-4">
          <CheckCircle2 className="w-10 h-10 text-success" />
        </div>
        <h3 className="text-lg font-semibold text-text-primary mb-2">Task Assigned!</h3>
        <p className="text-text-secondary text-center">
          Your agent will start working on this task automatically
        </p>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      <div className="bg-surface-hover/50 rounded-lg p-4 mb-4">
        <p className="text-sm text-text-secondary flex items-center gap-2">
          <Target className="w-4 h-4 text-accent" />
          Tasks define what your agent should accomplish
        </p>
      </div>

      <div>
        <label className="label">Task Title</label>
        <input
          type="text"
          value={taskData.title}
          onChange={(e) => setTaskData({ ...taskData, title: e.target.value })}
          className="input"
        />
      </div>

      <div>
        <label className="label">Description</label>
        <textarea
          value={taskData.description}
          onChange={(e) => setTaskData({ ...taskData, description: e.target.value })}
          className="input min-h-[80px] resize-y"
        />
      </div>

      <div>
        <label className="label">Priority: {taskData.priority}/10</label>
        <input
          type="range"
          min="1"
          max="10"
          value={taskData.priority}
          onChange={(e) => setTaskData({ ...taskData, priority: parseInt(e.target.value) })}
          className="w-full h-2 bg-surface-hover rounded-lg appearance-none cursor-pointer"
        />
        <div className="flex justify-between text-xs text-text-muted mt-1">
          <span>Low</span>
          <span>High</span>
        </div>
      </div>

      <button
        onClick={handleAssign}
        disabled={isAssigning}
        className="btn-primary w-full flex items-center justify-center gap-2"
      >
        {isAssigning ? (
          <>
            <RefreshCw className="w-4 h-4 animate-spin" />
            Assigning...
          </>
        ) : (
          <>
            <Target className="w-4 h-4" />
            Assign to Agent
          </>
        )}
      </button>
    </div>
  )
}

function MonitorProgressStep() {
  const [progress, setProgress] = useState(0)
  const [status, setStatus] = useState<'idle' | 'running' | 'completed'>('idle')
  const [logs, setLogs] = useState<string[]>([])

  const simulateProgress = () => {
    setProgress(0)
    setStatus('running')
    setLogs([])

    const logMessages = [
      '🤖 Agent started processing task...',
      '📂 Reading workspace files...',
      '🔍 Analyzing code structure...',
      '📊 Identifying main components...',
      '✨ Generating summary...',
    ]

    let currentProgress = 0
    const interval = setInterval(() => {
      currentProgress += 20
      setProgress(currentProgress)

      if (logMessages.length > 0) {
        setLogs(prev => [...prev, logMessages.shift()!])
      }

      if (currentProgress >= 100) {
        clearInterval(interval)
        setStatus('completed')
        setLogs(prev => [...prev, '✅ Task completed successfully!'])
      }
    }, 800)
  }

  return (
    <div className="space-y-4">
      <div className="bg-surface-hover/50 rounded-lg p-4 mb-4">
        <p className="text-sm text-text-secondary flex items-center gap-2">
          <Monitor className="w-4 h-4 text-accent" />
          Monitor your agent's progress in real-time
        </p>
      </div>

      {/* Status Card */}
      <div className="bg-surface-hover rounded-lg p-4">
        <div className="flex items-center justify-between mb-3">
          <div className="flex items-center gap-3">
            <div className={`w-3 h-3 rounded-full ${
              status === 'running' ? 'bg-accent animate-pulse' :
              status === 'completed' ? 'bg-success' : 'bg-text-muted'
            }`} />
            <span className="font-medium capitalize">{status}</span>
          </div>
          <span className="text-sm text-text-secondary">{progress}%</span>
        </div>
        
        <div className="h-2 bg-surface rounded-full overflow-hidden">
          <div 
            className={`h-full transition-all duration-300 ${
              status === 'completed' ? 'bg-success' : 'bg-accent'
            }`}
            style={{ width: `${progress}%` }}
          />
        </div>
      </div>

      {/* Activity Log */}
      <div className="bg-surface-hover rounded-lg p-4 min-h-[150px]">
        <div className="flex items-center gap-2 mb-3 text-text-secondary">
          <Terminal className="w-4 h-4" />
          <span className="text-sm font-medium">Activity Log</span>
        </div>
        <div className="font-mono text-sm space-y-1">
          {logs.length === 0 ? (
            <p className="text-text-muted italic">Click "Simulate Execution" to see activity...</p>
          ) : (
            logs.map((log, i) => (
              <div key={i} className="text-text-secondary animate-fade-in">
                {log}
              </div>
            ))
          )}
        </div>
      </div>

      <button
        onClick={simulateProgress}
        disabled={status === 'running'}
        className="btn-primary w-full flex items-center justify-center gap-2"
      >
        {status === 'running' ? (
          <>
            <RefreshCw className="w-4 h-4 animate-spin" />
            Executing...
          </>
        ) : (
          <>
            <Play className="w-4 h-4" />
            Simulate Execution
          </>
        )}
      </button>
    </div>
  )
}

function ReviewResultsStep() {
  const [showResults, setShowResults] = useState(false)

  const results = {
    summary: 'Your agent has successfully completed the codebase analysis.',
    details: [
      { label: 'Files Analyzed', value: '24' },
      { label: 'Components Found', value: '12' },
      { label: 'Lines of Code', value: '3,842' },
      { label: 'Execution Time', value: '12.3s' },
    ],
    output: `# Codebase Analysis Report

## Project Structure
- /src/components - React components
- /src/hooks - Custom React hooks
- /src/lib - Utility functions
- /src/types - TypeScript definitions

## Main Technologies
- React 18 with TypeScript
- TailwindCSS for styling
- Vite for bundling

## Recommendations
1. Consider adding more unit tests
2. Document API endpoints
3. Add error boundary components`
  }

  return (
    <div className="space-y-4">
      <div className="bg-surface-hover/50 rounded-lg p-4 mb-4">
        <p className="text-sm text-text-secondary flex items-center gap-2">
          <FileCheck className="w-4 h-4 text-accent" />
          Review the results and provide feedback
        </p>
      </div>

      {!showResults ? (
        <div className="flex flex-col items-center justify-center py-8">
          <div className="w-16 h-16 rounded-full bg-accent/20 flex items-center justify-center mb-4">
            <Eye className="w-8 h-8 text-accent" />
          </div>
          <h3 className="text-lg font-semibold text-text-primary mb-2">Ready for Review</h3>
          <p className="text-text-secondary text-center mb-4">
            Your agent has completed its task and is waiting for your review
          </p>
          <button
            onClick={() => setShowResults(true)}
            className="btn-primary flex items-center gap-2"
          >
            <FileCheck className="w-4 h-4" />
            View Results
          </button>
        </div>
      ) : (
        <div className="space-y-4 animate-fade-in">
          {/* Summary Card */}
          <div className="bg-success/10 border border-success/20 rounded-lg p-4">
            <p className="text-success flex items-center gap-2">
              <CheckCircle2 className="w-5 h-5" />
              {results.summary}
            </p>
          </div>

          {/* Stats Grid */}
          <div className="grid grid-cols-2 gap-3">
            {results.details.map((detail, i) => (
              <div key={i} className="bg-surface-hover rounded-lg p-3 text-center">
                <p className="text-2xl font-bold text-accent">{detail.value}</p>
                <p className="text-xs text-text-muted">{detail.label}</p>
              </div>
            ))}
          </div>

          {/* Output Preview */}
          <div className="bg-surface-hover rounded-lg p-4">
            <div className="flex items-center gap-2 mb-3 text-text-secondary">
              <FileCheck className="w-4 h-4" />
              <span className="text-sm font-medium">Agent Output</span>
            </div>
            <pre className="font-mono text-xs text-text-secondary whitespace-pre-wrap bg-surface rounded p-3 max-h-[150px] overflow-y-auto">
              {results.output}
            </pre>
          </div>

          {/* Actions */}
          <div className="flex gap-3">
            <button className="btn-secondary flex-1">Request Changes</button>
            <button className="btn-primary flex-1 bg-success hover:bg-success/90">Approve</button>
          </div>
        </div>
      )}
    </div>
  )
}

// Tutorial Steps Definition
const steps: TutorialStep[] = [
  {
    id: 1,
    title: 'Create an Agent',
    description: 'Learn how to create your first AI agent. Agents are autonomous assistants that can execute tasks on your behalf.',
    icon: Bot,
    content: <CreateAgentStep />,
    tips: [
      'Choose a descriptive name for easy identification',
      'Select the provider based on your needs',
      'GPT-4o is recommended for complex tasks',
      'Local models work offline but may be less capable'
    ]
  },
  {
    id: 2,
    title: 'Assign a Task',
    description: 'Give your agent a clear objective. Well-defined tasks lead to better results.',
    icon: Target,
    content: <AssignTaskStep />,
    tips: [
      'Be specific about what you want',
      'Include context and requirements',
      'Set appropriate priority levels',
      'Break complex tasks into smaller ones'
    ]
  },
  {
    id: 3,
    title: 'Monitor Progress',
    description: 'Watch your agent work in real-time. Track progress, view logs, and stay informed.',
    icon: Monitor,
    content: <MonitorProgressStep />,
    tips: [
      'Check activity logs for details',
      'You can pause/resume execution',
      'Monitor resource usage',
      'Intervene if something goes wrong'
    ]
  },
  {
    id: 4,
    title: 'Review Results',
    description: 'Evaluate the output, provide feedback, and approve or request changes.',
    icon: FileCheck,
    content: <ReviewResultsStep />,
    tips: [
      'Review the summary first',
      'Check the detailed output',
      'Provide feedback for improvements',
      'Approve to mark task complete'
    ]
  }
]
