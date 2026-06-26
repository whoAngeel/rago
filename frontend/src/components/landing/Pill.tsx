import type { ReactNode } from "react"

interface PillProps {
  children: ReactNode
  tone?: "white" | "pink" | "lime" | "lilac"
}

const toneBg: Record<string, string> = {
  pink: "bg-pink-soft",
  lime: "bg-lime-soft",
  lilac: "bg-lilac",
  white: "bg-white",
}

export function Pill({ children, tone = "white" }: PillProps) {
  return (
    <span className={`inline-flex items-center gap-2 ${toneBg[tone]} border border-ink rounded-pill px-3.5 py-1.5 text-sm font-medium text-black`}>
      {children}
    </span>
  )
}
