import React, { useEffect, useMemo, useState } from 'react'
import { NavLink, Outlet, useLocation, useNavigate } from 'react-router-dom'
import api from '../../lib/api'
import {
  LayoutDashboard,
  Layers,
  AlertTriangle,
  Wrench,
  Activity,
  Users,
  Shield,
  Settings,
  LogOut,
  ExternalLink,
  User,
  Webhook,
  PanelLeftClose,
  PanelLeftOpen,
  ChevronDown,
  ChevronRight,
} from 'lucide-react'
import type { UserRole } from '../../types'
import type { StatusPageSettings } from '../../types'
import { useApi } from '../../hooks/useApi'

const DEFAULT_PAGE_TITLE = 'Statora'

interface StoredAdminProfile {
  role?: UserRole
}

interface NavChildItem {
  to: string
  label: string
  end: boolean
}

interface NavItem {
  to: string
  label: string
  icon: React.ComponentType<{ className?: string }>
  end: boolean
  children?: NavChildItem[]
}

interface NavSection {
  label: string
  items: NavItem[]
}

function readStoredRole(): UserRole | null {
  try {
    const raw = localStorage.getItem('user_profile') || localStorage.getItem('admin_profile')
    if (!raw) return null
    const parsed = JSON.parse(raw) as StoredAdminProfile
    return parsed.role ?? null
  } catch {
    return null
  }
}

const navSections: NavSection[] = [
  {
    label: 'Monitoring',
    items: [
      { to: '/admin', label: 'Dashboard', icon: LayoutDashboard, end: true },
      { to: '/admin/monitors', label: 'Monitors', icon: Activity, end: false },
      {
        to: '/admin/components',
        label: 'Components',
        icon: Layers,
        end: false,
        children: [{ to: '/admin/subcomponents', label: 'Sub-Components', end: false }],
      },
    ],
  },
  {
    label: 'Operations',
    items: [
      { to: '/admin/incidents', label: 'Incidents', icon: AlertTriangle, end: false },
      { to: '/admin/maintenance', label: 'Maintenance', icon: Wrench, end: false },
    ],
  },
  {
    label: 'Notifications',
    items: [
      { to: '/admin/subscribers', label: 'Subscribers', icon: Users, end: false },
      { to: '/admin/webhook-channels', label: 'Webhook Channels', icon: Webhook, end: false },
    ],
  },
  {
    label: 'System',
    items: [
      { to: '/admin/users', label: 'Users', icon: Shield, end: false },
      { to: '/admin/settings', label: 'Settings', icon: Settings, end: false },
      { to: '/admin/profile', label: 'My Profile', icon: User, end: true },
    ],
  },
]

const OPERATOR_ALLOWED = new Set(['/admin/incidents', '/admin/maintenance'])
const ALWAYS_ALLOWED = new Set(['/admin/profile'])
const SIDEBAR_SECTION_STATE_KEY = 'admin_sidebar_section_state'

function sectionKey(label: string): string {
  return label.toLowerCase().replace(/[^a-z0-9]+/g, '_').replace(/^_+|_+$/g, '')
}

function isRouteActive(pathname: string, to: string, end: boolean): boolean {
  if (end) {
    return pathname === to
  }

  return pathname === to || pathname.startsWith(`${to}/`)
}

function isSectionActive(pathname: string, section: NavSection): boolean {
  return section.items.some(item => {
    if (isRouteActive(pathname, item.to, item.end)) {
      return true
    }

    if (!item.children || item.children.length === 0) {
      return false
    }

    return item.children.some(child => isRouteActive(pathname, child.to, child.end))
  })
}

