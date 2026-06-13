import type { Incident, IncidentUpdate, Maintenance, MaintenanceStatus, RichTextDocument } from '../types'

export function normalizeMaintenanceStatus(status: MaintenanceStatus): Exclude<MaintenanceStatus, 'in_progress'> {
  return status === 'in_progress' ? 'active' : status
}

export function getPlainTextFromRichText(document?: RichTextDocument): string {
  if (!document?.content?.length) {
    return ''
  }

  return extractText(document.content).join('\n').trim()
}

function extractText(nodes: Array<Record<string, unknown>>): string[] {
  return nodes.flatMap((node) => {
    const directText = typeof node.text === 'string' ? [node.text] : []
    const nestedContent = Array.isArray(node.content)
      ? extractText(node.content as Array<Record<string, unknown>>)
      : []

    return [...directText, ...nestedContent]
  })
}

export function getIncidentContent(incident: Pick<Incident, 'description' | 'descriptionJson'>): {
  text: string
  json?: RichTextDocument
} {
  const derivedText = getPlainTextFromRichText(incident.descriptionJson)

  return {
    text: derivedText || incident.description,
    json: incident.descriptionJson,
  }
}

export function getIncidentUpdateContent(update: Pick<IncidentUpdate, 'message' | 'messageJson'>): {
  text: string
  json?: RichTextDocument
} {
  const derivedText = getPlainTextFromRichText(update.messageJson)

  return {
    text: derivedText || update.message,
    json: update.messageJson,
  }
}

export function getMaintenanceContent(maintenance: Pick<Maintenance, 'description' | 'descriptionJson'>): {
  text: string
  json?: RichTextDocument
} {
  const derivedText = getPlainTextFromRichText(maintenance.descriptionJson)

  return {
    text: derivedText || maintenance.description,
    json: maintenance.descriptionJson,
  }
}
