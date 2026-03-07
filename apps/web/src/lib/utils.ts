import { clsx, type ClassValue } from 'clsx'
import type { ComponentStatus } from '../types'

export function cn(...inputs: ClassValue[]) {
  return clsx(inputs)
}

export const STATUS_LABELS: Record<ComponentStatus, string> = {
  operational: 'Operational',
  degraded_performance: 'Degraded Performance',
  partial_outage: 'Partial Outage',
  major_outage: 'Major Outage',
  maintenance: 'Under Maintenance',
}

export const STATUS_COLORS: Record<ComponentStatus, string> = {
  operational: 'bg-green-500',
  degraded_performance: 'bg-yellow-400',
  partial_outage: 'bg-orange-500',
  major_outage: 'bg-red-500',
  maintenance: 'bg-blue-500',
}

export const STATUS_TEXT_COLORS: Record<ComponentStatus, string> = {
  operational: 'text-green-600',
  degraded_performance: 'text-yellow-600',
  partial_outage: 'text-orange-600',
  major_outage: 'text-red-600',
  maintenance: 'text-blue-600',
}

export const STATUS_BG_LIGHT: Record<ComponentStatus, string> = {
  operational: 'bg-green-50 border-green-200',
  degraded_performance: 'bg-yellow-50 border-yellow-200',
  partial_outage: 'bg-orange-50 border-orange-200',
  major_outage: 'bg-red-50 border-red-200',
  maintenance: 'bg-blue-50 border-blue-200',
}

export const INCIDENT_STATUS_LABELS: Record<string, string> = {
  investigating: 'Investigating',
  identified: 'Identified',
  monitoring: 'Monitoring',
  resolved: 'Resolved',
}

export const INCIDENT_IMPACT_LABELS: Record<string, string> = {
  none: 'None',
  minor: 'Minor',
  major: 'Major',
  critical: 'Critical',
}

export function getOverallStatusLabel(status: ComponentStatus): string {
  const labels: Record<ComponentStatus, string> = {
    operational: 'All Systems Operational',
    degraded_performance: 'Degraded System Performance',
    partial_outage: 'Partial System Outage',
    major_outage: 'Major System Outage',
    maintenance: 'System Under Maintenance',
  }
  return labels[status] || 'Unknown Status'
}

export function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

export function formatDateShort(dateStr: string): string {
  return new Date(dateStr).toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
  })
}
