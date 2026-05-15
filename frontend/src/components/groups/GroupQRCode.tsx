import { useEffect, useRef } from "react"
import QRCode from "qrcode"
import { Button } from "../ui"
import { Download, Copy } from "lucide-react"

interface GroupQRCodeProps {
    slug?: string
}

export function GroupQRCode({ slug }: GroupQRCodeProps) {
    const canvasRef = useRef<HTMLCanvasElement>(null)

    const shareableLink = slug ? `${window.location.origin}/c/${slug}` : ""

    useEffect(() => {
        if (!slug || !canvasRef.current) return
        QRCode.toCanvas(canvasRef.current, shareableLink, { width: 200, margin: 1 })
    }, [slug, shareableLink])

    const downloadQR = () => {
        if (!canvasRef.current) return
        canvasRef.current.toBlob((blob) => {
            if (!blob) return
            const url = URL.createObjectURL(blob)
            const a = document.createElement("a")
            a.href = url
            a.download = `qr-${slug}.png`
            a.click()
            URL.revokeObjectURL(url)
        })
    }

    if (!slug) {
        return (
            <div className="flex flex-col gap-3 border-2 border-neutral-950 p-6 bg-neutral-50 rounded shadow-hard-md">
                <h4 className="text-lg font-bold text-neutral-950">Código QR</h4>
                <p className="text-sm text-neutral-400">No disponible</p>
            </div>
        )
    }

    return (
        <div className="flex flex-col gap-3 border-2 border-neutral-950 p-6 bg-neutral-50 rounded shadow-hard-md">
            <h4 className="text-lg font-bold text-neutral-950">Código QR</h4>
            <div className="flex flex-col gap-3 items-center">
                <canvas ref={canvasRef} className="border-2 border-neutral-950 rounded w-full" />
                <div className="flex gap-2 w-full">
                    <Button variant="secondary" className="flex-1" onClick={downloadQR}>
                        <Download size={16} /> Descargar
                    </Button>
                    {/*  */}
                </div>
            </div>
        </div>
    )
}
