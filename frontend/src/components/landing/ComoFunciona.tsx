import { Sparkles, FileText, FolderPlus, Link2 } from "lucide-react"
import { Pill } from "./Pill"

const steps = [
  {
    n: "01",
    title: "Procesa manuales",
    icon: FileText,
    body: "Sube PDFs o archivos de texto. Rago los divide y los convierte en vectores al instante.",
  },
  {
    n: "02",
    title: "Crea grupos",
    icon: FolderPlus,
    body: "Organiza tus documentos en grupos para diferentes productos, temas o audiencias.",
  },
  {
    n: "03",
    title: "Comparte el enlace",
    icon: Link2,
    body: "Rago genera un enlace público único. Compártelo y deja que chateen con tus documentos.",
  },
]

export function ComoFunciona() {
  return (
    <section id="como-funciona" className="bg-smoke mx-3 sm:mx-6 md:mx-8 mt-8 sm:mt-10 border border-ink rounded-lg">
      <div className="mx-auto max-w-[1312px] px-4 sm:px-6 md:px-10 py-12 sm:py-16 md:py-20">
        <div className="flex flex-col lg:flex-row lg:items-end justify-between gap-6">
          <div>
            <Pill tone="lime">
              <Sparkles className="w-4 h-4" /> Cómo funciona
            </Pill>
            <h2 className="mt-5 text-3xl sm:text-4xl md:text-5xl tracking-[-0.02em]">
              Un sistema RAG completo,
              <br /> de la subida al chat.
            </h2>
          </div>
          <p className="max-w-md text-black/70 text-base leading-6">
            Patrones modernos de arquitectura, una interfaz limpia y todo el flujo
            end-to-end para que lo veas funcionar en minutos.
          </p>
        </div>

        <div className="mt-12 grid md:grid-cols-3 gap-6">
          {steps.map((s) => (
            <div
              key={s.n}
              className="press-brutal bg-white border-2 border-[color:var(--shadow-ink)] rounded-xl shadow-brutal-lg p-5 sm:p-7"
            >
              <div className="flex items-center justify-between">
                <span className="text-2xl font-bold">{s.n}</span>
                <span className="grid place-items-center w-11 h-11 bg-lime border border-ink rounded-sm shadow-brutal-sm">
                  <s.icon className="w-5 h-5" />
                </span>
              </div>
              <div className="my-5 border-t border-dashed border-ink/40" />
              <h3 className="text-2xl tracking-[-0.02em]">{s.title}</h3>
              <p className="mt-3 text-black/70 leading-6">{s.body}</p>
            </div>
          ))}
        </div>
      </div>
    </section>
  )
}
