interface User {
    id: number
    name: string | undefined,
    email: string
    role: 'editor' | 'admin' | 'viewer'
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
    error_message: null
    retry_count: number
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
