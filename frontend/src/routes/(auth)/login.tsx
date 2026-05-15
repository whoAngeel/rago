import { createFileRoute, Link, useNavigate, useSearch } from '@tanstack/react-router'
import { useAuthStore, useIsAuthenticated } from '../../store/authStore'
import { useForm } from 'react-hook-form'
import { loginSchema, type LoginForm } from '../../lib/validations'
import { zodResolver } from '@hookform/resolvers/zod'
import { LoginForm as LoginFormComponent } from '../../components/ui/auth/LoginForm'
import { useEffect } from 'react'

export const Route = createFileRoute('/(auth)/login')({
    validateSearch: (search: Record<string, unknown>) => {
        if (search.redirect && typeof search.redirect === "string") {
            return { redirect: search.redirect }
        }
        return { redirect: "/dashboard" }
    },
    component: RouteComponent,
})

function RouteComponent() {
    const { login, isLoading } = useAuthStore()
    const isAuthenticated = useIsAuthenticated()

    const navigate = useNavigate()
    const search = useSearch({ from: "/(auth)/login" })

    const form = useForm<LoginForm>({
        resolver: zodResolver(loginSchema),
        mode: "onChange",
    })

    const onSubmit = async (data: LoginForm) => {
        await login(data.email, data.password)
    }

    useEffect(() => {
        if (isAuthenticated) {
            navigate({ to: search.redirect || '/dashboard' })
        }
    }, [isAuthenticated])


    if (isAuthenticated) return null

    return (
        <div className='flex items-center justify-center min-h-screen bg-accent-pink p-4 font-sans'>
            <div className='w-full max-w-md bg-white border-2 border-neutral-800 rounded-card shadow-hard-lg px-8 py-10 flex flex-col gap-8'>
                <div className="flex flex-col gap-2 text-center">
                    <h1 className='text-3xl leading-snug font-medium text-black tracking-[-0.64px]'>
                        Iniciar Sesión
                    </h1>
                    <p className='text-base leading-snug font-medium text-neutral-600'>
                        Bienvenido de nuevo
                    </p>
                </div>
                <LoginFormComponent
                    form={form}
                    onSubmit={onSubmit}
                    isLoading={isLoading}
                />
                <p className='text-base leading-snug font-medium text-neutral-600 text-center'>
                    ¿No tienes una cuenta?{" "}
                    <Link className="font-medium underline underline-offset-4 hover:text-black transition-colors" to="/register" search={{ redirect: search.redirect }}>Registrate</Link>
                </p>
            </div>
        </div>
    )
}
