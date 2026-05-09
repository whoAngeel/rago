interface User {
    id: number
    name: string | undefined,
    email: string
    role: 'editor' | 'admin' | 'viewer'
    max_documents: number
    document_count: number
}

interface Document {
    id: number
    user_id: number
    filename: string
    file_path: string
    content_type: string
    status: 'uploading' | 'pending' | 'processing' | 'completed' | 'failed'
    size: number
    created_at: string
    updated_at: string
    processing_started_at: string | null
    error_message: string | null
    retry_count: number
}

interface Group {
    id: number
    user_id: number
    name: string
    is_active: boolean
    slug: string
    allow_downloads: boolean
    chat_quota: number
    chat_quota_used: number
    chat_attempts: number
    document_ids: number[]
    created_at: string
    updated_at: string

}
interface ChatSession {
    id: number
    user_id: number
    title: string
    created_at: string
    updated_at: string
}

interface Source {
    content: string
    source: string
    score: number
}

interface ChatMessage {
    id: number
    session_id: number
    role: 'user' | 'assistant'
    content: string
    sources: string
    created_at: string
}

interface ProcessingStep {
    id: number
    document_id: number
    step_name: string
    status: 'started' | 'completed' | 'failed'
    error_message: string | null
    duration_ms: number | null
    created_at: string
}

export type { User, Document, ChatSession, Source, ChatMessage, ProcessingStep, Group }