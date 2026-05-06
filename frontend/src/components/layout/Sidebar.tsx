import { Link } from "@tanstack/react-router"
import { LayoutDashboard, FileText, Settings } from "lucide-react"

const links = [
    { to: "/dashboard", label: "Dashboard", icon: LayoutDashboard },
    { to: "/documents", label: "Documentos", icon: FileText },
    { to: "/settings", label: "Configuración", icon: Settings },
]


export const SideBar = () => {
    return (
        <aside className="w-64 h-screen bg-white border-r border-neutral-200 p-6 flex flex-col gap-2">

            <div className="text-xl font-bold text-neutral-900 mb-8">RAGO</div>
            <nav className="flex flex-col gap-1">
                {
                    links.map((link) => (
                        <Link key={link.to} to={link.to}
                            className="flex items-center gap-3 px-4 py-2.5 rounded-btn text-neutral-700 hover:bg-neutral-100 transition-colors"
                            activeOptions={{ exact: link.to === "/dashboard" }}
                            activeProps={{ className: "bg-primary text-primary-foreground" }}>
                            <link.icon size={20} />
                            {link.label}

                        </Link>
                    ))
                }

            </nav>
        </aside>
    )
}