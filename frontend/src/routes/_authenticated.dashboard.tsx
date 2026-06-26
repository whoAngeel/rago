import { createFileRoute } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import api from '../lib/api'
import type { User } from '../types'
import {
  FileText, Layers, HardDrive, MessageSquare,
  AlertTriangle, CheckCircle2, XCircle, Clock, Users, Zap,
} from 'lucide-react'

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

function StatCard({ icon: Icon, label, value, bgClass }: {
  icon: React.ComponentType<{ size?: number; className?: string }>
  label: string
  value: string | number
  bgClass: string
}) {
  return (
    <div className={`border-2 border-neutral-950 rounded-xl p-4 flex flex-col gap-3 shadow-hard-md hover:translate-x-px hover:translate-y-px hover:shadow-hard-sm transition-all ${bgClass}`}>
      <div className="w-8 h-8 border-2 border-neutral-950 rounded-lg flex items-center justify-center bg-neutral-950 shrink-0">
        <Icon size={15} className="text-white" strokeWidth={2.5} />
      </div>
      <div>
        <p className="text-3xl font-black text-neutral-950 tracking-tighter leading-none mb-1">{value}</p>
        <p className="text-[10px] font-bold text-neutral-500 uppercase tracking-wider leading-tight">{label}</p>
      </div>
    </div>
  )
}

function QuotaCard({ label, used, max, isUnlimited, icon: Icon, bgClass }: {
  label: string
  used: number
  max: number
  isUnlimited: boolean
  icon: React.ComponentType<any>
  bgClass: string
}) {
  const pct = isUnlimited ? 100 : (max > 0 ? Math.min((used / max) * 100, 100) : 0)
  const isHigh = pct > 80

  return (
    <div className={`flex-1 border-2 border-neutral-950 rounded-xl p-4 shadow-hard-md hover:translate-x-px hover:translate-y-px hover:shadow-hard-sm transition-all ${bgClass}`}>
      <div className="flex items-center gap-2 mb-3">
        <div className="w-7 h-7 border-2 border-neutral-950 rounded-lg flex items-center justify-center bg-neutral-950 shrink-0">
          <Icon size={13} className="text-white" />
        </div>
        <p className="text-[10px] font-bold text-neutral-600 uppercase tracking-wider truncate">{label}</p>
        {isUnlimited && (
          <span className="ml-auto text-[9px] bg-white border border-neutral-950 px-1.5 py-0.5 rounded font-bold uppercase shadow-hard-sm shrink-0">
            Ilimitado
          </span>
        )}
      </div>
      <div className="flex items-baseline gap-1.5 mb-3">
        <span className="text-4xl font-black text-neutral-950 tracking-tighter leading-none">{used}</span>
        {!isUnlimited && max > 0 && (
          <span className="text-sm font-bold text-neutral-400">/ {max}</span>
        )}
      </div>
      {!isUnlimited && max > 0 && (
        <div className="h-2.5 w-full bg-neutral-200 rounded-sm border border-neutral-300 overflow-hidden">
          <div
            className={`h-full transition-all duration-700 ${isHigh ? 'bg-accent-orange-deep' : 'bg-neutral-950'}`}
            style={{ width: `${pct}%` }}
          />
        </div>
      )}
    </div>
  )
}

