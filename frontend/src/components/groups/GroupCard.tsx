import { Copy, QrCode, FileText, MessageSquare, BarChart2, Trash2 } from "lucide-react"
import type { Group } from "../../types"

interface GroupCardProps {
    group: Group
    onToggle: (id: number) => void
    onDelete: (id: number) => void
}

export function GroupCard({ group, onToggle, onDelete }: GroupCardProps) {
    return (
        <div
            className="group bg-white border-2 border-neutral-950 rounded shadow-hard-lg overflow-hidden flex flex-col h-full hover:border-primary-500 hover:-translate-y-1 hover:-translate-x-1 hover:shadow-[8px_8px_0_0_#0a0a0d] cursor-pointer transition-all duration-300"
            onClick={() => console.log(`Clicked group ${group.id} (${group.name})`)}
        >
            {/* Header */}
            <div className="p-6 border-b-2 border-neutral-950 group-hover:border-primary-500 bg-neutral-50 flex justify-between items-start transition-colors">
                <div>
                    <h3 className="text-xl font-black text-neutral-950 break-all">{group.name}</h3>
                    <p className="text-xs font-bold text-neutral-500 uppercase tracking-wider mt-0.5">{group.slug}</p>
                </div>
                <span className={`px-3 py-1 border-2 border-neutral-950 rounded-md text-xs font-bold uppercase shadow-hard-sm ${group.is_active ? 'bg-primary-400' : 'bg-neutral-200'}`}>
                    {group.is_active ? "Activo" : "Inactivo"}
                </span>
            </div>

            {/* Stats */}
            <div className="p-6 flex-1 flex flex-col gap-4">
                <div className="grid grid-cols-2 gap-4">
                    <div className="p-3 bg-neutral-50 border-2 border-neutral-950 rounded shadow-hard-sm">
                        <div className="flex items-center gap-1.5 mb-1">
                            <FileText size={14} className="text-neutral-600" />
                            <span className="text-xs font-bold text-neutral-600 uppercase tracking-wider">Documentos</span>
                        </div>
                        <p className="text-2xl font-black text-neutral-950">{group.document_ids.length}</p>
                    </div>
                    <div className="p-3 bg-neutral-50 border-2 border-neutral-950 rounded shadow-hard-sm">
                        <div className="flex items-center gap-1.5 mb-1">
                            <MessageSquare size={14} className="text-neutral-600" />
                            <span className="text-xs font-bold text-neutral-600 uppercase tracking-wider">Mensajes</span>
                        </div>
                        <div className="flex items-baseline gap-1">
                            <span className="text-2xl font-black text-neutral-950">{group.chat_quota_used}</span>
                            <span className="text-xs font-bold text-neutral-500">/ {group.chat_quota}</span>
                        </div>
                    </div>
                </div>

                <div className="flex items-center justify-between p-3 bg-white border-2 border-neutral-950 rounded shadow-hard-sm">
                    <div className="flex items-center gap-1.5">
                        <BarChart2 size={16} className="text-neutral-600" />
                        <span className="text-xs font-bold text-neutral-600 uppercase tracking-wider">Interacciones</span>
                    </div>
                    <p className="text-lg font-black text-neutral-950">{group.chat_attempts} intentos</p>
                </div>
            </div>

            {/* Actions */}
            <div className="px-6 py-4 border-t-2 border-neutral-950 bg-neutral-50 flex items-center justify-between mt-auto">
                <div className="flex items-center gap-2">

                    <button onClick={(e) => e.stopPropagation()} className="p-2 border-2 border-neutral-950 rounded bg-white hover:bg-neutral-100 hover:shadow-hard-sm transition-all cursor-pointer" title="Compartir QR">
                        <QrCode size={18} />
                    </button>
                    <button onClick={(e) => e.stopPropagation()} className="p-2 border-2 border-neutral-950 rounded bg-white hover:bg-neutral-100 hover:shadow-hard-sm transition-all cursor-pointer" title="Copiar enlace">
                        <Copy size={18} />
                    </button>
                    <button
                        onClick={(e) => { e.stopPropagation(); onToggle(group.id); }}
                        className={`w-14 h-[34px] flex items-center border-2 border-neutral-950 rounded p-0.5 transition-colors cursor-pointer ${group.is_active ? 'bg-primary-400' : 'bg-neutral-200'}`}
                        title={group.is_active ? "Desactivar" : "Activar"}
                    >
                        <div
                            className={`w-5 h-[26px] bg-white border-2 border-neutral-950 rounded-sm shadow-hard-sm transition-transform duration-200 ${group.is_active ? 'translate-x-8' : 'translate-x-0'}`}
                        />
                    </button>
                </div>

                <button
                    onClick={(e) => { e.stopPropagation(); onDelete(group.id); }}
                    className="p-2 border-2 border-transparent hover:border-neutral-950 hover:bg-accent-red/10 rounded text-neutral-600 hover:text-accent-red-deep hover:shadow-hard-sm transition-all cursor-pointer"
                    title="Eliminar grupo"
                >
                    <Trash2 size={18} />
                </button>
            </div>
        </div>
    )
}