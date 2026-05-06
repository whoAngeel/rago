import { z } from 'zod'

export const loginSchema = z.object({
    email: z.email("Email invalido"),
    password: z.string().min(1, "La contrasena es requerida").min(8, "La contrasena debe tener al menos 8 caracteres")
})

export const registerSchema = z.object({
    name: z.
        string().
        min(1, 'El nombre es obligatorio').
        min(2, 'El nombre debe tener al menos 2 caracteres'),
    email: z.email('Email invalido'),
    password: z.
        string().
        min(1, "La contrasena es requerida").
        min(8, "La contrasena debe tener al menos 8 caracteres"),
    password_confirmation: z.
        string().
        min(1, "La confirmacion de la contrasena es requerida").
        min(8, "La confirmacion de la contrasena debe tener al menos 8 caracteres"),
}).refine((data) => data.password === data.password_confirmation, {
    message: "Las contraseñas no coinciden",
    path: ["password_confirmation"],
})

export type LoginForm = z.infer<typeof loginSchema>
export type RegisterForm = z.infer<typeof registerSchema>