function ProcessRow({ icon: Icon, label, val, total, color }: {
  icon: any
  label: string
  val: number
  total: number
  color: string
}) {
  const pct = total > 0 ? (val / total) * 100 : 0
  return (
    <div>
      <div className="flex justify-between items-center mb-1.5">
        <div className="flex items-center gap-1.5">
          <Icon size={12} className="text-neutral-500" />
          <span className="text-[10px] font-bold text-neutral-600 uppercase tracking-wider">{label}</span>
        </div>
        <span className="text-xs font-black text-neutral-950">{val}</span>
      </div>
      <div className="h-2 w-full bg-neutral-100 rounded-sm border border-neutral-200 overflow-hidden">
        <div className={`h-full ${color} transition-all duration-700`} style={{ width: `${pct}%` }} />
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

  return (
    <div className="p-4 sm:p-6 lg:h-full lg:overflow-hidden flex flex-col gap-4 max-w-7xl mx-auto w-full font-sans">

      {/* Header */}
      <div className="flex items-center gap-4 bg-white border-2 border-neutral-950 rounded-xl px-5 py-4 shadow-hard-md shrink-0">
        <div className="w-1.5 self-stretch bg-primary-400 rounded-full shrink-0" />
        <div>
          <h1 className="text-xl font-black text-neutral-950 tracking-tighter leading-tight">
            {user ? `Hola, ${user.name}` : 'Dashboard'}
          </h1>
          <p className="text-xs font-medium text-neutral-500">Resumen de tus agentes y documentos</p>
        </div>
      </div>

      {isLoading ? (
        <div className="flex-1 flex items-center justify-center">
          <div className="animate-spin text-neutral-400"><Clock size={32} /></div>
        </div>
      ) : (
        <div className="flex-1 grid grid-cols-1 lg:grid-cols-12 gap-4 min-h-0 overflow-y-auto lg:overflow-visible pb-4 lg:pb-0">

          {/* Column 1: stat cards 2×2 */}
          <div className="lg:col-span-4 grid grid-cols-2 gap-3">
            <StatCard icon={FileText}     label="Docs Listos"    value={completed}               bgClass="bg-primary-100" />
            <StatCard icon={Layers}       label="Grupos Activos" value={stats?.active_groups ?? 0} bgClass="bg-white" />
            <StatCard icon={MessageSquare} label="Consultas"     value={attempts}                bgClass="bg-white" />
            <StatCard icon={HardDrive}    label="Espacio"        value={formatBytes(storage)}    bgClass="bg-accent-blue" />
          </div>

          {/* Column 2: quotas */}
          <div className="lg:col-span-4 flex flex-col gap-3">
            <QuotaCard
              label="Cuota de Mensajes"
              used={user?.chat_quota_used ?? 0}
              max={user?.chat_quota ?? 0}
              isUnlimited={user?.chat_quota === 0}
              bgClass="bg-white"
              icon={Zap}
            />
            <QuotaCard
              label="Límite de Documentos"
              used={user?.document_count ?? 0}
              max={user?.max_documents ?? 0}
              isUnlimited={user?.max_documents === 0}
              bgClass="bg-primary-100"
              icon={FileText}
            />
          </div>

          {/* Column 3: processing status */}
          <div className="lg:col-span-4 flex flex-col gap-3">

            {unanswered > 0 && (
              <div className="bg-accent-orange border-2 border-neutral-950 rounded-xl px-4 py-3 flex items-center gap-3 shadow-hard-md shrink-0">
                <AlertTriangle size={16} className="text-neutral-950 shrink-0" />
                <div>
                  <p className="text-sm font-black text-neutral-950 tracking-tighter leading-tight">
                    {unanswered} sin respuesta
                  </p>
                  <p className="text-[10px] font-bold text-neutral-700 mt-0.5">
                    Sube más contexto a tu agente.
                  </p>
                </div>
              </div>
            )}

            <div className="flex-1 bg-white border-2 border-neutral-950 rounded-xl p-4 shadow-hard-md flex flex-col">
              <div className="flex items-center gap-2 mb-4 pb-2 border-b border-neutral-100">
                <p className="text-[10px] font-bold text-neutral-500 uppercase tracking-wider">
                  Estado de Procesamiento
                </p>
                <span className="ml-auto text-[10px] font-black text-neutral-950 bg-neutral-100 border border-neutral-200 px-2 py-0.5 rounded-full">
                  {total}
                </span>
              </div>

              {total > 0 ? (
                <div className="flex-1 flex flex-col justify-around gap-3">
                  <ProcessRow icon={CheckCircle2} label="Completados" val={completed}  total={total} color="bg-primary-400" />
                  <ProcessRow icon={Clock}        label="Procesando"  val={processing} total={total} color="bg-accent-amber-deep" />
                  <ProcessRow icon={HardDrive}    label="En Cola"     val={pending}    total={total} color="bg-neutral-300" />
                  <ProcessRow icon={XCircle}      label="Fallidos"    val={failed}     total={total} color="bg-accent-red-deep" />
                </div>
              ) : (
                <div className="flex-1 flex flex-col items-center justify-center text-center opacity-30">
                  <Users size={28} className="mb-2" />
                  <p className="text-xs font-bold">Sin actividad</p>
                </div>
              )}
            </div>

          </div>

        </div>
      )}
    </div>
  )
}
