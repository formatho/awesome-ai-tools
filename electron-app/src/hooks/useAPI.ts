import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { agentsAPI, todosAPI, cronAPI, configAPI, healthAPI, chatAPI } from '../lib/api'

// Query keys
export const queryKeys = {
  agents: ['agents'] as const,
  agent: (id: string) => ['agents', id] as const,
  agentLogs: (id: string) => ['agents', id, 'logs'] as const,
  chat: (agentId: string) => ['chat', agentId] as const,
  todos: (filters?: { status?: string; priority?: string }) => ['todos', filters] as const,
  todo: (id: string) => ['todos', id] as const,
  cronJobs: ['cron'] as const,
  cronJob: (id: string) => ['cron', id] as const,
  config: ['config'] as const,
  health: ['health'] as const,
}

// Agents hooks
export function useAgents() {
  return useQuery({
    queryKey: queryKeys.agents,
    queryFn: () => agentsAPI.list(),
    select: (response) => response.data,
  })
}

export function useAgent(id: string) {
  return useQuery({
    queryKey: queryKeys.agent(id),
    queryFn: () => agentsAPI.get(id),
    select: (response) => response.data,
    enabled: !!id,
  })
}

export function useAgentLogs(id: string) {
  return useQuery({
    queryKey: queryKeys.agentLogs(id),
    queryFn: () => agentsAPI.logs(id),
    select: (response) => response.data,
    enabled: !!id,
    refetchInterval: 2000, // Poll every 2 seconds for new logs
  })
}

export function useAgentMutations() {
  const queryClient = useQueryClient()

  const createMutation = useMutation({
    mutationFn: (data: unknown) => agentsAPI.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.agents })
    },
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: unknown }) => agentsAPI.update(id, data),
    onSuccess: (_, { id }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.agents })
      queryClient.invalidateQueries({ queryKey: queryKeys.agent(id) })
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => agentsAPI.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.agents })
    },
  })

  const startMutation = useMutation({
    mutationFn: (id: string) => agentsAPI.start(id),
    onSuccess: (_, id) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.agent(id) })
    },
  })

  const stopMutation = useMutation({
    mutationFn: (id: string) => agentsAPI.stop(id),
    onSuccess: (_, id) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.agent(id) })
    },
  })

  return {
    create: createMutation,
    update: updateMutation,
    delete: deleteMutation,
    start: startMutation,
    stop: stopMutation,
  }
}

// Chat hooks
export function useChat(agentId: string) {
  return useQuery({
    queryKey: queryKeys.chat(agentId),
    queryFn: () => chatAPI.history(agentId),
    select: (response) => response.data,
    enabled: !!agentId,
  })
}

export function useChatMutations() {
  const queryClient = useQueryClient()

  const sendMutation = useMutation({
    mutationFn: ({ agentId, message }: { agentId: string; message: string }) =>
      chatAPI.send(agentId, message),
    onSuccess: (_, { agentId }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.chat(agentId) })
    },
  })

  return {
    send: sendMutation,
  }
}

// TODOs hooks
export function useTODOs(filters?: { status?: string; priority?: string }) {
  return useQuery({
    queryKey: queryKeys.todos(filters),
    queryFn: () => todosAPI.list(filters),
    select: (response) => response.data,
  })
}

export function useTODO(id: string) {
  return useQuery({
    queryKey: queryKeys.todo(id),
    queryFn: () => todosAPI.get(id),
    select: (response) => response.data,
    enabled: !!id,
  })
}

export function useTODOMutations() {
  const queryClient = useQueryClient()

  const createMutation = useMutation({
    mutationFn: (data: unknown) => todosAPI.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.todos() })
    },
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: unknown }) => todosAPI.update(id, data),
    onSuccess: (_, { id }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.todos() })
      queryClient.invalidateQueries({ queryKey: queryKeys.todo(id) })
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => todosAPI.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.todos() })
    },
  })

  const completeMutation = useMutation({
    mutationFn: (id: string) => todosAPI.complete(id),
    onSuccess: (_, id) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.todos() })
      queryClient.invalidateQueries({ queryKey: queryKeys.todo(id) })
    },
  })

  return {
    create: createMutation,
    update: updateMutation,
    delete: deleteMutation,
    complete: completeMutation,
  }
}

// Cron hooks
export function useCronJobs() {
  return useQuery({
    queryKey: queryKeys.cronJobs,
    queryFn: () => cronAPI.list(),
    select: (response) => response.data,
  })
}

export function useCronJob(id: string) {
  return useQuery({
    queryKey: queryKeys.cronJob(id),
    queryFn: () => cronAPI.get(id),
    select: (response) => response.data,
    enabled: !!id,
  })
}

export function useCronMutations() {
  const queryClient = useQueryClient()

  const createMutation = useMutation({
    mutationFn: (data: unknown) => cronAPI.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.cronJobs })
    },
  })

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: unknown }) => cronAPI.update(id, data),
    onSuccess: (_, { id }) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.cronJobs })
      queryClient.invalidateQueries({ queryKey: queryKeys.cronJob(id) })
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => cronAPI.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.cronJobs })
    },
  })

  const pauseMutation = useMutation({
    mutationFn: (id: string) => cronAPI.pause(id),
    onSuccess: (_, id) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.cronJob(id) })
    },
  })

  const resumeMutation = useMutation({
    mutationFn: (id: string) => cronAPI.resume(id),
    onSuccess: (_, id) => {
      queryClient.invalidateQueries({ queryKey: queryKeys.cronJob(id) })
    },
  })

  return {
    create: createMutation,
    update: updateMutation,
    delete: deleteMutation,
    pause: pauseMutation,
    resume: resumeMutation,
  }
}

// Config hooks
export function useConfig() {
  return useQuery({
    queryKey: queryKeys.config,
    queryFn: () => configAPI.get(),
    select: (response) => response.data,
  })
}

export function useConfigMutation() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (data: unknown) => configAPI.update(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: queryKeys.config })
    },
  })
}

// Health check hook
export function useHealthCheck() {
  return useQuery({
    queryKey: queryKeys.health,
    queryFn: () => healthAPI.check(),
    select: (response) => response.data,
    refetchInterval: 30000, // Check every 30 seconds
  })
}
