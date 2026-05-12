import { useQueryClient } from "@tanstack/react-query";
import { useEffect } from "react";
import { useAuthStore } from "../store/authStore";
import { API_BASE_URL, API_VERSION } from "../lib/constants";
import type { ProcessingStep } from "../types";

export function useSSE() {
    const queryClient = useQueryClient()
    const token = useAuthStore.getState().accessToken

    useEffect(() => {
        if (!token) return;

        const url = `${API_BASE_URL}${API_VERSION}/stream?token=${token}`
        const eventSource = new EventSource(url)

        eventSource.addEventListener("document_status", (event) => {
            const data = JSON.parse(event.data) as { id: number; status: string; error: string }
            // Clear stale step cache whenever a new processing attempt starts
            if (data.status === 'processing') {
                queryClient.removeQueries({ queryKey: ["steps", data.id] })
            }
            queryClient.invalidateQueries({ queryKey: ["documents"] })
        })

        eventSource.addEventListener("document_step", (event) => {
            const data = JSON.parse(event.data) as {
                doc_id: number
                step_id: number
                step_name: string
                status: string
                error: string
            }
            queryClient.setQueryData<ProcessingStep[]>(
                ["steps", data.doc_id],
                (old) => {
                    const prev = old ?? []
                    const updated: ProcessingStep = {
                        id: data.step_id,
                        document_id: data.doc_id,
                        step_name: data.step_name,
                        status: data.status as ProcessingStep["status"],
                        duration_ms: null,
                        error_message: data.error || null,
                        created_at: new Date().toISOString(),
                    }
                    const idx = prev.findIndex(s => s.step_name === data.step_name)
                    if (idx >= 0) {
                        return prev.map((s, i) => i === idx ? updated : s)
                    }
                    return [...prev, updated]
                }
            )
        })

        eventSource.addEventListener("group_usage_updated", (event) => {
            const data = JSON.parse(event.data) as { group_id: number; chat_attempts: number }
            queryClient.invalidateQueries({ queryKey: ["group", String(data.group_id)] })
            queryClient.invalidateQueries({ queryKey: ["groups"] })
        })

        return () => eventSource.close()

    }, [token, queryClient])
}