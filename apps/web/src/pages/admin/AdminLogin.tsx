import React, { useEffect, useMemo, useState } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import api from '../../lib/api'
import { getApiErrorMessage } from '../../lib/apiError'
import { clearAuthSession, setAuthSession } from '../../lib/auth'
import { useApi } from '../../hooks/useApi'
import type { LoginResponse, MfaVerifyResponse, StatusPageSettings, User } from '../../types'

const DEFAULT_PAGE_TITLE = 'Statora'

export default function AdminLogin() {
  const navigate = useNavigate()
  const location = useLocation()
  const { data: settingsData } = useApi<StatusPageSettings>('/status/settings')
  const pageTitle = settingsData?.head?.title?.trim() || DEFAULT_PAGE_TITLE
  const ssoError = useMemo(() => {
    const code = new URLSearchParams(location.search).get('error')
    switch (code) {
      case 'sso_not_configured':
        return 'SSO is not configured.'
      case 'sso_disabled':
        return 'SSO is currently disabled.'
      case 'user_not_found':
        return 'No user account matches this SSO login.'
      case 'sso_not_allowed':
        return 'SSO is not enabled for this account.'
      case 'invalid_token':
        return 'The SSO token is invalid or expired.'
      default:
        return ''
    }
  }, [location.search])
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [mfaCode, setMfaCode] = useState('')
  const [mfaError, setMfaError] = useState('')
  const [mfaLoading, setMfaLoading] = useState(false)
  const [pendingToken, setPendingToken] = useState<string | null>(null)
  const [pendingUser, setPendingUser] = useState<User | null>(null)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    setError('')
    setLoading(true)
    try {
      const res = await api.post<LoginResponse>('/auth/login', { email, password })

      if (res.data.mfaRequired) {
        setPendingToken(res.data.token)
        setPendingUser(res.data.user)
        setMfaCode('')
        setMfaError('')
      } else {
        setAuthSession(res.data.token, {
          ...res.data.user,
          mfaVerified: true,
        })
        navigate('/admin')
      }
    } catch (err: unknown) {
      setError(getApiErrorMessage(err, 'Invalid credentials'))
    } finally {
      setLoading(false)
    }
  }

  async function handleMfaVerify(e: React.FormEvent) {
    e.preventDefault()
    if (!pendingToken) {
      setMfaError('Your MFA session expired. Please sign in again.')
      return
    }

    setMfaError('')
    setMfaLoading(true)

    try {
      const res = await api.post<MfaVerifyResponse>(
        '/auth/mfa/verify',
        { code: mfaCode },
        { headers: { Authorization: `Bearer ${pendingToken}` } }
      )

      setAuthSession(res.data.token, {
        ...res.data.user,
        mfaEnabled: true,
        mfaVerified: res.data.mfaVerified,
      })

      setPendingToken(null)
      setPendingUser(null)
      navigate('/admin')
    } catch (err: unknown) {
      setMfaError(getApiErrorMessage(err, 'Invalid verification code'))
    } finally {
      setMfaLoading(false)
    }
  }

  function handleMfaCancel() {
    setPendingToken(null)
    setPendingUser(null)
    setMfaCode('')
    setMfaError('')
    clearAuthSession()
  }

  useEffect(() => {
    document.title = `${pageTitle} - Admin Panel`
  }, [pageTitle])

  return (
    <>
      <div className="statusforge-shell flex items-center justify-center px-4 py-12">
        <div className="statusforge-card w-full max-w-sm p-8">
          <p className="statusforge-kicker mb-2">Admin access</p>
          <h1 className="text-2xl font-bold tracking-tight text-slate-950 mb-2">User Login</h1>
          <p className="statusforge-muted mb-6">{pageTitle}</p>

          {(error || ssoError) && (
            <div className="statusforge-alert-error mb-4">
              {error || ssoError}
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-4">
            <div>
              <label className="block text-sm font-medium text-slate-700 mb-1">Email</label>
              <input
                type="email"
                required
                value={email}
                onChange={e => setEmail(e.target.value)}
                className="statusforge-input"
                placeholder="admin@statusplatform.com"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-slate-700 mb-1">Password</label>
              <input
                type="password"
                required
                value={password}
                onChange={e => setPassword(e.target.value)}
                className="statusforge-input"
                placeholder="••••••••"
              />
            </div>
            <button
              type="submit"
              disabled={loading}
              className="statusforge-btn-primary w-full"
            >
              {loading ? 'Signing in...' : 'Sign In'}
            </button>
          </form>
        </div>
      </div>

      {pendingToken && (
        <div className="fixed inset-0 z-50 bg-slate-950/55 flex items-center justify-center px-4 backdrop-blur-sm">
          <div className="statusforge-card w-full max-w-sm p-6">
            <h2 className="text-xl font-semibold text-slate-950 mb-1">MFA Verification Required</h2>
            <p className="text-sm text-slate-600 mb-4">
              Enter the 6-digit code from your authenticator app for {pendingUser?.email ?? 'your account'}.
            </p>

            {mfaError && (
              <div className="statusforge-alert-error mb-4">
                {mfaError}
              </div>
            )}

            <form onSubmit={handleMfaVerify} className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-slate-700 mb-1">Authenticator Code</label>
                <input
                  type="text"
                  inputMode="numeric"
                  pattern="[0-9]{6}"
                  maxLength={6}
                  required
                  value={mfaCode}
                  onChange={(e) => setMfaCode(e.target.value.replace(/\D/g, '').slice(0, 6))}
                  className="statusforge-input tracking-widest"
                  placeholder="123456"
                />
              </div>

              <div className="flex items-center gap-3">
                <button
                  type="submit"
                  disabled={mfaLoading}
                  className="statusforge-btn-primary flex-1"
                >
                  {mfaLoading ? 'Verifying...' : 'Verify and Continue'}
                </button>
                <button
                  type="button"
                  onClick={handleMfaCancel}
                  disabled={mfaLoading}
                  className="statusforge-btn-secondary px-3"
                >
                  Cancel
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </>
  )
}
