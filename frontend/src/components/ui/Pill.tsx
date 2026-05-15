import type { ReactNode } from "react"

interface PillProps {
    children: ReactNode
    size?: "default" | "sm"
    className?: string
}

const sizeStyles = {
    default: "px-3.5 py-1.5 h-[38px] text-btn",    // 14px 6px 38px
    sm: "px-3.5 py-1 h-[34px] text-btn",           // 14px 4px 34px
} as const

export const Pill = ({
    children,
    size = "default",
    className = "",
}: PillProps) => {
    return (
        <span className={`
      inline-flex items-center gap-2
      bg-neutral-0 border border-neutral-900
      rounded-pill font-medium
      ${sizeStyles[size]}
      ${className}
    `}>
            {children}
        </span>
    )
}