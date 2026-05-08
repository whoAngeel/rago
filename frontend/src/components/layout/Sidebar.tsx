import { Link } from "@tanstack/react-router"
import { LayoutDashboard, FileText, Settings, Fish } from "lucide-react"

const links = [
    { to: "/dashboard", label: "Dashboard", icon: LayoutDashboard },
    { to: "/documents", label: "Documentos", icon: FileText },
    { to: "/settings", label: "Configuración", icon: Settings },
]

export const SideBar = () => {
    return (
        <aside className="w-60 h-screen bg-neutral-100 border-r-2 border-neutral-950 p-6 flex flex-col gap-8 font-sans">
            <div className="text-3xl font-bold text-black tracking-[-0.02em] flex items-center gap-3">
                <Fish size={32} className="text-primary-400" />
                RAGO
            </div>
            <div class="px-6 mb-4">
                <div class="flex items-center gap-2">
                    <div
                        class="w-10 h-10 bg-primary-container border-2 border-black flex items-center justify-center rounded-lg shadow-[2px_2px_0px_0px_rgba(0,0,0,1)]">
                        <span class="material-symbols-outlined text-black font-bold">bubble_chart</span>
                    </div>
                    <div>
                        <h1 class="font-h3-card text-[24px] font-black text-on-surface tracking-tighter">Rago</h1>
                        <p class="text-[12px] font-bold text-on-surface-variant uppercase tracking-widest">RAG Platform
                        </p>
                    </div>
                </div>
            </div>
            <nav className="flex flex-col gap-3">
                {
                    links.map((link) => (
                        <Link key={link.to} to={link.to}
                            className="flex items-center gap-3 px-4 py-3 rounded text-neutral-600 font-medium border-2 border-transparent hover:border-neutral-950 hover:bg-white hover:text-black hover:shadow-hard-md transition-all"
                            activeOptions={{ exact: link.to === "/dashboard" }}
                            activeProps={{ className: "bg-[#a3e635] text-black border-[#0a0a0d] shadow-[2px_2px_0_0_#0a0a0d]" }}>
                            <link.icon size={22} />
                            <span className="text-lg">{link.label}</span>
                        </Link>
                    ))
                }
            </nav>
        </aside>
    )
}