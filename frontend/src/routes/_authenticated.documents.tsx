import { createFileRoute } from '@tanstack/react-router'
import { DropZone } from '../components/documents/DropZone'

export const Route = createFileRoute('/_authenticated/documents')({
  component: RouteComponent,
})

function RouteComponent() {
  return (
    <div className='flex w-full h-full p-6 flex-col gap-6'>

      {/* Header */}
      <div>
        <h1 className='text-4xl font-black text-neutral-950 tracking-tighter'>Documentos</h1>
        <p className='text-neutral-500'>Gestión de documentos</p>
      </div>

      {/* Drag & Drop Area (Debajo del header) */}
      <div className='w-full'>
        <DropZone onUpload={(file) => { console.log(file.name) }} />
      </div>


    </div>
  )
}
