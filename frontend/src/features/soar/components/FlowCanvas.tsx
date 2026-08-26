import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import {
  Background,
  Controls,
  ReactFlow,
  ReactFlowProvider,
  addEdge,
  useEdgesState,
  useNodesState,
  useReactFlow,
  type Connection,
  type Edge,
  type EdgeChange,
  type Node,
  type NodeChange,
  type NodeTypes,
} from '@xyflow/react'
import '@xyflow/react/dist/style.css'
import { ChevronLeft, ChevronRight, PanelLeft, PanelRight } from 'lucide-react'
import { cn } from '@/shared/lib/utils'
import { useTheme } from '@/shared/hooks/useTheme'
import type { FlowNode, NodeKind } from '../types/soar.types'
import { NodePalette } from './NodePalette'
import { NodeInspector } from './NodeInspector'
import { DAGNode } from './nodes/DAGNode'
import { TriggerNode } from './nodes/TriggerNode'

const NODE_TYPES: NodeTypes = { dag: DAGNode as unknown as NodeTypes[string], trigger: TriggerNode as unknown as NodeTypes[string] }
const TRIGGER_ID = '__trigger__'

interface Props {
  roots: string[]
  nodes: Record<string, FlowNode>
  readOnly?: boolean
  onChange: (patch: { roots: string[]; nodes: Record<string, FlowNode> }) => void
}

/** Node-red style DAG editor for a SOAR flow. Nodes come from the flow's
 *  `nodes` map; edges are derived from each node's `onSuccess`/`onError`. A
 *  virtual trigger node hosts the roots list — dragging from it adds a root. */
export function FlowCanvas(props: Props) {
  return (
    <ReactFlowProvider>
      <FlowCanvasInner {...props} />
    </ReactFlowProvider>
  )
}

