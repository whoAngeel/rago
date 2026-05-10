import { FileText } from 'lucide-react'
import type { Document } from '../../types'

interface GroupDocumentsProps {
  allDocuments?: Document[]
  selectedIds: number[]
  onToggle: (docId: number, isSelected: boolean) => void
}

export function GroupDocuments({ allDocuments, selectedIds, onToggle }: GroupDocumentsProps) {
  if (!allDocuments) return <p className='text-sm text-neutral-500 font-medium'>Cargando documentos...</p>

  return (
    <div className='flex flex-col gap-3'>
      <h4 className='text-lg font-bold text-neutral-950'>Documentos</h4>
      <p className='text-sm text-neutral-600 font-medium mb-1'>Haz click en un documento para agregarlo o quitarlo del grupo.</p>
      
      <div className='grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3'>
        {allDocuments.map((doc) => {
          const isSelected = selectedIds.includes(doc.id)
          return (
            <div
              key={doc.id}
              onClick={() => onToggle(doc.id, isSelected)}
              className={`border-2 border-neutral-950 rounded p-3 flex items-center gap-2 shadow-hard-sm transition-all duration-200 cursor-pointer ${
                isSelected 
                  ? 'bg-blue-200 text-neutral-950 border-neutral-950' 
                  : 'bg-neutral-50 text-neutral-400 border-neutral-400 hover:border-neutral-950 hover:text-neutral-700 hover:bg-neutral-100'
              } hover:translate-x-px hover:translate-y-px hover:shadow-none`}
              title={isSelected ? 'Quitar del grupo' : 'Agregar al grupo'}
            >
              <FileText size={16} className={`${isSelected ? 'text-blue-700' : 'text-neutral-400'} shrink-0`} />
              <span className='text-sm font-bold truncate'>{doc.filename}</span>
            </div>
          )
        })}
      </div>
    </div>
  )
}
