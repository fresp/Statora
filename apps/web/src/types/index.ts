export interface PaginatedResponse<T> {
  items: T[];
  total: number;
  page: number;
  total_pages: number;
}

// API types for the status platform

export type ComponentStatus =
  | 'operational'
  | 'degraded_performance'
  | 'partial_outage'
  | 'major_outage'
  | 'maintenance'

export interface RichTextDocument {
  type: 'doc'
  content?: Array<Record<string, unknown>>
}

export type IncidentStatus = 'investigating' | 'identified' | 'monitoring' | 'resolved'
export type IncidentImpact = 'none' | 'minor' | 'major' | 'critical'
export type IncidentVisibilityState = 'draft' | 'published'
export type MaintenanceStatus = 'draft' | 'scheduled' | 'active' | 'completed' | 'in_progress'
export type MonitorType = 'http' | 'tcp' | 'dns' | 'ping' | 'ssl'
export type MonitorLogStatus = 'up' | 'down'

export interface Component {
  id: string
  name: string
  description: string
  status: ComponentStatus
  createdAt: string
  updatedAt: string
}

export interface SubComponent {
  id: string
  componentId: string
  name: string
  description: string
  status: ComponentStatus
}

export interface UptimeBar {
  date: string
  uptimePercent: number
  status: ComponentStatus
}

export interface ComponentWithSubs extends Component {
  subComponents: SubComponent[]
  uptimeHistory: UptimeBar[]
}

export interface CategoryServiceStatus {
  id: string
  name: string
  description: string
  status: ComponentStatus
  updatedAt: string
  uptime30d: number | null
  uptimeHistory: UptimeBar[]
}

export interface CategorySummary {
  prefix: string
  name: string
  description: string
  aggregateStatus: ComponentStatus
  uptime30d: number | null
  services: CategoryServiceStatus[]
  incidents: Incident[]
}

export interface Incident {
  id: string
  title: string
  description: string
  descriptionJson?: RichTextDocument
  visibilityState?: IncidentVisibilityState
  postmortemEnabled?: boolean
  postmortemContentJson?: RichTextDocument
  status: IncidentStatus
  impact: IncidentImpact
  creatorId?: string
  creatorUsername?: string
  affectedComponents: Component[]
  affectedComponentTargets?: IncidentAffectedComponentTarget[]
  createdAt: string
  updatedAt: string
  resolvedAt?: string
  updates?: IncidentUpdate[]
}

export interface IncidentAffectedComponentTarget {
  component: Component
  subComponents?: SubComponent[]
}

export interface IncidentUpdate {
  id: string
  incidentId: string
  message: string
  messageJson?: RichTextDocument
  status: IncidentStatus
  createdAt: string
}

export interface Maintenance {
  id: string
  title: string
  description: string
  descriptionJson?: RichTextDocument
  creatorId?: string
  creatorUsername?: string
  components: string[]
  startTime: string
  endTime: string
  status: MaintenanceStatus
  visibilityState?: IncidentVisibilityState
}

export type AuditResourceType = 'incident' | 'incident_update' | 'maintenance'
export type AuditAction =
  | 'created'
  | 'edited'
  | 'status_changed'
  | 'update_added'
  | 'resolved'
  | 'deleted'
  | 'published'
  | 'draft_saved'

export interface AuditLog {
  id: string
  resourceType: AuditResourceType
  resourceId: string
  action: AuditAction
  actorId?: string
  actorUsername?: string
  at: string
  summary?: string
  changes?: Record<string, unknown>
}

export interface Monitor {
  id: string
  name: string
  type: MonitorType
  target: string
  monitoring?: {
    advanced?: {
      domain_expiry?: boolean
      cert_expiry?: boolean
      ignore_tls_error?: boolean
    }
  }
  sslThresholds?: number[]
  intervalSeconds: number
  timeoutSeconds: number
  componentId: string
  subComponentId?: string
  lastStatus?: MonitorLogStatus
  sslWarning?: boolean
  sslDaysRemaining?: number
  sslTriggeredThreshold?: number
  domainWarning?: boolean
  domainDaysRemaining?: number
  domainTriggeredThreshold?: number
  lastCheckedAt?: string
  createdAt: string
}

export interface MonitorLog {
  id: string
  monitorId: string
  status: MonitorLogStatus
  responseTime: number
  statusCode: number
  checkedAt: string
  region?: string
}

export interface DailyUptime {
  id: string
  monitorId: string
  date: string
  totalChecks: number
  successfulChecks: number
  uptimePercent: number
}

export interface MonitorMetrics {
  latency?: {
    p90?: number | null
    p99?: number | null
  } | null
  availability?: {
    last30Days?: number | null
  } | null
  history: Array<{
    month: string
    latency?: {
      p90?: number | null
      p99?: number | null
    } | null
    availability?: number | null
  }>
}

export interface Subscriber {
  id: string
  email: string
  verified: boolean
  createdAt: string
}

