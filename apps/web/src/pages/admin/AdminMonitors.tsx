import React, { useState } from 'react'
import { Plus, Trash2, X, Activity, Edit2, CheckCircle2, XCircle, AlertCircle } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { useApi } from '../../hooks/useApi'
import { useAdminPagination } from '../../hooks/useAdminPagination'
import api from '../../lib/api'
import { getApiErrorMessage } from '../../lib/apiError'
import type { Monitor, Component, SubComponent, MonitorType } from '../../types'
import Modal from '../../components/Modal'
import AdminPaginationControls from '../../components/AdminPaginationControls'
import { AdminListCard, AdminTableEmptyRow } from '../../components/AdminTableShell'

const MONITOR_TYPES: MonitorType[] = ['http', 'tcp', 'dns', 'ping', 'ssl']

interface FormState {
  name: string
  type: MonitorType
  target: string
  sslThresholds: string
  domainExpiry: boolean
  certExpiry: boolean
  ignoreTLSError: boolean
  intervalSeconds: number
  timeoutSeconds: number
  componentId: string
  subComponentId: string
}


const DEFAULT_FORM: FormState = {
  name: '',
  type: 'http',
  target: '',
  sslThresholds: '30,14,7',
  domainExpiry: false,
  certExpiry: false,
  ignoreTLSError: false,
  intervalSeconds: 60,
  timeoutSeconds: 10,
  componentId: '',
  subComponentId: '',
}


const TYPE_PLACEHOLDERS: Record<MonitorType, string> = {
  http: 'https://example.com/health',
  tcp: 'example.com:443',
  dns: 'example.com',
  ping: 'example.com',
  ssl: 'example.com:443',
}

export function monitorLogsPath(id: string): string {
  return `/admin/monitors/${id}/logs`
}

