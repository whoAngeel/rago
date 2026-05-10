import { useState } from 'react'
import { FileText, Plus, ChevronDown, ChevronUp } from 'lucide-react'
import type { Document } from '../../types'

interface DocumentSelectorProps {
  availableDocs: Document[]
  onAdd: (docId: number) => void
}

export function DocumentSelector({ availableDocs, onAdd }: DocumentSelectorProps) {
  const [isOpen, setIsOpen] = useState(false)

  return (
    <div className='flex flex-col gap-3'>
      <button
        onClick={() => setIsOpen(!isOpen)}
        className='flex items-center gap-1.5 text-sm font-bold text-neutral-600 hover:text-neutral-950 uppercase tracking-wider transition-colors cursor-pointer w-fit'
      >
        {isOpen ? <ChevronUp size={16} /> : <ChevronDown size={16} />}
        Agregar documentos
      </button>

      {isOpen && (
        <div className='bg-neutral-50 border-2 border-neutral-950 rounded p-4 flex flex-col gap-3 shadow-hard-sm'>
          {availableDocs.length === 0 ? (
            <p className='text-sm text-neutral-500 font-medium'>No hay documentos disponibles para agregar</p>
          ) : (
            <div className='grid grid-cols-1 sm:grid-cols-2 gap-3'>
              {availableDocs.map((doc) => (
                <div
                  key={doc.id}
                  onClick={() => onAdd(doc.id)}
                  className='bg-white border-2 border-neutral-950 rounded p-3 flex items-center justify-between shadow-hard-sm hover:translate-x-px 
                  hover:translate-y-px hover:shadow-none transition-all duration-200 cursor-pointer group'
                  title='Hacer click para agregar'
                >
                  <div className='flex items-center gap-2 min-w-0'>
                    <FileText size={16} className='text-neutral-600 shrink-0' />
                    <span className='text-sm font-bold text-neutral-950 truncate'>{doc.filename}</span>
                  </div>
                  <div className='p-1 border-2 border-transparent group-hover:border-neutral-950 rounded bg-white transition-all'>
                    <Plus size={16} className='text-neutral-600 group-hover:text-primary-500' />
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  )
}
