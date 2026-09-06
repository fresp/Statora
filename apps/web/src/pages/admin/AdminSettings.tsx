import React, { useEffect, useState } from 'react'
import type {
  MailProviderType,
  MailSettings,
  SMTPSettings,
  SendGridSettings,
  StatusPageSettings,
  StatusPageSettingsPatchRequest,
  StatusPageThemePresetSummary,
} from '../../types'
import api from '../../lib/api'
import { getApiErrorMessage } from '../../lib/apiError'
import { getThemePresets, loadThemePresetStylesheet, normalizeThemePresetSelection } from '../../lib/themePresetLoader'
import { CheckCircle, Mail, Send } from 'lucide-react'

const DEFAULT_MAIL_SETTINGS: MailSettings = {
  provider: 'none',
  smtp: {
    host: '',
    port: 587,
    username: '',
    fromEmail: '',
    fromName: '',
    encryption: 'starttls',
  },
  sendgrid: {
    fromEmail: '',
    fromName: '',
  },
  baseUrl: '',
}

const ADMIN_TITLE_SUFFIX = ' - Admin Panel'

const DEFAULT_SETTINGS: StatusPageSettings = {
  head: {
    title: 'Statora',
    description: 'Live system status and incident updates.',
    keywords: 'status, uptime, incidents, maintenance',
    faviconUrl: '/vite.svg',
    metaTags: {},
  },
  branding: {
    siteName: 'Statora',
    logoUrl: '',
    backgroundImageUrl: '',
    heroImageUrl: '',
  },
  theme: {
    preset: 'default.css',
    mode: 'light',
  },
  footer: {
    text: '',
    showPoweredBy: true,
  },
  sso: {
    enabled: false,
    provider: '',
    issuer: '',
    audience: '',
    algorithm: 'HS256',
    publicKeyPem: '',
    sharedSecret: '',
    hasSecret: false,
  },
  customCss: '',
  updatedAt: new Date().toISOString(),
  createdAt: new Date().toISOString(),
}

function normalizeMailSettings(input: Partial<MailSettings> | null | undefined): MailSettings {
  if (!input) {
    return DEFAULT_MAIL_SETTINGS
  }
  return {
    provider: (input.provider ?? 'none') as MailProviderType,
    smtp: {
      host: input.smtp?.host ?? '',
      port: input.smtp?.port ?? 587,
      username: input.smtp?.username ?? '',
      hasPassword: input.smtp?.hasPassword ?? false,
      fromEmail: input.smtp?.fromEmail ?? '',
      fromName: input.smtp?.fromName ?? '',
      encryption: input.smtp?.encryption ?? 'starttls',
    },
    sendgrid: {
      hasApiKey: input.sendgrid?.hasApiKey ?? false,
      fromEmail: input.sendgrid?.fromEmail ?? '',
      fromName: input.sendgrid?.fromName ?? '',
    },
    baseUrl: input.baseUrl ?? '',
  }
}

function normalizeSettings(input: StatusPageSettings | null | undefined, presets: StatusPageThemePresetSummary[]): StatusPageSettings {
  if (!input) {
    return DEFAULT_SETTINGS
  }

  const normalizedPreset = normalizeThemePresetSelection(input.theme?.preset || '', presets)

  return {
    head: {
      title: input.head?.title ?? DEFAULT_SETTINGS.head.title,
      description: input.head?.description ?? DEFAULT_SETTINGS.head.description,
      keywords: input.head?.keywords ?? DEFAULT_SETTINGS.head.keywords,
      faviconUrl: input.head?.faviconUrl ?? DEFAULT_SETTINGS.head.faviconUrl,
      metaTags: input.head?.metaTags || {},
    },
    branding: {
      siteName: input.branding?.siteName ?? DEFAULT_SETTINGS.branding.siteName,
      logoUrl: input.branding?.logoUrl ?? '',
      backgroundImageUrl: input.branding?.backgroundImageUrl ?? '',
      heroImageUrl: input.branding?.heroImageUrl ?? '',
    },
    theme: {
      preset: normalizedPreset,
      mode: input.theme?.mode ?? 'light',
    },
    footer: {
      text: input.footer?.text ?? '',
      showPoweredBy: input.footer?.showPoweredBy ?? true,
    },
    sso: {
      enabled: input.sso?.enabled ?? false,
      provider: input.sso?.provider ?? '',
      issuer: input.sso?.issuer ?? '',
      audience: input.sso?.audience ?? '',
      algorithm: input.sso?.algorithm ?? 'HS256',
      publicKeyPem: input.sso?.publicKeyPem ?? '',
      sharedSecret: '',
    },
    mail: normalizeMailSettings(input.mail),
    customCss: input.customCss ?? '',
    updatedAt: input.updatedAt ?? new Date().toISOString(),
    createdAt: input.createdAt ?? new Date().toISOString(),
  }
}

