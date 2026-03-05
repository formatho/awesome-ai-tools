import { useParams, Link } from 'react-router-dom'
import { ArrowLeft, MessageSquare } from 'lucide-react'
import ChatWindow from './ChatWindow'
import { useAgent } from '../../hooks/useAPI'

export default function ChatPage() {
  const { id: agentId } = useParams<{ id: string }>()
  const { data: agent, isLoading } = useAgent(agentId || '')

  if (!agentId) {
    return (
      <div className="flex items-center justify-center h-full">
        <p className="text-error">No agent ID provided</p>
      </div>
    )
  }

  return (
    <div className="flex flex-col h-[calc(100vh-8rem)] animate-fade-in">
      {/* Header */}
      <div className="card mb-4">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-4">
            <Link 
              to="/agents" 
              className="p-2 hover:bg-surface-hover rounded-lg text-text-secondary hover:text-text-primary transition-colors"
            >
              <ArrowLeft className="w-5 h-5" />
            </Link>
            <div className="flex items-center gap-3">
              <div className="w-10 h-10 rounded-lg bg-accent/20 flex items-center justify-center">
                <MessageSquare className="w-5 h-5 text-accent" />
              </div>
              <div>
                {isLoading ? (
                  <p className="text-text-muted">Loading...</p>
                ) : (
                  <>
                    <h1 className="text-lg font-semibold text-text-primary">
                      {agent?.name || 'Agent Chat'}
                    </h1>
                    <p className="text-sm text-text-muted">
                      {agent?.type || 'Chat with agent'}
                    </p>
                  </>
                )}
              </div>
            </div>
          </div>
          
          {!isLoading && agent && (
            <div className="flex items-center gap-2">
              <span className={`status-dot ${agent.status === 'running' ? 'online' : 'offline'}`} />
              <span className="text-sm capitalize">{agent.status}</span>
            </div>
          )}
        </div>
      </div>

      {/* Chat Window */}
      <div className="card flex-1 overflow-hidden">
        <ChatWindow agentId={agentId} />
      </div>
    </div>
  )
}
