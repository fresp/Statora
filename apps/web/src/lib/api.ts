import axios from 'axios'
import { clearAuthSession, requiresMfa } from './auth'
import type { SetupSaveRequest, SetupStatusResponse } from '../types'

const API_BASE = import.meta.env.VITE_API_URL || '/api'

const api = axios.create({
  baseURL: API_BASE,
  headers: { 'Content-Type': 'application/json' },
})

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('user_token') || localStorage.getItem('admin_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

api.interceptors.response.use(
  (res) => res,
  (err) => {
    const status = err.response?.status
    if (status === 401) {
      const pathname = window.location.pathname
      const isAdminRoute = pathname.startsWith('/admin')
      
      if (isAdminRoute) {
        clearAuthSession()
        localStorage.removeItem('admin_token')
        localStorage.removeItem('admin_profile')
        window.location.href = '/admin/login'
      }
    }

    if (status === 403) {
      const pathname = window.location.pathname
      if (pathname.startsWith('/admin') && requiresMfa()) {
        window.location.href = '/admin/profile'
      }
    }

    return Promise.reject(err)
  }
)

export default api

export async function getSetupStatus(): Promise<SetupStatusResponse> {
  const res = await api.get<SetupStatusResponse>('/setup/status')
  return res.data
}

export async function validateMongoSetup(mongoUri: string, mongoDbName: string): Promise<{ valid: boolean }> {
  const res = await api.post<{ valid: boolean }>('/setup/validate/mongo', { mongoUri, mongoDbName })
  return res.data
}

export async function saveSetupConfig(payload: SetupSaveRequest): Promise<SetupStatusResponse> {
  const res = await api.post<SetupStatusResponse>('/setup/save', payload)
  return res.data
}
