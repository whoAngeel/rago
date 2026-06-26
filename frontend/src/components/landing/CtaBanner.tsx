import { Sparkles, ArrowRight } from "lucide-react"
import { Pill } from "./Pill"
import { BtnPrimary } from "./BtnPrimary"
import { BtnSecondary } from "./BtnSecondary"

export function CtaBanner() {
  return (
    <section className="mx-3 sm:mx-6 md:mx-8 mt-8 sm:mt-10">
      <div className="relative overflow-hidden bg-gradient-cta border-2 border-[color:var(--shadow-ink)] rounded-lg shadow-brutal-lg px-4 sm:px-10 md:px-16 py-12 sm:py-16 md:py-20 text-center">
        <Sparkles className="absolute top-6 left-4 sm:top-8 sm:left-10 w-4 h-4 sm:w-6 sm:h-6" />
        <Sparkles className="absolute bottom-8 right-4 sm:bottom-10 sm:right-12 w-5 h-5 sm:w-7 sm:h-7" />
        <div className="absolute top-10 right-1/4 w-2 h-2 sm:w-3 sm:h-3 rounded-full bg-black" />
        <div className="absolute bottom-12 left-1/3 w-1.5 h-1.5 sm:w-2 sm:h-2 rounded-full bg-black" />
        <Pill tone="white">¿Te animas?</Pill>
        <h2 className="mt-5 text-3xl sm:text-4xl md:text-5xl lg:text-6xl tracking-[-0.02em]">
          ¿Listo para probar el sistema?
        </h2>
        <p className="mt-5 text-lg text-black/70 max-w-2xl mx-auto">
          Inicia sesión en el dashboard, sube un documento de prueba y experimenta
          el chat por ti mismo.
        </p>
        <div className="mt-8 flex flex-wrap justify-center gap-4">
          <BtnPrimary href="https://rago.whoangel.work/dashboard" external>
            Ir al Dashboard <ArrowRight className="w-4 h-4" />
          </BtnPrimary>
          <BtnSecondary href="#como-funciona">Ver cómo funciona</BtnSecondary>
        </div>
      </div>
    </section>
  )
}
