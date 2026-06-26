import { createFileRoute, Link } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import api from '../lib/api'
import type { User } from '../types'
import {
  FileText, Layers, HardDrive, MessageSquare,
  AlertTriangle, CheckCircle2, XCircle, Clock, Zap,
  ArrowRight, RefreshCw,
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

function Skeleton({ className = '' }: { className?: string }) {
  return <div className={`animate-pulse bg-neutral-100 rounded-lg ${className}`} />
}

/* ──────────── StatCard ──────────── */

interface StatCardProps {
  icon: React.ComponentType<{ size?: number; className?: string }>
  label: string
  value: string | number
  to?: string
}

function StatCard({ icon: Icon, label, value, to }: StatCardProps) {
  const card = (
    <div className="group border border-neutral-200 rounded-xl p-4 flex items-center gap-4 bg-white hover:border-neutral-400 hover:shadow-sm transition-all">
      <div className="w-10 h-10 rounded-lg bg-neutral-100 flex items-center justify-center shrink-0 group-hover:bg-primary-50 transition-colors">
        <Icon size={18} className="text-neutral-500 group-hover:text-primary-600 transition-colors" strokeWidth={1.5} />
      </div>
      <div className="min-w-0">
        <p className="text-2xl font-bold text-neutral-900 leading-none mb-1">{value}</p>
        <p className="text-xs font-medium text-neutral-500 truncate">{label}</p>
      </div>
    </div>
  )

  if (to) {
    return (
      <Link to={to} className="block focus:outline-none focus:ring-2 focus:ring-primary-400 focus:ring-offset-2 rounded-xl">
        {card}
      </Link>
    )
  }
  return card
}

/* ──────────── QuotaRow ──────────── */

interface QuotaRowProps {
  icon: React.ComponentType<{ size?: number; className?: string }>
  label: string
  used: number
  max: number
  isUnlimited: boolean
}

function QuotaRow({ icon: Icon, label, used, max, isUnlimited }: QuotaRowProps) {
  const pct = isUnlimited ? 100 : (max > 0 ? Math.min((used / max) * 100, 100) : 0)
  const isHigh = pct > 80

  return (
    <div className="flex items-center gap-3">
      <div className="w-8 h-8 rounded-lg bg-neutral-100 flex items-center justify-center shrink-0">
        <Icon size={15} className="text-neutral-500" strokeWidth={1.5} />
      </div>
      <div className="flex-1 min-w-0">
        <div className="flex items-baseline justify-between mb-1">
          <span className="text-sm font-medium text-neutral-700">{label}</span>
          <span className="text-sm font-semibold text-neutral-900 tabular-nums">
            {used}
            {!isUnlimited && max > 0 && (
              <span className="font-normal text-neutral-400"> / {max}</span>
            )}
          </span>
        </div>
        <div className="h-1.5 w-full bg-neutral-100 rounded-full overflow-hidden">
          <div
            className={`h-full rounded-full transition-all duration-500 ${
              isHigh ? 'bg-accent-orange-deep' : 'bg-primary-400'
            }`}
            style={{ width: `${pct}%` }}
          />
        </div>
      </div>
    </div>
  )
}

/* ──────────── ProcessBar ──────────── */

interface ProcessBarProps {
  completed: number
  processing: number
  pending: number
  failed: number
  total: number
}

function ProcessBar({ completed, processing, pending, failed, total }: ProcessBarProps) {
  if (total === 0) return null

  const segments = [
    { value: completed, color: 'bg-primary-400', label: 'Completados' },
    { value: processing, color: 'bg-accent-amber-deep', label: 'Procesando' },
    { value: pending, color: 'bg-neutral-300', label: 'En cola' },
    { value: failed, color: 'bg-accent-red-deep', label: 'Fallidos' },
  ].filter(s => s.value > 0)

  return (
    <div className="space-y-3">
      <div className="flex h-2.5 w-full rounded-full overflow-hidden bg-neutral-100">
        {segments.map((s) => (
          <div
            key={s.color}
            className={`${s.color} transition-all duration-500 first:rounded-l-full last:rounded-r-full`}
            style={{ width: `${(s.value / total) * 100}%` }}
          />
        ))}
      </div>
      <div className="flex flex-wrap gap-x-5 gap-y-1">
        {segments.map((s) => (
          <div key={s.color} className="flex items-center gap-1.5">
            <span className={`w-2 h-2 rounded-full ${s.color} shrink-0`} />
            <span className="text-xs font-medium text-neutral-600">{s.label}</span>
            <span className="text-xs font-semibold text-neutral-900 tabular-nums">{s.value}</span>
          </div>
        ))}
      </div>
    </div>
  )
}

/* ──────────── Main ──────────── */

function DashboardPage() {
  const {
    data: user,
    isError: userError,
    refetch: refetchUser,
  } = useQuery<User>({
    queryKey: ['users', 'me'],
    queryFn: async () => {
      const { data } = await api.get('/users/me')
      return data
    },
    staleTime: 10000,
  })

  const {
    data: stats,
    isLoading,
    isError: statsError,
    refetch: refetchStats,
  } = useQuery<DashboardStats>({
    queryKey: ['users', 'me', 'stats'],
    queryFn: async () => {
      const { data } = await api.get('/users/me/stats')
      return data
    },
    staleTime: 30000,
  })

  const hasError = userError || statsError
  const total = stats?.total_documents ?? 0
  const completed = stats?.by_status?.completed ?? 0
  const failed = stats?.by_status?.failed ?? 0
  const pending = stats?.by_status?.pending ?? 0
  const processing = Math.max(0, total - completed - failed - pending)
  const storage = stats?.storage_used ?? 0
  const unanswered = stats?.unanswered_total ?? 0
  const attempts = stats?.total_attempts ?? 0

  return (
    <div className="max-w-5xl mx-auto w-full space-y-5">

      {/* ── Header ── */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-xl font-semibold text-neutral-900 tracking-tight">
            {user ? `Hola, ${user.name}` : 'Dashboard'}
          </h1>
          <p className="text-sm text-neutral-500 mt-0.5">
            Resumen de tus documentos y grupos
          </p>
        </div>
        {hasError && (
          <button
            onClick={() => { refetchUser(); refetchStats() }}
            className="inline-flex items-center gap-1.5 text-sm font-medium text-neutral-600 hover:text-neutral-900 transition-colors"
          >
            <RefreshCw size={14} />
            Reintentar
          </button>
        )}
      </div>

      {/* ── Error state ── */}
      {hasError && !isLoading && (
        <div className="border border-accent-red-deep/20 bg-accent-red/20 rounded-xl px-5 py-4 flex items-start gap-3">
          <AlertTriangle size={18} className="text-accent-red-deep shrink-0 mt-0.5" />
          <div>
            <p className="text-sm font-semibold text-neutral-900">No se pudieron cargar los datos</p>
            <p className="text-sm text-neutral-600 mt-0.5">
              Verifica tu conexión e intenta de nuevo.
            </p>
          </div>
        </div>
      )}

      {/* ── Loading ── */}
      {isLoading && !hasError && (
        <div className="space-y-5">
          <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
            <Skeleton className="h-[88px]" />
            <Skeleton className="h-[88px]" />
            <Skeleton className="h-[88px]" />
            <Skeleton className="h-[88px]" />
          </div>
          <div className="grid lg:grid-cols-5 gap-4">
            <Skeleton className="lg:col-span-3 h-[120px]" />
            <Skeleton className="lg:col-span-2 h-[120px]" />
          </div>
        </div>
      )}

      {/* ── Content ── */}
      {!isLoading && !hasError && (
        <>
          {/* Stats row */}
          <div className="grid grid-cols-2 lg:grid-cols-4 gap-3">
            <StatCard icon={FileText}     label="Docs listos"    value={completed}               to="/documents" />
            <StatCard icon={Layers}       label="Grupos activos" value={stats?.active_groups ?? 0} to="/groups" />
            <StatCard icon={MessageSquare} label="Consultas"     value={attempts} />
            <StatCard icon={HardDrive}    label="Almacenamiento" value={formatBytes(storage)} />
          </div>

          {/* Bottom row: process + quotas */}
          <div className="grid lg:grid-cols-5 gap-4">
            {/* Process overview */}
            <div className="lg:col-span-3 border border-neutral-200 rounded-xl p-5 bg-white space-y-4">
              <div className="flex items-center justify-between">
                <h2 className="text-sm font-semibold text-neutral-700">
                  Estado de procesamiento
                </h2>
                <span className="text-xs font-medium text-neutral-400 tabular-nums">
                  {total} total
                </span>
              </div>

              {total > 0 ? (
                <ProcessBar
                  completed={completed}
                  processing={processing}
                  pending={pending}
                  failed={failed}
                  total={total}
                />
              ) : (
                <div className="flex flex-col items-center justify-center py-6 text-center">
                  <FileText size={24} className="text-neutral-300 mb-2" />
                  <p className="text-sm font-medium text-neutral-400">
                    Sin documentos todavía
                  </p>
                  <Link
                    to="/documents"
                    className="mt-2 inline-flex items-center gap-1 text-sm font-medium text-primary-600 hover:text-primary-700 transition-colors"
                  >
                    Subir documentos <ArrowRight size={14} />
                  </Link>
                </div>
              )}
            </div>

            {/* Quotas + alerts */}
            <div className="lg:col-span-2 space-y-4">
              {/* Unanswered alert */}
              {unanswered > 0 && (
                <Link
                  to="/documents"
                  className="block border border-accent-amber-deep/20 bg-accent-amber/40 rounded-xl px-4 py-3 hover:border-accent-amber-deep/40 transition-colors"
                >
                  <div className="flex items-start gap-3">
                    <AlertTriangle size={16} className="text-accent-amber-deep shrink-0 mt-0.5" />
                    <div className="min-w-0">
                      <p className="text-sm font-semibold text-neutral-900">
                        {unanswered} consulta{unanswered !== 1 ? 's' : ''} sin respuesta
                      </p>
                      <p className="text-xs text-neutral-600 mt-0.5">
                        Sube más documentos con contexto para tus agentes.
                      </p>
                    </div>
                    <ArrowRight size={14} className="text-neutral-400 shrink-0 mt-1" />
                  </div>
                </Link>
              )}

              {/* Quotas */}
              <div className="border border-neutral-200 rounded-xl p-5 bg-white space-y-4">
                <h2 className="text-sm font-semibold text-neutral-700">Uso del plan</h2>
                <QuotaRow
                  icon={Zap}
                  label="Mensajes"
                  used={user?.chat_quota_used ?? 0}
                  max={user?.chat_quota ?? 0}
                  isUnlimited={user?.chat_quota === 0}
                />
                <QuotaRow
                  icon={FileText}
                  label="Documentos"
                  used={user?.document_count ?? 0}
                  max={user?.max_documents ?? 0}
                  isUnlimited={user?.max_documents === 0}
                />
              </div>
            </div>
          </div>
        </>
      )}
    </div>
  )
}
