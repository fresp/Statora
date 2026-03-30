import { describe, expect, it } from 'vitest'
import type { CategorySummary } from '../types'
import { getMonitoredServiceCount, getMonitoringViewState, hasMonitoringData } from './statusCategoryMonitoring'

function createSummary(services: CategorySummary['services']): CategorySummary {
  return {
    prefix: 'ulw',
    name: 'ULW',
    description: 'Category',
    aggregateStatus: 'operational',
    uptime90d: 100,
    services,
    incidents: [],
  }
}

describe('statusCategoryMonitoring', () => {
  it('returns false for monitoring when services array is empty', () => {
    const summary = createSummary([])

    expect(hasMonitoringData(summary)).toBe(false)
    expect(getMonitoredServiceCount(summary)).toBe(0)
    expect(getMonitoringViewState(summary)).toEqual({
      hasMonitoring: false,
      monitoredServiceCount: 0,
      totalServiceCount: 0,
      showPartialMonitoringNote: false,
    })
  })

  it('returns false when all services have empty uptime history', () => {
    const summary = createSummary([
      {
        id: 'svc-1',
        name: 'API',
        description: 'API service',
        status: 'operational',
        uptime90d: 99.9,
        uptimeHistory: [],
      },
      {
        id: 'svc-2',
        name: 'Worker',
        description: 'Worker service',
        status: 'operational',
        uptime90d: 99.8,
        uptimeHistory: [],
      },
    ])

    expect(hasMonitoringData(summary)).toBe(false)
    expect(getMonitoredServiceCount(summary)).toBe(0)
    expect(getMonitoringViewState(summary).showPartialMonitoringNote).toBe(false)
  })

  it('returns true and partial note when only some services are monitored', () => {
    const summary = createSummary([
      {
        id: 'svc-1',
        name: 'API',
        description: 'API service',
        status: 'operational',
        uptime90d: 99.9,
        uptimeHistory: [{ date: '2026-03-29', uptimePercent: 100, status: 'operational' }],
      },
      {
        id: 'svc-2',
        name: 'Worker',
        description: 'Worker service',
        status: 'operational',
        uptime90d: 99.8,
        uptimeHistory: [],
      },
    ])

    expect(hasMonitoringData(summary)).toBe(true)
    expect(getMonitoredServiceCount(summary)).toBe(1)
    expect(getMonitoringViewState(summary)).toEqual({
      hasMonitoring: true,
      monitoredServiceCount: 1,
      totalServiceCount: 2,
      showPartialMonitoringNote: true,
    })
  })

  it('returns no partial note when all services are monitored', () => {
    const summary = createSummary([
      {
        id: 'svc-1',
        name: 'API',
        description: 'API service',
        status: 'operational',
        uptime90d: 99.9,
        uptimeHistory: [{ date: '2026-03-29', uptimePercent: 100, status: 'operational' }],
      },
      {
        id: 'svc-2',
        name: 'Worker',
        description: 'Worker service',
        status: 'operational',
        uptime90d: 99.8,
        uptimeHistory: [{ date: '2026-03-29', uptimePercent: 100, status: 'operational' }],
      },
    ])

    expect(getMonitoringViewState(summary)).toEqual({
      hasMonitoring: true,
      monitoredServiceCount: 2,
      totalServiceCount: 2,
      showPartialMonitoringNote: false,
    })
  })
})
