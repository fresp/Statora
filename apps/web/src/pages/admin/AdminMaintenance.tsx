import React, { useState } from 'react'
import { Plus, Pencil, Trash2, Eye } from 'lucide-react'
import { useApi } from '../../hooks/useApi'
import { useAdminPagination } from '../../hooks/useAdminPagination'
import api from '../../lib/api'
import { getApiErrorMessage } from '../../lib/apiError'
import type { Maintenance, Component, MaintenanceStatus, IncidentVisibilityState, RichTextDocument } from '../../types'
import { formatDate } from '../../lib/utils'
import { getMaintenanceContent, getPlainTextFromRichText, normalizeMaintenanceStatus } from '../../lib/contentModel'
import Modal from '../../components/Modal'
import AdminPaginationControls from '../../components/AdminPaginationControls'
import { AdminListCard, AdminTableEmptyRow, textOrEmDash } from '../../components/AdminTableShell'
import RichTextEditor from '../../components/editor/RichTextEditor'
import ContentRenderer from '../../components/content/ContentRenderer'

const STATUSES: MaintenanceStatus[] = ['scheduled', 'in_progress', 'completed']

const STATUS_LABELS: Record<MaintenanceStatus, string> = {
  scheduled: 'Scheduled',
  in_progress: 'In Progress',
  completed: 'Completed',
  draft: 'Draft',
  active: 'Active',
}

const STATUS_COLORS: Record<MaintenanceStatus, string> = {
  scheduled: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400',
  in_progress: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400',
  completed: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400',
  draft: 'bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-400',
  active: 'bg-emerald-100 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400',
}

const VISIBILITY_LABELS: Record<IncidentVisibilityState, string> = {
  draft: 'Draft',
  published: 'Published',
}

const VISIBILITY_COLORS: Record<IncidentVisibilityState, string> = {
  draft: 'bg-slate-100 text-slate-600 dark:bg-slate-800 dark:text-slate-400',
  published: 'bg-emerald-50 text-emerald-700 dark:bg-emerald-900/30 dark:text-emerald-400',
}

interface FormState {
  title: string
  description: string
  descriptionJson?: RichTextDocument
  components: string[]
  startTime: string
  endTime: string
  status: MaintenanceStatus
  visibilityState: IncidentVisibilityState
}

const DEFAULT_FORM: FormState = {
  title: '',
  description: '',
  components: [],
  startTime: '',
  endTime: '',
  status: 'scheduled',
  visibilityState: 'published',
}

function toDatetimeLocal(iso: string) {
  if (!iso) return ''
  return iso.slice(0, 16)
}

