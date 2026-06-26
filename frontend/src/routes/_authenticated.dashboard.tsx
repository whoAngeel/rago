import { createFileRoute, Link } from '@tanstack/react-router'
import { useQuery } from '@tanstack/react-query'
import api from '../lib/api'
import type { User } from '../types'
import {
  FileText, Layers, HardDrive, MessageSquare,
  AlertTriangle, Zap,
  ArrowRight, RefreshCw, Loader,
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

/* ──────────── DonutChart ──────────── */

interface DonutSegment {
  value: number
  color: string
  label: string
}

interface DonutChartProps {
  segments: DonutSegment[]
  total: number
  centerValue: string
  centerLabel: string
  size?: number
  strokeWidth?: number
}

function DonutChart({
  segments,
  total,
  centerValue,
  centerLabel,
  size = 160,
  strokeWidth = 18,
}: DonutChartProps) {
  const radius = (size - strokeWidth) / 2
  const circumference = 2 * Math.PI * radius
  const visible = segments.filter((s) => s.value > 0)

  let cumulative = 0

  return (
    <div className="flex items-center gap-6 sm:gap-8">
      <div className="relative shrink-0" style={{ width: size, height: size }}>
        <svg
          viewBox={`0 0 ${size} ${size}`}
          className="transform -rotate-90 w-full h-full"
        >
          {/* background ring */}
          <circle
            cx={size / 2}
            cy={size / 2}
            r={radius}
            fill="none"
            stroke="#f5f5f5"
            strokeWidth={strokeWidth}
          />
          {visible.map((seg) => {
            const pct = seg.value / total
            const dashLength = pct * circumference
            const offset = cumulative * circumference
            cumulative += pct
            return (
              <circle
                key={seg.label}
                cx={size / 2}
                cy={size / 2}
                r={radius}
                fill="none"
                stroke={seg.color}
                strokeWidth={strokeWidth}
                strokeDasharray={`${dashLength} ${circumference - dashLength}`}
                strokeDashoffset={-offset}
                strokeLinecap={pct > 0.05 ? 'round' : 'butt'}
                className="transition-all duration-700"
              />
            )
          })}
        </svg>
        <div className="absolute inset-0 flex flex-col items-center justify-center">
          <span className="text-2xl sm:text-3xl font-bold text-neutral-900 leading-none">
            {centerValue}
          </span>
          <span className="text-[11px] font-medium text-neutral-500 mt-1">
            {centerLabel}
          </span>
        </div>
      </div>

      <div className="space-y-2 min-w-0">
        {visible.map((seg) => (
          <div key={seg.label} className="flex items-center gap-2.5">
            <span
              className="w-2.5 h-2.5 rounded-full shrink-0"
              style={{ backgroundColor: seg.color }}
            />
            <span className="text-sm text-neutral-600 truncate">{seg.label}</span>
            <span className="text-sm font-semibold text-neutral-900 tabular-nums ml-auto">
              {seg.value}
            </span>
          </div>
        ))}
      </div>
    </div>
  )
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
        <div className="flex items-baseline justify-between mb-1.5">
          <span className="text-sm font-medium text-neutral-700">{label}</span>
          <span className="text-sm font-semibold text-neutral-900 tabular-nums">
            {used}
            {!isUnlimited && max > 0 && (
              <span className="font-normal text-neutral-400"> / {max}</span>
            )}
          </span>
        </div>
        <div className="h-2 w-full bg-neutral-100 rounded-full overflow-hidden">
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
  const activeGroups = stats?.active_groups ?? 0
  const totalGroups = stats?.total_groups ?? 0

  const successRate = total > 0 ? Math.round((completed / total) * 100) : 0

  const donutSegments: DonutSegment[] = [
    { value: completed, color: '#a3e635', label: 'Completados' },
    { value: processing, color: '#fbbf25', label: 'Procesando' },
    { value: pending, color: '#d4d4d4', label: 'En cola' },
    { value: failed, color: '#f4405e', label: 'Fallidos' },
  ]

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
          <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-3">
            <Skeleton className="h-[88px]" />
            <Skeleton className="h-[88px]" />
            <Skeleton className="h-[88px]" />
            <Skeleton className="h-[88px]" />
            <Skeleton className="h-[88px]" />
          </div>
          <div className="grid lg:grid-cols-5 gap-4">
            <Skeleton className="lg:col-span-3 h-[200px]" />
            <Skeleton className="lg:col-span-2 h-[200px]" />
          </div>
        </div>
      )}

      {/* ── Content ── */}
      {!isLoading && !hasError && (
        <>
          {/* Stats row — 5 cards */}
          <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-3">
            <StatCard
              icon={FileText}
              label="Docs listos"
              value={completed}
              to="/documents"
            />
            <StatCard
              icon={Loader}
              label="En proceso"
              value={processing + pending}
              to={processing + pending > 0 ? '/documents' : undefined}
            />
            <StatCard
              icon={Layers}
              label="Grupos activos"
              value={`${activeGroups}`}
              to="/groups"
            />
            <StatCard
              icon={MessageSquare}
              label="Consultas"
              value={attempts}
            />
            <StatCard
              icon={HardDrive}
              label="Almacenamiento"
              value={formatBytes(storage)}
            />
          </div>

          {/* Bottom row: donut chart + quotas */}
          <div className="grid lg:grid-cols-5 gap-4">
            {/* Donut chart — document status */}
            <div className="lg:col-span-3 border border-neutral-200 rounded-xl p-5 sm:p-6 bg-white">
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-sm font-semibold text-neutral-700">
                  Estado de documentos
                </h2>
                <span className="text-xs font-medium text-neutral-400 tabular-nums">
                  {total} total
                </span>
              </div>

              {total > 0 ? (
                <DonutChart
                  segments={donutSegments}
                  total={total}
                  centerValue={`${successRate}%`}
                  centerLabel="completado"
                />
              ) : (
                <div className="flex flex-col items-center justify-center py-10 text-center">
                  <FileText size={28} className="text-neutral-300 mb-3" />
                  <p className="text-sm font-medium text-neutral-500">
                    Sin documentos todavía
                  </p>
                  <Link
                    to="/documents"
                    className="mt-2 inline-flex items-center gap-1 text-sm font-medium text-primary-600 hover:text-primary-700 transition-colors"
                  >
                    Subir el primero <ArrowRight size={14} />
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
                        Agrega más documentos con contexto.
                      </p>
                    </div>
                    <ArrowRight size={14} className="text-neutral-400 shrink-0 mt-1" />
                  </div>
                </Link>
              )}

              {/* Quotas */}
              <div className="border border-neutral-200 rounded-xl p-5 bg-white space-y-5">
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

                {totalGroups > 0 && (
                  <div className="pt-3 border-t border-neutral-100">
                    <div className="flex items-center justify-between text-sm">
                      <span className="text-neutral-500">Total grupos</span>
                      <span className="font-semibold tabular-nums">{totalGroups}</span>
                    </div>
                    <div className="flex items-center justify-between text-sm mt-1">
                      <span className="text-neutral-500">Inactivos</span>
                      <span className="font-medium text-neutral-400 tabular-nums">
                        {Math.max(0, totalGroups - activeGroups)}
                      </span>
                    </div>
                  </div>
                )}
              </div>
            </div>
          </div>
        </>
      )}
    </div>
  )
}
