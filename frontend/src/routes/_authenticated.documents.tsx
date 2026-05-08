import { createFileRoute } from '@tanstack/react-router'
import { DropZone } from '../components/documents/DropZone'
import api from '../lib/api'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { type Document as DocumentType } from '../types'
import { DocumentTable } from '../components/documents/DocumentTable'
import { useEffect, useState } from 'react'
import { useToastStore } from '../store/toastStore'
import { useSSE } from '../hooks/useSSE'

export const Route = createFileRoute('/_authenticated/documents')({
  component: RouteComponent,
})

function RouteComponent() {

  useSSE();
  const addToast = useToastStore((s) => s.add)
  const queryClient = useQueryClient()

  const { data, isLoading, error } = useQuery<DocumentType[]>({
    queryKey: ["documents"],
    queryFn: async () => {
      const { data } = await api.get('/documents')
      console.log('Fetch documents:', data)
      return data.items
    },
    refetchInterval: (query) => {
      const docs = query.state.data
      if (!docs) return false
      const hasActive = docs.some(d => d.status === 'uploading' || d.status === 'pending' || d.status === 'processing')
      return hasActive ? 3000 : false
    },
  })

  useEffect(() => {
    if (error) {
      addToast("Error", `No se pudieron cargar los documentos: ${(error as Error).message}`, "error")
    }
  }, [error, addToast])

  const [deletingIds, setDeletingIds] = useState<number[]>([])

  useEffect(() => {
    if (data) {
      setDeletingIds(prev => prev.filter(id => data.some(doc => doc.id === id)))
    }
  }, [data])

  const uploadMutation = useMutation({
    mutationFn: async (file: File) => {
      const form = new FormData()
      form.append("file", file)
      return api.post("/documents/", form)
    },
    onMutate: (file: File) => {
      addToast("Subiendo", `Subiendo ${file.name}...`, "info")
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["documents"] })
      addToast("Éxito", "Documento subido correctamente", "success")
    },
    onError: (err: any) => {
      addToast("Error", `No se pudo subir el documento: ${err.message}`, "error")
    }
  })

  const deleteMutation = useMutation({
    mutationFn: async (id: number) => {
      return api.delete(`/documents/${id}/`)
    },
    onMutate: (id: number) => {
      setDeletingIds(prev => [...prev, id])
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["documents"] })
      addToast("Éxito", "Documento borrado correctamente", "success")
    },
    onError: (err: any, id: number) => {
      setDeletingIds(prev => prev.filter(itemId => itemId !== id))
      addToast("Error", `No se pudo borrar el documento: ${err.message}`, "error")
    }
  })

  return (
    <div className='flex w-full h-full p-6 flex-col gap-6'>

      {/* Header */}
      <div>
        <h1 className='text-4xl font-black text-neutral-950 tracking-tighter'>Documentos</h1>
        <p className='text-neutral-500'>Gestión de documentos</p>
      </div>

      {/* Drag & Drop */}
      <div className='w-full'>
        <DropZone onUpload={(file) => uploadMutation.mutate(file)} />
      </div>



      {/* Contenido Principal: Tabla o Estado de Carga Inicial */}
      {isLoading ? (
        <div className="flex-1 bg-white border-2 border-neutral-950 rounded-card p-12 flex flex-col items-center justify-center gap-4 shadow-hard-lg">
          <div className="w-12 h-12 bg-primary-400 border-2 border-neutral-950 rounded-full flex items-center justify-center animate-spin shadow-hard-sm">
            <span className="text-xl font-black">@</span>
          </div>
          <span className="text-neutral-950 font-bold">Cargando tus documentos...</span>
        </div>
      ) : (
        <DocumentTable documents={data || []}
          onDelete={(id) => deleteMutation.mutate(id)}
          onDownload={(id) => console.log(`download ${id}`)}
          deletingIds={deletingIds} />
      )}

    </div>
  )
}
