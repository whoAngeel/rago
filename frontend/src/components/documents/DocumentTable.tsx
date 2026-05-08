import { formatSize } from "../../lib/formatter"
import type { Document } from "../../types"
import { Download, FileText, Trash2, FileSpreadsheet, FileJson, AlertCircle } from "lucide-react"
import { formatDistanceToNow } from "date-fns"
import { es } from "date-fns/locale"

interface DocumentsProps {
    documents: Document[]
    onDelete: (id: number) => void
    onDownload: (id: number) => void
    deletingIds?: number[]
}

const statusConfig = {
    uploading: { label: "Subiendo", bg: "bg-blue-pastel", dot: "bg-blue-600" },
    pending: { label: "Pendiente", bg: "bg-neutral-100", dot: "bg-neutral-500" },
    processing: { label: "Procesando", bg: "bg-accent-blue/30", dot: "bg-accent-blue animate-pulse" },
    completed: { label: "Completado", bg: "bg-primary-400/30", dot: "bg-primary-500" },
    failed: { label: "Fallido", bg: "bg-accent-red/30", dot: "bg-accent-red-deep" },
}

const getFileIcon = (contentType: string) => {
    const ct = contentType.toLowerCase()
    if (ct.includes("pdf")) return <FileText size={20} className="text-black" />
    if (ct.includes("csv") || ct.includes("excel") || ct.includes("spreadsheet") || ct.includes("openxmlformats")) return <FileSpreadsheet size={20} className="text-black" />
    if (ct.includes("json")) return <FileJson size={20} className="text-black" />
    return <FileText size={20} className="text-black" />
}

const getFileIconBg = (contentType: string) => {
    const ct = contentType.toLowerCase()
    if (ct.includes("pdf")) return "bg-accent-red-deep/20"
    if (ct.includes("csv") || ct.includes("excel") || ct.includes("spreadsheet") || ct.includes("openxmlformats")) return "bg-primary-300"
    if (ct.includes("json")) return "bg-accent-amber"
    return "bg-accent-blue"
}

const getFileTypeLabel = (contentType: string) => {
    const ct = contentType.toLowerCase()
    if (ct.includes("pdf")) return "PDF"
    if (ct.includes("csv")) return "CSV"
    if (ct.includes("json")) return "JSON"
    if (ct.includes("excel") || ct.includes("spreadsheet") || ct.includes("openxmlformats")) return "XLSX"
    if (ct.includes("text") || ct.includes("plain")) return "TXT"
    if (ct.includes("word") || ct.includes("officedocument.wordprocessingml")) return "DOCX"
    return "FILE"
}

const renderProgressDots = (status: string) => {
    const steps = ['download', 'parse', 'chunk', 'embed', 'upsert']
    let completedCount = 0
    let activeIndex = -1
    let isFailed = false

    if (status === "completed") {
        completedCount = 5
    } else if (status === "processing") {
        completedCount = 2 // Simulamos que va por el paso 3
        activeIndex = 2 // chunk
    } else if (status === "uploading") {
        completedCount = 0
        activeIndex = 0 // download
    } else if (status === "pending") {
        completedCount = 0
    } else if (status === "failed") {
        completedCount = 1 // Simulamos que falló en el parse
        activeIndex = 1
        isFailed = true
    }


    return (
        <div className="flex flex-col gap-1">
            <div className="flex items-center gap-1.5">
                {steps.map((step, i) => {
                    let className = "w-3 h-3 rounded-full border border-neutral-950 "
                    if (i < completedCount) {
                        className += "bg-primary-500"
                    } else if (i === activeIndex) {
                        if (isFailed) {
                            className += "bg-accent-red-deep"
                        } else {
                            className += "bg-accent-blue animate-pulse"
                        }
                    } else {
                        className += "bg-neutral-100"
                    }
                    return <div key={step} className={className} title={step} />
                })}
            </div>
            {activeIndex !== -1 && (
                <span className={`text-[10px] font-bold uppercase tracking-wider ${isFailed ? "text-accent-red-deep" : "text-neutral-500"}`}>
                    {isFailed ? `Fallo en: ${steps[activeIndex]}` : steps[activeIndex]}
                </span>
            )}
        </div>
    )
}

