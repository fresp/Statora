import React from 'react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { renderToStaticMarkup } from 'react-dom/server'

import type { CategorySummary, StatusPageSettings } from '../types'

vi.mock('../hooks/useApi', () => ({
  useApi: vi.fn(),
  useCategorySummary: vi.fn(),
}))

vi.mock('../hooks/useWebSocket', () => ({
  useWebSocket: vi.fn(),
}))

vi.mock('../components/status/UptimeTimeline', () => ({
  UptimeTimeline: ({ history }: { history: Array<unknown> }) => React.createElement('div', { 'data-testid': 'uptime-timeline', 'data-bar-count': String(history.length) }, 'timeline'),
}))

vi.mock('../components/IncidentTimeline', () => ({
  IncidentTimeline: () => React.createElement('div', { 'data-testid': 'incident-timeline' }, 'incident timeline'),
}))

vi.mock('../components/content/ContentRenderer', () => ({
  default: ({ text }: { text?: string }) => React.createElement('div', null, text ?? ''),
}))

vi.mock('react-router-dom', () => ({
  Link: ({ to, children, ...props }: { to: string; children: React.ReactNode }) => React.createElement('a', { href: to, ...props }, children),
  useParams: vi.fn(),
  useLocation: vi.fn(),
}))

import { useApi, useCategorySummary } from '../hooks/useApi'
import { useLocation, useParams } from 'react-router-dom'
import StatusCategoryPage from './StatusCategoryPage'

const DEFAULT_SETTINGS: StatusPageSettings = {
  head: {
    title: 'Statora',
    description: 'Live system status and incident updates.',
    keywords: 'status, uptime, incidents, maintenance',
    faviconUrl: '/vite.svg',
    metaTags: {},
  },
  branding: {
    siteName: 'Statora',
    logoUrl: '',
    backgroundImageUrl: '',
    heroImageUrl: '',
  },
  theme: {
    preset: 'statora.css',
  },
  footer: {
    text: '',
    showPoweredBy: true,
  },
  customCss: '',
  updatedAt: '',
  createdAt: '',
}

function createHistory(days: number): CategorySummary['services'][number]['uptimeHistory'] {
  return Array.from({ length: days }, (_, index) => ({
    date: `2026-06-${String(index + 1).padStart(2, '0')}`,
    uptimePercent: 100,
    status: 'operational' as const,
  }))
}

function createSummary(services: CategorySummary['services']): CategorySummary {
  return {
    prefix: 'api-platform',
    name: 'API Platform',
    description: 'Category summary',
    aggregateStatus: 'operational',
    uptime90d: 99.95,
    services,
    incidents: [],
  }
}

describe('StatusCategoryPage', () => {
  const mockedUseApi = vi.mocked(useApi)
  const mockedUseCategorySummary = vi.mocked(useCategorySummary)
  const mockedUseParams = vi.mocked(useParams)
  const mockedUseLocation = vi.mocked(useLocation)

  beforeEach(() => {
    vi.clearAllMocks()
    vi.useFakeTimers()
    vi.setSystemTime(new Date('2026-06-13T12:00:00.000Z'))

    mockedUseParams.mockReturnValue({ categoryPrefix: 'api-platform' })
    mockedUseLocation.mockReturnValue({ pathname: '/status/api-platform', search: '', hash: '', state: null, key: 'test' })
    mockedUseApi.mockReturnValue({
      data: DEFAULT_SETTINGS,
      total: 0,
      page: 1,
      totalPages: 1,
      loading: false,
      error: null,
      refetch: async () => undefined,
    })
  })

  it('renders monitoring cards unchanged with uptime timeline and metric footer', () => {
    mockedUseCategorySummary.mockReturnValue({
      data: createSummary([
        {
          id: 'svc-monitoring',
          name: 'API Gateway',
          description: 'Monitored service',
          status: 'operational',
          updatedAt: '2026-06-13T10:00:00.000Z',
          uptime90d: 99.95,
          uptimeHistory: createHistory(30),
        },
      ]),
      loading: false,
      error: null,
      refetch: async () => undefined,
    })

    const html = renderToStaticMarkup(<StatusCategoryPage />)

    expect(html).toContain('data-testid="uptime-timeline"')
    expect(html).toContain('data-bar-count="30"')
    expect(html).toContain('99.95%')
    expect(html).toContain('30-Day Uptime')
    expect(html).not.toContain('NO ACTIVE INCIDENTS')
    expect(html).not.toContain('Last updated')
  })

  it('renders manual cards with large status typography and relative update time', () => {
    mockedUseCategorySummary.mockReturnValue({
      data: createSummary([
        {
          id: 'svc-manual',
          name: 'Email API',
          description: 'Manual service',
          status: 'operational',
          updatedAt: '2026-06-13T10:00:00.000Z',
          uptime90d: 100,
          uptimeHistory: [],
        },
      ]),
      loading: false,
      error: null,
      refetch: async () => undefined,
    })

    const html = renderToStaticMarkup(<StatusCategoryPage />)

    expect(html).toContain('NO ACTIVE INCIDENTS')
    expect(html).toContain('All systems operating normally')
    expect(html).toContain('Last updated')
    expect(html).toContain('2 hours ago')
    expect(html).not.toContain('data-testid="incident-timeline"')
  })
})
