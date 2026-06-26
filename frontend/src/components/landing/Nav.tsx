import { Bot, ArrowRight } from "lucide-react"

export function Nav() {
  return (
    <header className="w-full px-3 sm:px-6 md:px-8 pt-4 sm:pt-6">
      <nav className="mx-auto max-w-[1376px] bg-white border border-ink rounded-pill shadow-brutal flex items-center justify-between px-4 sm:px-6 md:px-12 h-[60px] sm:h-[68px]">
        <a href="#" className="flex items-center gap-2 sm:gap-2.5 shrink-0">
          <span className="grid place-items-center w-8 h-8 sm:w-9 sm:h-9 bg-lime border border-ink rounded-sm shadow-brutal-sm">
            <Bot className="w-4 h-4 sm:w-5 sm:h-5" strokeWidth={2.25} />
          </span>
          <span className="text-lg sm:text-xl font-medium tracking-tight">Rago</span>
        </a>
        <div className="hidden md:flex items-center gap-8 text-base">
          <a href="#como-funciona" className="hover:underline underline-offset-4">Cómo Funciona</a>
          <a href="#porque" className="hover:underline underline-offset-4">Por qué Rago</a>
          <a href="#stack" className="hover:underline underline-offset-4">Stack</a>
        </div>
        <div className="flex items-center gap-2 sm:gap-3">
          <a
            href="https://rago.whoangel.work/dashboard"
            className="press-brutal hidden sm:inline-flex items-center gap-2 bg-white border border-ink rounded-sm shadow-brutal px-3 sm:px-4 py-2 text-xs sm:text-sm font-medium"
          >
            Sign in
          </a>
          <a
            href="https://rago.whoangel.work/dashboard"
            className="press-brutal inline-flex items-center gap-1.5 sm:gap-2 bg-lime border border-ink rounded-sm shadow-brutal px-3 sm:px-4 py-2 text-xs sm:text-sm font-medium"
          >
            Probar ahora <ArrowRight className="w-3.5 h-3.5 sm:w-4 sm:h-4" />
          </a>
        </div>
      </nav>
    </header>
  )
}