export interface StatusSummary {
  overallStatus: ComponentStatus
  componentCounts: Record<ComponentStatus, number>
  activeIncidents: number
  scheduledMaintenance: number
}

export interface User {
  id: string
  username: string
  email: string
  role: UserRole
  mfaEnabled?: boolean
  mfaVerified?: boolean
}

export type UserRole = 'admin' | 'operator'

export type UserStatus = 'active' | 'disabled' | 'invited'

export interface UserMember {
  id: string
  username: string
  email: string
  role: UserRole
  status: UserStatus
  ssoEnabled: boolean
}

export interface UserInvitation {
  id: string
  email: string
  role: UserRole
  expiresAt: string
  createdAt: string
  isExpired: boolean
}

export interface LoginResponse {
  token: string
  user: User
  mfaRequired?: boolean
}

export interface AuthMeResponse {
  userId: string
  username: string
  email: string
  role: UserRole
  mfaEnabled: boolean
  mfaVerified: boolean
}

export interface StoredUserProfile {
  id: string
  username: string
  email: string
  role?: UserRole
  mfaEnabled?: boolean
  mfaVerified?: boolean
}

export interface MfaSetupResponse {
  secret: string
  otpauthUrl: string
  recoveryCodes: string[]
}

export interface MfaVerifyRequest {
  code: string
}

export interface MfaVerifyResponse {
  token: string
  mfaVerified: boolean
  user: User
}

export interface ProfileUpdateRequest {
  username: string
  currentPassword?: string
  newPassword?: string
}

export interface StatusPageSSOSettings {
  enabled: boolean
  provider: string
  issuer: string
  audience: string
  algorithm: 'HS256' | 'RS256'
  publicKeyPem: string
  sharedSecret?: string
  hasSecret?: boolean
}

export interface StatusPageSettings {
  head: {
    title: string
    description: string
    keywords: string
    faviconUrl: string
    metaTags: Record<string, string>
  }
  branding: {
    siteName: string
    logoUrl: string
    backgroundImageUrl: string
    heroImageUrl: string
  }
  theme: {
    preset: string
    appliedPreset?: string
    mode?: 'light' | 'dark' | 'system'
    overrides?: {
      light?: {
        primaryColor?: string
        backgroundColor?: string
        textColor?: string
        accentColor?: string
      }
      dark?: {
        primaryColor?: string
        backgroundColor?: string
        textColor?: string
        accentColor?: string
      }
      typography?: {
        fontFamily?: string
        fontScale?: 'sm' | 'md' | 'lg'
      }
      ui?: {
        cardBackground?: string
        borderColor?: string
      }
    }
    resolved?: {
      light: {
        primaryColor: string
        backgroundColor: string
        textColor: string
        accentColor: string
      }
      dark: {
        primaryColor: string
        backgroundColor: string
        textColor: string
        accentColor: string
      }
      typography: {
        fontFamily: string
        fontScale: 'sm' | 'md' | 'lg'
      }
      ui: {
        cardBackground: string
        borderColor: string
      }
    }
  }
  layout?: {
    variant: 'classic' | 'compact' | 'minimal' | 'cards'
  }
  footer: {
    text: string
    showPoweredBy: boolean
  }
  sso?: StatusPageSSOSettings
  customCss: string
  updatedAt: string
  createdAt: string
}

export interface StatusPageThemePresetSummary {
  key: string
  label: string
}

export interface StatusPageThemePresetCatalog {
  presets: StatusPageThemePresetSummary[]
}

export interface StatusPageSettingsPatchRequest {
  head?: {
    title?: string
    description?: string
    keywords?: string
    faviconUrl?: string
    metaTags?: Record<string, string>
  }
  branding?: {
    siteName?: string
    logoUrl?: string
    backgroundImageUrl?: string
    heroImageUrl?: string
  }
  theme?: {
    preset?: string
    mode?: 'light' | 'dark' | 'system'
    light?: {
      primaryColor?: string
      backgroundColor?: string
      textColor?: string
      accentColor?: string
    }
    dark?: {
      primaryColor?: string
      backgroundColor?: string
      textColor?: string
      accentColor?: string
    }
    typography?: {
      fontFamily?: string
      fontScale?: 'sm' | 'md' | 'lg'
    }
    ui?: {
      cardBackground?: string
      borderColor?: string
    }
  }
  layout?: {
    variant?: 'classic' | 'compact' | 'minimal' | 'cards'
  }
  footer?: {
    text?: string
    showPoweredBy?: boolean
  }
  sso?: {
    enabled?: boolean
    provider?: string
    issuer?: string
    audience?: string
    algorithm?: 'HS256' | 'RS256'
    sharedSecret?: string
    publicKeyPem?: string
  }
  customCss?: string
}

export interface WebhookChannel {
  id: string
  name: string
  url: string
  enabled: boolean
  createdAt: string
}
