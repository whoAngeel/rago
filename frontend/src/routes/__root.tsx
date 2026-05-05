import { createRootRoute, Link, Outlet } from "@tanstack/react-router";
import { TanStackRouterDevtools } from "@tanstack/react-router-devtools"

export const Route = createRootRoute({
    component: RootComponent,
    notFoundComponent: () => {
        return (
            <div>
                <p>404 - Pagina no encontrada</p>
                <Link to="/">Volver al inicio</Link>
            </div>
        )
    }
})

function RootComponent() {
    return (
        <>
            <div className="p-2 flex gap-2 text-lg border-b">
                <Link to="/"
                    activeProps={{
                        className: "font-bold"
                    }}
                    activeOptions={{ exact: true }}
                >
                    Home
                </Link> {' '}
                <Link to="/login">
                    Login
                </Link>{' '}

                <Link to="/register">
                    Register
                </Link>{' '}

            </div>
            <hr />
            <Outlet />
            <TanStackRouterDevtools position="bottom-left" />
        </>
    )
}