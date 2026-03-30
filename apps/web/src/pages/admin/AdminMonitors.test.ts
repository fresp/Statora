import { describe, expect, it } from 'vitest'

import { monitorLogsPath } from './AdminMonitors'

describe('AdminMonitors', () => {
  describe('monitorLogsPath', () => {
    it('builds monitor logs route from monitor id', () => {
      expect(monitorLogsPath('abc123')).toBe('/admin/monitors/abc123/logs')
    })
  })
})
