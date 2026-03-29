import type { CategorySummary } from '../types'

export function hasMonitoringData(data: CategorySummary): boolean {
  return data.services.some((service) => service.uptimeHistory.length > 0)
}

export function getMonitoredServiceCount(data: CategorySummary): number {
  return data.services.filter((service) => service.uptimeHistory.length > 0).length
}

export interface MonitoringViewState {
  hasMonitoring: boolean
  monitoredServiceCount: number
  totalServiceCount: number
  showPartialMonitoringNote: boolean
}

export function getMonitoringViewState(data: CategorySummary): MonitoringViewState {
  const totalServiceCount = data.services.length
  const monitoredServiceCount = getMonitoredServiceCount(data)
  const hasMonitoring = monitoredServiceCount > 0

  return {
    hasMonitoring,
    monitoredServiceCount,
    totalServiceCount,
    showPartialMonitoringNote: hasMonitoring && monitoredServiceCount < totalServiceCount,
  }
}
