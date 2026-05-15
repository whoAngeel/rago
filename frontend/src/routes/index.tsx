import { createFileRoute, Link } from "@tanstack/react-router";
import { useRef } from "react";
import gsap from "gsap";
import { useGSAP } from "@gsap/react";
import {
  ArrowRight,
  Bot,
  Database,
  Link as LinkIcon,
  MessageSquare,
  ChevronRight,
  GitGraph,
} from "lucide-react";

gsap.registerPlugin(useGSAP);

export const Route = createFileRoute("/")({
  component: LandingPage,
});

function LandingPage() {
  const container = useRef<HTMLDivElement>(null);

  useGSAP(() => {
    // Hero Elements Animation
    gsap.from(".hero-elem", {
      y: 30,
      opacity: 0,
      duration: 0.8,
      stagger: 0.15,
      ease: "power3.out",
    });

    // Chat Mockup Animation
    gsap.from(".chat-mockup", {
      x: 40,
      opacity: 0,
      duration: 1,
      delay: 0.4,
      ease: "power3.out",
    });

    // Feature Cards Stagger
    gsap.from(".feature-card", {
      y: 40,
      opacity: 0,
      duration: 0.6,
      stagger: 0.15,
      ease: "back.out(1.5)",
      delay: 0.6,
    });
  }, { scope: container });

  return (
    <div ref={container} className="min-h-screen bg-white text-black selection:bg-[#bef265]">
      {/* Navigation */}
      <nav className="fixed top-0 w-full z-50 bg-white border-b-2 border-neutral-950">
        <div className="max-w-[1376px] mx-auto px-6 h-20 flex items-center justify-between">
          <div className="flex items-center gap-3 font-bold text-2xl tracking-tight">
            <div className="w-10 h-10 rounded-[var(--radius-btn)] bg-[#84cc17] border-2 border-neutral-950 flex items-center justify-center shadow-[var(--shadow-hard-sm)]">
              <Bot className="w-6 h-6 text-neutral-950" />
            </div>
            Rago
          </div>
          <div className="flex items-center gap-6">
            <Link
              to="/dashboard"
              className="font-bold bg-white text-black border-2 border-neutral-950 px-6 py-2.5 rounded-[var(--radius-btn)] shadow-[var(--shadow-hard-sm)] shadow-hover transition-colors"
            >
              Go to Dashboard
            </Link>
          </div>
        </div>
      </nav>

      {/* Hero Section */}
      <div className="relative pt-40 pb-24 overflow-hidden bg-[#fafafa] border-b-2 border-neutral-950">
        <div className="max-w-[1312px] mx-auto px-6 grid lg:grid-cols-2 gap-16 items-center">
          <div className="max-w-2xl">
            <div className="hero-elem inline-flex items-center gap-2 px-4 py-1.5 rounded-[var(--radius-pill)] bg-[#f5d1fe] border-2 border-neutral-950 text-sm font-bold mb-8 shadow-[var(--shadow-hard-sm)]">
              <span className="flex h-2.5 w-2.5 rounded-full bg-neutral-950"></span>
              Go + React
            </div>

            <h1 className="hero-elem text-5xl sm:text-7xl font-extrabold tracking-tight mb-8 leading-[var(--text-h1--line-height)]">
              Porque ya nadie lee las instrucciones.
            </h1>

            <p className="hero-elem text-xl text-neutral-700 mb-10 leading-relaxed font-medium">
              Transforma tus aburridos manuales y PDFs en asistentes de IA
              públicos e interactivos. Agrupa tus documentos, obtén un enlace y
              deja que tus usuarios simplemente pregunten.
            </p>

            <div className="hero-elem flex flex-col sm:flex-row items-center gap-4">
              <Link
                to="/dashboard"
                className="flex items-center justify-center gap-2 bg-[#84cc17] text-neutral-950 border-2 border-neutral-950 px-8 py-4 rounded-[var(--radius-btn)] font-bold text-lg shadow-[var(--shadow-hard-md)] shadow-hover w-full sm:w-auto"
              >
                Probar el Dashboard
                <ArrowRight className="w-5 h-5" />
              </Link>
              <a
                href="#features"
                className="flex items-center justify-center gap-2 bg-white text-neutral-950 border-2 border-neutral-950 px-8 py-4 rounded-[var(--radius-btn)] font-bold text-lg shadow-[var(--shadow-hard-md)] shadow-hover w-full sm:w-auto"
              >
                Ver Cómo Funciona
              </a>
            </div>
          </div>

          <div className="relative hidden lg:block chat-mockup">
            <div className="absolute inset-0 bg-[#35d399] rounded-[var(--radius-card)] translate-x-4 translate-y-4 border-2 border-neutral-950"></div>
            <div className="relative bg-white border-2 border-neutral-950 rounded-[var(--radius-card)] p-8 shadow-[var(--shadow-hard-lg)] flex flex-col h-[400px]">
              <div className="flex items-center gap-4 border-b-2 border-neutral-200 pb-4 mb-4">
                <div className="w-12 h-12 bg-[#84cc17] border-2 border-neutral-950 rounded-[var(--radius-btn)] flex items-center justify-center shadow-[var(--shadow-hard-sm)]">
                  <Bot className="w-6 h-6 text-neutral-950" />
                </div>
                <div>
                  <div className="font-bold text-lg">Bot de Manuales</div>
                  <div className="text-sm text-neutral-500 font-medium">
                    Pregúntame sobre la Cafetera V2
                  </div>
                </div>
              </div>
              <div className="flex-1 flex flex-col justify-end space-y-4">
                <div className="bg-[#f5f5f5] border-2 border-neutral-950 rounded-[var(--radius-btn)] p-4 max-w-[85%] self-end">
                  <p className="font-medium">
                    ¿Cómo descalcifico la máquina? La luz roja parpadea.
                  </p>
                </div>
                <div className="bg-[#d2fae5] border-2 border-neutral-950 rounded-[var(--radius-btn)] p-4 max-w-[90%] shadow-[var(--shadow-hard-sm)]">
                  <p className="font-medium">
                    ¡La luz roja significa que es hora de descalcificar! Mezcla
                    500ml de agua con la solución, llena el tanque y mantén
                    presionado el botón por 5 segundos para iniciar el ciclo.
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Features Section */}
      <div
        id="features"
        className="py-24 bg-white border-b-2 border-neutral-950"
      >
        <div className="max-w-[1312px] mx-auto px-6">
          <div className="text-center mb-20 max-w-2xl mx-auto">
            <h2 className="text-4xl font-bold mb-6">Cómo Funciona Rago</h2>
            <p className="text-xl text-neutral-600 font-medium">
              Un sistema RAG completo (end-to-end) construido para demostrar
              patrones de arquitectura modernos y una interfaz limpia.
            </p>
          </div>

          <div className="grid md:grid-cols-3 gap-8">
            <FeatureCard
              icon={<Database className="w-8 h-8 text-neutral-950" />}
              bgColor="bg-[#edeafe]"
              title="1. Procesa Manuales"
              description="Sube PDFs o archivos de texto. RAGO toma tus archivos, los divide y los convierte en vectores al instante."
            />
            <FeatureCard
              icon={<MessageSquare className="w-8 h-8 text-neutral-950" />}
              bgColor="bg-[#fef3c8]"
              title="2. Crea Grupos"
              description="Organiza tus documentos subidos en grupos específicos para diferentes productos o temas."
            />
            <FeatureCard
              icon={<LinkIcon className="w-8 h-8 text-neutral-950" />}
              bgColor="bg-[#dceafe]"
              title="3. Comparte el Enlace"
              description="Rago genera un enlace público único. Compártelo con tus clientes para que chateen con tus documentos."
            />
          </div>
        </div>
      </div>

      {/* CTA Section */}
      <div className="bg-gradient-cta border-b-2 border-neutral-950 py-32">
        <div className="max-w-[1312px] mx-auto px-6 text-center">
          <h2 className="text-5xl font-bold mb-8">
            ¿Listo para probar el sistema?
          </h2>
          <p className="text-xl text-neutral-800 font-medium mb-12 max-w-2xl mx-auto">
            Inicia sesión en el dashboard, sube un documento de prueba y
            experimenta el chat por ti mismo.
          </p>
          <Link
            to="/dashboard"
            className="inline-flex items-center justify-center gap-3 bg-white text-neutral-950 border-2 border-neutral-950 px-10 py-5 rounded-[var(--radius-btn)] font-bold text-xl shadow-[var(--shadow-hard-lg)] shadow-hover"
          >
            Ir al Dashboard
            <ChevronRight className="w-6 h-6" />
          </Link>
        </div>
      </div>

      {/* Footer */}
      <footer className="bg-white py-12 border-t-2 border-neutral-950 font-medium">
        <div className="max-w-[1376px] mx-auto px-6 flex flex-col md:flex-row items-center justify-between">
          <div className="flex items-center gap-3 mb-6 md:mb-0">
            <div className="w-8 h-8 rounded-[var(--radius-btn)] bg-[#84cc17] border-2 border-neutral-950 flex items-center justify-center shadow-[var(--shadow-hard-sm)]">
              <Bot className="w-4 h-4 text-neutral-950" />
            </div>
            <span className="font-bold">
              © {new Date().getFullYear()} Rago Platform.
            </span>
          </div>
          <div className="flex items-center gap-8 text-neutral-600">
            <a
              href="#"
              className="hover:text-black hover:underline transition-colors"
            >
              Documentation
            </a>
            <a
              href="#"
              className="hover:text-black hover:underline transition-colors"
            >
              Privacy
            </a>
            <a
              href="#"
              className="hover:text-black hover:underline transition-colors"
            >
              Terms
            </a>
          </div>
        </div>
      </footer>
    </div>
  );
}

function FeatureCard({
  icon,
  bgColor,
  title,
  description,
}: {
  icon: React.ReactNode;
  bgColor: string;
  title: string;
  description: string;
}) {
  return (
    <div
      className={`feature-card bg-white border-2 border-neutral-950 p-8 rounded-[var(--radius-card)] shadow-[var(--shadow-hard-lg)] shadow-hover transition-transform`}
    >
      <div
        className={`w-16 h-16 ${bgColor} border-2 border-neutral-950 rounded-[var(--radius-btn)] flex items-center justify-center mb-8 shadow-[var(--shadow-hard-sm)]`}
      >
        {icon}
      </div>
      <h3 className="text-2xl font-bold mb-4">{title}</h3>
      <p className="text-neutral-700 font-medium leading-relaxed text-lg">
        {description}
      </p>
    </div>
  );
}
