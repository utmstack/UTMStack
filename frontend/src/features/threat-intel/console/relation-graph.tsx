import { useEffect, useMemo, useRef, useState } from 'react'
import { useThemeContext } from '@/app/providers'
import { cn } from '@/shared/lib/utils'
import { useTranslation } from 'react-i18next'
import type { RelationNode, RelationsResult } from './types'

interface SimNode {
  id: string
  label: string
  type: string
  reputation: number | null
  x: number
  y: number
  vx: number
  vy: number
  degree: number
}

interface RelationGraphProps {
  graph: RelationsResult
  rootId?: string
  onSelect?(node: RelationNode): void
  className?: string
  height?: number
}

const REPULSION = 9000
const SPRING = 0.012
const IDEAL_EDGE_LENGTH = 90
const DAMPING = 0.82
const CENTER_PULL = 0.006

interface GraphPalette {
  danger: string
  warning: string
  safe: string
  unknown: string
  edge: string
  label: string
  root: string
  hover: string
}

function readPalette(dark: boolean): GraphPalette {
  return {
    danger: dark ? 'hsl(0 72% 58%)' : 'hsl(0 82% 57%)',
    warning: 'hsl(38 92% 50%)',
    safe: dark ? 'hsl(142 70% 45%)' : 'hsl(142 71% 35%)',
    unknown: dark ? 'hsl(220 10% 58%)' : 'hsl(218 16% 44%)',
    edge: dark ? 'hsl(220 10% 58% / 0.28)' : 'hsl(218 16% 44% / 0.28)',
    label: dark ? 'hsl(220 10% 58%)' : 'hsl(218 16% 44%)',
    root: dark ? 'hsl(211 100% 50%)' : 'hsl(211 100% 47%)',
    hover: dark ? 'hsl(220 15% 93%)' : 'hsl(222 42% 10%)',
  }
}

function reputationColor(reputation: number | null, palette: GraphPalette): string {
  if (reputation === null) return palette.unknown
  if (reputation <= -1) return palette.danger
  if (reputation < 1) return palette.warning
  return palette.safe
}

