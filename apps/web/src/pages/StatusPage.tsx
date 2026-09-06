import { useState, useCallback, useEffect, FormEvent } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { ArrowRight, Calendar, CheckCircle, XCircle } from 'lucide-react'
import { useApi } from '../hooks/useApi'
import { useWebSocket } from '../hooks/useWebSocket'
import type { ComponentStatus, ComponentWithSubs, Incident, Maintenance, StatusPageSettings, StatusSummary } from '../types'
import { formatDate } from '../lib/utils'
import { getIncidentContent, getMaintenanceContent } from '../lib/contentModel'
import { getApiErrorMessage } from '../lib/apiError'
import api from '../lib/api'
import Modal from '../components/Modal'
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

  const [searchParams, setSearchParams] = useSearchParams()
  const [subscribeOpen, setSubscribeOpen] = useState(false)
  const [subscribeEmail, setSubscribeEmail] = useState('')
  const [subscribeSubmitting, setSubscribeSubmitting] = useState(false)
  const [subscribeSuccess, setSubscribeSuccess] = useState<string | null>(null)
  const [subscribeError, setSubscribeError] = useState<string | null>(null)

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
    if (settingsData) {
      const normalized = normalizeStatusPageSettings(settingsData)
      setSettings(normalized)
      cacheStatusPageSettings(normalized)
    }
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

  const verifiedBanner = searchParams.get('verified') === 'true'
  const unsubscribedBanner = searchParams.get('unsubscribed') === 'true'
  const bannerError = searchParams.get('error')

  function dismissBanners() {
    const next = new URLSearchParams(searchParams)
    next.delete('verified')
    next.delete('unsubscribed')
    next.delete('error')
    setSearchParams(next, { replace: true })
  }

  function openSubscribe() {
    setSubscribeSuccess(null)
    setSubscribeError(null)
    setSubscribeEmail('')
    setSubscribeOpen(true)
  }

  async function handleSubscribeSubmit(e: FormEvent) {
    e.preventDefault()
    const email = subscribeEmail.trim()
    if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(email)) {
      setSubscribeError('Please enter a valid email address.')
      return
    }
    setSubscribeSubmitting(true)
    setSubscribeError(null)
    try {
      await api.post('/subscribe', { email })
      setSubscribeSuccess("Check your inbox! We've sent a verification link to confirm your subscription.")
      setSubscribeEmail('')
    } catch (err: unknown) {
      setSubscribeError(getApiErrorMessage(err, 'Failed to subscribe. Please try again.'))
    } finally {
      setSubscribeSubmitting(false)
    }
  }

  return (
    <StatusPageFrame settings={settings} onSubscribeClick={openSubscribe}>
      {(verifiedBanner || unsubscribedBanner || bannerError) && (
        <div className="mx-auto flex max-w-[768px] items-center justify-between gap-3 px-4 pt-6 md:px-8">
          {verifiedBanner && (
            <div className="flex flex-1 items-center gap-2 rounded-lg border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-800 dark:border-emerald-800 dark:bg-emerald-900/30 dark:text-emerald-300">
              <CheckCircle className="h-4 w-4 shrink-0" />
              Your subscription is confirmed. You will now receive status updates.
            </div>
          )}
          {unsubscribedBanner && (
            <div className="flex flex-1 items-center gap-2 rounded-lg border border-slate-200 bg-slate-100 px-4 py-3 text-sm text-slate-700 dark:border-slate-700 dark:bg-slate-800/60 dark:text-slate-300">
              <XCircle className="h-4 w-4 shrink-0" />
              You have been unsubscribed. You will no longer receive status updates.
            </div>
          )}
          {bannerError && (
            <div className="flex flex-1 items-center gap-2 rounded-lg border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-800 dark:border-amber-800 dark:bg-amber-900/30 dark:text-amber-300">
              <XCircle className="h-4 w-4 shrink-0" />
              {bannerError === 'invalid_or_expired_token'
                ? 'That verification link is invalid or has expired. Please subscribe again.'
                : 'That link is invalid. Please check the link or subscribe again.'}
            </div>
          )}
          <button
            type="button"
            onClick={dismissBanners}
            className="rounded-md px-2 py-1 text-xs font-medium text-slate-500 hover:bg-slate-100 dark:text-slate-400 dark:hover:bg-slate-800"
          >
            Dismiss
          </button>
        </div>
      )}

      {subscribeOpen && (
        <Modal
          title="Subscribe to updates"
          onClose={() => setSubscribeOpen(false)}
          footer={
            <div className="flex justify-end gap-2">
              <button
                type="button"
                onClick={() => setSubscribeOpen(false)}
                className="rounded-lg border border-slate-200 px-4 py-2 text-sm font-medium text-slate-600 hover:bg-slate-50 dark:border-slate-700 dark:text-slate-300 dark:hover:bg-slate-800"
              >
                Close
              </button>
              <button
                type="submit"
                form="subscribe-form"
                disabled={subscribeSubmitting}
                className="rounded-lg bg-emerald-700 px-4 py-2 text-sm font-medium text-white hover:bg-emerald-800 disabled:opacity-60 dark:bg-emerald-600 dark:hover:bg-emerald-700"
              >
                {subscribeSubmitting ? 'Subscribing...' : 'Subscribe'}
              </button>
            </div>
          }
        >
          <form id="subscribe-form" onSubmit={handleSubscribeSubmit} className="space-y-4">
            <p className="text-sm text-slate-600 dark:text-slate-400">
              Get email notifications when we post incidents or maintenance updates.
            </p>
            <input
              type="email"
              required
              placeholder="you@example.com"
              className="w-full rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm text-slate-900 focus:outline-none focus:ring-2 focus:ring-emerald-500 dark:border-slate-700 dark:bg-slate-800 dark:text-slate-100"
              value={subscribeEmail}
              onChange={(e) => setSubscribeEmail(e.target.value)}
            />
            {subscribeSuccess && (
              <p className="flex items-center gap-2 rounded-lg bg-emerald-50 px-3 py-2 text-sm text-emerald-800 dark:bg-emerald-900/30 dark:text-emerald-300">
                <CheckCircle className="h-4 w-4 shrink-0" />
                {subscribeSuccess}
              </p>
            )}
            {subscribeError && (
              <p className="rounded-lg bg-red-50 px-3 py-2 text-sm text-red-700 dark:bg-red-900/30 dark:text-red-400">
                {subscribeError}
              </p>
            )}
          </form>
        </Modal>
      )}

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
