import React from 'react'
import type { IncidentUpdate } from '../types'
import { INCIDENT_STATUS_LABELS, formatDate } from '../lib/utils'
import { getIncidentUpdateContent } from '../lib/contentModel'
import ContentRenderer from '../components/content/ContentRenderer'

interface IncidentTimelineProps {
  updates: IncidentUpdate[]
}

function getIncidentStatusToken(status: string): string {
  switch (status) {
    case 'investigating':
      return '--warning'
    case 'identified':
      return '--partial'
    case 'monitoring':
      return '--info'
    case 'resolved':
      return '--success'
    default:
      return '--text-subtle'
  }
}

const TimelineItem = ({ update }: { update: IncidentUpdate }) => {
  const statusToken = getIncidentStatusToken(update.status)
  const content = getIncidentUpdateContent(update)

  return (
    <div className="relative pb-3" key={update.id}>
      <div className="absolute left-4 top-0 h-full w-0.5 -ml-px" style={{ backgroundColor: 'var(--border)' }}></div>

      <div className="relative flex items-start mb-2">
        <div
          className="flex-shrink-0 w-8 h-8 rounded-full flex items-center justify-center"
          style={{
            backgroundColor: `color-mix(in srgb, var(${statusToken}) 14%, var(--surface))`,
            border: `1px solid color-mix(in srgb, var(${statusToken}) 36%, var(--border))`,
          }}
        >
          <div className="w-3 h-3 rounded-full" style={{ backgroundColor: `var(${statusToken})` }}></div>
        </div>

        <div className="ml-4 flex-1 min-w-0">
          <div className="flex flex-wrap items-baseline gap-x-2">
            <span className="font-medium text-sm" style={{ color: `var(${statusToken})` }}>
              {INCIDENT_STATUS_LABELS[update.status]}
            </span>
            <span className="text-xs" style={{ color: 'var(--text-subtle)' }}>
              {formatDate(update.createdAt)}
            </span>
          </div>
          <div className="text-sm" style={{ color: 'var(--text-muted)' }}>
            <ContentRenderer text={content.text} json={content.json} />
          </div>
        </div>
      </div>
    </div>
  )
}

export function IncidentTimeline({ updates }: IncidentTimelineProps) {
  return (
    <div>
      {updates.length === 0 ? (
        <p className="text-sm" style={{ color: 'var(--text-subtle)' }}>No updates yet.</p>
      ) : (
        updates.map((update) => (
          <TimelineItem key={update.id} update={update} />
        ))
      )}
    </div>
  )
}

export default IncidentTimeline
