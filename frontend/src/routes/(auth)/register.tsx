import { createFileRoute, Link, useNavigate, useSearch } from '@tanstack/react-router'
import { useAuthStore } from '../../store/authStore'
import { registerSchema, type RegisterForm } from '../../lib/validations'
import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'
import { RegisterForm as RegisterFormComponent } from '../../components/ui/auth/RegisterForm'
import { useEffect } from 'react'

export const Route = createFileRoute("/(auth)/register")({
  validateSearch: (search: Record<string, unknown>) => {
    if (search.redirect && typeof search.redirect === "string") {
      return { redirect: search.redirect }
    }
    return { redirect: "/dashboard" }
  },
  component: RouteComponent,
})

function RouteComponent() {
  const { register, isLoading, isAuthenticated } = useAuthStore()
  const navigate = useNavigate()
  const search = useSearch({ from: "/(auth)/register" })

  const form = useForm<RegisterForm>({
    resolver: zodResolver(registerSchema),
  })

  const onSubmit = async (data: RegisterForm) => {
    await register(data.email, data.password, data.name, "viewer")
  }

  useEffect(() => {
    if (isAuthenticated) {
      navigate({ to: search.redirect || "/dashboard" })
    }
  }, [isAuthenticated])

  if (isAuthenticated) return null

  return (
    <div className='flex items-center justify-center min-h-screen bg-accent-pink p-4 font-sans'>
      <div className='w-full max-w-md bg-white border-2 border-neutral-800 rounded-card shadow-hard-lg px-8 py-10 flex flex-col gap-8'>
        <div className="flex flex-col gap-2 text-center">
          <h1 className='text-3xl leading-snug font-medium text-black tracking-[-0.64px]'>
            Crear Cuenta
          </h1>
          <p className='text-base leading-snug font-medium text-neutral-600'>
            ¿Ya tienes una cuenta?{" "}
            <Link className="font-medium underline underline-offset-4 hover:text-black transition-colors" to="/login" search={{ redirect: search.redirect }}>Inicia Sesión</Link>
          </p>
        </div>
        <RegisterFormComponent
          form={form}
          onSubmit={onSubmit}
          isLoading={isLoading}
        />
      </div>
    </div>
  )
}
