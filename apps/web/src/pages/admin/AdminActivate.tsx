import React, { useEffect, useMemo, useState } from 'react'
import { useNavigate, useSearchParams } from 'react-router-dom'
import api from '../../lib/api'
import { getApiErrorMessage } from '../../lib/apiError'
import { useApi } from '../../hooks/useApi'
import type { StatusPageSettings } from '../../types'

const DEFAULT_PAGE_TITLE = 'Statora'

export default function AdminActivate() {
  const [searchParams] = useSearchParams()
  const navigate = useNavigate()
  const token = useMemo(() => searchParams.get('token') || '', [searchParams])

  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [confirmPassword, setConfirmPassword] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')
  const { data: settingsData } = useApi<StatusPageSettings>('/status/settings')
  const pageTitle = settingsData?.head?.title?.trim() || DEFAULT_PAGE_TITLE

  useEffect(() => {
    document.title = `${pageTitle} - Admin Panel`
  }, [pageTitle])

  async function onSubmit(e: React.FormEvent<HTMLFormElement>) {
    e.preventDefault()
    setError('')

    if (!token) {
      setError('Activation token is missing.')
      return
    }

    if (password !== confirmPassword) {
      setError('Passwords do not match.')
      return
    }

    setSubmitting(true)
    try {
      await api.post('/users/invitations/activate', {
        token,
        username,
        password,
      })
      navigate('/admin/login', { replace: true })
    } catch (err: unknown) {
      setError(getApiErrorMessage(err, 'Activation failed. Please try again.'))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="statora-shell flex items-center justify-center p-4">
      <div className="statora-card w-full max-w-md p-6">
        <p className="statora-kicker mb-2">Invitation setup</p>
        <h1 className="text-xl font-semibold tracking-tight text-slate-950 dark:text-slate-55 mb-1">Activate User Account</h1>
        <p className="statora-muted mb-6">Create your username and password to complete activation.</p>

        {error && (
          <div className="statora-alert-error mb-4">
            {error}
          </div>
        )}

        <form className="space-y-4" onSubmit={onSubmit}>
          <div>
            <label className="mb-1 block text-sm font-medium text-slate-700 dark:text-slate-300">Username</label>
            <input
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              required
              disabled={submitting}
              className="statora-input"
            />
          </div>

          <div>
            <label className="mb-1 block text-sm font-medium text-slate-700 dark:text-slate-300">Password</label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              minLength={8}
              required
              disabled={submitting}
              className="statora-input"
            />
          </div>

          <div>
            <label className="mb-1 block text-sm font-medium text-slate-700 dark:text-slate-300">Confirm Password</label>
            <input
              type="password"
              value={confirmPassword}
              onChange={(e) => setConfirmPassword(e.target.value)}
              minLength={8}
              required
              disabled={submitting}
              className="statora-input"
            />
          </div>

          <button
            type="submit"
            disabled={submitting}
            className="statora-btn-primary w-full"
          >
            {submitting ? 'Activating...' : 'Activate Account'}
          </button>
        </form>
      </div>
    </div>
  )
}
