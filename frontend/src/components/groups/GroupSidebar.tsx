import { useState } from 'react'
import { Switch } from '../ui'
import { Copy, Check, ExternalLink } from 'lucide-react'
import { GroupQRCode } from './GroupQRCode'

interface GroupSidebarProps {
  slug?: string
  allowDownloads: boolean
  isActive: boolean
  onUpdate: (updates: { allow_downloads?: boolean; is_active?: boolean }) => void
}

export function GroupSidebar({ slug, allowDownloads, isActive, onUpdate }: GroupSidebarProps) {
  const shareableLink = slug ? `${window.location.origin}/c/${slug}` : '—'

  const [copied, setCopied] = useState(false)

  const copyToClipboard = () => {
    if (slug) {
      navigator.clipboard.writeText(shareableLink)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    }
  }

  return (
    <div className='flex flex-col gap-4'>
      {/* Link Card */}
      <div className='flex flex-col gap-3 border-2 border-neutral-950 p-6 bg-neutral-50 rounded shadow-hard-md'>
        <h4 className='text-lg font-bold text-neutral-950'>Link Del Chat</h4>
        <p className='text-xs font-bold uppercase tracking-wider text-neutral-600'>URL DE ACCESO</p>

        <div className='p-2 border-2 border-neutral-950 rounded bg-neutral-100 flex items-center justify-between gap-2'>
          <p className='text-sm font-medium truncate'>{shareableLink}</p>
          <button onClick={copyToClipboard} className='p-1.5 border-2 border-transparent hover:border-neutral-950 rounded hover:bg-white transition-all cursor-pointer' title='Copiar link'>
            {copied ? <Check size={16} className="text-primary-600" /> : <Copy size={16} />}
          </button>
        </div>

        <a 
          href={shareableLink} 
          target="_blank" 
          rel="noopener noreferrer"
          className="w-full border-2 border-neutral-950 p-2 bg-white rounded shadow-hard-sm font-bold text-sm text-center hover:translate-x-px hover:translate-y-px hover:shadow-none transition-all cursor-pointer flex items-center justify-center gap-2"
        >
          <ExternalLink size={16} />
          Ver Chat Público
        </a>

        <div className='border-t-2 border-neutral-950 my-2' />

        <div className='flex flex-col gap-4'>
          <div className='flex items-center justify-between'>
            <span className='text-sm font-bold text-neutral-700'>Permitir Descargas</span>
            <Switch
              checked={allowDownloads}
              onChange={(checked) => onUpdate({ allow_downloads: checked })}
            />
          </div>
          <div className='flex items-center justify-between'>
            <span className='text-sm font-bold text-neutral-700'>Activar Chat</span>
            <Switch
              checked={isActive}
              onChange={(checked) => onUpdate({ is_active: checked })}
            />
          </div>
        </div>
      </div>

      <GroupQRCode slug={slug} />
    </div>
  )
}