export default function AdminMonitors() {
  const navigate = useNavigate()
  const { page, limit, apiParams, setPage, setLimit } = useAdminPagination()
  const { data: monitors, total: totalMonitors, page: currentPage, totalPages, loading: monitorsLoading, error: monitorsError, refetch } = useApi<Monitor[]>('/monitors', [], apiParams)
  const { data: components, total: totalComponents, loading: componentsLoading, error: componentsError } = useApi<Component[]>('/components', [], { page: 1, limit: 10 })
  const { data: subcomponents, total: totalSubcomponents, loading: subcomponentsLoading, error: subcomponentsError } = useApi<SubComponent[]>('/subcomponents', [], { page: 1, limit: 10 })
  const [showModal, setShowModal] = useState(false)
  const [editingId, setEditingId] = useState<string | null>(null)
  const [form, setForm] = useState<FormState>(DEFAULT_FORM)
  const [saving, setSaving] = useState(false)
  const [testing, setTesting] = useState(false)
  const [testResult, setTestResult] = useState<{ status: string; responseTime: number; sslWarning?: boolean; sslDaysRemaining?: number; sslTriggeredThreshold?: number; domainWarning?: boolean; domainDaysRemaining?: number; domainTriggeredThreshold?: number; } | null>(null)
  const [error, setError] = useState('')

  const supportsDomainExpiry = form.type === 'http' || form.type === 'ssl'
  const supportsCertExpiry = form.type === 'http' || form.type === 'ssl'
  const supportsIgnoreTLSError = form.type === 'http'
  const needsThresholds = form.type === 'ssl' || form.domainExpiry || form.certExpiry

  function openCreate() {
    setEditingId(null)
    setForm({ ...DEFAULT_FORM, componentId: (components || [])[0]?.id || '', subComponentId: '' })
    setError('')
    setTestResult(null)
    setShowModal(true)
  }

  function openEdit(m: Monitor) {
    setEditingId(m.id || null)
    setForm({
      name: m.name,
      type: m.type,
      target: m.target,
      sslThresholds: (m.sslThresholds && m.sslThresholds.length > 0) ? m.sslThresholds.join(',') : '30,14,7',
      domainExpiry: Boolean(m.monitoring?.advanced?.domain_expiry),
      certExpiry: Boolean(m.monitoring?.advanced?.cert_expiry),
      ignoreTLSError: Boolean(m.monitoring?.advanced?.ignore_tls_error),
      intervalSeconds: m.intervalSeconds,
      timeoutSeconds: m.timeoutSeconds,
      componentId: m.componentId || '',
      subComponentId: m.subComponentId || '',
    })
    setError('')
    setTestResult(null)
    setShowModal(true)
  }

  function closeModal() {
    setShowModal(false)
    setEditingId(null)
    setTestResult(null)
  }

  async function handleTest() {
    setTesting(true)
    setTestResult(null)
    setError('')
    try {
      const res = await api.post('/monitors/test', {
        ...form,
        sslThresholds:
          form.type === 'ssl' || form.domainExpiry || form.certExpiry
            ? form.sslThresholds.split(',').map(v => parseInt(v.trim(), 10)).filter(v => Number.isFinite(v) && v > 0)
            : undefined,
        monitoring: {
          advanced: {
            domain_expiry: form.domainExpiry,
            cert_expiry: form.certExpiry,
            ignore_tls_error: form.ignoreTLSError,
          },
        },
      })
      setTestResult(res.data)
    } catch (err: unknown) {
      setError(getApiErrorMessage(err, 'Failed to test monitor'))
    } finally {
      setTesting(false)
    }
  }

  async function handleSave(e: React.FormEvent) {
    e.preventDefault()
    setSaving(true)
    setError('')
    try {
      const payload = {
        ...form,
        sslThresholds:
          form.type === 'ssl' || form.domainExpiry || form.certExpiry
            ? form.sslThresholds.split(',').map(v => parseInt(v.trim(), 10)).filter(v => Number.isFinite(v) && v > 0)
            : undefined,
        monitoring: {
          advanced: {
            domain_expiry: form.domainExpiry,
            cert_expiry: form.certExpiry,
            ignore_tls_error: form.ignoreTLSError,
          },
        },
      }

      if (editingId) {
        await api.put(`/monitors/${editingId}`, payload)
      } else {
        await api.post('/monitors', payload)
      }
      await refetch()
      closeModal()
    } catch (err: unknown) {
      setError(getApiErrorMessage(err, `Failed to ${editingId ? 'update' : 'create'} monitor`))
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete(m: Monitor) {
    if (!confirm(`Delete monitor "${m.name}"?`)) return
    try {
      await api.delete(`/monitors/${m.id}`)
      await refetch()
    } catch (err: unknown) {
      alert(getApiErrorMessage(err, 'Failed to delete'))
    }
  }

  function getComponentNameAndSubcomponentText(monitor: Monitor) {
    if (monitor.subComponentId) {
      // Monitor is associated with a subcomponent
      const subcomponent = (subcomponents || []).find(sc => sc.id === monitor.subComponentId);
      const component = (components || []).find(c => c.id === monitor.componentId);
      if (subcomponent && component) {
        return `${subcomponent.name} (Subcomponent of ${component.name})`;
      } else if (subcomponent) {
        return `${subcomponent.name} (SubComponent)`;
      } else {
        return `SubComponent: ${monitor.subComponentId}`; // Not found in the list
      }
    } else if (monitor.componentId) {
      // Traditional component-only case
      return (components || []).find(c => c.id === monitor.componentId)?.name || monitor.componentId;
    }
    return 'Unassigned';
  }

  return (
    <div className="p-8">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-slate-100">Monitors</h1>
          <p className="text-sm text-gray-500 dark:text-slate-400 mt-1">{totalMonitors ?? 0} active monitors</p>
        </div>
        <button
          onClick={openCreate}
          className="flex items-center gap-2 bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium px-4 py-2 rounded-lg transition-colors"
        >
          <Plus className="w-4 h-4" /> Add Monitor
        </button>
      </div>

       <AdminListCard>
         <table className="w-full text-sm">
          <thead className="bg-gray-50 border-b border-gray-100 dark:bg-slate-800/50 dark:border-slate-700">
            <tr>
              <th className="text-left px-6 py-3 font-medium text-gray-600 dark:text-slate-400">Name</th>
              <th className="text-left px-6 py-3 font-medium text-gray-600 dark:text-slate-400">Type</th>
              <th className="text-left px-6 py-3 font-medium text-gray-600 dark:text-slate-400">Target</th>
              <th className="text-left px-6 py-3 font-medium text-gray-600 dark:text-slate-400">Status</th>
              <th className="text-left px-6 py-3 font-medium text-gray-600 dark:text-slate-400">Component</th>
              <th className="text-left px-6 py-3 font-medium text-gray-600 dark:text-slate-400">Interval</th>
              <th className="px-6 py-3" />
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-50 dark:divide-slate-800">
            {(monitors || []).map(m => (
              <tr
                key={m.id}
                className="cursor-pointer hover:bg-gray-50 dark:hover:bg-slate-800/50"
                onClick={() => navigate(monitorLogsPath(m.id))}
              >
                <td className="px-6 py-4 font-medium text-gray-900 dark:text-slate-100">{m.name}</td>
                <td className="px-6 py-4">
                  <span className="flex items-center gap-1.5">
                    <Activity className="w-3.5 h-3.5 text-purple-500 dark:text-purple-400" />
                    <span className="uppercase text-xs font-medium text-purple-700 dark:text-purple-400">{m.type}</span>
                  </span>
                </td>
                <td className="px-6 py-4 text-gray-500 dark:text-slate-400 max-w-xs truncate font-mono text-xs">{m.target}</td>
                <td className="px-6 py-4">
                  {m.lastStatus === 'up' ? (
                    <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium bg-green-50 text-green-700 border border-green-200 dark:bg-green-950/40 dark:text-green-400 dark:border-green-700/40">
                      <CheckCircle2 className="w-3.5 h-3.5" />
                      Up
                    </span>
                  ) : m.lastStatus === 'down' ? (
                    <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium bg-red-50 text-red-700 border border-red-200 dark:bg-red-950/40 dark:text-red-400 dark:border-red-700/40">
                      <XCircle className="w-3.5 h-3.5" />
                      Down
                    </span>
                  ) : (
                    <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium bg-gray-50 text-gray-600 border border-gray-200 dark:bg-slate-800 dark:text-slate-400 dark:border-slate-600">
                      <AlertCircle className="w-3.5 h-3.5" />
                      Pending
                    </span>
                  )}
                </td>
                <td className="px-6 py-4 text-gray-500 dark:text-slate-400">{getComponentNameAndSubcomponentText(m)}</td>
                <td className="px-6 py-4 text-gray-500 dark:text-slate-400">{m.intervalSeconds}s</td>
                <td className="px-6 py-4">
                  <div className="flex items-center justify-end gap-2">
                    <button
                      onClick={(e) => {
                        e.stopPropagation()
                        openEdit(m)
                      }}
                      className="text-gray-400 hover:text-blue-600 transition-colors dark:text-slate-500 dark:hover:text-blue-400"
                    >
                      <Edit2 className="w-4 h-4" />
                    </button>
                    <button
                      onClick={(e) => {
                        e.stopPropagation()
                        void handleDelete(m)
                      }}
                      className="text-gray-400 hover:text-red-600 transition-colors dark:text-slate-500 dark:hover:text-red-400"
                    >
                      <Trash2 className="w-4 h-4" />
                    </button>
                  </div>
                </td>
              </tr>
            ))}
             {(monitors || []).length === 0 && (
               <AdminTableEmptyRow colSpan={7}>
                 No monitors configured. Add one to start tracking uptime.
               </AdminTableEmptyRow>
             )}
          </tbody>
        </table>

        <AdminPaginationControls
          page={currentPage || page}
          totalPages={totalPages}
          total={totalMonitors}
          limit={limit}
          loading={monitorsLoading}
          onPageChange={setPage}
           onLimitChange={setLimit}
         />
       </AdminListCard>

      {showModal && (
        <Modal
          title={editingId ? 'Edit Monitor' : 'New Monitor'}
          onClose={closeModal}
          footer={(
            <div className="flex gap-3">
              <button type="button" onClick={closeModal} className="flex-1 border border-gray-300 text-gray-700 rounded-lg py-2 text-sm hover:bg-gray-50 dark:border-slate-600 dark:text-slate-300 dark:hover:bg-slate-800">
                Cancel
              </button>
              <button type="button" onClick={handleTest} disabled={testing || !form.target} className="flex-1 border border-blue-200 text-blue-700 bg-blue-50 hover:bg-blue-100 disabled:opacity-60 rounded-lg py-2 text-sm font-medium dark:border-blue-700/40 dark:text-blue-400 dark:bg-blue-950/30 dark:hover:bg-blue-950/50">
                {testing ? 'Testing...' : 'Test Target'}
              </button>
              <button type="submit" form="monitor-form" disabled={saving} className="flex-1 bg-blue-600 hover:bg-blue-700 disabled:opacity-60 text-white rounded-lg py-2 text-sm font-medium">
                {saving ? 'Saving...' : (editingId ? 'Update Monitor' : 'Create Monitor')}
              </button>
            </div>
          )}
        >
          {error && <p className="mb-4 text-sm text-red-600 bg-red-50 rounded-lg px-3 py-2 dark:text-red-400 dark:bg-red-950/30">{error}</p>}
          <form id="monitor-form" onSubmit={handleSave} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1">Name</label>
              <input
                required
                value={form.name}
                onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
                className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:border-slate-600 dark:bg-slate-800 dark:text-slate-200 dark:placeholder:text-slate-500"
                placeholder="API Health Check"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1">Type</label>
              <select
                value={form.type}
                onChange={e =>
                  setForm(f => {
                    const nextType = e.target.value as MonitorType
                    const canUseDomainExpiry = nextType === 'http' || nextType === 'ssl'
                    const canUseCertExpiry = nextType === 'http' || nextType === 'ssl'
                    const canUseIgnoreTLSError = nextType === 'http'

                    const nextDomainExpiry = canUseDomainExpiry ? f.domainExpiry : false
                    const nextCertExpiry = canUseCertExpiry ? f.certExpiry : false
                    let nextIgnoreTLSError = canUseIgnoreTLSError ? f.ignoreTLSError : false

                    if (nextCertExpiry && nextIgnoreTLSError) {
                      nextIgnoreTLSError = false
                    }

                    return {
                      ...f,
                      type: nextType,
                      target: '',
                      domainExpiry: nextDomainExpiry,
                      certExpiry: nextCertExpiry,
                      ignoreTLSError: nextIgnoreTLSError,
                    }
                  })
                }
                className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:border-slate-600 dark:bg-slate-800 dark:text-slate-200"
              >
                {MONITOR_TYPES.map(t => (
                  <option key={t} value={t}>{t.toUpperCase()}</option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1">Target</label>
              <input
                required
                value={form.target}
                onChange={e => setForm(f => ({ ...f, target: e.target.value }))}
                className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 font-mono dark:border-slate-600 dark:bg-slate-800 dark:text-slate-200 dark:placeholder:text-slate-500"
                placeholder={TYPE_PLACEHOLDERS[form.type]}
              />
            </div>
            {needsThresholds && (
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1">Alert Thresholds (days)</label>
                <input
                  value={form.sslThresholds}
                  onChange={e => setForm(f => ({ ...f, sslThresholds: e.target.value }))}
                  className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 font-mono dark:border-slate-600 dark:bg-slate-800 dark:text-slate-200 dark:placeholder:text-slate-500"
                  placeholder="30,14,7"
                />
                <p className="text-xs text-gray-500 dark:text-slate-400 mt-1">
                  Used for certificate and domain expiry warnings (highest threshold first).
                </p>
              </div>
            )}

            {(supportsDomainExpiry || supportsCertExpiry || supportsIgnoreTLSError) && (
              <div className="rounded-lg border border-gray-200 bg-gray-50 p-3 space-y-3 dark:border-slate-700 dark:bg-slate-800/50">
                <div>
                  <p className="text-sm font-medium text-gray-800 dark:text-slate-200">Advanced Monitoring</p>
                  <p className="text-xs text-gray-500 dark:text-slate-400 mt-0.5">
                    Enable optional checks depending on monitor type compatibility.
                  </p>
                </div>

                {supportsDomainExpiry && (
                  <label className="flex items-start gap-2 text-sm text-gray-700 dark:text-slate-300">
                    <input
                      type="checkbox"
                      checked={form.domainExpiry}
                      onChange={e => setForm(f => ({ ...f, domainExpiry: e.target.checked }))}
                      className="mt-0.5"
                    />
                    <span>Domain expiry monitoring</span>
                  </label>
                )}

                {supportsCertExpiry && (
                  <label className="flex items-start gap-2 text-sm text-gray-700 dark:text-slate-300">
                    <input
                      type="checkbox"
                      checked={form.certExpiry}
                      onChange={e =>
                        setForm(f => ({
                          ...f,
                          certExpiry: e.target.checked,
                          ignoreTLSError: e.target.checked ? false : f.ignoreTLSError,
                        }))
                      }
                      className="mt-0.5"
                    />
                    <span>Certificate expiry monitoring</span>
                  </label>
                )}

                {supportsIgnoreTLSError && (
                  <label className="flex items-start gap-2 text-sm text-gray-700 dark:text-slate-300">
                    <input
                      type="checkbox"
                      checked={form.ignoreTLSError}
                      disabled={form.certExpiry}
                      onChange={e =>
                        setForm(f => ({
                          ...f,
                          ignoreTLSError: e.target.checked,
                          certExpiry: e.target.checked ? false : f.certExpiry,
                        }))
                      }
                      className="mt-0.5"
                    />
                    <span>Ignore TLS certificate errors (HTTP only)</span>
                  </label>
                )}

                {form.type === 'http' && (form.certExpiry || form.ignoreTLSError) && (
                  <p className="text-xs text-amber-700 bg-amber-50 border border-amber-200 rounded px-2 py-1 dark:text-amber-400 dark:bg-amber-950/30 dark:border-amber-700/40">
                    HTTP target must use <span className="font-mono">https://</span> when certificate expiry or ignore TLS errors is enabled.
                  </p>
                )}
              </div>
            )}
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1">Component / Subcomponent</label>
              <select
                value={form.componentId}
                onChange={e => setForm(f => ({ ...f, componentId: e.target.value, subComponentId: '' }))} // Clear subComponentId when component changes
                className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 mb-2 dark:border-slate-600 dark:bg-slate-800 dark:text-slate-200"
              >
                {(components || []).map(c => (
                  <option key={c.id} value={c.id}>{c.name}</option>
                ))}
              </select>

              {/* Subcomponent selection */}
              {form.componentId && (
                <select
                  value={form.subComponentId}
                  onChange={e => setForm(f => ({ ...f, subComponentId: e.target.value }))}
                  className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:border-slate-600 dark:bg-slate-800 dark:text-slate-200"
                >
                  <option value="">Select subcomponent (optional)...</option>
                  {(subcomponents || [])
                    .filter(sc => sc.componentId === form.componentId)
                    .map(sc => (
                      <option key={sc.id} value={sc.id}>{sc.name}</option>
                    ))
                  }
                </select>
              )}
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1">Interval (seconds)</label>
                <input
                  type="number"
                  min={10}
                  max={3600}
                  value={form.intervalSeconds}
                  onChange={e => setForm(f => ({ ...f, intervalSeconds: parseInt(e.target.value) }))}
                  className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:border-slate-600 dark:bg-slate-800 dark:text-slate-200"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1">Timeout (seconds)</label>
                <input
                  type="number"
                  min={1}
                  max={60}
                  value={form.timeoutSeconds}
                  onChange={e => setForm(f => ({ ...f, timeoutSeconds: parseInt(e.target.value) }))}
                  className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 dark:border-slate-600 dark:bg-slate-800 dark:text-slate-200"
                />
              </div>
            </div>
            {testResult && (
              <div className={`px-3 py-2 rounded-lg text-sm flex items-center justify-between ${testResult.status === 'up' ? 'bg-green-50 text-green-700 border border-green-200 dark:bg-green-950/40 dark:text-green-400 dark:border-green-700/40' : 'bg-red-50 text-red-700 border border-red-200 dark:bg-red-950/40 dark:text-red-400 dark:border-red-700/40'}`}>
                <span className="font-medium">
                  Test Result: {testResult.status.toUpperCase()}
                  {testResult.sslWarning || testResult.domainWarning ? ' (WARNING)' : ''}
                </span>
                <span>
                  {typeof testResult.domainDaysRemaining === 'number'
                    ? `Domain: ${testResult.domainDaysRemaining}d left`
                    : typeof testResult.sslDaysRemaining === 'number'
                      ? `Cert: ${testResult.sslDaysRemaining}d left`
                      : `${testResult.responseTime}ms`}
                </span>
              </div>
            )}
          </form>
        </Modal>
      )}
    </div>
  )
}
