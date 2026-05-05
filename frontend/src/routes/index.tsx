import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute('/')({
    component: Home,
})

function Home() {
    return (
        <div className="p-2">
            <h3>
                welcome home
            </h3></div>
    )
}