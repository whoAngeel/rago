import { Outlet } from "@tanstack/react-router";
import { SideBar } from "./Sidebar";
import { Navbar } from "./Navbar";

export function AppLayout() {
    return (
        <div className="flex min-h-screen bg-neutral-100">
            <SideBar />
            <div className="flex-1 flex flex-col">
                <Navbar />
                <main className="flex-1 overflow-auto py-6">
                    <Outlet />
                </main>
            </div>
        </div>
    )
}