import React, { useState } from 'react'
import { Search, Loader2 } from 'lucide-react'
import api from '../lib/api'
import { getApiErrorMessage, getApiErrorStatus } from '../lib/apiError'
import type { UserMember } from '../types'
import { AdminListCard } from './AdminTableShell'

interface UserSearchProps {
  onSearchResult?: (user: UserMember | null) => void
}

export default function UserSearch({ onSearchResult }: UserSearchProps) {
  const [email, setEmail] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [user, setUser] = useState<UserMember | null>(null)
  const [searched, setSearched] = useState(false)

  const handleSearch = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!email.trim() || !email.includes('@')) {
      setError('Please enter a valid full email address')
      return
    }

    setLoading(true)
    setError('')
    setSearched(false)
    setUser(null)

    try {
      const res = await api.get<UserMember>(`/v1/users/search?email=${encodeURIComponent(email)}`)
      setUser(res.data)
      onSearchResult?.(res.data)
    } catch (err: unknown) {
      if (getApiErrorStatus(err) === 404) {
        // Not found is an expected state, not an error banner state
        setUser(null)
      } else {
        setError(getApiErrorMessage(err, 'Failed to search user'))
      }
      onSearchResult?.(null)
    } finally {
      setLoading(false)
      setSearched(true)
    }
  }

  const handleClear = () => {
    setEmail('')
    setUser(null)
    setSearched(false)
    setError('')
    onSearchResult?.(null)
  }

  return (
    <div className="mb-8">
      <h2 className="mb-2 text-lg font-semibold tracking-tight text-slate-950">Search Users</h2>
      
      <form onSubmit={handleSearch} className="flex gap-2 max-w-md mb-2">
        <div className="relative flex-1">
          <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
            <Search className="h-4 w-4 text-slate-400" />
          </div>
          <input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="Search by full email address..."
            className="statusforge-input-icon"
            required
          />
        </div>
        <button
          type="submit"
          disabled={loading || !email.trim()}
          className="statusforge-btn-primary"
        >
          {loading ? <Loader2 className="h-4 w-4 animate-spin" /> : 'Search'}
        </button>
        {(searched || email) && (
          <button
            type="button"
            onClick={handleClear}
            className="statusforge-btn-secondary px-3"
          >
            Clear
          </button>
        )}
      </form>
      
      <p className="mb-4 text-xs text-slate-500">Search requires full email</p>

      {error && (
        <div className="statusforge-alert-error mb-4">
          {error}
        </div>
      )}

      {searched && !loading && !user && !error && (
        <div className="statusforge-empty">
          No user found with email "{email}"
        </div>
      )}

      {user && (
        <AdminListCard>
          <table className="w-full text-sm">
            <thead className="statusforge-table-head">
              <tr>
                <th className="px-6 py-3 text-left font-semibold">User</th>
                <th className="px-6 py-3 text-left font-semibold">Role</th>
                <th className="px-6 py-3 text-left font-semibold">Status</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              <tr className="transition-colors hover:bg-slate-50">
                <td className="px-6 py-4">
                  <div>
                    <p className="font-medium text-slate-950">{user.username}</p>
                    <p className="text-xs text-slate-500">{user.email}</p>
                  </div>
                </td>
                <td className="px-6 py-4">
                  <span className="badge badge-neutral capitalize">
                    {user.role}
                  </span>
                </td>
                <td className="px-6 py-4">
                  <span className={`badge capitalize ${
                    user.status === 'active' ? 'badge-success' :
                    user.status === 'invited' ? 'badge-warning' :
                    'badge-error'
                  }`}>
                    {user.status}
                  </span>
                </td>
              </tr>
            </tbody>
          </table>
        </AdminListCard>
      )}
    </div>
  )
}
