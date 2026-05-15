import { Outlet } from "@tanstack/react-router";
import { SideBar } from "./Sidebar";
import { Navbar } from "./Navbar";

export function AppLayout() {
    return (
        <div className="flex h-screen overflow-hidden bg-neutral-100">
            <SideBar />
            <div className="flex-1 flex flex-col h-full overflow-hidden">
                <Navbar />
                <main className="flex-1 overflow-auto px-8 py-6">
                    <Outlet />
                </main>
            </div>
        </div>
    )
}