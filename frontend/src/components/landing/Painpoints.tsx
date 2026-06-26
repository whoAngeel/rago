import { Pill } from "./Pill"

const items = [
  "Manuales PDF que nadie abre",
  "Soporte respondiendo lo mismo 100 veces",
  "Documentación dispersa en mil enlaces",
  "Usuarios frustrados buscando en un buscador roto",
]

export function Painpoints() {
  return (
    <section id="porque" className="mx-3 sm:mx-6 md:mx-8 mt-8 sm:mt-10 bg-white border border-ink rounded-lg">
      <div className="mx-auto max-w-[1312px] px-4 sm:px-6 md:px-10 py-12 sm:py-16 md:py-20 grid lg:grid-cols-2 gap-8 lg:gap-14 items-center">
        <div>
          <Pill tone="pink">
            ✦ Stop losing users
          </Pill>
          <h2 className="mt-5 text-3xl sm:text-4xl md:text-5xl tracking-[-0.02em]">
            La documentación no se lee.
            <br /> Pero sí se pregunta.
          </h2>
          <p className="mt-5 text-black/70 leading-6 max-w-md">
            Rago convierte cualquier base de conocimiento estática en una
            conversación: usuarios contentos, soporte aliviado.
          </p>
        </div>

        <div className="space-y-5">
          {items.map((t, i) => (
            <div
              key={i}
              className="relative ml-10 bg-white border border-ink rounded-md shadow-brutal-lg pl-10 pr-5 py-4 flex items-center gap-3"
            >
              <span className="absolute -left-6 top-1/2 -translate-y-1/2 grid place-items-center w-12 h-12 bg-destructive border border-ink rounded-full shadow-brutal-sm text-white font-bold">
                ✕
              </span>
              <span className="text-base sm:text-lg">{t}</span>
            </div>
          ))}
        </div>
      </div>
    </section>
  )
}
