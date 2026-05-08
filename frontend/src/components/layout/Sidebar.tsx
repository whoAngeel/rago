import { Link, useLocation } from "@tanstack/react-router"
import { LayoutDashboard, FileText, Settings, Fish } from "lucide-react"

const links = [
    { to: "/dashboard", label: "Dashboard", icon: LayoutDashboard, subtitle: "Visión general del sistema" },
    { to: "/documents", label: "Documentos", icon: FileText, subtitle: "Gestión de documentos" },
    { to: "/settings", label: "Configuración", icon: Settings, subtitle: "Preferencias del sistema" },
]

export const SideBar = () => {
    const { pathname } = useLocation()

    return (
        <aside className="w-64 h-screen bg-neutral-100 border-r-2 border-neutral-950 p-6 flex flex-col gap-8 font-sans">
            <div className="flex items-center gap-3 px-4">
                <div className="w-10 h-10 bg-primary-400 border-2 border-neutral-950 flex items-center justify-center rounded-lg shadow-hard-md">
                    <Fish size={24} className="text-black" />
                </div>
                <div>
                    <h1 className="text-2xl font-bold text-neutral-950 tracking-tighter">Rago</h1>
                    <p className="text-xs font-bold text-neutral-500 uppercase tracking-widest">RAG Platform</p>
                </div>
            </div>
            <nav className="flex flex-col gap-3">
                {
                    links.map((link) => (
                        <Link key={link.to} to={link.to}
                            className="flex items-center gap-3 px-4 py-3 rounded text-neutral-600 font-medium border-2 border-transparent hover:border-neutral-950 hover:bg-white hover:text-black hover:shadow-hard-md transition-all"
                            activeOptions={{ exact: link.to === "/dashboard" }}
                            activeProps={{ className: "bg-primary-400 text-black border-neutral-950 shadow-hard-md" }}>
                            <link.icon size={22} />
                            <span className="text-lg">{link.label}</span>
                        </Link>
                    ))
                }
            </nav>
        </aside>
    )
}