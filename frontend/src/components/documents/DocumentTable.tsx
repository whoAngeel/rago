import { useState, useRef } from "react"
import { createPortal } from "react-dom"
import { useQuery } from "@tanstack/react-query"
import { formatSize } from "../../lib/formatter"
import type { Document, ProcessingStep } from "../../types"
import { Download, FileText, Trash2, FileSpreadsheet, FileJson, AlertCircle, Check, X, RefreshCw, Circle } from "lucide-react"
import { formatDistanceToNow } from "date-fns"
import { es } from "date-fns/locale"
import api from "../../lib/api"

interface DocumentsProps {
    documents: Document[]
    onDelete: (id: number) => void
    onDownload: (id: number) => void
    deletingIds?: number[]
}

const statusConfig = {
    uploading: { label: "Subiendo", bg: "bg-blue-pastel", dot: "bg-blue-600" },
    pending: { label: "Pendiente", bg: "bg-neutral-100", dot: "bg-neutral-500" },
    processing: { label: "Procesando", bg: "bg-accent-blue/20 animate-pulse", dot: "bg-blue-700" },
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

const STEP_NAMES = ['download', 'parse', 'chunk', 'embed', 'upsert'] as const

const formatDuration = (ms: number | null | undefined): string => {
    if (ms == null) return '—'
    if (ms < 1000) return `${ms}ms`
    return `${(ms / 1000).toFixed(1)}s`
}

function ProgressDots({ doc }: { doc: Document }) {
    const [isHovered, setIsHovered] = useState(false)
    const [tooltipPos, setTooltipPos] = useState({ top: 0, left: 0 })
    const ref = useRef<HTMLDivElement>(null)

    const canHaveSteps = doc.status === 'completed' || doc.status === 'failed'

    const { data: steps } = useQuery<ProcessingStep[]>({
        queryKey: ["documents", doc.id, "steps"],
        queryFn: async () => {
            const { data } = await api.get(`/documents/${doc.id}/steps`)
            return data
        },
        enabled: isHovered && canHaveSteps,
        staleTime: Infinity,
    })

    const handleMouseEnter = () => {
        if (ref.current && canHaveSteps) {
            const rect = ref.current.getBoundingClientRect()
            setTooltipPos({ top: rect.bottom + 6, left: rect.left })
        }
        setIsHovered(true)
    }

    const stepMap = new Map(steps?.map(s => [s.step_name, s]) ?? [])

    const getDotClass = (name: string, i: number): string => {
        const base = "w-3 h-3 rounded-full border border-neutral-950 "
        if (steps !== undefined) {
            const step = stepMap.get(name)
            if (step?.status === 'completed') return base + "bg-primary-500"
            if (step?.status === 'failed') return base + "bg-accent-red-deep"
            if (step?.status === 'started') return base + "bg-blue-700 animate-pulse"
            if (doc.status === 'completed') return base + "bg-primary-500"
            return base + "bg-neutral-100"
        }
        // Fallback simulated state while steps not loaded
        if (doc.status === 'completed') return base + "bg-primary-500"
        if (doc.status === 'processing') {
            if (i < 2) return base + "bg-primary-500"
            if (i === 2) return base + "bg-blue-700 animate-pulse"
        }
        if (doc.status === 'failed') {
            if (i < 1) return base + "bg-primary-500"
            if (i === 1) return base + "bg-accent-red-deep"
        }
        return base + "bg-neutral-100"
    }

    return (
        <>
            <div
                ref={ref}
                className={`flex items-center gap-1.5 w-fit ${canHaveSteps ? 'cursor-help' : 'cursor-default'}`}
                onMouseEnter={handleMouseEnter}
                onMouseLeave={() => setIsHovered(false)}
            >
                {STEP_NAMES.map((name, i) => (
                    <div key={name} className={getDotClass(name, i)} />
                ))}
            </div>

            {isHovered && canHaveSteps && createPortal(
                <div
                    style={{ top: tooltipPos.top, left: tooltipPos.left }}
                    className="fixed z-50 pointer-events-none bg-white border-2 border-neutral-950 rounded shadow-hard-sm w-44"
                >
                    <div className="px-2 py-1 bg-neutral-100 border-b-2 border-neutral-950">
                        <span className="text-[10px] font-bold uppercase tracking-wider text-neutral-950">Pipeline</span>
                    </div>
                    {steps !== undefined ? (
                        <div className="divide-y divide-neutral-200">
                            {STEP_NAMES.map(name => {
                                const step = stepMap.get(name)
                                const isCompleted = step?.status === 'completed' || (doc.status === 'completed' && !step)
                                const isFailed = step?.status === 'failed'
                                const isActive = step?.status === 'started'
                                return (
                                    <div key={name} className="px-2 py-1 flex items-center justify-between">
                                        <div className="flex items-center gap-1.5">
                                            <span className={`w-4 h-4 flex items-center justify-center ${isCompleted ? 'text-primary-500' : isFailed ? 'text-accent-red-deep' : isActive ? 'text-blue-600' : 'text-neutral-400'}`}>
                                                {isCompleted ? <Check size={12} strokeWidth={3} /> : isFailed ? <X size={12} strokeWidth={3} /> : isActive ? <RefreshCw size={10} className="animate-spin" /> : <Circle size={4} fill="currentColor" />}
                                            </span>
                                            <span className="text-xs font-bold capitalize text-neutral-800">{name}</span>
                                        </div>
                                        <span className="text-[10px] font-mono font-bold text-neutral-500">
                                            {formatDuration(step?.duration_ms)}
                                        </span>
                                    </div>
                                )
                            })}
                        </div>
                    ) : (
                        <div className="px-2 py-2 text-[10px] font-bold text-neutral-500 flex items-center gap-1">
                            <RefreshCw size={10} className="animate-spin" />
                            Cargando...
                        </div>
                    )}
                </div>,
                document.body
            )}
        </>
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
                                    <div className="flex flex-col gap-1">
                                        <ProgressDots doc={document} />
                                        {document.status === "failed" && (
                                            <div className="flex items-center gap-1 text-accent-red-deep font-bold text-xs mt-1">
                                                <AlertCircle size={12} />
                                                <span className="break-all">{document.error_message || "Error de procesamiento"}</span>
                                            </div>
                                        )}
                                    </div>
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
