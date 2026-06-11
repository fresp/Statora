import { useMemo, useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { ArrowLeft, Calendar as CalendarIcon, ChevronDown, ChevronLeft, ChevronRight, Filter, Search } from 'lucide-react'
import { useApi } from '../hooks/useApi'
import { IncidentTimeline } from '../components/IncidentTimeline'
import type { Incident, StatusPageSettings } from '../types'
import { INCIDENT_STATUS_LABELS, formatDate } from '../lib/utils'
import { DEFAULT_STATUS_PAGE_SETTINGS, normalizeStatusPageSettings } from '../lib/statusPageSettings'
import { DEFAULT_THEME_PRESET, getThemePresets, loadThemePresetStylesheet } from '../lib/themePresetLoader'
import { StatusPageBadge, StatusPageFrame, incidentImpactToSeverity } from '../components/statuspage/StatusPagePrimitives'

interface QuarterCursor {
  year: number
  quarter: number
}

interface MonthBucket {
  monthIndex: number
  monthLabel: string
  incidents: Incident[]
}

const MONTH_SHORT = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec']

function toLocalBoundaryIso(date: Date, endOfDay: boolean): string {
  const boundary = new Date(date)
  if (endOfDay) boundary.setHours(23, 59, 59, 999)
  else boundary.setHours(0, 0, 0, 0)
  return boundary.toISOString()
}

function getCurrentQuarterCursor(reference = new Date()): QuarterCursor {
  return {
    year: reference.getFullYear(),
    quarter: Math.floor(reference.getMonth() / 3),
  }
}

function toQuarterIndex(cursor: QuarterCursor): number {
  return cursor.year * 4 + cursor.quarter
}

function shiftQuarter(cursor: QuarterCursor, delta: number): QuarterCursor {
  const absoluteQuarter = toQuarterIndex(cursor) + delta
  const year = Math.floor(absoluteQuarter / 4)
  const quarter = ((absoluteQuarter % 4) + 4) % 4
  return { year, quarter }
}

function formatQuarterLabel(cursor: QuarterCursor): string {
  const startMonth = cursor.quarter * 3
  const endMonth = startMonth + 2
  return `${MONTH_SHORT[startMonth]} – ${MONTH_SHORT[endMonth]} ${cursor.year}`
}

function getQuarterDateRange(cursor: QuarterCursor, today: Date): { startDate: string; endDate: string } {
  const quarterStartMonth = cursor.quarter * 3
  const start = new Date(cursor.year, quarterStartMonth, 1)
  const isCurrentQuarter = cursor.year === today.getFullYear() && cursor.quarter === Math.floor(today.getMonth() / 3)
  const end = isCurrentQuarter
    ? new Date(today.getFullYear(), today.getMonth(), today.getDate())
    : new Date(cursor.year, quarterStartMonth + 3, 0)

  return {
    startDate: toLocalBoundaryIso(start, false),
    endDate: toLocalBoundaryIso(end, true),
  }
}

function groupIncidentsByQuarterMonths(incidents: Incident[], cursor: QuarterCursor): MonthBucket[] {
  const quarterStartMonth = cursor.quarter * 3

  return Array.from({ length: 3 }, (_, offset) => {
    const monthIndex = quarterStartMonth + (2 - offset)
    const monthLabel = new Date(cursor.year, monthIndex, 1).toLocaleString('en-US', { month: 'long' })
    const monthIncidents = incidents
      .filter((incident) => {
        const createdAt = new Date(incident.createdAt)
        return createdAt.getFullYear() === cursor.year && createdAt.getMonth() === monthIndex
      })
      .sort((a, b) => new Date(b.createdAt).getTime() - new Date(a.createdAt).getTime())

    return { monthIndex, monthLabel, incidents: monthIncidents }
  })
}

export default function HistoryPage() {
  const today = useMemo(() => new Date(), [])
  const currentQuarter = useMemo(() => getCurrentQuarterCursor(today), [today])
  const [selectedQuarter, setSelectedQuarter] = useState<QuarterCursor>(currentQuarter)
  const [expandedId, setExpandedId] = useState<string | null>(null)

  const quarterRange = useMemo(() => getQuarterDateRange(selectedQuarter, today), [selectedQuarter, today])
  const incidentsUrl = useMemo(
    () => `/status/incidents?start_date=${encodeURIComponent(quarterRange.startDate)}&end_date=${encodeURIComponent(quarterRange.endDate)}`,
    [quarterRange.endDate, quarterRange.startDate]
  )

  const { data: incidentData, loading: incidentsLoading, error: incidentsError } = useApi<{ active: Incident[]; resolved: Incident[] }>(incidentsUrl, [incidentsUrl])
  const { data: settingsData } = useApi<StatusPageSettings>('/status/settings')

  const settings = normalizeStatusPageSettings(settingsData ?? DEFAULT_STATUS_PAGE_SETTINGS)
  const allIncidents = useMemo(() => [...(incidentData?.active ?? []), ...(incidentData?.resolved ?? [])], [incidentData])
  const monthBuckets = useMemo(() => groupIncidentsByQuarterMonths(allIncidents, selectedQuarter), [allIncidents, selectedQuarter])
  const canGoNext = toQuarterIndex(selectedQuarter) < toQuarterIndex(currentQuarter)
  const themePreset = (settingsData?.theme?.preset?.trim() || DEFAULT_THEME_PRESET).endsWith('.css')
    ? settingsData?.theme?.preset?.trim() || DEFAULT_THEME_PRESET
    : `${settingsData?.theme?.preset?.trim() || DEFAULT_THEME_PRESET}.css`

  useEffect(() => {
    const pageTitle = settingsData?.head?.title?.trim() || 'Status Page'
    document.title = `${pageTitle} - Incident History`
  }, [settingsData?.head?.title])

  useEffect(() => {
    const presets = getThemePresets().presets
    loadThemePresetStylesheet(themePreset, presets).catch(() => { })
  }, [themePreset])

  return (
    <StatusPageFrame settings={settings}>
      <main className="mx-auto max-w-[1024px] px-4 py-8 md:px-8 md:py-12">
        <div className="mb-10">
          <nav className="mb-4">
            <Link to="/" className="inline-flex items-center gap-1.5 font-mono text-sm text-slate-600 transition-colors hover:text-emerald-700 dark:text-slate-400 dark:hover:text-emerald-400">
              <ArrowLeft className="h-4 w-4" /> Back to Dashboard
            </Link>
          </nav>
          <h1 className="mb-2 text-3xl font-bold">Incident History</h1>
          <p className="text-lg text-slate-600 dark:text-slate-400">A detailed record of service interruptions and maintenance events.</p>
        </div>

        <div className="mb-10 flex flex-col items-center justify-between gap-4 rounded-xl border border-slate-200 bg-white p-4 shadow-sm transition-colors dark:border-slate-800 dark:bg-slate-900 md:flex-row">
          <div className="relative w-full md:w-96">
            <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-slate-400" />
            <input
              type="text"
              placeholder="Search incidents..."
              className="w-full rounded-lg border border-slate-200 bg-slate-50 py-2 pl-9 pr-4 font-mono text-sm outline-none transition-colors focus:ring-2 focus:ring-emerald-500 dark:border-slate-700 dark:bg-slate-800"
              readOnly
            />
          </div>
          <div className="flex w-full gap-2 overflow-x-auto md:w-auto">
            <button className="flex flex-1 items-center justify-center gap-2 whitespace-nowrap rounded-lg border border-slate-200 px-4 py-2 font-mono text-xs font-semibold text-slate-700 transition-colors hover:bg-slate-50 dark:border-slate-700 dark:text-slate-300 dark:hover:bg-slate-800 md:flex-none">
              <Filter className="h-4 w-4" /> Filter by Severity
            </button>
            <button className="flex flex-1 items-center justify-center gap-2 whitespace-nowrap rounded-lg border border-slate-200 px-4 py-2 font-mono text-xs font-semibold text-slate-700 transition-colors hover:bg-slate-50 dark:border-slate-700 dark:text-slate-300 dark:hover:bg-slate-800 md:flex-none">
              <CalendarIcon className="h-4 w-4" /> {formatQuarterLabel(selectedQuarter)}
            </button>
          </div>
        </div>

        <section className="mb-10 rounded-xl border border-slate-200 bg-white p-4 shadow-sm dark:border-slate-800 dark:bg-slate-900">
          <div className="flex items-center justify-between gap-4">
            <button
              onClick={() => setSelectedQuarter((prev) => shiftQuarter(prev, -1))}
              className="inline-flex items-center gap-1 rounded-lg border border-slate-200 bg-slate-50 px-3 py-2 text-sm font-medium transition-colors hover:bg-white dark:border-slate-700 dark:bg-slate-800 dark:hover:bg-slate-700"
            >
              <ChevronLeft className="h-4 w-4" /> Prev
            </button>
            <div className="text-center font-mono text-sm font-semibold">{formatQuarterLabel(selectedQuarter)}</div>
            <button
              onClick={() => setSelectedQuarter((prev) => shiftQuarter(prev, 1))}
              disabled={!canGoNext}
              className="inline-flex items-center gap-1 rounded-lg border border-slate-200 bg-slate-50 px-3 py-2 text-sm font-medium transition-colors hover:bg-white disabled:cursor-not-allowed disabled:opacity-50 dark:border-slate-700 dark:bg-slate-800 dark:hover:bg-slate-700"
            >
              Next <ChevronRight className="h-4 w-4" />
            </button>
          </div>
        </section>

        {incidentsLoading ? (
          <div className="rounded-xl border border-slate-200 bg-white p-6 font-mono text-sm text-slate-500 shadow-sm dark:border-slate-800 dark:bg-slate-900 dark:text-slate-400">Loading incident history...</div>
        ) : incidentsError ? (
          <div className="rounded-xl border border-slate-200 bg-white p-6 text-sm text-slate-500 shadow-sm dark:border-slate-800 dark:bg-slate-900 dark:text-slate-400">Unable to load incidents for this quarter.</div>
        ) : (
          <div className="space-y-10">
            <section>
              <div className="sticky top-[73px] z-10 mb-6 border-b border-slate-200 bg-slate-50/90 py-3 backdrop-blur-md transition-colors dark:border-slate-800 dark:bg-slate-900/90">
                <h2 className="text-xl font-semibold">Incident Log</h2>
              </div>

              <div className="space-y-8">
                {monthBuckets.map((monthGroup) => (
                  <div key={`${selectedQuarter.year}-${monthGroup.monthIndex}`} className="space-y-6">
                    <h3 className="font-mono text-sm font-semibold uppercase tracking-wider text-slate-500 dark:text-slate-400">{monthGroup.monthLabel}</h3>
                    {monthGroup.incidents.length === 0 ? (
                      <div className="rounded-xl border border-dashed border-slate-200 bg-white p-6 text-sm text-slate-500 shadow-sm dark:border-slate-800 dark:bg-slate-900 dark:text-slate-400">No incidents reported in this month.</div>
                    ) : monthGroup.incidents.map((incident) => {
                      const isExpanded = expandedId === incident.id
                      return (
                        <article key={incident.id} className="rounded-xl border border-slate-200 bg-white p-6 shadow-sm transition-colors hover:border-slate-300 dark:border-slate-800 dark:bg-slate-900 dark:hover:border-slate-700">
                          <div className="mb-4 flex flex-col justify-between gap-4 md:flex-row md:items-start">
                            <div>
                              <div className="mb-2 flex flex-wrap items-center gap-3">
                                <h3 className="text-xl font-semibold">{incident.title}</h3>
                                <StatusPageBadge status={incidentImpactToSeverity(incident.impact)} />
                              </div>
                              <p className="font-mono text-xs uppercase tracking-wider text-slate-500 dark:text-slate-400">
                                {incident.id} &bull; Started: {formatDate(incident.createdAt)}
                              </p>
                            </div>
                            <StatusPageBadge status={incident.status === 'resolved' ? 'operational' : incidentImpactToSeverity(incident.impact)} solid={incident.status === 'resolved'} label={INCIDENT_STATUS_LABELS[incident.status]} />
                          </div>

                          <div className="mb-4">
                            <p className="mb-4 text-base leading-relaxed text-slate-700 dark:text-slate-300">{incident.description}</p>
                            <div className="flex flex-wrap gap-2">
                              {(incident.affectedComponentTargets && incident.affectedComponentTargets.length > 0
                                ? incident.affectedComponentTargets.map((target) => target.component.name)
                                : incident.affectedComponents.map((component) => component.name)
                              ).map((name) => (
                                <span key={name} className="rounded border border-slate-200 bg-slate-100 px-2 py-1 font-mono text-xs font-semibold text-slate-600 dark:border-slate-700 dark:bg-slate-800 dark:text-slate-300">
                                  {name}
                                </span>
                              ))}
                            </div>
                          </div>

                          <div className="mt-2 border-t border-slate-100 pt-4 dark:border-slate-800/60">
                            <button
                              onClick={() => setExpandedId((prev) => prev === incident.id ? null : incident.id)}
                              className="flex items-center gap-1 font-mono text-sm font-semibold text-cyan-700 hover:underline focus:outline-none dark:text-cyan-400"
                            >
                              {isExpanded ? 'Hide' : 'View'} Updates
                              <ChevronDown className={`h-4 w-4 transition-transform ${isExpanded ? 'rotate-180' : ''}`} />
                            </button>
                            {isExpanded && <div className="mt-6"><IncidentTimeline updates={incident.updates || []} /></div>}
                          </div>
                        </article>
                      )
                    })}
                  </div>
                ))}
              </div>
            </section>
          </div>
        )}
      </main>
    </StatusPageFrame>
  )
}
