import { Send, Loader2 } from 'lucide-react'
import type { SyntheticEvent } from 'react'
import ReactMarkdown from "react-markdown"

interface ChatMessage {
  role: 'user' | 'assistant'
  content: string
  timestamp?: string
}

interface PublicChatAreaProps {
  messages: ChatMessage[]
  input: string
  setInput: (value: string) => void
  isSending: boolean
  isQuotaExceeded: boolean
  isDisabled?: boolean
  onSubmit: (e: SyntheticEvent<HTMLFormElement>) => void
}

export function PublicChatArea({ messages, input, setInput, isSending, isQuotaExceeded, isDisabled, onSubmit }: PublicChatAreaProps) {
  return (
    <div className='col-span-full lg:col-span-8 flex flex-col h-full bg-white'>
      {/* Messages Area */}
      <div className='flex-1 overflow-y-auto p-4 lg:p-6 flex flex-col gap-4'>
        {isDisabled && (
          <div className='self-center bg-amber-100 p-4 border-2 border-neutral-950 rounded shadow-hard-sm max-w-full lg:max-w-[90%]'>
            <p className='text-sm font-bold text-neutral-900'>Este chat está desactivado por el administrador del grupo.</p>
            <p className='text-xs text-neutral-600 mt-1'>Mientras el grupo esté inactivo no se pueden enviar mensajes.</p>
          </div>
        )}
        {messages.length === 0 && !isDisabled ? (
          <div className='flex-1 flex flex-col items-center justify-center text-neutral-500'>
            <p className='text-lg font-bold mb-1'>¡Bienvenido al chat!</p>
            <p className='text-sm font-medium'>Haz una pregunta sobre los documentos del grupo.</p>
          </div>
        ) : (
          messages.map((msg, index) => (
            <div
              key={index}
              className={`max-w-[85%] lg:max-w-[70%] p-4 border-2 border-neutral-950 shadow-hard-sm relative ${
                msg.role === 'user'
                  ? 'self-end bg-primary-200 rounded-t-lg rounded-bl-lg rounded-br-none'
                  : 'self-start bg-neutral-50 rounded-t-lg rounded-br-lg rounded-bl-none'
              }`}
            >
              {/* Bubble Tail */}
              {msg.role === 'user' ? (
                <div className="absolute -bottom-2 right-4 w-4 h-4 bg-primary-200 border-b-2 border-r-2 border-neutral-950 transform rotate-45"></div>
              ) : (
                <div className="absolute -bottom-2 left-4 w-4 h-4 bg-neutral-50 border-b-2 border-r-2 border-neutral-950 transform rotate-45"></div>
              )}
              <p className="text-xs font-bold text-neutral-500 mb-1">
                {msg.role === 'user' ? 'Tú' : 'Asistente'}
                {msg.timestamp && <> · {new Date(msg.timestamp).toLocaleTimeString()}</>}
              </p>
              {msg.role === 'assistant' ? (
                <div className="text-sm font-medium text-neutral-800 [&_p]:mb-2 [&_ul]:list-disc [&_ul]:ml-4 [&_ol]:list-decimal [&_ol]:ml-4 [&_strong]:font-bold [&_em]:italic">
                  <ReactMarkdown>{msg.content}</ReactMarkdown>
                </div>
              ) : (
                <p className='text-sm font-medium text-neutral-950 whitespace-pre-wrap'>{msg.content}</p>
              )}
            </div>
          ))
        )}
        {isSending && (
          <div className='self-start bg-neutral-50 p-4 border-2 border-neutral-950 rounded-t-lg rounded-br-lg rounded-bl-none shadow-hard-sm max-w-[85%] lg:max-w-[70%]'>
            <div className='flex items-center gap-2'>
              <Loader2 className='animate-spin text-neutral-600' size={16} />
              <span className='text-sm font-bold text-neutral-600'>Generando respuesta...</span>
            </div>
          </div>
        )}
        {isQuotaExceeded && (
          <div className='self-center bg-red-100 p-4 border-2 border-neutral-950 rounded shadow-hard-sm max-w-full lg:max-w-[90%] mt-2'>
            <p className='text-sm font-bold text-red-700'>
              Se ha excedido la cuota de mensajes para este grupo. No se pueden procesar más mensajes en este momento.
            </p>
          </div>
        )}
      </div>

      {/* Input Area */}
      <div className='p-4 border-t-2 border-neutral-950 bg-white'>
        <form onSubmit={onSubmit} className='flex gap-2 lg:gap-3'>
          <input
            value={input}
            onChange={(e) => setInput(e.target.value)}
            placeholder={isDisabled ? "El chat está desactivado" : "Escribe tu mensaje aquí..."}
            disabled={isSending || isQuotaExceeded || isDisabled}
            className='flex-1 min-w-0 border-2 border-neutral-950 p-3 rounded focus:outline-none focus:border-primary-500 shadow-hard-sm disabled:bg-neutral-100 disabled:cursor-not-allowed font-medium text-sm'
          />
          <button
            type='submit'
            disabled={isSending || isQuotaExceeded || isDisabled || !input.trim()}
            className='border-2 border-neutral-950 p-3 bg-primary-400 rounded shadow-hard-sm font-bold hover:translate-x-px hover:translate-y-px hover:shadow-none transition-all disabled:bg-neutral-200 disabled:cursor-not-allowed disabled:shadow-none cursor-pointer shrink-0'
          >
            <Send size={18} />
          </button>
        </form>
      </div>
    </div>
  )
}
