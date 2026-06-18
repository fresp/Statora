import React, { useState } from 'react'
import { Plus, ChevronDown, ChevronUp, Pencil, Trash2, Eye, CheckCircle } from 'lucide-react'
import { useApi } from '../../hooks/useApi'
import { useAdminPagination } from '../../hooks/useAdminPagination'
import api from '../../lib/api'
import { getApiErrorMessage } from '../../lib/apiError'
import type {
  Incident,
  IncidentUpdate,
  Component,
  IncidentStatus,
  IncidentImpact,
  SubComponent,
  RichTextDocument,
} from '../../types'
import { INCIDENT_STATUS_LABELS, INCIDENT_IMPACT_LABELS, formatDate } from '../../lib/utils'
import { getIncidentContent, getIncidentUpdateContent, getPlainTextFromRichText } from '../../lib/contentModel'
import Modal from '../../components/Modal'
import AdminPaginationControls from '../../components/AdminPaginationControls'
import { AdminListCard, AdminTableEmptyRow, textOrEmDash } from '../../components/AdminTableShell'
import RichTextEditor from '../../components/editor/RichTextEditor'
import ContentRenderer from '../../components/content/ContentRenderer'

const STATUSES: IncidentStatus[] = ['investigating', 'identified', 'monitoring', 'resolved']
const IMPACTS: IncidentImpact[] = ['none', 'minor', 'major', 'critical']

interface IncidentForm {
  title: string
  description: string
  descriptionJson?: RichTextDocument
  status: IncidentStatus
  impact: IncidentImpact
  affectedComponentTargets: Array<{
    componentId: string
    subComponentIds: string[]
  }>
}

const DEFAULT_FORM: IncidentForm = {
  title: '',
  description: '',
  status: 'investigating',
  impact: 'minor',
  affectedComponentTargets: [],
}

