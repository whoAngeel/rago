import { Button, Switch } from '../ui'
import { QrCode, Copy } from 'lucide-react'

interface GroupSidebarProps {
  slug?: string
  allowDownloads: boolean
  isActive: boolean
  onUpdate: (updates: { allow_downloads?: boolean; is_active?: boolean }) => void
}

export function GroupSidebar({ slug, allowDownloads, isActive, onUpdate }: GroupSidebarProps) {
  const shareableLink = slug ? `${window.location.origin}/chat/${slug}` : '—'

  const copyToClipboard = () => {
    if (slug) {
      navigator.clipboard.writeText(shareableLink)
      // Podríamos agregar un toast aquí si pasamos la función por props
    }
  }

  return (
    <div className='flex flex-col gap-4'>
      {/* Link Card */}
      <div className='flex flex-col gap-3 border-2 border-neutral-950 p-6 bg-neutral-50 rounded shadow-hard-md'>
        <h4 className='text-lg font-bold text-neutral-950'>Link Compartible</h4>
        <p className='text-xs font-bold uppercase tracking-wider text-neutral-600'>URL DE ACCESO</p>

        <div className='p-2 border-2 border-neutral-950 rounded bg-neutral-100 flex items-center justify-between gap-2'>
          <p className='text-sm font-medium truncate'>{shareableLink}</p>
          <button onClick={copyToClipboard} className='p-1.5 border-2 border-transparent hover:border-neutral-950 rounded hover:bg-white transition-all cursor-pointer' title='Copiar link'>
            <Copy size={16} />
          </button>
        </div>

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

      {/* QR Card */}
      <div className='flex flex-col gap-3 border-2 border-neutral-950 p-6 bg-neutral-50 rounded shadow-hard-md'>
        <h4 className='text-lg font-bold text-neutral-950'>Código QR</h4>
        <Button variant='secondary' className='w-full flex items-center justify-center gap-2'>
          <QrCode size={18} />
          <span>Compartir QR</span>
        </Button>
      </div>
    </div>
  )
}
