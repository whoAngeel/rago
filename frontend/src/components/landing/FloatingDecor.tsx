import { Sparkles } from "lucide-react"

export function FloatingDecor() {
  return (
    <>
      <div className="absolute top-16 left-10 w-3 h-3 rounded-full bg-black" />
      <div className="absolute top-40 left-24 w-2 h-2 rounded-full bg-black" />
      <div className="absolute bottom-24 left-16 w-4 h-4 rounded-full bg-black" />
      <div className="absolute top-24 right-32 w-2 h-2 rounded-full bg-black" />
      <div className="absolute bottom-40 right-10 w-3 h-3 rounded-full bg-black" />
      <Sparkles className="absolute top-12 right-1/3 w-6 h-6 text-black" />
      <Sparkles className="absolute bottom-20 left-1/3 w-5 h-5 text-black" />
    </>
  )
}
