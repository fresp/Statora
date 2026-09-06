import React, { useState } from 'react'
import { Trash2, Mail, RefreshCw } from 'lucide-react'
import { useApi } from '../../hooks/useApi'
import { useAdminPagination } from '../../hooks/useAdminPagination'
import api from '../../lib/api'
import { getApiErrorMessage } from '../../lib/apiError'
import type { Subscriber } from '../../types'
import { formatDate } from '../../lib/utils'
import AdminPaginationControls from '../../components/AdminPaginationControls'
import { AdminListCard, AdminTableEmptyRow } from '../../components/AdminTableShell'

export default function AdminSubscribers() {
  const { page, limit, apiParams, setPage, setLimit } = useAdminPagination()
  const { data: subscribers, total: totalSubscribers, totalPages, loading, refetch } = useApi<Subscriber[]>('/subscribers', [], apiParams)
  const [deleting, setDeleting] = useState<string | null>(null)
  const [resending, setResending] = useState<string | null>(null)

  async function handleDelete(s: Subscriber) {
    if (!confirm(`Remove subscriber ${s.email}?`)) return
    setDeleting(s.id)
    try {
      await api.delete(`/subscribers/${s.id}`)
      await refetch()
    } catch (err: unknown) {
      alert(getApiErrorMessage(err, 'Failed to delete'))
    } finally {
      setDeleting(null)
    }
  }

  const verified = subscribers?.filter(s => s.verified && !s.unsubscribed) || []
  const unverified = subscribers?.filter(s => !s.verified && !s.unsubscribed) || []
  const unsubscribed = subscribers?.filter(s => s.unsubscribed) || []

  async function handleResend(s: Subscriber) {
    setResending(s.id)
    try {
      await api.post(`/subscribers/${s.id}/resend-verification`)
      await refetch()
    } catch (err: unknown) {
      alert(getApiErrorMessage(err, 'Failed to resend verification email'))
    } finally {
      setResending(null)
    }
  }

  return (
    <div className="p-8">
      <div className="flex items-center justify-between mb-6">
        <div>
          <h1 className="text-2xl font-bold text-gray-900 dark:text-slate-100">Subscribers</h1>
            <p className="text-sm text-gray-500 dark:text-slate-400 mt-1">
            {totalSubscribers} total · {verified.length} verified · {unverified.length} pending · {unsubscribed.length} unsubscribed
            </p>
        </div>
      </div>

      <AdminListCard>
        <table className="w-full text-sm">
          <thead className="bg-gray-50 border-b border-gray-100 dark:bg-slate-800/50 dark:border-slate-700">
            <tr>
              <th className="text-left px-6 py-3 font-medium text-gray-600 dark:text-slate-300">Email</th>
              <th className="text-left px-6 py-3 font-medium text-gray-600 dark:text-slate-300">Status</th>
              <th className="text-left px-6 py-3 font-medium text-gray-600 dark:text-slate-300">Subscribed</th>
              <th className="px-6 py-3" />
            </tr>
          </thead>
          <tbody className="divide-y divide-gray-50 dark:divide-slate-800/50">
            {(subscribers || []).map(s => (
              <tr key={s.id} className="hover:bg-gray-50 dark:hover:bg-slate-800/50">
                <td className="px-6 py-4">
                  <div className="flex items-center gap-2">
                    <Mail className="w-4 h-4 text-gray-400 dark:text-slate-500" />
                    <span className="text-gray-900 dark:text-slate-100">{s.email}</span>
                  </div>
                </td>
                <td className="px-6 py-4">
                  <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${
                    s.unsubscribed
                      ? 'bg-gray-100 text-gray-600 dark:bg-slate-800 dark:text-slate-400'
                      : s.verified
                        ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
                        : 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400'
                  }`}>
                    {s.unsubscribed ? 'Unsubscribed' : s.verified ? 'Verified' : 'Pending'}
                  </span>
                </td>
                <td className="px-6 py-4 text-gray-500 dark:text-slate-400">{formatDate(s.createdAt)}</td>
                <td className="px-6 py-4">
                  <div className="flex items-center justify-end gap-1">
                    {!s.verified && !s.unsubscribed && (
                      <button
                        onClick={() => handleResend(s)}
                        disabled={resending === s.id}
                        title="Resend verification email"
                        className="text-gray-400 dark:text-slate-500 hover:text-blue-600 dark:hover:text-blue-400 transition-colors disabled:opacity-40"
                      >
                        <RefreshCw className={`w-4 h-4 ${resending === s.id ? 'animate-spin' : ''}`} />
                      </button>
                    )}
                    <button
                      onClick={() => handleDelete(s)}
                      disabled={deleting === s.id}
                      className="text-gray-400 dark:text-slate-500 hover:text-red-600 dark:hover:text-red-400 transition-colors disabled:opacity-40"
                    >
                      <Trash2 className="w-4 h-4" />
                    </button>
                  </div>
                </td>
              </tr>
            ))}
             {(subscribers || []).length === 0 && (
                <AdminTableEmptyRow colSpan={4}>
                  No subscribers yet. Subscribers sign up from the public status page.
                </AdminTableEmptyRow>
              )}
          </tbody>
        </table>

        <AdminPaginationControls
          page={page}
          totalPages={totalPages}
          total={totalSubscribers}
          limit={limit}
          loading={loading}
          onPageChange={setPage}
          onLimitChange={setLimit}
        />
      </AdminListCard>
    </div>
  )
}
