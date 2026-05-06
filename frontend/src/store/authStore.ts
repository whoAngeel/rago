import { create } from "zustand"
import { persist } from "zustand/middleware"
import axios from "axios"
import type { User } from "../types"
import { API_BASE_URL, API_VERSION } from "../lib/constants"
import { useToastStore } from "./toastStore"


interface AuthState {
    user: User | null
    accessToken: string | null
    refreshToken: string | null
    isAuthenticated: boolean
    isLoading: boolean
    isRefreshing: boolean

    login: (email: string, password: string) => Promise<void>
    register: (email: string, password: string, name: string, role: string) => Promise<void>
    logout: () => Promise<void>
    refresh: () => Promise<void>
    setUser: (user: User) => void
}

export const useAuthStore = create<AuthState>()(
    persist(
        (set, get) => ({
            user: null,
            accessToken: null,
            refreshToken: null,
            isAuthenticated: false,
            isLoading: false,
            isRefreshing: false,

            login: async (email, password) => {
                set({ isLoading: true })
                try {
                    const res = await axios.post(
                        `${API_BASE_URL}${API_VERSION}/auth/login`,
                        { email, password }
                    )
                    const {
                        access_token: accessToken,
                        refresh_token: refreshToken,
                        user,
                    } = res.data
                    set({ accessToken, refreshToken, user, isAuthenticated: true })
                } catch (error) {
                    useToastStore.getState().add("Error de Autenticación", "Credenciales inválidas", "error")
                    throw error
                } finally {
                    set({ isLoading: false })
                }
            },

            register: async (email, password, name, role) => {
                set({ isLoading: true })
                try {
                    const res = await axios.post(
                        `${API_BASE_URL}${API_VERSION}/auth/register`,
                        { email, password, name, role }
                    )
                    const {
                        access_token: accessToken,
                        refresh_token: refreshToken,
                        user,
                    } = res.data
                    set({ accessToken, refreshToken, user, isAuthenticated: true })
                } catch (error) {
                    useToastStore.getState().add("Error de Registro", "Credenciales inválidas", "error")
                    throw error
                } finally {
                    set({ isLoading: false })
                }
            },

            logout: async () => {
                set({ isLoading: true })
                try {
                    await axios.post(
                        `${API_BASE_URL}${API_VERSION}/auth/logout`,
                        { refresh_token: get().refreshToken }
                    )
                } catch (error) {
                    console.error(error)
                    useToastStore.getState().add("Error de Cierre de Sesión", "Error al cerrar sesión", "error")

                } finally {
                    set({
                        user: null,
                        accessToken: null,
                        refreshToken: null,
                        isAuthenticated: false,
                        isRefreshing: false,

                    })
                }
            },

            refresh: async () => {
                set({ isRefreshing: true })
                try {
                    const res = await axios.post(
                        `${API_BASE_URL}${API_VERSION}/auth/refresh`,
                        { refresh_token: get().refreshToken }
                    )
                    const {
                        access_token: accessToken,
                        refresh_token: refreshToken,
                        user,
                    } = res.data
                    set({ accessToken, refreshToken, user, isAuthenticated: true, isRefreshing: false })
                } catch (error) {
                    useToastStore.getState().add("Error de Autenticación", "Error al refrescar token", "error")
                    set({
                        user: null,
                        accessToken: null,
                        refreshToken: null,
                        isAuthenticated: false,
                        isRefreshing: false,
                    })

                    throw error
                } finally {
                    set({ isRefreshing: false })
                }
            },

            setUser: (user) => set({ user }),
        }),
        {
            name: "auth-storage",
            partialize: (state) => ({ refreshToken: state.refreshToken }),
        }
    )
)

export const useUser = () => useAuthStore((s) => s.user)
export const useIsAuthenticated = () => useAuthStore((s) => s.isAuthenticated)
export const useAccessToken = () => useAuthStore((s) => s.accessToken)
export const useRefreshToken = () => useAuthStore((s) => s.refreshToken)
