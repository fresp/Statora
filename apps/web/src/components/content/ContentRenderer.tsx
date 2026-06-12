import React from 'react'
import type { RichTextDocument } from '../../types'

interface ContentRendererProps {
  text: string
  json?: RichTextDocument
  className?: string
}

interface TipTapNode {
  type: string
  attrs?: Record<string, unknown>
  content?: TipTapNode[]
  text?: string
  marks?: Array<{ type: string; attrs?: Record<string, unknown> }>
}

function renderMarks(node: TipTapNode, children: React.ReactNode): React.ReactNode {
  if (!node.marks || node.marks.length === 0) return children

  return node.marks.reduce<React.ReactNode>((acc, mark) => {
    switch (mark.type) {
      case 'bold':
        return <strong key="bold">{acc}</strong>
      case 'italic':
        return <em key="italic">{acc}</em>
      case 'strike':
        return <s key="strike">{acc}</s>
      case 'code':
        return <code key="code" className="rounded bg-slate-100 px-1 py-0.5 font-mono text-xs text-slate-700">{acc}</code>
      case 'link': {
        const href = (mark.attrs?.href as string) || '#'
        return (
          <a key="link" href={href} target="_blank" rel="noopener noreferrer" className="text-emerald-700 underline">
            {acc}
          </a>
        )
      }
      case 'highlight':
        return <mark key="highlight" className="rounded-sm bg-yellow-200 px-0.5">{acc}</mark>
      default:
        return acc
    }
  }, children)
}

function renderNode(node: TipTapNode, index: number): React.ReactNode {
  switch (node.type) {
    case 'doc':
      return (
        <React.Fragment key={index}>
          {(node.content || []).map((child, i) => renderNode(child, i))}
        </React.Fragment>
      )
    case 'paragraph': {
      const align = (node.attrs?.textAlign as string | undefined) || undefined
      return (
        <p key={index} className="mb-2 last:mb-0" style={align ? { textAlign: align as React.CSSProperties['textAlign'] } : undefined}>
          {(node.content || []).map((child, i) => renderNode(child, i))}
        </p>
      )
    }
    case 'heading': {
      const level = (node.attrs?.level as number) || 2
      const align = (node.attrs?.textAlign as string | undefined) || undefined
      const Tag = `h${Math.min(Math.max(level, 1), 6)}` as keyof JSX.IntrinsicElements
      return (
        <Tag key={index} className="mt-3 mb-2 font-semibold text-slate-900 first:mt-0" style={align ? { textAlign: align as React.CSSProperties['textAlign'] } : undefined}>
          {(node.content || []).map((child, i) => renderNode(child, i))}
        </Tag>
      )
    }
    case 'bulletList':
      return (
        <ul key={index} className="mb-2 list-disc pl-5 last:mb-0">
          {(node.content || []).map((child, i) => renderNode(child, i))}
        </ul>
      )
    case 'orderedList':
      return (
        <ol key={index} className="mb-2 list-decimal pl-5 last:mb-0">
          {(node.content || []).map((child, i) => renderNode(child, i))}
        </ol>
      )
    case 'listItem':
      return (
        <li key={index} className="mb-1 last:mb-0">
          {(node.content || []).map((child, i) => renderNode(child, i))}
        </li>
      )
    case 'blockquote':
      return (
        <blockquote key={index} className="mb-2 border-l-4 border-slate-200 pl-4 italic text-slate-600 last:mb-0">
          {(node.content || []).map((child, i) => renderNode(child, i))}
        </blockquote>
      )
    case 'codeBlock':
      return (
        <pre key={index} className="mb-2 overflow-x-auto rounded-lg bg-slate-900 p-3 text-xs text-slate-100 last:mb-0">
          <code>{(node.content || []).map((child, i) => renderNode(child, i))}</code>
        </pre>
      )
    case 'hardBreak':
      return <br key={index} />
    case 'horizontalRule':
      return <hr key={index} className="my-3 border-slate-200" />
    case 'text':
      return renderMarks(node, node.text || '')
    default:
      return null
  }
}

export default function ContentRenderer({ text, json, className = '' }: ContentRendererProps) {
  const hasJson = json && Array.isArray(json.content) && json.content.length > 0

  if (!hasJson) {
    return <p className={`text-sm leading-relaxed text-slate-700 dark:text-slate-300 ${className}`}>{text || ''}</p>
  }

  return (
    <div className={`rich-text-content text-sm leading-relaxed text-slate-700 dark:text-slate-300 ${className}`}>
      {renderNode(json as unknown as TipTapNode, 0)}
    </div>
  )
}
