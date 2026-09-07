import { describe, expect, it } from 'vitest'
import { mergeAttributes } from '@tiptap/core'
import LinkifyIt from 'linkify-it'

describe('TipTap & Linkify security smoke tests', () => {
  it('prevents prototype pollution via mergeAttributes', () => {
    const payload: Record<string, unknown> = JSON.parse('{"__proto__": {"polluted": true}}')
    const result = mergeAttributes({}, payload)

    expect((Object.prototype as Record<string, unknown>).polluted).toBeUndefined()
    expect((result as Record<string, unknown>).polluted).toBeUndefined()
  })

  it('safely scans mailto links with linkify-it 5.0.2', () => {
    const linkify = new LinkifyIt()
    const matches = linkify.match('Contact mailto:support@example.com for help')

    expect(matches).not.toBeNull()
    expect(matches![0].url).toBe('mailto:support@example.com')
  })
})