export default function AdminMaintenance() {
  const { page, limit, apiParams, setPage, setLimit } = useAdminPagination()
  const { data: maintenance, total: totalMaintenance, totalPages, loading, refetch } = useApi<Maintenance[]>('/maintenance', [], apiParams)
  const { data: components } = useApi<Component[]>('/components', [], { page: 1, limit: 100 })
  const [showModal, setShowModal] = useState(false)
  const [editing, setEditing] = useState<Maintenance | null>(null)
  const [form, setForm] = useState<FormState>(DEFAULT_FORM)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  const [viewing, setViewing] = useState<Maintenance | null>(null)
  const [deleting, setDeleting] = useState<Maintenance | null>(null)

  function openCreate() {
    setEditing(null)
    setForm(DEFAULT_FORM)
    setError('')
    setShowModal(true)
  }

  function openEdit(m: Maintenance) {
    setEditing(m)
    setForm({
      title: m.title,
      description: m.description,
      descriptionJson: m.descriptionJson,
      components: m.components,
      startTime: toDatetimeLocal(m.startTime),
      endTime: toDatetimeLocal(m.endTime),
      status: normalizeMaintenanceStatus(m.status),
      visibilityState: m.visibilityState || 'published',
    })
    setError('')
    setShowModal(true)
  }

  function closeModal() {
    setShowModal(false)
    setEditing(null)
  }

  function toggleComponent(id: string) {
    setForm((f) => ({
      ...f,
      components: f.components.includes(id) ? f.components.filter((c) => c !== id) : [...f.components, id],
    }))
  }

  async function handleSave(e: React.FormEvent) {
    e.preventDefault()
    setSaving(true)
    setError('')
    try {
      const payload = {
        ...form,
        startTime: new Date(form.startTime).toISOString(),
        endTime: new Date(form.endTime).toISOString(),
        description: getPlainTextFromRichText(form.descriptionJson) || form.description,
        descriptionJson: form.descriptionJson,
      }
      if (editing) {
        await api.patch(`/maintenance/${editing.id}`, payload)
      } else {
        await api.post('/maintenance', payload)
      }
      await refetch()
      closeModal()
    } catch (err: unknown) {
      setError(getApiErrorMessage(err, 'Failed to save'))
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete() {
    if (!deleting) return
    setSaving(true)
    try {
      await api.delete(`/maintenance/${deleting.id}`)
      setDeleting(null)
      await refetch()
    } catch (err: unknown) {
      alert(getApiErrorMessage(err, 'Failed to delete maintenance'))
    } finally {
      setSaving(false)
    }
  }

  const previewContent = getMaintenanceContent(form)

  return (
    <div className="p-8">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-slate-100">Maintenance</h1>
          <p className="text-sm text-gray-500 dark:text-slate-400 mt-1">{totalMaintenance} total</p>
        </div>
        <button
          onClick={openCreate}
          className="flex items-center gap-2 bg-blue-600 hover:bg-blue-700 text-white text-sm font-medium px-4 py-2 rounded-lg transition-colors"
        >
          <Plus className="w-4 h-4" /> Schedule Maintenance
        </button>
      </div>

      <AdminListCard>
        <table className="w-full text-sm">
          <thead className="bg-gray-50 border-b border-gray-100 dark:bg-slate-800/50 dark:border-slate-700">
            <tr>
              <th className="text-left px-6 py-3 font-medium text-gray-600 dark:text-slate-300">Title</th>
              <th className="text-left px-6 py-3 font-medium text-gray-600 dark:text-slate-300">Status</th>
              <th className="text-left px-6 py-3 font-medium text-gray-600 dark:text-slate-300">Visibility</th>
              <th className="text-left px-6 py-3 font-medium text-gray-600 dark:text-slate-300">Start</th>
              <th className="text-left px-6 py-3 font-medium text-gray-600 dark:text-slate-300">End</th>
              <th className="text-left px-6 py-3 font-medium text-gray-600 dark:text-slate-300">Creator</th>
              <th className="px-6 py-3" />
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-50 dark:divide-slate-800/50">
            {(maintenance || []).map((m) => {
              const normalizedStatus = normalizeMaintenanceStatus(m.status)
              return (
                <tr key={m.id} className="hover:bg-gray-50 dark:hover:bg-slate-800/50">
                  <td className="px-6 py-4 font-medium text-gray-900 dark:text-slate-100">{m.title}</td>
                  <td className="px-6 py-4">
                    <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${STATUS_COLORS[normalizedStatus]}`}>
                      {STATUS_LABELS[normalizedStatus]}
                    </span>
                  </td>
                  <td className="px-6 py-4">
                    <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${VISIBILITY_COLORS[m.visibilityState || 'published']}`}>
                      {VISIBILITY_LABELS[m.visibilityState || 'published']}
                    </span>
                  </td>
                  <td className="px-6 py-4 text-gray-500 dark:text-slate-400">{formatDate(m.startTime)}</td>
                  <td className="px-6 py-4 text-gray-500 dark:text-slate-400">{formatDate(m.endTime)}</td>
                  <td className="px-6 py-4 text-sm text-gray-500 dark:text-slate-400">{textOrEmDash(m.creatorUsername)}</td>
                  <td className="px-6 py-4">
                    <div className="flex items-center justify-end gap-2">
                      <button
                        onClick={() => setViewing(m)}
                        className="text-gray-400 dark:text-slate-500 hover:text-emerald-600 dark:hover:text-emerald-400 transition-colors"
                        title="View"
                      >
                        <Eye className="w-4 h-4" />
                      </button>
                      <button
                        onClick={() => openEdit(m)}
                        className="text-gray-400 dark:text-slate-500 hover:text-blue-600 dark:hover:text-blue-400 transition-colors"
                        title="Edit"
                      >
                        <Pencil className="w-4 h-4" />
                      </button>
                      <button
                        onClick={() => setDeleting(m)}
                        className="text-gray-400 dark:text-slate-500 hover:text-red-600 dark:hover:text-red-400 transition-colors"
                        title="Delete"
                      >
                        <Trash2 className="w-4 h-4" />
                      </button>
                    </div>
                  </td>
                </tr>
              )
            })}
            {(maintenance || []).length === 0 && (
              <AdminTableEmptyRow colSpan={7}>No maintenance windows scheduled.</AdminTableEmptyRow>
            )}
          </tbody>
        </table>

        <AdminPaginationControls
          page={page}
          totalPages={totalPages}
          total={totalMaintenance}
          limit={limit}
          loading={loading}
          onPageChange={setPage}
          onLimitChange={setLimit}
        />
      </AdminListCard>

      {showModal && (
        <Modal title={editing ? 'Edit Maintenance' : 'Schedule Maintenance'} onClose={closeModal} size="lg">
          {error && <p className="mb-4 text-sm text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-900/20 rounded-lg px-3 py-2">{error}</p>}
          <form onSubmit={handleSave} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1">Title</label>
              <input
                required
                value={form.title}
                onChange={(e) => setForm((f) => ({ ...f, title: e.target.value }))}
                className="w-full border border-gray-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-slate-100 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1">Description</label>
              <RichTextEditor
                value={form.descriptionJson || form.description}
                onChange={(json) => setForm((f) => ({ ...f, descriptionJson: json, description: '' }))}
                placeholder="Details about the maintenance window..."
                minHeight="120px"
              />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1">Start Time</label>
                <input
                  type="datetime-local"
                  required
                  value={form.startTime}
                  onChange={(e) => setForm((f) => ({ ...f, startTime: e.target.value }))}
                  className="w-full border border-gray-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-slate-100 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1">End Time</label>
                <input
                  type="datetime-local"
                  required
                  value={form.endTime}
                  onChange={(e) => setForm((f) => ({ ...f, endTime: e.target.value }))}
                  className="w-full border border-gray-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-slate-100 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1">Status</label>
                <select
                  value={form.status}
                  onChange={(e) => setForm((f) => ({ ...f, status: e.target.value as MaintenanceStatus }))}
                  className="w-full border border-gray-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-slate-100 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  {STATUSES.map((s) => (
                    <option key={s} value={s}>
                      {STATUS_LABELS[s]}
                    </option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1">Visibility</label>
                <select
                  value={form.visibilityState}
                  onChange={(e) =>
                    setForm((f) => ({ ...f, visibilityState: e.target.value as IncidentVisibilityState }))
                  }
                  className="w-full border border-gray-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-slate-100 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  <option value="draft">{VISIBILITY_LABELS.draft}</option>
                  <option value="published">{VISIBILITY_LABELS.published}</option>
                </select>
              </div>
            </div>
            {(components || []).length > 0 && (
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-2">Affected Components</label>
                <div className="space-y-1">
                  {(components || []).map((c) => (
                    <label key={c.id} className="flex items-center gap-2 text-sm cursor-pointer text-gray-800 dark:text-slate-300">
                      <input
                        type="checkbox"
                        checked={form.components.includes(c.id)}
                        onChange={() => toggleComponent(c.id)}
                        className="rounded"
                      />
                      {c.name}
                    </label>
                  ))}
                </div>
              </div>
            )}
            <div className="rounded-lg border border-slate-100 dark:border-slate-800 bg-slate-50 dark:bg-slate-900 p-3">
              <p className="text-xs font-medium text-slate-500 dark:text-slate-400 mb-2">Preview</p>
              <ContentRenderer text={previewContent.text} json={previewContent.json} />
            </div>
            <div className="flex gap-3 pt-2">
              <button
                type="button"
                onClick={closeModal}
                className="flex-1 border border-gray-300 dark:border-slate-700 text-gray-700 dark:text-slate-300 rounded-lg py-2 text-sm hover:bg-gray-50 dark:hover:bg-slate-800"
              >
                Cancel
              </button>
              <button
                type="submit"
                disabled={saving}
                className="flex-1 bg-blue-600 hover:bg-blue-700 disabled:opacity-60 text-white rounded-lg py-2 text-sm font-medium"
              >
                {saving ? 'Saving...' : editing ? 'Update' : 'Schedule'}
              </button>
            </div>
          </form>
        </Modal>
      )}

      {viewing && (
        <Modal title={viewing.title} onClose={() => setViewing(null)} size="lg">
          <div className="space-y-4">
            <div className="flex flex-wrap items-center gap-2">
              <span
                className={`text-xs px-2 py-0.5 rounded-full font-medium ${STATUS_COLORS[normalizeMaintenanceStatus(viewing.status)]}`}
              >
                {STATUS_LABELS[normalizeMaintenanceStatus(viewing.status)]}
              </span>
              <span
                className={`text-xs px-2 py-0.5 rounded-full font-medium ${VISIBILITY_COLORS[viewing.visibilityState || 'published']}`}
              >
                {VISIBILITY_LABELS[viewing.visibilityState || 'published']}
              </span>
            </div>
            <div className="text-gray-900 dark:text-slate-100">
              <ContentRenderer text={getMaintenanceContent(viewing).text} json={getMaintenanceContent(viewing).json} />
            </div>
            <div className="grid grid-cols-2 gap-3 text-xs text-gray-500 dark:text-slate-400">
              <div>
                <span className="font-medium">Start:</span> {formatDate(viewing.startTime)}
              </div>
              <div>
                <span className="font-medium">End:</span> {formatDate(viewing.endTime)}
              </div>
            </div>
            {viewing.components.length > 0 && (
              <div className="flex flex-wrap gap-2">
                {viewing.components.map((id) => {
                  const comp = (components || []).find((c) => c.id === id)
                  return (
                    <span key={id} className="rounded border border-slate-200 dark:border-slate-800 bg-slate-100 dark:bg-slate-900 px-2 py-1 font-mono text-xs text-slate-600 dark:text-slate-400">
                      {comp?.name || id}
                    </span>
                  )
                })}
              </div>
            )}
          </div>
        </Modal>
      )}

      {deleting && (
        <Modal title="Delete Maintenance" onClose={() => setDeleting(null)}>
          <div className="space-y-4">
            <p className="text-sm text-gray-700 dark:text-slate-300">
              Are you sure you want to delete <strong>{deleting.title}</strong>? This action cannot be undone.
            </p>
            <div className="flex gap-3">
              <button
                type="button"
                onClick={() => setDeleting(null)}
                className="flex-1 border border-gray-300 dark:border-slate-700 text-gray-700 dark:text-slate-300 rounded-lg py-2 text-sm hover:bg-gray-50 dark:hover:bg-slate-800"
              >
                Cancel
              </button>
              <button
                type="button"
                onClick={handleDelete}
                disabled={saving}
                className="flex-1 bg-red-600 hover:bg-red-700 disabled:opacity-60 text-white rounded-lg py-2 text-sm font-medium"
              >
                {saving ? 'Deleting...' : 'Delete'}
              </button>
            </div>
          </div>
        </Modal>
      )}
    </div>
  )
}
