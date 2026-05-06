import { createFileRoute, Outlet, redirect, useNavigate } from '@tanstack/react-router'
import { Button } from '../components/ui'
import { useAuthStore } from '../store/authStore'

export const Route = createFileRoute('/_authenticated')({
  beforeLoad: async () => {
    const { accessToken } = await import('../store/authStore').then(m => m.useAuthStore.getState())
    if (!accessToken) {
      throw redirect({
        to: '/login',
        search: {
          redirect: window.location.pathname
        }
      })
    }

  },
  component: AuthenticatedLayout,
})

function AuthenticatedLayout() {
  const { logout } = useAuthStore()
  const navigate = useNavigate()

  const handleLogout = () => {
    logout()
    navigate({ to: '/login', search: { redirect: undefined } })
  }

  return (
    <div className="min-h-screen bg-neutral-100 p-4 font-sans">
      <header className="max-w-7xl mx-auto bg-white border border-neutral-800 rounded-full shadow-hard-md px-12 py-3 flex items-center justify-between mb-8">
        <div className="font-medium text-xl text-neutral-800">Rago</div>
        <Button variant="secondary" onClick={handleLogout}>
          Cerrar sesión
        </Button>
      </header>
      <main className="max-w-7xl mx-auto">
        <Outlet />
      </main>
    </div>
  )
}