export function RelationGraph({ graph, rootId, onSelect, className, height = 420 }: RelationGraphProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const wrapperRef = useRef<HTMLDivElement>(null)
  const frameRef = useRef<number>(0)
  const nodesRef = useRef<SimNode[]>([])
  const hoverRef = useRef<string | null>(null)
  const viewRef = useRef({ scale: 1, offsetX: 0, offsetY: 0 })
  const dragRef = useRef<{ x: number; y: number } | null>(null)
  const { t } = useTranslation()
  const { theme } = useThemeContext()
  const [hovered, setHovered] = useState<string | null>(null)
  const palette = useMemo(() => readPalette(theme === 'dark'), [theme])

  const { nodes, edges } = useMemo(() => {
    const degree = new Map<string, number>()
    for (const edge of graph.edges) {
      degree.set(edge.source, (degree.get(edge.source) ?? 0) + 1)
      degree.set(edge.target, (degree.get(edge.target) ?? 0) + 1)
    }
    const simNodes: SimNode[] = graph.nodes.map((n, i) => {
      const angle = (i / Math.max(graph.nodes.length, 1)) * Math.PI * 2
      const radius = 120 + (i % 5) * 24
      return {
        id: n.id,
        label: n.value || n.id,
        type: n.type,
        reputation: n.reputation,
        x: Math.cos(angle) * radius,
        y: Math.sin(angle) * radius,
        vx: 0,
        vy: 0,
        degree: degree.get(n.id) ?? 0,
      }
    })
    const index = new Map(simNodes.map((n) => [n.id, n]))
    const simEdges = graph.edges
      .map((e) => ({ a: index.get(e.source), b: index.get(e.target) }))
      .filter((e): e is { a: SimNode; b: SimNode } => Boolean(e.a && e.b))
    return { nodes: simNodes, edges: simEdges }
  }, [graph])

  useEffect(() => {
    nodesRef.current = nodes
    viewRef.current = { scale: 1, offsetX: 0, offsetY: 0 }
  }, [nodes])

  useEffect(() => {
    const canvas = canvasRef.current
    const wrapper = wrapperRef.current
    if (!canvas || !wrapper) return
    const ctx = canvas.getContext('2d')
    if (!ctx) return

    let width = wrapper.clientWidth
    let alpha = 1

    const resize = () => {
      width = wrapper.clientWidth
      const dpr = window.devicePixelRatio || 1
      canvas.width = width * dpr
      canvas.height = height * dpr
      canvas.style.width = `${width}px`
      canvas.style.height = `${height}px`
      ctx.setTransform(dpr, 0, 0, dpr, 0, 0)
    }
    resize()
    const observer = new ResizeObserver(resize)
    observer.observe(wrapper)

    const tick = () => {
      const sim = nodesRef.current

      if (alpha > 0.005) {
        for (let i = 0; i < sim.length; i++) {
          const a = sim[i]
          for (let j = i + 1; j < sim.length; j++) {
            const b = sim[j]
            let dx = a.x - b.x
            let dy = a.y - b.y
            let distSq = dx * dx + dy * dy
            if (distSq < 0.01) {
              dx = (Math.random() - 0.5) * 0.1
              dy = (Math.random() - 0.5) * 0.1
              distSq = dx * dx + dy * dy
            }
            const force = REPULSION / distSq
            const dist = Math.sqrt(distSq)
            const fx = (dx / dist) * force
            const fy = (dy / dist) * force
            a.vx += fx
            a.vy += fy
            b.vx -= fx
            b.vy -= fy
          }
        }

        for (const { a, b } of edges) {
          const dx = b.x - a.x
          const dy = b.y - a.y
          const dist = Math.sqrt(dx * dx + dy * dy) || 1
          const force = (dist - IDEAL_EDGE_LENGTH) * SPRING
          const fx = (dx / dist) * force
          const fy = (dy / dist) * force
          a.vx += fx
          a.vy += fy
          b.vx -= fx
          b.vy -= fy
        }

        for (const n of sim) {
          n.vx -= n.x * CENTER_PULL
          n.vy -= n.y * CENTER_PULL
          n.vx *= DAMPING
          n.vy *= DAMPING
          n.x += n.vx * alpha
          n.y += n.vy * alpha
        }
        alpha *= 0.994
      }

      const { scale, offsetX, offsetY } = viewRef.current
      ctx.clearRect(0, 0, width, height)
      ctx.save()
      ctx.translate(width / 2 + offsetX, height / 2 + offsetY)
      ctx.scale(scale, scale)

      ctx.strokeStyle = palette.edge
      ctx.lineWidth = 1
      for (const { a, b } of edges) {
        ctx.beginPath()
        ctx.moveTo(a.x, a.y)
        ctx.lineTo(b.x, b.y)
        ctx.stroke()
      }

      for (const n of sim) {
        const isRoot = n.id === rootId
        const isHover = n.id === hoverRef.current
        const radius = Math.min(5 + n.degree * 1.4, 14) * (isRoot ? 1.35 : 1)
        ctx.beginPath()
        ctx.arc(n.x, n.y, radius, 0, Math.PI * 2)
        ctx.fillStyle = reputationColor(n.reputation, palette)
        ctx.fill()
        if (isRoot || isHover) {
          ctx.strokeStyle = isRoot ? palette.root : palette.hover
          ctx.lineWidth = 2
          ctx.stroke()
        }
        if (isHover || isRoot || n.degree > 2 || sim.length <= 18) {
          ctx.font = '10px "JetBrains Mono", monospace'
          ctx.fillStyle = palette.label
          ctx.textAlign = 'center'
          const label = n.label.length > 26 ? `${n.label.slice(0, 24)}…` : n.label
          ctx.fillText(label, n.x, n.y - radius - 5)
        }
      }
      ctx.restore()

      frameRef.current = requestAnimationFrame(tick)
    }

    frameRef.current = requestAnimationFrame(tick)
    return () => {
      cancelAnimationFrame(frameRef.current)
      observer.disconnect()
    }
  }, [edges, height, rootId, palette])

  const toGraphSpace = (clientX: number, clientY: number) => {
    const canvas = canvasRef.current
    if (!canvas) return null
    const rect = canvas.getBoundingClientRect()
    const { scale, offsetX, offsetY } = viewRef.current
    return {
      x: (clientX - rect.left - rect.width / 2 - offsetX) / scale,
      y: (clientY - rect.top - rect.height / 2 - offsetY) / scale,
    }
  }

  const nodeAt = (clientX: number, clientY: number): SimNode | null => {
    const point = toGraphSpace(clientX, clientY)
    if (!point) return null
    let best: SimNode | null = null
    let bestDist = Infinity
    for (const n of nodesRef.current) {
      const dx = n.x - point.x
      const dy = n.y - point.y
      const dist = Math.sqrt(dx * dx + dy * dy)
      const radius = Math.min(5 + n.degree * 1.4, 14) + 6
      if (dist < radius && dist < bestDist) {
        best = n
        bestDist = dist
      }
    }
    return best
  }

  return (
    <div ref={wrapperRef} className={cn('relative w-full', className)}>
      <canvas
        ref={canvasRef}
        className="cursor-grab rounded-md border border-border bg-card active:cursor-grabbing"
        onMouseDown={(e) => {
          dragRef.current = { x: e.clientX, y: e.clientY }
        }}
        onMouseUp={() => {
          dragRef.current = null
        }}
        onMouseLeave={() => {
          dragRef.current = null
          hoverRef.current = null
          setHovered(null)
        }}
        onMouseMove={(e) => {
          if (dragRef.current) {
            viewRef.current.offsetX += e.clientX - dragRef.current.x
            viewRef.current.offsetY += e.clientY - dragRef.current.y
            dragRef.current = { x: e.clientX, y: e.clientY }
            return
          }
          const node = nodeAt(e.clientX, e.clientY)
          hoverRef.current = node?.id ?? null
          setHovered(node?.id ?? null)
        }}
        onWheel={(e) => {
          const next = viewRef.current.scale * (e.deltaY < 0 ? 1.12 : 0.89)
          viewRef.current.scale = Math.min(Math.max(next, 0.25), 4)
        }}
        onClick={(e) => {
          const node = nodeAt(e.clientX, e.clientY)
          if (!node || !onSelect) return
          onSelect({ id: node.id, type: node.type, value: node.label, reputation: node.reputation })
        }}
      />
      <div className="absolute left-2 top-2 flex items-center gap-3 font-mono text-[10px] text-muted-foreground">
        <span>{t('tiAnalyst.graph.nodes_one', { count: graph.nodes.length })}</span>
        <span>{t('tiAnalyst.graph.edges_one', { count: graph.edges.length })}</span>
        {hovered && <span className="text-primary">{t('tiAnalyst.graph.clickToPivot')}</span>}
      </div>
      <div className="absolute right-2 top-2 flex items-center gap-2 font-mono text-[10px] text-muted-foreground">
        <Legend className="bg-destructive" label={t('tiAnalyst.graph.malicious')} />
        <Legend className="bg-amber-500" label={t('tiAnalyst.graph.unknown')} />
        <Legend className="bg-emerald-500" label={t('tiAnalyst.graph.benign')} />
      </div>
    </div>
  )
}

function Legend({ className, label }: { className: string; label: string }) {
  return (
    <span className="flex items-center gap-1">
      <span className={cn('h-2 w-2 rounded-full', className)} />
      {label}
    </span>
  )
}
