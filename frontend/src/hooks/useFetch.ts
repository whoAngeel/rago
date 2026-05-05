import { useEffect, useState } from "react"

type Data<T> = T | null
type ErrorType = string | null

interface Params<T> {
    data: Data<T>
    loading: boolean
    error: ErrorType
}

export function useFetch<T>(url: string): Params<T> {
    const [data, setData] = useState<Data<T>>(null)
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<ErrorType>(null)

    useEffect(() => {
        const controller = new AbortController()
        setLoading(true)

        const fetchData = async () => {
            try {
                const response = await fetch(url, { signal: controller.signal })
                if (!response.ok) {
                    throw new Error("Error en la peticion")
                }
                const json = await response.json()
                setData(json)
                setError(null)
            } catch (error: any) {
                if (error.name === "AbortError") {
                    return // Ignoramos el error si fue por cancelar la petición
                }
                if (error instanceof Error) {
                    setError(error.message)
                } else {
                    setError("Error en la peticion")
                }
                setData(null)
            } finally {
                if (!controller.signal.aborted) {
                    setLoading(false)
                }
            }
        }
        fetchData()

        return () => {
            controller.abort()
        }
    }, [url])


    return { data, loading, error }

}

