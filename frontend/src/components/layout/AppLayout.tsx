import { Outlet, Link } from "@tanstack/react-router";
import { SideBar } from "./Sidebar";
import { Navbar } from "./Navbar";
import { LayoutDashboard, FileText, Settings, Group } from "lucide-react";

const links = [
    { to: "/dashboard", label: "Dashboard", icon: LayoutDashboard },
    { to: "/documents", label: "Documentos", icon: FileText },
    { to: "/groups", label: "Grupos", icon: Group },
    { to: "/settings", label: "Config", icon: Settings },
];

export function AppLayout() {
    return (
        <div className="flex h-screen overflow-hidden bg-neutral-100 flex-col md:flex-row">
            <div className="hidden md:block">
                <SideBar />
            </div>
            <div className="flex-1 flex flex-col h-full overflow-hidden">
                <Navbar />
                <main className="flex-1 overflow-auto px-4 sm:px-8 py-4 sm:py-6 pb-20 md:pb-6">
                    <Outlet />
                </main>

                {/* Mobile Bottom Navigation */}
                <nav className="md:hidden fixed bottom-0 w-full bg-white border-t-2 border-neutral-950 flex items-center justify-around p-2 pb-[calc(env(safe-area-inset-bottom)+0.5rem)] z-50">
                    {links.map((link) => (
                        <Link
                            key={link.to}
                            to={link.to}
                            className="flex flex-col items-center justify-center p-2 text-neutral-500 font-bold transition-colors w-16"
                            activeOptions={{ exact: link.to === "/dashboard" }}
                            activeProps={{ className: "text-[#84cc17]" }}
                        >
                            <link.icon size={24} className="mb-1" />
                            <span className="text-[10px] leading-tight">{link.label}</span>
                        </Link>
                    ))}
                </nav>
            </div>
        </div>
    )
}