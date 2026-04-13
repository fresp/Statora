import React, { useState } from 'react'
import { Search, Loader2 } from 'lucide-react'
import api from '../lib/api'
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
    } catch (err: any) {
      if (err.response?.status === 404) {
        // Not found is an expected state, not an error banner state
        setUser(null)
      } else {
        setError(err.response?.data?.error || 'Failed to search user')
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
      <h2 className="text-lg font-semibold text-gray-900 mb-2">Search Users</h2>
      
      <form onSubmit={handleSearch} className="flex gap-2 max-w-md mb-2">
        <div className="relative flex-1">
          <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
            <Search className="h-4 w-4 text-gray-400" />
          </div>
          <input
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="Search by full email address..."
            className="block w-full pl-10 pr-3 py-2 border border-gray-300 rounded-lg leading-5 bg-white placeholder-gray-500 focus:outline-none focus:placeholder-gray-400 focus:ring-1 focus:ring-blue-500 focus:border-blue-500 sm:text-sm transition duration-150 ease-in-out"
            required
          />
        </div>
        <button
          type="submit"
          disabled={loading || !email.trim()}
          className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-lg shadow-sm text-white bg-gray-900 hover:bg-gray-800 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-gray-900 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          {loading ? <Loader2 className="h-4 w-4 animate-spin" /> : 'Search'}
        </button>
        {(searched || email) && (
          <button
            type="button"
            onClick={handleClear}
            className="inline-flex items-center px-3 py-2 border border-gray-300 shadow-sm text-sm font-medium rounded-lg text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
          >
            Clear
          </button>
        )}
      </form>
      
      <p className="text-xs text-gray-500 mb-4">Search requires full email</p>

      {error && (
        <div className="mb-4 rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700">
          {error}
        </div>
      )}

      {searched && !loading && !user && !error && (
        <div className="rounded-lg border border-gray-200 bg-gray-50 px-4 py-6 text-center text-sm text-gray-500">
          No user found with email "{email}"
        </div>
      )}

      {user && (
        <AdminListCard>
          <table className="w-full text-sm">
            <thead className="bg-gray-50 border-b border-gray-100">
              <tr>
                <th className="text-left px-6 py-3 font-medium text-gray-600">User</th>
                <th className="text-left px-6 py-3 font-medium text-gray-600">Role</th>
                <th className="text-left px-6 py-3 font-medium text-gray-600">Status</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-50">
              <tr className="hover:bg-gray-50">
                <td className="px-6 py-4">
                  <div>
                    <p className="font-medium text-gray-900">{user.username}</p>
                    <p className="text-xs text-gray-500">{user.email}</p>
                  </div>
                </td>
                <td className="px-6 py-4">
                  <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium capitalize bg-gray-100 text-gray-800">
                    {user.role}
                  </span>
                </td>
                <td className="px-6 py-4">
                  <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium capitalize ${
                    user.status === 'active' ? 'bg-green-100 text-green-800' : 
                    user.status === 'invited' ? 'bg-yellow-100 text-yellow-800' : 
                    'bg-red-100 text-red-800'
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
