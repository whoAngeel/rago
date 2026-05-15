import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/_authenticated/chat')({
  component: ChatPage,
})

function ChatPage() {
  return (
    <div className="p-8">
      <h1 className="text-h2">Chat RAG</h1>
      <p className='text-body text-neutral-600 mt-4'>
        Asistente Conversacional
      </p>
    </div>
  )
}
