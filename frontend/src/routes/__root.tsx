import { createRootRoute, Outlet } from "@tanstack/react-router";
import { TanStackRouterDevtools } from "@tanstack/react-router-devtools"
import { useEffect } from "react";
import { useAuthStore } from "../store/authStore";
import { ToastContainer } from "../components/ui/ToastContainer";

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
    const { refresh, isLoading, isRefreshing, refreshToken, accessToken } = useAuthStore()
    useEffect(() => {
        if (refreshToken && !accessToken) {
            refresh()
        }
    }, [])

    if (isRefreshing) return (
        <div className="flex items-center justify-center min-h-screen bg-neutral-50">
            <div className="flex flex-col items-center gap-6">
                <div className="bg-white border-2 border-neutral-950 rounded-card shadow-hard-lg p-6 flex items-end gap-2 h-20">
                    <div className="w-5 bg-primary-400 border-2 border-neutral-950 rounded animate-[loader-bar_1s_ease-in-out_infinite]" style={{ animationDelay: '0ms' }} />
                    <div className="w-5 bg-primary-400 border-2 border-neutral-950 rounded animate-[loader-bar_1s_ease-in-out_infinite]" style={{ animationDelay: '120ms' }} />
                    <div className="w-5 bg-primary-400 border-2 border-neutral-950 rounded animate-[loader-bar_1s_ease-in-out_infinite]" style={{ animationDelay: '240ms' }} />
                    <div className="w-5 bg-primary-400 border-2 border-neutral-950 rounded animate-[loader-bar_1s_ease-in-out_infinite]" style={{ animationDelay: '360ms' }} />
                    <div className="w-5 bg-primary-400 border-2 border-neutral-950 rounded animate-[loader-bar_1s_ease-in-out_infinite]" style={{ animationDelay: '480ms' }} />
                </div>
                <p className="font-bold text-neutral-600 uppercase tracking-widest text-xs">Cargando</p>
            </div>
        </div>
    )
    return (
        <>
            <ToastContainer />
            <Outlet />
            <TanStackRouterDevtools position="bottom-left" />
        </>
    )
}