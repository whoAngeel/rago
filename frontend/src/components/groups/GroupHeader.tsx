import { useState } from 'react'
import { Pencil, Check, X } from 'lucide-react'
import { formatDistanceToNow } from 'date-fns'
import { es } from 'date-fns/locale'

interface GroupHeaderProps {
  name: string
  createdAt?: string
  updatedAt?: string
  documentCount: number
  onSaveName: (newName: string) => void
}

export function GroupHeader({ name, createdAt, updatedAt, documentCount, onSaveName }: GroupHeaderProps) {
  const [editing, setEditing] = useState(false)
  const [draft, setDraft] = useState(name)

  const handleSave = () => {
    if (draft.trim() && draft.trim() !== name) {
      onSaveName(draft.trim())
    }
    setEditing(false)
  }

  const formatTimeAgo = (dateString?: string) => {
    if (!dateString) return ''
    try {
      const date = new Date(dateString)
      return formatDistanceToNow(date, { addSuffix: true, locale: es })
    } catch (e) {
      return ''
    }
  }

  return (
    <div className='flex flex-col gap-1'>
      <div className='flex items-center gap-3'>
        {editing ? (
          <div className='flex items-center gap-2'>
            <input
              value={draft}
              onChange={(e) => setDraft(e.target.value)}
              className='border-2 border-neutral-950 px-2 py-1 text-2xl font-black rounded focus:outline-none focus:border-primary-500'
              autoFocus
            />
            <button onClick={handleSave} className='p-1.5 border-2 border-neutral-950 rounded bg-white hover:bg-neutral-100 cursor-pointer shadow-hard-sm' title='Guardar'>
              <Check size={20} />
            </button>
            <button onClick={() => { setEditing(false); setDraft(name); }} className='p-1.5 border-2 border-neutral-950 rounded bg-white hover:bg-neutral-100 cursor-pointer shadow-hard-sm' title='Cancelar'>
              <X size={20} />
            </button>
          </div>
        ) : (
          <div className='flex items-center gap-2'>
            <h1 className='text-3xl font-black text-neutral-950'>{name}</h1>
            <button onClick={() => setEditing(true)} className='p-1.5 border-2 border-transparent hover:border-neutral-950 rounded hover:bg-neutral-100 cursor-pointer transition-all' title='Editar nombre'>
              <Pencil size={18} />
            </button>
          </div>
        )}
      </div>
      <div className='flex flex-col flex-wrap items-start gap-x-1 gap-y-1 text-xs font-bold text-neutral-500 uppercase tracking-wider mt-1'>
        <span>{createdAt && <span>Creado el {createdAt.slice(0, 10)}</span>}</span>

        <span>{updatedAt && <span>Actualizado {formatTimeAgo(updatedAt)}</span>}</span>

        <span>{documentCount} documentos</span>
      </div>
    </div>
  )
}
