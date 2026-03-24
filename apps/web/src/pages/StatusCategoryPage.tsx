import { Link, useParams } from 'react-router-dom'
import { AlertCircle, AlertTriangle, CheckCircle, ChevronRight, Wrench, XCircle } from 'lucide-react'
import { useCategorySummary } from '../hooks/useApi'
import { STATUS_LABELS, formatDate } from '../lib/utils'
import type { ComponentStatus } from '../types'

function getStatusToken(status: ComponentStatus): string {
  switch (status) {
    case 'operational':
      return '--status-operational'
    case 'degraded_performance':
      return '--status-degraded'
    case 'partial_outage':
      return '--status-partial'
    case 'major_outage':
      return '--status-major'
    case 'maintenance':
      return '--status-maintenance'
    default:
      return '--status-operational'
  }
}

function StatusIcon({ status }: { status: ComponentStatus }) {
  const cls = 'w-5 h-5'
  const color = `var(${getStatusToken(status)})`

  switch (status) {
    case 'operational':
      return <CheckCircle className={cls} style={{ color }} />
    case 'degraded_performance':
      return <AlertTriangle className={cls} style={{ color }} />
    case 'partial_outage':
      return <AlertCircle className={cls} style={{ color }} />
    case 'major_outage':
      return <XCircle className={cls} style={{ color }} />
    case 'maintenance':
      return <Wrench className={cls} style={{ color }} />
    default:
      return <CheckCircle className={cls} style={{ color }} />
  }
}

function UptimeBar({ bars }: { bars: { date: string; uptimePercent: number; status: ComponentStatus }[] }) {
  return (
    <div className="flex gap-px items-end h-8 mt-2">
      {bars.map((bar, i) => (
        <div
          key={`${bar.date}-${i}`}
          className="flex-1 rounded-sm opacity-80 hover:opacity-100 transition-opacity"
          style={{
            backgroundColor: `var(${getStatusToken(bar.status)})`,
            height: `${Math.max(20, (bar.uptimePercent / 100) * 32)}px`,
          }}
          title={`${bar.date}: ${bar.uptimePercent.toFixed(2)}% uptime`}
        />
      ))}
    </div>
  )
}

export default function StatusCategoryPage() {
  const { categoryPrefix } = useParams<{ categoryPrefix: string }>()
  const { data, loading, error } = useCategorySummary(categoryPrefix)

  if (loading) {
    return (
      <div className="min-h-screen bg-[var(--bg)] text-[var(--text)]">
        <div className="max-w-5xl mx-auto px-4 py-10">Loading category status…</div>
      </div>
    )
  }

  if (error || !data) {
    return (
      <div className="min-h-screen bg-[var(--bg)] text-[var(--text)]">
        <div className="max-w-5xl mx-auto px-4 py-10 space-y-3">
          <nav className="text-sm text-[var(--text-muted)] flex items-center gap-2">
            <Link to="/" className="hover:underline">Status</Link>
            <ChevronRight className="w-4 h-4" />
            <span>{categoryPrefix ?? 'Unknown category'}</span>
          </nav>
          <div className="rounded-xl border border-[var(--border)] bg-[var(--surface)] p-6">
            <h1 className="text-xl font-semibold mb-2">Category unavailable</h1>
            <p className="text-sm text-[var(--text-muted)]">{error ?? 'Unable to load this category right now.'}</p>
          </div>
        </div>
      </div>
    )
  }

  const aggregateStatus = data.aggregateStatus

  return (
    <div className="min-h-screen bg-[var(--bg)] text-[var(--text)]">
      <div className="max-w-5xl mx-auto px-4 py-10 space-y-8">
        <nav className="text-sm text-[var(--text-muted)] flex items-center gap-2">
          <Link to="/" className="hover:underline">Status</Link>
          <ChevronRight className="w-4 h-4" />
          <span>{data.name}</span>
        </nav>

        <header className="rounded-xl border border-[var(--border)] bg-[var(--surface)] p-6">
          <div className="flex items-center justify-between gap-4">
            <div>
              <h1 className="text-2xl font-bold">{data.name}</h1>
              {data.description && <p className="text-sm mt-1 text-[var(--text-muted)]">{data.description}</p>}
            </div>
            <div className="flex items-center gap-2">
              <StatusIcon status={aggregateStatus} />
              <span className="text-sm font-semibold" style={{ color: `var(${getStatusToken(aggregateStatus)})` }}>
                {STATUS_LABELS[aggregateStatus]}
              </span>
            </div>
          </div>
          <p className="mt-4 text-sm text-[var(--text-muted)]">90-day category uptime: {data.uptime90d.toFixed(2)}%</p>
        </header>

        <section className="space-y-4">
          <h2 className="text-xl font-semibold">Services</h2>
          {data.services.length === 0 ? (
            <div className="rounded-xl border border-[var(--border)] bg-[var(--surface)] p-6 text-sm text-[var(--text-muted)]">
              No services are mapped to this category yet.
            </div>
          ) : (
            <div className="space-y-4">
              {data.services.map((service) => (
                <article key={service.id} className="rounded-xl border border-[var(--border)] bg-[var(--surface)] p-5">
                  <div className="flex items-center justify-between gap-4">
                    <div>
                      <h3 className="text-lg font-semibold">{service.name}</h3>
                      {service.description && (
                        <p className="text-sm text-[var(--text-muted)] mt-1">{service.description}</p>
                      )}
                    </div>
                    <div className="flex items-center gap-2">
                      <StatusIcon status={service.status} />
                      <span className="text-sm font-medium" style={{ color: `var(${getStatusToken(service.status)})` }}>
                        {STATUS_LABELS[service.status]}
                      </span>
                    </div>
                  </div>
                  <div className="mt-3 text-xs text-[var(--text-muted)]">90-day uptime: {service.uptime90d.toFixed(2)}%</div>
                  {service.uptimeHistory.length > 0 && <UptimeBar bars={service.uptimeHistory} />}
                </article>
              ))}
            </div>
          )}
        </section>

        <section className="space-y-4">
          <h2 className="text-xl font-semibold">Incidents</h2>
          {data.incidents.length === 0 ? (
            <div className="rounded-xl border border-[var(--border)] bg-[var(--surface)] p-6 text-sm text-[var(--text-muted)]">
              No incidents linked to this category in the selected history window.
            </div>
          ) : (
            <div className="space-y-3">
              {data.incidents.map((incident) => (
                <article key={incident.id} className="rounded-xl border border-[var(--border)] bg-[var(--surface)] p-5">
                  <div className="flex items-center justify-between gap-4">
                    <h3 className="font-semibold">{incident.title}</h3>
                    <span className="text-xs uppercase tracking-wide text-[var(--text-muted)]">{incident.status}</span>
                  </div>
                  <p className="text-sm mt-2 text-[var(--text-muted)]">{incident.description}</p>
                  <p className="text-xs mt-3 text-[var(--text-subtle)]">Started: {formatDate(incident.createdAt)}</p>
                </article>
              ))}
            </div>
          )}
        </section>
      </div>
    </div>
  )
}
