import { useState, useCallback, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { AlertTriangle, ArrowRight, Calendar, CheckCircle } from 'lucide-react'
import { useApi } from '../hooks/useApi'
import { useWebSocket } from '../hooks/useWebSocket'
import type { ComponentStatus, ComponentWithSubs, Incident, Maintenance, StatusPageSettings, StatusSummary } from '../types'
import { formatDate, getOverallStatusLabel } from '../lib/utils'
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
import { UptimeTimeline } from '../components/status/UptimeTimeline'
import {
  StatusPageBadge,
  StatusPageFrame,
  componentStatusToSeverity,
  incidentImpactToSeverity,
} from '../components/statuspage/StatusPagePrimitives'

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
      <main className="mx-auto flex max-w-[1024px] flex-col gap-8 px-4 py-8 md:gap-10 md:px-8 md:py-12">
        <section className={`flex flex-col items-start justify-between gap-6 rounded-xl p-6 shadow-sm transition-colors md:flex-row md:items-center md:p-8 ${hasActiveIncidents
          ? 'border border-amber-200 bg-amber-50 text-amber-900 dark:border-amber-800/50 dark:bg-amber-900/20 dark:text-amber-100'
          : 'bg-emerald-600 text-white'
          }`}>
          <div className="flex items-start gap-4 md:items-center">
            {hasActiveIncidents ? (
              <AlertTriangle className="mt-1 h-8 w-8 shrink-0 md:mt-0 md:h-10 md:w-10" />
            ) : (
              <CheckCircle className="mt-1 h-8 w-8 shrink-0 md:mt-0 md:h-10 md:w-10" />
            )}
            <div>
              <h1 className="text-2xl font-bold tracking-tight md:text-3xl">{getOverallStatusLabel(overallStatus)}</h1>
              <p className="mt-1 max-w-2xl text-base opacity-90">
                {hasActiveIncidents
                  ? `${activeIncidents.length || summary?.activeIncidents || 1} active incident${(activeIncidents.length || summary?.activeIncidents || 1) === 1 ? '' : 's'} currently being tracked.`
                  : `${settings.branding.siteName || 'Statora'} is operating normally. No active incidents detected.`}
              </p>
            </div>
          </div>
          <div className="grid w-full grid-cols-3 gap-3 text-left font-mono text-xs uppercase tracking-wider opacity-90 md:w-auto md:min-w-72 md:text-right">
            <div>
              <div className="text-2xl font-bold normal-case tracking-tight">{components?.length ?? 0}</div>
              <div>Groups</div>
            </div>
            <div>
              <div className="text-2xl font-bold normal-case tracking-tight">{totalServices}</div>
              <div>Services</div>
            </div>
            <div>
              <div className="text-2xl font-bold normal-case tracking-tight">{summary?.scheduledMaintenance ?? upcomingMaintenance.length}</div>
              <div>Maint.</div>
            </div>
          </div>
        </section>

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
                          <StatusPageBadge status={componentStatusToSeverity(serviceStatus)} />
                        </div>
                      </div>
                      {component.uptimeHistory.length > 0 && (
                        <>
                          <div className="mb-2">
                            <UptimeTimeline history={component.uptimeHistory} showLabels={false} />
                          </div>
                          <div className="flex justify-between font-mono text-xs text-slate-500 dark:text-slate-400">
                            <span>90 days ago</span>
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
