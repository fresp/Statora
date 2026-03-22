import React, { useEffect, useState } from 'react'
import type { StatusPageSettings, StatusPageSettingsPatchRequest } from '../../types'
import api from '../../lib/api'
import { CheckCircle, AlertTriangle, AlertCircle, XCircle, Wrench } from 'lucide-react'

type ThemePalette = StatusPageSettings['theme']['light']

const THEME_PRESETS: Record<string, { label: string; light: ThemePalette; dark: ThemePalette }> = {
  default: {
    label: 'Default',
    light: { primaryColor: '#16a34a', backgroundColor: '#f9fafb', textColor: '#111827', accentColor: '#0ea5e9' },
    dark: { primaryColor: '#22c55e', backgroundColor: '#0b1220', textColor: '#e5e7eb', accentColor: '#38bdf8' },
  },
  ocean: {
    label: 'Ocean',
    light: { primaryColor: '#0ea5e9', backgroundColor: '#f0f9ff', textColor: '#0f172a', accentColor: '#14b8a6' },
    dark: { primaryColor: '#38bdf8', backgroundColor: '#082f49', textColor: '#e0f2fe', accentColor: '#2dd4bf' },
  },
  graphite: {
    label: 'Graphite',
    light: { primaryColor: '#334155', backgroundColor: '#f8fafc', textColor: '#0f172a', accentColor: '#6366f1' },
    dark: { primaryColor: '#64748b', backgroundColor: '#020617', textColor: '#e2e8f0', accentColor: '#818cf8' },
  },
}

const DEFAULT_SETTINGS: StatusPageSettings = {
  head: {
    title: 'Status Platform',
    description: 'Live system status and incident updates.',
    keywords: 'status, uptime, incidents, maintenance',
    faviconUrl: '/vite.svg',
    metaTags: {},
  },
  branding: {
    siteName: 'System Status',
    logoUrl: '',
    backgroundImageUrl: '',
    heroImageUrl: '',
  },
  theme: {
    preset: 'default',
    mode: 'system',
    light: {
      primaryColor: '#16a34a',
      backgroundColor: '#f9fafb',
      textColor: '#111827',
      accentColor: '#0ea5e9',
    },
    dark: {
      primaryColor: '#22c55e',
      backgroundColor: '#0b1220',
      textColor: '#e5e7eb',
      accentColor: '#38bdf8',
    },
    typography: {
      fontFamily: 'Inter, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif',
      fontScale: 'md',
    },
  },
  layout: { variant: 'classic' },
  footer: {
    text: '',
    showPoweredBy: true,
  },
  customCss: '',
  updatedAt: new Date().toISOString(),
  createdAt: new Date().toISOString(),
}

