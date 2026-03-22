import { useState, useEffect } from 'react'
import { Minus, Square, X, Maximize2, PlayCircle } from 'lucide-react'
import { startTour, isTourCompleted } from '../../lib/productTour'

declare global {
  interface Window {
    electronAPI?: {
      minimizeWindow: () => void
      maximizeWindow: () => void
      closeWindow: () => void
      isMaximized: () => Promise<boolean>
    }
  }
}

export default function Header() {
  const [isMaximized, setIsMaximized] = useState(false)
  const [tourCompleted, setTourCompleted] = useState(false)

  useEffect(() => {
    const checkMaximized = async () => {
      if (window.electronAPI) {
        const maximized = await window.electronAPI.isMaximized()
        setIsMaximized(maximized)
      }
    }
    checkMaximized()
    setTourCompleted(isTourCompleted())
  }, [])

  const handleMaximize = async () => {
    if (window.electronAPI) {
      window.electronAPI.maximizeWindow()
      const maximized = await window.electronAPI.isMaximized()
      setIsMaximized(maximized)
    }
  }

  const handleStartTour = () => {
    startTour()
  }

  return (
    <header className="h-12 bg-surface border-b border-border flex items-center justify-between px-4 drag-region">
      {/* Window Title */}
      <div className="flex items-center gap-3">
        <span className="text-sm text-text-secondary">Agent Orchestrator</span>
        {!tourCompleted && (
          <button
            onClick={handleStartTour}
            className="no-drag flex items-center gap-1.5 px-2 py-1 text-xs bg-accent/10 text-accent rounded-md hover:bg-accent/20 transition-colors"
          >
            <PlayCircle className="w-3.5 h-3.5" />
            Start Tour
          </button>
        )}
      </div>

      {/* Window Controls */}
      <div className="flex items-center gap-1 no-drag">
        <button
          onClick={() => window.electronAPI?.minimizeWindow()}
          className="w-8 h-8 flex items-center justify-center text-text-muted hover:bg-surface-hover rounded transition-colors"
          aria-label="Minimize"
        >
          <Minus className="w-4 h-4" />
        </button>
        <button
          onClick={handleMaximize}
          className="w-8 h-8 flex items-center justify-center text-text-muted hover:bg-surface-hover rounded transition-colors"
          aria-label="Maximize"
        >
          {isMaximized ? <Square className="w-3.5 h-3.5" /> : <Maximize2 className="w-3.5 h-3.5" />}
        </button>
        <button
          onClick={() => window.electronAPI?.closeWindow()}
          className="w-8 h-8 flex items-center justify-center text-text-muted hover:bg-error hover:text-white rounded transition-colors"
          aria-label="Close"
        >
          <X className="w-4 h-4" />
        </button>
      </div>
    </header>
  )
}
