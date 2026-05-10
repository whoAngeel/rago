import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { createFileRoute } from '@tanstack/react-router'
import type { Document, Group } from '../types'
import api from '../lib/api'
import { Button } from '../components/ui'
import { Loader2 } from 'lucide-react'
import { GroupHeader } from '../components/groups/GroupHeader'
import { GroupDocuments } from '../components/groups/GroupDocuments'
import { GroupSidebar } from '../components/groups/GroupSidebar'

export const Route = createFileRoute('/_authenticated/groups/$id')({
  component: RouteComponent,
})

function RouteComponent() {
  const { id } = Route.useParams()
  const queryClient = useQueryClient()

  const { data: group, isLoading, error } = useQuery<Group>({
    queryKey: ['group', id],
    queryFn: async () => {
      const { data } = await api.get(`/groups/${id}`)
      return data
    },
  })

  const { data: availableDocs } = useQuery<Document[]>({
    queryKey: ["documents", "select"],
    queryFn: async () => {
      const { data } = await api.get("/documents/select")
      return data
    },
  })

  const docsNotInGroup = availableDocs?.filter(
    (doc) => !group?.document_ids?.includes(doc.id)
  ) ?? []

  const updateMutation = useMutation({
    mutationFn: async (body: Record<string, unknown>) => {
      return api.patch(`/groups/${id}/`, body)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['group', id] })
      queryClient.invalidateQueries({ queryKey: ['groups'] })
    },
  })

  const addDocMutation = useMutation({
    mutationFn: async (docIds: number[]) => {
      return api.post(`/groups/${id}/documents`, { document_ids: docIds })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['group', id] })
    },
  })

  const removeDocMutation = useMutation({
    mutationFn: async (docId: number) => {
      return api.delete(`/groups/${id}/documents/${docId}`)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['group', id] })
    },
  })

  if (isLoading) return <div className="flex items-center justify-center w-full h-full"><Button disabled><Loader2 className="animate-spin" /> Cargando grupo...</Button></div>
  if (error) return <div className="flex items-center justify-center w-full h-full text-neutral-600 font-bold">Error al cargar el grupo</div>

  return (
    <div className="w-full h-full p-6 gap-6 grid grid-cols-12 font-sans">
      <div className="col-span-9 flex flex-col gap-6">
        <GroupHeader
          name={group?.name ?? ''}
          createdAt={group?.created_at}
          updatedAt={group?.updated_at}
          documentCount={group?.document_ids?.length ?? 0}
          onSaveName={(newName) => updateMutation.mutate({ name: newName })}
        />

        <div className="border-t-2 border-neutral-950 my-2" />

        <GroupDocuments
          allDocuments={availableDocs}
          selectedIds={group?.document_ids ?? []}
          onToggle={(docId, isSelected) => {
            if (isSelected) {
              removeDocMutation.mutate(docId)
            } else {
              addDocMutation.mutate([docId])
            }
          }}
        />

        <div className="border-t-2 border-neutral-950 my-2" />

        {/* Stats Section */}
        <div className="border-2 border-neutral-950 p-6 bg-white rounded shadow-hard-md">
          <h4 className="font-bold mb-4 text-lg text-neutral-950">Estadísticas de Uso</h4>
          <div className="grid grid-cols-2 gap-4">
            {/* Quota Usage */}
            <div className="p-4 bg-neutral-50 border-2 border-neutral-950 rounded shadow-hard-sm flex flex-col gap-2">
              <div className="flex justify-between items-baseline">
                <p className="text-xs font-bold text-neutral-600 uppercase tracking-wider mb-1">Cuota de Mensajes</p>
                <p className="text-sm font-black text-neutral-950">{group?.chat_quota_used} / {group?.chat_quota}</p>
              </div>
              <div className="w-full h-4 bg-white border-2 border-neutral-950 rounded-full overflow-hidden shadow-hard-sm">
                <div 
                  className="h-full bg-primary-400 border-r-2 border-neutral-950" 
                  style={{ width: `${Math.min(100, (group?.chat_quota_used ?? 0) / (group?.chat_quota ?? 1) * 100)}%` }}
                />
              </div>
              <p className="text-xs font-bold text-neutral-500 uppercase mt-auto">Mensajes completados con éxito</p>
            </div>

            {/* Traffic & Effectiveness */}
            <div className="p-4 bg-neutral-50 border-2 border-neutral-950 rounded shadow-hard-sm flex flex-col gap-2">
              <p className="text-xs font-bold text-neutral-600 uppercase tracking-wider mb-1">Tráfico y Efectividad</p>
              <div className="flex items-baseline gap-2">
                <p className="text-3xl font-black text-neutral-950">{group?.chat_attempts}</p>
                <p className="text-xs font-bold text-neutral-500 uppercase">Intentos totales</p>
              </div>
              
              <div className="flex gap-6 mt-1 border-t-2 border-neutral-200 pt-2">
                <div className="flex flex-col">
                  <span className="text-xs font-bold text-neutral-500 uppercase">Completados</span>
                  <span className="text-xl font-black text-primary-600">{group?.chat_quota_used}</span>
                </div>
                <div className="flex flex-col">
                  <span className="text-xs font-bold text-neutral-500 uppercase">Rechazados</span>
                  <span className="text-xl font-black text-red-600">{(group?.chat_attempts ?? 0) - (group?.chat_quota_used ?? 0)}</span>
                </div>
              </div>
              <p className="text-xs font-medium text-neutral-500 mt-2 border-t border-neutral-200 pt-1">
                * Los rechazos incluyen intentos por cuota de grupo agotada o grupo inactivo.
              </p>
            </div>
          </div>
        </div>
      </div>

      <div className="col-span-3">
        <GroupSidebar
          slug={group?.slug}
          allowDownloads={group?.allow_downloads ?? false}
          isActive={group?.is_active ?? false}
          onUpdate={(updates) => updateMutation.mutate(updates)}
        />
      </div>
    </div>
  )
}
