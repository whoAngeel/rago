import { createFileRoute } from "@tanstack/react-router"
import { Nav, Hero, ComoFunciona, Painpoints, Stack, CtaBanner, Footer } from "../components/landing"

export const Route = createFileRoute("/")({
  component: Landing,
})

function Landing() {
  return (
    <div className="min-h-screen bg-white text-black overflow-x-hidden">
      <Nav />
      <main>
        <Hero />
        <ComoFunciona />
        <Painpoints />
        <Stack />
        <CtaBanner />
      </main>
      <Footer />
    </div>
  )
}