export default function AdminLayout() {
  const navigate = useNavigate()
  const location = useLocation()
  const [isSidebarCollapsed, setIsSidebarCollapsed] = useState(false)
  const [openSections, setOpenSections] = useState<Record<string, boolean>>({})
  const role = readStoredRole()
  const { data: settingsData } = useApi<StatusPageSettings>('/settings/status-page')
  const pageTitle = settingsData?.head?.title?.trim() || DEFAULT_PAGE_TITLE
  const visibleNavSections: NavSection[] = role === 'operator'
    ? navSections
      .map(section => {
        const items: NavItem[] = []

        section.items.forEach(item => {
          const visibleChildren = item.children?.filter(
            child => OPERATOR_ALLOWED.has(child.to) || ALWAYS_ALLOWED.has(child.to),
          )
          const isItemAllowed = OPERATOR_ALLOWED.has(item.to) || ALWAYS_ALLOWED.has(item.to)

          if (!isItemAllowed && (!visibleChildren || visibleChildren.length === 0)) {
            return
          }

          const nextItem: NavItem = visibleChildren
            ? { ...item, children: visibleChildren }
            : { ...item }

          items.push(nextItem)
        })

        return {
          ...section,
          items,
        }
      })
      .filter(section => section.items.length > 0)
    : navSections

  async function handleLogout() {
    try {
      await api.post('/auth/logout')
    } catch {
      // ignore logout request failures and still clear local state
    }

    localStorage.removeItem('user_token')
    localStorage.removeItem('user_profile')
    localStorage.removeItem('admin_token')
    localStorage.removeItem('admin_profile')
    navigate('/admin/login')
  }

  const sidebarWidthClass = isSidebarCollapsed ? 'w-16' : 'w-64'
  const sidebarOffsetClass = isSidebarCollapsed ? 'ml-16' : 'ml-64'
  const visibleSectionKeys = useMemo(
    () => visibleNavSections.map(section => sectionKey(section.label)),
    [visibleNavSections],
  )

  useEffect(() => {
    let stored: Record<string, boolean> = {}

    try {
      const raw = localStorage.getItem(SIDEBAR_SECTION_STATE_KEY)
      if (raw) {
        const parsed = JSON.parse(raw) as Record<string, boolean>
        if (parsed && typeof parsed === 'object') {
          stored = parsed
        }
      }
    } catch {
      stored = {}
    }

    setOpenSections(prev => {
      const next: Record<string, boolean> = {}

      visibleSectionKeys.forEach(key => {
        if (typeof prev[key] === 'boolean') {
          next[key] = prev[key]
        } else if (typeof stored[key] === 'boolean') {
          next[key] = stored[key]
        } else {
          next[key] = true
        }
      })

      return next
    })
  }, [visibleSectionKeys])

  useEffect(() => {
    if (visibleSectionKeys.length === 0) return

    const hasCompleteSectionState = visibleSectionKeys.every(
      key => typeof openSections[key] === 'boolean',
    )

    if (!hasCompleteSectionState) return

    const next: Record<string, boolean> = {}
    visibleSectionKeys.forEach(key => {
      next[key] = openSections[key] ?? true
    })

    localStorage.setItem(SIDEBAR_SECTION_STATE_KEY, JSON.stringify(next))
  }, [openSections, visibleSectionKeys])

  useEffect(() => {
    setOpenSections(prev => {
      let changed = false
      const next = { ...prev }

      visibleNavSections.forEach(section => {
        const key = sectionKey(section.label)
        if (isSectionActive(location.pathname, section) && next[key] === false) {
          next[key] = true
          changed = true
        }
      })

      return changed ? next : prev
    })
  }, [location.pathname, visibleNavSections])

  useEffect(() => {
    document.title = `${pageTitle} - Admin Panel`
  }, [pageTitle])

  function toggleSection(key: string) {
    setOpenSections(prev => ({
      ...prev,
      [key]: !(prev[key] ?? true),
    }))
  }

  return (
    <div className="min-h-screen bg-slate-50 text-slate-900 dark:bg-slate-950 dark:text-slate-100">
      {/* Sidebar */}
      <aside
        className={`fixed inset-y-0 left-0 z-30 ${sidebarWidthClass} flex flex-col overflow-hidden border-r border-slate-200 bg-white/90 text-slate-600 shadow-sm backdrop-blur-md transition-[width] duration-300 ease-in-out dark:border-slate-800 dark:bg-slate-900/90 dark:text-slate-400`}
      >
        <div className="border-b border-slate-200 bg-white/80 backdrop-blur-sm dark:border-slate-800 dark:bg-slate-900/80">
          <div className="flex items-start justify-between gap-2 px-4 py-5 min-h-[80px]">
            <div className={`min-w-0 transition-opacity duration-200 ${isSidebarCollapsed ? 'opacity-0 pointer-events-none' : 'opacity-100'}`}>
              <h1 className="truncate text-base font-bold tracking-wide text-emerald-700 dark:text-emerald-400">{pageTitle}</h1>
              <p className="mt-1 truncate font-mono text-[11px] font-semibold uppercase tracking-wider text-slate-500 dark:text-slate-400">Operations Console</p>
            </div>
            <button
              type="button"
              onClick={() => setIsSidebarCollapsed(prev => !prev)}
              aria-label={isSidebarCollapsed ? 'Expand sidebar' : 'Collapse sidebar'}
              className="inline-flex h-8 w-8 items-center justify-center rounded-full text-slate-500 transition-all duration-200 hover:bg-slate-100 hover:text-slate-900 dark:text-slate-400 dark:hover:bg-slate-800 dark:hover:text-slate-100"
            >
              {isSidebarCollapsed ? <PanelLeftOpen className="h-[18px] w-[18px]" /> : <PanelLeftClose className="h-[18px] w-[18px]" />}
            </button>
          </div>
        </div>

        <nav className="flex-1 min-h-0 overflow-y-auto px-3 pt-5 pb-8 [scrollbar-width:thin] [scrollbar-color:rgb(148_163_184)_transparent] [&::-webkit-scrollbar]:w-1.5 [&::-webkit-scrollbar-track]:bg-transparent [&::-webkit-scrollbar-thumb]:rounded-full [&::-webkit-scrollbar-thumb]:bg-slate-300 hover:[&::-webkit-scrollbar-thumb]:bg-slate-400 dark:[scrollbar-color:rgb(51_65_85)_transparent] dark:[&::-webkit-scrollbar-thumb]:bg-slate-700/50 dark:hover:[&::-webkit-scrollbar-thumb]:bg-slate-600/80">
          {visibleNavSections.map(section => (
            <div key={section.label} className="mb-6 last:mb-0 space-y-1">
              {!isSidebarCollapsed && (
                <button
                  type="button"
                  onClick={() => toggleSection(sectionKey(section.label))}
                  aria-expanded={openSections[sectionKey(section.label)] ?? true}
                  className="group flex w-full items-center justify-between px-3 py-2 font-mono text-[10px] font-bold uppercase tracking-[0.15em] text-slate-500 transition-colors hover:text-emerald-700 dark:text-slate-500 dark:hover:text-emerald-400"
                >
                  <span>{section.label}</span>
                  {(openSections[sectionKey(section.label)] ?? true) ? (
                    <ChevronDown className="h-3 w-3 opacity-50 group-hover:opacity-100 transition-opacity" aria-hidden="true" />
                  ) : (
                    <ChevronRight className="h-3 w-3 opacity-50 group-hover:opacity-100 transition-opacity" aria-hidden="true" />
                  )}
                </button>
              )}

              <div
                className={`grid transition-[grid-template-rows,opacity] duration-300 ease-in-out ${(openSections[sectionKey(section.label)] ?? true) || isSidebarCollapsed ? 'grid-rows-[1fr] opacity-100' : 'grid-rows-[0fr] opacity-0'}`}
              >
                <div className="overflow-hidden space-y-0.5">
                  {section.items.map(({ to, label, icon: Icon, end, children }) => (
                    <div key={to} className="space-y-0.5">
                      <NavLink
                        to={to}
                        end={end}
                        title={isSidebarCollapsed ? label : undefined}
                        className={({ isActive }) =>
                          `flex items-center ${isSidebarCollapsed ? 'justify-center px-0 w-10 mx-auto' : 'gap-3 px-3'} py-2 rounded-lg font-mono text-[13px] leading-5 font-semibold transition-all duration-200 ${isActive
                            ? 'bg-emerald-100 text-emerald-800 shadow-[inset_2px_0_0_0_rgb(4,120,87)] dark:bg-emerald-950/40 dark:text-emerald-400 dark:shadow-[inset_2px_0_0_0_rgb(52,211,153)]'
                            : 'border-transparent text-slate-600 hover:bg-slate-100 hover:text-emerald-700 dark:text-slate-400 dark:hover:bg-slate-800/60 dark:hover:text-emerald-400'
                          }`
                        }
                      >
                        <Icon className={`w-[18px] h-[18px] flex-shrink-0 ${isSidebarCollapsed ? '' : 'opacity-80'}`} />
                        {!isSidebarCollapsed && <span className="truncate">{label}</span>}
                      </NavLink>

                      {!isSidebarCollapsed && children && children.length > 0 && (
                        <div className="ml-[22px] mt-0.5 space-y-0.5 border-l border-slate-200 pl-3 dark:border-slate-800">
                          {children.map(child => (
                            <NavLink
                              key={child.to}
                              to={child.to}
                              end={child.end}
                              className={({ isActive }) =>
                                `block rounded-md px-3 py-1.5 font-mono text-[13px] font-semibold leading-5 transition-colors ${isActive
                                  ? 'bg-emerald-100 text-emerald-800 dark:bg-emerald-950/40 dark:text-emerald-400'
                                  : 'text-slate-500 hover:bg-slate-100 hover:text-emerald-700 dark:hover:bg-slate-800/40 dark:hover:text-emerald-400'
                                }`
                              }
                            >
                              {child.label}
                            </NavLink>
                          ))}
                        </div>
                      )}
                    </div>
                  ))}
                </div>
              </div>
            </div>
          ))}
        </nav>

        <footer className="space-y-1 border-t border-slate-200 bg-slate-50/80 p-4 dark:border-slate-800 dark:bg-slate-900/80">
          <a
            href="/"
            target="_blank"
            rel="noopener noreferrer"
            title={isSidebarCollapsed ? 'View Status Page' : undefined}
            className={`flex items-center ${isSidebarCollapsed ? 'justify-center px-0 w-10 mx-auto' : 'gap-3 px-3'} rounded-lg py-2 font-mono text-[13px] font-semibold text-slate-600 transition-colors hover:bg-slate-100 hover:text-emerald-700 dark:text-slate-400 dark:hover:bg-slate-800/60 dark:hover:text-emerald-400`}
          >
            <ExternalLink className="w-[18px] h-[18px] opacity-80" />
            {!isSidebarCollapsed && 'Status Page'}
          </a>
          <button
            type="button"
            onClick={handleLogout}
            title={isSidebarCollapsed ? 'Logout' : undefined}
            className={`flex w-full items-center ${isSidebarCollapsed ? 'justify-center px-0 w-10 mx-auto' : 'gap-3 px-3'} rounded-lg py-2 font-mono text-[13px] font-semibold text-slate-600 transition-colors hover:bg-rose-50 hover:text-rose-600 dark:text-slate-400 dark:hover:bg-rose-950/30 dark:hover:text-rose-400`}
          >
            <LogOut className="w-[18px] h-[18px] opacity-80" />
            {!isSidebarCollapsed && 'Logout'}
          </button>
        </footer>
      </aside>

      {/* Main content */}
      <main className={`${sidebarOffsetClass} h-screen overflow-auto transition-[margin] duration-300 ease-in-out`}>
        <div className="max-w-7xl mx-auto p-6 md:p-10 lg:p-12">
          <Outlet />
        </div>
      </main>
    </div>
  )
}
