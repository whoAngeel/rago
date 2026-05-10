import { createFileRoute } from '@tanstack/react-router'
import { useState } from 'react'
import type { SyntheticEvent } from 'react'
import type { GroupPublicInfo } from '../types'
import api from '../lib/api'
import { useMutation, useQuery } from '@tanstack/react-query'
import { Loader2 } from 'lucide-react'
import { PublicChatNavbar } from '../components/groups/PublicChatNavbar'
import { PublicChatArea } from '../components/groups/PublicChatArea'
import { PublicDocumentsPanel } from '../components/groups/PublicDocumentsPanel'

interface ChatMessage {
  role: 'user' | 'assistant'
  content: string
  timestamp: string
}

export const Route = createFileRoute('/c/$slug')({
  component: RouteComponent,
})

function RouteComponent() {
  const { slug } = Route.useParams()
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [input, setInput] = useState("")
  const [isSendig, setIsSending] = useState<boolean>(false)
  const [isQuotaExceeded, setIsQuotaExceeded] = useState(false)

  const { data: group, isLoading, error } = useQuery<GroupPublicInfo>({
    queryKey: ["public_group", slug],
    queryFn: async () => {
      const { data } = await api.get(`/public/groups/${slug}`)
      return data
    }
  })

  const chatMutation = useMutation({
    mutationFn: async (message: string) => {
      const { data } = await api.post(`/public/groups/${slug}/chat`, { message })
      return data.answer
    },
    onMutate: (message) => {
      setMessages(prev => [...prev, { role: "user", content: message, timestamp: new Date().toISOString() }])
      setInput("")
      setIsSending(true)
    },
    onSuccess: (answer) => {
      if (answer) {
        setMessages(prev => [...prev, { role: "assistant", content: answer, timestamp: new Date().toISOString() }])
      }
    },
    onError: (err: any) => {
      console.error(err)
      if (err?.response?.data?.error === "quota_exceeded") {
        setIsQuotaExceeded(true)
      }
    },
    onSettled: () => { setIsSending(false) }
  })

  const handleSubmit = (e: SyntheticEvent<HTMLFormElement>) => {
    e.preventDefault()
    if (input.trim() && !isSendig) {
      chatMutation.mutate(input.trim())
    }
  }

  if (isLoading) return <div className="flex items-center justify-center min-h-screen bg-neutral-50"><Loader2 className="animate-spin text-neutral-600" size={32} /></div>
  if (error) {
    const isNotFound = (error as any)?.response?.status === 404
    return (
      <div className="flex items-center justify-center min-h-screen bg-neutral-50 text-neutral-950 font-bold">
        {isNotFound ? "Este chat no está disponible" : "Error al cargar el grupo"}
      </div>
    )
  }

  return (
    <div className="flex flex-col h-screen bg-neutral-50 font-sans">
      <PublicChatNavbar groupName={group?.name} />

      <div className='grid grid-cols-12 w-full flex-1 overflow-hidden'>
        <PublicChatArea
          messages={messages}
          input={input}
          setInput={setInput}
          isSending={isSendig}
          isQuotaExceeded={isQuotaExceeded}
          onSubmit={handleSubmit}
        />

        <PublicDocumentsPanel
          documents={group?.documents}
          allowDownloads={group?.allow_downloads ?? false}
          slug={slug}
        />
      </div>
    </div>
  )
}
