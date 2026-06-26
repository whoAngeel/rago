import { ArrowRight, Bot, MessageSquareText } from "lucide-react"
import { Pill } from "./Pill"
import { BtnPrimary } from "./BtnPrimary"
import { BtnSecondary } from "./BtnSecondary"
import { FloatingDecor } from "./FloatingDecor"

export function Hero() {
  return (
    <section className="relative overflow-hidden bg-pink mx-3 sm:mx-6 md:mx-8 mt-4 sm:mt-6 border border-ink rounded-lg">
      <FloatingDecor />
      <div className="relative mx-auto max-w-[1312px] px-4 sm:px-6 md:px-10 py-12 sm:py-16 md:py-24 grid lg:grid-cols-2 gap-8 lg:gap-12 items-center">
        <div>
          <Pill tone="white">
            <span className="w-2 h-2 rounded-full bg-black" />
            Go + React · Open source
          </Pill>
          <h1 className="mt-6 text-4xl sm:text-5xl md:text-6xl lg:text-[64px] leading-[1.05] tracking-[-0.02em] font-medium">
            Porque ya nadie lee
            <br /> las instrucciones.
          </h1>
          <p className="mt-6 max-w-xl text-lg leading-7 text-black/80">
            Transforma tus aburridos manuales y PDFs en asistentes de IA
            públicos e interactivos. Agrupa tus documentos, obtén un enlace y
            deja que tus usuarios simplemente <strong>pregunten</strong>.
          </p>
          <div className="mt-8 flex flex-wrap gap-4">
            <BtnPrimary href="https://rago.whoangel.work/dashboard" external>
              Probar el Dashboard <ArrowRight className="w-4 h-4" />
            </BtnPrimary>
            <BtnSecondary href="#como-funciona">Ver cómo funciona</BtnSecondary>
          </div>
          <div className="mt-10 flex flex-wrap gap-2">
            <Pill tone="lime">⚡ Vectores al instante</Pill>
            <Pill tone="lilac">🔗 Enlace público</Pill>
            <Pill tone="pink">🧠 RAG end-to-end</Pill>
          </div>
        </div>

        <div className="relative hidden lg:block">
          <div className="absolute inset-0 translate-x-2 translate-y-2 bg-mint border-2 border-[color:var(--shadow-ink)] rounded-xl" />
          <div className="relative bg-white border-2 border-[color:var(--shadow-ink)] rounded-xl shadow-brutal-lg p-6">
            <div className="flex items-center gap-3 pb-4 border-b border-ink/10">
              <span className="grid place-items-center w-11 h-11 bg-lime border border-ink rounded-sm">
                <Bot className="w-5 h-5" />
              </span>
              <div>
                <div className="font-medium">Bot de Manuales</div>
                <div className="text-sm text-black/60">Pregúntame sobre la Cafetera V2</div>
              </div>
            </div>

            <div className="mt-5 flex justify-end">
              <div className="max-w-[85%] bg-smoke border border-ink rounded-md px-4 py-3 text-[15px]">
                ¿Cómo descalcifico la máquina? La luz roja parpadea.
              </div>
            </div>

            <div className="mt-3 flex">
              <div className="max-w-[92%] bg-lime-soft border border-ink rounded-md px-4 py-3 text-[15px] leading-6">
                <strong>¡La luz roja significa descalcificar!</strong> Mezcla
                500 ml de agua con la solución, llena el tanque y mantén
                presionado el botón por 5 segundos para iniciar el ciclo.
              </div>
            </div>

            <div className="mt-5 flex items-center gap-2 border border-ink rounded-sm px-3 py-2.5">
              <MessageSquareText className="w-4 h-4 text-black/60" />
              <input
                className="flex-1 bg-transparent outline-none text-sm placeholder:text-black/40"
                placeholder="Escribe una pregunta…"
                readOnly
              />
              <button className="bg-lime border border-ink rounded-sm px-3 py-1 text-sm font-medium shadow-brutal-sm">
                Enviar
              </button>
            </div>
          </div>
        </div>
      </div>

      <svg
        viewBox="0 0 1440 60"
        className="absolute bottom-0 left-0 w-full text-lime"
        preserveAspectRatio="none"
      >
        <path
          fill="currentColor"
          d="M0,30 C240,60 480,0 720,20 C960,40 1200,60 1440,20 L1440,60 L0,60 Z"
        />
      </svg>
    </section>
  )
}
