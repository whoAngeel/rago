import { createRootRoute, Outlet } from "@tanstack/react-router";
import { TanStackRouterDevtools } from "@tanstack/react-router-devtools"
import { useEffect } from "react";
import { useAuthStore } from "../store/authStore";

export const Route = createRootRoute({
    component: RootComponent,
    notFoundComponent: () => {
        return (
            <div className="flex flex-col items-center justify-center min-h-screen gap-4">
                <p className="text-h2">404</p>
                <p className="text-body">Página no encontrada</p>
                {/* <Link to="/" className="text-body underline">Volver al inicio</Link> */}
            </div>
        )
    }
})

function RootComponent() {
    const { refresh, isLoading, refreshToken, accessToken } = useAuthStore()
    useEffect(() => {
        if (refreshToken && !accessToken) {
            refresh()
        }
    }, [])

    if (isLoading) return (
        <div className="flex items-center justify-center min-h-screen">
            <p className="text-body text-neutral-600">Cargando...</p>
        </div>
    )
    return (
        <>
            <Outlet />
            <TanStackRouterDevtools position="bottom-left" />
        </>
    )
}