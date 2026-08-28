import React from 'react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { renderToStaticMarkup } from 'react-dom/server'

import type { AvailabilityResponse } from '../../types'

vi.mock('../../hooks/useApi', () => ({
  useApi: vi.fn(),
}))

import { useApi } from '../../hooks/useApi'
import { AvailabilityMetrics } from './AvailabilityMetrics'

const useApiMock = vi.mocked(useApi)

function createResponse(overrides: Partial<AvailabilityResponse> = {}): AvailabilityResponse {
  return {
    period: {
      label: 'Last 30 Days',
      start: '2026-07-29T14:30:00Z',
      end: '2026-08-28T14:30:00Z',
    },
    overall: {
      availability: 99.9144,
      totalMinutes: 43200,
      downtimeMinutes: 37,
      incidentCount: 1,
    },
    incidents: [
      {
        id: 'incident-1',
        title: 'Intermittent WhatsApp API',
        impact: 'minor',
        status: 'resolved',
        startedAt: '2026-08-15T10:00:00Z',
        resolvedAt: '2026-08-15T10:37:00Z',
        effectiveDowntimeMinutes: 37,
        affectedComponents: [{ id: 'component-1', name: 'WhatsApp API' }],
      },
    ],
    services: [
      {
        componentId: 'component-1',
        name: 'WhatsApp API',
        availability: 99.9144,
        downtimeMinutes: 37,
        incidentCount: 1,
      },
    ],
    ...overrides,
  }
}

function renderComponent(): string {
  return renderToStaticMarkup(React.createElement(AvailabilityMetrics))
}

beforeEach(() => {
  useApiMock.mockReset()
  useApiMock.mockReturnValue({
    data: null,
    total: 0,
    page: 1,
    totalPages: 1,
    loading: true,
    error: null,
    refetch: vi.fn(),
  })
})

describe('AvailabilityMetrics', () => {
  it('renders loading skeleton while fetching', () => {
    const markup = renderComponent()
    expect(markup).toContain('data-testid="availability-loading"')
    expect(markup).toContain('animate-pulse')
  })

  it('renders error state with retry button', () => {
    useApiMock.mockReturnValue({
      data: null,
      total: 0,
      page: 1,
      totalPages: 1,
      loading: false,
      error: 'Request failed',
      refetch: vi.fn(),
    })

    const markup = renderComponent()
    expect(markup).toContain('data-testid="availability-error"')
    expect(markup).toContain('Failed to load availability metrics')
    expect(markup).toContain('Retry')
  })

  it('renders summary cards and breakdowns with data', () => {
    useApiMock.mockReturnValue({
      data: createResponse(),
      total: 0,
      page: 1,
      totalPages: 1,
      loading: false,
      error: null,
      refetch: vi.fn(),
    })

    const markup = renderComponent()
    expect(markup).toContain('data-testid="availability-percent"')
    expect(markup).toContain('99.91%')
    expect(markup).toContain('37m')
    expect(markup).toContain('Intermittent WhatsApp API')
    expect(markup).toContain('WhatsApp API')
    expect(markup).toContain('Downtime by Incident')
    expect(markup).toContain('Availability by Service')
    expect(markup).not.toContain('data-testid="availability-active-note"')
  })

  it('shows 100 percent and empty state when no incidents', () => {
    useApiMock.mockReturnValue({
      data: createResponse({
        overall: {
          availability: 100,
          totalMinutes: 43200,
          downtimeMinutes: 0,
          incidentCount: 0,
        },
        incidents: [],
        services: [],
      }),
      total: 0,
      page: 1,
      totalPages: 1,
      loading: false,
      error: null,
      refetch: vi.fn(),
    })

    const markup = renderComponent()
    expect(markup).toContain('data-testid="availability-percent"')
    expect(markup).toContain('100%')
    expect(markup).toContain('No incidents in this period')
    expect(markup).not.toContain('Downtime by Incident')
    expect(markup).not.toContain('Availability by Service')
  })

  it('shows live availability note for active incidents', () => {
    useApiMock.mockReturnValue({
      data: createResponse({
        incidents: [
          {
            id: 'incident-2',
            title: 'Active Incident',
            impact: 'major',
            status: 'investigating',
            startedAt: '2026-08-28T10:00:00Z',
            resolvedAt: null,
            effectiveDowntimeMinutes: 270,
            affectedComponents: [],
          },
        ],
      }),
      total: 0,
      page: 1,
      totalPages: 1,
      loading: false,
      error: null,
      refetch: vi.fn(),
    })

    const markup = renderComponent()
    expect(markup).toContain('data-testid="availability-active-note"')
  })

  it('requests availability for the selected period', () => {
    useApiMock.mockReturnValue({
      data: createResponse(),
      total: 0,
      page: 1,
      totalPages: 1,
      loading: false,
      error: null,
      refetch: vi.fn(),
    })

    renderComponent()
    expect(useApiMock).toHaveBeenCalledWith('/status/availability', [], { period: '30d' })
  })

  it('renders period selector with all presets', () => {
    useApiMock.mockReturnValue({
      data: createResponse(),
      total: 0,
      page: 1,
      totalPages: 1,
      loading: false,
      error: null,
      refetch: vi.fn(),
    })

    const markup = renderComponent()
    for (const label of ['24h', '7d', '30d', '90d', 'YTD']) {
      expect(markup).toContain(`>${label}<`)
    }
  })
})
