import { createPortal } from 'react-dom'
import { AlertTriangle, X } from 'lucide-react'

interface ConfirmDialogProps {
  open: boolean
  title: string
  message: string
  confirmLabel?: string
  cancelLabel?: string
  variant?: 'danger' | 'warning' | 'default'
  onConfirm: () => void
  onCancel: () => void
}

export function ConfirmDialog({
  open,
  title,
  message,
  confirmLabel = 'Confirmar',
  cancelLabel = 'Cancelar',
  variant = 'default',
  onConfirm,
  onCancel,
}: ConfirmDialogProps) {
  if (!open) return null

  const variantStyles = {
    danger: 'bg-accent-red-deep text-white hover:bg-accent-red-deep/90',
    warning: 'bg-accent-amber-deep text-black hover:bg-accent-amber-deep/90',
    default: 'bg-primary-400 text-black hover:bg-primary-300',
  }

  return createPortal(
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <div
        className="absolute inset-0 bg-neutral-950/40"
        onClick={onCancel}
      />
      <div className="relative bg-white border-2 border-neutral-950 rounded-card shadow-hard-lg max-w-sm w-full mx-4 overflow-hidden">
        <div className="p-5 flex flex-col gap-4">
          <div className="flex items-start gap-3">
            <div className="w-10 h-10 bg-accent-amber border-2 border-neutral-950 rounded-full flex items-center justify-center shrink-0 shadow-hard-sm">
              <AlertTriangle size={20} className="text-neutral-950" />
            </div>
            <div className="flex-1 min-w-0">
              <h3 className="font-black text-neutral-950 text-lg leading-tight">
                {title}
              </h3>
              <p className="text-sm font-medium text-neutral-600 mt-1">
                {message}
              </p>
            </div>
            <button
              onClick={onCancel}
              className="p-1 text-neutral-400 hover:text-neutral-950 transition-colors cursor-pointer shrink-0"
            >
              <X size={18} />
            </button>
          </div>

          <div className="flex items-center gap-2 justify-end">
            <button
              onClick={onCancel}
              className="px-4 py-2 border-2 border-neutral-950 rounded bg-white text-neutral-950 font-bold text-sm hover:bg-neutral-100 hover:shadow-hard-sm transition-all cursor-pointer"
            >
              {cancelLabel}
            </button>
            <button
              onClick={onConfirm}
              className={`px-4 py-2 border-2 border-neutral-950 rounded font-bold text-sm hover:shadow-hard-sm transition-all cursor-pointer ${variantStyles[variant]}`}
            >
              {confirmLabel}
            </button>
          </div>
        </div>
      </div>
    </div>,
    document.body,
  )
}
