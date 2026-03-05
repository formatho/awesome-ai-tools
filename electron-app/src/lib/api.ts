import axios, { type AxiosInstance, AxiosError } from 'axios'

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:18765/api'

// Create axios instance with default config
const api: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Request interceptor
api.interceptors.request.use(
  (config) => {
    // Add auth token if available
    const token = localStorage.getItem('authToken')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor
api.interceptors.response.use(
  (response) => response,
  (error: AxiosError) => {
    if (error.response) {
      // Server responded with error status
      const { status, data } = error.response

      switch (status) {
        case 401:
          // Unauthorized - clear token and redirect to login
          localStorage.removeItem('authToken')
          window.location.href = '/login'
          break
        case 403:
          console.error('Access forbidden')
          break
        case 404:
          console.error('Resource not found')
          break
        case 500:
          console.error('Server error')
          break
        default:
          console.error('API Error:', data)
      }
    } else if (error.request) {
      // Request was made but no response received
      console.error('No response from server')
    } else {
      console.error('Request error:', error.message)
    }

    return Promise.reject(error)
  }
)

// API endpoints
export const agentsAPI = {
  list: () => api.get('/agents'),
  get: (id: string) => api.get(`/agents/${id}`),
  create: (data: unknown) => api.post('/agents', data),
  update: (id: string, data: unknown) => api.put(`/agents/${id}`, data),
  delete: (id: string) => api.delete(`/agents/${id}`),
  start: (id: string) => api.post(`/agents/${id}/start`),
  stop: (id: string) => api.post(`/agents/${id}/stop`),
  logs: (id: string) => api.get(`/agents/${id}/logs`),
}

export const todosAPI = {
  list: (params?: { status?: string; priority?: string }) => api.get('/todos', { params }),
  get: (id: string) => api.get(`/todos/${id}`),
  create: (data: unknown) => api.post('/todos', data),
  update: (id: string, data: unknown) => api.put(`/todos/${id}`, data),
  delete: (id: string) => api.delete(`/todos/${id}`),
  complete: (id: string) => api.post(`/todos/${id}/complete`),
}

export const cronAPI = {
  list: () => api.get('/cron'),
  get: (id: string) => api.get(`/cron/${id}`),
  create: (data: unknown) => api.post('/cron', data),
  update: (id: string, data: unknown) => api.put(`/cron/${id}`, data),
  delete: (id: string) => api.delete(`/cron/${id}`),
  pause: (id: string) => api.post(`/cron/${id}/pause`),
  resume: (id: string) => api.post(`/cron/${id}/resume`),
  history: (id: string) => api.get(`/cron/${id}/history`),
}

export const configAPI = {
  get: () => api.get('/config'),
  update: (data: unknown) => api.put('/config', data),
}

export const healthAPI = {
  check: () => api.get('/health'),
}

export const chatAPI = {
  history: (agentId: string) => api.get(`/agents/${agentId}/chat`),
  send: (agentId: string, message: string) => api.post(`/agents/${agentId}/chat`, { message }),
}

export default api