function IncidentRow({
  incident,
  onRefetch,
  onEdit,
  onView,
  onDelete,
  onResolve,
}: {
  incident: Incident
  onRefetch: () => void
  onEdit: (incident: Incident) => void
  onView: (incident: Incident) => void
  onDelete: (incident: Incident) => void
  onResolve: (incident: Incident) => void
}) {
  const [expanded, setExpanded] = useState(false)
  const [showUpdateModal, setShowUpdateModal] = useState(false)
  const [updateMsg, setUpdateMsg] = useState('')
  const [updateMsgJson, setUpdateMsgJson] = useState<RichTextDocument | undefined>()
  const [updateStatus, setUpdateStatus] = useState<IncidentStatus>('investigating')
  const [updates, setUpdates] = useState<IncidentUpdate[] | null>(null)
  const [saving, setSaving] = useState(false)

  async function loadUpdates() {
    if (!expanded) {
      const res = await api.get<IncidentUpdate[]>(`/incidents/${incident.id}/updates`)
      setUpdates(res.data)
    }
    setExpanded((e) => !e)
  }

  async function submitUpdate(e: React.FormEvent) {
    e.preventDefault()
    setSaving(true)
    try {
      const plainMessage = getPlainTextFromRichText(updateMsgJson) || updateMsg
      await api.post(`/incidents/${incident.id}/update`, {
        message: plainMessage,
        messageJson: updateMsgJson,
        status: updateStatus,
      })
      setShowUpdateModal(false)
      setUpdateMsg('')
      setUpdateMsgJson(undefined)
      onRefetch()
      const res = await api.get<IncidentUpdate[]>(`/incidents/${incident.id}/updates`)
      setUpdates(res.data)
    } catch {
      alert('Failed to add update')
    } finally {
      setSaving(false)
    }
  }

  const statusColor =
    incident.status === 'resolved'
      ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
      : incident.status === 'monitoring'
        ? 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400'
        : 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'

  const impactColor: Record<IncidentImpact, string> = {
    none: 'bg-gray-100 text-gray-600 dark:bg-slate-800 dark:text-slate-400',
    minor: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400',
    major: 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400',
    critical: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400',
  }

  const descriptionContent = getIncidentContent(incident)

  return (
    <>
      <tr className="hover:bg-gray-50 dark:hover:bg-slate-800/50">
        <td className="px-6 py-4 font-medium text-gray-900 dark:text-slate-100 max-w-xs truncate">{incident.title}</td>
        <td className="px-6 py-4">
          <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${statusColor}`}>
            {INCIDENT_STATUS_LABELS[incident.status]}
          </span>
        </td>
        <td className="px-6 py-4">
          <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${impactColor[incident.impact]}`}>
            {INCIDENT_IMPACT_LABELS[incident.impact]}
          </span>
        </td>
        <td className="px-6 py-4 text-sm text-gray-500 dark:text-slate-400">{formatDate(incident.createdAt)}</td>
        <td className="px-6 py-4 text-sm text-gray-500 dark:text-slate-400">{textOrEmDash(incident.creatorUsername)}</td>
        <td className="px-6 py-4">
          <div className="flex items-center justify-end gap-2">
            <button
              onClick={() => onView(incident)}
              className="text-gray-400 dark:text-slate-500 hover:text-emerald-600 dark:hover:text-emerald-400 transition-colors"
              title="View"
            >
              <Eye className="w-4 h-4" />
            </button>
            <button
              onClick={() => onEdit(incident)}
              className="text-gray-400 dark:text-slate-500 hover:text-blue-600 dark:hover:text-blue-400 transition-colors"
              title="Edit"
            >
              <Pencil className="w-4 h-4" />
            </button>
            {incident.status !== 'resolved' && (
              <button
                onClick={() => onResolve(incident)}
                className="text-gray-400 dark:text-slate-500 hover:text-green-600 dark:hover:text-green-400 transition-colors"
                title="Resolve"
              >
                <CheckCircle className="w-4 h-4" />
              </button>
            )}
            <button
              onClick={() => onDelete(incident)}
              className="text-gray-400 dark:text-slate-500 hover:text-red-600 dark:hover:text-red-400 transition-colors"
              title="Delete"
            >
              <Trash2 className="w-4 h-4" />
            </button>
            <button
              onClick={() => setShowUpdateModal(true)}
              className="text-xs text-blue-600 dark:text-blue-400 hover:underline"
            >
              Add Update
            </button>
            <button onClick={loadUpdates} className="text-gray-400 dark:text-slate-500 hover:text-gray-600 dark:hover:text-slate-300">
              {expanded ? <ChevronUp className="w-4 h-4" /> : <ChevronDown className="w-4 h-4" />}
            </button>
          </div>
        </td>
      </tr>
      {expanded && (
        <tr>
          <td colSpan={6} className="px-6 pb-4 bg-gray-50 dark:bg-slate-800/40">
            <div className="pl-4 border-l-2 border-gray-200 dark:border-slate-700 space-y-3 mt-1">
              <div className="text-sm text-gray-700 dark:text-slate-300">
                <ContentRenderer text={descriptionContent.text} json={descriptionContent.json} />
              </div>
              {(updates || []).length === 0 ? (
                <p className="text-sm text-gray-400 dark:text-slate-500">No updates yet.</p>
              ) : (
                <div className="space-y-2">
                  {(updates || []).map((u) => {
                    const uc = getIncidentUpdateContent(u)
                    return (
                      <div key={u.id} className="text-sm">
                        <span className="font-medium text-gray-700 dark:text-slate-300">{INCIDENT_STATUS_LABELS[u.status]}</span>
                        <span className="text-gray-400 dark:text-slate-500 ml-2 text-xs">{formatDate(u.createdAt)}</span>
                        <div className="mt-1 text-gray-600 dark:text-slate-300">
                          <ContentRenderer text={uc.text} json={uc.json} />
                        </div>
                      </div>
                    )
                  })}
                </div>
              )}
            </div>
          </td>
        </tr>
      )}

      {showUpdateModal && (
        <Modal title="Add Incident Update" onClose={() => setShowUpdateModal(false)} size="lg">
          <form onSubmit={submitUpdate} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1">Status</label>
              <select
                value={updateStatus}
                onChange={(e) => setUpdateStatus(e.target.value as IncidentStatus)}
                className="w-full border border-gray-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-slate-100 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                {STATUSES.map((s) => (
                  <option key={s} value={s}>
                    {INCIDENT_STATUS_LABELS[s]}
                  </option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1">Message</label>
              <RichTextEditor
                value={updateMsgJson}
                onChange={(json) => {
                  setUpdateMsgJson(json)
                  setUpdateMsg('')
                }}
                placeholder="Describe what's happening..."
                minHeight="120px"
              />
            </div>
            <div className="flex gap-3">
              <button
                type="button"
                onClick={() => setShowUpdateModal(false)}
                className="flex-1 border border-gray-300 dark:border-slate-700 text-gray-700 dark:text-slate-300 rounded-lg py-2 text-sm hover:bg-gray-50 dark:hover:bg-slate-800"
              >
                Cancel
              </button>
              <button
                type="submit"
                disabled={saving}
                className="flex-1 bg-blue-600 hover:bg-blue-700 disabled:opacity-60 text-white rounded-lg py-2 text-sm font-medium"
              >
                {saving ? 'Posting...' : 'Post Update'}
              </button>
            </div>
          </form>
        </Modal>
      )}
    </>
  )
}