export const DocumentTable = ({ documents, onDelete, onDownload, deletingIds = [] }: DocumentsProps) => {
    return (
        <div className="bg-white border-2 border-neutral-950 rounded-card shadow-hard-lg overflow-hidden">
            <table className="w-full text-sm border-collapse">
                <thead>
                    <tr className="bg-neutral-200 border-b-2 border-neutral-950">
                        <th className="px-6 py-4 font-bold uppercase text-xs tracking-widest text-neutral-950 text-left">Nombre</th>
                        <th className="px-6 py-4 font-bold uppercase text-xs tracking-widest text-neutral-950 text-left">Tamaño</th>
                        <th className="px-6 py-4 font-bold uppercase text-xs tracking-widest text-neutral-950 text-left">Estado</th>
                        <th className="px-6 py-4 font-bold uppercase text-xs tracking-widest text-neutral-950 text-left">Progreso</th>
                        <th className="px-6 py-4 font-bold uppercase text-xs tracking-widest text-neutral-950 text-left">Subido</th>
                        <th className="px-6 py-4 font-bold uppercase text-xs tracking-widest text-neutral-950 text-center">Acciones</th>
                    </tr>
                </thead>
                <tbody className="divide-y divide-neutral-200">
                    {documents.length === 0 && (
                        <tr>
                            <td colSpan={6} className="px-6 py-12 text-center">
                                <div className="flex flex-col items-center justify-center gap-3">
                                    <div className="w-16 h-16 bg-neutral-100 border-2 border-neutral-950 rounded-full flex items-center justify-center shadow-hard-sm">
                                        <FileText size={32} className="text-neutral-400" />
                                    </div>
                                    <div>
                                        <p className="font-bold text-neutral-950 text-lg">No hay documentos aún</p>
                                        <p className="text-sm text-neutral-500">Sube un archivo para empezar a procesarlo</p>
                                    </div>
                                </div>
                            </td>
                        </tr>
                    )}
                    {documents.map(document => {
                        const isDeleting = deletingIds.includes(document.id)
                        return (
                            <tr key={document.id} className={`hover:bg-neutral-50 transition-colors ${isDeleting ? 'opacity-50 pointer-events-none' : ''}`}>
                                {/* Nombre */}
                                <td className="px-6 py-5">
                                    <div className="flex items-center gap-3">
                                        <div className={`w-10 h-10 ${getFileIconBg(document.content_type)} border-2 border-neutral-950 rounded flex items-center justify-center shadow-hard-sm shrink-0`}>
                                            {getFileIcon(document.content_type)}
                                        </div>
                                        <div>
                                            <p className="font-bold text-neutral-950 break-all">{document.filename}</p>
                                            <p className="text-xs text-neutral-500 font-medium">{getFileTypeLabel(document.content_type)}</p>
                                        </div>
                                    </div>
                                </td>

                                {/* Tamaño */}
                                <td className="px-6 py-5 font-medium text-neutral-950">
                                    {formatSize(document.size)}
                                </td>

                                {/* Estado */}
                                <td className="px-6 py-5">
                                    <span className={`px-3 py-1 ${statusConfig[document.status].bg} border border-neutral-950 rounded-md text-xs font-bold uppercase flex items-center gap-1 w-fit shadow-hard-sm`}>
                                        <span className={`w-2 h-2 ${statusConfig[document.status].dot} rounded-full`}></span>
                                        {statusConfig[document.status].label}
                                    </span>
                                </td>

                                {/* Progreso */}
                                <td className="px-6 py-5">
                                    {document.status === "failed" ? (
                                        <div className="flex items-center gap-1 text-accent-red-deep font-bold text-xs">
                                            <AlertCircle size={14} />
                                            <span className="break-all">{document.error_message || "Error de procesamiento"}</span>
                                        </div>
                                    ) : (
                                        renderProgressDots(document.status)
                                    )}
                                </td>

                                {/* Subido */}
                                <td className="px-6 py-5 text-neutral-600 font-medium text-xs">
                                    {formatDistanceToNow(new Date(document.created_at), { addSuffix: true, locale: es })}
                                </td>

                                {/* Acciones */}
                                <td className="px-6 py-5">
                                    <div className="flex items-center gap-2 justify-center">
                                        <button className="p-2 text-neutral-600 hover:text-neutral-950 border-2 
                                        border-transparent hover:border-neutral-950 hover:bg-white 
                                        hover:shadow-hard-sm rounded transition-all cursor-pointer"
                                            title="Descargar"
                                            onClick={() => onDownload(document.id)}
                                        >
                                            <Download size={16} />
                                        </button>
                                        <button className="p-2 text-neutral-600 hover:text-accent-red-deep 
                                        border-2 border-transparent hover:border-neutral-950 hover:bg-accent-red/10 
                                        hover:shadow-hard-sm rounded transition-all cursor-pointer" title="Eliminar"
                                            onClick={() => onDelete(document.id)}
                                        >
                                            <Trash2 size={16} />
                                        </button>
                                    </div>
                                </td>
                            </tr>
                        )
                    })}
                </tbody>
            </table>
        </div>
    )
}