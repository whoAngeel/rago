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
        <header className="h-20 bg-white border-b-2 border-[#0a0a0d] px-8 flex items-center justify-between font-sans">
            <div className="relative max-w-md w-full">
                <Search size={20} className="absolute left-3 top-1/2 -translate-y-1/2 text-[#525252]" />
                <input
                    placeholder="Buscar archivos..."
                    className="w-full pl-10 pr-4 py-3 bg-white border border-[#737373] focus:border-[#0a0a0d] rounded text-base font-medium text-black placeholder:text-[#525252] focus:outline-none transition-colors"
                />
            </div>

            <div className="flex items-center gap-6">
                <button className="text-[#0a0a0d] hover:bg-[#a3e635] p-2 rounded-full border-2 border-transparent hover:border-[#0a0a0d] hover:shadow-[2px_2px_0_0_#0a0a0d] transition-all">
                    <Bell size={22} />
                </button>
                <button className="text-[#0a0a0d] hover:bg-[#a3e635] p-2 rounded-full border-2 border-transparent hover:border-[#0a0a0d] hover:shadow-[2px_2px_0_0_#0a0a0d] transition-all">
                    <HelpCircle size={22} />
                </button>

                <div className="relative">
                    <button
                        onClick={() => setShowMenu(!showMenu)}
                        className="w-11 h-11 rounded-full bg-[#a3e635] text-black font-bold text-base flex items-center justify-center border-2 border-[#0a0a0d] shadow-[2px_2px_0_0_#0a0a0d] hover:translate-x-[2px] hover:translate-y-[2px] hover:shadow-none transition-all"
                    >
                        {initials}
                    </button>

                    {showMenu && (
                        <>
                            <div className="fixed inset-0 z-10" onClick={() => setShowMenu(false)} />
                            <div className="absolute right-0 top-14 z-20 bg-white border-2 border-[#0a0a0d] rounded-[10px] shadow-[4px_4px_0_0_#0a0a0d] p-3 min-w-56 flex flex-col gap-2">
                                <div className="px-2 py-2 border-b-2 border-[#e5e5e5] mb-1">
                                    <p className="text-base font-bold text-black tracking-[-0.02em]">{user?.name}</p>
                                    <p className="text-sm font-medium text-[#525252]">{user?.email}</p>
                                </div>
                                <button
                                    onClick={handleLogout}
                                    className="w-full text-left px-3 py-2 text-base font-bold text-[#f4405e] hover:bg-[#ffe5e6] border-2 border-transparent hover:border-[#f4405e] rounded transition-colors"
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