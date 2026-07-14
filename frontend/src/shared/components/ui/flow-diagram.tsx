import type { ReactNode } from 'react'

// Dependency-free building blocks for a horizontal "A → B → C" flow diagram —
// rounded node cards connected by an animated SVG connector (native
// <animateMotion>, no JS animation library). Originally built for the
// integrations ForwarderGuide; shared here so other pipeline-style diagrams
// (e.g. the parsing-filters / alerting-rules dataType flow) can reuse it.

export const TONES = {
  neutral: 'border-border bg-card',
  accent: 'border-primary/40 bg-primary/[0.06] text-primary',
  brand: 'border-emerald-500/40 bg-emerald-500/[0.06] text-emerald-600 dark:text-emerald-400',
} as const

export function FlowNode({
  icon,
  title,
  sub,
  tone,
  onClick,
  size = 'sm',
}: {
  icon: ReactNode
  title: string
  sub: string
  tone: keyof typeof TONES
  onClick?: () => void
  /** 'lg' is for standalone hero diagrams (e.g. the filters/rules pipeline page) — 'sm' (default) keeps every existing inline setup-guide diagram unchanged. */
  size?: 'sm' | 'lg'
}) {
  const sizeCls =
    size === 'lg'
      ? 'w-[112px] gap-1.5 p-3.5 sm:w-[140px]'
      : 'w-[92px] gap-1 p-2.5 sm:w-[112px]'
  const titleCls = size === 'lg' ? 'text-[13px]' : 'text-[11px]'
  const subCls = size === 'lg' ? 'text-[10.5px]' : 'text-[9px]'
  const className = `flex shrink-0 flex-col items-center rounded-lg border text-center ${sizeCls} ${TONES[tone]} ${
    onClick ? 'cursor-pointer transition-transform hover:scale-[1.03]' : ''
  }`
  const content = (
    <>
      {icon}
      <span className={`font-semibold leading-tight text-foreground ${titleCls}`}>{title}</span>
      <span className={`leading-tight text-muted-foreground ${subCls}`}>{sub}</span>
    </>
  )
  if (onClick) {
    return (
      <button type="button" onClick={onClick} className={className}>
        {content}
      </button>
    )
  }
  return <div className={className}>{content}</div>
}

// A connector with little packets flowing along the line (SVG animateMotion, no deps).
export function FlowEdge({
  label,
  size = 'sm',
  orientation = 'horizontal',
}: {
  label?: string
  size?: 'sm' | 'lg'
  /** 'vertical' stacks nodes top-to-bottom instead of left-to-right. */
  orientation?: 'horizontal' | 'vertical'
}) {
  if (orientation === 'vertical') {
    const heightCls = size === 'lg' ? 'h-10 sm:h-14' : 'h-6 sm:h-8'
    return (
      <div className="flex shrink-0 items-center justify-center">
        {label && (
          <span className="mr-1 text-[8px] font-medium uppercase tracking-wide text-muted-foreground">
            {label}
          </span>
        )}
        <svg viewBox="0 0 12 64" className={`w-3 ${heightCls}`} aria-hidden="true">
          <line
            x1="6"
            y1="2"
            x2="6"
            y2="62"
            className="stroke-border"
            strokeWidth="2"
            strokeDasharray="2 3"
            strokeLinecap="round"
          />
          <circle r="2.6" className="fill-primary">
            <animateMotion dur="1.6s" repeatCount="indefinite" path="M6 2 V62" />
          </circle>
          <circle r="2.6" className="fill-primary">
            <animateMotion dur="1.6s" begin="0.8s" repeatCount="indefinite" path="M6 2 V62" />
          </circle>
        </svg>
      </div>
    )
  }
  const widthCls = size === 'lg' ? 'w-14 sm:w-24' : 'w-10 sm:w-16'
  return (
    <div className="flex shrink-0 flex-col items-center justify-center">
      {label && (
        <span className="mb-1 text-[8px] font-medium uppercase tracking-wide text-muted-foreground">
          {label}
        </span>
      )}
      <svg viewBox="0 0 64 12" className={`h-3 ${widthCls}`} aria-hidden="true">
        <line
          x1="2"
          y1="6"
          x2="62"
          y2="6"
          className="stroke-border"
          strokeWidth="2"
          strokeDasharray="2 3"
          strokeLinecap="round"
        />
        <circle r="2.6" className="fill-primary">
          <animateMotion dur="1.6s" repeatCount="indefinite" path="M2 6 H62" />
        </circle>
        <circle r="2.6" className="fill-primary">
          <animateMotion dur="1.6s" begin="0.8s" repeatCount="indefinite" path="M2 6 H62" />
        </circle>
      </svg>
    </div>
  )
}
