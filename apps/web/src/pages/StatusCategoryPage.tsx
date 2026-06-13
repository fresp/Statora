import { useCallback, useMemo } from 'react'
import { Link, useParams } from 'react-router-dom'
import { ArrowLeft, CheckCircle, MessageSquare } from 'lucide-react'
import { useApi, useCategorySummary } from '../hooks/useApi'
import type { CategoryServiceStatus, ComponentStatus, Incident, StatusPageSettings } from '../types'
import { useWebSocket } from '../hooks/useWebSocket'
import { UptimeTimeline } from '../components/status/UptimeTimeline'
import { IncidentTimeline } from '../components/IncidentTimeline'
import { DEFAULT_STATUS_PAGE_SETTINGS, normalizeStatusPageSettings } from '../lib/statusPageSettings'
import { getIncidentContent } from '../lib/contentModel'
import { formatRelativeTime } from '../lib/utils'
import ContentRenderer from '../components/content/ContentRenderer'
import {
  StatusPageBadge,
  StatusPageFrame,
  componentStatusToSeverity,
  incidentImpactToSeverity,
} from '../components/statuspage/StatusPagePrimitives'

const EMPTY_INCIDENTS: Incident[] = []
const EMPTY_SERVICES: CategoryServiceStatus[] = []

function isIncidentActive(status: string): boolean {
  const normalized = status.toLowerCase()
  return normalized !== 'resolved' && normalized !== 'completed' && normalized !== 'closed' && normalized !== 'postmortem'
}

function impactRank(impact: string): number {
  switch (impact.toLowerCase()) {
    case 'critical':
      return 3
    case 'major':
      return 2
    case 'minor':
      return 1
    default:
      return 0
  }
}

function impactToStatus(impact: string): ComponentStatus {
  switch (impact.toLowerCase()) {
    case 'minor':
      return 'degraded_performance'
    case 'major':
      return 'partial_outage'
    case 'critical':
      return 'major_outage'
    default:
      return 'operational'
  }
}

function incidentAffectsService(incident: Incident, service: CategoryServiceStatus): boolean {
  const serviceName = service.name.trim().toLowerCase()

  if (incident.affectedComponentTargets && incident.affectedComponentTargets.length > 0) {
    return incident.affectedComponentTargets.some((target) => {
      const targetName = target.component.name.trim().toLowerCase()
      if (target.component.id === service.id || targetName === serviceName) return true

      if (target.subComponents && target.subComponents.length > 0) {
        return target.subComponents.some((subComponent) => {
          const subComponentName = subComponent.name.trim().toLowerCase()
          return subComponent.id === service.id || subComponentName === serviceName
        })
      }

      return false
    })
  }

  if (incident.affectedComponents.length > 0) {
    return incident.affectedComponents.some((component) => {
      const componentName = component.name.trim().toLowerCase()
      return component.id === service.id || componentName === serviceName
    })
  }

  return false
}

