import { useState } from 'react'
import { Activity } from 'lucide-react'
import { useApi } from '../../hooks/useApi'
import type { AvailabilityResponse } from '../../types'
import { StatusPageBadge, incidentImpactToSeverity } from '../statuspage/StatusPagePrimitives'

type PeriodPreset = '24h' | '7d' | '30d' | '90d' | 'ytd'

const PERIOD_PRESETS: Array<{ value: PeriodPreset; label: string }> = [
  { value: '24h', label: '24h' },
  { value: '7d', label: '7d' },
  { value: '30d', label: '30d' },
  { value: '90d', label: '90d' },
  { value: 'ytd', label: 'YTD' },
]

function formatDuration(minutes: number): string {
  if (minutes < 1) return '<1 min'
  const wholeMinutes = Math.round(minutes)
  const days = Math.floor(wholeMinutes / 1440)
  const hours = Math.floor((wholeMinutes % 1440) / 60)
  const mins = wholeMinutes % 60
  const parts: string[] = []
  if (days > 0) parts.push(`${days}d`)
  if (hours > 0) parts.push(`${hours}h`)
  if (mins > 0) parts.push(`${mins}m`)
  return parts.length > 0 ? parts.join(' ') : '0m'
}

function formatPercent(value: number): string {
  return `${value.toFixed(2)}%`
}

