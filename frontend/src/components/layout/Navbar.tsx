import { Search, Bell, HelpCircle } from "lucide-react"
import { useAuthStore, useUser } from "../../store/authStore"
import { useNavigate } from "@tanstack/react-router"
import { useState } from "react"

export function Navbar() {
    const user = useUser()
    const logout = useAuthStore((s) => s.logout)
    const navigate = useNavigate()
    const [showMenu, setShowMenu] = useState(false)

    const initials = user?.name
        ? user.name.split(" ").map((n) => n[0]).join("").toUpperCase().slice(0, 2)
        : "?"

    const handleLogout = async () => {
        await logout()
        navigate({ to: "/login", search: { redirect: "/dashboard" } })
    }

    return (
        <header className="h-16 bg-white border-b border-neutral-200 px-6 flex items-center justify-between">
            <div className="relative max-w-md w-full">
                <Search size={18} className="absolute left-3 top-1/2 -translate-y-1/2 text-neutral-400" />
                <input
                    placeholder="Buscar archivos..."
                    className="w-full pl-10 pr-4 py-2 bg-neutral-100 border border-neutral-200 rounded-btn text-sm text-neutral-900 placeholder:text-neutral-400 focus:border-neutral-900 focus:outline-none"
                />
            </div>

            <div className="flex items-center gap-4">
                <Bell size={20} className="text-neutral-600 cursor-pointer" />
                <HelpCircle size={20} className="text-neutral-600 cursor-pointer" />

                <div className="relative">
                    <button
                        onClick={() => setShowMenu(!showMenu)}
                        className="w-9 h-9 rounded-full bg-primary-400 text-black font-bold text-sm flex items-center justify-center"
                    >
                        {initials}
                    </button>

                    {showMenu && (
                        <>
                            <div className="fixed inset-0 z-10" onClick={() => setShowMenu(false)} />
                            <div className="absolute right-0 top-12 z-20 bg-white border border-neutral-200 rounded-card shadow-hard-md p-2 min-w-48">
                                <div className="px-3 py-2 border-b border-neutral-100">
                                    <p className="text-sm font-medium text-neutral-900">{user?.name}</p>
                                    <p className="text-xs text-neutral-500">{user?.email}</p>
                                </div>
                                <button
                                    onClick={handleLogout}
                                    className="w-full text-left px-3 py-2 text-sm text-red-600 hover:bg-red-50 rounded-btn mt-1"
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