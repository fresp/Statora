import React, { useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { ChevronDown, ChevronUp } from 'lucide-react'
import { useApi } from '../hooks/useApi'
import { IncidentTimeline } from '../components/IncidentTimeline'
import type { Incident, StatusPageSettings } from '../types'
import {
  formatDate,
  INCIDENT_IMPACT_LABELS,
  INCIDENT_STATUS_LABELS,
  groupIncidentsByYearQuarterMonth,
  type HistoryYearGroup,
} from '../lib/utils'
import { loadThemePresetStylesheet, getThemePresets, DEFAULT_THEME_PRESET } from '../lib/themePresetLoader'

const DEFAULT_HISTORY_PAGE_SIZE = 2

export default function HistoryPage() {
  const { data: incidentData } = useApi<{ active: Incident[]; resolved: Incident[] }>('/status/incidents')
  const { data: settingsData } = useApi<StatusPageSettings>('/status/settings')
  const [expandedIncidents, setExpandedIncidents] = useState<Set<string>>(new Set())
  const [visibleYears, setVisibleYears] = useState(DEFAULT_HISTORY_PAGE_SIZE)

  const allIncidents = useMemo(
    () => [...(incidentData?.active ?? []), ...(incidentData?.resolved ?? [])],
    [incidentData]
  )

  const groupedHistory = useMemo<HistoryYearGroup[]>(
    () => groupIncidentsByYearQuarterMonth(allIncidents),
    [allIncidents]
  )

  const paginatedHistory = groupedHistory.slice(0, visibleYears)
  const hasMoreYears = visibleYears < groupedHistory.length

  const themePreset = (settingsData?.theme?.preset?.trim() || DEFAULT_THEME_PRESET).endsWith('.css')
    ? settingsData?.theme?.preset?.trim() || DEFAULT_THEME_PRESET
    : `${settingsData?.theme?.preset?.trim() || DEFAULT_THEME_PRESET}.css`

  React.useEffect(() => {
    const presets = getThemePresets().presets
    loadThemePresetStylesheet(themePreset, presets).catch(() => {})
  }, [themePreset])

  return (
    <div
      className="min-h-screen"
      style={{
        backgroundColor: 'var(--bg)',
        color: 'var(--text)',
        fontFamily: 'var(--font-family)',
      }}
    >
      <div
        className="py-10 px-4 border-b"
        style={{
          borderColor: 'var(--border)',
          backgroundColor: 'var(--surface)',
        }}
      >
        <div className="max-w-5xl mx-auto flex items-center justify-between gap-4">
          <div>
            <h1 className="text-3xl font-bold">Incident History</h1>
            <p className="text-sm mt-1" style={{ color: 'var(--text-muted)' }}>
              Full archive grouped by year, quarter, and month.
            </p>
          </div>
          <Link
            to="/"
            className="inline-flex items-center rounded-lg px-4 py-2 text-sm font-medium border"
            style={{
              borderColor: 'var(--border)',
              color: 'var(--text)',
              backgroundColor: 'var(--surface)',
            }}
          >
            Back to Status
          </Link>
        </div>
      </div>

      <div className="max-w-5xl mx-auto px-4 py-8 space-y-8">
        {paginatedHistory.length === 0 ? (
          <div
            className="rounded-xl border p-6"
            style={{
              borderColor: 'var(--border)',
              backgroundColor: 'var(--surface)',
            }}
          >
            <p className="text-sm" style={{ color: 'var(--text-muted)' }}>No incident history available yet.</p>
          </div>
        ) : (
          paginatedHistory.map((yearGroup) => (
            <section key={yearGroup.year} className="space-y-4">
              <h2 className="text-2xl font-semibold">{yearGroup.year}</h2>

              {yearGroup.quarters.map((quarterGroup) => (
                <div
                  key={`${yearGroup.year}-${quarterGroup.quarter}`}
                  className="rounded-xl border p-5"
                  style={{
                    borderColor: 'var(--border)',
                    backgroundColor: 'var(--surface)',
                  }}
                >
                  <h3 className="text-lg font-semibold mb-4">{quarterGroup.quarter}</h3>

                  <div className="space-y-4">
                    {quarterGroup.months.map((monthGroup) => (
                      <div
                        key={`${yearGroup.year}-${quarterGroup.quarter}-${monthGroup.monthIndex}`}
                        className="rounded-lg border p-4"
                        style={{
                          borderColor: 'var(--border)',
                          backgroundColor: 'var(--surface-uptime)',
                        }}
                      >
                        <div className="flex items-center justify-between mb-3">
                          <h4 className="font-medium">{monthGroup.monthLabel}</h4>
                          <span className="text-xs" style={{ color: 'var(--text-subtle)' }}>
                            {monthGroup.incidents.length} incident{monthGroup.incidents.length === 1 ? '' : 's'}
                          </span>
                        </div>

                        {monthGroup.incidents.length === 0 ? (
                          <p className="text-sm" style={{ color: 'var(--text-muted)' }}>
                            No incidents reported in this month.
                          </p>
                        ) : (
                          <div className="space-y-3">
                            {monthGroup.incidents.map((incident) => {
                              const isExpanded = expandedIncidents.has(incident.id)
                              const isResolved = incident.status === 'resolved'

                              return (
                                <div
                                  key={incident.id}
                                  className="rounded-lg border p-4"
                                  style={{
                                    borderColor: 'var(--border-incident)',
                                    backgroundColor: 'var(--surface-incident)',
                                  }}
                                >
                                  <div className="flex items-start justify-between gap-3">
                                    <div>
                                      <h5 className="font-medium">{incident.title}</h5>
                                      <p className="text-sm mt-1" style={{ color: 'var(--text-muted)' }}>
                                        {incident.description}
                                      </p>
                                    </div>

                                    <div className="flex items-center gap-3">
                                      <span
                                        className="text-xs px-2 py-1 rounded-full font-medium"
                                        style={
                                          isResolved
                                            ? { backgroundColor: 'var(--status-resolved-bg)', color: 'var(--status-resolved-text)' }
                                            : { backgroundColor: 'var(--status-active-bg)', color: 'var(--status-active-text)' }
                                        }
                                      >
                                        {INCIDENT_STATUS_LABELS[incident.status]}
                                      </span>

                                      <button
                                        onClick={() => {
                                          setExpandedIncidents((prev) => {
                                            const next = new Set(prev)
                                            if (next.has(incident.id)) {
                                              next.delete(incident.id)
                                            } else {
                                              next.add(incident.id)
                                            }
                                            return next
                                          })
                                        }}
                                        className="flex-shrink-0 transition-colors"
                                        style={{ color: 'var(--text-subtle)' }}
                                      >
                                        {isExpanded ? <ChevronUp className="w-5 h-5" /> : <ChevronDown className="w-5 h-5" />}
                                      </button>
                                    </div>
                                  </div>

                                  <div className="flex flex-wrap gap-4 mt-3 text-xs" style={{ color: 'var(--text-subtle)' }}>
                                    <span>Impact: {INCIDENT_IMPACT_LABELS[incident.impact]}</span>
                                    <span>Created: {formatDate(incident.createdAt)}</span>
                                    {incident.resolvedAt && <span>Resolved: {formatDate(incident.resolvedAt)}</span>}
                                    {incident.creatorUsername && <span>Created by: {incident.creatorUsername}</span>}
                                  </div>

                                  {isExpanded && <IncidentTimeline updates={incident.updates || []} />}
                                </div>
                              )
                            })}
                          </div>
                        )}
                      </div>
                    ))}
                  </div>
                </div>
              ))}
            </section>
          ))
        )}

        {hasMoreYears && (
          <div className="flex justify-center pt-2">
            <button
              onClick={() => setVisibleYears((prev) => prev + DEFAULT_HISTORY_PAGE_SIZE)}
              className="inline-flex items-center rounded-lg px-4 py-2 text-sm font-medium border"
              style={{
                borderColor: 'var(--border)',
                color: 'var(--text)',
                backgroundColor: 'var(--surface)',
              }}
            >
              Load More Years
            </button>
          </div>
        )}
      </div>
    </div>
  )
}
