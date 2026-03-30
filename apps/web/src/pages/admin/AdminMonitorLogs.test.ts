import { describe, expect, it } from 'vitest'

import { latencyBarWidth, statusBadgeClass } from './AdminMonitorLogs'

describe('AdminMonitorLogs', () => {
  describe('statusBadgeClass', () => {
    it('returns green class for up status', () => {
      expect(statusBadgeClass('up')).toContain('text-green-700')
    })

    it('returns red class for down status', () => {
      expect(statusBadgeClass('down')).toContain('text-red-700')
    })

    it('returns gray class for unknown status', () => {
      expect(statusBadgeClass('pending')).toContain('text-gray-700')
    })
  })

  describe('latencyBarWidth', () => {
    it('returns percentage for normal latency', () => {
      expect(latencyBarWidth(1000)).toBe(50)
    })

    it('clamps below zero to zero', () => {
      expect(latencyBarWidth(-10)).toBe(0)
    })

    it('clamps above max to one hundred', () => {
      expect(latencyBarWidth(9999)).toBe(100)
    })
  })
})
