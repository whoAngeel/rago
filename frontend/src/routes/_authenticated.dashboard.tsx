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
  return <div className={`animate-pulse bg-neutral-200 rounded-lg border-2 border-neutral-950 ${className}`} />
}

/* ──────────── IconBox (colorize + overdrive) ──────────── */

function IconBox({
  children,
  size = 'md',
}: {
  children: React.ReactNode
  size?: 'sm' | 'md'
}) {
  const sizeClass = size === 'sm'
    ? 'w-7 h-7'
    : 'w-10 h-10'

  return (
    <span
      className={`
        ${sizeClass} grid place-items-center shrink-0
        bg-lime border border-ink rounded-lg
        shadow-brutal-sm
        transition-shadow duration-200
        group-hover:shadow-[1px_1px_0_0_#0a0a0d,0_0_16px_4px_rgba(163,230,53,0.30)]
      `}
    >
      {children}
    </span>
  )
}

/* ──────────── StatCard ──────────── */

interface StatCardProps {
  icon: React.ComponentType<{ size?: number; className?: string }>
  label: string
  value: string | number
  to?: string
  bgClass?: string
}

function StatCard({ icon: Icon, label, value, to, bgClass = 'bg-white' }: StatCardProps) {
  const card = (
    <div className={`group shadow-hover border-2 border-neutral-950 rounded-xl p-4 flex items-center gap-4 ${bgClass} shadow-hard-md`}>
      <IconBox>
        <Icon size={18} className="text-black" strokeWidth={2.25} />
      </IconBox>
      <div className="min-w-0">
        <p className="text-2xl font-black text-neutral-950 tracking-tighter leading-none mb-1">{value}</p>
        <p className="text-[11px] font-bold text-neutral-500 uppercase tracking-wider leading-tight">{label}</p>
      </div>
    </div>
  )

  if (to) {
    return (
      <Link to={to} className="block">
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
  strokeWidth = 22,
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
          <circle
            cx={size / 2}
            cy={size / 2}
            r={radius}
            fill="none"
            stroke="#e5e5e5"
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
                strokeLinecap={pct > 0.08 ? 'round' : 'butt'}
                className="transition-all duration-700"
              />
            )
          })}
        </svg>
        <div className="absolute inset-0 flex flex-col items-center justify-center">
          <span className="text-3xl font-black tabular-nums tracking-tighter leading-none"
            style={{ color: '#65a30e' }}
          >
            {centerValue}
          </span>
          <span className="text-[11px] font-bold text-neutral-500 uppercase tracking-wider mt-1">
            {centerLabel}
          </span>
        </div>
      </div>

      <div className="space-y-2 min-w-0">
        {visible.map((seg) => (
          <div key={seg.label} className="flex items-center gap-2">
            <span
              className="w-3 h-3 rounded-sm shrink-0 border border-neutral-950"
              style={{ backgroundColor: seg.color }}
            />
            <span className="text-sm font-bold text-neutral-700 truncate">{seg.label}</span>
            <span className="text-sm font-black text-neutral-950 tabular-nums ml-auto">
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
    <div>
      <div className="flex items-center justify-between mb-2">
        <div className="flex items-center gap-2">
          <IconBox size="sm">
            <Icon size={13} className="text-black" strokeWidth={2.25} />
          </IconBox>
          <span className="text-sm font-bold text-neutral-700">{label}</span>
        </div>
        <span className="text-sm font-black text-neutral-950 tabular-nums">
          {used}
          {!isUnlimited && max > 0 && (
            <span className="font-bold text-neutral-400"> / {max}</span>
          )}
          {isUnlimited && (
            <span className="ml-1.5 text-[9px] bg-white border border-neutral-950 px-1.5 py-0.5 rounded font-bold uppercase shadow-hard-sm">
              Ilimitado
            </span>
          )}
        </span>
      </div>
      {!isUnlimited && max > 0 && (
        <div className="h-2.5 w-full bg-neutral-200 rounded-sm border border-neutral-300 overflow-hidden">
          <div
            className={`h-full transition-all duration-500 ${
              isHigh ? 'bg-accent-orange-deep' : 'bg-neutral-950'
            }`}
            style={{ width: `${pct}%` }}
          />
        </div>
      )}
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

  const successRate = total > 0 ? Math.round((completed / total) * 100) : 0

  const donutSegments: DonutSegment[] = [
    { value: completed, color: '#84cc17', label: 'Completados' },
    { value: processing, color: '#fbbf25', label: 'Procesando' },
    { value: pending, color: '#d4d4d4', label: 'En cola' },
    { value: failed, color: '#f4405e', label: 'Fallidos' },
  ]

  return (
    <div className="max-w-5xl mx-auto w-full space-y-5">

      {/* ── Header ── */}
      <div className="flex items-center gap-4 bg-white border-2 border-neutral-950 rounded-xl px-5 py-4 shadow-hard-md">
        <div className="w-1.5 self-stretch bg-primary-400 rounded-full shrink-0" />
        <div className="flex-1 flex items-center justify-between">
          <div>
            <h1 className="text-xl font-black text-neutral-950 tracking-tighter leading-tight">
              {user ? `Hola, ${user.name}` : 'Dashboard'}
            </h1>
            <p className="text-xs font-medium text-neutral-500">Resumen de tus documentos y grupos</p>
          </div>
          {hasError && (
            <button
              onClick={() => { refetchUser(); refetchStats() }}
              className="press-brutal inline-flex items-center gap-1.5 bg-white border-2 border-neutral-950 rounded-sm shadow-brutal px-3 py-1.5 text-xs font-bold"
            >
              <RefreshCw size={13} />
              Reintentar
            </button>
          )}
        </div>
      </div>

      {/* ── Error state ── */}
      {hasError && !isLoading && (
        <div className="bg-accent-red border-2 border-neutral-950 rounded-xl px-5 py-4 flex items-start gap-3 shadow-hard-md">
          <AlertTriangle size={18} className="text-neutral-950 shrink-0 mt-0.5" />
          <div>
            <p className="text-sm font-black text-neutral-950 tracking-tighter">
              Error al cargar el dashboard
            </p>
            <p className="text-xs font-bold text-neutral-700 mt-0.5">
              Revisa tu conexión y vuelve a intentarlo.
            </p>
          </div>
        </div>
      )}

      {/* ── Loading ── */}
      {isLoading && !hasError && (
        <div className="space-y-4">
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
          {/* Stats row */}
          <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-3">
            <StatCard
              icon={FileText}
              label="Documentos listos"
              value={completed}
              to="/documents"
              bgClass="bg-primary-100"
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
              value={activeGroups}
              to="/groups"
              bgClass="bg-accent-blue"
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
              bgClass="bg-accent-purple"
            />
          </div>

          {/* Bottom row: donut + quotas */}
          <div className="grid lg:grid-cols-5 gap-4">
            {/* Donut chart */}
            <div className="lg:col-span-3 bg-white border-2 border-neutral-950 rounded-xl p-5 sm:p-6 shadow-hard-lg">
              <div className="flex items-center justify-between mb-4 pb-2 border-b border-neutral-100">
                <h2 className="text-[10px] font-bold text-neutral-500 uppercase tracking-wider">
                  Estado de documentos
                </h2>
                <span className="text-[10px] font-black text-neutral-950 bg-neutral-100 border border-neutral-200 px-2 py-0.5 rounded-full tabular-nums">
                  {total}
                </span>
              </div>

              {total > 0 ? (
                <DonutChart
                  segments={donutSegments}
                  total={total}
                  centerValue={`${successRate}%`}
                  centerLabel="tasa de éxito"
                />
              ) : (
                <div className="flex flex-col items-center justify-center py-10 text-center">
                  <FileText size={28} className="text-neutral-300 mb-3" />
                  <p className="text-sm font-bold text-neutral-500">
                    Aún no hay documentos
                  </p>
                  <Link
                    to="/documents"
                    className="mt-3 press-brutal inline-flex items-center gap-1.5 bg-lime border-2 border-neutral-950 rounded-sm shadow-brutal px-4 py-2 text-sm font-bold"
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
                  className="block shadow-hover bg-accent-orange border-2 border-neutral-950 rounded-xl px-4 py-3 shadow-hard-md"
                >
                  <div className="flex items-start gap-3">
                    <AlertTriangle size={16} className="text-neutral-950 shrink-0 mt-0.5" />
                    <div className="min-w-0 flex-1">
                      <p className="text-sm font-black text-neutral-950 tracking-tighter leading-tight">
                        {unanswered} sin respuesta
                      </p>
                      <p className="text-[10px] font-bold text-neutral-700 mt-0.5">
                        Agrega más documentos con contexto.
                      </p>
                    </div>
                    <ArrowRight size={14} className="text-neutral-950 shrink-0 mt-1" />
                  </div>
                </Link>
              )}

              {/* Quotas */}
              <div className="bg-white border-2 border-neutral-950 rounded-xl p-5 shadow-hard-md space-y-5">
                <div className="flex items-center gap-2 pb-2 border-b border-neutral-100">
                  <p className="text-[10px] font-bold text-neutral-500 uppercase tracking-wider">
                    Uso del plan
                  </p>
                </div>

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
