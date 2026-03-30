import React from 'react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { renderToStaticMarkup } from 'react-dom/server'

import type { Monitor, MonitorLog } from '../../types'

vi.mock('../../hooks/useApi', () => ({
  useApi: vi.fn(),
}))

vi.mock('react-router-dom', () => ({
  Link: ({ to, children, ...props }: { to: string; children: React.ReactNode }) =>
    React.createElement('a', { href: to, ...props }, children),
  useParams: vi.fn(),
}))

import { useApi } from '../../hooks/useApi'
import { useParams } from 'react-router-dom'
import AdminMonitorLogs from './AdminMonitorLogs'

function makeApiResult<T>(data: T, loading = false, error: string | null = null) {
  return {
    data,
    total: 0,
    page: 1,
    totalPages: 1,
    loading,
    error,
    refetch: async () => undefined,
  }
}

const MONITOR: Monitor = {
  id: 'm1',
  name: 'API Health',
  type: 'http',
  target: 'https://example.com/health',
  intervalSeconds: 60,
  timeoutSeconds: 10,
  sslThresholds: [30, 14, 7],
  monitoring: {
    advanced: {
      cert_expiry: false,
      domain_expiry: false,
      ignore_tls_error: false,
    },
  },
  componentId: '',
  subComponentId: '',
  createdAt: '2026-01-01T00:00:00.000Z',
}

const LOG: MonitorLog = {
  id: 'l1',
  monitorId: 'm1',
  status: 'up',
  responseTime: 120,
  statusCode: 200,
  checkedAt: '2026-01-01T00:00:00.000Z',
}

describe('AdminMonitorLogs states', () => {
  const mockedUseApi = vi.mocked(useApi)
  const mockedUseParams = vi.mocked(useParams)

  beforeEach(() => {
    vi.clearAllMocks()
    mockedUseParams.mockReturnValue({ id: 'm1' })
  })

  it('renders loading state for logs', () => {
    mockedUseApi.mockImplementation((url: string) => {
      if (url === '/monitors') {
        return makeApiResult<Monitor[]>([MONITOR], false)
      }
      return makeApiResult<MonitorLog[] | null>(null, true)
    })

    const html = renderToStaticMarkup(<AdminMonitorLogs />)
    expect(html).toContain('Loading logs...')
  })

  it('renders error state for logs', () => {
    mockedUseApi.mockImplementation((url: string) => {
      if (url === '/monitors') {
        return makeApiResult<Monitor[]>([MONITOR], false)
      }
      return makeApiResult<MonitorLog[]>([], false, 'boom')
    })

    const html = renderToStaticMarkup(<AdminMonitorLogs />)
    expect(html).toContain('Failed to load logs.')
  })

  it('renders empty state when no logs exist', () => {
    mockedUseApi.mockImplementation((url: string) => {
      if (url === '/monitors') {
        return makeApiResult<Monitor[]>([MONITOR], false)
      }
      return makeApiResult<MonitorLog[]>([], false)
    })

    const html = renderToStaticMarkup(<AdminMonitorLogs />)
    expect(html).toContain('No logs yet for this monitor.')
  })

  it('renders logs table with fallback region', () => {
    mockedUseApi.mockImplementation((url: string) => {
      if (url === '/monitors') {
        return makeApiResult<Monitor[]>([MONITOR], false)
      }
      return makeApiResult<MonitorLog[]>([LOG], false)
    })

    const html = renderToStaticMarkup(<AdminMonitorLogs />)
    expect(html).toContain('Status Code')
    expect(html).toContain('global')
    expect(html).toContain('UP')
  })
})
