import axios from "axios"
import { API_BASE_URL, API_VERSION } from "./constants"
import { useAuthStore } from "../store/authStore"

const api = axios.create({
  baseURL: `${API_BASE_URL}${API_VERSION}`,
  withCredentials: true,
  headers: {
    // "Content-Type": "application/json",
    Accept: "application/json",
  },
})

api.interceptors.request.use(
  (config) => {
    const token = useAuthStore.getState().accessToken
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => Promise.reject(error)
)

let refreshPromise: Promise<void> | null = null

function getRefreshPromise(): Promise<void> {
  if (refreshPromise) return refreshPromise

  refreshPromise = useAuthStore.getState().refresh()
    .finally(() => { refreshPromise = null })
  return refreshPromise
}

api.interceptors.response.use(
  (response) => response,
  async (error) => {
    if (error.response?.status === 401 && !error.config._retry) {
      error.config._retry = true

      try {
        await getRefreshPromise()
      } catch {
        // refresh failed — all queued 401s share this single failure
      }

      const newToken = useAuthStore.getState().accessToken
      if (newToken) {
        error.config.headers.Authorization = `Bearer ${newToken}`
        return api.request(error.config)
      }
    }

    return Promise.reject(error)
  }
)

export default api
