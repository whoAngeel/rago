import { Menu } from 'lucide-react'

interface PublicChatNavbarProps {
  groupName?: string
  onToggleDocs?: () => void
}

export function PublicChatNavbar({ groupName, onToggleDocs }: PublicChatNavbarProps) {
  return (
    <div className='border-b-2 border-neutral-950 bg-white p-4 flex items-center justify-between shadow-hard-sm z-20'>
      <div className='flex items-center gap-3'>
        {onToggleDocs && (
          <button
            onClick={onToggleDocs}
            className='lg:hidden p-2 border-2 border-neutral-950 rounded hover:bg-neutral-100 transition-all cursor-pointer'
            title='Documentos'
          >
            <Menu size={18} />
          </button>
        )}
        <h3 className='text-xl font-black text-neutral-950 truncate'>
          {groupName}
        </h3>
      </div>
      <div className='flex items-center gap-2 shrink-0'>
        <span className='text-xs font-bold uppercase tracking-wider text-neutral-600 hidden sm:inline'>Chat Público</span>
        <div className='w-2 h-2 bg-green-500 rounded-full' />
      </div>
    </div>
  )
}
