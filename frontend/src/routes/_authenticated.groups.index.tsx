import { createFileRoute } from '@tanstack/react-router'
import { useToastStore } from '../store/toastStore'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import api from '../lib/api'
import { useEffect, useState } from 'react'
import { GroupList } from '../components/groups/GroupList'
import type { Group } from '../types'
import { Plus, ChevronLeft, ChevronRight } from 'lucide-react'
import { Button } from '../components/ui'
import { CreateGroupModal } from '../components/groups/CreateGroupModal'

interface GroupsResponse {
  items: Group[]
  total: number
  page: number
  limit: number
}

export const Route = createFileRoute('/_authenticated/groups/')({
  component: RouteComponent,
})

function RouteComponent() {
  const addToast = useToastStore(s => s.add)
  const queryClient = useQueryClient()
  const [loadingIds, setLoadingIds] = useState<number[]>([])
  const [page, setPage] = useState(1)
  const [showCreateModal, setShowCreateModal] = useState(false)
  const limit = 4

  const { data, isLoading, error } = useQuery<GroupsResponse>({
    queryKey: ['groups', page],
    queryFn: async () => {
      const { data } = await api.get(`/groups/?page=${page}&limit=${limit}`)
      return data
    }
  })

  const totalPages = data ? Math.ceil(data.total / data.limit) : 0

  useEffect(() => {
    if (error) {
      addToast("Error", `No se pudieron cargar los grupos: ${(error as Error).message}`, "error")
    }
  }, [error, addToast])

  const deleteMutation = useMutation({
    mutationFn: async (id: number) => {
      return api.delete(`/groups/${id}/`)
    },
    onMutate: (id) => {
      setLoadingIds(prev => [...prev, id])
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['groups'] })
      addToast("Éxito", "Grupo eliminado correctamente", "success")
    },
    onError: (err: any) => {
      addToast("Error", `No se pudo eliminar el grupo: ${err.message}`, "error")
    },
    onSettled: () => {
      setLoadingIds([])
    }
  })

  const toggleMutation = useMutation({
    mutationFn: async ({ id, is_active, name }: { id: number; is_active: boolean; name: string }) => {
      return api.patch(`/groups/${id}/`, {
        name,
        is_active,
        allow_downloads: true,
      })
    },
    onMutate: ({ id }) => {
      setLoadingIds(prev => [...prev, id])
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['groups'] })
    },
    onError: (err: any) => {
      addToast("Error", `No se pudo actualizar el grupo: ${err.message}`, "error")
    },
    onSettled: () => {
      setLoadingIds([])
    }
  })


  const createMutation = useMutation({
    mutationFn: async (name: string) => {
      return api.post(`/groups/`, { name, document_ids: [] })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['groups'] })
      addToast("Éxito", "Grupo creado correctamente", "success")
    },
    onError: (err: any) => {
      addToast("Error", `No se pudo crear el grupo: ${err.message}`, "error")
    }
  })

  const handleOpenModal = () => {
    setShowCreateModal(true)
  }

  const handleCloseModal = () => {
    setShowCreateModal(false)
  }

  const handleCreateGroup = (name: string) => {
    createMutation.mutate(name)
  }


  return (
    <div className='flex w-full h-full p-6 flex-col gap-6'>

      {isLoading ? (
        <div className="flex-1 bg-white border-2 border-neutral-950 rounded-card p-12 flex flex-col items-center justify-center gap-4 shadow-hard-lg">
          <div className="w-12 h-12 bg-primary-400 border-2 border-neutral-950 rounded-full flex items-center justify-center animate-spin shadow-hard-sm">
            <span className="text-xl font-black">@</span>
          </div>
          <span className="text-neutral-950 font-bold">Cargando grupos...</span>
        </div>
      ) : (
        <div className="flex-1 min-h-0 flex flex-col gap-6">
          <div className='flex w-full shrink-0 justify-end'>
            <Button variant='primary' size='lg' className="flex items-center gap-2" onClick={handleOpenModal}>
              <Plus size={16} strokeWidth={3} />
              Nuevo grupo
            </Button>
          </div>

          <div className="flex-1 min-h-0 flex flex-col">
            <div className='flex-1 min-h-0 overflow-y-auto overflow-x-visible p-2'>
              <GroupList
                groups={data?.items ?? []}
                onToggle={(id, is_active, name) => toggleMutation.mutate({ id, is_active, name })}
                onDelete={(id) => deleteMutation.mutate(id)}
                loadingIds={loadingIds}
              />
            </div>

            {totalPages > 1 && (
              <div className="flex items-center justify-between py-4">
                <p className="text-xs font-bold text-neutral-600">{data?.total ?? 0} grupos</p>
                <div className="flex items-center gap-2">
                  <button onClick={() => setPage(Math.max(1, page - 1))} disabled={page <= 1}
                    className="p-1.5 border-2 border-neutral-950 rounded disabled:opacity-30 disabled:pointer-events-none hover:bg-white transition-all cursor-pointer"
                  ><ChevronLeft size={16} /></button>
                  <span className="text-xs font-bold text-neutral-700 px-2">Página {page} de {totalPages}</span>
                  <button onClick={() => setPage(Math.min(totalPages, page + 1))} disabled={page >= totalPages}
                    className="p-1.5 border-2 border-neutral-950 rounded disabled:opacity-30 disabled:pointer-events-none hover:bg-white transition-all cursor-pointer"
                  ><ChevronRight size={16} /></button>
                </div>
              </div>
            )}
          </div>
        </div>

      )}

      <CreateGroupModal isOpen={showCreateModal} onClose={handleCloseModal} onCreate={handleCreateGroup} isPending={createMutation.isPending} />
    </div>
  )
}
