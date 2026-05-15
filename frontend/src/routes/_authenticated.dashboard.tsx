import { createFileRoute } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import api from '../lib/api'
import type { User } from '../types'
import { FileText, Layers, HardDrive, MessageSquare, AlertTriangle, CheckCircle2, XCircle, Clock, Users } from 'lucide-react'

interface DashboardStats {
  total_documents: number
  by_status: {
    completed: number
    failed: number
    pending: number
  }
  storage_used: number
  total_groups: number
  active_groups: number
  total_attempts: number
  unanswered_total: number
}

export const Route = createFileRoute('/_authenticated/dashboard')({
  component: DashboardPage,
})

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`
}

function StatCard({ icon: Icon, label, value, sub, variant }: {
  icon: React.ComponentType<{ size?: number; className?: string }>
  label: string
  value: string
  sub?: string
  variant?: 'default' | 'warning'
}) {
  return (
    <div className={`bg-white border-2 rounded-xl p-4 sm:p-5 flex items-start gap-3 sm:gap-4 transition-colors ${variant === 'warning' ? 'border-orange-200' : 'border-neutral-200 hover:border-neutral-400'}`}>
      <div className={`w-10 h-10 sm:w-12 sm:h-12 border-2 border-neutral-950 rounded-lg flex items-center justify-center shrink-0 ${variant === 'warning' ? 'bg-orange-100' : 'bg-neutral-100'}`}>
        <Icon size={20} className={variant === 'warning' ? 'text-orange-700' : 'text-neutral-700'} />
      </div>
      <div className="flex-1 min-w-0">
        <p className="text-xs font-bold text-neutral-500 uppercase tracking-wider">{label}</p>
        <p className="text-xl sm:text-2xl font-black text-neutral-950 mt-0.5">{value}</p>
        {sub && <p className="text-xs text-neutral-500 mt-0.5">{sub}</p>}
      </div>
    </div>
  )
}

function DashboardPage() {
  const { data: user } = useQuery<User>({
    queryKey: ['users', 'me'],
    queryFn: async () => {
      const { data } = await api.get('/users/me')
      return data
    },
    staleTime: 10000,
  })

  const { data: stats, isLoading } = useQuery<DashboardStats>({
    queryKey: ['users', 'me', 'stats'],
    queryFn: async () => {
      const { data } = await api.get('/users/me/stats')
      return data
    },
    staleTime: 30000,
  })

  const total = stats?.total_documents ?? 0
  const completed = stats?.by_status?.completed ?? 0
  const failed = stats?.by_status?.failed ?? 0
  const pending = stats?.by_status?.pending ?? 0
  const processing = Math.max(0, total - completed - failed - pending)
  const storage = stats?.storage_used ?? 0
  const unanswered = stats?.unanswered_total ?? 0
  const attempts = stats?.total_attempts ?? 0
  const unansweredRate = attempts > 0 ? Math.round((unanswered / attempts) * 100) : 0

  return (
    <div className="p-4 sm:p-6 max-w-5xl mx-auto flex flex-col gap-4 sm:gap-6">
      <div>
        <h1 className="text-2xl font-bold text-neutral-900">
          {user ? `Hola, ${user.name}` : 'Dashboard'}
        </h1>
        <p className="text-neutral-600 mt-1 text-sm sm:text-base">
          Resumen de tu actividad
        </p>
      </div>

      {isLoading ? (
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-3 sm:gap-4">
          {Array.from({ length: 4 }).map((_, i) => (
            <div key={i} className="bg-white border-2 border-neutral-200 rounded-xl p-5 h-24 animate-pulse">
              <div className="h-3 bg-neutral-100 rounded w-1/2 mb-3" />
              <div className="h-6 bg-neutral-100 rounded w-1/3" />
            </div>
          ))}
        </div>
      ) : (
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-3 sm:gap-4">
          <StatCard
            icon={FileText}
            label="Documentos"
            value={String(total)}
            sub={`${completed} completados`}
          />
          <StatCard
            icon={Layers}
            label="Grupos"
            value={String(stats?.active_groups ?? 0)}
            sub={`de ${stats?.total_groups ?? 0} totales`}
          />
          <StatCard
            icon={MessageSquare}
            label="Consultas"
            value={String(attempts)}
            sub={`${unansweredRate > 0 ? `${unansweredRate}% sin respuesta` : 'todas respondidas'}`}
          />
          <StatCard
            icon={HardDrive}
            label="Almacenamiento"
            value={formatBytes(storage)}
            sub={user ? `${user.document_count} / ${user.max_documents > 0 ? user.max_documents : '∞'} archivos` : ''}
          />
        </div>
      )}

      {stats && unanswered > 0 && (
        <div className="bg-orange-50 border-2 border-orange-200 rounded-xl p-4 flex items-center gap-3">
          <AlertTriangle size={20} className="text-orange-600 shrink-0" />
          <div>
            <p className="text-sm font-bold text-orange-800">{unanswered} preguntas sin respuesta suficiente</p>
            <p className="text-xs text-orange-600">
              El agente no encontró suficiente contexto en los documentos del grupo. Revisa los documentos subidos.
            </p>
          </div>
        </div>
      )}

      {stats && total > 0 && (
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 sm:gap-4">
          <div className="bg-white border-2 border-green-200 rounded-xl p-4 flex items-center gap-3">
            <CheckCircle2 size={18} className="text-green-600 shrink-0" />
            <div>
              <p className="text-xs font-bold text-neutral-500 uppercase tracking-wider">Completados</p>
              <p className="text-lg font-black text-green-700">{completed}</p>
            </div>
          </div>
          <div className="bg-white border-2 border-orange-200 rounded-xl p-4 flex items-center gap-3">
            <Clock size={18} className="text-orange-600 shrink-0" />
            <div>
              <p className="text-xs font-bold text-neutral-500 uppercase tracking-wider">Procesando</p>
              <p className="text-lg font-black text-orange-700">{processing}</p>
            </div>
          </div>
          <div className="bg-white border-2 border-amber-200 rounded-xl p-4 flex items-center gap-3">
            <Clock size={18} className="text-amber-600 shrink-0" />
            <div>
              <p className="text-xs font-bold text-neutral-500 uppercase tracking-wider">Pendientes</p>
              <p className="text-lg font-black text-amber-700">{pending}</p>
            </div>
          </div>
          <div className="bg-white border-2 border-red-200 rounded-xl p-4 flex items-center gap-3">
            <XCircle size={18} className="text-red-600 shrink-0" />
            <div>
              <p className="text-xs font-bold text-neutral-500 uppercase tracking-wider">Fallidos</p>
              <p className="text-lg font-black text-red-700">{failed}</p>
            </div>
          </div>
        </div>
      )}

      {stats && total === 0 && stats?.total_groups === 0 && (
        <div className="bg-white border-2 border-neutral-200 rounded-xl p-12 text-center">
          <Users size={40} className="text-neutral-300 mx-auto mb-4" />
          <p className="font-bold text-neutral-600 text-lg">Sin actividad aún</p>
          <p className="text-sm text-neutral-500 mt-1">
            Crea tu primer grupo y sube documentos para empezar
          </p>
        </div>
      )}
    </div>
  )
}
