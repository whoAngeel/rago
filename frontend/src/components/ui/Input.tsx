import type { Ref, ReactNode } from 'react'

interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
    ref?: Ref<HTMLInputElement>
    className?: string
    label?: string
    error?: string
    rightElement?: ReactNode
}

export const Input = ({
    label,
    error,
    ref,
    className = '',
    required,
    rightElement,
    ...rest
}: InputProps) => {
    return <div className="flex flex-col gap-1.5">
        {label && (
            <label className="text-xs font-bold text-neutral-600 uppercase tracking-wider">
                {label} {required && <span className="text-accent-red-deep text-sm font-bold">*</span>}
            </label>
        )}
        <div className="relative">
            <input
                ref={ref}
                className={`
                    bg-white
                    border-2 border-neutral-950
                    rounded
                    px-3 py-3
                    w-full
                    text-base font-medium text-neutral-900
                    placeholder:text-neutral-400
                    focus:outline-none focus:border-primary-500 focus:shadow-hard-lg
                    transition-all
                    ${rightElement ? 'pr-10' : ''}
                    ${error ? 'border-accent-red-deep' : ''}
                    ${className}
                `}
                {...rest}
            />
            {rightElement && (
                <div className="absolute right-3 top-1/2 -translate-y-1/2">
                    {rightElement}
                </div>
            )}
        </div>
        {error && <p className="text-accent-red-deep text-sm font-bold mt-0.5">{error}</p>}
    </div>
}