function parseMetaTagsText(value: string): Record<string, string> {
  const lines = value
    .split('\n')
    .map(line => line.trim())
    .filter(Boolean)

  const tags: Record<string, string> = {}
  for (const line of lines) {
    const separatorIndex = line.indexOf(':')
    if (separatorIndex <= 0) {
      continue
    }
    const key = line.slice(0, separatorIndex).trim()
    const tagValue = line.slice(separatorIndex + 1).trim()
    if (!key) {
      continue
    }
    tags[key] = tagValue
  }
  return tags
}

function metaTagsToText(metaTags: Record<string, string>): string {
  return Object.entries(metaTags)
    .map(([key, value]) => `${key}: ${value}`)
    .join('\n')
}

export default function AdminSettings() {
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState<string | null>(null)
  const [settings, setSettings] = useState<StatusPageSettings>(DEFAULT_SETTINGS)
  const [themePresets, setThemePresets] = useState<StatusPageThemePresetSummary[]>(() => getThemePresets().presets)
  const [themePresetNotice, setThemePresetNotice] = useState<string | null>(null)
  const [metaTagsText, setMetaTagsText] = useState('')
  const [smtpPassword, setSmtpPassword] = useState('')
  const [sendGridApiKey, setSendGridApiKey] = useState('')
  const [testEmailTo, setTestEmailTo] = useState('')
  const [testEmailSending, setTestEmailSending] = useState(false)
  const [testEmailResult, setTestEmailResult] = useState<{ ok: boolean; message: string } | null>(null)

  const previewStyle: React.CSSProperties = {
    backgroundColor: 'var(--bg)',
    color: 'var(--text)',
    fontFamily: 'var(--font-family)',
    backgroundImage: settings.branding.backgroundImageUrl
      ? `linear-gradient(var(--bg-image-overlay), var(--bg-image-overlay)), url(${settings.branding.backgroundImageUrl})`
      : undefined,
    backgroundSize: settings.branding.backgroundImageUrl ? 'cover' : undefined,
    backgroundPosition: settings.branding.backgroundImageUrl ? 'center' : undefined,
  }

  async function loadSettings() {
    try {
      setLoading(true)
      setError(null)
      const settingsRes = await api.get<StatusPageSettings>('/settings/status-page')
      const { presets, hasErrors } = getThemePresets()
      setThemePresetNotice(hasErrors ? 'Some local theme files are invalid or missing fields. Falling back to default values.' : null)
      setThemePresets(presets)
      const normalized = normalizeSettings(settingsRes.data, presets)
      if (normalized.sso) {
        normalized.sso.sharedSecret = ''
      }
      setSettings(normalized)
      setMetaTagsText(metaTagsToText(normalized.head.metaTags || {}))
    } catch (err: unknown) {
      setError(getApiErrorMessage(err, 'Failed to load settings'))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadSettings()
  }, [])

  useEffect(() => {
    document.title = `${settings.head.title}${ADMIN_TITLE_SUFFIX}`
  }, [settings.head.title])

  useEffect(() => {
    if (themePresets.length === 0) {
      return
    }

    loadThemePresetStylesheet(settings.theme.preset, themePresets).catch(() => { })
  }, [settings.theme.preset, themePresets])

  useEffect(() => {
    if (settings.theme.mode) {
      const mode = settings.theme.mode
      const root = document.documentElement
      if (mode === 'dark') {
        root.classList.add('dark')
      } else if (mode === 'light') {
        root.classList.remove('dark')
      } else if (mode === 'system') {
        const systemDark = window.matchMedia('(prefers-color-scheme: dark)').matches
        if (systemDark) {
          root.classList.add('dark')
        } else {
          root.classList.remove('dark')
        }
      }
    }
  }, [settings.theme.mode])

  async function handleSave(e: React.FormEvent) {
    e.preventDefault()
    setSaving(true)
    setError(null)
    setSuccess(null)

    try {
      const mailPatch: NonNullable<StatusPageSettingsPatchRequest['mail']> = {
        provider: settings.mail?.provider,
        baseUrl: settings.mail?.baseUrl,
        smtp: {
          host: settings.mail?.smtp.host,
          port: settings.mail?.smtp.port,
          username: settings.mail?.smtp.username,
          fromEmail: settings.mail?.smtp.fromEmail,
          fromName: settings.mail?.smtp.fromName,
          encryption: settings.mail?.smtp.encryption,
        },
        sendgrid: {
          fromEmail: settings.mail?.sendgrid.fromEmail,
          fromName: settings.mail?.sendgrid.fromName,
        },
      }
      if (smtpPassword.trim() !== '') {
        mailPatch.smtp = { ...mailPatch.smtp, password: smtpPassword.trim() }
      }
      if (sendGridApiKey.trim() !== '') {
        mailPatch.sendgrid = { ...mailPatch.sendgrid, apiKey: sendGridApiKey.trim() }
      }

      const payload: StatusPageSettingsPatchRequest = {
        head: {
          title: settings.head.title,
          description: settings.head.description,
          keywords: settings.head.keywords,
          faviconUrl: settings.head.faviconUrl,
          metaTags: parseMetaTagsText(metaTagsText),
        },
        branding: {
          siteName: settings.branding.siteName,
          logoUrl: settings.branding.logoUrl,
          backgroundImageUrl: settings.branding.backgroundImageUrl,
          heroImageUrl: settings.branding.heroImageUrl,
        },
        theme: {
          preset: settings.theme.preset,
          mode: settings.theme.mode || 'light',
        },
        footer: {
          text: settings.footer.text,
          showPoweredBy: settings.footer.showPoweredBy,
        },
        sso: {
          enabled: settings.sso?.enabled,
          provider: settings.sso?.provider,
          issuer: settings.sso?.issuer,
          audience: settings.sso?.audience,
          algorithm: settings.sso?.algorithm,
          sharedSecret: settings.sso?.sharedSecret,
          publicKeyPem: settings.sso?.publicKeyPem,
        },
        mail: mailPatch,
        customCss: settings.customCss,
      }
      const res = await api.patch<StatusPageSettings>('/settings/status-page', payload)
      const normalized = normalizeSettings(res.data, themePresets)
      if (normalized.sso) {
        normalized.sso.sharedSecret = ''
      }
      setSettings(normalized)
      setSmtpPassword('')
      setSendGridApiKey('')
      setMetaTagsText(metaTagsToText(normalized.head.metaTags || {}))
      setSuccess('Settings saved successfully')
    } catch (err: unknown) {
      setError(getApiErrorMessage(err, 'Failed to save settings'))
    } finally {
      setSaving(false)
    }
  }

  async function handleSendTestEmail() {
    if (testEmailTo.trim() === '') {
      setTestEmailResult({ ok: false, message: 'Enter a recipient email address.' })
      return
    }
    setTestEmailSending(true)
    setTestEmailResult(null)
    try {
      const res = await api.post<{ message: string }>('/settings/mail/test', { to: testEmailTo.trim() })
      setTestEmailResult({ ok: true, message: res.data.message || 'Test email sent' })
    } catch (err: unknown) {
      setTestEmailResult({ ok: false, message: getApiErrorMessage(err, 'Failed to send test email') })
    } finally {
      setTestEmailSending(false)
    }
  }

  if (loading) {
    return <div className="p-8 text-sm text-gray-500 dark:text-slate-400">Loading settings...</div>
  }

  return (
    <div className="p-8 max-w-5xl">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900 dark:text-slate-100">Status Page Settings</h1>
        <p className="text-sm text-gray-500 dark:text-slate-400 mt-1">
          Configure SEO, branding, footer, theme, and custom CSS for the public status page.
        </p>
      </div>

      {error && <p className="mb-4 text-sm text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-900/20 rounded-lg px-3 py-2">{error}</p>}
      {themePresetNotice && <p className="mb-4 text-sm text-amber-700 dark:text-amber-400 bg-amber-50 dark:bg-amber-900/20 rounded-lg px-3 py-2">{themePresetNotice}</p>}
      {success && <p className="mb-4 text-sm text-green-700 dark:text-green-400 bg-green-50 dark:bg-green-905/20 rounded-lg px-3 py-2">{success}</p>}

      <form onSubmit={handleSave} className="space-y-6">
        <section className="bg-white dark:bg-slate-900 border border-gray-200 dark:border-slate-800 rounded-xl p-5 space-y-4">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-slate-100">Head & SEO</h2>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1">Page Title</label>
            <input
              className="w-full border border-gray-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-slate-100 rounded-lg px-3 py-2 text-sm"
              value={settings.head.title}
              onChange={(e) => setSettings(prev => ({ ...prev, head: { ...prev.head, title: e.target.value } }))}
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1">Description</label>
            <input
              className="w-full border border-gray-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-slate-100 rounded-lg px-3 py-2 text-sm"
              value={settings.head.description}
              onChange={(e) => setSettings(prev => ({ ...prev, head: { ...prev.head, description: e.target.value } }))}
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1">Keywords</label>
            <input
              className="w-full border border-gray-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-slate-100 rounded-lg px-3 py-2 text-sm"
              value={settings.head.keywords}
              onChange={(e) => setSettings(prev => ({ ...prev, head: { ...prev.head, keywords: e.target.value } }))}
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1">Favicon URL</label>
            <input
              className="w-full border border-gray-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-slate-100 rounded-lg px-3 py-2 text-sm"
              value={settings.head.faviconUrl}
              onChange={(e) => setSettings(prev => ({ ...prev, head: { ...prev.head, faviconUrl: e.target.value } }))}
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1">Additional Meta Tags</label>
            <textarea
              className="w-full border border-gray-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-slate-100 rounded-lg px-3 py-2 text-sm min-h-[110px]"
              placeholder={'og:title: My Status Page\nog:site_name: Statora'}
              value={metaTagsText}
              onChange={(e) => setMetaTagsText(e.target.value)}
            />
            <p className="text-xs text-gray-500 dark:text-slate-400 mt-1">One tag per line using format: key: value</p>
          </div>
        </section>

        <section className="bg-white dark:bg-slate-900 border border-gray-200 dark:border-slate-800 rounded-xl p-5 space-y-4">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-slate-100">Branding Assets</h2>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1">Site Name</label>
            <input
              className="w-full border border-gray-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-slate-100 rounded-lg px-3 py-2 text-sm"
              value={settings.branding.siteName}
              onChange={(e) => setSettings(prev => ({ ...prev, branding: { ...prev.branding, siteName: e.target.value } }))}
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1">Logo URL</label>
            <input
              className="w-full border border-gray-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-slate-100 rounded-lg px-3 py-2 text-sm"
              value={settings.branding.logoUrl}
              onChange={(e) => setSettings(prev => ({ ...prev, branding: { ...prev.branding, logoUrl: e.target.value } }))}
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1">Background Image URL</label>
            <input
              className="w-full border border-gray-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-slate-100 rounded-lg px-3 py-2 text-sm"
              value={settings.branding.backgroundImageUrl}
              onChange={(e) => setSettings(prev => ({ ...prev, branding: { ...prev.branding, backgroundImageUrl: e.target.value } }))}
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1">Hero Image URL</label>
            <input
              className="w-full border border-gray-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-slate-100 rounded-lg px-3 py-2 text-sm"
              value={settings.branding.heroImageUrl}
              onChange={(e) => setSettings(prev => ({ ...prev, branding: { ...prev.branding, heroImageUrl: e.target.value } }))}
            />
          </div>
        </section>

        <section className="bg-white dark:bg-slate-900 border border-gray-200 dark:border-slate-800 rounded-xl p-5 space-y-4">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-slate-100">Visual Theme</h2>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1">Preset</label>
              <select
                className="w-full border border-gray-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-slate-100 rounded-lg px-3 py-2 text-sm"
                value={settings.theme.preset}
                onChange={(e) => {
                  const selectedPreset = normalizeThemePresetSelection(e.target.value, themePresets)
                  setSettings(prev => ({
                    ...prev,
                    theme: {
                      ...prev.theme,
                      preset: selectedPreset,
                    },
                  }))
                }}
              >
                {themePresets.map((preset) => (
                  <option key={preset.key} value={preset.key}>{preset.label}</option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1">Theme Mode</label>
              <select
                className="w-full border border-gray-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-slate-100 rounded-lg px-3 py-2 text-sm"
                value={settings.theme.mode || 'light'}
                onChange={(e) => {
                  const selectedMode = e.target.value as 'light' | 'dark' | 'system'
                  setSettings(prev => ({
                    ...prev,
                    theme: {
                      ...prev.theme,
                      mode: selectedMode,
                    },
                  }))
                }}
              >
                <option value="light">Light</option>
                <option value="dark">Dark</option>
                <option value="system">System Preference</option>
              </select>
            </div>
          </div>

          <div className="rounded-xl border border-gray-200 dark:border-slate-800 p-4 space-y-3">
            <h3 className="text-sm font-semibold text-gray-900 dark:text-slate-100">Live Preview</h3>
            <div className="rounded-lg p-4" style={previewStyle}>
              <div
                className="rounded-lg p-3 items-center justify-between"
                style={{
                  backgroundColor: 'var(--status-operational)',
                  color: 'var(--on-primary)',
                  boxShadow: 'inset 0 -3px 0 var(--color-accent)',
                }}
              >
                <div className="flex items-center gap-2">
                  {settings.branding.logoUrl && <img src={settings.branding.logoUrl} alt="logo" className="w-6 h-6 rounded object-contain" />}
                  <span className="font-semibold">{settings.branding.siteName || 'Statora'}</span>
                </div>
                <div className="flex items-center gap-2 text-xs">
                  <CheckCircle className="w-4 h-4" />
                  <span>All systems operational</span>
                  <div className="flex items-center gap-3 text-xl">
                    {settings.branding.heroImageUrl && (
                      <img
                        src={settings.branding.heroImageUrl}
                        alt="hero"
                        className="w-full h-24 object-cover rounded-md mt-3 border"
                        style={{ borderColor: 'var(--hero-image-border)' }}
                      />
                    )}
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>

        <section className="bg-white dark:bg-slate-900 border border-gray-200 dark:border-slate-800 rounded-xl p-5 space-y-4">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-slate-100">SSO Settings</h2>
          <div className="flex items-center gap-2">
            <input
              id="sso-enabled"
              type="checkbox"
              checked={settings.sso?.enabled ?? false}
              className="rounded"
              onChange={(e) => setSettings(prev => ({
                ...prev,
                sso: {
                  enabled: e.target.checked,
                  provider: prev.sso?.provider ?? '',
                  issuer: prev.sso?.issuer ?? '',
                  audience: prev.sso?.audience ?? '',
                  algorithm: prev.sso?.algorithm ?? 'HS256',
                  publicKeyPem: prev.sso?.publicKeyPem ?? '',
                  sharedSecret: prev.sso?.sharedSecret ?? '',
                  hasSecret: prev.sso?.hasSecret ?? false,
                },
              }))}
            />
            <label htmlFor="sso-enabled" className="text-sm text-gray-700 dark:text-slate-300 cursor-pointer">Enable SSO</label>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1">Provider</label>
            <input
              className="w-full border border-gray-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-slate-100 rounded-lg px-3 py-2 text-sm"
              value={settings.sso?.provider ?? ''}
              onChange={(e) => setSettings(prev => ({ ...prev, sso: { ...prev.sso!, provider: e.target.value } }))}
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1">Issuer</label>
            <input
              className="w-full border border-gray-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-slate-100 rounded-lg px-3 py-2 text-sm"
              value={settings.sso?.issuer ?? ''}
              onChange={(e) => setSettings(prev => ({ ...prev, sso: { ...prev.sso!, issuer: e.target.value } }))}
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1">Audience</label>
            <input
              className="w-full border border-gray-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-slate-100 rounded-lg px-3 py-2 text-sm"
              value={settings.sso?.audience ?? ''}
              onChange={(e) => setSettings(prev => ({ ...prev, sso: { ...prev.sso!, audience: e.target.value } }))}
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1">Algorithm</label>
            <select
              className="w-full border border-gray-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-slate-100 rounded-lg px-3 py-2 text-sm"
              value={settings.sso?.algorithm ?? 'HS256'}
              onChange={(e) => setSettings(prev => ({ ...prev, sso: { ...prev.sso!, algorithm: e.target.value as 'HS256' | 'RS256' } }))}
            >
              <option value="HS256">HS256</option>
              <option value="RS256">RS256</option>
            </select>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1">Shared Secret</label>
            <input
              type="password"
              className="w-full border border-gray-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-slate-100 rounded-lg px-3 py-2 text-sm"
              value={settings.sso?.sharedSecret ?? ''}
              placeholder={settings.sso?.hasSecret ? 'Stored secret exists; enter to replace' : 'Enter shared secret'}
              onChange={(e) => setSettings(prev => ({ ...prev, sso: { ...prev.sso!, sharedSecret: e.target.value } }))}
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1">RSA Public Key (PEM)</label>
            <textarea
              className="w-full border border-gray-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-slate-100 rounded-lg px-3 py-2 text-sm min-h-[120px] font-mono"
              value={settings.sso?.publicKeyPem ?? ''}
              onChange={(e) => setSettings(prev => ({ ...prev, sso: { ...prev.sso!, publicKeyPem: e.target.value } }))}
            />
          </div>
        </section>

        <section className="bg-white dark:bg-slate-900 border border-gray-200 dark:border-slate-800 rounded-xl p-5 space-y-4">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-slate-100">Mail Provider Settings</h2>
          <p className="text-sm text-gray-500 dark:text-slate-400">
            Used for subscriber verification emails and event notifications. Secrets are encrypted at rest.
          </p>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1">Provider</label>
              <select
                className="w-full border border-gray-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-slate-100 rounded-lg px-3 py-2 text-sm"
                value={settings.mail?.provider ?? 'none'}
                onChange={(e) => setSettings(prev => ({
                  ...prev,
                  mail: normalizeMailSettings({
                    provider: e.target.value as MailProviderType,
                    smtp: prev.mail?.smtp,
                    sendgrid: prev.mail?.sendgrid,
                    baseUrl: prev.mail?.baseUrl,
                  }),
                }))}
              >
                <option value="none">None</option>
                <option value="smtp">SMTP</option>
                <option value="sendgrid">SendGrid</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1">Status Page Base URL</label>
              <input
                className="w-full border border-gray-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-slate-100 rounded-lg px-3 py-2 text-sm"
                placeholder="https://status.example.com"
                value={settings.mail?.baseUrl ?? ''}
                onChange={(e) => setSettings(prev => ({
                  ...prev,
                  mail: { ...prev.mail!, baseUrl: e.target.value },
                }))}
              />
              <p className="text-xs text-gray-500 dark:text-slate-400 mt-1">Used to build verification and unsubscribe links.</p>
            </div>
          </div>

          {settings.mail?.provider === 'smtp' && (
            <div className="rounded-xl border border-gray-200 dark:border-slate-800 p-4 space-y-3">
              <h3 className="text-sm font-semibold text-gray-900 dark:text-slate-100">SMTP</h3>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1">Host</label>
                  <input
                    className="w-full border border-gray-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-slate-100 rounded-lg px-3 py-2 text-sm"
                    value={settings.mail.smtp.host}
                    onChange={(e) => setSettings(prev => ({ ...prev, mail: { ...prev.mail!, smtp: { ...prev.mail!.smtp, host: e.target.value } } }))}
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1">Port</label>
                  <input
                    type="number"
                    className="w-full border border-gray-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-slate-100 rounded-lg px-3 py-2 text-sm"
                    value={settings.mail.smtp.port}
                    onChange={(e) => setSettings(prev => ({ ...prev, mail: { ...prev.mail!, smtp: { ...prev.mail!.smtp, port: Number(e.target.value) || 0 } } }))}
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1">Username</label>
                  <input
                    className="w-full border border-gray-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-slate-100 rounded-lg px-3 py-2 text-sm"
                    value={settings.mail.smtp.username}
                    onChange={(e) => setSettings(prev => ({ ...prev, mail: { ...prev.mail!, smtp: { ...prev.mail!.smtp, username: e.target.value } } }))}
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1">Password</label>
                  <input
                    type="password"
                    className="w-full border border-gray-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-slate-100 rounded-lg px-3 py-2 text-sm"
                    placeholder={settings.mail.smtp.hasPassword ? 'Stored secret exists; enter to replace' : 'Enter password'}
                    value={smtpPassword}
                    onChange={(e) => setSmtpPassword(e.target.value)}
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1">From Email</label>
                  <input
                    type="email"
                    className="w-full border border-gray-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-slate-100 rounded-lg px-3 py-2 text-sm"
                    value={settings.mail.smtp.fromEmail}
                    onChange={(e) => setSettings(prev => ({ ...prev, mail: { ...prev.mail!, smtp: { ...prev.mail!.smtp, fromEmail: e.target.value } } }))}
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1">From Name</label>
                  <input
                    className="w-full border border-gray-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-slate-100 rounded-lg px-3 py-2 text-sm"
                    value={settings.mail.smtp.fromName}
                    onChange={(e) => setSettings(prev => ({ ...prev, mail: { ...prev.mail!, smtp: { ...prev.mail!.smtp, fromName: e.target.value } } }))}
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1">Encryption</label>
                  <select
                    className="w-full border border-gray-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-slate-100 rounded-lg px-3 py-2 text-sm"
                    value={settings.mail.smtp.encryption}
                    onChange={(e) => setSettings(prev => ({ ...prev, mail: { ...prev.mail!, smtp: { ...prev.mail!.smtp, encryption: e.target.value as 'starttls' | 'tls' | 'none' } } }))}
                  >
                    <option value="starttls">STARTTLS</option>
                    <option value="tls">TLS</option>
                    <option value="none">None</option>
                  </select>
                </div>
              </div>
            </div>
          )}

          {settings.mail?.provider === 'sendgrid' && (
            <div className="rounded-xl border border-gray-200 dark:border-slate-800 p-4 space-y-3">
              <h3 className="text-sm font-semibold text-gray-900 dark:text-slate-100">SendGrid</h3>
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1">API Key</label>
                  <input
                    type="password"
                    className="w-full border border-gray-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-slate-100 rounded-lg px-3 py-2 text-sm"
                    placeholder={settings.mail.sendgrid.hasApiKey ? 'Stored secret exists; enter to replace' : 'Enter API key'}
                    value={sendGridApiKey}
                    onChange={(e) => setSendGridApiKey(e.target.value)}
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1">From Email</label>
                  <input
                    type="email"
                    className="w-full border border-gray-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-slate-100 rounded-lg px-3 py-2 text-sm"
                    value={settings.mail.sendgrid.fromEmail}
                    onChange={(e) => setSettings(prev => ({ ...prev, mail: { ...prev.mail!, sendgrid: { ...prev.mail!.sendgrid, fromEmail: e.target.value } } }))}
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1">From Name</label>
                  <input
                    className="w-full border border-gray-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-slate-100 rounded-lg px-3 py-2 text-sm"
                    value={settings.mail.sendgrid.fromName}
                    onChange={(e) => setSettings(prev => ({ ...prev, mail: { ...prev.mail!, sendgrid: { ...prev.mail!.sendgrid, fromName: e.target.value } } }))}
                  />
                </div>
              </div>
            </div>
          )}

          <div className="rounded-xl border border-gray-200 dark:border-slate-800 p-4 space-y-3">
            <h3 className="text-sm font-semibold text-gray-900 dark:text-slate-100">Send Test Email</h3>
            <div className="flex items-center gap-2">
              <input
                type="email"
                className="flex-1 border border-gray-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-slate-100 rounded-lg px-3 py-2 text-sm"
                placeholder="you@example.com"
                value={testEmailTo}
                onChange={(e) => setTestEmailTo(e.target.value)}
              />
              <button
                type="button"
                onClick={handleSendTestEmail}
                disabled={testEmailSending || settings.mail?.provider === 'none'}
                className="flex items-center gap-2 bg-blue-600 hover:bg-blue-700 disabled:opacity-60 text-white rounded-lg px-4 py-2 text-sm font-medium"
              >
                <Send className="w-4 h-4" />
                {testEmailSending ? 'Sending...' : 'Send Test'}
              </button>
            </div>
            {settings.mail?.provider === 'none' && (
              <p className="text-xs text-gray-500 dark:text-slate-400">Select and save a mail provider first.</p>
            )}
            {testEmailResult && (
              <p className={`text-sm rounded-lg px-3 py-2 ${testEmailResult.ok ? 'text-green-700 dark:text-green-400 bg-green-50 dark:bg-green-900/20' : 'text-red-600 dark:text-red-400 bg-red-50 dark:bg-red-900/20'}`}>
                {testEmailResult.message}
              </p>
            )}
          </div>
        </section>

        <section className="bg-white dark:bg-slate-900 border border-gray-200 dark:border-slate-800 rounded-xl p-5 space-y-4">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-slate-100">Footer & Custom CSS</h2>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1">Footer Text</label>
            <input
              className="w-full border border-gray-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-slate-100 rounded-lg px-3 py-2 text-sm"
              value={settings.footer.text}
              onChange={(e) => setSettings(prev => ({ ...prev, footer: { ...prev.footer, text: e.target.value } }))}
            />
          </div>
          <div className="flex items-center gap-2">
            <input
              id="show-powered"
              type="checkbox"
              checked={settings.footer.showPoweredBy}
              className="rounded"
              onChange={(e) => setSettings(prev => ({ ...prev, footer: { ...prev.footer, showPoweredBy: e.target.checked } }))}
            />
            <label htmlFor="show-powered" className="text-sm text-gray-700 dark:text-slate-300 cursor-pointer">Show “Powered by” text</label>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 dark:text-slate-300 mb-1">Custom CSS</label>
            <textarea
              className="w-full border border-gray-300 dark:border-slate-700 bg-white dark:bg-slate-800 text-gray-900 dark:text-slate-100 rounded-lg px-3 py-2 text-sm min-h-[160px] font-mono"
              value={settings.customCss}
              onChange={(e) => setSettings(prev => ({ ...prev, customCss: e.target.value }))}
            />
          </div>
        </section>


        <div className="flex justify-end">
          <button
            type="submit"
            disabled={saving}
            className="bg-blue-600 hover:bg-blue-700 disabled:opacity-60 text-white rounded-lg px-5 py-2 text-sm font-medium"
          >
            {saving ? 'Saving...' : 'Save Settings'}
          </button>
        </div>
      </form>
    </div>
  )
}
