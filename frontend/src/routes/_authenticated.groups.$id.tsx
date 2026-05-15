import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { createFileRoute } from '@tanstack/react-router'
import type { Document, Group, User, GroupUsage } from '../types'
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

  const { data: userProfile } = useQuery<User>({
    queryKey: ['me'],
    queryFn: async () => {
      const { data } = await api.get('/users/me')
      return data
    },
  })

  const { data: usage } = useQuery<GroupUsage>({
    queryKey: ['group-usage', id],
    queryFn: async () => {
      const { data } = await api.get(`/groups/${id}/usage`)
      return data
    },
  })

  const rejectedAttempts = (group?.chat_attempts ?? 0) - (group?.chat_quota_used ?? 0)

  const isGroupQuotaExhausted =
    (group?.chat_quota ?? 0) > 0 &&
    (group?.chat_quota_used ?? 0) >= (group?.chat_quota ?? 0)

  const isUserQuotaExhausted =
    (userProfile?.chat_quota ?? 0) > 0 &&
    (userProfile?.chat_quota_used ?? 0) >= (userProfile?.chat_quota ?? 0)

  const isBlocked = isGroupQuotaExhausted || isUserQuotaExhausted

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
    <div className="w-full h-full p-4 sm:p-6 gap-6 flex flex-col xl:grid xl:grid-cols-12 font-sans overflow-auto">
      <div className="xl:col-span-9 flex flex-col gap-6">
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

          {isBlocked && (
            <div className="mb-4 p-3 border-2 border-neutral-950 rounded bg-amber-100">
              <p className="text-sm font-bold text-neutral-950">
                {isUserQuotaExhausted
                  ? "Tu cuota global de mensajes se agotó. Todos tus grupos están bloqueados."
                  : "La cuota de mensajes de este grupo se agotó. Los visitantes no pueden chatear."}
              </p>
            </div>
          )}

          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            {/* Card 1 — Cuota del grupo */}
            <div className="p-4 bg-neutral-50 border-2 border-neutral-950 rounded shadow-hard-sm flex flex-col gap-2">
              <p className="text-xs font-bold text-neutral-600 uppercase tracking-wider">Cuota del Grupo</p>
              <p className="text-sm font-black text-neutral-950">{group?.chat_quota_used} / {group?.chat_quota}</p>
              <div className="w-full h-4 bg-white border-2 border-neutral-950 rounded-full overflow-hidden shadow-hard-sm">
                <div
                  className={`h-full border-r-2 border-neutral-950 ${isGroupQuotaExhausted ? 'bg-accent-red-deep' : 'bg-primary-400'}`}
                  style={{ width: `${Math.min(100, (group?.chat_quota_used ?? 0) / (group?.chat_quota ?? 1) * 100)}%` }}
                />
              </div>
            </div>

            {/* Card 2 — Cuota del usuario */}
            <div className="p-4 bg-neutral-50 border-2 border-neutral-950 rounded shadow-hard-sm flex flex-col gap-2">
              <p className="text-xs font-bold text-neutral-600 uppercase tracking-wider">Cuota Global</p>
              <p className="text-sm font-black text-neutral-950">{userProfile?.chat_quota_used} / {userProfile?.chat_quota}</p>
              <div className="w-full h-4 bg-white border-2 border-neutral-950 rounded-full overflow-hidden shadow-hard-sm">
                <div
                  className={`h-full border-r-2 border-neutral-950 ${isUserQuotaExhausted ? 'bg-accent-red-deep' : 'bg-primary-400'}`}
                  style={{ width: `${Math.min(100, (userProfile?.chat_quota_used ?? 0) / (userProfile?.chat_quota ?? 1) * 100)}%` }}
                />
              </div>
            </div>

            {/* Card 3 — Tráfico */}
            <div className="p-4 bg-neutral-50 border-2 border-neutral-950 rounded shadow-hard-sm flex flex-col gap-2">
              <p className="text-xs font-bold text-neutral-600 uppercase tracking-wider">Tráfico y Efectividad</p>
              <p className="text-2xl font-black text-neutral-950">{group?.chat_attempts} intentos</p>
              <div className="flex gap-4 text-xs font-bold">
                <span className="text-primary-600">{group?.chat_quota_used} completados</span>
                <span className="text-accent-red-deep">{rejectedAttempts} rechazados</span>
              </div>
            </div>

            {/* Card 4 — Uso LLM */}
            <div className="p-4 bg-neutral-50 border-2 border-neutral-950 rounded shadow-hard-sm flex flex-col gap-2">
              <p className="text-xs font-bold text-neutral-600 uppercase tracking-wider">Uso LLM</p>
              <p className="text-2xl font-black text-neutral-950">{usage?.total_calls ?? 0} llamadas</p>
              <div className="flex gap-4 text-xs font-bold">
                <span className="text-neutral-600">{usage?.input_tokens ?? 0} tokens entrada</span>
                <span className="text-neutral-600">{usage?.output_tokens ?? 0} tokens salida</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div className="xl:col-span-3">
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
