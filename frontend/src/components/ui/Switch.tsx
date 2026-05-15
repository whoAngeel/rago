import React from 'react'

interface SwitchProps {
  checked: boolean
  onChange: (checked: boolean) => void
  disabled?: boolean
  className?: string
}

export function Switch({ checked, onChange, disabled, className = '' }: SwitchProps) {
  const handleClick = (e: React.MouseEvent) => {
    e.stopPropagation()
    if (!disabled) {
      onChange(!checked)
    }
  }

  return (
    <button
      type='button'
      onClick={handleClick}
      disabled={disabled}
      className={`w-14 h-7 border-2 border-neutral-950 rounded-btn flex items-center transition-all cursor-pointer disabled:opacity-50 disabled:pointer-events-none ${checked ? 'bg-primary-400 justify-end' : 'bg-neutral-200 justify-start'} ${className}`}
    >
      <div className='w-5 h-5 bg-white border-2 border-neutral-950 rounded-btn mx-0.5 shadow-hard-sm' />
    </button>
  )
}
