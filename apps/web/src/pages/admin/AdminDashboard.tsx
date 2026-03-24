import React from 'react'
import { Link } from 'react-router-dom'
import { AlertTriangle, Layers, Activity, Users, Wrench, TrendingUp } from 'lucide-react'
import { useApi } from '../../hooks/useApi'
import type { Component, Incident, Monitor, Subscriber, Maintenance } from '../../types'
import { STATUS_LABELS, STATUS_COLORS, INCIDENT_STATUS_LABELS, formatDate } from '../../lib/utils'

function StatCard({
  title,
  value,
  icon: Icon,
  color,
  to,
}: {
  title: string
  value: number | string
  icon: React.ElementType
  color: string
  to: string
}) {
  return (
    <Link
      to={to}
      className="bg-white rounded-xl border border-gray-200 p-5 flex items-center gap-4 hover:shadow-md transition-shadow"
    >
      <div className={`w-12 h-12 rounded-lg flex items-center justify-center ${color}`}>
        <Icon className="w-6 h-6 text-white" />
      </div>
      <div>
        <p className="text-2xl font-bold text-gray-900">{value}</p>
        <p className="text-sm text-gray-500">{title}</p>
      </div>
    </Link>
  )
}

export default function AdminDashboard() {
  const { data: components, total: totalComponents } = useApi<Component[]>('/components')
  const { data: incidents, total: totalIncidents } = useApi<Incident[]>('/incidents')
  const { data: monitors, total: totalMonitors } = useApi<Monitor[]>('/monitors')
  const { data: subscribers, total: totalSubscribers } = useApi<Subscriber[]>('/subscribers')
  const { data: maintenance } = useApi<Maintenance[]>('/maintenance')

  const activeIncidents = incidents?.filter(i => i.status !== 'resolved') || []
  const scheduledMaintenance = Array.isArray(maintenance)
    ? maintenance.filter(m => m.status === 'scheduled')
    : []

  return (
    <div className="p-8">
      <div className="mb-8">
        <h1 className="text-2xl font-bold text-gray-900">Dashboard</h1>
        <p className="text-sm text-gray-500 mt-1">System overview</p>
      </div>

      {/* Stats grid */}
      <div className="grid grid-cols-2 lg:grid-cols-3 gap-4 mb-8">
        <StatCard
          title="Components"
          value={totalComponents}
          icon={Layers}
          color="bg-blue-500"
          to="/admin/components"
        />
        <StatCard
          title="Active Incidents"
          value={activeIncidents.length}
          icon={AlertTriangle}
          color={activeIncidents.length > 0 ? 'bg-red-500' : 'bg-green-500'}
          to="/admin/incidents"
        />
        <StatCard
          title="Monitors"
          value={totalMonitors}
          icon={Activity}
          color="bg-purple-500"
          to="/admin/monitors"
        />
        <StatCard
          title="Subscribers"
          value={totalSubscribers}
          icon={Users}
          color="bg-indigo-500"
          to="/admin/subscribers"
        />
        <StatCard
          title="Scheduled Maintenance"
          value={scheduledMaintenance.length}
          icon={Wrench}
          color="bg-yellow-500"
          to="/admin/maintenance"
        />
        <StatCard
          title="Total Incidents"
          value={totalIncidents}
          icon={TrendingUp}
          color="bg-gray-500"
          to="/admin/incidents"
        />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Component Status */}
        <div className="bg-white rounded-xl border border-gray-200 p-5">
          <h2 className="font-semibold text-gray-900 mb-4">Component Status</h2>
          {(components || []).length === 0 ? (
            <p className="text-sm text-gray-400">No components yet.</p>
          ) : (
            <div className="space-y-2">
              {(components || []).map(c => (
                <div key={c.id} className="flex items-center justify-between py-1.5 border-b border-gray-50 last:border-0">
                  <span className="text-sm text-gray-700">{c.name}</span>
                  <span className={`flex items-center gap-1.5 text-xs font-medium`}>
                    <span className={`w-2 h-2 rounded-full ${STATUS_COLORS[c.status]}`} />
                    {STATUS_LABELS[c.status]}
                  </span>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Recent Incidents */}
        <div className="bg-white rounded-xl border border-gray-200 p-5">
          <h2 className="font-semibold text-gray-900 mb-4">Recent Incidents</h2>
          {(incidents || []).length === 0 ? (
            <p className="text-sm text-gray-400">No incidents.</p>
          ) : (
            <div className="space-y-3">
              {(incidents || []).slice(0, 5).map(inc => (
                <div key={inc.id} className="border-b border-gray-50 pb-3 last:border-0 last:pb-0">
                  <div className="flex items-start justify-between gap-2">
                    <p className="text-sm font-medium text-gray-900">{inc.title}</p>
                    <span className={`text-xs px-2 py-0.5 rounded-full flex-shrink-0 ${inc.status === 'resolved'
                        ? 'bg-green-100 text-green-700'
                        : 'bg-red-100 text-red-700'
                      }`}>
                      {INCIDENT_STATUS_LABELS[inc.status]}
                    </span>
                  </div>
                  <p className="text-xs text-gray-400 mt-0.5">{formatDate(inc.createdAt)}</p>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
