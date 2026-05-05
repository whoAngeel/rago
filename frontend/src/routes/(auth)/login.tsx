import { createFileRoute } from '@tanstack/react-router'
import { Button } from '../../components/ui'
import { useAuthStore, useIsAuthenticated } from '../../store/authStore'

export const Route = createFileRoute('/(auth)/login')({
    component: RouteComponent,
})

function RouteComponent() {
    const { login, isLoading } = useAuthStore()
    const isAuthenticated = useIsAuthenticated()

    if (isAuthenticated) return <Button variant='secondary' onClick={useAuthStore.getState().logout}>Logout</Button>
    return <div>

        <Button onClick={() => login('whoangel.agl@gmail.com', 'whoangel')} isLoading={isLoading}>
            {isLoading ? "Cargando..." : "Iniciar sesion"}
        </Button>
    </div>
}
