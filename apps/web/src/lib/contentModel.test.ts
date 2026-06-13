import { describe, expect, it } from 'vitest'
import type { Incident, MaintenanceStatus, RichTextDocument } from '../types'
import { normalizeMaintenanceStatus } from './contentModel'

describe('content model contracts', () => {
  it('allows legacy string fields and optional rich json companions', () => {
    const descriptionJson: RichTextDocument = { type: 'doc', content: [] }

    const incident: Incident = {
      id: 'inc-1',
      title: 'API outage',
      description: 'Legacy text',
      descriptionJson,
      status: 'investigating',
      impact: 'major',
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
      affectedComponents: [],
    }

    expect(incident.description).toBe('Legacy text')
    expect(incident.descriptionJson?.type).toBe('doc')
  })

  it('supports new maintenance lifecycle states', () => {
    const statuses: MaintenanceStatus[] = ['draft', 'scheduled', 'active', 'completed']

    expect(statuses).toContain('draft')
    expect(statuses).toContain('active')
  })

  it('normalizes legacy in_progress maintenance state to active behavior', () => {
    expect(normalizeMaintenanceStatus('in_progress')).toBe('active')
    expect(normalizeMaintenanceStatus('scheduled')).toBe('scheduled')
  })
})
