import type { ReactNode, Ref } from "react"

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
    variant?: 'primary' | 'secondary'
    size?: 'default' | 'lg'
    isLoading?: boolean
    children: ReactNode
    ref?: Ref<HTMLButtonElement>
}

const variantStyles = {
    primary: "bg-primary-400 text-neutral-900",
    secondary: "bg-neutral-0 text-neutral-900"
} as const

const sizeStyles = {
    default: "text-btn px-4 py-2 h-[42px]",
    lg: "text-btn-lg px-6 py-3 h-[52px]"
} as const

export const Button = ({
    variant = "primary",
    size = "default",
    isLoading = false,
    children,
    disabled,
    ref,
    className = "",
    ...rest

}: ButtonProps) => {
    return <>
        <button
            ref={ref}
            disabled={disabled || isLoading}
            className={`
            inline-flex items-center gap-2 border-2 border-neutral-900 rounded-b-btn 
            shadow-hard-md shadow-hover font-medium 
            cursor-pointer 
            transition-all
            ${variantStyles[variant]}
            ${sizeStyles[size]}
            ${disabled || isLoading ? 'opacity-50 cursor-not-allowed' : ''}
            ${className}
            `}
            {...rest}
        >
            {isLoading ? "Cargando..." : children}

        </button>
    </>
}