import { useState, useCallback, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { ArrowRight, Calendar } from 'lucide-react'
import { useApi } from '../hooks/useApi'
import { useWebSocket } from '../hooks/useWebSocket'
import type { ComponentStatus, ComponentWithSubs, Incident, Maintenance, StatusPageSettings, StatusSummary } from '../types'
import { formatDate } from '../lib/utils'
import { getIncidentContent, getMaintenanceContent } from '../lib/contentModel'
import ContentRenderer from '../components/content/ContentRenderer'
import {
  DEFAULT_STATUS_PAGE_SETTINGS,
  applyStatusPageHeadSettings,
  applyStatusPageThemePreset,
  cacheStatusPageSettings,
  getBootstrappedStatusPageSettings,
  normalizeStatusPageSettings,
  parseStatusPageSettingsPayload,
  readCachedStatusPageSettings,
} from '../lib/statusPageSettings'
import { AvailabilityMetrics } from '../components/status/AvailabilityMetrics'
import { UptimeTimeline } from '../components/status/UptimeTimeline'
import {
  StatusPageBadge,
  StatusPageFrame,
  componentStatusToSeverity,
  incidentImpactToSeverity,
} from '../components/statuspage/StatusPagePrimitives'
import StatusHero from '../components/statuspage/StatusHero'

function toCategoryPrefix(name: string): string {
  return name
    .toLowerCase()
    .trim()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
}

function getServiceUptime(component: ComponentWithSubs): number {
  if (component.uptimeHistory.length === 0) return component.status === 'operational' ? 100 : 0
  return component.uptimeHistory.reduce((sum, day) => sum + day.uptimePercent, 0) / component.uptimeHistory.length
}

function hasIncidentSeverity(status: ComponentStatus): boolean {
  return status !== 'operational'
}

export default function StatusPage() {
  const { data: summary, refetch: refetchSummary } = useApi<StatusSummary>('/status/summary')
  const { data: components, refetch: refetchComponents } = useApi<ComponentWithSubs[]>('/status/components')
  const { data: incidentData, refetch: refetchIncidents } = useApi<{ active: Incident[]; resolved: Incident[] }>('/status/incidents')
  const { data: settingsData } = useApi<StatusPageSettings>('/status/settings')
  const { data: maintenanceData } = useApi<Maintenance[]>('/status/maintenance')

  const [settings, setSettings] = useState<StatusPageSettings>(() => (
    getBootstrappedStatusPageSettings()
    ?? readCachedStatusPageSettings()
    ?? DEFAULT_STATUS_PAGE_SETTINGS
  ))

  const handleWsMessage = useCallback((event: { type: string; data: unknown }) => {
    if (['component_updated', 'component_created'].includes(event.type)) {
      refetchComponents()
      refetchSummary()
    }
    if (['incident_created', 'incident_updated', 'incident_resolved', 'incident_update_added'].includes(event.type)) {
      refetchIncidents()
      refetchSummary()
    }
    if (event.type === 'status_page_settings_updated') {
      const nextSettings = parseStatusPageSettingsPayload(event.data)
      if (nextSettings) {
        setSettings(nextSettings)
        cacheStatusPageSettings(nextSettings)
      }
    }
  }, [refetchComponents, refetchSummary, refetchIncidents])

  useWebSocket(handleWsMessage)

  useEffect(() => {
    if (!settingsData) return
    const nextSettings = normalizeStatusPageSettings(settingsData)
    setSettings(nextSettings)
    cacheStatusPageSettings(nextSettings)
  }, [settingsData])

  useEffect(() => {
    applyStatusPageHeadSettings(settings)
  }, [settings])

  useEffect(() => {
    applyStatusPageThemePreset(settings)
  }, [settings])

  const overallStatus = summary?.overallStatus ?? 'operational'
  const activeIncidents = incidentData?.active ?? []
  const upcomingMaintenance = maintenanceData?.filter((maintenance) => maintenance.status !== 'completed') ?? []
  const hasActiveIncidents = activeIncidents.length > 0 || hasIncidentSeverity(overallStatus)
  const totalServices = components?.reduce((count, component) => count + Math.max(component.subComponents.length, 1), 0) ?? 0

  return (
    <StatusPageFrame settings={settings}>
      <main className="mx-auto flex max-w-[768px] flex-col gap-8 px-4 py-8 md:gap-10 md:px-8 md:py-12">
        <StatusHero
          overallStatus={overallStatus}
          activeIncidentCount={hasActiveIncidents ? (activeIncidents.length || summary?.activeIncidents || 1) : 0}
          groupCount={components?.length ?? 0}
          serviceCount={totalServices}
          maintenanceCount={summary?.scheduledMaintenance ?? upcomingMaintenance.length}
        />

        <AvailabilityMetrics />

        {upcomingMaintenance.map((maintenance) => (
          <section key={maintenance.id} className="flex flex-col items-start gap-5 rounded-xl border border-slate-200 bg-slate-100 p-6 transition-colors dark:border-slate-800 dark:bg-slate-800/50 md:flex-row">
            <div className="rounded-lg bg-slate-200 p-3 text-slate-700 dark:bg-slate-800 dark:text-slate-300">
              <Calendar className="h-6 w-6" />
            </div>
            <div className="flex-1">
              <h2 className="mb-2 text-xl font-semibold">Scheduled Maintenance: {maintenance.title}</h2>
              <div className="mb-4 text-base text-slate-600 dark:text-slate-400">
                <ContentRenderer text={getMaintenanceContent(maintenance).text} json={getMaintenanceContent(maintenance).json} />
              </div>
              <div className="flex flex-wrap items-center gap-2 font-mono text-xs">
                <span className="rounded-md border border-slate-200 bg-white px-3 py-1.5 shadow-sm dark:border-slate-700 dark:bg-slate-900">{formatDate(maintenance.startTime)}</span>
                <ArrowRight className="h-4 w-4 text-slate-400" />
                <span className="rounded-md border border-slate-200 bg-white px-3 py-1.5 shadow-sm dark:border-slate-700 dark:bg-slate-900">{formatDate(maintenance.endTime)}</span>
              </div>
            </div>
          </section>
        ))}

        {activeIncidents.length > 0 && (
          <section className="space-y-4 rounded-xl border border-slate-200 bg-white p-6 shadow-sm transition-colors dark:border-slate-800 dark:bg-slate-900">
            <div className="flex flex-col gap-2 md:flex-row md:items-start md:justify-between">
              <div>
                <h2 className="text-xl font-semibold">Active Incidents</h2>
                <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">Live incident updates are shown without leaving the status overview.</p>
              </div>
              <StatusPageBadge status={incidentImpactToSeverity(activeIncidents[0]?.impact ?? 'minor')} />
            </div>
            <div className="divide-y divide-slate-100 dark:divide-slate-800/60">
              {activeIncidents.map((incident) => (
                <Link key={incident.id} to="/history" className="block py-5 transition-colors hover:bg-slate-50 dark:hover:bg-slate-800/30">
                  <div className="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
                    <div>
                      <h3 className="text-lg font-semibold">{incident.title}</h3>
                      <div className="mt-1 text-sm leading-6 text-slate-600 dark:text-slate-400">
                        <ContentRenderer text={getIncidentContent(incident).text} json={getIncidentContent(incident).json} />
                      </div>
                      <p className="mt-2 font-mono text-xs uppercase tracking-wider text-slate-500 dark:text-slate-400">Started: {formatDate(incident.createdAt)}</p>
                    </div>
                    <StatusPageBadge status={incidentImpactToSeverity(incident.impact)} />
                  </div>
                </Link>
              ))}
            </div>
          </section>
        )}

        <section className="flex flex-col gap-6">
          {(components ?? []).map((component) => (
            <div key={component.id} className="overflow-hidden rounded-xl border border-slate-200 bg-white shadow-sm transition-colors dark:border-slate-800 dark:bg-slate-900">
              <Link to={`/status/${toCategoryPrefix(component.name)}`} className="flex items-center justify-between border-b border-slate-200 bg-slate-50 px-6 py-4 transition-colors hover:bg-slate-100 dark:border-slate-800 dark:bg-slate-800/50 dark:hover:bg-slate-800">
                <div>
                  <h3 className="text-xl font-semibold">{component.name}</h3>
                  {component.description && <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">{component.description}</p>}
                </div>
                <StatusPageBadge status={componentStatusToSeverity(component.status)} />
              </Link>

              <div className="divide-y divide-slate-100 dark:divide-slate-800/60">
                {(component.subComponents.length > 0 ? component.subComponents : [component]).map((service) => {
                  const serviceStatus = service.status
                  const serviceUptime = 'uptimeHistory' in service ? getServiceUptime(service) : getServiceUptime(component)
                  return (
                    <Link to={`/status/${toCategoryPrefix(component.name)}`} key={service.id} className="block px-6 py-5 transition-colors hover:bg-slate-50 dark:hover:bg-slate-800/30">
                      <div className="mb-4 flex flex-col justify-between gap-3 md:flex-row md:items-center">
                        <div>
                          <h4 className="text-lg font-semibold">{service.name}</h4>
                          {service.description && <p className="mt-0.5 text-sm text-slate-500 dark:text-slate-400">{service.description}</p>}
                        </div>
                        <div className="self-start md:self-center">
                          {serviceStatus !== 'operational' ? <StatusPageBadge status={componentStatusToSeverity(serviceStatus)} /> : "No known issues"}
                        </div>
                      </div>
                      {component.uptimeHistory.length > 0 && (
                        <>
                          <div className="mb-2">
                            <UptimeTimeline history={component.uptimeHistory} showLabels={false} />
                          </div>
                          <div className="flex justify-between font-mono text-xs text-slate-500 dark:text-slate-400">
                            <span>Last 30 Days</span>
                            <span className={`font-semibold ${serviceUptime === 100 ? 'text-emerald-700 dark:text-emerald-400' : 'text-slate-900 dark:text-slate-300'}`}>
                              {serviceUptime.toFixed(2)}% Uptime
                            </span>
                            <span>Today</span>
                          </div>
                        </>
                      )}
                    </Link>
                  )
                })}
              </div>
            </div>
          ))}
        </section>
      </main>
    </StatusPageFrame>
  )
}
