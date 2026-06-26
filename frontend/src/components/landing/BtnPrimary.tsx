import type { ReactNode } from "react"
import { Link } from "@tanstack/react-router"

interface BtnPrimaryProps {
  children: ReactNode
  href: string
  external?: boolean
}

export function BtnPrimary({ children, href, external }: BtnPrimaryProps) {
  const cls = "press-brutal inline-flex items-center gap-2 bg-lime text-black border border-ink rounded-sm shadow-brutal px-6 py-3 text-base font-medium"
  if (external) {
    return <a href={href} className={cls}>{children}</a>
  }
  return <Link to={href} className={cls}>{children}</Link>
}
