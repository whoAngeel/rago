import { useQuery, useQueryClient } from '@tanstack/react-query'
import { createFileRoute, Link } from '@tanstack/react-router'
import type { Group } from '../types'
import api from '../lib/api'
import { Button } from '../components/ui'
import { ArrowLeft, Loader2 } from 'lucide-react'

export const Route = createFileRoute('/_authenticated/groups/$id')({
  component: RouteComponent,
})

function RouteComponent() {
  const { id } = Route.useParams()
  const queryClient = useQueryClient()

  const { data, isLoading, error } = useQuery<Group>({
    queryKey: ["group", id],
    queryFn: async () => {
      const { data } = await api.get(`/groups/${id}`)
      return data
    },
  })

  if (isLoading) return <div className="flex items-center justify-center w-full h-full"> <Button disabled><Loader2 /> Cargando grupo...</Button></div>

  if (error) return <div className="flex items-center justify-center w-full h-full">Error al cargar el grupo</div>

  return (
    <div className='flex w-full h-full p-6 flex-col gap-6'>


    </div>
  )
}
