import { mkdir } from 'node:fs/promises'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { chromium } from 'playwright'

const __filename = fileURLToPath(import.meta.url)
const __dirname = path.dirname(__filename)

const BASE_WEB = process.env.DOCS_SCREENSHOTS_BASE_WEB || 'http://localhost:3000'
const BASE_API = process.env.DOCS_SCREENSHOTS_BASE_API || 'http://localhost:8080/api'
const ADMIN_EMAIL = process.env.DOCS_SCREENSHOTS_ADMIN_EMAIL || 'admin@statusplatform.com'
const ADMIN_PASSWORD = process.env.DOCS_SCREENSHOTS_ADMIN_PASSWORD || 'admin123'
const OUTPUT_DIR = path.resolve(__dirname, '../../../docs/screenshots')
const VIEWPORT = { width: 1440, height: 1800 }

async function apiRequest(route, { method = 'GET', token, body } = {}) {
  const headers = {}
  if (token) headers.Authorization = `Bearer ${token}`
  if (body !== undefined) headers['Content-Type'] = 'application/json'

  const response = await fetch(`${BASE_API}${route}`, {
    method,
    headers,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  })

  const text = await response.text()
  let data = null
  try {
    data = text ? JSON.parse(text) : null
  } catch {
    data = text
  }

  return { ok: response.ok, status: response.status, data }
}

function withLoopbackFallback(url) {
  if (!url.includes('localhost')) return [url]
  return [url, url.replace('localhost', '127.0.0.1')]
}

async function fetchFirstReachable(url, options) {
  let lastError = null
  for (const candidate of withLoopbackFallback(url)) {
    try {
      return await fetch(candidate, options)
    } catch (error) {
      lastError = error
    }
  }

  throw lastError ?? new Error(`Failed to reach ${url}`)
}

async function assertServerAvailability() {
  const backend = await fetchFirstReachable(`${BASE_API.replace(/\/api$/, '')}/health`)
  if (!backend.ok) {
    throw new Error(`Backend health check failed (${backend.status}). Expected service at ${BASE_API.replace(/\/api$/, '')}`)
  }

  const frontend = await fetchFirstReachable(BASE_WEB)
  if (!frontend.ok) {
    throw new Error(`Frontend check failed (${frontend.status}). Expected Vite app at ${BASE_WEB}`)
  }
}

async function loginAdmin() {
  const login = await apiRequest('/auth/login', {
    method: 'POST',
    body: {
      email: ADMIN_EMAIL,
      password: ADMIN_PASSWORD,
    },
  })

  if (!login.ok || !login.data?.token) {
    throw new Error(`Admin login failed (${login.status}): ${JSON.stringify(login.data)}`)
  }

  return {
    token: login.data.token,
    profile: {
      ...(login.data.admin ?? login.data.user ?? {}),
      role: 'admin',
      mfaVerified: true,
    },
  }
}

async function ensureDemoData(adminToken) {
  const componentsRes = await apiRequest('/components', { token: adminToken })
  if (!componentsRes.ok || !Array.isArray(componentsRes.data)) {
    throw new Error(`Failed to read components (${componentsRes.status})`)
  }

  let component = componentsRes.data.find((item) => item.name === 'Docs Demo API')
  if (!component) {
    const created = await apiRequest('/components', {
      method: 'POST',
      token: adminToken,
      body: {
        name: 'Docs Demo API',
        description: 'Synthetic component for README screenshots',
        status: 'operational',
      },
    })

    if (!created.ok) {
      throw new Error(`Failed to create demo component (${created.status}): ${JSON.stringify(created.data)}`)
    }

    component = created.data
  }

  const incidentsRes = await apiRequest('/incidents', { token: adminToken })
  if (!incidentsRes.ok || !Array.isArray(incidentsRes.data)) {
    throw new Error(`Failed to read incidents (${incidentsRes.status})`)
  }

  const demoIncidentExists = incidentsRes.data.some((item) => item.title === 'Docs Demo Incident')
  if (!demoIncidentExists) {
    const created = await apiRequest('/incidents', {
      method: 'POST',
      token: adminToken,
      body: {
        title: 'Docs Demo Incident',
        description: 'Synthetic incident to showcase history and incident UI in docs screenshots.',
        status: 'investigating',
        impact: 'minor',
        affectedComponents: [component.id],
      },
    })

    if (!created.ok) {
      throw new Error(`Failed to create demo incident (${created.status}): ${JSON.stringify(created.data)}`)
    }
  }

  const monitorsRes = await apiRequest('/monitors', { token: adminToken })
  if (!monitorsRes.ok || !Array.isArray(monitorsRes.data)) {
    throw new Error(`Failed to read monitors (${monitorsRes.status})`)
  }

  const demoMonitorExists = monitorsRes.data.some((item) => item.name === 'Docs Demo HTTPS Monitor')
  if (!demoMonitorExists) {
    const created = await apiRequest('/monitors', {
      method: 'POST',
      token: adminToken,
      body: {
        name: 'Docs Demo HTTPS Monitor',
        type: 'http',
        target: 'https://example.com',
        intervalSeconds: 60,
        timeoutSeconds: 10,
        componentId: component.id,
      },
    })

    if (!created.ok) {
      throw new Error(`Failed to create demo monitor (${created.status}): ${JSON.stringify(created.data)}`)
    }
  }
}

