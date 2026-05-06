import { useToastStore } from "../../store/toastStore"
import { ToastItem } from "./Toast"

export function ToastContainer() {
    const toasts = useToastStore((s) => s.toasts)

    if (toasts.length === 0) return null

    return (
        <div className="fixed bottom-6 right-6 flex flex-col gap-4 z-50 w-full max-w-[380px] pointer-events-none">
            {toasts.map((toast) => (
                <div key={toast.id} className="pointer-events-auto">
                    <ToastItem toast={toast} />
                </div>
            ))}
        </div>
    )
}