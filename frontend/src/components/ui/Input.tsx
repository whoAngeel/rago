import type { Ref } from "react"

interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
    ref?: Ref<HTMLInputElement>
    className?: string
    label?: string
    error?: string
}

export const Input = ({
    label,
    error,
    ref,
    className = "",
    required,
    ...rest                     // onChange, onBlur, name, type, placeholder, etc.
}: InputProps) => {

    return <div className="flex flex-col gap-1">
        {
            label && (
                <label className="text-sm font-medium text-neutral-900">
                    {label}: {required && <span className="text-accent-red-deep text-sm font-bold">*</span>}
                </label>
            )
        }
        <input ref={ref}
            className={`
            bg-neutral-0
            border border-neutral-500
            rounded-btn
            px-3 py-3 
            text-base font-medium text-neutral-900
            placeholder:text-neutral-400
            focus:border-neutral-900 focus:outline-none
            ${error ? "border-accent-red-deep" : ""}
            ${className}
            
        `} {...rest} />

        {error && <p className="text-accent-red-deep text-sm">{error}</p>}
    </div>

}