import { Bot, ExternalLink } from "lucide-react"

export function Footer() {
  return (
    <footer className="mx-3 sm:mx-6 md:mx-8 mt-8 sm:mt-10 mb-6 sm:mb-8 bg-gradient-footer border border-ink rounded-lg overflow-hidden">
      <div className="mx-auto max-w-[1312px] px-4 sm:px-6 md:px-10 py-12 sm:py-16">
        <div className="flex flex-col md:flex-row gap-10 justify-between">
          <div className="max-w-sm">
            <div className="flex items-center gap-2.5">
              <span className="grid place-items-center w-10 h-10 bg-lime border border-ink rounded-sm shadow-brutal-sm">
                <Bot className="w-5 h-5" />
              </span>
              <span className="text-2xl font-medium tracking-tight">Rago</span>
            </div>
            <p className="mt-4 text-black/70 leading-6">
              Sistema RAG end-to-end. Convierte tus documentos en chats.
            </p>
            <div className="mt-5 inline-flex items-center gap-2 bg-white border border-ink rounded-sm shadow-brutal-sm px-4 py-2 text-sm">
              <span className="w-2 h-2 rounded-full bg-lime border border-ink" />
              All systems operational
            </div>
          </div>

          <div className="grid grid-cols-2 sm:grid-cols-3 gap-8 text-sm">
            <div>
              <div className="font-bold mb-3">Producto</div>
              <ul className="space-y-2">
                <li><a className="hover:underline" href="#como-funciona">Cómo funciona</a></li>
                <li><a className="hover:underline" href="#stack">Stack</a></li>
                <li><a className="hover:underline" href="https://rago.whoangel.work/dashboard">Dashboard</a></li>
              </ul>
            </div>
            <div>
              <div className="font-bold mb-3">Recursos</div>
              <ul className="space-y-2">
                <li><a className="hover:underline" href="#">Docs</a></li>
                <li><a className="hover:underline" href="#">Ejemplos</a></li>
              </ul>
            </div>
            <div>
              <div className="font-bold mb-3">Comunidad</div>
              <ul className="space-y-2">
                <li>
                  <a className="hover:underline inline-flex items-center gap-1.5" href="#">
                    <ExternalLink className="w-4 h-4" /> GitHub
                  </a>
                </li>
              </ul>
            </div>
          </div>
        </div>

        <div className="mt-12 pt-6 border-t border-ink/20 flex flex-col items-center sm:flex-row sm:justify-between gap-3 text-sm text-center sm:text-left">
          <span>© {new Date().getFullYear()} Rago. Built with Go + React.</span>
          <span className="text-black/70">Hecho con cariño para developers cansados de PDFs.</span>
        </div>
      </div>
    </footer>
  )
}
