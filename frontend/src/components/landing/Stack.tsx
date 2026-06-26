import { Zap, ShieldCheck, Bot, Link2 } from "lucide-react"
import { Pill } from "./Pill"

const features = [
  { icon: Zap, title: "Indexación instantánea", body: "Embeddings en segundos al subir tus archivos." },
  { icon: ShieldCheck, title: "Aislamiento por grupo", body: "Cada link consulta solo los documentos de su grupo." },
  { icon: Bot, title: "Respuestas con cita", body: "El modelo responde con el contexto recuperado." },
  { icon: Link2, title: "Comparte sin fricción", body: "Un enlace público, sin registro para el usuario final." },
]

export function Stack() {
  return (
    <section id="stack" className="bg-smoke mx-3 sm:mx-6 md:mx-8 mt-8 sm:mt-10 border border-ink rounded-lg">
      <div className="mx-auto max-w-[1312px] px-4 sm:px-6 md:px-10 py-12 sm:py-16 md:py-20">
        <div className="flex flex-col lg:flex-row lg:items-end justify-between gap-6">
          <div>
            <Pill tone="lilac">⚙︎ Bajo el capó</Pill>
            <h2 className="mt-5 text-3xl sm:text-4xl md:text-5xl tracking-[-0.02em]">
              Construido con Go + React.
            </h2>
          </div>
          <p className="max-w-md text-black/70">
            Backend en Go para procesamiento veloz, frontend en React para una UI
            limpia. Open patterns, código mantenible.
          </p>
        </div>

        <div className="mt-12 grid sm:grid-cols-2 lg:grid-cols-4 gap-5">
          {features.map((f) => (
            <div
              key={f.title}
              className="press-brutal bg-white border border-ink rounded-md shadow-brutal-lg p-6"
            >
              <span className="grid place-items-center w-11 h-11 bg-lime border border-ink rounded-sm shadow-brutal-sm">
                <f.icon className="w-5 h-5" />
              </span>
              <h3 className="mt-5 text-xl tracking-[-0.02em]">{f.title}</h3>
              <p className="mt-2 text-sm text-black/70 leading-6">{f.body}</p>
            </div>
          ))}
        </div>
      </div>
    </section>
  )
}
