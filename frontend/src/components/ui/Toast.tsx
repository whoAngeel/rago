import type { Toast } from "../../store/toastStore"
import { useToastStore } from "../../store/toastStore"
import { CheckCircle2, AlertCircle, Info } from "lucide-react"
import type { ReactNode } from "react"

interface ToastProps {
    toast: Toast
}

const typeStyles: Record<Toast["type"], { bg: string, icon: ReactNode }> = {
    success: { bg: "bg-[#a3e635] text-black", icon: <CheckCircle2 className="w-6 h-6" /> },
    error: { bg: "bg-[#f4405e] text-white", icon: <AlertCircle className="w-6 h-6" /> },
    info: { bg: "bg-[#3c82f6] text-white", icon: <Info className="w-6 h-6" /> },
}

export function ToastItem({ toast }: ToastProps) {
    const remove = useToastStore((s) => s.remove)
    const style = typeStyles[toast.type]

    return (
        <div className={`relative flex items-start gap-4 w-full border-2 border-[#0a0a0d] rounded-[10px] shadow-[4px_4px_0_0_#0a0a0d] px-5 py-4 ${style.bg} font-sans`}>
            <div className="text-2xl leading-none mt-0.5">{style.icon}</div>
            <div className="flex-1 pr-4">
                {toast.title && <h4 className="font-bold text-base leading-snug mb-1 tracking-[-0.02em]">{toast.title}</h4>}
                <p className="text-sm font-medium opacity-90">{toast.message}</p>
            </div>
            <button 
                onClick={() => remove(toast.id)} 
                className="absolute top-4 right-4 text-2xl leading-none font-medium opacity-70 hover:opacity-100 transition-opacity"
                aria-label="Cerrar"
            >
                &times;
            </button>
        </div>
    )
}