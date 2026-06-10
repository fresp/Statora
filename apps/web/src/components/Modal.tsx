import React, { useEffect } from 'react'
import { X } from 'lucide-react'

type ModalSize = 'md' | 'lg'

interface ModalProps {
  title: string
  onClose: () => void
  children: React.ReactNode
  footer?: React.ReactNode
  size?: ModalSize
}

const SIZE_CLASS: Record<ModalSize, string> = {
  md: 'max-w-md',
  lg: 'max-w-lg',
}

export default function Modal({ title, onClose, children, footer, size = 'md' }: ModalProps) {
  useEffect(() => {
    const previousOverflow = document.body.style.overflow
    document.body.style.overflow = 'hidden'

    return () => {
      document.body.style.overflow = previousOverflow
    }
  }, [])

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/55 p-4 backdrop-blur-sm">
      <div className={`statusforge-dialog ${SIZE_CLASS[size]} flex max-h-[90vh] flex-col`}>
        <div className="statusforge-dialog-header">
          <h2 className="font-semibold tracking-tight text-slate-950">{title}</h2>
          <button type="button" onClick={onClose} className="rounded-full p-1 text-slate-400 transition-colors hover:bg-white hover:text-slate-700 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500 focus-visible:ring-offset-2" aria-label="Close modal">
            <X className="w-5 h-5" />
          </button>
        </div>

        <div className="p-6 flex-1 min-h-0 overflow-y-auto">{children}</div>

        {footer && <div className="statusforge-dialog-footer">{footer}</div>}
      </div>
    </div>
  )
}
