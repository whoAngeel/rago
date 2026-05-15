import { Bell, HelpCircle, ArrowLeft } from "lucide-react"
import { useAuthStore, useUser } from "../../store/authStore"
import { useNavigate, useLocation } from "@tanstack/react-router"
import { useState } from "react"

const pageTitles: Record<string, { label: string; subtitle: string; parent?: string }> = {
    "/dashboard": { label: "Dashboard", subtitle: "Visión general del sistema" },
    "/documents": { label: "Documentos", subtitle: "Gestión de documentos" },
    "/settings": { label: "Configuración", subtitle: "Preferencias del sistema" },
    "/groups": { label: "Grupos", subtitle: "Gestión de grupos de documentos" },
}

function getCurrentPage(pathname: string) {
    if (pageTitles[pathname]) return pageTitles[pathname]
    if (pathname.startsWith("/groups/") && pathname !== "/groups") {
        return { label: "Detalle del Grupo", subtitle: "", parent: "/groups" as const }
    }
    return null
}

export function Navbar() {
    const user = useUser()
    const logout = useAuthStore((s) => s.logout)
    const navigate = useNavigate()
    const { pathname } = useLocation()
    const [showMenu, setShowMenu] = useState(false)
    const current = getCurrentPage(pathname)

    const initials = user?.name
        ? user.name.split(" ").map((n) => n[0]).join("").toUpperCase().slice(0, 2)
        : "?"

    const handleLogout = async () => {
        await logout()
        navigate({ to: "/login", search: { redirect: "/dashboard" } })
    }

    return (
        <header className="h-20 bg-white border-b-2 border-neutral-950 px-4 sm:px-8 flex items-center justify-between font-sans shrink-0">
            <div className="flex items-center gap-2 sm:gap-4">
                {current?.parent && (
                    <button
                        onClick={() => navigate({ to: current.parent })}
                        className="p-2 border-2 border-neutral-950 rounded hover:bg-neutral-100 hover:shadow-hard-sm transition-all cursor-pointer"
                    >
                        <ArrowLeft size={18} />
                    </button>
                )}
                {current && (
                    <div>
                        <h1 className="text-xl font-bold text-neutral-950">{current.label}</h1>
                        <p className="text-xs font-medium text-neutral-500">{current.subtitle}</p>
                    </div>
                )}
            </div>
            <div className="flex items-center gap-6">
                <button className="text-neutral-950 hover:bg-primary-400 p-2 rounded-full border-2 border-transparent hover:border-[#0a0a0d] hover:shadow-[2px_2px_0_0_#0a0a0d] transition-all">
                    <Bell size={22} />
                </button>
                <button className="text-neutral-950 hover:bg-primary-400 p-2 rounded-full border-2 border-transparent hover:border-[#0a0a0d] hover:shadow-[2px_2px_0_0_#0a0a0d] transition-all">
                    <HelpCircle size={22} />
                </button>

                <div className="relative">
                    <button
                        onClick={() => setShowMenu(!showMenu)}
                        className="w-11 h-11 rounded-full bg-primary-400 text-black font-bold text-base flex items-center justify-center border-2 border-[#0a0a0d] shadow-[2px_2px_0_0_#0a0a0d] hover:translate-x-[2px] hover:translate-y-[2px] hover:shadow-none transition-all"
                    >
                        {initials}
                    </button>

                    {showMenu && (
                        <>
                            <div className="fixed inset-0 z-10" onClick={() => setShowMenu(false)} />
                            <div className="absolute right-0 top-14 z-20 bg-white border-2 border-[#0a0a0d] rounded-[10px] shadow-[4px_4px_0_0_#0a0a0d] p-3 min-w-56 flex flex-col gap-2">
                                <div className="px-2 py-2 border-b-2 border-neutral-200 mb-1">
                                    <p className="text-base font-bold text-black tracking-[-0.02em]">{user?.name}</p>
                                    <p className="text-sm font-medium text-neutral-600">{user?.email}</p>
                                </div>
                                <button
                                    onClick={handleLogout}
                                    className="w-full text-left px-3 py-2 text-base font-bold text-[#f4405e] hover:bg-accent-red border-2 border-transparent hover:border-accent-red-deep rounded transition-colors"
                                >
                                    Cerrar sesión
                                </button>
                            </div>
                        </>
                    )}
                </div>
            </div>
        </header>
    )
}