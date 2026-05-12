import { useState } from 'react'
import { Eye, EyeOff } from 'lucide-react'
import type { LoginForm } from '../../../lib/validations'
import type { UseFormReturn } from 'react-hook-form'
import { Input } from '../Input'
import { Button } from '../Button'

interface LoginFormProps {
    form: UseFormReturn<LoginForm>
    onSubmit: (data: LoginForm) => void | Promise<void>
    isLoading: boolean
}

export function LoginForm({
    form, onSubmit, isLoading
}: LoginFormProps) {
    const [showPassword, setShowPassword] = useState(false)
    const { errors } = form.formState

    return <form onSubmit={form.handleSubmit(onSubmit)} className="flex flex-col gap-5 w-full">
        <div className="flex flex-col gap-1">
            <Input
                label="Email"
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
        <Button
            type="submit"
            isLoading={isLoading}
            className="w-full mt-2 bg-primary-400 hover:bg-primary-300 text-black border border-neutral-900 rounded shadow-hard-md hover:translate-x-0.5 hover:translate-y-0.5 hover:shadow-none transition-all duration-150 font-medium py-3 text-base flex items-center justify-center gap-2"
        >
            Iniciar Sesión
        </Button>
    </form>
}