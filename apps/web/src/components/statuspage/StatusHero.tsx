import { AlertTriangle, Gauge, ShieldCheck, Wrench, XCircle } from 'lucide-react'
import type { ComponentStatus } from '../../types'
import { getOverallStatusLabel, getStatusIllustration } from '../../lib/utils'

interface StatusHeroProps {
  overallStatus: ComponentStatus
  activeIncidentCount: number
  groupCount: number
  serviceCount: number
  maintenanceCount: number
}

function getStatusAccent(status: ComponentStatus) {
  switch (status) {
    case 'operational':
      return {
        dot: 'bg-[#10B981]',
        icon: 'text-[#10B981]',
        iconSurface: 'border-[#A7F3D0] bg-white/70 text-[#047857]',
        card: 'border-[#BBF7D0] bg-[#ECFDF5]',
      }
    case 'degraded_performance':
      return {
        dot: 'bg-[#F59E0B]',
        icon: 'text-[#F59E0B]',
        iconSurface: 'border-[#FDE68A] bg-white/70 text-[#B45309]',
        card: 'border-[#FDE68A] bg-[#FFFBEB]',
      }
    case 'partial_outage':
      return {
        dot: 'bg-[#F97316]',
        icon: 'text-[#F97316]',
        iconSurface: 'border-[#FED7AA] bg-white/70 text-[#C2410C]',
        card: 'border-[#FED7AA] bg-[#FFF7ED]',
      }
    case 'major_outage':
      return {
        dot: 'bg-[#EF4444]',
        icon: 'text-[#EF4444]',
        iconSurface: 'border-[#FECACA] bg-white/70 text-[#B91C1C]',
        card: 'border-[#FECACA] bg-[#FEF2F2]',
      }
    case 'maintenance':
      return {
        dot: 'bg-[#3B82F6]',
        icon: 'text-[#3B82F6]',
        iconSurface: 'border-[#BFDBFE] bg-white/70 text-[#1D4ED8]',
        card: 'border-[#BFDBFE] bg-[#EFF6FF]',
      }
  }
}

function getStatusMetricClasses(status: ComponentStatus) {
  switch (status) {
    case 'operational':
      return 'text-[#047857]'
    case 'degraded_performance':
      return 'text-[#B45309]'
    case 'partial_outage':
      return 'text-[#C2410C]'
    case 'major_outage':
      return 'text-[#B91C1C]'
    case 'maintenance':
      return 'text-[#1D4ED8]'
  }
}

function getStatusIcon(status: ComponentStatus) {
  switch (status) {
    case 'operational':
      return ShieldCheck
    case 'degraded_performance':
      return Gauge
    case 'partial_outage':
      return AlertTriangle
    case 'major_outage':
      return XCircle
    case 'maintenance':
      return Wrench
  }
}

function getStatusDescription(status: ComponentStatus) {
  switch (status) {
    case 'operational':
      return 'All services are running normally.'
    case 'degraded_performance':
      return 'Some services are experiencing slower performance.'
    case 'partial_outage':
      return 'Some services are currently unavailable.'
    case 'major_outage':
      return 'Major disruption detected across services.'
    case 'maintenance':
      return 'We are performing scheduled maintenance.'
  }
}

export default function StatusHero({
  overallStatus,
  activeIncidentCount,
  groupCount,
  serviceCount,
  maintenanceCount,
}: StatusHeroProps) {
  const accent = getStatusAccent(overallStatus)
  const StatusIcon = getStatusIcon(overallStatus)

  return (
    <section className={`overflow-hidden rounded-[20px] border p-8 shadow-sm transition-colors dark:border-slate-800 dark:bg-slate-900 ${accent.card}`}>
      <div className="grid items-center gap-8 md:grid-cols-[55%_45%]">
        <div className="order-1 flex flex-col md:justify-between">
          <div>

            <h1 className="text-[32px] font-extrabold tracking-tight text-slate-950 dark:text-slate-50 md:text-[48px] md:leading-[1.02]">
              {getOverallStatusLabel(overallStatus)}
            </h1>
            <p className="mt-3 max-w-xl text-base leading-7 text-slate-700 dark:text-slate-300">
              {getStatusDescription(overallStatus)}
            </p>
            <p className={`mt-4 flex items-center gap-2 text-sm font-semibold ${getStatusMetricClasses(overallStatus)}`}>
              <span className={`h-2.5 w-2.5 rounded-full ${accent.dot}`} aria-hidden="true" />
              <span>{activeIncidentCount} active incident{activeIncidentCount === 1 ? '' : 's'}</span>
            </p>
          </div>
        </div>

        <div className="order-2 flex items-center justify-center md:justify-end">
          <div className="w-full max-w-[560px]">
            <img
              src={getStatusIllustration(overallStatus)}
              alt="System status illustration"
              className="h-auto w-full object-contain"
            />
          </div>
        </div>
      </div>
    </section>
  )
}