function FlowCanvasInner({ roots, nodes, readOnly, onChange }: Props) {
  const { t } = useTranslation()
  const wrapperRef = useRef<HTMLDivElement>(null)
  const { screenToFlowPosition } = useReactFlow()
  const { theme } = useTheme()

  // Layout positions are stashed by node id so they survive re-derivation
  // whenever the flow model updates. Fresh nodes get a top-down default.
  const layoutRef = useRef<Record<string, { x: number; y: number }>>({})
  const [selectedId, setSelectedId] = useState<string | null>(null)
  const [paletteOpen, setPaletteOpen] = useState(true)
  const [inspectorOpen, setInspectorOpen] = useState(true)

  const { rfNodes, rfEdges } = useMemo(() => {
    const posFor = (id: string, fallback: { x: number; y: number }) =>
      layoutRef.current[id] ?? (layoutRef.current[id] = fallback)

    const rfN: Node[] = [
      {
        id: TRIGGER_ID,
        type: 'trigger',
        position: posFor(TRIGGER_ID, { x: 120, y: 0 }),
        data: {},
        draggable: !readOnly,
        deletable: false,
      },
    ]
    let i = 0
    for (const [id, n] of Object.entries(nodes)) {
      rfN.push({
        id,
        type: 'dag',
        position: posFor(id, { x: (i % 3) * 260, y: 180 + Math.floor(i / 3) * 180 }),
        data: { nodeId: id, ...n } as unknown as Record<string, unknown>,
        selected: id === selectedId,
        draggable: !readOnly,
      })
      i++
    }

    const rfE: Edge[] = []
    roots.forEach((r) =>
      rfE.push({
        id: `${TRIGGER_ID}->${r}`,
        source: TRIGGER_ID,
        sourceHandle: 'trigger',
        target: r,
        animated: true,
        style: { stroke: '#f59e0b', strokeWidth: 2 },
      }),
    )
    for (const [id, n] of Object.entries(nodes)) {
      for (const child of n.onSuccess ?? []) {
        rfE.push({
          id: `${id}-s->${child}`,
          source: id,
          sourceHandle: 'success',
          target: child,
          style: { stroke: '#10b981', strokeWidth: 2 },
        })
      }
      for (const child of n.onError ?? []) {
        rfE.push({
          id: `${id}-e->${child}`,
          source: id,
          sourceHandle: 'error',
          target: child,
          style: { stroke: '#ef4444', strokeWidth: 2, strokeDasharray: '5 3' },
        })
      }
    }
    return { rfNodes: rfN, rfEdges: rfE }
  }, [roots, nodes, readOnly, selectedId])

  const [flowNodes, setFlowNodes, onNodesChange] = useNodesState(rfNodes)
  const [flowEdges, setFlowEdges, onEdgesChange] = useEdgesState(rfEdges)

  useEffect(() => {
    setFlowNodes(rfNodes)
    setFlowEdges(rfEdges)
  }, [rfNodes, rfEdges, setFlowNodes, setFlowEdges])

  const handleNodesChange = useCallback(
    (changes: NodeChange[]) => {
      for (const c of changes) {
        if (c.type === 'position' && c.position) layoutRef.current[c.id] = c.position
        if (c.type === 'select') {
          if (c.selected && c.id !== TRIGGER_ID) setSelectedId(c.id)
          else if (!c.selected && selectedId === c.id) setSelectedId(null)
        }
      }
      onNodesChange(changes)
    },
    [onNodesChange, selectedId],
  )

  const handleEdgesChange = useCallback(
    (changes: EdgeChange[]) => {
      let modelDirty = false
      const nextRoots = [...roots]
      const nextNodes: Record<string, FlowNode> = Object.fromEntries(
        Object.entries(nodes).map(([id, n]) => [id, { ...n, onSuccess: [...(n.onSuccess ?? [])], onError: [...(n.onError ?? [])] }]),
      )
      for (const c of changes) {
        if (c.type !== 'remove') continue
        const edge = flowEdges.find((e) => e.id === c.id)
        if (!edge) continue
        if (edge.source === TRIGGER_ID) {
          const i = nextRoots.indexOf(edge.target)
          if (i >= 0) {
            nextRoots.splice(i, 1)
            modelDirty = true
          }
          continue
        }
        const src = nextNodes[edge.source]
        if (!src) continue
        if (edge.sourceHandle === 'success') {
          src.onSuccess = src.onSuccess?.filter((t) => t !== edge.target)
          modelDirty = true
        } else if (edge.sourceHandle === 'error') {
          src.onError = src.onError?.filter((t) => t !== edge.target)
          modelDirty = true
        }
      }
      if (modelDirty) onChange({ roots: nextRoots, nodes: nextNodes })
      onEdgesChange(changes)
    },
    [flowEdges, roots, nodes, onChange, onEdgesChange],
  )

  const onConnect = useCallback(
    (conn: Connection) => {
      if (!conn.source || !conn.target) return
      if (conn.source === conn.target) return
      const nextRoots = [...roots]
      const nextNodes: Record<string, FlowNode> = Object.fromEntries(
        Object.entries(nodes).map(([id, n]) => [id, { ...n, onSuccess: [...(n.onSuccess ?? [])], onError: [...(n.onError ?? [])] }]),
      )
      if (conn.source === TRIGGER_ID) {
        if (!nextRoots.includes(conn.target)) nextRoots.push(conn.target)
      } else {
        const src = nextNodes[conn.source]
        if (!src) return
        const list = conn.sourceHandle === 'error' ? (src.onError ??= []) : (src.onSuccess ??= [])
        if (!list.includes(conn.target)) list.push(conn.target)
      }
      onChange({ roots: nextRoots, nodes: nextNodes })
      setFlowEdges((es) => addEdge(conn, es))
    },
    [roots, nodes, onChange, setFlowEdges],
  )

  // Drag the endpoint of an existing edge off any handle to disconnect. If
  // it's dropped onto a valid handle we re-route instead.
  const reconnectOk = useRef(true)
  const removeEdgeFromModel = useCallback(
    (edge: Edge, base: { roots: string[]; nodes: Record<string, FlowNode> }) => {
      if (edge.source === TRIGGER_ID) {
        base.roots = base.roots.filter((r) => r !== edge.target)
        return
      }
      const src = base.nodes[edge.source]
      if (!src) return
      if (edge.sourceHandle === 'error') src.onError = src.onError?.filter((t) => t !== edge.target)
      else src.onSuccess = src.onSuccess?.filter((t) => t !== edge.target)
    },
    [],
  )
  const addConnToModel = useCallback((conn: Connection, base: { roots: string[]; nodes: Record<string, FlowNode> }) => {
    if (!conn.source || !conn.target || conn.source === conn.target) return
    if (conn.source === TRIGGER_ID) {
      if (!base.roots.includes(conn.target)) base.roots.push(conn.target)
      return
    }
    const src = base.nodes[conn.source]
    if (!src) return
    const list = conn.sourceHandle === 'error' ? (src.onError ??= []) : (src.onSuccess ??= [])
    if (!list.includes(conn.target)) list.push(conn.target)
  }, [])
  const cloneModel = useCallback(
    () => ({
      roots: [...roots],
      nodes: Object.fromEntries(
        Object.entries(nodes).map(([id, n]) => [id, { ...n, onSuccess: [...(n.onSuccess ?? [])], onError: [...(n.onError ?? [])] }]),
      ),
    }),
    [roots, nodes],
  )
  const onReconnectStart = useCallback(() => {
    reconnectOk.current = false
  }, [])
  const onReconnect = useCallback(
    (oldEdge: Edge, newConnection: Connection) => {
      reconnectOk.current = true
      const base = cloneModel()
      removeEdgeFromModel(oldEdge, base)
      addConnToModel(newConnection, base)
      onChange(base)
    },
    [cloneModel, removeEdgeFromModel, addConnToModel, onChange],
  )
  const onReconnectEnd = useCallback(
    (_: unknown, edge: Edge) => {
      if (reconnectOk.current) return
      const base = cloneModel()
      removeEdgeFromModel(edge, base)
      onChange(base)
    },
    [cloneModel, removeEdgeFromModel, onChange],
  )

  const onDragOver = useCallback((event: React.DragEvent) => {
    event.preventDefault()
    event.dataTransfer.dropEffect = 'move'
  }, [])

  const onDrop = useCallback(
    (event: React.DragEvent) => {
      event.preventDefault()
      if (readOnly) return
      const raw = event.dataTransfer.getData('application/soar-node')
      if (!raw) return
      let payload: { executor: string; kind: NodeKind; paramsPlaceholder?: unknown }
      try {
        payload = JSON.parse(raw)
      } catch {
        return
      }
      const position = screenToFlowPosition({ x: event.clientX, y: event.clientY })
      const id = uniqueId(payload.executor, nodes)
      const newNode: FlowNode = {
        kind: payload.kind,
        executor: payload.executor,
        params: payload.paramsPlaceholder,
      }
      layoutRef.current[id] = position
      onChange({ roots, nodes: { ...nodes, [id]: newNode } })
      setSelectedId(id)
    },
    [readOnly, screenToFlowPosition, roots, nodes, onChange],
  )

  const selected = selectedId ? nodes[selectedId] : null

  const renameNode = (nextId: string) => {
    if (!selectedId || nextId === selectedId || nodes[nextId]) return
    const nextNodes: Record<string, FlowNode> = {}
    for (const [id, n] of Object.entries(nodes)) {
      const copy: FlowNode = {
        ...n,
        onSuccess: n.onSuccess?.map((t) => (t === selectedId ? nextId : t)),
        onError: n.onError?.map((t) => (t === selectedId ? nextId : t)),
      }
      nextNodes[id === selectedId ? nextId : id] = copy
    }
    const nextRoots = roots.map((r) => (r === selectedId ? nextId : r))
    layoutRef.current[nextId] = layoutRef.current[selectedId]
    delete layoutRef.current[selectedId]
    onChange({ roots: nextRoots, nodes: nextNodes })
    setSelectedId(nextId)
  }

  const patchNode = (patch: Partial<FlowNode>) => {
    if (!selectedId || !nodes[selectedId]) return
    onChange({ roots, nodes: { ...nodes, [selectedId]: { ...nodes[selectedId], ...patch } } })
  }

  const deleteNode = () => {
    if (!selectedId) return
    const nextNodes: Record<string, FlowNode> = {}
    for (const [id, n] of Object.entries(nodes)) {
      if (id === selectedId) continue
      nextNodes[id] = {
        ...n,
        onSuccess: n.onSuccess?.filter((t) => t !== selectedId),
        onError: n.onError?.filter((t) => t !== selectedId),
      }
    }
    const nextRoots = roots.filter((r) => r !== selectedId)
    delete layoutRef.current[selectedId]
    onChange({ roots: nextRoots, nodes: nextNodes })
    setSelectedId(null)
  }

  return (
    <div className="flex h-[560px] overflow-hidden rounded-lg border border-border">
      {paletteOpen ? (
        <div className="relative">
          <NodePalette readOnly={readOnly} />
          <button
            type="button"
            onClick={() => setPaletteOpen(false)}
            className="absolute right-1 top-1 flex h-6 w-6 items-center justify-center rounded text-muted-foreground hover:bg-muted hover:text-foreground"
            title={t('soar.editor.canvas.hidePalette')}
          >
            <ChevronLeft size={13} />
          </button>
        </div>
      ) : (
        <CollapsedRail side="left" label={t('soar.editor.canvas.paletteLabel')} onClick={() => setPaletteOpen(true)} />
      )}

      <div ref={wrapperRef} className="relative flex-1" onDragOver={onDragOver} onDrop={onDrop}>
        <ReactFlow
          nodes={flowNodes}
          edges={flowEdges}
          nodeTypes={NODE_TYPES}
          onNodesChange={handleNodesChange}
          onEdgesChange={handleEdgesChange}
          onConnect={onConnect}
          onReconnect={onReconnect}
          onReconnectStart={onReconnectStart}
          onReconnectEnd={onReconnectEnd}
          onPaneClick={() => setSelectedId(null)}
          colorMode={theme}
          nodesDraggable={!readOnly}
          nodesConnectable={!readOnly}
          edgesFocusable={!readOnly}
          fitView
          fitViewOptions={{ padding: 0.2 }}
          proOptions={{ hideAttribution: true }}
          style={{
            '--xy-background-color-default': 'var(--background)',
            '--xy-controls-button-background-color-default': 'var(--card)',
            '--xy-controls-button-background-color-hover-default': 'var(--muted)',
            '--xy-controls-button-color-default': 'var(--foreground)',
            '--xy-controls-button-color-hover-default': 'var(--foreground)',
            '--xy-controls-button-border-color-default': 'var(--border)',
          } as React.CSSProperties}
        >
          <Background gap={16} size={1} />
          <Controls showInteractive={false} />
        </ReactFlow>
      </div>

      {selected && selectedId ? (
        inspectorOpen ? (
          <div className="relative">
            <NodeInspector
              nodeId={selectedId}
              node={selected}
              nodes={nodes}
              readOnly={readOnly}
              onRename={renameNode}
              onChange={patchNode}
              onDelete={deleteNode}
            />
            <button
              type="button"
              onClick={() => setInspectorOpen(false)}
              className="absolute left-1 top-1 flex h-6 w-6 items-center justify-center rounded text-muted-foreground hover:bg-muted hover:text-foreground"
              title={t('soar.editor.canvas.hideInspector')}
            >
              <ChevronRight size={13} />
            </button>
          </div>
        ) : (
          <CollapsedRail side="right" label={t('soar.editor.canvas.nodeShort')} onClick={() => setInspectorOpen(true)} />
        )
      ) : null}
    </div>
  )
}

function CollapsedRail({ side, label, onClick }: { side: 'left' | 'right'; label: string; onClick: () => void }) {
  const { t } = useTranslation()
  const Icon = side === 'left' ? PanelLeft : PanelRight
  return (
    <button
      type="button"
      onClick={onClick}
      className={cn(
        'flex w-8 shrink-0 flex-col items-center gap-2 border-border bg-card py-2 text-muted-foreground hover:bg-muted hover:text-foreground',
        side === 'left' ? 'border-r' : 'border-l',
      )}
      title={t('soar.editor.canvas.showPanel', { name: label })}
    >
      <Icon size={13} />
      <span
        className="text-[10px] font-semibold uppercase tracking-wider"
        style={{ writingMode: 'vertical-rl', transform: 'rotate(180deg)' }}
      >
        {label}
      </span>
    </button>
  )
}

function uniqueId(executor: string, existing: Record<string, FlowNode>): string {
  let i = 1
  let id = executor
  while (existing[id]) {
    id = `${executor}_${i++}`
  }
  return id
}
