import { create } from "zustand"

export type ToastType = "success" | "error" | "info"

export interface Toast {
    id: string
    title: string
    message: string
    type: ToastType
}

interface ToastState {
    toasts: Toast[]
    add: (title: string, message: string, type?: ToastType) => void
    remove: (id: string) => void
}

let count = 0

export const useToastStore = create<ToastState>()((set) => ({
    toasts: [],

    add: (title, message, type = "error") => {
        const id = `toast-${Date.now()}-${count++}`
        set((state) => ({
            toasts: [...state.toasts, { id, title, message, type }],
        }))
        setTimeout(() => {
            set((state) => ({
                toasts: state.toasts.filter((t) => t.id !== id),
            }))
        }, 4000)
    },

    remove: (id) => {
        set((state) => ({
            toasts: state.toasts.filter((t) => t.id !== id),
        }))
    },
}))