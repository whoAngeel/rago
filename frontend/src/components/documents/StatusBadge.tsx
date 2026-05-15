import type { Document } from "../../types"

interface StatusBadgeProps {
    status: Document["status"]
}

const statusConfig = {
    uploading: { label: "Subiendo", bg: "bg-blue-100 text-blue-900" },
    pending: { label: "Pendiente", bg: "bg-yellow-100 text-yellow-900" },
    processing: { label: "Procesando", bg: "bg-amber-100 text-amber-900" },
    completed: { label: "Completado", bg: "bg-green-100 text-green-900" },
    failed: { label: "Fallido", bg: "bg-red-100 text-red-900" },
}

const StatusBadge = ({ status }: StatusBadgeProps) => {

    const config = statusConfig[status]

    return (
        <span className={`inline-block px-2.5 py-1 rounded-btn text-xs font-medium ${config.bg}`}>
            {config.label}
        </span>
    )

}

export default StatusBadge