import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/_authenticated/dashboard')({
  component: DashboardPage,
})


function DashboardPage() {
  return (
    <div className="p-8">
      <h1 className='text-h2'>Dashboard</h1>
      <p className="text-body">
        Gestion de documentos
      </p>
    </div>
  )
}
