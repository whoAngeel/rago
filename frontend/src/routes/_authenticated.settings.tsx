import { createFileRoute } from '@tanstack/react-router'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useEffect, useState, useRef } from 'react'
import { Settings, Terminal, Info, AlertCircle, RotateCcw, Check, Loader2 } from 'lucide-react'
import api from '../lib/api'
import { useToastStore } from '../store/toastStore'

export const Route = createFileRoute('/_authenticated/settings')({
  component: RouteComponent,
})

interface SystemPromptResponse {
  value: string
}

const DEFAULT_PROMPT = "Eres un asistente de IA experto. Tu objetivo es responder preguntas utilizando únicamente los documentos proporcionados en el contexto. Si no sabes la respuesta, dilo claramente."

function RouteComponent() {
  const addToast = useToastStore((s) => s.add)
  const queryClient = useQueryClient()
  const [prompt, setPrompt] = useState('')
  const [isEditing, setIsEditing] = useState(false)
  const [saveStatus, setSaveStatus] = useState<'idle' | 'saving' | 'saved'>('idle')
  const textareaRef = useRef<HTMLTextAreaElement>(null)

  const { data, isLoading } = useQuery<SystemPromptResponse>({
    queryKey: ['config', 'system-prompt'],
    queryFn: async () => {
      const { data } = await api.get('/config/system-prompt')
      return data
    },
  })

  useEffect(() => {
    if (data?.value && !isEditing) {
      setPrompt(data.value)
    }
  }, [data, isEditing])

  const mutation = useMutation({
    mutationFn: async (value: string) => {
      await api.put('/config/system-prompt', { value })
    },
    onMutate: () => {
      setSaveStatus('saving')
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['config', 'system-prompt'] })
      setSaveStatus('saved')
      setTimeout(() => setSaveStatus('idle'), 2000)
    },
    onError: () => {
      setSaveStatus('idle')
      addToast('Error', 'No se pudo guardar el system prompt', 'error')
    },
  })

  const handleBlur = () => {
    setIsEditing(false)
    if (data && prompt !== data.value) {
      mutation.mutate(prompt)
    }
  }

  const handleReset = () => {
    setPrompt(DEFAULT_PROMPT)
    mutation.mutate(DEFAULT_PROMPT)
  }

  return (
    <div className="p-8 max-w-4xl mx-auto flex flex-col gap-8 animate-in fade-in duration-300">
      {/* Header */}
      <div className="flex items-center gap-4 border-b-2 border-neutral-200 pb-6">
        <div className="w-12 h-12 bg-[#84cc17] border-2 border-neutral-950 rounded-[var(--radius-btn)] flex items-center justify-center shadow-[var(--shadow-hard-sm)]">
          <Settings className="w-6 h-6 text-neutral-950" />
        </div>
        <div>
          <h1 className="text-3xl font-bold text-neutral-950">Configuración</h1>
          <p className="text-neutral-600 font-medium mt-1">
            Personaliza el comportamiento general del asistente y la plataforma.
          </p>
        </div>
      </div>

      <div className="grid lg:grid-cols-3 gap-8">
        {/* Sidebar Nav */}
        <div className="hidden lg:flex flex-col gap-2">
          <button className="flex items-center gap-3 px-4 py-3 bg-neutral-950 text-white rounded-[var(--radius-btn)] font-bold text-left border-2 border-neutral-950 shadow-[var(--shadow-hard-sm)] transition-all">
            <Terminal className="w-5 h-5" />
            Asistente IA
          </button>
          <button className="flex items-center gap-3 px-4 py-3 bg-white text-neutral-600 rounded-[var(--radius-btn)] font-bold text-left border-2 border-transparent hover:border-neutral-200 transition-all">
            <AlertCircle className="w-5 h-5" />
            Próximamente
          </button>
        </div>

        {/* Main Content Area */}
        <div className="lg:col-span-2 flex flex-col gap-6">
          <div className="bg-white border-2 border-neutral-950 rounded-[var(--radius-card)] p-8 shadow-[var(--shadow-hard-lg)] flex flex-col gap-6 relative overflow-hidden">
            {/* Decorative background element */}
            <div className="absolute -top-10 -right-10 w-32 h-32 bg-[#f5d1fe] rounded-full blur-3xl opacity-50 pointer-events-none" />

            <div className="relative">
              <div className="flex items-center justify-between mb-2">
                <h2 className="text-xl font-bold text-neutral-950 flex items-center gap-3">
                  System Prompt Global
                  {saveStatus === 'saving' && (
                    <span className="flex items-center gap-1 text-sm font-medium text-neutral-500">
                      <Loader2 className="w-4 h-4 animate-spin" /> Guardando...
                    </span>
                  )}
                  {saveStatus === 'saved' && (
                    <span className="flex items-center gap-1 text-sm font-medium text-[#84cc17]">
                      <Check className="w-4 h-4" /> Guardado
                    </span>
                  )}
                </h2>
                
                <button
                  onClick={handleReset}
                  className="flex items-center gap-1.5 text-sm font-bold text-neutral-500 hover:text-neutral-950 transition-colors"
                  title="Restablecer al prompt por defecto"
                >
                  <RotateCcw className="w-4 h-4" />
                  Resetear
                </button>
              </div>
              <p className="text-neutral-600 font-medium text-sm mb-6 flex items-start gap-2 bg-[#f5f5f5] p-3 rounded-md border border-neutral-200">
                <Info className="w-5 h-5 text-[#84cc17] shrink-0" />
                <span>
                  Haz clic en el texto para editarlo. Se guardará automáticamente cuando termines de escribir y hagas clic fuera.
                </span>
              </p>

              {isLoading ? (
                <div className="animate-pulse h-48 bg-neutral-100 rounded-[var(--radius-btn)] border-2 border-neutral-200" />
              ) : (
                <div className="group relative">
                  <textarea
                    ref={textareaRef}
                    value={prompt}
                    onFocus={() => setIsEditing(true)}
                    onBlur={handleBlur}
                    onChange={(e) => setPrompt(e.target.value)}
                    className="w-full min-h-[12rem] p-4 text-sm font-mono text-neutral-800 bg-[#fafafa] focus:bg-white border-2 border-transparent hover:border-neutral-300 focus:border-neutral-950 rounded-[var(--radius-btn)] focus:outline-none transition-all resize-y shadow-none focus:shadow-[var(--shadow-hard-sm)] cursor-pointer focus:cursor-text"
                    placeholder="Haz clic para escribir el prompt..."
                  />
                  {!isEditing && (
                    <div className="absolute top-4 right-4 pointer-events-none opacity-0 group-hover:opacity-100 transition-opacity">
                      <div className="bg-neutral-900 text-white text-xs font-bold px-2 py-1 rounded">
                        Haz clic para editar
                      </div>
                    </div>
                  )}
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
