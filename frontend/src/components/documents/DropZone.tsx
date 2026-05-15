import { useCallback } from "react"
import { useDropzone } from "react-dropzone"
import { Upload, Info } from "lucide-react"

interface DropZoneProps {
    onUpload: (file: File) => void
    disabled?: boolean
}

export function DropZone({ onUpload, disabled }: DropZoneProps) {
    const onDrop = useCallback(
        (acceptedFiles: File[]) => {
            acceptedFiles.forEach((file) => onUpload(file))
        },
        [onUpload]
    )

    const { getRootProps, getInputProps, isDragActive } = useDropzone({
        onDrop,
        disabled,
        multiple: true,
    })

    return (
        <div className="bg-white border-2 border-neutral-950 rounded-card p-4 shadow-hard-md relative overflow-hidden w-full">
            {/* Decoración de fondo */}
            <div className="absolute -top-12 -right-12 w-48 h-48 bg-primary-400/20 rounded-full blur-3xl"></div>
            <div className="absolute -bottom-8 -left-8 w-32 h-32 bg-accent-purple/40 rounded-full blur-2xl"></div>
            
            <div
                {...getRootProps()}
                className={`relative border-2 border-dashed rounded-xl p-6 flex flex-col items-center justify-center text-center transition-colors cursor-pointer group
                ${isDragActive ? "border-primary-400 bg-primary-50/50" : "border-neutral-950 bg-neutral-50/30 hover:bg-neutral-50/80"}
                ${disabled ? "opacity-50 cursor-not-allowed" : ""}
                `}
            >
                <input {...getInputProps()} />
                
                <div className="w-12 h-12 bg-primary-400 border-2 border-neutral-950 rounded-full flex items-center justify-center mb-3 shadow-hard-md group-hover:scale-110 transition-transform">
                    <Upload size={24} className="text-black" />
                </div>
                
                <h3 className="text-xl font-bold mb-1 text-neutral-950 tracking-tighter">
                    {isDragActive ? "Suelta los archivos aquí" : "Arrastra tus archivos aquí o haz click"}
                </h3>
                
                <p className="text-neutral-600 font-medium mb-4 text-sm">
                    Formatos soportados: <span className="text-neutral-950 font-bold">.txt, .pdf, .csv, .json, .docx, .xlsx, .md</span>
                </p>
                
                <div className="flex items-center gap-2 px-3 py-1 bg-accent-amber border border-neutral-950 rounded-md text-xs font-bold shadow-hard-sm">
                    <Info size={14} className="text-black" />
                    <span>Tamaño máximo: 50MB</span>
                </div>
            </div>
        </div>
    )
}