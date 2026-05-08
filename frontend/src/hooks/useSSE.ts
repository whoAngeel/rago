import { useQueryClient } from "@tanstack/react-query";
import { useEffect } from "react";
import { useAuthStore } from "../store/authStore";
import { API_BASE_URL, API_VERSION } from "../lib/constants";
import type { Document as DocumentType } from "../types";

export function useSSE() {
    const queryClient = useQueryClient()
    const token = useAuthStore.getState().accessToken

    useEffect(() => {
        if (!token) return;

        const url = `${API_BASE_URL}${API_VERSION}/stream?token=${token}`
        const eventSource = new EventSource(url)

        eventSource.addEventListener("document_status", (event) => {
            const data = JSON.parse(event.data)
            queryClient.setQueryData<DocumentType[]>(["documents"], (old) => {
                if (!old) return old
                return old.map((doc) =>
                    doc.id === data.id
                        ? { ...doc, status: data.status, error_message: data.error ?? null }
                        : doc
                )
            })
            queryClient.invalidateQueries({ queryKey: ["documents"] })
        })

        return () => eventSource.close()

    }, [token, queryClient])
}