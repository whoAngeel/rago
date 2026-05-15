import { useState } from 'react'
import { Eye, EyeOff, Check, X } from 'lucide-react'
import type { UseFormReturn } from 'react-hook-form'
import type { RegisterForm } from '../../../lib/validations'
import { Input } from '../Input'
import { Button } from '../Button'

interface RegisterFormProps {
    form: UseFormReturn<RegisterForm>
    onSubmit: (data: RegisterForm) => void | Promise<void>
    isLoading: boolean
}

export function RegisterForm({
    form, onSubmit, isLoading
}: RegisterFormProps) {
    const [showPassword, setShowPassword] = useState(false)
    const [showConfirm, setShowConfirm] = useState(false)
    const { errors } = form.formState
    const password = form.watch('password')
    const confirm = form.watch('password_confirmation')
    const passwordsMatch = password && confirm && password === confirm
    const hasBoth = password && confirm

    return <form onSubmit={form.handleSubmit(onSubmit)} className="flex flex-col gap-5 w-full">
        <div className="flex flex-col gap-1">
            <Input
                label="Nombre"
                error={errors.name?.message}
                {...form.register('name')}
                className="bg-white border border-neutral-500 focus:border-neutral-900 rounded px-3 py-3 text-base font-medium transition-colors outline-none w-full"
            />
        </div>
        <div className="flex flex-col gap-1">
            <Input
                label="Email"
                type="email"
                error={errors.email?.message}
                {...form.register('email')}
                className="bg-white border border-neutral-500 focus:border-neutral-900 rounded px-3 py-3 text-base font-medium transition-colors outline-none w-full"
            />
        </div>
        <div className="flex flex-col gap-1">
            <Input
                label="Password"
                type={showPassword ? 'text' : 'password'}
                error={errors.password?.message}
                {...form.register('password')}
                className="bg-white border border-neutral-500 focus:border-neutral-900 rounded px-3 py-3 text-base font-medium transition-colors outline-none w-full"
                rightElement={
                    <button
                        type="button"
                        onClick={() => setShowPassword(!showPassword)}
                        className="text-neutral-500 hover:text-neutral-900 transition-colors"
                        tabIndex={-1}
                    >
                        {showPassword ? <EyeOff size={20} /> : <Eye size={20} />}
                    </button>
                }
            />
        </div>
        <div className="flex flex-col gap-1">
            <Input
                label="Confirm Password"
                type={showConfirm ? 'text' : 'password'}
                error={errors.password_confirmation?.message}
                {...form.register('password_confirmation')}
                className="bg-white border border-neutral-500 focus:border-neutral-900 rounded px-3 py-3 text-base font-medium transition-colors outline-none w-full"
                rightElement={
                    <div className="flex items-center gap-1">
                        {hasBoth && (
                            passwordsMatch
                                ? <Check size={18} className="text-green-600" />
                                : <X size={18} className="text-accent-red-deep" />
                        )}
                        <button
                            type="button"
                            onClick={() => setShowConfirm(!showConfirm)}
                            className="text-neutral-500 hover:text-neutral-900 transition-colors"
                            tabIndex={-1}
                        >
                            {showConfirm ? <EyeOff size={20} /> : <Eye size={20} />}
                        </button>
                    </div>
                }
            />
        </div>
        <Button
            type="submit"
            isLoading={isLoading}
            className="w-full mt-2 bg-primary-400 hover:bg-primary-300 text-black border border-neutral-900 rounded shadow-hard-md hover:translate-x-0.5 hover:translate-y-0.5 hover:shadow-none transition-all duration-150 font-medium py-3 text-base flex items-center justify-center gap-2"
        >
            Registrar
        </Button>
    </form>
}