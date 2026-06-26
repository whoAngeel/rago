import type { ReactNode } from "react"

interface BtnSecondaryProps {
  children: ReactNode
  href: string
}

export function BtnSecondary({ children, href }: BtnSecondaryProps) {
  return (
    <a href={href} className="press-brutal inline-flex items-center gap-2 bg-white text-black border border-ink rounded-sm shadow-brutal px-6 py-3 text-base font-medium">
      {children}
    </a>
  )
}