function formatDateTime(iso: string): string {
  return new Date(iso).toLocaleString(undefined, {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}

export function AvailabilityMetrics() {
  const [period, setPeriod] = useState<PeriodPreset>('30d')
  const [customRange, setCustomRange] = useState(false)
  const [customFrom, setCustomFrom] = useState('')
  const [customTo, setCustomTo] = useState('')

  const apiParams = customRange && customFrom && customTo
    ? { from: customFrom, to: customTo }
    : { period }

  const { data, loading, error, refetch } = useApi<AvailabilityResponse>('/status/availability', [], apiParams)

  const activeIncidentCount = (data?.incidents ?? []).filter((incident) => incident.resolvedAt === null).length

  return (
    <section
      data-testid="availability-metrics"
      className="space-y-5 rounded-xl border border-slate-200 bg-white p-6 shadow-sm transition-colors dark:border-slate-800 dark:bg-slate-900"
    >
      <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
        <div className="flex items-center gap-2">
          <Activity className="h-5 w-5 text-slate-500 dark:text-slate-400" />
          <h2 className="text-xl font-semibold">Availability</h2>
        </div>
        <div className="flex flex-wrap items-center gap-1 rounded-lg border border-slate-200 bg-slate-100 p-1 dark:border-slate-700 dark:bg-slate-800">
          {PERIOD_PRESETS.map((preset) => (
            <button
              key={preset.value}
              type="button"
              aria-pressed={!customRange && period === preset.value}
              onClick={() => { setCustomRange(false); setPeriod(preset.value) }}
              className={`rounded-md px-3 py-1.5 text-sm font-medium transition-colors ${
                !customRange && period === preset.value
                  ? 'bg-white text-slate-900 shadow-sm dark:bg-slate-900 dark:text-slate-100'
                  : 'text-slate-500 hover:text-slate-900 dark:text-slate-400 dark:hover:text-slate-100'
              }`}
            >
              {preset.label}
            </button>
          ))}
          <button
            type="button"
            aria-pressed={customRange}
            onClick={() => setCustomRange(true)}
            className={`rounded-md px-3 py-1.5 text-sm font-medium transition-colors ${
              customRange
                ? 'bg-white text-slate-900 shadow-sm dark:bg-slate-900 dark:text-slate-100'
                : 'text-slate-500 hover:text-slate-900 dark:text-slate-400 dark:hover:text-slate-100'
            }`}
          >
            Custom
          </button>
        </div>
      </div>

      {customRange && (
        <div className="flex flex-wrap items-end gap-3" data-testid="availability-custom-range">
          <label className="flex flex-col gap-1">
            <span className="text-xs font-medium text-slate-500 dark:text-slate-400">From</span>
            <input
              type="date"
              value={customFrom}
              onChange={(e) => setCustomFrom(e.target.value)}
              className="rounded-md border border-slate-300 bg-white px-3 py-1.5 text-sm dark:border-slate-600 dark:bg-slate-800 dark:text-slate-100"
            />
          </label>
          <label className="flex flex-col gap-1">
            <span className="text-xs font-medium text-slate-500 dark:text-slate-400">To</span>
            <input
              type="date"
              value={customTo}
              onChange={(e) => setCustomTo(e.target.value)}
              className="rounded-md border border-slate-300 bg-white px-3 py-1.5 text-sm dark:border-slate-600 dark:bg-slate-800 dark:text-slate-100"
            />
          </label>
          <button
            type="button"
            onClick={() => refetch()}
            disabled={!customFrom || !customTo}
            className="rounded-md border border-slate-300 bg-white px-3 py-1.5 text-sm font-medium text-slate-700 transition-colors hover:bg-slate-50 disabled:opacity-50 dark:border-slate-600 dark:bg-slate-800 dark:text-slate-300 dark:hover:bg-slate-700"
          >
            Apply
          </button>
        </div>
      )}

      {loading && (
        <div className="space-y-4" data-testid="availability-loading">
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
            {[0, 1, 2].map((index) => (
              <div key={index} className="h-20 animate-pulse rounded-lg bg-slate-100 dark:bg-slate-800" />
            ))}
          </div>
          <div className="h-24 animate-pulse rounded-lg bg-slate-100 dark:bg-slate-800" />
        </div>
      )}

      {!loading && error && (
        <div className="flex flex-col items-start gap-2 rounded-lg border border-red-200 bg-red-50 p-4 dark:border-red-900/50 dark:bg-red-950/30" data-testid="availability-error">
          <p className="text-sm text-red-700 dark:text-red-300">Failed to load availability metrics: {error}</p>
          <button
            type="button"
            onClick={() => refetch()}
            className="rounded-md border border-red-300 bg-white px-3 py-1.5 text-sm font-medium text-red-700 transition-colors hover:bg-red-50 dark:border-red-800 dark:bg-slate-900 dark:text-red-300 dark:hover:bg-red-950/50"
          >
            Retry
          </button>
        </div>
      )}

      {!loading && !error && data && (
        <div className="space-y-5">
          <p className="text-xs font-mono uppercase tracking-wider text-slate-500 dark:text-slate-400">
            {data.period.label} &middot; {formatDateTime(data.period.start)} &rarr; {formatDateTime(data.period.end)}
          </p>

          <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
            <div className="rounded-lg border border-slate-200 p-4 dark:border-slate-700" data-testid="availability-percent">
              <p className={`text-2xl font-semibold ${data.overall.availability >= 99.9 ? 'text-emerald-700 dark:text-emerald-400' : 'text-slate-900 dark:text-slate-100'}`}>
                {data.overall.incidentCount === 0 ? '100%' : formatPercent(data.overall.availability)}
              </p>
              <p className="mt-1 text-xs uppercase tracking-wide text-slate-500 dark:text-slate-400">
                {data.overall.incidentCount === 0 ? 'No incidents in this period' : 'Uptime'}
              </p>
            </div>
            <div className="rounded-lg border border-slate-200 p-4 dark:border-slate-700">
              <p className="text-2xl font-semibold text-slate-900 dark:text-slate-100">{formatDuration(data.overall.downtimeMinutes)}</p>
              <p className="mt-1 text-xs uppercase tracking-wide text-slate-500 dark:text-slate-400">Downtime</p>
            </div>
            <div className="rounded-lg border border-slate-200 p-4 dark:border-slate-700">
              <p className="text-2xl font-semibold text-slate-900 dark:text-slate-100">{data.overall.incidentCount}</p>
              <p className="mt-1 text-xs uppercase tracking-wide text-slate-500 dark:text-slate-400">Incidents</p>
            </div>
          </div>

          {activeIncidentCount > 0 && (
            <p className="text-xs text-amber-600 dark:text-amber-400" data-testid="availability-active-note">
              Includes {activeIncidentCount === 1 ? 'an active incident' : `${activeIncidentCount} active incidents`} &mdash; availability is live and updates as incidents progress.
            </p>
          )}

          {data.incidents.length > 0 && (
            <div data-testid="availability-incidents">
              <h3 className="mb-2 text-sm font-semibold uppercase tracking-wider text-slate-500 dark:text-slate-400">Downtime by Incident</h3>
              <div className="divide-y divide-slate-100 dark:divide-slate-800/60">
                {data.incidents.map((incident) => (
                  <div key={incident.id} className="flex flex-col gap-1 py-3 sm:flex-row sm:items-center sm:justify-between">
                    <div className="min-w-0">
                      <div className="flex items-center gap-2">
                        <StatusPageBadge status={incidentImpactToSeverity(incident.impact)} />
                        <span className="truncate text-sm font-medium text-slate-900 dark:text-slate-100">{incident.title}</span>
                      </div>
                      <p className="mt-0.5 font-mono text-xs text-slate-500 dark:text-slate-400">
                        {formatDateTime(incident.startedAt)}
                        {incident.resolvedAt && <> &rarr; {formatDateTime(incident.resolvedAt)}</>}
                        {incident.affectedComponents.length > 0 && <> &middot; {incident.affectedComponents.map((component) => component.name).join(', ')}</>}
                      </p>
                    </div>
                    <span className="shrink-0 font-mono text-sm text-slate-900 dark:text-slate-100">{formatDuration(incident.effectiveDowntimeMinutes)}</span>
                  </div>
                ))}
                <div className="flex items-center justify-between py-2 text-sm font-semibold text-slate-900 dark:text-slate-100">
                  <span>Total</span>
                  <span className="font-mono">{formatDuration(data.overall.downtimeMinutes)}</span>
                </div>
              </div>
            </div>
          )}

          {data.services.length > 0 && (
            <div data-testid="availability-services">
              <h3 className="mb-2 text-sm font-semibold uppercase tracking-wider text-slate-500 dark:text-slate-400">Availability by Service</h3>
              <div className="divide-y divide-slate-100 dark:divide-slate-800/60">
                {data.services.map((service) => (
                  <div key={service.componentId} className="flex items-center justify-between py-3">
                    <span className="truncate text-sm font-medium text-slate-900 dark:text-slate-100">{service.name}</span>
                    <div className="flex items-center gap-3">
                      <span className="font-mono text-xs text-slate-500 dark:text-slate-400">{formatDuration(service.downtimeMinutes)} down</span>
                      <span className={`font-mono text-sm font-semibold ${service.availability >= 99.9 ? 'text-emerald-700 dark:text-emerald-400' : 'text-slate-900 dark:text-slate-100'}`}>
                        {formatPercent(service.availability)}
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      )}
    </section>
  )
}