function ServiceHealthCard({ service, incidents }: { service: CategoryServiceStatus; incidents: Incident[] }) {
  const activeIncidents = incidents.filter((incident) => isIncidentActive(incident.status))
  const highestImpact = activeIncidents.reduce<string>((current, incident) => (
    impactRank(incident.impact) > impactRank(current) ? incident.impact : current
  ), '')
  const displayStatus = highestImpact ? impactToStatus(highestImpact) : service.status
  const hasMonitoringData = service.uptimeHistory.length > 0
  const statusAccentClass = (() => {
    switch (componentStatusToSeverity(displayStatus)) {
      case 'critical':
        return 'text-red-700 dark:text-red-400'
      case 'major':
        return 'text-orange-700 dark:text-orange-400'
      case 'minor':
        return 'text-amber-700 dark:text-amber-400'
      case 'operational':
        return 'text-emerald-700 dark:text-emerald-400'
    }
  })()
  const primaryStatusLabel = activeIncidents.length > 0
    ? `${activeIncidents.length} ACTIVE INCIDENT${activeIncidents.length === 1 ? '' : 'S'}`
    : 'NO ACTIVE INCIDENTS'
  const secondaryStatusLabel = activeIncidents.length > 0
    ? 'Incident updates are in progress'
    : 'All systems operating normally'

  return (
    <div className="min-w-0 overflow-hidden flex flex-col justify-between rounded-xl border border-slate-200 bg-white p-6 shadow-sm transition-colors dark:border-slate-800 dark:bg-slate-900">
      <div>
        <div className="mb-6 flex items-start justify-between gap-4">
          <div>
            <span className="text-lg font-semibold">{service.name}</span>
            {service.description && <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">{service.description}</p>}
          </div>
          <StatusPageBadge status={componentStatusToSeverity(displayStatus)} />
        </div>
        {hasMonitoringData && (
          <div className="mb-6 min-w-0 overflow-hidden">
            <UptimeTimeline history={service.uptimeHistory} showLabels={false} />
          </div>
        )}
        {!hasMonitoringData && (
          <div className="mb-6 flex min-h-[96px] flex-col items-center justify-center text-center">
            <p className={`font-semibold uppercase leading-none tracking-[-0.04em] ${statusAccentClass}`}>
              {primaryStatusLabel}
            </p>
            <p className="mt-3 text-sm text-slate-500 dark:text-slate-400">{secondaryStatusLabel}</p>
          </div>
        )}
        {hasMonitoringData && (
          <div className="flex justify-between font-mono text-sm">
            <span className="text-slate-500 dark:text-slate-400">Uptime</span>
            <span className="font-bold">{service.uptime30d?.toFixed(2)}%</span>
          </div>
        )}
      </div>
      <div className="flex justify-between border-t border-slate-100 pt-4 font-mono text-sm dark:border-slate-800/60">
        <span className="text-slate-500 dark:text-slate-400">Last updated</span>
        <span className="font-bold text-slate-700 dark:text-slate-300">{formatRelativeTime(service.updatedAt)}</span>
      </div>
      {hasMonitoringData && activeIncidents.length > 0 && (
        <div className="mt-5 rounded-lg border border-slate-200 bg-slate-50 p-4 dark:border-slate-800 dark:bg-slate-800/50">
          <p className="mb-3 font-mono text-xs font-semibold uppercase tracking-wider text-slate-500 dark:text-slate-400">Active incident context</p>
          <div className="space-y-4">
            {activeIncidents.map((incident) => (
              <div key={incident.id}>
                <div className="mb-2 flex flex-wrap items-center gap-2">
                  <span className="text-sm font-semibold">{incident.title}</span>
                  <StatusPageBadge status={incidentImpactToSeverity(incident.impact)} />
                </div>
                <div className="mb-3 text-sm text-slate-600 dark:text-slate-400">
                  <ContentRenderer text={getIncidentContent(incident).text} json={getIncidentContent(incident).json} />
                </div>
                <IncidentTimeline updates={incident.updates || []} />
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}

export default function StatusCategoryPage() {
  const { categoryPrefix } = useParams<{ categoryPrefix: string }>()
  const { data, loading, error, refetch } = useCategorySummary(categoryPrefix)
  const { data: settingsData } = useApi<StatusPageSettings>('/status/settings')

  const settings = normalizeStatusPageSettings(settingsData ?? DEFAULT_STATUS_PAGE_SETTINGS)
  const incidents = data?.incidents ?? EMPTY_INCIDENTS
  const services = data?.services ?? EMPTY_SERVICES
  const aggregateStatus: ComponentStatus = data?.aggregateStatus ?? 'operational'
  const categoryHasMonitoring = data?.hasMonitoring ?? false

  const incidentsByService = useMemo(() => {
    const serviceIncidentMap = new Map<string, Incident[]>()
    for (const service of services) {
      serviceIncidentMap.set(service.id, incidents.filter((incident) => incidentAffectsService(incident, service)))
    }
    return serviceIncidentMap
  }, [incidents, services])

  const handleWsMessage = useCallback((event: { type: string; data: unknown }) => {
    if (['component_updated', 'component_created', 'incident_created', 'incident_updated', 'incident_resolved', 'incident_update_added'].includes(event.type)) {
      refetch()
    }
  }, [refetch])

  useWebSocket(handleWsMessage)

  if (loading) {
    return (
      <StatusPageFrame settings={settings}>
        <main className="mx-auto max-w-[768px] px-4 py-10 md:px-8">
          <div className="rounded-xl border border-slate-200 bg-white p-6 font-mono text-sm text-slate-500 shadow-sm dark:border-slate-800 dark:bg-slate-900 dark:text-slate-400">Loading service health...</div>
        </main>
      </StatusPageFrame>
    )
  }

  if (error || !data) {
    return (
      <StatusPageFrame settings={settings}>
        <main className="mx-auto max-w-[768px] space-y-6 px-4 py-10 md:px-8">
          <Link to="/" className="inline-flex items-center gap-1.5 font-mono text-sm text-slate-600 transition-colors hover:text-emerald-700 dark:text-slate-400 dark:hover:text-emerald-400">
            <ArrowLeft className="h-4 w-4" /> Global Status
          </Link>
          <div className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm dark:border-slate-800 dark:bg-slate-900">
            <h1 className="text-xl font-semibold">Category unavailable</h1>
            <p className="mt-2 text-sm text-slate-500 dark:text-slate-400">{error ?? `Unable to load ${categoryPrefix ?? 'this category'} right now.`}</p>
          </div>
        </main>
      </StatusPageFrame>
    )
  }

  return (
    <StatusPageFrame settings={settings}>
      <main className="mx-auto max-w-[768px] px-4 py-8 md:px-8 md:py-12">
        <div className="mb-6">
          <Link to="/" className="inline-flex items-center gap-1.5 font-mono text-sm text-slate-600 transition-colors hover:text-emerald-700 dark:text-slate-400 dark:hover:text-emerald-400">
            <ArrowLeft className="h-4 w-4" /> Global Status
          </Link>
        </div>

        <div className="mb-10 rounded-xl border border-slate-200 bg-white p-6 shadow-sm transition-colors dark:border-slate-800 dark:bg-slate-900 md:p-8">
          <div className="flex flex-col items-start justify-between gap-6 md:flex-row md:items-center">
            <div>
              <div className="mb-2 flex items-center gap-3">
                <MessageSquare className="h-8 w-8 text-emerald-700 dark:text-emerald-400" />
                <h1 className="text-3xl font-bold">{data.name}</h1>
              </div>
              <p className="text-lg text-slate-600 dark:text-slate-400">{data.description || 'Real-time service health and operational context.'}</p>
            </div>
            <div className="flex w-full flex-col items-start md:w-auto md:items-end">
              <div className="mb-2">
                <StatusPageBadge status={componentStatusToSeverity(aggregateStatus)} />
              </div>
              <div className="mt-2 text-left md:text-right">
                {data?.uptime30d != null ? (
                  <>
                    <span className="text-5xl font-bold tracking-tight text-slate-900 dark:text-white">
                      {data.uptime30d.toFixed(2)}<span className="text-2xl font-normal text-slate-400">%</span>
                    </span>
                    <p className="mt-1 font-mono text-xs uppercase tracking-wider text-slate-500 dark:text-slate-400">30-Day Uptime</p>
                  </>
                ) : (
                  <>
                    <p className="text-2xl font-bold uppercase leading-none tracking-tight text-slate-900 dark:text-white">Incident</p>
                    <p className="mt-1 font-mono text-xs uppercase tracking-wider text-slate-500 dark:text-slate-400">Tracking</p>
                  </>
                )}
              </div>
            </div>
          </div>
        </div>

        <section className="mb-10">
          <div className="mb-4 flex items-center justify-between">
            <h2 className="text-xl font-semibold">Sub-service Health</h2>
            <span className="font-mono text-xs text-slate-500">Live updates enabled</span>
          </div>
          {services.length > 0 ? (
            <div className="grid grid-cols-1 gap-6 md:grid-cols-2">
              {services.map((service) => (
                <ServiceHealthCard key={service.id} service={service} incidents={incidentsByService.get(service.id) ?? EMPTY_INCIDENTS} />
              ))}
            </div>
          ) : (
            <div className="rounded-xl border border-slate-200 bg-white p-6 text-sm text-slate-500 shadow-sm dark:border-slate-800 dark:bg-slate-900 dark:text-slate-400">No services are configured for this category yet.</div>
          )}
        </section>

        {/* <section className="mb-10">
          <div className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm transition-colors dark:border-slate-800 dark:bg-slate-900">
            <div className="mb-8 flex flex-col items-start justify-between gap-4 md:flex-row md:items-center">
              <div>
                <h2 className="text-xl font-semibold">Latency Trend</h2>
                <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">Average response time over the last 24 hours.</p>
              </div>
              <div className="flex rounded-lg bg-slate-100 p-1 dark:bg-slate-800">
                <button className="rounded-md bg-white px-4 py-1.5 font-mono text-xs font-semibold shadow-sm dark:bg-slate-700">24h</button>
                <button className="px-4 py-1.5 font-mono text-xs font-semibold text-slate-500 transition-colors hover:text-slate-900 dark:text-slate-400 dark:hover:text-white">7d</button>
              </div>
            </div>
            <div className="flex h-64 items-center justify-center rounded-xl border border-dashed border-slate-200 bg-slate-50 text-center dark:border-slate-800 dark:bg-slate-800/40">
              <div>
                <CheckCircle className="mx-auto mb-3 h-8 w-8 text-emerald-600 dark:text-emerald-400" />
                <p className="font-mono text-sm font-semibold text-slate-700 dark:text-slate-300">Latency metrics are not exposed by the current public API.</p>
                <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">Service uptime and incident context remain live.</p>
              </div>
            </div>
          </div>
        </section> */}
      </main>
    </StatusPageFrame>
  )
}
