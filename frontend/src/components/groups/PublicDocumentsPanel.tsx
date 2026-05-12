import { FileText, Download, Eye } from 'lucide-react'
import type { PublicDocument } from '../../types'

interface PublicDocumentsPanelProps {
  documents?: PublicDocument[]
  allowDownloads: boolean
  isActive?: boolean
  slug: string
}

const isPdf = (filename: string) => filename.toLowerCase().endsWith('.pdf')

export function PublicDocumentsPanel({ documents, allowDownloads, isActive = true, slug }: PublicDocumentsPanelProps) {
  return (
    <div className='col-span-4 border-r-2 border-neutral-950 bg-neutral-50 p-6 flex flex-col gap-4 overflow-y-auto'>
      <h4 className='font-bold text-lg text-neutral-950 mb-2'>Documentos Adjuntos</h4>

      {!isActive && (
        <div className='p-3 bg-amber-100 border-2 border-neutral-950 rounded shadow-hard-sm'>
          <p className='text-xs font-bold text-neutral-900'>Grupo desactivado.</p>
          <p className='text-xs text-neutral-600 mt-1'>Los documentos se muestran aquí cuando el grupo esté activo.</p>
        </div>
      )}

      {documents && documents.length > 0 ? (
        <div className='flex flex-col gap-3'>
          {documents.map(doc => (
            <div
              key={doc.id}
              className='bg-white border-2 border-neutral-950 p-4 rounded shadow-hard-sm flex items-center justify-between hover:translate-x-px hover:translate-y-px hover:shadow-none transition-all'
            >
              <div className='flex items-center gap-2 min-w-0'>
                <FileText size={16} className='text-neutral-600 shrink-0' />
                <span className='text-sm font-bold text-neutral-950 truncate'>{doc.filename}</span>
              </div>

              {allowDownloads && (
                <div className='flex items-center gap-2 shrink-0'>
                  {isPdf(doc.filename) ? (
                    <a
                      href={`/api/v1/public/groups/${slug}/documents/${doc.id}/view`}
                      target="_blank"
                      rel="noopener noreferrer"
                      className='p-1.5 border-2 border-transparent hover:border-neutral-950 rounded hover:bg-neutral-100 transition-all cursor-pointer'
                      title='Vista previa'
                    >
                      <Eye size={16} className='text-neutral-700' />
                    </a>
                  ) : (
                    <a
                      href={`/api/v1/public/groups/${slug}/documents/${doc.id}/download`}
                      download
                      className='p-1.5 border-2 border-transparent hover:border-neutral-950 rounded hover:bg-neutral-100 transition-all cursor-pointer'
                      title='Descargar documento'
                    >
                      <Download size={16} className='text-neutral-700' />
                    </a>
                  )}
                </div>
              )}
            </div>
          ))}
        </div>
      ) : (
        <p className='text-sm text-neutral-500 font-medium'>No hay documentos en este grupo.</p>
      )}
    </div>
  )
}
