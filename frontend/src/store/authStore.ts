import { create } from "zustand"
import { persist } from "zustand/middleware"
import axios from "axios"
import type { User } from "../types"
import { API_BASE_URL, API_VERSION } from "../lib/constants"

interface AuthState {
    user: User | null
    accessToken: string | null
    refreshToken: string | null
    isAuthenticated: boolean
    isLoading: boolean

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

            login: async (email: string, password: string) => {
                set({ isLoading: true })
                try {
                    const res = await axios.post(`${API_BASE_URL}${API_VERSION}/auth/login`, { email, password })
                    const { access_token: accessToken, refresh_token: refreshToken } = res.data
                    set({ accessToken, refreshToken, isAuthenticated: true, user: res.data?.user })
                } catch (error) {
                    console.error(error)
                } finally {
                    set({ isLoading: false })
                }
            },

            logout: async () => {
                set({ isLoading: true })
                try {
                    await axios.post(`${API_BASE_URL}${API_VERSION}/auth/logout`, { refresh_token: get().refreshToken })
                } catch (error) {
                    console.error(error)
                } finally {
                    set({ user: null, accessToken: null, refreshToken: null, isAuthenticated: false, isLoading: false })
                }
            },

            refresh: async () => {
                set({ isLoading: true })
                try {
                    const res = await axios.post(`${API_BASE_URL}${API_VERSION}/auth/refresh`, { refresh_token: get().refreshToken })
                    const { access_token: accessToken, refresh_token: refreshToken } = res.data
                    set({ accessToken, refreshToken, isAuthenticated: true })
                } catch (error) {
                    console.error(error)
                } finally {
                    set({ isLoading: false })
                }
            },
            register: async (email: string, password: string, name: string, role: string) => {
                // TODO
            },

            setUser: (user: User) => { set({ user }) },
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