import React, { useState, useEffect } from 'react'
import { Plus, Pencil } from 'lucide-react'
import { useApi } from '../../hooks/useApi'
import api from '../../lib/api'
import type { SubComponent, Component, ComponentStatus } from '../../types'
import { STATUS_LABELS, STATUS_COLORS } from '../../lib/utils'
import Modal from '../../components/Modal'

const STATUSES: ComponentStatus[] = ['operational', 'degraded_performance', 'partial_outage', 'major_outage', 'maintenance']

interface FormState {
  componentId: string
  name: string
  description: string
  status: ComponentStatus
}

const DEFAULT_FORM: FormState = { componentId: '', name: '', description: '', status: 'operational' }

export default function AdminSubComponents() {
  // Fetch components but not subcomponents initially
  const { data: components, refetch: refetchComponents } = useApi<Component[]>('/components')
  const [subComponents, setSubComponents] = useState<SubComponent[] | undefined>()
  const [loadingAll, setLoadingAll] = useState(false)
  const [showModal, setShowModal] = useState(false)
  const [editing, setEditing] = useState<SubComponent | null>(null)
  


  const [form, setForm] = useState<FormState>(DEFAULT_FORM)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState('')
  
  // Fetch all subcomponents by fetching each component's subcomponents
  const fetchAllSubcomponents = async () => {
    if (!components || components.length === 0) {
      setSubComponents([])
      return
    }
    
    setLoadingAll(true)
    try {
      // Fetch subcomponents for each component
      const allSubComponents = await Promise.all(
        components.map(async (component) => {
          const response = await api.get<SubComponent[]>(`/components/${component.id}/subcomponents`)
          return response.data;
        })
      )
      
      // Flatten and set the combined results
      const flattened = allSubComponents.flat()
      setSubComponents(flattened) 
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to fetch subcomponents')
      setSubComponents(undefined)
    } finally {
      setLoadingAll(false)
    }
  }
  
  // Refetch both components and subcomponents
  const refetch = async () => {
    await refetchComponents()
    fetchAllSubcomponents() // Fetch subcomponents after components update
  }
  // Fetch all subcomponents when components are loaded
  useEffect(() => {
    if (components) {
      fetchAllSubcomponents();
    }
  }, [components]);
  

  function openCreate() {
    setEditing(null)
    setForm({ ...DEFAULT_FORM, componentId: components?.[0]?.id || '' })
    setError('')
    setShowModal(true)
  }

  function openEdit(s: SubComponent) {
    setEditing(s)
    setForm({ componentId: s.componentId, name: s.name, description: s.description, status: s.status })
    setError('')
    setShowModal(true)
  }

  function closeModal() {
    setShowModal(false)
    setEditing(null)
  }

  async function handleSave(e: React.FormEvent) {
    e.preventDefault()
    setSaving(true)
    setError('')
    try {
      if (editing) {
        await api.patch(`/subcomponents/${editing.id}`, form)
      } else {
        await api.post('/subcomponents', form)
      }
      await refetch()
      closeModal()
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to save')
    } finally {
      setSaving(false)
    }
  }

  function getComponentName(id: string) {
    return components?.find(c => c.id === id)?.name || id
  }

  return (
    <div className="p-8">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900">Sub-Components</h1>
          <p className="text-sm text-gray-500 mt-1">{subComponents?.length ?? 0} total</p>
        </div>
        <button
          onClick={openCreate}
          disabled={!components?.length}
          className="flex items-center gap-2 bg-blue-600 hover:bg-blue-700 disabled:opacity-50 text-white text-sm font-medium px-4 py-2 rounded-lg transition-colors"
        >
          <Plus className="w-4 h-4" /> Add Sub-Component
        </button>
      </div>

      {!components?.length && (
        <div className="bg-yellow-50 border border-yellow-200 rounded-lg px-4 py-3 text-sm text-yellow-700 mb-4">
          Create at least one component before adding sub-components.
        </div>
      )}

      <div className="bg-white rounded-xl border border-gray-200 overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-gray-50 border-b border-gray-100">
            <tr>
              <th className="text-left px-6 py-3 font-medium text-gray-600">Name</th>
              <th className="text-left px-6 py-3 font-medium text-gray-600">Parent Component</th>
              <th className="text-left px-6 py-3 font-medium text-gray-600">Status</th>
              <th className="px-6 py-3" />
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-50">
            {(subComponents || []).map(s => (
              <tr key={s.id} className="hover:bg-gray-50">
                <td className="px-6 py-4 font-medium text-gray-900">{s.name}</td>
                <td className="px-6 py-4 text-gray-500">{getComponentName(s.componentId)}</td>
                <td className="px-6 py-4">
                  <span className="flex items-center gap-1.5">
                    <span className={`w-2 h-2 rounded-full ${STATUS_COLORS[s.status]}`} />
                    {STATUS_LABELS[s.status]}
                  </span>
                </td>
                <td className="px-6 py-4">
                  <div className="flex items-center justify-end">
                    <button onClick={() => openEdit(s)} className="text-gray-400 hover:text-blue-600 transition-colors">
                      <Pencil className="w-4 h-4" />
                    </button>
                  </div>
                </td>
              </tr>
            ))}
            {loadingAll && (
              <tr>
                <td colSpan={4} className="px-6 py-12 text-center text-gray-500">Loading subcomponents...</td>
              </tr>
            )}
            {!loadingAll && subComponents !== undefined && (subComponents || []).length === 0 && (
              <tr>
                <td colSpan={4} className="px-6 py-12 text-center text-gray-400">No sub-components yet.</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      {showModal && (
        <Modal title={editing ? 'Edit Sub-Component' : 'New Sub-Component'} onClose={closeModal}>
          {error && <p className="mb-4 text-sm text-red-600 bg-red-50 rounded-lg px-3 py-2">{error}</p>}
          <form onSubmit={handleSave} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Parent Component</label>
              <select
                required
                value={form.componentId}
                onChange={e => setForm(f => ({ ...f, componentId: e.target.value }))}
                className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="">Select component...</option>
                {(components || []).map(c => (
                  <option key={c.id} value={c.id}>{c.name}</option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Name</label>
              <input
                required
                value={form.name}
                onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
                className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
              <input
                value={form.description}
                onChange={e => setForm(f => ({ ...f, description: e.target.value }))}
                className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Status</label>
              <select
                value={form.status}
                onChange={e => setForm(f => ({ ...f, status: e.target.value as ComponentStatus }))}
                className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                {STATUSES.map(s => (
                  <option key={s} value={s}>{STATUS_LABELS[s]}</option>
                ))}
              </select>
            </div>
            <div className="flex gap-3 pt-2">
              <button type="button" onClick={closeModal} className="flex-1 border border-gray-300 text-gray-700 rounded-lg py-2 text-sm hover:bg-gray-50">
                Cancel
              </button>
              <button type="submit" disabled={saving} className="flex-1 bg-blue-600 hover:bg-blue-700 disabled:opacity-60 text-white rounded-lg py-2 text-sm font-medium">
                {saving ? 'Saving...' : editing ? 'Update' : 'Create'}
              </button>
            </div>
          </form>
        </Modal>
      )}
    </div>
  )
}
