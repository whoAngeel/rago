import { createFileRoute } from '@tanstack/react-router'
import { ArrowLeft } from 'lucide-react'

export const Route = createFileRoute('/_authenticated/groups/new')({
  validateSearch: (search: Record<string, unknown>) => {
    return {
      name: typeof search.name === 'string' ? search.name : undefined
    }
  },
  component: RouteComponent,
})

function RouteComponent() {
  const { name } = Route.useSearch()
  return (
    <div className='flex w-full h-full p-6 gap-6'>
      <div>
        {/** GroupName */}

      </div>


    </div>
  )
}