function normalizeSettings(input?: StatusPageSettings | null): StatusPageSettings {
  if (!input) {
    return DEFAULT_SETTINGS
  }

  const mode = input.theme?.mode === 'light' || input.theme?.mode === 'dark' || input.theme?.mode === 'system'
    ? input.theme.mode
    : DEFAULT_SETTINGS.theme.mode

  const fontScale = input.theme?.typography?.fontScale === 'sm' || input.theme?.typography?.fontScale === 'md' || input.theme?.typography?.fontScale === 'lg'
    ? input.theme.typography.fontScale
    : DEFAULT_SETTINGS.theme.typography.fontScale

  const variant = input.layout?.variant === 'classic' || input.layout?.variant === 'compact' || input.layout?.variant === 'minimal' || input.layout?.variant === 'cards'
    ? input.layout.variant
    : DEFAULT_SETTINGS.layout.variant

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
      preset: input.theme?.preset ?? DEFAULT_SETTINGS.theme.preset,
      mode,
      light: {
        primaryColor: input.theme?.light?.primaryColor ?? DEFAULT_SETTINGS.theme.light.primaryColor,
        backgroundColor: input.theme?.light?.backgroundColor ?? DEFAULT_SETTINGS.theme.light.backgroundColor,
        textColor: input.theme?.light?.textColor ?? DEFAULT_SETTINGS.theme.light.textColor,
        accentColor: input.theme?.light?.accentColor ?? DEFAULT_SETTINGS.theme.light.accentColor,
      },
      dark: {
        primaryColor: input.theme?.dark?.primaryColor ?? DEFAULT_SETTINGS.theme.dark.primaryColor,
        backgroundColor: input.theme?.dark?.backgroundColor ?? DEFAULT_SETTINGS.theme.dark.backgroundColor,
        textColor: input.theme?.dark?.textColor ?? DEFAULT_SETTINGS.theme.dark.textColor,
        accentColor: input.theme?.dark?.accentColor ?? DEFAULT_SETTINGS.theme.dark.accentColor,
      },
      typography: {
        fontFamily: input.theme?.typography?.fontFamily ?? DEFAULT_SETTINGS.theme.typography.fontFamily,
        fontScale,
      },
    },
    layout: {
      variant,
    },
    footer: {
      text: input.footer?.text ?? '',
      showPoweredBy: input.footer?.showPoweredBy ?? true,
    },
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
  const [metaTagsText, setMetaTagsText] = useState('')
  const [systemPrefersDark, setSystemPrefersDark] = useState<boolean>(() => {
    if (typeof window === 'undefined' || typeof window.matchMedia !== 'function') {
      return false
    }
    return window.matchMedia('(prefers-color-scheme: dark)').matches
  })

  useEffect(() => {
    if (typeof window === 'undefined' || typeof window.matchMedia !== 'function') {
      return
    }

    const media = window.matchMedia('(prefers-color-scheme: dark)')
    const listener = (event: MediaQueryListEvent) => {
      setSystemPrefersDark(event.matches)
    }

    media.addEventListener('change', listener)
    return () => media.removeEventListener('change', listener)
  }, [])

  const previewMode = settings.theme.mode === 'system'
    ? (systemPrefersDark ? 'dark' : 'light')
    : settings.theme.mode
  const previewPalette = previewMode === 'dark' ? settings.theme.dark : settings.theme.light
  const previewStyle: React.CSSProperties = {
    backgroundColor: previewPalette.backgroundColor,
    color: previewPalette.textColor,
    fontFamily: settings.theme.typography.fontFamily,
    backgroundImage: settings.branding.backgroundImageUrl
      ? `linear-gradient(rgba(0,0,0,0.16), rgba(0,0,0,0.16)), url(${settings.branding.backgroundImageUrl})`
      : undefined,
    backgroundSize: settings.branding.backgroundImageUrl ? 'cover' : undefined,
    backgroundPosition: settings.branding.backgroundImageUrl ? 'center' : undefined,
  }

  async function loadSettings() {
    try {
      setLoading(true)
      setError(null)
      const res = await api.get<StatusPageSettings>('/settings/status-page')
      const normalized = normalizeSettings(res.data)
      setSettings(normalized)
      setMetaTagsText(metaTagsToText(normalized.head.metaTags || {}))
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to load settings')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadSettings()
  }, [])

  async function handleSave(e: React.FormEvent) {
    e.preventDefault()
    setSaving(true)
    setError(null)
    setSuccess(null)

    try {
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
          mode: settings.theme.mode,
          light: {
            primaryColor: settings.theme.light.primaryColor,
            backgroundColor: settings.theme.light.backgroundColor,
            textColor: settings.theme.light.textColor,
            accentColor: settings.theme.light.accentColor,
          },
          dark: {
            primaryColor: settings.theme.dark.primaryColor,
            backgroundColor: settings.theme.dark.backgroundColor,
            textColor: settings.theme.dark.textColor,
            accentColor: settings.theme.dark.accentColor,
          },
          typography: {
            fontFamily: settings.theme.typography.fontFamily,
            fontScale: settings.theme.typography.fontScale,
          },
        },
        layout: {
          variant: settings.layout.variant,
        },
        footer: {
          text: settings.footer.text,
          showPoweredBy: settings.footer.showPoweredBy,
        },
        customCss: settings.customCss,
      }

      const res = await api.patch<StatusPageSettings>('/settings/status-page', payload)
      const normalized = normalizeSettings(res.data)
      setSettings(normalized)
      setMetaTagsText(metaTagsToText(normalized.head.metaTags || {}))
      setSuccess('Settings saved successfully')
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to save settings')
    } finally {
      setSaving(false)
    }
  }

  if (loading) {
    return <div className="p-8 text-sm text-gray-500">Loading settings...</div>
  }

  return (
    <div className="p-8 max-w-5xl">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900">Status Page Settings</h1>
        <p className="text-sm text-gray-500 mt-1">
          Configure SEO, branding, layout, footer, theme, and custom CSS for the public status page.
        </p>
      </div>

      {error && <p className="mb-4 text-sm text-red-600 bg-red-50 rounded-lg px-3 py-2">{error}</p>}
      {success && <p className="mb-4 text-sm text-green-700 bg-green-50 rounded-lg px-3 py-2">{success}</p>}

      <form onSubmit={handleSave} className="space-y-6">
        <section className="bg-white border border-gray-200 rounded-xl p-5 space-y-4">
          <h2 className="text-lg font-semibold text-gray-900">Head & SEO</h2>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Page Title</label>
            <input
              className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm"
              value={settings.head.title}
              onChange={(e) => setSettings(prev => ({ ...prev, head: { ...prev.head, title: e.target.value } }))}
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Description</label>
            <input
              className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm"
              value={settings.head.description}
              onChange={(e) => setSettings(prev => ({ ...prev, head: { ...prev.head, description: e.target.value } }))}
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Keywords</label>
            <input
              className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm"
              value={settings.head.keywords}
              onChange={(e) => setSettings(prev => ({ ...prev, head: { ...prev.head, keywords: e.target.value } }))}
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Favicon URL</label>
            <input
              className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm"
              value={settings.head.faviconUrl}
              onChange={(e) => setSettings(prev => ({ ...prev, head: { ...prev.head, faviconUrl: e.target.value } }))}
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Additional Meta Tags</label>
            <textarea
              className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm min-h-[110px]"
              placeholder={'og:title: My Status Page\nog:site_name: StatusForge'}
              value={metaTagsText}
              onChange={(e) => setMetaTagsText(e.target.value)}
            />
            <p className="text-xs text-gray-500 mt-1">One tag per line using format: key: value</p>
          </div>
        </section>

        <section className="bg-white border border-gray-200 rounded-xl p-5 space-y-4">
          <h2 className="text-lg font-semibold text-gray-900">Branding Assets</h2>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Site Name</label>
            <input
              className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm"
              value={settings.branding.siteName}
              onChange={(e) => setSettings(prev => ({ ...prev, branding: { ...prev.branding, siteName: e.target.value } }))}
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Logo URL</label>
            <input
              className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm"
              value={settings.branding.logoUrl}
              onChange={(e) => setSettings(prev => ({ ...prev, branding: { ...prev.branding, logoUrl: e.target.value } }))}
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Background Image URL</label>
            <input
              className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm"
              value={settings.branding.backgroundImageUrl}
              onChange={(e) => setSettings(prev => ({ ...prev, branding: { ...prev.branding, backgroundImageUrl: e.target.value } }))}
            />
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Hero Image URL</label>
            <input
              className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm"
              value={settings.branding.heroImageUrl}
              onChange={(e) => setSettings(prev => ({ ...prev, branding: { ...prev.branding, heroImageUrl: e.target.value } }))}
            />
          </div>
        </section>

        <section className="bg-white border border-gray-200 rounded-xl p-5 space-y-4">
          <h2 className="text-lg font-semibold text-gray-900">Visual Theme</h2>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Preset</label>
              <select
                className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm"
                value={settings.theme.preset}
                onChange={(e) => {
                  const selectedPreset = THEME_PRESETS[e.target.value] || THEME_PRESETS.default
                  setSettings(prev => ({
                    ...prev,
                    theme: {
                      ...prev.theme,
                      preset: e.target.value,
                      light: { ...selectedPreset.light },
                      dark: { ...selectedPreset.dark },
                    },
                  }))
                }}
              >
                {Object.entries(THEME_PRESETS).map(([value, preset]) => (
                  <option key={value} value={value}>{preset.label}</option>
                ))}
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Mode</label>
              <select
                className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm"
                value={settings.theme.mode}
                onChange={(e) => setSettings(prev => ({ ...prev, theme: { ...prev.theme, mode: e.target.value as 'light' | 'dark' | 'system' } }))}
              >
                <option value="system">System</option>
                <option value="light">Light</option>
                <option value="dark">Dark</option>
              </select>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Layout Variant</label>
              <select
                className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm"
                value={settings.layout.variant}
                onChange={(e) => setSettings(prev => ({ ...prev, layout: { variant: e.target.value as 'classic' | 'compact' | 'minimal' | 'cards' } }))}
              >
                <option value="classic">Classic</option>
                <option value="compact">Compact</option>
                <option value="minimal">Minimal</option>
                <option value="cards">Cards</option>
              </select>
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">Light Palette</label>
              <div className="grid grid-cols-2 gap-3">
                <input className="border border-gray-300 rounded-lg px-3 py-2 text-sm" placeholder="Primary" value={settings.theme.light.primaryColor} onChange={(e) => setSettings(prev => ({ ...prev, theme: { ...prev.theme, light: { ...prev.theme.light, primaryColor: e.target.value } } }))} />
                <input className="border border-gray-300 rounded-lg px-3 py-2 text-sm" placeholder="Background" value={settings.theme.light.backgroundColor} onChange={(e) => setSettings(prev => ({ ...prev, theme: { ...prev.theme, light: { ...prev.theme.light, backgroundColor: e.target.value } } }))} />
                <input className="border border-gray-300 rounded-lg px-3 py-2 text-sm" placeholder="Text" value={settings.theme.light.textColor} onChange={(e) => setSettings(prev => ({ ...prev, theme: { ...prev.theme, light: { ...prev.theme.light, textColor: e.target.value } } }))} />
                <input className="border border-gray-300 rounded-lg px-3 py-2 text-sm" placeholder="Accent" value={settings.theme.light.accentColor} onChange={(e) => setSettings(prev => ({ ...prev, theme: { ...prev.theme, light: { ...prev.theme.light, accentColor: e.target.value } } }))} />
              </div>
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">Dark Palette</label>
              <div className="grid grid-cols-2 gap-3">
                <input className="border border-gray-300 rounded-lg px-3 py-2 text-sm" placeholder="Primary" value={settings.theme.dark.primaryColor} onChange={(e) => setSettings(prev => ({ ...prev, theme: { ...prev.theme, dark: { ...prev.theme.dark, primaryColor: e.target.value } } }))} />
                <input className="border border-gray-300 rounded-lg px-3 py-2 text-sm" placeholder="Background" value={settings.theme.dark.backgroundColor} onChange={(e) => setSettings(prev => ({ ...prev, theme: { ...prev.theme, dark: { ...prev.theme.dark, backgroundColor: e.target.value } } }))} />
                <input className="border border-gray-300 rounded-lg px-3 py-2 text-sm" placeholder="Text" value={settings.theme.dark.textColor} onChange={(e) => setSettings(prev => ({ ...prev, theme: { ...prev.theme, dark: { ...prev.theme.dark, textColor: e.target.value } } }))} />
                <input className="border border-gray-300 rounded-lg px-3 py-2 text-sm" placeholder="Accent" value={settings.theme.dark.accentColor} onChange={(e) => setSettings(prev => ({ ...prev, theme: { ...prev.theme, dark: { ...prev.theme.dark, accentColor: e.target.value } } }))} />
              </div>
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Font Family</label>
              <input
                className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm"
                value={settings.theme.typography.fontFamily}
                onChange={(e) => setSettings(prev => ({ ...prev, theme: { ...prev.theme, typography: { ...prev.theme.typography, fontFamily: e.target.value } } }))}
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">Font Scale</label>
              <select
                className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm"
                value={settings.theme.typography.fontScale}
                onChange={(e) => setSettings(prev => ({ ...prev, theme: { ...prev.theme, typography: { ...prev.theme.typography, fontScale: e.target.value as 'sm' | 'md' | 'lg' } } }))}
              >
                <option value="sm">Small</option>
                <option value="md">Medium</option>
                <option value="lg">Large</option>
              </select>
            </div>
          </div>

          <div className="rounded-xl border border-gray-200 p-4 space-y-3">
            <h3 className="text-sm font-semibold text-gray-900">Live Preview</h3>
            <div className="rounded-lg p-4" style={previewStyle}>
              <div className="rounded-lg p-3 text-white flex items-center justify-between" style={{ backgroundColor: previewPalette.primaryColor }}>
                <div className="flex items-center gap-2">
                  {settings.branding.logoUrl && <img src={settings.branding.logoUrl} alt="logo" className="w-6 h-6 rounded object-contain" />}
                  <span className="font-semibold">{settings.branding.siteName || 'System Status'}</span>
                </div>
                <div className="flex items-center gap-2 text-xs">
                  <CheckCircle className="w-4 h-4" />
                  <span>All systems operational</span>
                </div>
              </div>
              {settings.branding.heroImageUrl && <img src={settings.branding.heroImageUrl} alt="hero" className="w-full h-24 object-cover rounded-md mt-3" />}
              <div className="grid grid-cols-2 md:grid-cols-5 gap-2 mt-3 text-xs">
                <div className="rounded p-2 border" style={{ borderColor: previewPalette.accentColor }}><CheckCircle className="w-4 h-4" /> Operational</div>
                <div className="rounded p-2 border" style={{ borderColor: previewPalette.accentColor }}><AlertTriangle className="w-4 h-4" /> Degraded</div>
                <div className="rounded p-2 border" style={{ borderColor: previewPalette.accentColor }}><AlertCircle className="w-4 h-4" /> Partial</div>
                <div className="rounded p-2 border" style={{ borderColor: previewPalette.accentColor }}><XCircle className="w-4 h-4" /> Major</div>
                <div className="rounded p-2 border" style={{ borderColor: previewPalette.accentColor }}><Wrench className="w-4 h-4" /> Maint.</div>
              </div>
            </div>
          </div>
        </section>

        <section className="bg-white border border-gray-200 rounded-xl p-5 space-y-4">
          <h2 className="text-lg font-semibold text-gray-900">Footer & Custom CSS</h2>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Footer Text</label>
            <input
              className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm"
              value={settings.footer.text}
              onChange={(e) => setSettings(prev => ({ ...prev, footer: { ...prev.footer, text: e.target.value } }))}
            />
          </div>
          <div className="flex items-center gap-2">
            <input
              id="show-powered"
              type="checkbox"
              checked={settings.footer.showPoweredBy}
              onChange={(e) => setSettings(prev => ({ ...prev, footer: { ...prev.footer, showPoweredBy: e.target.checked } }))}
            />
            <label htmlFor="show-powered" className="text-sm text-gray-700">Show “Powered by” text</label>
          </div>
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-1">Custom CSS</label>
            <textarea
              className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm min-h-[160px] font-mono"
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
