import React from 'react'
import { Link, useLocation } from 'react-router-dom'
import { Bell, CheckCircle, FileText, Lock, Mail, Shield } from 'lucide-react'
import type { ComponentStatus, StatusPageSettings } from '../../types'

export type StatusPageSeverity = 'operational' | 'minor' | 'major' | 'critical'

export function componentStatusToSeverity(status: ComponentStatus): StatusPageSeverity {
  switch (status) {
    case 'degraded_performance':
      return 'minor'
    case 'partial_outage':
      return 'major'
    case 'major_outage':
      return 'critical'
    case 'maintenance':
      return 'minor'
    case 'operational':
      return 'operational'
  }
}

export function incidentImpactToSeverity(impact: string): StatusPageSeverity {
  switch (impact.toLowerCase()) {
    case 'minor':
      return 'minor'
    case 'major':
      return 'major'
    case 'critical':
      return 'critical'
    default:
      return 'operational'
  }
}

export function StatusPageBadge({ status, solid = false, label }: { status: StatusPageSeverity; solid?: boolean; label?: string }) {
  const styles = (() => {
    switch (status) {
      case 'operational':
        return solid
          ? 'bg-emerald-600 text-white'
          : 'border border-emerald-200 bg-emerald-100 text-emerald-800 dark:border-emerald-800 dark:bg-emerald-950/40 dark:text-emerald-400'
      case 'minor':
        return solid
          ? 'bg-amber-500 text-white'
          : 'border border-amber-200 bg-amber-100 text-amber-800 dark:border-amber-800 dark:bg-amber-950/40 dark:text-amber-400'
      case 'major':
        return solid
          ? 'bg-orange-500 text-white'
          : 'border border-orange-200 bg-orange-100 text-orange-800 dark:border-orange-800 dark:bg-orange-950/40 dark:text-orange-400'
      case 'critical':
        return solid
          ? 'bg-red-600 text-white'
          : 'border border-red-200 bg-red-100 text-red-800 dark:border-red-800 dark:bg-red-950/40 dark:text-red-400'
    }
  })()

  const text = label ?? (() => {
    switch (status) {
      case 'minor':
        return 'Degraded'
      case 'major':
        return 'Partial Outage'
      case 'critical':
        return 'Major Outage'
      case 'operational':
        return 'Operational'
    }
  })()

  return (
    <span className={`rounded-full px-2.5 py-0.5 font-mono text-[11px] font-semibold uppercase tracking-wider ${styles}`}>
      {text}
    </span>
  )
}

export function StatusPageHeader({ settings }: { settings: StatusPageSettings }) {
  const location = useLocation()
  const navLinks = [
    { name: 'Current Status', path: '/' },
    { name: 'Incidents', path: '/history' },
  ]

  return (
    <header className="sticky top-0 z-50 border-b border-slate-200 bg-white/80 backdrop-blur-md transition-colors duration-200 dark:border-slate-800 dark:bg-slate-900/80">
      <div className="mx-auto flex max-w-[768px] items-center justify-center px-4 py-4 md:px-8">
        <div className="flex min-w-0 items-center gap-8">
          <Link to="/" className="flex min-w-0 items-center gap-3 text-xl font-bold text-emerald-700 dark:text-emerald-400">
            {settings.branding.logoUrl ? (
              <img src={settings.branding.logoUrl} alt={`${settings.branding.siteName} logo`} className="h-8 w-8 rounded-lg object-contain" />
            ) : (
              <CheckCircle className="h-7 w-7 shrink-0" />
            )}
            <span className="truncate">{settings.branding.siteName || 'Statuspage'}</span>
          </Link>
        </div>
        {/* temporary hidden */}
        {/* <Link
          to="/history"
          className="flex items-center gap-2 rounded-lg bg-emerald-700 px-4 py-2 font-mono text-sm font-medium text-white transition-all hover:bg-emerald-800 dark:bg-emerald-600 dark:hover:bg-emerald-700 sm:px-5"
        >
          <Bell className="h-4 w-4" />
          <span className="hidden sm:inline">Subscribe</span>
        </Link> */}
      </div>
    </header>
  )
}

export function StatusPageFooter({ centerText, showPoweredBy }: { centerText?: string; showPoweredBy?: boolean }) {
  const trimmedCenterText = centerText?.trim() ?? ''

  return (
    <footer className="mt-12 border-t border-slate-200 bg-slate-50 py-8 transition-colors duration-200 dark:border-slate-800 dark:bg-slate-900">
      <div className="mx-auto flex max-w-[768px] flex-col items-center justify-between gap-6 px-4 md:flex-row md:px-8">
        <div className="text-center md:text-left">
          {/* <p className="font-mono text-sm font-bold text-slate-900 dark:text-slate-100">Statora</p> */}
          <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">
            {trimmedCenterText || (showPoweredBy ? 'Powered by Statora Infrastructure' : 'Reliable infrastructure status')}
          </p>
        </div>
        <div className="flex flex-wrap justify-center gap-6">
          <Link to="/history" className="flex items-center gap-2 font-mono text-sm text-slate-600 transition-colors hover:text-cyan-700 dark:text-slate-400 dark:hover:text-cyan-400">
            <FileText className="h-4 w-4" /> History
          </Link>
          {/* <a href="#" className="flex items-center gap-2 font-mono text-sm text-slate-600 transition-colors hover:text-cyan-700 dark:text-slate-400 dark:hover:text-cyan-400">
            <Lock className="h-4 w-4" /> Privacy Policy
          </a>
          <a href="#" className="flex items-center gap-2 font-mono text-sm text-slate-600 transition-colors hover:text-cyan-700 dark:text-slate-400 dark:hover:text-cyan-400">
            <Shield className="h-4 w-4" /> Security
          </a>
          <a href="mailto:support@example.com" className="flex items-center gap-2 font-mono text-sm text-slate-600 transition-colors hover:text-cyan-700 dark:text-slate-400 dark:hover:text-cyan-400">
            <Mail className="h-4 w-4" /> Contact Support
          </a> */}
        </div>
      </div>
    </footer>
  )
}

export function StatusPageFrame({ settings, children }: { settings: StatusPageSettings; children: React.ReactNode }) {
  React.useEffect(() => {
    const mode = settings.theme?.mode || 'light'
    const root = document.documentElement
    if (mode === 'dark') {
      root.classList.add('dark')
    } else if (mode === 'light') {
      root.classList.remove('dark')
    } else if (mode === 'system') {
      const systemDark = window.matchMedia('(prefers-color-scheme: dark)').matches
      if (systemDark) {
        root.classList.add('dark')
      } else {
        root.classList.remove('dark')
      }
    }
  }, [settings.theme?.mode])

  return (
    <div className="flex min-h-screen flex-col bg-slate-50 text-slate-900 transition-colors dark:bg-slate-950 dark:text-slate-100" style={{ fontFamily: 'var(--font-family)' }}>
      <StatusPageHeader settings={settings} />
      <div className="flex-1">{children}</div>
      <StatusPageFooter centerText={settings.footer.text} showPoweredBy={settings.footer.showPoweredBy} />
    </div>
  )
}
