import { createFileRoute } from '@tanstack/react-router'
import { useToastStore } from '../store/toastStore'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import api from '../lib/api'
import { useEffect } from 'react'
import { GroupList } from '../components/groups/GroupList'
import type { Group } from '../types'
import { Plus } from 'lucide-react'
import { Button } from '../components/ui'

export const Route = createFileRoute('/_authenticated/groups')({
  component: RouteComponent,
})

interface GroupsResponse {
  items: Group[]

}

function RouteComponent() {
  const addToast = useToastStore(s => s.add)
  const queryClient = useQueryClient()

  const { data, isLoading, error } = useQuery<GroupsResponse>({
    queryKey: ['groups'],
    queryFn: async () => {
      const { data } = await api.get(`/groups/`)
      console.log(data)
      return data
    }

  })

  useEffect(() => {
    if (error) {
      addToast("Error", `No se pudieron cargar los grupos: ${(error as Error).message}`, "error")
    }
  }, [error, addToast])

  function toggleGroup(id: number) {
    console.log(`toggle ${id}`)
  }

  function deleteGroup(id: number) {
    console.log(`delete ${id}`)
  }

  return (
    <div className='flex w-full h-full p-6 flex-col gap-6'>

      {isLoading ? (
        <div className="flex-1 bg-white border-2 border-neutral-950 rounded-card p-12 flex flex-col items-center justify-center gap-4 shadow-hard-lg">
          <div className="w-12 h-12 bg-primary-400 border-2 border-neutral-950 rounded-full flex items-center justify-center animate-spin shadow-hard-sm">
            <span className="text-xl font-black">@</span>
          </div>
          <span className="text-neutral-950 font-bold">Cargando tus documentos...</span>
        </div>
      ) : (<div>

        <div className='flex w-full h-auto gap-4 justify-end'>

          <Button variant='primary' size='lg' className="flex items-center gap-2">
            <Plus size={16} strokeWidth={3} />
            Nuevo grupo
          </Button>
        </div>
        <div className='mt-8'>
          <GroupList groups={data?.items ?? []} onToggle={toggleGroup} onDelete={deleteGroup} />

        </div>


      </div>)}

    </div>
  )

}