async function captureScreenshots(adminAuth) {
  await mkdir(OUTPUT_DIR, { recursive: true })

  const browser = await chromium.launch({ headless: true })
  try {
    const publicContext = await browser.newContext({ viewport: VIEWPORT })
    const publicPage = await publicContext.newPage()

    await publicPage.goto(BASE_WEB, { waitUntil: 'domcontentloaded' })
    await publicPage.waitForTimeout(800)
    await publicPage.screenshot({ path: path.join(OUTPUT_DIR, 'status-page.png'), fullPage: true })

    await publicPage.goto(`${BASE_WEB}/history`, { waitUntil: 'domcontentloaded' })
    await publicPage.waitForTimeout(800)
    await publicPage.screenshot({ path: path.join(OUTPUT_DIR, 'incident-history.png'), fullPage: true })

    await publicContext.close()

    const adminContext = await browser.newContext({ viewport: VIEWPORT })
    const adminPage = await adminContext.newPage()
    await adminPage.goto(BASE_WEB, { waitUntil: 'domcontentloaded' })
    await adminPage.evaluate((auth) => {
      localStorage.setItem('user_token', auth.token)
      localStorage.setItem('user_profile', JSON.stringify(auth.profile))
    }, adminAuth)

    await adminPage.goto(`${BASE_WEB}/admin`, { waitUntil: 'domcontentloaded' })
    await adminPage.waitForSelector('text=Dashboard', { timeout: 15000 })
    await adminPage.waitForTimeout(500)
    await adminPage.screenshot({ path: path.join(OUTPUT_DIR, 'admin-dashboard.png'), fullPage: true })

    await adminPage.goto(`${BASE_WEB}/admin/monitors`, { waitUntil: 'domcontentloaded' })
    await adminPage.waitForSelector('text=Monitors', { timeout: 15000 })
    await adminPage.waitForTimeout(500)
    await adminPage.screenshot({ path: path.join(OUTPUT_DIR, 'monitoring-table.png'), fullPage: true })

    await adminPage.goto(`${BASE_WEB}/admin/settings`, { waitUntil: 'domcontentloaded' })
    await adminPage.waitForSelector('text=Status Page Settings', { timeout: 15000 })
    await adminPage.waitForTimeout(500)
    await adminPage.screenshot({ path: path.join(OUTPUT_DIR, 'theme-settings.png'), fullPage: true })

    await adminContext.close()
  } finally {
    await browser.close()
  }
}

async function main() {
  await assertServerAvailability()
  const adminAuth = await loginAdmin()
  await ensureDemoData(adminAuth.token)
  await captureScreenshots(adminAuth)

  const generated = [
    'status-page.png',
    'incident-history.png',
    'admin-dashboard.png',
    'monitoring-table.png',
    'theme-settings.png',
  ].map((name) => path.join('docs/screenshots', name))

  console.log('Generated screenshots:')
  for (const item of generated) {
    console.log(`- ${item}`)
  }
}

main().catch((error) => {
  console.error(error)
  process.exit(1)
})
