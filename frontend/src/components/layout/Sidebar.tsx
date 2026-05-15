import { Link, useLocation } from "@tanstack/react-router"
import { LayoutDashboard, FileText, Settings, Fish, HardDrive, Group } from "lucide-react"
import { useQuery } from "@tanstack/react-query"
import api from "../../lib/api"
import type { User } from "../../types"

const links = [
    { to: "/dashboard", label: "Dashboard", icon: LayoutDashboard },
    { to: "/documents", label: "Documentos", icon: FileText },
    { to: "/settings", label: "Configuración", icon: Settings },
    { to: "/groups", label: "Grupos", icon: Group },

]

export const SideBar = () => {
    const { pathname } = useLocation()

    const { data: user } = useQuery<User>({
        queryKey: ["users", "me"],
        queryFn: async () => {
            const { data } = await api.get("/users/me")
            return data
        },
        staleTime: 30000,
    })

    const isUnlimited = user?.max_documents === 0
    const used = user?.document_count ?? 0
    const max = user?.max_documents ?? 0
    const percent = isUnlimited ? 100 : Math.min((used / max) * 100, 100)

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
            <nav className="flex flex-col gap-3 flex-1">
                {
                    links.map((link) => (
                        <Link key={link.to} to={link.to}
                            className="flex items-center gap-3 px-4 py-3 rounded text-neutral-600 font-medium border-2 border-transparent hover:border-neutral-950 hover:bg-white hover:text-black hover:shadow-hard-md transition-all"
                            activeOptions={{ exact: link.to === "/dashboard" }}
                            activeProps={{ className: "bg-primary-400 text-black border-neutral-950 shadow-hard-md" }}>
                            <link.icon size={32} />
                            <span className="text-lg">{link.label}</span>
                        </Link>
                    ))
                }
            </nav>

            <div className="px-4 py-4 bg-white border-2 border-neutral-950 rounded shadow-hard-sm mb-8">
                <div className="flex items-center gap-2 mb-3">
                    <FileText size={16} className="text-neutral-600" />
                    <span className="text-xs font-bold text-neutral-600 uppercase tracking-wider">Documentos</span>
                </div>
                {isUnlimited ? (
                    <div className="flex items-baseline justify-between">
                        <p className="text-3xl font-black text-neutral-950">{used}</p>
                        <span className="px-2 py-0.5 bg-primary-400 border border-neutral-950 rounded text-[10px] font-bold uppercase shadow-hard-sm">Ilimitado</span>
                    </div>
                ) : (
                    <>
                        <div className="flex items-baseline justify-between mb-1">
                            <p className="text-3xl font-black text-neutral-950">{used}</p>
                            <p className="text-xs font-bold text-neutral-500 uppercase">Límite: {max}</p>
                        </div>
                        <div className="w-full h-3 bg-neutral-100 border-2 border-neutral-950 rounded-full overflow-hidden">
                            <div
                                className={`h-full ${percent > 90 ? 'bg-accent-red-deep' : 'bg-primary-400'} transition-all duration-300`}
                                style={{ width: `${percent}%` }}
                            />
                        </div>
                    </>
                )}
            </div>
        </aside>
    )
}