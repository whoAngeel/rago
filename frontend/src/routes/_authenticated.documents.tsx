import { createFileRoute } from '@tanstack/react-router'
import { DropZone } from '../components/documents/DropZone'
import api from '../lib/api'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { type Document as DocumentType, type User } from '../types'
import { DocumentTable } from '../components/documents/DocumentTable'
import { useEffect, useState } from 'react'
import { useToastStore } from '../store/toastStore'
import { useSSE } from '../hooks/useSSE'
import { Search, SlidersHorizontal, ChevronLeft, ChevronRight } from "lucide-react"

interface DocumentsResponse {
  items: DocumentType[]
  total: number
  page: number
  limit: number
}

export const Route = createFileRoute('/_authenticated/documents')({
  component: RouteComponent,
})

function RouteComponent() {

  useSSE();
  const addToast = useToastStore((s) => s.add)
  const queryClient = useQueryClient()

  const [page, setPage] = useState(1)
  const limit = 8

  const { data: currentUser } = useQuery<User>({
    queryKey: ["users", "me"],
    queryFn: async () => {
      const { data } = await api.get("/users/me")
      return data
    },
    staleTime: 10000,
  })

  const isAtLimit = (currentUser?.max_documents ?? 0) > 0 &&
    (currentUser?.document_count ?? 0) >= (currentUser?.max_documents ?? 0)

  const { data, isLoading, error } = useQuery<DocumentsResponse>({
    queryKey: ["documents", page],
    queryFn: async () => {
      const { data } = await api.get(`/documents/?page=${page}&limit=${limit}`)
      return data
    },
  })

  const totalPages = data ? Math.ceil(data.total / data.limit) : 0

  useEffect(() => {
    if (error) {
      addToast("Error", `No se pudieron cargar los documentos: ${(error as Error).message}`, "error")
    }
  }, [error, addToast])

  const [deletingIds, setDeletingIds] = useState<number[]>([])
  const [reprocessingIds, setReprocessingIds] = useState<number[]>([])

  useEffect(() => {
    if (data) {
      setDeletingIds(prev => prev.filter(id => data.items.some(doc => doc.id === id)))
      setReprocessingIds(prev => prev.filter(id => data.items.some(doc => doc.id === id)))
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
      queryClient.invalidateQueries({ queryKey: ["users", "me"] })
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
      queryClient.invalidateQueries({ queryKey: ["users", "me"] })
      addToast("Éxito", "Documento borrado correctamente", "success")
    },
    onError: (err: any, id: number) => {
      setDeletingIds(prev => prev.filter(itemId => itemId !== id))
      addToast("Error", `No se pudo borrar el documento: ${err.message}`, "error")
    }
  })

  const reprocessMutation = useMutation({
    mutationFn: async (id: number) => {
      return api.post(`/documents/${id}/reprocess`)
    },
    onMutate: (id: number) => {
      setReprocessingIds(prev => [...prev, id])
    },
    onSuccess: (_, id: number) => {
      setReprocessingIds(prev => prev.filter(itemId => itemId !== id))
      queryClient.invalidateQueries({ queryKey: ["documents"] })
      addToast("Éxito", "Documento puesto en cola para reprocesar", "success")
    },
    onError: (err: any, id: number) => {
      setReprocessingIds(prev => prev.filter(itemId => itemId !== id))
      addToast("Error", `No se pudo reprocesar el documento: ${err.message}`, "error")
    }
  })

  const documents = data?.items ?? []

  return (
    <div className='flex w-full h-full p-4 sm:p-6 flex-col gap-4 sm:gap-6'>

      <div className='w-full'>
        <DropZone onUpload={(file) => uploadMutation.mutate(file)} disabled={isAtLimit} />
      </div>
      {isAtLimit && (
        <p className="text-sm text-accent-red-deep font-bold">
          Has alcanzado el límite de {currentUser?.max_documents} documentos.
        </p>
      )}

      <div className="flex flex-col sm:flex-row items-stretch sm:items-center justify-between w-full gap-3">
        <div className="relative flex-1 max-w-md">
          <Search size={18} className="absolute left-3 top-1/2 -translate-y-1/2 text-neutral-500" />
          <input
            placeholder="Buscar archivos..."
            className="w-full pl-10 pr-4 py-2.5 bg-white border-2 border-neutral-950 rounded text-sm font-bold text-neutral-900 placeholder:text-neutral-400 focus:outline-none focus:shadow-hard-sm focus:bg-neutral-50 transition-all"
          />
        </div>
        <button className="flex items-center justify-center gap-2 px-4 py-2.5 bg-white border-2 border-neutral-950 rounded text-sm font-black text-neutral-950 hover:bg-neutral-100 hover:shadow-hard-sm transition-all cursor-pointer">
          <SlidersHorizontal size={16} strokeWidth={3} />
          Filtros
        </button>
      </div>

      {isLoading ? (
        <div className="flex-1 bg-white border-2 border-neutral-950 rounded-card p-12 flex flex-col items-center justify-center gap-4 shadow-hard-lg">
          <div className="w-12 h-12 bg-primary-400 border-2 border-neutral-950 rounded-full flex items-center justify-center animate-spin shadow-hard-sm">
            <span className="text-xl font-black">@</span>
          </div>
          <span className="text-neutral-950 font-bold">Cargando tus documentos...</span>
        </div>
      ) : (
        <div className="flex-1 min-h-0 flex flex-col bg-white border-2 border-neutral-950 rounded-card shadow-hard-lg overflow-hidden">
          <div className="flex-1 min-h-0 overflow-auto">
            <DocumentTable documents={documents}
              onDelete={(id) => deleteMutation.mutate(id)}
              onDownload={(id) => console.log(`download ${id}`)}
              onReprocess={(id) => reprocessMutation.mutate(id)}
              deletingIds={deletingIds}
              reprocessingIds={reprocessingIds} />
          </div>

          {totalPages > 1 && (
            <div className="flex flex-col sm:flex-row items-center justify-between px-4 sm:px-6 py-4 border-t-2 border-neutral-950 bg-neutral-100 shrink-0 gap-4">
              <p className="text-xs font-bold text-neutral-600 w-full sm:w-auto text-center sm:text-left">
                {data?.total ?? 0} documentos
              </p>
              <div className="flex items-center gap-2 max-w-full overflow-x-auto pb-1 sm:pb-0 hide-scrollbar">
                <button
                  onClick={() => setPage(Math.max(1, page - 1))}
                  disabled={page <= 1}
                  className="p-1.5 border-2 border-neutral-950 rounded disabled:opacity-30 disabled:pointer-events-none hover:bg-white transition-all cursor-pointer shrink-0"
                >
                  <ChevronLeft size={16} />
                </button>

                <div className="flex items-center gap-1.5">
                  {Array.from({ length: totalPages }, (_, i) => i + 1).map((p) => (
                    <button
                      key={p}
                      onClick={() => setPage(p)}
                      className={`px-3 py-1 border-2 border-neutral-950 rounded text-xs font-bold transition-all cursor-pointer shrink-0 ${
                        p === page
                          ? "bg-primary-400 text-black shadow-hard-sm"
                          : "bg-white text-neutral-950 hover:shadow-hard-sm"
                      }`}
                    >
                      {p}
                    </button>
                  ))}
                </div>

                <button
                  onClick={() => setPage(Math.min(totalPages, page + 1))}
                  disabled={page >= totalPages}
                  className="p-1.5 border-2 border-neutral-950 rounded disabled:opacity-30 disabled:pointer-events-none hover:bg-white transition-all cursor-pointer shrink-0"
                >
                  <ChevronRight size={16} />
                </button>
              </div>
            </div>
          )}
        </div>
      )}

    </div>
  )
}