export default function AdminIncidents() {
  const { page, limit, apiParams, setPage, setLimit } = useAdminPagination()
  const { data: incidents, total: totalIncidents, totalPages, loading, refetch } = useApi<Incident[]>('/incidents', [], apiParams)
  const { data: components } = useApi<Component[]>('/components', [], { page: 1, limit: 100 })
  const { data: subComponents } = useApi<SubComponent[]>('/subcomponents', [], { page: 1, limit: 100 })
  const [showModal, setShowModal] = useState(false)
  const [editingId, setEditingId] = useState<string | null>(null)
  const [form, setForm] = useState<IncidentForm>(DEFAULT_FORM)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  const [filter, setFilter] = useState<'all' | 'active' | 'resolved'>('all')
  const [viewing, setViewing] = useState<Incident | null>(null)
  const [deleting, setDeleting] = useState<Incident | null>(null)
  const [resolving, setResolving] = useState<Incident | null>(null)

  function openCreate() {
    setEditingId(null)
    setForm(DEFAULT_FORM)
    setError('')
    setShowModal(true)
  }

  function openEdit(incident: Incident) {
    setEditingId(incident.id)
    setForm({
      title: incident.title,
      description: incident.description,
      descriptionJson: incident.descriptionJson,
      status: incident.status,
      impact: incident.impact,
      affectedComponentTargets:
        incident.affectedComponentTargets?.map((t) => ({
          componentId: t.component.id,
          subComponentIds: t.subComponents?.map((s) => s.id) ?? [],
        })) ??
        incident.affectedComponents.map((c) => ({
          componentId: c.id,
          subComponentIds: [],
        })),
    })
    setError('')
    setShowModal(true)
  }

  function closeModal() {
    setShowModal(false)
    setEditingId(null)
  }

  async function handleSave(e: React.FormEvent) {
    e.preventDefault()
    setSaving(true)
    setError('')
    try {
      const payload = {
        ...form,
        description: getPlainTextFromRichText(form.descriptionJson) || form.description,
        descriptionJson: form.descriptionJson,
      }
      if (editingId) {
        await api.patch(`/incidents/${editingId}`, payload)
      } else {
        await api.post('/incidents', payload)
      }
      await refetch()
      closeModal()
    } catch (err: unknown) {
      setError(getApiErrorMessage(err, 'Failed to save incident'))
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete() {
    if (!deleting) return
    setSaving(true)
    try {
      await api.delete(`/incidents/${deleting.id}`)
      setDeleting(null)
      await refetch()
    } catch (err: unknown) {
      setError(getApiErrorMessage(err, 'Failed to delete incident'))
      alert(error || 'Failed to delete incident')
    } finally {
      setSaving(false)
    }
  }

  async function handleResolve() {
    if (!resolving) return
    setSaving(true)
    try {
      await api.post(`/incidents/${resolving.id}/resolve`)
      setResolving(null)
      await refetch()
    } catch (err: unknown) {
      alert(getApiErrorMessage(err, 'Failed to resolve incident'))
    } finally {
      setSaving(false)
    }
  }

  function toggleComponent(componentId: string) {
    setForm((f) => ({
      ...f,
      affectedComponentTargets: f.affectedComponentTargets.some((t) => t.componentId === componentId)
        ? f.affectedComponentTargets.filter((t) => t.componentId !== componentId)
        : [...f.affectedComponentTargets, { componentId, subComponentIds: [] }],
    }))
  }

  function toggleSubComponent(componentId: string, subComponentId: string) {
    setForm((current) => {
      const index = current.affectedComponentTargets.findIndex((target) => target.componentId === componentId)
      if (index === -1) {
        return {
          ...current,
          affectedComponentTargets: [
            ...current.affectedComponentTargets,
            { componentId, subComponentIds: [subComponentId] },
          ],
        }
      }

      const target = current.affectedComponentTargets[index]
      const hasSub = target.subComponentIds.includes(subComponentId)
      const nextSubComponentIds = hasSub
        ? target.subComponentIds.filter((id) => id !== subComponentId)
        : [...target.subComponentIds, subComponentId]

      const nextTargets = [...current.affectedComponentTargets]
      nextTargets[index] = {
        ...target,
        subComponentIds: nextSubComponentIds,
      }

      return {
        ...current,
        affectedComponentTargets: nextTargets,
      }
    })
  }

  function getTarget(componentId: string) {
    return form.affectedComponentTargets.find((t) => t.componentId === componentId)
  }

  const filtered = (incidents || []).filter((i) => {
    if (filter === 'active') return i.status !== 'resolved'
    if (filter === 'resolved') return i.status === 'resolved'
    return true
  })

  const previewContent = getIncidentContent(form)

  return (
    <div className="p-8">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-slate-100">Incidents</h1>
          <p className="text-sm text-gray-500 dark:text-slate-400 mt-1">{totalIncidents} total</p>
        </div>
        <button
          onClick={openCreate}
          className="flex items-center gap-2 bg-red-600 hover:bg-red-700 text-white text-sm font-medium px-4 py-2 rounded-lg transition-colors"
        >
          <Plus className="w-4 h-4" /> Create Incident
        </button>
      </div>

      <div className="flex gap-2 mb-4">
        {(['all', 'active', 'resolved'] as const).map((f) => (
          <button
            key={f}
            onClick={() => setFilter(f)}
            className={`px-3 py-1.5 rounded-lg text-sm font-medium capitalize transition-colors ${
              filter === f
                ? 'bg-gray-900 text-white dark:bg-slate-100 dark:text-slate-900'
                : 'bg-white border border-gray-200 text-gray-600 hover:bg-gray-50 dark:bg-slate-800 dark:border-slate-700 dark:text-slate-300 dark:hover:bg-slate-700'
            }`}
          >
            {f}
          </button>
        ))}
      </div>

      <AdminListCard>
        <table className="w-full text-sm">
          <thead className="bg-gray-50 border-b border-gray-100 dark:bg-slate-800/50 dark:border-slate-700">
            <tr>
              <th className="text-left px-6 py-3 font-medium text-gray-600 dark:text-slate-300">Title</th>
              <th className="text-left px-6 py-3 font-medium text-gray-600 dark:text-slate-300">Status</th>
              <th className="text-left px-6 py-3 font-medium text-gray-600 dark:text-slate-300">Impact</th>
              <th className="text-left px-6 py-3 font-medium text-gray-600 dark:text-slate-300">Created</th>
              <th className="text-left px-6 py-3 font-medium text-gray-600 dark:text-slate-300">Creator</th>
              <th className="px-6 py-3" />
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-50 dark:divide-slate-800/50">
            {filtered.map((inc) => (
              <IncidentRow
                key={inc.id}
                incident={inc}
                onRefetch={refetch}
                onEdit={openEdit}
                onView={setViewing}
                onDelete={setDeleting}
                onResolve={setResolving}
              />
            ))}
            {filtered.length === 0 && (
              <AdminTableEmptyRow colSpan={6}>No incidents found.</AdminTableEmptyRow>
            )}
          </tbody>
        </table>

        <AdminPaginationControls
          page={page}
          totalPages={totalPages}
          total={totalIncidents}
          limit={limit}
          loading={loading}
          onPageChange={setPage}
          onLimitChange={setLimit}
        />
      </AdminListCard>

      {showModal && (
        <Modal title={editingId ? 'Edit Incident' : 'Create Incident'} onClose={closeModal} size="lg">
          {error && <p className="mb-4 text-sm text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-900/20 rounded-lg px-3 py-2">{error}</p>}
          <form onSubmit={handleSave} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1">Title</label>
              <input
                required
                value={form.title}
                onChange={(e) => setForm((f) => ({ ...f, title: e.target.value }))}
                className="w-full border border-gray-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-slate-100 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="Brief incident title"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1">Description</label>
              <RichTextEditor
                value={form.descriptionJson || form.description}
                onChange={(json) => setForm((f) => ({ ...f, descriptionJson: json, description: '' }))}
                placeholder="What is happening?"
                minHeight="140px"
              />
            </div>
            <div className="grid grid-cols-2 gap-3">
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1">Status</label>
                <select
                  value={form.status}
                  onChange={(e) => setForm((f) => ({ ...f, status: e.target.value as IncidentStatus }))}
                  className="w-full border border-gray-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-slate-100 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  {STATUSES.map((s) => (
                    <option key={s} value={s}>
                      {INCIDENT_STATUS_LABELS[s]}
                    </option>
                  ))}
                </select>
              </div>
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1">Impact</label>
                <select
                  value={form.impact}
                  onChange={(e) => setForm((f) => ({ ...f, impact: e.target.value as IncidentImpact }))}
                  className="w-full border border-gray-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-slate-100 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  {IMPACTS.map((i) => (
                    <option key={i} value={i}>
                      {INCIDENT_IMPACT_LABELS[i]}
                    </option>
                  ))}
                </select>
              </div>
            </div>
            {(components || []).length > 0 && (
              <div>
                <label className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-2">Affected Components</label>
                <div className="space-y-1">
                  {(components || []).map((component) => {
                    const target = getTarget(component.id)
                    const checked = Boolean(target)
                    const relatedSubComponents = (subComponents || []).filter(
                      (subComponent) => subComponent.componentId === component.id,
                    )

                    return (
                      <div key={component.id} className="rounded-lg border border-gray-200 dark:border-slate-700 px-3 py-2">
                        <label className="flex items-center gap-2 text-sm cursor-pointer font-medium text-gray-800 dark:text-slate-200">
                          <input
                            type="checkbox"
                            checked={checked}
                            onChange={() => toggleComponent(component.id)}
                            className="rounded"
                          />
                          {component.name}
                        </label>

                        {checked && relatedSubComponents.length > 0 && (
                          <div className="mt-2 pl-6 space-y-1">
                            {relatedSubComponents.map((subComponent) => {
                              const isSubChecked = target?.subComponentIds.includes(subComponent.id) || false
                              return (
                                <label
                                  key={subComponent.id}
                                  className="flex items-center gap-2 text-xs text-gray-600 dark:text-slate-400 cursor-pointer"
                                >
                                  <input
                                    type="checkbox"
                                    checked={isSubChecked}
                                    onChange={() => toggleSubComponent(component.id, subComponent.id)}
                                    className="rounded"
                                  />
                                  {subComponent.name}
                                </label>
                              )
                            })}
                            <p className="text-[11px] text-gray-500 dark:text-slate-500 pt-1">
                              Leave sub-components unchecked to affect the whole component.
                            </p>
                          </div>
                        )}
                      </div>
                    )
                  })}
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
                className="flex-1 bg-red-600 hover:bg-red-700 disabled:opacity-60 text-white rounded-lg py-2 text-sm font-medium"
              >
                {saving ? 'Saving...' : editingId ? 'Update' : 'Create Incident'}
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
                className={`text-xs px-2 py-0.5 rounded-full font-medium ${
                  viewing.status === 'resolved'
                    ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
                    : viewing.status === 'monitoring'
                      ? 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400'
                      : 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
                }`}
              >
                {INCIDENT_STATUS_LABELS[viewing.status]}
              </span>
              <span
                className={`text-xs px-2 py-0.5 rounded-full font-medium ${
                  {
                    none: 'bg-gray-100 text-gray-600 dark:bg-slate-800 dark:text-slate-400',
                    minor: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400',
                    major: 'bg-orange-100 text-orange-700 dark:bg-orange-900/30 dark:text-orange-400',
                    critical: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400',
                  }[viewing.impact]
                }`}
              >
                {INCIDENT_IMPACT_LABELS[viewing.impact]}
              </span>
            </div>
            <div className="text-gray-900 dark:text-slate-100">
              <ContentRenderer
                text={getIncidentContent(viewing).text}
                json={getIncidentContent(viewing).json}
              />
            </div>
            <div className="text-xs text-gray-400 dark:text-slate-500">
              Created {formatDate(viewing.createdAt)}
              {viewing.resolvedAt ? ` · Resolved ${formatDate(viewing.resolvedAt)}` : ''}
            </div>
            {(viewing.updates || []).length > 0 && (
              <div className="border-t border-slate-100 dark:border-slate-800 pt-3 space-y-3">
                <p className="text-xs font-semibold uppercase tracking-wider text-slate-500 dark:text-slate-400">Updates</p>
                {viewing.updates?.map((u) => {
                  const uc = getIncidentUpdateContent(u)
                  return (
                    <div key={u.id} className="text-sm">
                      <div className="flex items-center gap-2 mb-1">
                        <span className="font-medium text-gray-700 dark:text-slate-300">{INCIDENT_STATUS_LABELS[u.status]}</span>
                        <span className="text-gray-400 dark:text-slate-500 text-xs">{formatDate(u.createdAt)}</span>
                      </div>
                      <div className="text-gray-600 dark:text-slate-300">
                        <ContentRenderer text={uc.text} json={uc.json} />
                      </div>
                    </div>
                  )
                })}
              </div>
            )}
          </div>
        </Modal>
      )}

      {deleting && (
        <Modal title="Delete Incident" onClose={() => setDeleting(null)}>
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

      {resolving && (
        <Modal title="Resolve Incident" onClose={() => setResolving(null)}>
          <div className="space-y-4">
            <p className="text-sm text-gray-700 dark:text-slate-300">
              Resolve <strong>{resolving.title}</strong>? This will mark the incident as resolved.
            </p>
            <div className="flex gap-3">
              <button
                type="button"
                onClick={() => setResolving(null)}
                className="flex-1 border border-gray-300 dark:border-slate-700 text-gray-700 dark:text-slate-300 rounded-lg py-2 text-sm hover:bg-gray-50 dark:hover:bg-slate-800"
              >
                Cancel
              </button>
              <button
                type="button"
                onClick={handleResolve}
                disabled={saving}
                className="flex-1 bg-green-600 hover:bg-green-700 disabled:opacity-60 text-white rounded-lg py-2 text-sm font-medium"
              >
                {saving ? 'Resolving...' : 'Resolve'}
              </button>
            </div>
          </div>
        </Modal>
      )}
    </div>
  )
}
