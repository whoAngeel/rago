import { useState } from "react"
import { X } from "lucide-react"
import { Button } from "../ui"

interface CreateGroupModalProps {
    isOpen: boolean
    onClose: () => void
    onCreate: (name: string) => void
    isPending?: boolean
}

export const CreateGroupModal = ({ isOpen, onClose, onCreate, isPending }: CreateGroupModalProps) => {
    const [name, setName] = useState("")

    if (!isOpen) return null

    const handleSubmit = (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault()
        if (name.trim()) {
            onCreate(name.trim())
            setName("")
            onClose()
        }
    }

    return (
        <div className="fixed inset-0 bg-neutral-950/50 flex items-center justify-center z-50 p-4">
            <div className="bg-white border-2 border-neutral-950 rounded-card shadow-hard-lg max-w-md w-full overflow-hidden">
                <div className="p-6 border-b-2 border-neutral-950 bg-neutral-50 flex justify-between items-center">
                    <h2 className="text-xl font-black text-neutral-950">Crear Nuevo Grupo</h2>
                    <button onClick={onClose}
                        className="p-1 border-2 border-transparent hover:border-neutral-950 rounded hover:bg-neutral-100 transition-all cursor-pointer"
                    ><X size={20} /></button>
                </div>

                <form onSubmit={handleSubmit}>
                    <div className="p-6 flex flex-col gap-4">
                        <div className="flex flex-col gap-1.5">
                            <label htmlFor="group-name" className="text-xs font-bold text-neutral-600 uppercase tracking-wider">
                                Nombre del Grupo
                            </label>
                            <input id="group-name" type="text" value={name}
                                onChange={(e) => setName(e.target.value)}
                                placeholder="Ej: Equipo de Marketing"
                                className="w-full p-3 border-2 border-neutral-950 rounded font-medium focus:outline-none focus:border-primary-500 focus:shadow-[4px_4px_0_0_#0a0a0d] transition-all"
                                autoFocus
                            />
                        </div>
                    </div>

                    <div className="px-6 py-4 border-t-2 border-neutral-950 bg-neutral-50 flex justify-end gap-3">
                        <Button variant="secondary" type="button" onClick={onClose}>Cancelar</Button>
                        <Button variant="primary" type="submit" disabled={!name.trim() || isPending} isLoading={isPending}>
                            {isPending ? "Creando..." : "Crear Grupo"}
                        </Button>
                    </div>
                </form>
            </div>
        </div>
    )
}
